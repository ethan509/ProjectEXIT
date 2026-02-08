package zamhistory

import "time"

// DailyHistory represents a single aggregated row in zam_daily_history.
type DailyHistory struct {
	UserID      int64     `json:"user_id"`
	HistoryDate time.Time `json:"history_date"`
	TxType      string    `json:"tx_type"`
	TotalAmount int64     `json:"total_amount"`
	TxCount     int       `json:"tx_count"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// bufferKey is the composite key for in-memory aggregation.
type bufferKey struct {
	UserID      int64
	HistoryDate string // "2006-01-02" format
	TxType      string
}

// bufferValue holds the accumulated amounts and counts.
type bufferValue struct {
	TotalAmount int64
	TxCount     int
}
