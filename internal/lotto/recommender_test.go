package lotto

import (
	"math"
	"testing"
)

// 테스트용 AnalysisStat 더미 데이터 생성 (번호 1~45)
func makeTestStats() []AnalysisStat {
	stats := make([]AnalysisStat, 45)
	for i := 0; i < 45; i++ {
		num := i + 1
		stats[i] = AnalysisStat{
			Number:       num,
			TotalProb:    float64(num) / 1000.0,    // 0.001 ~ 0.045
			ReappearProb: float64(46-num) / 1000.0,  // 0.045 ~ 0.001 (역순)
			FirstProb:    float64(num%10+1) / 100.0,  // 주기적 패턴
			LastProb:     float64(num%7+1) / 100.0,
			BayesianPost: float64(num*2%45+1) / 1000.0,
			BonusProb:    float64(num) / 500.0,
		}
	}
	return stats
}

func TestGetMethodProbabilities(t *testing.T) {
	r := &Recommender{}
	stats := makeTestStats()

	tests := []struct {
		code     string
		checkNum int
		wantProb float64
	}{
		{"NUMBER_FREQUENCY", 10, 0.010},  // TotalProb: 10/1000
		{"NUMBER_FREQUENCY", 45, 0.045},
		{"REAPPEAR_PROB", 1, 0.045},       // ReappearProb: (46-1)/1000
		{"REAPPEAR_PROB", 45, 0.001},
		{"BAYESIAN", 10, 0.021},           // BayesianPost: (10*2%45+1)/1000 = 21/1000
		{"FIRST_POSITION", 10, 0.01},      // FirstProb: (10%10+1)/100 = 1/100
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			probMap := r.getMethodProbabilities(tt.code, stats)

			if len(probMap) != 45 {
				t.Errorf("expected 45 entries, got %d", len(probMap))
			}

			got := probMap[tt.checkNum]
			if math.Abs(got-tt.wantProb) > 0.0001 {
				t.Errorf("number %d: got prob %.6f, want %.6f", tt.checkNum, got, tt.wantProb)
			}
		})
	}
}

func TestCombineSimpleAverage(t *testing.T) {
	r := &Recommender{}

	t.Run("single method", func(t *testing.T) {
		probMaps := []map[int]float64{
			{1: 0.10, 2: 0.20, 3: 0.30},
		}
		result := r.combineSimpleAverage(probMaps)

		if math.Abs(result[1]-0.10) > 0.0001 {
			t.Errorf("number 1: got %.6f, want 0.10", result[1])
		}
		if math.Abs(result[2]-0.20) > 0.0001 {
			t.Errorf("number 2: got %.6f, want 0.20", result[2])
		}
	})

	t.Run("two methods average", func(t *testing.T) {
		probMaps := []map[int]float64{
			{1: 0.10, 2: 0.40, 3: 0.30},
			{1: 0.30, 2: 0.20, 3: 0.10},
		}
		result := r.combineSimpleAverage(probMaps)

		// (0.10 + 0.30) / 2 = 0.20
		if math.Abs(result[1]-0.20) > 0.0001 {
			t.Errorf("number 1: got %.6f, want 0.20", result[1])
		}
		// (0.40 + 0.20) / 2 = 0.30
		if math.Abs(result[2]-0.30) > 0.0001 {
			t.Errorf("number 2: got %.6f, want 0.30", result[2])
		}
		// (0.30 + 0.10) / 2 = 0.20
		if math.Abs(result[3]-0.20) > 0.0001 {
			t.Errorf("number 3: got %.6f, want 0.20", result[3])
		}
	})

	t.Run("three methods average", func(t *testing.T) {
		probMaps := []map[int]float64{
			{1: 0.30, 2: 0.10},
			{1: 0.60, 2: 0.20},
			{1: 0.00, 2: 0.30},
		}
		result := r.combineSimpleAverage(probMaps)

		// (0.30 + 0.60 + 0.00) / 3 = 0.30
		if math.Abs(result[1]-0.30) > 0.0001 {
			t.Errorf("number 1: got %.6f, want 0.30", result[1])
		}
		// (0.10 + 0.20 + 0.30) / 3 = 0.20
		if math.Abs(result[2]-0.20) > 0.0001 {
			t.Errorf("number 2: got %.6f, want 0.20", result[2])
		}
	})

	t.Run("empty input", func(t *testing.T) {
		result := r.combineSimpleAverage(nil)
		if len(result) != 0 {
			t.Errorf("expected empty map, got %d entries", len(result))
		}
	})
}

