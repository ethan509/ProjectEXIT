package zamhistory

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/example/LottoSmash/internal/logger"
)

// --- mock store ---

type upsertCall struct {
	UserID      int64
	HistoryDate time.Time
	TxType      string
	Amount      int64
	Count       int
}

type mockStore struct {
	mu    sync.Mutex
	calls []upsertCall
	err   error // if non-nil, UpsertDailyHistory returns this error
}

func (m *mockStore) UpsertDailyHistory(_ context.Context, userID int64, historyDate time.Time, txType string, amount int64, count int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, upsertCall{userID, historyDate, txType, amount, count})
	return m.err
}

func (m *mockStore) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.calls)
}

func (m *mockStore) getCalls() []upsertCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	dst := make([]upsertCall, len(m.calls))
	copy(dst, m.calls)
	return dst
}

// --- helpers ---

func newTestLogger(t *testing.T) *logger.Logger {
	t.Helper()
	lg, err := logger.New(t.TempDir(), "debug")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	t.Cleanup(func() { lg.Close() })
	return lg
}

func newTestBuffer(t *testing.T, s store) *Buffer {
	t.Helper()
	return &Buffer{
		data:     make(map[bufferKey]*bufferValue),
		repo:     s,
		log:      newTestLogger(t),
		loc:      time.UTC,
		interval: 50 * time.Millisecond, // fast for tests
	}
}

// =============================================================
// Scenario 1: 단일 Record → Flush → DB에 정확히 1건 UPSERT
// =============================================================
func TestRecord_SingleEntry(t *testing.T) {
	ms := &mockStore{}
	buf := newTestBuffer(t, ms)

	buf.Record(1, 100, "REGISTER_BONUS")

	// buffer에 1건 존재
	buf.mu.Lock()
	if len(buf.data) != 1 {
		t.Fatalf("expected 1 entry in buffer, got %d", len(buf.data))
	}
	buf.mu.Unlock()

	buf.Flush(context.Background())

	calls := ms.getCalls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 upsert call, got %d", len(calls))
	}
	c := calls[0]
	if c.UserID != 1 || c.Amount != 100 || c.Count != 1 || c.TxType != "REGISTER_BONUS" {
		t.Fatalf("unexpected call: %+v", c)
	}

	// flush 후 buffer는 비어야 함
	buf.mu.Lock()
	if len(buf.data) != 0 {
		t.Fatalf("expected empty buffer after flush, got %d", len(buf.data))
	}
	buf.mu.Unlock()
}

// =============================================================
// Scenario 2: 같은 유저+같은 날+같은 타입 → 합산 (aggregation)
// =============================================================
func TestRecord_SameKeyAggregation(t *testing.T) {
	ms := &mockStore{}
	buf := newTestBuffer(t, ms)

	buf.Record(1, 100, "DAILY_LOGIN")
	buf.Record(1, 100, "DAILY_LOGIN")
	buf.Record(1, 100, "DAILY_LOGIN")

	buf.mu.Lock()
	if len(buf.data) != 1 {
		t.Fatalf("expected 1 entry (aggregated), got %d", len(buf.data))
	}
	buf.mu.Unlock()

	buf.Flush(context.Background())

	calls := ms.getCalls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 upsert call, got %d", len(calls))
	}
	if calls[0].Amount != 300 {
		t.Fatalf("expected aggregated amount 300, got %d", calls[0].Amount)
	}
	if calls[0].Count != 3 {
		t.Fatalf("expected aggregated count 3, got %d", calls[0].Count)
	}
}

// =============================================================
// Scenario 3: 다른 유저/타입 → 별도 entry로 관리
// =============================================================
func TestRecord_DifferentKeys(t *testing.T) {
	ms := &mockStore{}
	buf := newTestBuffer(t, ms)

	buf.Record(1, 100, "REGISTER_BONUS")
	buf.Record(2, 200, "REGISTER_BONUS")
	buf.Record(1, 10, "DAILY_LOGIN")

	buf.mu.Lock()
	if len(buf.data) != 3 {
		t.Fatalf("expected 3 distinct entries, got %d", len(buf.data))
	}
	buf.mu.Unlock()

	buf.Flush(context.Background())

	if ms.callCount() != 3 {
		t.Fatalf("expected 3 upsert calls, got %d", ms.callCount())
	}
}

// =============================================================
// Scenario 4: 빈 buffer flush → DB 호출 없음
// =============================================================
func TestFlush_EmptyBuffer(t *testing.T) {
	ms := &mockStore{}
	buf := newTestBuffer(t, ms)

	buf.Flush(context.Background())

	if ms.callCount() != 0 {
		t.Fatalf("expected 0 upsert calls for empty buffer, got %d", ms.callCount())
	}
}

