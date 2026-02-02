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

// InsertDraw 당첨번호 저장
func (r *Repository) InsertDraw(ctx context.Context, draw *LottoDraw) error {
	query := `
		INSERT INTO lotto_draws (
			draw_no, draw_date, num1, num2, num3, num4, num5, num6,
			bonus_num, first_prize, first_winners, first_per_game,
			second_prize, second_winners, second_per_game,
			third_prize, third_winners, third_per_game,
			fourth_prize, fourth_winners, fourth_per_game,
			fifth_prize, fifth_winners, fifth_per_game,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24,
			NOW(), NOW()
		) ON CONFLICT (draw_no) DO UPDATE SET
			draw_date = EXCLUDED.draw_date,
			num1 = EXCLUDED.num1, num2 = EXCLUDED.num2, num3 = EXCLUDED.num3,
			num4 = EXCLUDED.num4, num5 = EXCLUDED.num5, num6 = EXCLUDED.num6,
			bonus_num = EXCLUDED.bonus_num,
			first_prize = EXCLUDED.first_prize, first_winners = EXCLUDED.first_winners, first_per_game = EXCLUDED.first_per_game,
			second_prize = EXCLUDED.second_prize, second_winners = EXCLUDED.second_winners, second_per_game = EXCLUDED.second_per_game,
			third_prize = EXCLUDED.third_prize, third_winners = EXCLUDED.third_winners, third_per_game = EXCLUDED.third_per_game,
			fourth_prize = EXCLUDED.fourth_prize, fourth_winners = EXCLUDED.fourth_winners, fourth_per_game = EXCLUDED.fourth_per_game,
			fifth_prize = EXCLUDED.fifth_prize, fifth_winners = EXCLUDED.fifth_winners, fifth_per_game = EXCLUDED.fifth_per_game,
			updated_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query,
		draw.DrawNo, draw.DrawDate,
		draw.Num1, draw.Num2, draw.Num3, draw.Num4, draw.Num5, draw.Num6,
		draw.BonusNum,
		draw.FirstPrize, draw.FirstWinners, draw.FirstPerGame,
		draw.SecondPrize, draw.SecondWinners, draw.SecondPerGame,
		draw.ThirdPrize, draw.ThirdWinners, draw.ThirdPerGame,
		draw.FourthPrize, draw.FourthWinners, draw.FourthPerGame,
		draw.FifthPrize, draw.FifthWinners, draw.FifthPerGame,
	)
	return err
}

// InsertDraws 여러 당첨번호 일괄 저장
func (r *Repository) InsertDraws(ctx context.Context, draws []LottoDraw) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO lotto_draws (
			draw_no, draw_date, num1, num2, num3, num4, num5, num6,
			bonus_num, first_prize, first_winners, first_per_game,
			second_prize, second_winners, second_per_game,
			third_prize, third_winners, third_per_game,
			fourth_prize, fourth_winners, fourth_per_game,
			fifth_prize, fifth_winners, fifth_per_game,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24,
			NOW(), NOW()
		) ON CONFLICT (draw_no) DO NOTHING
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, draw := range draws {
		_, err := stmt.ExecContext(ctx,
			draw.DrawNo, draw.DrawDate,
			draw.Num1, draw.Num2, draw.Num3, draw.Num4, draw.Num5, draw.Num6,
			draw.BonusNum,
			draw.FirstPrize, draw.FirstWinners, draw.FirstPerGame,
			draw.SecondPrize, draw.SecondWinners, draw.SecondPerGame,
			draw.ThirdPrize, draw.ThirdWinners, draw.ThirdPerGame,
			draw.FourthPrize, draw.FourthWinners, draw.FourthPerGame,
			draw.FifthPrize, draw.FifthWinners, draw.FifthPerGame,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// UnclaimedPrize Methods

// GetUnclaimedPrizes 미수령 당첨금 조회 (모두)
func (r *Repository) GetUnclaimedPrizes(ctx context.Context) ([]UnclaimedPrize, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, draw_no, prize_rank, amount, winner_name, winning_date, expiration_date, status, created_at, updated_at
		 FROM unclaimed_prizes WHERE status = 'unclaimed' ORDER BY expiration_date ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prizes []UnclaimedPrize
	for rows.Next() {
		var prize UnclaimedPrize
		if err := rows.Scan(
			&prize.ID, &prize.DrawNo, &prize.PrizeRank, &prize.Amount, &prize.WinnerName,
			&prize.WinningDate, &prize.ExpirationDate, &prize.Status, &prize.CreatedAt, &prize.UpdatedAt,
		); err != nil {
			return nil, err
		}
		prizes = append(prizes, prize)
	}
	return prizes, rows.Err()
}

// GetUnclaimedPrizesByDrawNo 특정 회차의 미수령 당첨금 조회
func (r *Repository) GetUnclaimedPrizesByDrawNo(ctx context.Context, drawNo int) ([]UnclaimedPrize, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, draw_no, prize_rank, amount, winner_name, winning_date, expiration_date, status, created_at, updated_at
		 FROM unclaimed_prizes WHERE draw_no = $1 AND status = 'unclaimed'`, drawNo,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prizes []UnclaimedPrize
	for rows.Next() {
		var prize UnclaimedPrize
		if err := rows.Scan(
			&prize.ID, &prize.DrawNo, &prize.PrizeRank, &prize.Amount, &prize.WinnerName,
			&prize.WinningDate, &prize.ExpirationDate, &prize.Status, &prize.CreatedAt, &prize.UpdatedAt,
		); err != nil {
			return nil, err
		}
		prizes = append(prizes, prize)
	}
	return prizes, rows.Err()
}