func TestCombineSimpleAverage_FullRange(t *testing.T) {
	r := &Recommender{}
	stats := makeTestStats()

	// NUMBER_FREQUENCY와 REAPPEAR_PROB은 역순이므로 평균은 모든 번호에서 비슷해야 함
	prob1 := r.getMethodProbabilities("NUMBER_FREQUENCY", stats)
	prob2 := r.getMethodProbabilities("REAPPEAR_PROB", stats)

	combined := r.combineSimpleAverage([]map[int]float64{prob1, prob2})

	// 모든 번호의 합이 (TotalProb합 + ReappearProb합) / 2와 같은지 확인
	totalSum := 0.0
	for num := 1; num <= 45; num++ {
		totalSum += combined[num]
	}

	expectedSum := 0.0
	for _, s := range stats {
		expectedSum += (s.TotalProb + s.ReappearProb) / 2
	}

	if math.Abs(totalSum-expectedSum) > 0.0001 {
		t.Errorf("total sum mismatch: got %.6f, want %.6f", totalSum, expectedSum)
	}

	// 번호 1: TotalProb=0.001, ReappearProb=0.045 -> 평균=0.023
	if math.Abs(combined[1]-0.023) > 0.0001 {
		t.Errorf("number 1: got %.6f, want 0.023", combined[1])
	}

	// 번호 45: TotalProb=0.045, ReappearProb=0.001 -> 평균=0.023
	if math.Abs(combined[45]-0.023) > 0.0001 {
		t.Errorf("number 45: got %.6f, want 0.023", combined[45])
	}
}

func TestCalculateCombineConfidence(t *testing.T) {
	r := &Recommender{}

	t.Run("zero methods", func(t *testing.T) {
		got := r.calculateCombineConfidence([]int{1, 2, 3}, map[int]float64{1: 0.5}, 0)
		if got != 0.0 {
			t.Errorf("expected 0.0, got %.6f", got)
		}
	})

	t.Run("empty numbers", func(t *testing.T) {
		got := r.calculateCombineConfidence([]int{}, map[int]float64{1: 0.5}, 1)
		if got != 0.0 {
			t.Errorf("expected 0.0, got %.6f", got)
		}
	})

	t.Run("high confidence", func(t *testing.T) {
		// 기대확률 1/45 ≈ 0.0222, 상한 = 0.0222 * 3 = 0.0667
		// 점수가 0.05이면 confidence = 0.05 / 0.0667 ≈ 0.75
		scores := map[int]float64{1: 0.05, 2: 0.05, 3: 0.05, 4: 0.05, 5: 0.05, 6: 0.05}
		got := r.calculateCombineConfidence([]int{1, 2, 3, 4, 5, 6}, scores, 2)
		if got < 0.5 || got > 1.0 {
			t.Errorf("expected confidence between 0.5~1.0, got %.6f", got)
		}
	})

	t.Run("capped at 1.0", func(t *testing.T) {
		scores := map[int]float64{1: 1.0, 2: 1.0}
		got := r.calculateCombineConfidence([]int{1, 2}, scores, 1)
		if got != 1.0 {
			t.Errorf("expected 1.0 (capped), got %.6f", got)
		}
	})
}

func TestSelectTopNumbers_WithCombinedScores(t *testing.T) {
	r := &Recommender{
		rng: nil, // selectTopNumbers에서 동점 시 rng 사용하지만, 점수가 다르면 불필요
	}

	// 명확히 다른 점수를 가진 맵
	scores := map[int]float64{
		1: 0.10, 2: 0.50, 3: 0.30, 4: 0.90, 5: 0.70,
		6: 0.20, 7: 0.80, 8: 0.60, 9: 0.40, 10: 0.15,
	}

	numbers := r.selectTopNumbers(scores, 6)

	if len(numbers) != 6 {
		t.Fatalf("expected 6 numbers, got %d", len(numbers))
	}

	// 상위 6개: 4(0.90), 7(0.80), 5(0.70), 8(0.60), 2(0.50), 9(0.40)
	expectedTop := map[int]bool{4: true, 7: true, 5: true, 8: true, 2: true, 9: true}
	for _, n := range numbers {
		if !expectedTop[n] {
			t.Errorf("unexpected number %d in top 6, numbers: %v", n, numbers)
		}
	}
}

