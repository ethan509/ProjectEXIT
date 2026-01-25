package lotto

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrDrawNotFound = errors.New("draw not found")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetLatestDrawNo DB에서 최신 회차 번호 조회
func (r *Repository) GetLatestDrawNo(ctx context.Context) (int, error) {
	var drawNo int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(draw_no), 0) FROM lotto_draws",
	).Scan(&drawNo)
	if err != nil {
		return 0, err
	}
	return drawNo, nil
}

// GetDrawByNo 특정 회차 당첨번호 조회
func (r *Repository) GetDrawByNo(ctx context.Context, drawNo int) (*LottoDraw, error) {
	draw := &LottoDraw{}
	err := r.db.QueryRowContext(ctx,
		`SELECT draw_no, draw_date, num1, num2, num3, num4, num5, num6,
		        bonus_num, first_prize, first_winners, created_at, updated_at
		 FROM lotto_draws WHERE draw_no = $1`, drawNo,
	).Scan(
		&draw.DrawNo, &draw.DrawDate,
		&draw.Num1, &draw.Num2, &draw.Num3, &draw.Num4, &draw.Num5, &draw.Num6,
		&draw.BonusNum, &draw.FirstPrize, &draw.FirstWinners,
		&draw.CreatedAt, &draw.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrDrawNotFound
	}
	if err != nil {
		return nil, err
	}
	return draw, nil
}

// GetDraw GetDrawByNo의 별칭
func (r *Repository) GetDraw(ctx context.Context, drawNo int) (*LottoDraw, error) {
	return r.GetDrawByNo(ctx, drawNo)
}

// GetDraws 당첨번호 목록 조회 (최신순)
func (r *Repository) GetDraws(ctx context.Context, limit, offset int) ([]LottoDraw, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT draw_no, draw_date, num1, num2, num3, num4, num5, num6,
		        bonus_num, first_prize, first_winners, created_at, updated_at
		 FROM lotto_draws ORDER BY draw_no DESC LIMIT $1 OFFSET $2`, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var draws []LottoDraw
	for rows.Next() {
		var draw LottoDraw
		if err := rows.Scan(
			&draw.DrawNo, &draw.DrawDate,
			&draw.Num1, &draw.Num2, &draw.Num3, &draw.Num4, &draw.Num5, &draw.Num6,
			&draw.BonusNum, &draw.FirstPrize, &draw.FirstWinners,
			&draw.CreatedAt, &draw.UpdatedAt,
		); err != nil {
			return nil, err
		}
		draws = append(draws, draw)
	}
	return draws, rows.Err()
}

// GetAllDraws 모든 당첨번호 조회 (회차순)
func (r *Repository) GetAllDraws(ctx context.Context) ([]*LottoDraw, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT draw_no, draw_date, num1, num2, num3, num4, num5, num6,
		        bonus_num, first_prize, first_winners, created_at, updated_at
		 FROM lotto_draws ORDER BY draw_no ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var draws []*LottoDraw
	for rows.Next() {
		draw := &LottoDraw{}
		if err := rows.Scan(
			&draw.DrawNo, &draw.DrawDate,
			&draw.Num1, &draw.Num2, &draw.Num3, &draw.Num4, &draw.Num5, &draw.Num6,
			&draw.BonusNum, &draw.FirstPrize, &draw.FirstWinners,
			&draw.CreatedAt, &draw.UpdatedAt,
		); err != nil {
			return nil, err
		}
		draws = append(draws, draw)
	}
	return draws, rows.Err()
}

// GetTotalDrawCount 전체 당첨번호 개수 조회
func (r *Repository) GetTotalDrawCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM lotto_draws").Scan(&count)
	return count, err
}

// SaveDraw 당첨번호 저장
func (r *Repository) SaveDraw(ctx context.Context, draw *LottoDraw) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO lotto_draws (draw_no, draw_date, num1, num2, num3, num4, num5, num6,
		                          bonus_num, first_prize, first_winners)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 ON CONFLICT (draw_no) DO UPDATE SET
		     draw_date = EXCLUDED.draw_date,
		     num1 = EXCLUDED.num1, num2 = EXCLUDED.num2, num3 = EXCLUDED.num3,
		     num4 = EXCLUDED.num4, num5 = EXCLUDED.num5, num6 = EXCLUDED.num6,
		     bonus_num = EXCLUDED.bonus_num,
		     first_prize = EXCLUDED.first_prize, first_winners = EXCLUDED.first_winners,
		     updated_at = NOW()`,
		draw.DrawNo, draw.DrawDate,
		draw.Num1, draw.Num2, draw.Num3, draw.Num4, draw.Num5, draw.Num6,
		draw.BonusNum, draw.FirstPrize, draw.FirstWinners,
	)
	return err
}

