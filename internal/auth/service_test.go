package auth

import (
	"sync"
	"testing"
)

// --- mock recorder ---

type recordCall struct {
	UserID int64
	Amount int64
	TxType string
}

type mockRecorder struct {
	mu    sync.Mutex
	calls []recordCall
}

func (m *mockRecorder) Record(userID int64, amount int64, txType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, recordCall{userID, amount, txType})
}

func (m *mockRecorder) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.calls)
}

func (m *mockRecorder) getCalls() []recordCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	dst := make([]recordCall, len(m.calls))
	copy(dst, m.calls)
	return dst
}

// =============================================================
// Scenario 1: SetZamHistoryRecorder → recordZamHistory 호출 시 기록
// =============================================================
func TestRecordZamHistory_WithRecorder(t *testing.T) {
	mr := &mockRecorder{}
	svc := &Service{}
	svc.SetZamHistoryRecorder(mr)

	svc.recordZamHistory(1, 100, "REGISTER_BONUS")
	svc.recordZamHistory(2, 200, "DAILY_LOGIN")

	if mr.callCount() != 2 {
		t.Fatalf("expected 2 record calls, got %d", mr.callCount())
	}

	calls := mr.getCalls()
	if calls[0].UserID != 1 || calls[0].Amount != 100 || calls[0].TxType != "REGISTER_BONUS" {
		t.Fatalf("unexpected first call: %+v", calls[0])
	}
	if calls[1].UserID != 2 || calls[1].Amount != 200 || calls[1].TxType != "DAILY_LOGIN" {
		t.Fatalf("unexpected second call: %+v", calls[1])
	}
}

// =============================================================
// Scenario 2: recorder 미설정(nil) → recordZamHistory는 no-op (패닉 없음)
// =============================================================
func TestRecordZamHistory_NilRecorder(t *testing.T) {
	svc := &Service{} // zamHistoryRecorder == nil

	// 패닉 없이 정상 동작해야 함
	svc.recordZamHistory(1, 100, "REGISTER_BONUS")
	svc.recordZamHistory(2, 200, "DAILY_LOGIN")
}

// =============================================================
// Scenario 3: SetZamHistoryRecorder로 교체 가능
// =============================================================
func TestSetZamHistoryRecorder_Replace(t *testing.T) {
	mr1 := &mockRecorder{}
	mr2 := &mockRecorder{}
	svc := &Service{}

	svc.SetZamHistoryRecorder(mr1)
	svc.recordZamHistory(1, 100, "REGISTER_BONUS")

	svc.SetZamHistoryRecorder(mr2)
	svc.recordZamHistory(2, 200, "DAILY_LOGIN")

	if mr1.callCount() != 1 {
		t.Fatalf("expected 1 call to first recorder, got %d", mr1.callCount())
	}
	if mr2.callCount() != 1 {
		t.Fatalf("expected 1 call to second recorder, got %d", mr2.callCount())
	}
}

// =============================================================
// Scenario 4: 다양한 tx_type 기록 검증
// =============================================================
func TestRecordZamHistory_AllTxTypes(t *testing.T) {
	mr := &mockRecorder{}
	svc := &Service{}
	svc.SetZamHistoryRecorder(mr)

	types := []struct {
		txType string
		amount int64
	}{
		{"REGISTER_BONUS", 100},
		{"DAILY_LOGIN", 10},
		{"PURCHASE", -50},
		{"REFUND", 50},
		{"REWARD", 1000},
		{"ADMIN", 999},
	}

	for _, tt := range types {
		svc.recordZamHistory(1, tt.amount, tt.txType)
	}

	calls := mr.getCalls()
	if len(calls) != len(types) {
		t.Fatalf("expected %d calls, got %d", len(types), len(calls))
	}
	for i, tt := range types {
		if calls[i].TxType != tt.txType {
			t.Fatalf("call %d: expected type %s, got %s", i, tt.txType, calls[i].TxType)
		}
		if calls[i].Amount != tt.amount {
			t.Fatalf("call %d: expected amount %d, got %d", i, tt.amount, calls[i].Amount)
		}
	}
}

// =============================================================
// Scenario 5: 음수 금액(차감)도 정상 기록
// =============================================================
func TestRecordZamHistory_NegativeAmount(t *testing.T) {
	mr := &mockRecorder{}
	svc := &Service{}
	svc.SetZamHistoryRecorder(mr)

	svc.recordZamHistory(1, -500, "PURCHASE")

	calls := mr.getCalls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Amount != -500 {
		t.Fatalf("expected amount -500, got %d", calls[0].Amount)
	}
}
