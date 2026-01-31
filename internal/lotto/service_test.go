package lotto

import (
	"strings"
	"testing"
	"time"

	"github.com/example/LottoSmash/internal/logger"
)

func TestService_parseCSV(t *testing.T) {
	// 로거 초기화 (테스트용 임시 디렉토리 사용)
	tmpDir := t.TempDir()
	l, err := logger.New(tmpDir, "debug")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer l.Close()

	svc := &Service{log: l}

	// 테스트용 CSV 데이터
	// 실제 CSV 구조: No,회차,당첨번호(6개),보너스,당첨게임수,1게임당 당첨금액 ("순위" 컬럼 제거됨)
	csvContent := `No,회차,당첨번호,,,,,,보너스,당첨게임수,1게임당 당첨금액
1,"1,208",6,27,30,36,38,42,25,6 명,"5,001,713,625 원"
2,"1,207",10,22,24,27,38,45,11,17 명,"1,733,202,949 원"
3,999,1,2,3,4,5,6,7,5명(dirty),"1000원!"
`

	r := strings.NewReader(csvContent)

	draws, err := svc.parseCSV(r)
	t.Logf("parseCSV result - draws: %d, error: %v", len(draws), err)

	if err != nil {
		t.Fatalf("parseCSV failed: %v", err)
	}

	if len(draws) != 3 {
		t.Fatalf("expected 3 draws, got %d", len(draws))
	}

	// 1208회 검증
	d1 := draws[0]
	t.Logf("draw 1 (1208): %+v", d1)

	if d1.DrawNo != 1208 {
		t.Errorf("draw 1: expected DrawNo 1208, got %d", d1.DrawNo)
	}
	if d1.FirstWinners != 6 {
		t.Errorf("draw 1: expected FirstWinners 6, got %d", d1.FirstWinners)
	}
	if d1.FirstPrize != 5001713625 {
		t.Errorf("draw 1: expected FirstPrize 5001713625, got %d", d1.FirstPrize)
	}

	// 날짜 계산 검증 (1208회)
	// 1회: 2002-12-07
	// 1208회: 2002-12-07 + 1207주
	expectedDateStr := time.Date(2002, 12, 7, 0, 0, 0, 0, time.UTC).AddDate(0, 0, (1208-1)*7).Format("2006.01.02")
	if d1.DrawDate != expectedDateStr {
		t.Errorf("draw 1: expected DrawDate %s, got %s", expectedDateStr, d1.DrawDate)
	}

	// 1207회 검증
	d2 := draws[1]
	t.Logf("draw 2 (1207): %+v", d2)

	if d2.DrawNo != 1207 {
		t.Errorf("draw 2: expected DrawNo 1207, got %d", d2.DrawNo)
	}
	if d2.FirstWinners != 17 {
		t.Errorf("draw 2: expected FirstWinners 17, got %d", d2.FirstWinners)
	}
	if d2.FirstPrize != 1733202949 {
		t.Errorf("draw 2: expected FirstPrize 1733202949, got %d", d2.FirstPrize)
	}

	// 999회 (Dirty Data) 검증
	d3 := draws[2]
	t.Logf("draw 3 (999 dirty): %+v", d3)

	if d3.DrawNo != 999 {
		t.Errorf("draw 3: expected DrawNo 999, got %d", d3.DrawNo)
	}
	if d3.FirstWinners != 5 {
		t.Errorf("draw 3: expected FirstWinners 5, got %d", d3.FirstWinners)
	}
	if d3.FirstPrize != 1000 {
		t.Errorf("draw 3: expected FirstPrize 1000, got %d", d3.FirstPrize)
	}
}