func TestCombineWeightedAverage(t *testing.T) {
	r := &Recommender{}

	t.Run("basic weighted average", func(t *testing.T) {
		// method A: 번호1=0.10, 번호2=0.40
		// method B: 번호1=0.50, 번호2=0.20
		// weights: A=0.7, B=0.3
		// 번호1: (0.7*0.10 + 0.3*0.50) / (0.7+0.3) = (0.07+0.15)/1.0 = 0.22
		// 번호2: (0.7*0.40 + 0.3*0.20) / (0.7+0.3) = (0.28+0.06)/1.0 = 0.34
		probMaps := []map[int]float64{
			{1: 0.10, 2: 0.40},
			{1: 0.50, 2: 0.20},
		}
		codes := []string{"A", "B"}
		weights := map[string]float64{"A": 0.7, "B": 0.3}

		result := r.combineWeightedAverage(probMaps, codes, weights)

		if math.Abs(result[1]-0.22) > 0.0001 {
			t.Errorf("number 1: got %.6f, want 0.22", result[1])
		}
		if math.Abs(result[2]-0.34) > 0.0001 {
			t.Errorf("number 2: got %.6f, want 0.34", result[2])
		}
	})

	t.Run("equal weights equals simple average", func(t *testing.T) {
		probMaps := []map[int]float64{
			{1: 0.10, 2: 0.40, 3: 0.30},
			{1: 0.30, 2: 0.20, 3: 0.10},
		}
		codes := []string{"A", "B"}
		weights := map[string]float64{"A": 1.0, "B": 1.0}

		weighted := r.combineWeightedAverage(probMaps, codes, weights)
		simple := r.combineSimpleAverage(probMaps)

		for num := 1; num <= 3; num++ {
			if math.Abs(weighted[num]-simple[num]) > 0.0001 {
				t.Errorf("number %d: weighted=%.6f != simple=%.6f with equal weights", num, weighted[num], simple[num])
			}
		}
	})

	t.Run("unnormalized weights auto-normalize", func(t *testing.T) {
		// weights 합이 1이 아닌 경우에도 정규화
		// A=2.0, B=3.0 → 정규화 후 A=0.4, B=0.6
		// 번호1: (2.0*0.10 + 3.0*0.50) / 5.0 = (0.20+1.50)/5.0 = 0.34
		probMaps := []map[int]float64{
			{1: 0.10},
			{1: 0.50},
		}
		codes := []string{"A", "B"}
		weights := map[string]float64{"A": 2.0, "B": 3.0}

		result := r.combineWeightedAverage(probMaps, codes, weights)

		if math.Abs(result[1]-0.34) > 0.0001 {
			t.Errorf("number 1: got %.6f, want 0.34", result[1])
		}
	})

	t.Run("high weight dominates", func(t *testing.T) {
		// A=0.99, B=0.01 → A의 값에 거의 수렴
		probMaps := []map[int]float64{
			{1: 0.80},
			{1: 0.10},
		}
		codes := []string{"A", "B"}
		weights := map[string]float64{"A": 0.99, "B": 0.01}

		result := r.combineWeightedAverage(probMaps, codes, weights)

		// (0.99*0.80 + 0.01*0.10) / 1.0 = 0.793
		if math.Abs(result[1]-0.793) > 0.001 {
			t.Errorf("number 1: got %.6f, want ~0.793", result[1])
		}
	})

	t.Run("three methods", func(t *testing.T) {
		// A=0.5, B=0.3, C=0.2
		// 번호1: (0.5*0.60 + 0.3*0.20 + 0.2*0.10) / 1.0 = 0.30+0.06+0.02 = 0.38
		probMaps := []map[int]float64{
			{1: 0.60},
			{1: 0.20},
			{1: 0.10},
		}
		codes := []string{"A", "B", "C"}
		weights := map[string]float64{"A": 0.5, "B": 0.3, "C": 0.2}

		result := r.combineWeightedAverage(probMaps, codes, weights)

		if math.Abs(result[1]-0.38) > 0.0001 {
			t.Errorf("number 1: got %.6f, want 0.38", result[1])
		}
	})

	t.Run("empty input", func(t *testing.T) {
		result := r.combineWeightedAverage(nil, nil, nil)
		if len(result) != 0 {
			t.Errorf("expected empty map, got %d entries", len(result))
		}
	})

	t.Run("zero weights fallback to simple average", func(t *testing.T) {
		probMaps := []map[int]float64{
			{1: 0.10, 2: 0.40},
			{1: 0.30, 2: 0.20},
		}
		codes := []string{"A", "B"}
		weights := map[string]float64{"A": 0.0, "B": 0.0}

		weighted := r.combineWeightedAverage(probMaps, codes, weights)
		simple := r.combineSimpleAverage(probMaps)

		for num := 1; num <= 2; num++ {
			if math.Abs(weighted[num]-simple[num]) > 0.0001 {
				t.Errorf("number %d: zero-weight result should match simple avg", num)
			}
		}
	})

	t.Run("partial weights missing key treated as zero", func(t *testing.T) {
		// B에 가중치가 없으면 0으로 처리 → A만 반영
		// 번호1: (0.5*0.80 + 0.0*0.10) / 0.5 = 0.80
		probMaps := []map[int]float64{
			{1: 0.80},
			{1: 0.10},
		}
		codes := []string{"A", "B"}
		weights := map[string]float64{"A": 0.5}

		result := r.combineWeightedAverage(probMaps, codes, weights)

		if math.Abs(result[1]-0.80) > 0.0001 {
			t.Errorf("number 1: got %.6f, want 0.80 (only A reflected)", result[1])
		}
	})
}

