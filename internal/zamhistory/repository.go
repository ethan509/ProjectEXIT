package zamhistory

import (
	"context"
	"database/sql"
	"time"
)

// Repository handles database operations for zam_daily_history.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new zamhistory Repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// UpsertDailyHistory inserts or updates a daily history aggregation row.
func (r *Repository) UpsertDailyHistory(ctx context.Context, userID int64, historyDate time.Time, txType string, amount int64, count int) error {
	query := `
		INSERT INTO zam_daily_history (user_id, history_date, tx_type, total_amount, tx_count, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (user_id, history_date, tx_type) DO UPDATE SET
			total_amount = zam_daily_history.total_amount + EXCLUDED.total_amount,
			tx_count     = zam_daily_history.tx_count + EXCLUDED.tx_count,
			updated_at   = NOW()
	`
	_, err := r.db.ExecContext(ctx, query, userID, historyDate, txType, amount, count)
	return err
}

// GetDailyHistory retrieves daily history for a user within a date range.
func (r *Repository) GetDailyHistory(ctx context.Context, userID int64, from, to time.Time) ([]DailyHistory, error) {
	query := `
		SELECT user_id, history_date, tx_type, total_amount, tx_count, updated_at
		FROM zam_daily_history
		WHERE user_id = $1 AND history_date >= $2 AND history_date <= $3
		ORDER BY history_date DESC, tx_type ASC
	`
	rows, err := r.db.QueryContext(ctx, query, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []DailyHistory
	for rows.Next() {
		var h DailyHistory
		if err := rows.Scan(&h.UserID, &h.HistoryDate, &h.TxType, &h.TotalAmount, &h.TxCount, &h.UpdatedAt); err != nil {
			return nil, err
		}
		results = append(results, h)
	}
	return results, rows.Err()
}
