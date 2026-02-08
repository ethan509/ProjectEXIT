package zamhistory

import (
	"context"
	"sync"
	"time"

	"github.com/example/LottoSmash/internal/logger"
)

const defaultFlushInterval = 5 * time.Minute

// store abstracts the DB operations needed by Buffer for testability.
type store interface {
	UpsertDailyHistory(ctx context.Context, userID int64, historyDate time.Time, txType string, amount int64, count int) error
}

// Buffer accumulates Zam events in memory and periodically flushes
// aggregated totals to the zam_daily_history table.
type Buffer struct {
	mu       sync.Mutex
	data     map[bufferKey]*bufferValue
	repo     store
	log      *logger.Logger
	loc      *time.Location
	interval time.Duration
}

// NewBuffer creates a new Buffer with the given repository, logger, and timezone location.
func NewBuffer(repo *Repository, log *logger.Logger, loc *time.Location) *Buffer {
	if loc == nil {
		loc = time.UTC
	}
	return &Buffer{
		data:     make(map[bufferKey]*bufferValue),
		repo:     repo,
		log:      log,
		loc:      loc,
		interval: defaultFlushInterval,
	}
}

// Record adds a Zam event to the in-memory buffer.
// Safe for concurrent use.
func (b *Buffer) Record(userID int64, amount int64, txType string) {
	key := bufferKey{
		UserID:      userID,
		HistoryDate: time.Now().In(b.loc).Format("2006-01-02"),
		TxType:      txType,
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	val, ok := b.data[key]
	if !ok {
		val = &bufferValue{}
		b.data[key] = val
	}
	val.TotalAmount += amount
	val.TxCount++
}

// Start begins the periodic flush goroutine. Blocks until ctx is cancelled.
func (b *Buffer) Start(ctx context.Context) {
	ticker := time.NewTicker(b.interval)
	defer ticker.Stop()

	b.log.Infof("zam history buffer started (flush interval: %s)", b.interval)

	for {
		select {
		case <-ctx.Done():
			b.log.Infof("zam history buffer stopping")
			return
		case <-ticker.C:
			b.flush(ctx)
		}
	}
}

// Flush drains the buffer and writes all accumulated data to the database.
// Exported for explicit shutdown flush.
func (b *Buffer) Flush(ctx context.Context) {
	b.flush(ctx)
}

// flush swaps the buffer and writes the snapshot to DB.
func (b *Buffer) flush(ctx context.Context) {
	b.mu.Lock()
	if len(b.data) == 0 {
		b.mu.Unlock()
		return
	}
	snapshot := b.data
	b.data = make(map[bufferKey]*bufferValue, len(snapshot))
	b.mu.Unlock()

	b.log.Infof("flushing %d zam history entries", len(snapshot))

	var failedKeys []bufferKey
	var failedVals []*bufferValue

	for key, val := range snapshot {
		historyDate, err := time.Parse("2006-01-02", key.HistoryDate)
		if err != nil {
			b.log.Errorf("invalid history date %s: %v", key.HistoryDate, err)
			continue
		}

		if err := b.repo.UpsertDailyHistory(ctx, key.UserID, historyDate, key.TxType, val.TotalAmount, val.TxCount); err != nil {
			b.log.Errorf("failed to upsert zam daily history (user=%d, date=%s, type=%s): %v",
				key.UserID, key.HistoryDate, key.TxType, err)
			failedKeys = append(failedKeys, key)
			failedVals = append(failedVals, val)
		}
	}

	// Re-merge failed entries back into the live buffer for retry.
	if len(failedKeys) > 0 {
		b.mu.Lock()
		for i, key := range failedKeys {
			if existing, ok := b.data[key]; ok {
				existing.TotalAmount += failedVals[i].TotalAmount
				existing.TxCount += failedVals[i].TxCount
			} else {
				b.data[key] = failedVals[i]
			}
		}
		b.mu.Unlock()
		b.log.Warnf("re-merged %d failed entries back into buffer", len(failedKeys))
	}
}