func TestCombineWeightedAverage_FullRange(t *testing.T) {
	r := &Recommender{}
	stats := makeTestStats()

	prob1 := r.getMethodProbabilities("NUMBER_FREQUENCY", stats) // TotalProb: ascending
	prob2 := r.getMethodProbabilities("REAPPEAR_PROB", stats)    // ReappearProb: descending

	// 가중치 0.8:0.2 → 높은 번호가 유리 (TotalProb이 높으므로)
	codes := []string{"NUMBER_FREQUENCY", "REAPPEAR_PROB"}
	weights := map[string]float64{"NUMBER_FREQUENCY": 0.8, "REAPPEAR_PROB": 0.2}

	combined := r.combineWeightedAverage([]map[int]float64{prob1, prob2}, codes, weights)

	// 번호 45가 번호 1보다 높아야 함 (TotalProb 가중치가 더 크므로)
	if combined[45] <= combined[1] {
		t.Errorf("number 45 (%.6f) should be > number 1 (%.6f) with 0.8 weight on TotalProb", combined[45], combined[1])
	}
}

func TestGetCombineMethods(t *testing.T) {
	svc := &Service{}
	resp := svc.GetCombineMethods()

	if resp.TotalCount != 5 {
		t.Errorf("expected 5 combine methods, got %d", resp.TotalCount)
	}

	// 활성화 상태 확인 (SIMPLE_AVG, WEIGHTED_AVG)
	activeCount := 0
	for _, m := range resp.Methods {
		if m.IsActive {
			activeCount++
		}
	}
	if activeCount != 2 {
		t.Errorf("expected 2 active methods, got %d", activeCount)
	}

	// SIMPLE_AVG가 활성화 상태인지 확인
	if resp.Methods[0].Code != CombineSimpleAvg {
		t.Errorf("expected first method to be SIMPLE_AVG, got %s", resp.Methods[0].Code)
	}
	if !resp.Methods[0].IsActive {
		t.Errorf("SIMPLE_AVG should be active")
	}
}

func TestMaxMethodCodesValidation(t *testing.T) {
	if MaxMethodCodes != 3 {
		t.Errorf("MaxMethodCodes should be 3, got %d", MaxMethodCodes)
	}
}

func TestCombineMethodConstants(t *testing.T) {
	codes := []string{CombineSimpleAvg, CombineWeightedAvg, CombineBayesian, CombineGeometricMean, CombineMinMax}
	expected := []string{"SIMPLE_AVG", "WEIGHTED_AVG", "BAYESIAN_COMBINE", "GEOMETRIC_MEAN", "MIN_MAX"}

	for i, code := range codes {
		if code != expected[i] {
			t.Errorf("constant %d: got %s, want %s", i, code, expected[i])
		}
	}
}

func TestCombineSimpleAverage_OrderIndependence(t *testing.T) {
	r := &Recommender{}

	mapA := map[int]float64{1: 0.10, 2: 0.40, 3: 0.70}
	mapB := map[int]float64{1: 0.50, 2: 0.20, 3: 0.30}

	// A, B 순서
	result1 := r.combineSimpleAverage([]map[int]float64{mapA, mapB})
	// B, A 순서
	result2 := r.combineSimpleAverage([]map[int]float64{mapB, mapA})

	for num := 1; num <= 3; num++ {
		if math.Abs(result1[num]-result2[num]) > 0.0001 {
			t.Errorf("number %d: order matters! AB=%.6f, BA=%.6f", num, result1[num], result2[num])
		}
	}
}