// InsertUnclaimedPrize 미수령 당첨금 저장
func (r *Repository) InsertUnclaimedPrize(ctx context.Context, prize *UnclaimedPrize) error {
	query := `
		INSERT INTO unclaimed_prizes (
			draw_no, prize_rank, amount, winner_name, winning_date, expiration_date, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		ON CONFLICT DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query,
		prize.DrawNo, prize.PrizeRank, prize.Amount, prize.WinnerName,
		prize.WinningDate, prize.ExpirationDate, prize.Status,
	)
	return err
}

// InsertUnclaimedPrizes 미수령 당첨금 일괄 저장
func (r *Repository) InsertUnclaimedPrizes(ctx context.Context, prizes []UnclaimedPrize) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO unclaimed_prizes (
			draw_no, prize_rank, amount, winner_name, winning_date, expiration_date, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, prize := range prizes {
		_, err := stmt.ExecContext(ctx,
			prize.DrawNo, prize.PrizeRank, prize.Amount, prize.WinnerName,
			prize.WinningDate, prize.ExpirationDate, prize.Status,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// UpdateUnclaimedPrizeStatus 미수령 당첨금 상태 업데이트
func (r *Repository) UpdateUnclaimedPrizeStatus(ctx context.Context, prizeID int64, status string) error {
	query := `UPDATE unclaimed_prizes SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, prizeID)
	return err
}