// =============================================================
// Scenario 5: Flush 실패 시 re-merge → 다음 flush에서 재시도
// =============================================================
func TestFlush_FailureRemerge(t *testing.T) {
	ms := &mockStore{err: errors.New("db connection error")}
	buf := newTestBuffer(t, ms)

	buf.Record(1, 500, "REWARD")

	// 1차 flush: 실패 → re-merge
	buf.Flush(context.Background())

	if ms.callCount() != 1 {
		t.Fatalf("expected 1 upsert attempt, got %d", ms.callCount())
	}

	// buffer에 re-merge되어 있어야 함
	buf.mu.Lock()
	if len(buf.data) != 1 {
		t.Fatalf("expected 1 re-merged entry, got %d", len(buf.data))
	}
	buf.mu.Unlock()

	// DB 복구 후 2차 flush: 성공
	ms.err = nil
	buf.Flush(context.Background())

	if ms.callCount() != 2 {
		t.Fatalf("expected 2 total upsert attempts, got %d", ms.callCount())
	}

	// 성공 후 buffer 비어야 함
	buf.mu.Lock()
	if len(buf.data) != 0 {
		t.Fatalf("expected empty buffer after successful retry, got %d", len(buf.data))
	}
	buf.mu.Unlock()
}

// =============================================================
// Scenario 6: Flush 실패 + 새 Record → re-merge가 새 데이터와 합산
// =============================================================
func TestFlush_FailureRemergeWithNewData(t *testing.T) {
	ms := &mockStore{err: errors.New("timeout")}
	buf := newTestBuffer(t, ms)

	buf.Record(1, 100, "DAILY_LOGIN")

	// flush 실패 → re-merge
	buf.Flush(context.Background())

	// 실패 후 추가 Record
	buf.Record(1, 100, "DAILY_LOGIN")

	// buffer에는 re-merge(100) + 새 Record(100) = 200이어야 함
	buf.mu.Lock()
	key := bufferKey{UserID: 1, HistoryDate: time.Now().In(time.UTC).Format("2006-01-02"), TxType: "DAILY_LOGIN"}
	val, ok := buf.data[key]
	if !ok {
		t.Fatal("expected entry in buffer after re-merge + new record")
	}
	if val.TotalAmount != 200 {
		t.Fatalf("expected merged amount 200, got %d", val.TotalAmount)
	}
	if val.TxCount != 2 {
		t.Fatalf("expected merged count 2, got %d", val.TxCount)
	}
	buf.mu.Unlock()

	// DB 복구 → 2차 flush 성공
	ms.err = nil
	buf.Flush(context.Background())

	calls := ms.getCalls()
	last := calls[len(calls)-1]
	if last.Amount != 200 || last.Count != 2 {
		t.Fatalf("expected retried flush with merged data (200, 2), got (%d, %d)", last.Amount, last.Count)
	}
}

// =============================================================
// Scenario 7: swap-and-flush — flush 중 Record는 새 buffer에 기록
// =============================================================
func TestFlush_SwapAndFlush(t *testing.T) {
	ms := &mockStore{}
	buf := newTestBuffer(t, ms)

	buf.Record(1, 100, "REGISTER_BONUS")
	buf.Flush(context.Background())

	// flush 후 새로운 Record
	buf.Record(1, 200, "REGISTER_BONUS")

	buf.mu.Lock()
	key := bufferKey{UserID: 1, HistoryDate: time.Now().In(time.UTC).Format("2006-01-02"), TxType: "REGISTER_BONUS"}
	val := buf.data[key]
	if val.TotalAmount != 200 {
		t.Fatalf("expected new buffer entry with 200, got %d", val.TotalAmount)
	}
	if val.TxCount != 1 {
		t.Fatalf("expected count 1 in new buffer, got %d", val.TxCount)
	}
	buf.mu.Unlock()

	// 2차 flush
	buf.Flush(context.Background())

	if ms.callCount() != 2 {
		t.Fatalf("expected 2 separate upsert calls, got %d", ms.callCount())
	}
}

// =============================================================
// Scenario 8: 동시 Record (concurrent goroutines) → race condition 없음
// =============================================================
func TestRecord_ConcurrentSafety(t *testing.T) {
	ms := &mockStore{}
	buf := newTestBuffer(t, ms)

	const goroutines = 100
	const recordsPerGoroutine = 50
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(userID int64) {
			defer wg.Done()
			for j := 0; j < recordsPerGoroutine; j++ {
				buf.Record(userID, 10, "DAILY_LOGIN")
			}
		}(int64(i % 10)) // 10명의 유저로 분산
	}
	wg.Wait()

	buf.Flush(context.Background())

	// 총 amount 검증: 100 goroutines × 50 records × 10 = 50,000
	totalAmount := int64(0)
	totalCount := 0
	for _, c := range ms.getCalls() {
		totalAmount += c.Amount
		totalCount += c.Count
	}
	expectedAmount := int64(goroutines * recordsPerGoroutine * 10)
	expectedCount := goroutines * recordsPerGoroutine
	if totalAmount != expectedAmount {
		t.Fatalf("expected total amount %d, got %d", expectedAmount, totalAmount)
	}
	if totalCount != expectedCount {
		t.Fatalf("expected total count %d, got %d", expectedCount, totalCount)
	}
}

