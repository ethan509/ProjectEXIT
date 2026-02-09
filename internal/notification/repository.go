package notification

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ========================================
// DeviceToken
// ========================================

// UpsertDeviceToken 디바이스 토큰 등록 또는 갱신
func (r *Repository) UpsertDeviceToken(ctx context.Context, userID int64, token, platform string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO device_tokens (user_id, token, platform, is_active, updated_at)
		VALUES ($1, $2, $3, true, NOW())
		ON CONFLICT (user_id, token)
		DO UPDATE SET platform = $3, is_active = true, updated_at = NOW()`,
		userID, token, platform,
	)
	return err
}

// DeactivateDeviceToken 디바이스 토큰 비활성화
func (r *Repository) DeactivateDeviceToken(ctx context.Context, userID int64, token string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE device_tokens SET is_active = false, updated_at = NOW()
		WHERE user_id = $1 AND token = $2`,
		userID, token,
	)
	return err
}

// GetActiveTokensByUserID 사용자의 활성 디바이스 토큰 조회
func (r *Repository) GetActiveTokensByUserID(ctx context.Context, userID int64) ([]DeviceToken, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, token, platform, is_active, created_at, updated_at
		FROM device_tokens
		WHERE user_id = $1 AND is_active = true`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []DeviceToken
	for rows.Next() {
		var t DeviceToken
		if err := rows.Scan(&t.ID, &t.UserID, &t.Token, &t.Platform, &t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, rows.Err()
}

// ========================================
// WinningCheck
// ========================================

// SaveWinningCheck 당첨 확인 결과 저장
func (r *Repository) SaveWinningCheck(ctx context.Context, wc *WinningCheck) error {
	return r.db.QueryRowContext(ctx, `
		INSERT INTO winning_checks (recommendation_id, user_id, draw_no, matched_numbers, matched_count, bonus_matched, prize_rank)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`,
		wc.RecommendationID, wc.UserID, wc.DrawNo,
		pq.Array(wc.MatchedNumbers), wc.MatchedCount, wc.BonusMatched, wc.PrizeRank,
	).Scan(&wc.ID, &wc.CreatedAt)
}

// GetWinnersByDrawNo 특정 회차의 당첨자 조회 (1~5등)
func (r *Repository) GetWinnersByDrawNo(ctx context.Context, drawNo int) ([]WinningCheck, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, recommendation_id, user_id, draw_no, matched_numbers, matched_count, bonus_matched, prize_rank, created_at
		FROM winning_checks
		WHERE draw_no = $1 AND prize_rank IS NOT NULL
		ORDER BY prize_rank ASC`,
		drawNo,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanWinningChecks(rows)
}

// GetWinningsByUserID 사용자의 당첨 결과 조회
func (r *Repository) GetWinningsByUserID(ctx context.Context, userID int64, limit, offset int) ([]WinningCheck, int, error) {
	if limit <= 0 {
		limit = 20
	}

	var totalCount int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM winning_checks
		WHERE user_id = $1 AND prize_rank IS NOT NULL`,
		userID,
	).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, recommendation_id, user_id, draw_no, matched_numbers, matched_count, bonus_matched, prize_rank, created_at
		FROM winning_checks
		WHERE user_id = $1 AND prize_rank IS NOT NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	checks, err := scanWinningChecks(rows)
	if err != nil {
		return nil, 0, err
	}
	return checks, totalCount, nil
}

// CheckAlreadyProcessed 해당 회차가 이미 처리되었는지 확인
func (r *Repository) CheckAlreadyProcessed(ctx context.Context, drawNo int) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM winning_checks WHERE draw_no = $1`,
		drawNo,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func scanWinningChecks(rows *sql.Rows) ([]WinningCheck, error) {
	var checks []WinningCheck
	for rows.Next() {
		var wc WinningCheck
		if err := rows.Scan(
			&wc.ID, &wc.RecommendationID, &wc.UserID, &wc.DrawNo,
			pq.Array(&wc.MatchedNumbers), &wc.MatchedCount, &wc.BonusMatched,
			&wc.PrizeRank, &wc.CreatedAt,
		); err != nil {
			return nil, err
		}
		checks = append(checks, wc)
	}
	return checks, rows.Err()
}

// ========================================
// PushNotification
// ========================================

// SavePushNotification 푸시 알림 기록 저장
func (r *Repository) SavePushNotification(ctx context.Context, pn *PushNotification) error {
	return r.db.QueryRowContext(ctx, `
		INSERT INTO push_notifications (user_id, type, title, body, data, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`,
		pn.UserID, pn.Type, pn.Title, pn.Body, pn.Data, pn.Status,
	).Scan(&pn.ID, &pn.CreatedAt)
}

// UpdatePushStatus 푸시 알림 상태 업데이트
func (r *Repository) UpdatePushStatus(ctx context.Context, id int64, status string, errMsg *string) error {
	var sentAt *time.Time
	if status == "sent" {
		now := time.Now()
		sentAt = &now
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE push_notifications SET status = $2, error_message = $3, sent_at = $4
		WHERE id = $1`,
		id, status, errMsg, sentAt,
	)
	return err
}

// GetNotificationsByUserID 사용자의 알림 목록 조회
func (r *Repository) GetNotificationsByUserID(ctx context.Context, userID int64, limit, offset int) ([]PushNotification, int, error) {
	if limit <= 0 {
		limit = 20
	}

	var totalCount int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM push_notifications WHERE user_id = $1`,
		userID,
	).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, type, title, body, data, status, error_message, created_at, sent_at
		FROM push_notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var notifications []PushNotification
	for rows.Next() {
		var pn PushNotification
		if err := rows.Scan(
			&pn.ID, &pn.UserID, &pn.Type, &pn.Title, &pn.Body,
			&pn.Data, &pn.Status, &pn.ErrorMessage, &pn.CreatedAt, &pn.SentAt,
		); err != nil {
			return nil, 0, err
		}
		notifications = append(notifications, pn)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return notifications, totalCount, nil
}

// ========================================
// Recommendations 조회 (lotto_recommendations 테이블)
// ========================================

// GetRecentRecommendations 최근 7일간의 추천 기록 조회
func (r *Repository) GetRecentRecommendations(ctx context.Context, since time.Time) ([]RecommendationRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, method_codes, numbers, bonus_number, confidence, created_at
		FROM lotto_recommendations
		WHERE created_at >= $1 AND user_id IS NOT NULL
		ORDER BY created_at DESC`,
		since,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recs []RecommendationRow
	for rows.Next() {
		var rec RecommendationRow
		if err := rows.Scan(
			&rec.ID, &rec.UserID, pq.Array(&rec.MethodCodes), pq.Array(&rec.Numbers),
			&rec.BonusNumber, &rec.Confidence, &rec.CreatedAt,
		); err != nil {
			return nil, err
		}
		recs = append(recs, rec)
	}
	return recs, rows.Err()
}

// RecommendationRow lotto_recommendations 테이블 조회 결과
type RecommendationRow struct {
	ID          int64
	UserID      *int64
	MethodCodes []string
	Numbers     []int
	BonusNumber *int
	Confidence  float64
	CreatedAt   time.Time
}