// DeleteExpiredUnclaimedPrizes 만기 지난 미수령 당첨금 삭제
func (r *Repository) DeleteExpiredUnclaimedPrizes(ctx context.Context) error {
	query := `DELETE FROM unclaimed_prizes WHERE expiration_date < NOW() AND status = 'unclaimed'`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

// Bayesian Stats Methods

// GetLatestBayesianStats 가장 최근 회차의 베이지안 통계 조회 (45개 번호 전체)
func (r *Repository) GetLatestBayesianStats(ctx context.Context) ([]BayesianStat, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, draw_no, number, total_count, total_draws, prior, posterior, appeared, calculated_at
		 FROM lotto_bayesian_stats
		 WHERE draw_no = (SELECT COALESCE(MAX(draw_no), 0) FROM lotto_bayesian_stats)
		 ORDER BY number ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []BayesianStat
	for rows.Next() {
		var stat BayesianStat
		if err := rows.Scan(
			&stat.ID, &stat.DrawNo, &stat.Number, &stat.TotalCount, &stat.TotalDraws,
			&stat.Prior, &stat.Posterior, &stat.Appeared, &stat.CalculatedAt,
		); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	return stats, rows.Err()
}

// GetBayesianStatsByDrawNo 특정 회차의 베이지안 통계 조회
func (r *Repository) GetBayesianStatsByDrawNo(ctx context.Context, drawNo int) ([]BayesianStat, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, draw_no, number, total_count, total_draws, prior, posterior, appeared, calculated_at
		 FROM lotto_bayesian_stats
		 WHERE draw_no = $1
		 ORDER BY number ASC`, drawNo,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []BayesianStat
	for rows.Next() {
		var stat BayesianStat
		if err := rows.Scan(
			&stat.ID, &stat.DrawNo, &stat.Number, &stat.TotalCount, &stat.TotalDraws,
			&stat.Prior, &stat.Posterior, &stat.Appeared, &stat.CalculatedAt,
		); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	return stats, rows.Err()
}

// UpsertBayesianStats 베이지안 통계 일괄 저장/업데이트
func (r *Repository) UpsertBayesianStats(ctx context.Context, stats []BayesianStat) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO lotto_bayesian_stats (draw_no, number, total_count, total_draws, prior, posterior, appeared, calculated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		 ON CONFLICT (draw_no, number) DO UPDATE SET
		     total_count = EXCLUDED.total_count,
		     total_draws = EXCLUDED.total_draws,
		     prior = EXCLUDED.prior,
		     posterior = EXCLUDED.posterior,
		     appeared = EXCLUDED.appeared,
		     calculated_at = NOW(),
		     updated_at = NOW()`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, stat := range stats {
		_, err := stmt.ExecContext(ctx,
			stat.DrawNo, stat.Number, stat.TotalCount, stat.TotalDraws,
			stat.Prior, stat.Posterior, stat.Appeared,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetBayesianStatsHistory 특정 번호의 확률 변화 히스토리 조회
func (r *Repository) GetBayesianStatsHistory(ctx context.Context, number int, limit int) ([]BayesianStat, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, draw_no, number, total_count, total_draws, prior, posterior, appeared, calculated_at
		 FROM lotto_bayesian_stats
		 WHERE number = $1
		 ORDER BY draw_no DESC
		 LIMIT $2`, number, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []BayesianStat
	for rows.Next() {
		var stat BayesianStat
		if err := rows.Scan(
			&stat.ID, &stat.DrawNo, &stat.Number, &stat.TotalCount, &stat.TotalDraws,
			&stat.Prior, &stat.Posterior, &stat.Appeared, &stat.CalculatedAt,
		); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	return stats, rows.Err()
}

// GetLatestBayesianDrawNo 베이지안 통계가 계산된 가장 최근 회차 번호 조회
func (r *Repository) GetLatestBayesianDrawNo(ctx context.Context) (int, error) {
	var drawNo int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(draw_no), 0) FROM lotto_bayesian_stats",
	).Scan(&drawNo)
	if err != nil {
		return 0, err
	}
	return drawNo, nil
}

// Unified Analysis Stats Methods

// GetLatestAnalysisStats 가장 최근 회차의 통합 분석 통계 조회 (45개 번호 전체)
func (r *Repository) GetLatestAnalysisStats(ctx context.Context) ([]AnalysisStat, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT draw_no, number, total_count, total_prob, bonus_count, bonus_prob, first_count, last_count,
		        reappear_total, reappear_count, reappear_prob,
		        bayesian_prior, bayesian_post, appeared, calculated_at
		 FROM lotto_analysis_stats
		 WHERE draw_no = (SELECT COALESCE(MAX(draw_no), 0) FROM lotto_analysis_stats)
		 ORDER BY number ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []AnalysisStat
	for rows.Next() {
		var stat AnalysisStat
		var totalProb, bonusProb sql.NullFloat64
		var bayesianPrior, bayesianPost sql.NullFloat64
		if err := rows.Scan(
			&stat.DrawNo, &stat.Number, &stat.TotalCount, &totalProb, &stat.BonusCount, &bonusProb,
			&stat.FirstCount, &stat.LastCount,
			&stat.ReappearTotal, &stat.ReappearCount, &stat.ReappearProb,
			&bayesianPrior, &bayesianPost, &stat.Appeared, &stat.CalculatedAt,
		); err != nil {
			return nil, err
		}
		if totalProb.Valid {
			stat.TotalProb = totalProb.Float64
		}
		if bonusProb.Valid {
			stat.BonusProb = bonusProb.Float64
		}
		if bayesianPrior.Valid {
			stat.BayesianPrior = bayesianPrior.Float64
		}
		if bayesianPost.Valid {
			stat.BayesianPost = bayesianPost.Float64
		}
		stats = append(stats, stat)
	}
	return stats, rows.Err()
}

// GetAnalysisStatsByDrawNo 특정 회차의 통합 분석 통계 조회
func (r *Repository) GetAnalysisStatsByDrawNo(ctx context.Context, drawNo int) ([]AnalysisStat, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT draw_no, number, total_count, total_prob, bonus_count, bonus_prob, first_count, last_count,
		        reappear_total, reappear_count, reappear_prob,
		        bayesian_prior, bayesian_post, appeared, calculated_at
		 FROM lotto_analysis_stats
		 WHERE draw_no = $1
		 ORDER BY number ASC`, drawNo,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []AnalysisStat
	for rows.Next() {
		var stat AnalysisStat
		var totalProb, bonusProb sql.NullFloat64
		var bayesianPrior, bayesianPost sql.NullFloat64
		if err := rows.Scan(
			&stat.DrawNo, &stat.Number, &stat.TotalCount, &totalProb, &stat.BonusCount, &bonusProb,
			&stat.FirstCount, &stat.LastCount,
			&stat.ReappearTotal, &stat.ReappearCount, &stat.ReappearProb,
			&bayesianPrior, &bayesianPost, &stat.Appeared, &stat.CalculatedAt,
		); err != nil {
			return nil, err
		}
		if totalProb.Valid {
			stat.TotalProb = totalProb.Float64
		}
		if bonusProb.Valid {
			stat.BonusProb = bonusProb.Float64
		}
		if bayesianPrior.Valid {
			stat.BayesianPrior = bayesianPrior.Float64
		}
		if bayesianPost.Valid {
			stat.BayesianPost = bayesianPost.Float64
		}
		stats = append(stats, stat)
	}
	return stats, rows.Err()
}

// UpsertAnalysisStats 통합 분석 통계 일괄 저장/업데이트
func (r *Repository) UpsertAnalysisStats(ctx context.Context, stats []AnalysisStat) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO lotto_analysis_stats (
			draw_no, number, total_count, total_prob, bonus_count, bonus_prob, first_count, last_count,
			reappear_total, reappear_count, reappear_prob,
			bayesian_prior, bayesian_post, appeared, calculated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW())
		ON CONFLICT (draw_no, number) DO UPDATE SET
			total_count = EXCLUDED.total_count,
			total_prob = EXCLUDED.total_prob,
			bonus_count = EXCLUDED.bonus_count,
			bonus_prob = EXCLUDED.bonus_prob,
			first_count = EXCLUDED.first_count,
			last_count = EXCLUDED.last_count,
			reappear_total = EXCLUDED.reappear_total,
			reappear_count = EXCLUDED.reappear_count,
			reappear_prob = EXCLUDED.reappear_prob,
			bayesian_prior = EXCLUDED.bayesian_prior,
			bayesian_post = EXCLUDED.bayesian_post,
			appeared = EXCLUDED.appeared,
			calculated_at = NOW(),
			updated_at = NOW()`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, stat := range stats {
		_, err := stmt.ExecContext(ctx,
			stat.DrawNo, stat.Number, stat.TotalCount, stat.TotalProb, stat.BonusCount, stat.BonusProb,
			stat.FirstCount, stat.LastCount,
			stat.ReappearTotal, stat.ReappearCount, stat.ReappearProb,
			stat.BayesianPrior, stat.BayesianPost, stat.Appeared,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetAnalysisStatsHistory 특정 번호의 분석 통계 히스토리 조회
func (r *Repository) GetAnalysisStatsHistory(ctx context.Context, number int, limit int) ([]AnalysisStat, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT draw_no, number, total_count, total_prob, bonus_count, bonus_prob, first_count, last_count,
		        reappear_total, reappear_count, reappear_prob,
		        bayesian_prior, bayesian_post, appeared, calculated_at
		 FROM lotto_analysis_stats
		 WHERE number = $1
		 ORDER BY draw_no DESC
		 LIMIT $2`, number, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []AnalysisStat
	for rows.Next() {
		var stat AnalysisStat
		var totalProb, bonusProb sql.NullFloat64
		var bayesianPrior, bayesianPost sql.NullFloat64
		if err := rows.Scan(
			&stat.DrawNo, &stat.Number, &stat.TotalCount, &totalProb, &stat.BonusCount, &bonusProb,
			&stat.FirstCount, &stat.LastCount,
			&stat.ReappearTotal, &stat.ReappearCount, &stat.ReappearProb,
			&bayesianPrior, &bayesianPost, &stat.Appeared, &stat.CalculatedAt,
		); err != nil {
			return nil, err
		}
		if totalProb.Valid {
			stat.TotalProb = totalProb.Float64
		}
		if bonusProb.Valid {
			stat.BonusProb = bonusProb.Float64
		}
		if bayesianPrior.Valid {
			stat.BayesianPrior = bayesianPrior.Float64
		}
		if bayesianPost.Valid {
			stat.BayesianPost = bayesianPost.Float64
		}
		stats = append(stats, stat)
	}
	return stats, rows.Err()
}

// GetLatestAnalysisDrawNo 통합 분석 통계가 계산된 가장 최근 회차 번호 조회
func (r *Repository) GetLatestAnalysisDrawNo(ctx context.Context) (int, error) {
	var drawNo int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(draw_no), 0) FROM lotto_analysis_stats",
	).Scan(&drawNo)
	if err != nil {
		return 0, err
	}
	return drawNo, nil
}

// GetAnalysisStatsWithZeroProb total_prob이 0인 행 조회 (수정 필요한 행)
func (r *Repository) GetAnalysisStatsWithZeroProb(ctx context.Context) ([]AnalysisStat, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT draw_no, number, total_count, total_prob, bonus_count, bonus_prob, first_count, last_count,
		        reappear_total, reappear_count, reappear_prob,
		        bayesian_prior, bayesian_post, appeared, calculated_at
		 FROM lotto_analysis_stats
		 WHERE total_prob = 0 OR total_prob IS NULL
		 ORDER BY draw_no ASC, number ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []AnalysisStat
	for rows.Next() {
		var stat AnalysisStat
		var totalProb, bonusProb sql.NullFloat64
		var bayesianPrior, bayesianPost sql.NullFloat64
		if err := rows.Scan(
			&stat.DrawNo, &stat.Number, &stat.TotalCount, &totalProb, &stat.BonusCount, &bonusProb,
			&stat.FirstCount, &stat.LastCount,
			&stat.ReappearTotal, &stat.ReappearCount, &stat.ReappearProb,
			&bayesianPrior, &bayesianPost, &stat.Appeared, &stat.CalculatedAt,
		); err != nil {
			return nil, err
		}
		if totalProb.Valid {
			stat.TotalProb = totalProb.Float64
		}
		if bonusProb.Valid {
			stat.BonusProb = bonusProb.Float64
		}
		if bayesianPrior.Valid {
			stat.BayesianPrior = bayesianPrior.Float64
		}
		if bayesianPost.Valid {
			stat.BayesianPost = bayesianPost.Float64
		}
		stats = append(stats, stat)
	}
	return stats, rows.Err()
}

// UpdateAnalysisStatsTotalProb total_prob 일괄 업데이트
func (r *Repository) UpdateAnalysisStatsTotalProb(ctx context.Context, updates []AnalysisStat) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`UPDATE lotto_analysis_stats
		 SET total_prob = $1, updated_at = NOW()
		 WHERE draw_no = $2 AND number = $3`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, stat := range updates {
		_, err := stmt.ExecContext(ctx, stat.TotalProb, stat.DrawNo, stat.Number)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetAnalysisStatsWithZeroBonusProb bonus_prob이 0인 행 조회 (수정 필요한 행)
func (r *Repository) GetAnalysisStatsWithZeroBonusProb(ctx context.Context) ([]AnalysisStat, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT draw_no, number, total_count, total_prob, bonus_count, bonus_prob, first_count, last_count,
		        reappear_total, reappear_count, reappear_prob,
		        bayesian_prior, bayesian_post, appeared, calculated_at
		 FROM lotto_analysis_stats
		 WHERE bonus_prob = 0 OR bonus_prob IS NULL
		 ORDER BY draw_no ASC, number ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []AnalysisStat
	for rows.Next() {
		var stat AnalysisStat
		var totalProb, bonusProb sql.NullFloat64
		var bayesianPrior, bayesianPost sql.NullFloat64
		if err := rows.Scan(
			&stat.DrawNo, &stat.Number, &stat.TotalCount, &totalProb, &stat.BonusCount, &bonusProb,
			&stat.FirstCount, &stat.LastCount,
			&stat.ReappearTotal, &stat.ReappearCount, &stat.ReappearProb,
			&bayesianPrior, &bayesianPost, &stat.Appeared, &stat.CalculatedAt,
		); err != nil {
			return nil, err
		}
		if totalProb.Valid {
			stat.TotalProb = totalProb.Float64
		}
		if bonusProb.Valid {
			stat.BonusProb = bonusProb.Float64
		}
		if bayesianPrior.Valid {
			stat.BayesianPrior = bayesianPrior.Float64
		}
		if bayesianPost.Valid {
			stat.BayesianPost = bayesianPost.Float64
		}
		stats = append(stats, stat)
	}
	return stats, rows.Err()
}

// UpdateAnalysisStatsBonusProb bonus_prob 일괄 업데이트
func (r *Repository) UpdateAnalysisStatsBonusProb(ctx context.Context, updates []AnalysisStat) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`UPDATE lotto_analysis_stats
		 SET bonus_prob = $1, updated_at = NOW()
		 WHERE draw_no = $2 AND number = $3`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, stat := range updates {
		_, err := stmt.ExecContext(ctx, stat.BonusProb, stat.DrawNo, stat.Number)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