// =============================================================
// Scenario 9: Start → 주기적 flush 동작 확인 → ctx cancel로 종료
// =============================================================
func TestStart_PeriodicFlushAndCancel(t *testing.T) {
	ms := &mockStore{}
	buf := newTestBuffer(t, ms) // interval = 50ms

	ctx, cancel := context.WithCancel(context.Background())

	buf.Record(1, 100, "REGISTER_BONUS")

	done := make(chan struct{})
	go func() {
		buf.Start(ctx)
		close(done)
	}()

	// 주기적 flush가 실행될 때까지 대기
	deadline := time.After(2 * time.Second)
	for {
		if ms.callCount() > 0 {
			break
		}
		select {
		case <-deadline:
			t.Fatal("periodic flush did not fire within timeout")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	// ctx cancel → goroutine 종료 확인
	cancel()
	select {
	case <-done:
		// ok
	case <-time.After(2 * time.Second):
		t.Fatal("Start goroutine did not exit after cancel")
	}
}

// =============================================================
// Scenario 10: 서버 종료 시나리오 — cancel 후 Flush로 잔여 데이터 반영
// =============================================================
func TestShutdown_FlushRemainingData(t *testing.T) {
	ms := &mockStore{}
	buf := newTestBuffer(t, ms)
	buf.interval = 10 * time.Minute // 주기적 flush가 안 되도록 길게 설정

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		buf.Start(ctx)
		close(done)
	}()

	// 데이터 기록
	buf.Record(1, 100, "REGISTER_BONUS")
	buf.Record(2, 200, "DAILY_LOGIN")
	buf.Record(1, 50, "DAILY_LOGIN")

	// 주기적 flush 안 됐으므로 DB 호출 없어야 함
	time.Sleep(100 * time.Millisecond)
	if ms.callCount() != 0 {
		t.Fatalf("expected no flush yet, but got %d calls", ms.callCount())
	}

	// 서버 종료 시뮬레이션
	cancel()
	<-done

	// 명시적 Flush (main.go의 shutdown 흐름과 동일)
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	buf.Flush(shutdownCtx)

	// 3개 entry가 모두 DB에 반영되어야 함
	if ms.callCount() != 3 {
		t.Fatalf("expected 3 upsert calls after shutdown flush, got %d", ms.callCount())
	}

	// buffer 비어야 함
	buf.mu.Lock()
	if len(buf.data) != 0 {
		t.Fatalf("expected empty buffer after shutdown flush, got %d", len(buf.data))
	}
	buf.mu.Unlock()
}

// =============================================================
// Scenario 11: 연속 Flush 2번 → 첫 flush 후 buffer 비어있으므로 두 번째는 no-op
// =============================================================
func TestFlush_DoubleFlushIdempotent(t *testing.T) {
	ms := &mockStore{}
	buf := newTestBuffer(t, ms)

	buf.Record(1, 100, "REGISTER_BONUS")
	buf.Flush(context.Background())
	buf.Flush(context.Background()) // no-op

	if ms.callCount() != 1 {
		t.Fatalf("expected 1 upsert call (second flush should be no-op), got %d", ms.callCount())
	}
}

// =============================================================
// Scenario 12: 타임존 반영 — Asia/Seoul 기준 날짜가 올바르게 기록되는지
// =============================================================
func TestRecord_TimezoneAwareness(t *testing.T) {
	ms := &mockStore{}
	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		t.Fatalf("failed to load timezone: %v", err)
	}

	buf := &Buffer{
		data:     make(map[bufferKey]*bufferValue),
		repo:     ms,
		log:      newTestLogger(t),
		loc:      loc,
		interval: time.Minute,
	}

	buf.Record(1, 100, "DAILY_LOGIN")

	expectedDate := time.Now().In(loc).Format("2006-01-02")

	buf.mu.Lock()
	found := false
	for k := range buf.data {
		if k.HistoryDate == expectedDate {
			found = true
		}
	}
	buf.mu.Unlock()

	if !found {
		t.Fatalf("expected history_date in KST (%s), but not found in buffer keys", expectedDate)
	}
}