// SaveDrawBatch 당첨번호 일괄 저장
func (r *Repository) SaveDrawBatch(ctx context.Context, draws []LottoDraw) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO lotto_draws (draw_no, draw_date, num1, num2, num3, num4, num5, num6,
		                          bonus_num, first_prize, first_winners)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 ON CONFLICT (draw_no) DO UPDATE SET
		     draw_date = EXCLUDED.draw_date,
		     num1 = EXCLUDED.num1, num2 = EXCLUDED.num2, num3 = EXCLUDED.num3,
		     num4 = EXCLUDED.num4, num5 = EXCLUDED.num5, num6 = EXCLUDED.num6,
		     bonus_num = EXCLUDED.bonus_num,
		     first_prize = EXCLUDED.first_prize, first_winners = EXCLUDED.first_winners,
		     updated_at = NOW()`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, draw := range draws {
		_, err := stmt.ExecContext(ctx,
			draw.DrawNo, draw.DrawDate,
			draw.Num1, draw.Num2, draw.Num3, draw.Num4, draw.Num5, draw.Num6,
			draw.BonusNum, draw.FirstPrize, draw.FirstWinners,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetAllNumberStats 모든 번호 통계 조회
func (r *Repository) GetAllNumberStats(ctx context.Context) ([]NumberStat, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, number, total_count, bonus_count, last_draw_no, calculated_at
		 FROM lotto_number_stats ORDER BY number ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []NumberStat
	for rows.Next() {
		var stat NumberStat
		var lastDrawNo sql.NullInt64
		if err := rows.Scan(
			&stat.ID, &stat.Number, &stat.TotalCount, &stat.BonusCount,
			&lastDrawNo, &stat.CalculatedAt,
		); err != nil {
			return nil, err
		}
		if lastDrawNo.Valid {
			stat.LastDrawNo = int(lastDrawNo.Int64)
		}
		stats = append(stats, stat)
	}
	return stats, rows.Err()
}

// UpsertNumberStats 번호 통계 일괄 저장/업데이트
func (r *Repository) UpsertNumberStats(ctx context.Context, stats []NumberStat) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO lotto_number_stats (number, total_count, bonus_count, last_draw_no, calculated_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (number) DO UPDATE SET
		     total_count = EXCLUDED.total_count,
		     bonus_count = EXCLUDED.bonus_count,
		     last_draw_no = EXCLUDED.last_draw_no,
		     calculated_at = EXCLUDED.calculated_at,
		     updated_at = NOW()`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now()
	for _, stat := range stats {
		_, err := stmt.ExecContext(ctx,
			stat.Number, stat.TotalCount, stat.BonusCount, stat.LastDrawNo, now,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetAllReappearStats 모든 재등장 통계 조회
func (r *Repository) GetAllReappearStats(ctx context.Context) ([]ReappearStat, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT number, total_appear, reappear_count, probability
		 FROM lotto_reappear_stats ORDER BY number ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []ReappearStat
	for rows.Next() {
		var stat ReappearStat
		if err := rows.Scan(
			&stat.Number, &stat.TotalAppear, &stat.ReappearCount, &stat.Probability,
		); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	return stats, rows.Err()
}

// UpsertReappearStats 재등장 통계 일괄 저장/업데이트
func (r *Repository) UpsertReappearStats(ctx context.Context, stats []ReappearStat) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO lotto_reappear_stats (number, total_appear, reappear_count, probability, calculated_at)
		 VALUES ($1, $2, $3, $4, NOW())
		 ON CONFLICT (number) DO UPDATE SET
		     total_appear = EXCLUDED.total_appear,
		     reappear_count = EXCLUDED.reappear_count,
		     probability = EXCLUDED.probability,
		     calculated_at = NOW()`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, stat := range stats {
		_, err := stmt.ExecContext(ctx,
			stat.Number, stat.TotalAppear, stat.ReappearCount, stat.Probability,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
