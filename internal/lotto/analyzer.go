package lotto

import (
	"context"
	"fmt"
	"sort"

	"github.com/example/LottoSmash/internal/constants"
	"github.com/example/LottoSmash/internal/logger"
)

type Analyzer struct {
	repo *Repository
	log  *logger.Logger
}

func NewAnalyzer(repo *Repository, log *logger.Logger) *Analyzer {
	return &Analyzer{repo: repo, log: log}
}

// CalculateNumberStats 각 번호별 당첨 횟수 계산
func (a *Analyzer) CalculateNumberStats(ctx context.Context) ([]NumberStat, error) {
	draws, err := a.repo.GetAllDraws(ctx)
	if err != nil {
		a.log.Errorf("CalculateNumberStats: failed to get all draws: %v", err)
		return nil, err
	}

	if len(draws) == 0 {
		return nil, nil
	}

	// 번호별 통계 초기화
	statsMap := make(map[int]*NumberStat)
	for i := constants.MinLottoNumber; i <= constants.MaxLottoNumber; i++ {
		statsMap[i] = &NumberStat{
			Number:     i,
			TotalCount: 0,
			BonusCount: 0,
			LastDrawNo: 0,
		}
	}

	// 모든 회차 순회하며 통계 집계
	for _, draw := range draws {
		// 일반 번호 집계
		for _, num := range draw.Numbers() {
			stat := statsMap[num]
			stat.TotalCount++
			if draw.DrawNo > stat.LastDrawNo {
				stat.LastDrawNo = draw.DrawNo
			}
		}

		// 보너스 번호 집계
		bonusStat := statsMap[draw.BonusNum]
		bonusStat.BonusCount++
		if draw.DrawNo > bonusStat.LastDrawNo {
			bonusStat.LastDrawNo = draw.DrawNo
		}
	}

	// 맵을 슬라이스로 변환
	stats := make([]NumberStat, 0, constants.TotalLottoNumber)
	for i := constants.MinLottoNumber; i <= constants.MaxLottoNumber; i++ {
		stats = append(stats, *statsMap[i])
	}

	return stats, nil
}

// CalculateReappearProbability 번호 재등장 확률 계산
// 각 번호가 어떤 회차에 나왔을 때, 다음 회차에도 나올 확률
func (a *Analyzer) CalculateReappearProbability(ctx context.Context) ([]ReappearStat, error) {
	draws, err := a.repo.GetAllDraws(ctx)
	if err != nil {
		a.log.Errorf("CalculateReappearProbability: failed to get all draws: %v", err)
		return nil, err
	}

	if len(draws) < 2 {
		return nil, nil
	}

	// 번호별 재등장 통계 초기화
	statsMap := make(map[int]*ReappearStat)
	for i := constants.MinLottoNumber; i <= constants.MaxLottoNumber; i++ {
		statsMap[i] = &ReappearStat{
			Number:        i,
			TotalAppear:   0,
			ReappearCount: 0,
			Probability:   0,
		}
	}

	// 연속된 회차 쌍을 순회
	for i := 0; i < len(draws)-1; i++ {
		currentDraw := draws[i]
		nextDraw := draws[i+1]

		// 현재 회차 번호들
		currentNumbers := currentDraw.Numbers()

		// 다음 회차 번호들을 set으로 변환
		nextNumbersSet := make(map[int]bool)
		for _, num := range nextDraw.Numbers() {
			nextNumbersSet[num] = true
		}

		// 현재 회차의 각 번호에 대해
		for _, num := range currentNumbers {
			stat := statsMap[num]
			stat.TotalAppear++

			// 다음 회차에도 나왔는지 확인
			if nextNumbersSet[num] {
				stat.ReappearCount++
			}
		}
	}

	// 확률 계산 및 슬라이스 변환
	stats := make([]ReappearStat, 0, constants.TotalLottoNumber)
	for i := constants.MinLottoNumber; i <= constants.MaxLottoNumber; i++ {
		stat := statsMap[i]
		if stat.TotalAppear > 0 {
			stat.Probability = float64(stat.ReappearCount) / float64(stat.TotalAppear)
		}
		stats = append(stats, *stat)
	}

	return stats, nil
}

// CalculateFirstLastStats 첫번째/마지막 번호 확률 계산
// 첫번째 번호: 정렬된 6개 번호 중 가장 작은 번호 (Num1)
// 마지막 번호: 정렬된 6개 번호 중 가장 큰 번호 (Num6)
func (a *Analyzer) CalculateFirstLastStats(ctx context.Context) (*FirstLastStatsResponse, error) {
	draws, err := a.repo.GetAllDraws(ctx)
	if err != nil {
		a.log.Errorf("CalculateFirstLastStats: failed to get all draws: %v", err)
		return nil, err
	}

	if len(draws) == 0 {
		return nil, nil
	}

	totalDraws := len(draws)

	// 첫번째 번호(Num1) 통계
	firstCountMap := make(map[int]int)
	// 마지막 번호(Num6) 통계
	lastCountMap := make(map[int]int)

	latestDrawNo := 0
	for _, draw := range draws {
		firstCountMap[draw.Num1]++
		lastCountMap[draw.Num6]++
		if draw.DrawNo > latestDrawNo {
			latestDrawNo = draw.DrawNo
		}
	}

	// 첫번째 번호 통계 변환 (1~45 중 가능한 범위: 1~40)
	firstStats := make([]PositionStat, 0)
	for num := constants.MinLottoNumber; num <= constants.MaxLottoNumber-5; num++ {
		count := firstCountMap[num]
		prob := 0.0
		if totalDraws > 0 {
			prob = float64(count) / float64(totalDraws)
		}
		firstStats = append(firstStats, PositionStat{
			Number:      num,
			Count:       count,
			Probability: prob,
		})
	}

	// 마지막 번호 통계 변환 (1~45 중 가능한 범위: 6~45)
	lastStats := make([]PositionStat, 0)
	for num := constants.MinLottoNumber + 5; num <= constants.MaxLottoNumber; num++ {
		count := lastCountMap[num]
		prob := 0.0
		if totalDraws > 0 {
			prob = float64(count) / float64(totalDraws)
		}
		lastStats = append(lastStats, PositionStat{
			Number:      num,
			Count:       count,
			Probability: prob,
		})
	}

	return &FirstLastStatsResponse{
		FirstStats:   firstStats,
		LastStats:    lastStats,
		TotalDraws:   totalDraws,
		LatestDrawNo: latestDrawNo,
	}, nil
}

// CalculatePairStats 번호 쌍 동반 출현 통계 계산
// 두 번호가 같은 회차에 함께 나온 횟수를 계산
func (a *Analyzer) CalculatePairStats(ctx context.Context, topN int) (*PairStatsResponse, error) {
	draws, err := a.repo.GetAllDraws(ctx)
	if err != nil {
		a.log.Errorf("CalculatePairStats: failed to get all draws: %v", err)
		return nil, err
	}

	if len(draws) == 0 {
		return nil, nil
	}

	totalDraws := len(draws)

	// 번호 쌍별 동반 출현 횟수 계산
	// key: "작은번호-큰번호" 형태로 저장
	pairCountMap := make(map[[2]int]int)

	latestDrawNo := 0
	for _, draw := range draws {
		nums := draw.Numbers()
		if draw.DrawNo > latestDrawNo {
			latestDrawNo = draw.DrawNo
		}

		// 6개 번호 중 2개씩 조합 (6C2 = 15개)
		for i := 0; i < len(nums); i++ {
			for j := i + 1; j < len(nums); j++ {
				// 항상 작은 번호가 먼저 오도록
				n1, n2 := nums[i], nums[j]
				if n1 > n2 {
					n1, n2 = n2, n1
				}
				pairCountMap[[2]int{n1, n2}]++
			}
		}
	}

	// PairStat 슬라이스로 변환
	allPairs := make([]PairStat, 0, len(pairCountMap))
	for pair, count := range pairCountMap {
		prob := float64(count) / float64(totalDraws)
		allPairs = append(allPairs, PairStat{
			Number1:     pair[0],
			Number2:     pair[1],
			Count:       count,
			Probability: prob,
		})
	}

	// 출현 횟수로 정렬
	sort.Slice(allPairs, func(i, j int) bool {
		return allPairs[i].Count > allPairs[j].Count
	})

	// 상위 N개, 하위 N개 추출
	if topN > len(allPairs) {
		topN = len(allPairs)
	}

	topPairs := allPairs[:topN]
	bottomPairs := make([]PairStat, topN)
	copy(bottomPairs, allPairs[len(allPairs)-topN:])

	// 하위는 오름차순으로 정렬 (가장 적게 나온 것부터)
	sort.Slice(bottomPairs, func(i, j int) bool {
		return bottomPairs[i].Count < bottomPairs[j].Count
	})

	return &PairStatsResponse{
		TopPairs:     topPairs,
		BottomPairs:  bottomPairs,
		TotalDraws:   totalDraws,
		LatestDrawNo: latestDrawNo,
	}, nil
}

// CalculateConsecutiveStats 연번 패턴 통계 계산
// 연속된 번호(예: 5-6-7)가 포함된 패턴 분석
func (a *Analyzer) CalculateConsecutiveStats(ctx context.Context) (*ConsecutiveStatsResponse, error) {
	draws, err := a.repo.GetAllDraws(ctx)
	if err != nil {
		a.log.Errorf("CalculateConsecutiveStats: failed to get all draws: %v", err)
		return nil, err
	}

	if len(draws) == 0 {
		return nil, nil
	}

	totalDraws := len(draws)

	// 연번 개수별 카운트 (0, 2, 3, 4, 5, 6)
	// 0: 연번 없음, 2: 2연번(예: 5-6), 3: 3연번(예: 5-6-7), ...
	countMap := make(map[int]int)

	// 최근 연번 포함 예시 (연번 2개 이상)
	var examples []ConsecutiveExample

	latestDrawNo := 0
	for _, draw := range draws {
		nums := draw.Numbers()
		if draw.DrawNo > latestDrawNo {
			latestDrawNo = draw.DrawNo
		}

		// 연번 개수 계산
		consecutiveCount := countConsecutive(nums)
		countMap[consecutiveCount]++

		// 연번이 있으면 예시에 추가 (최근 10개만)
		if consecutiveCount >= 2 && len(examples) < 10 {
			examples = append(examples, ConsecutiveExample{
				DrawNo:           draw.DrawNo,
				Numbers:          nums,
				ConsecutiveCount: consecutiveCount,
			})
		}
	}

	// 예시를 최신순으로 정렬
	sort.Slice(examples, func(i, j int) bool {
		return examples[i].DrawNo > examples[j].DrawNo
	})
	if len(examples) > 10 {
		examples = examples[:10]
	}

	// 통계 변환
	countStats := make([]ConsecutiveCountStat, 0)
	for _, cnt := range []int{0, 2, 3, 4, 5, 6} {
		drawCount := countMap[cnt]
		prob := 0.0
		if totalDraws > 0 {
			prob = float64(drawCount) / float64(totalDraws)
		}
		countStats = append(countStats, ConsecutiveCountStat{
			ConsecutiveCount: cnt,
			DrawCount:        drawCount,
			Probability:      prob,
		})
	}

	return &ConsecutiveStatsResponse{
		CountStats:     countStats,
		RecentExamples: examples,
		TotalDraws:     totalDraws,
		LatestDrawNo:   latestDrawNo,
	}, nil
}

// countConsecutive 6개 번호에서 최대 연속 번호 개수 계산
func countConsecutive(nums []int) int {
	if len(nums) < 2 {
		return 0
	}

	// 정렬된 상태라고 가정 (Num1~Num6은 이미 정렬됨)
	maxConsec := 1
	currentConsec := 1

	for i := 1; i < len(nums); i++ {
		if nums[i] == nums[i-1]+1 {
			currentConsec++
			if currentConsec > maxConsec {
				maxConsec = currentConsec
			}
		} else {
			currentConsec = 1
		}
	}

	// 연번이 없으면 0 반환, 있으면 최대 연속 개수 반환
	if maxConsec == 1 {
		return 0
	}
	return maxConsec
}

// CalculateRatioStats 홀짝/고저 비율 통계 계산
func (a *Analyzer) CalculateRatioStats(ctx context.Context) (*RatioStatsResponse, error) {
	draws, err := a.repo.GetAllDraws(ctx)
	if err != nil {
		a.log.Errorf("CalculateRatioStats: failed to get all draws: %v", err)
		return nil, err
	}

	if len(draws) == 0 {
		return nil, nil
	}

	totalDraws := len(draws)

	// 홀짝 비율별 카운트 (key: "홀:짝")
	oddEvenMap := make(map[string]int)
	// 고저 비율별 카운트 (key: "고:저", 고=23~45, 저=1~22)
	highLowMap := make(map[string]int)

	latestDrawNo := 0
	for _, draw := range draws {
		nums := draw.Numbers()
		if draw.DrawNo > latestDrawNo {
			latestDrawNo = draw.DrawNo
		}

		// 홀짝 계산
		oddCount := 0
		for _, n := range nums {
			if n%2 == 1 {
				oddCount++
			}
		}
		evenCount := 6 - oddCount
		oddEvenKey := fmt.Sprintf("%d:%d", oddCount, evenCount)
		oddEvenMap[oddEvenKey]++

		// 고저 계산 (고=23~45, 저=1~22)
		highCount := 0
		for _, n := range nums {
			if n >= 23 {
				highCount++
			}
		}
		lowCount := 6 - highCount
		highLowKey := fmt.Sprintf("%d:%d", highCount, lowCount)
		highLowMap[highLowKey]++
	}

	// 홀짝 통계 변환 (정렬: 6:0, 5:1, 4:2, 3:3, 2:4, 1:5, 0:6)
	oddEvenStats := make([]RatioStat, 0)
	for odd := 6; odd >= 0; odd-- {
		even := 6 - odd
		key := fmt.Sprintf("%d:%d", odd, even)
		count := oddEvenMap[key]
		prob := 0.0
		if totalDraws > 0 {
			prob = float64(count) / float64(totalDraws)
		}
		oddEvenStats = append(oddEvenStats, RatioStat{
			Ratio:       key,
			Count:       count,
			Probability: prob,
		})
	}

	// 고저 통계 변환 (정렬: 6:0, 5:1, 4:2, 3:3, 2:4, 1:5, 0:6)
	highLowStats := make([]RatioStat, 0)
	for high := 6; high >= 0; high-- {
		low := 6 - high
		key := fmt.Sprintf("%d:%d", high, low)
		count := highLowMap[key]
		prob := 0.0
		if totalDraws > 0 {
			prob = float64(count) / float64(totalDraws)
		}
		highLowStats = append(highLowStats, RatioStat{
			Ratio:       key,
			Count:       count,
			Probability: prob,
		})
	}

	return &RatioStatsResponse{
		OddEvenStats: oddEvenStats,
		HighLowStats: highLowStats,
		TotalDraws:   totalDraws,
		LatestDrawNo: latestDrawNo,
	}, nil
}

// getColorForNumber 번호에 해당하는 색상 반환
// 한국 로또 공식 색상: 1~10(Y노랑), 11~20(B파랑), 21~30(R빨강), 31~40(G회색), 41~45(E초록)
func getColorForNumber(num int) string {
	switch {
	case num >= 1 && num <= 10:
		return "Y" // Yellow 노랑
	case num >= 11 && num <= 20:
		return "B" // Blue 파랑
	case num >= 21 && num <= 30:
		return "R" // Red 빨강
	case num >= 31 && num <= 40:
		return "G" // Gray 회색
	case num >= 41 && num <= 45:
		return "E" // grEen 초록
	default:
		return "?"
	}
}

// CalculateColorStats 색상 패턴 통계 계산
func (a *Analyzer) CalculateColorStats(ctx context.Context, topN int) (*ColorStatsResponse, error) {
	draws, err := a.repo.GetAllDraws(ctx)
	if err != nil {
		a.log.Errorf("CalculateColorStats: failed to get all draws: %v", err)
		return nil, err
	}

	if len(draws) == 0 {
		return nil, nil
	}

	totalDraws := len(draws)

	// 색상 패턴별 카운트
	patternMap := make(map[string]int)
	// 각 색상별 총 출현 횟수
	colorCounts := map[string]int{"Y": 0, "B": 0, "R": 0, "G": 0, "E": 0}

	latestDrawNo := 0
	for _, draw := range draws {
		nums := draw.Numbers()
		if draw.DrawNo > latestDrawNo {
			latestDrawNo = draw.DrawNo
		}

		// 색상 패턴 생성
		var pattern string
		for _, num := range nums {
			color := getColorForNumber(num)
			pattern += color
			colorCounts[color]++
		}
		patternMap[pattern]++
	}

	// 패턴을 슬라이스로 변환하고 정렬
	patterns := make([]ColorPatternStat, 0, len(patternMap))
	for pattern, count := range patternMap {
		prob := float64(count) / float64(totalDraws)
		patterns = append(patterns, ColorPatternStat{
			Pattern:     pattern,
			Count:       count,
			Probability: prob,
		})
	}

	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Count > patterns[j].Count
	})

	if topN > len(patterns) {
		topN = len(patterns)
	}

	return &ColorStatsResponse{
		TopPatterns:  patterns[:topN],
		ColorCounts:  colorCounts,
		TotalDraws:   totalDraws,
		LatestDrawNo: latestDrawNo,
	}, nil
}

// getRowCol 번호의 7x7 격자 좌표 반환 (1-indexed)
func getRowCol(num int) (row, col int) {
	row = (num-1)/7 + 1
	col = (num-1)%7 + 1
	return
}

// CalculateRowColStats 행/열 분포 통계 계산 (7x7 격자 기준)
func (a *Analyzer) CalculateRowColStats(ctx context.Context, topN int) (*RowColStatsResponse, error) {
	draws, err := a.repo.GetAllDraws(ctx)
	if err != nil {
		a.log.Errorf("CalculateRowColStats: failed to get all draws: %v", err)
		return nil, err
	}

	if len(draws) == 0 {
		return nil, nil
	}

	totalDraws := len(draws)

	// 각 행/열별 총 출현 횟수
	rowCounts := make([]int, 8) // 1~7 사용 (0 무시)
	colCounts := make([]int, 8)

	// 행/열 분포 패턴별 카운트
	rowPatternMap := make(map[string]int)
	colPatternMap := make(map[string]int)

	latestDrawNo := 0
	for _, draw := range draws {
		nums := draw.Numbers()
		if draw.DrawNo > latestDrawNo {
			latestDrawNo = draw.DrawNo
		}

		// 이번 회차의 행/열 분포
		rowDist := make([]int, 8)
		colDist := make([]int, 8)

		for _, num := range nums {
			row, col := getRowCol(num)
			rowCounts[row]++
			colCounts[col]++
			rowDist[row]++
			colDist[col]++
		}

		// 분포 패턴 문자열 생성
		rowPattern := fmt.Sprintf("%d:%d:%d:%d:%d:%d:%d", rowDist[1], rowDist[2], rowDist[3], rowDist[4], rowDist[5], rowDist[6], rowDist[7])
		colPattern := fmt.Sprintf("%d:%d:%d:%d:%d:%d:%d", colDist[1], colDist[2], colDist[3], colDist[4], colDist[5], colDist[6], colDist[7])

		rowPatternMap[rowPattern]++
		colPatternMap[colPattern]++
	}

	// 행별 통계
	rowStats := make([]LineStat, 0, 7)
	for i := 1; i <= 7; i++ {
		prob := float64(rowCounts[i]) / float64(totalDraws*6) // 6개 번호 * 총 회차
		rowStats = append(rowStats, LineStat{
			Line:        i,
			Count:       rowCounts[i],
			Probability: prob,
		})
	}

	// 열별 통계
	colStats := make([]LineStat, 0, 7)
	for i := 1; i <= 7; i++ {
		prob := float64(colCounts[i]) / float64(totalDraws*6)
		colStats = append(colStats, LineStat{
			Line:        i,
			Count:       colCounts[i],
			Probability: prob,
		})
	}

	// 행 분포 패턴 정렬
	rowPatterns := make([]LineDistStat, 0, len(rowPatternMap))
	for pattern, count := range rowPatternMap {
		prob := float64(count) / float64(totalDraws)
		rowPatterns = append(rowPatterns, LineDistStat{
			Distribution: pattern,
			Count:        count,
			Probability:  prob,
		})
	}
	sort.Slice(rowPatterns, func(i, j int) bool {
		return rowPatterns[i].Count > rowPatterns[j].Count
	})

	// 열 분포 패턴 정렬
	colPatterns := make([]LineDistStat, 0, len(colPatternMap))
	for pattern, count := range colPatternMap {
		prob := float64(count) / float64(totalDraws)
		colPatterns = append(colPatterns, LineDistStat{
			Distribution: pattern,
			Count:        count,
			Probability:  prob,
		})
	}
	sort.Slice(colPatterns, func(i, j int) bool {
		return colPatterns[i].Count > colPatterns[j].Count
	})

	if topN > len(rowPatterns) {
		topN = len(rowPatterns)
	}
	topRowN := topN
	if topN > len(colPatterns) {
		topN = len(colPatterns)
	}
	topColN := topN

	return &RowColStatsResponse{
		RowStats:       rowStats,
		ColStats:       colStats,
		TopRowPatterns: rowPatterns[:topRowN],
		TopColPatterns: colPatterns[:topColN],
		TotalDraws:     totalDraws,
		LatestDrawNo:   latestDrawNo,
	}, nil
}

// RunFullAnalysis 전체 분석 실행 및 저장
func (a *Analyzer) RunFullAnalysis(ctx context.Context) error {
	a.log.Infof("RunFullAnalysis: starting full analysis")

	// 번호별 통계 계산 및 저장
	numberStats, err := a.CalculateNumberStats(ctx)
	if err != nil {
		a.log.Errorf("RunFullAnalysis: failed to calculate number stats: %v", err)
		return err
	}
	if len(numberStats) > 0 {
		if err := a.repo.UpsertNumberStats(ctx, numberStats); err != nil {
			a.log.Errorf("RunFullAnalysis: failed to upsert number stats: %v", err)
			return err
		}
	}

	// 재등장 확률 계산 및 저장
	reappearStats, err := a.CalculateReappearProbability(ctx)
	if err != nil {
		a.log.Errorf("RunFullAnalysis: failed to calculate reappear probability: %v", err)
		return err
	}
	if len(reappearStats) > 0 {
		if err := a.repo.UpsertReappearStats(ctx, reappearStats); err != nil {
			a.log.Errorf("RunFullAnalysis: failed to upsert reappear stats: %v", err)
			return err
		}
	}

	// 베이지안 통계 계산 및 저장 (점진적 업데이트) - 기존 테이블용
	if err := a.CalculateIncrementalBayesianStats(ctx); err != nil {
		a.log.Errorf("RunFullAnalysis: failed to calculate bayesian stats: %v", err)
		return err
	}

	// 통합 분석 통계 계산 및 저장 (점진적 업데이트)
	if err := a.CalculateUnifiedStats(ctx); err != nil {
		a.log.Errorf("RunFullAnalysis: failed to calculate unified stats: %v", err)
		return err
	}

	a.log.Infof("RunFullAnalysis: completed successfully")
	return nil
}

// CalculateBayesianStats 베이지안 추론 기반 번호별 확률 계산
// Prior: P(θ) = 1/45 (균등 분포)
// Likelihood: 최근 windowSize 회차에서의 출현 빈도
// Posterior: P(θ|D) ∝ P(D|θ)P(θ) (Beta-Binomial 모델)
func (a *Analyzer) CalculateBayesianStats(ctx context.Context, windowSize int) (*BayesianStatsResponse, error) {
	draws, err := a.repo.GetAllDraws(ctx)
	if err != nil {
		a.log.Errorf("CalculateBayesianStats: failed to get all draws: %v", err)
		return nil, err
	}

	if len(draws) == 0 {
		return nil, nil
	}

	totalDraws := len(draws)
	if windowSize <= 0 || windowSize > totalDraws {
		windowSize = 50 // 기본값: 최근 50회차
	}

	// 회차 번호 기준 내림차순 정렬 (최신 회차가 먼저)
	sort.Slice(draws, func(i, j int) bool {
		return draws[i].DrawNo > draws[j].DrawNo
	})

	latestDrawNo := draws[0].DrawNo

	// 최근 windowSize 회차만 사용
	recentDraws := draws
	if len(draws) > windowSize {
		recentDraws = draws[:windowSize]
	}
	actualWindow := len(recentDraws)

	// 상수
	const totalNumbers = 45
	const numbersPerDraw = 6
	prior := 1.0 / float64(totalNumbers) // P(θ) = 1/45

	// 각 회차에서 6개 번호가 나오므로, 전체 시행 횟수
	totalTrials := actualWindow * numbersPerDraw
	// 각 번호의 기대 출현 횟수 (균등 분포 가정)
	expectedCount := float64(totalTrials) / float64(totalNumbers)

	// 번호별 출현 횟수 카운트
	recentCountMap := make(map[int]int)
	lastAppearMap := make(map[int]int)
	for i := 1; i <= totalNumbers; i++ {
		recentCountMap[i] = 0
		lastAppearMap[i] = 0
	}

	for _, draw := range recentDraws {
		for _, num := range draw.Numbers() {
			recentCountMap[num]++
			if lastAppearMap[num] == 0 || draw.DrawNo > lastAppearMap[num] {
				lastAppearMap[num] = draw.DrawNo
			}
		}
	}

	// 전체 데이터에서 마지막 출현 회차 계산 (최근 윈도우 밖도 포함)
	for _, draw := range draws {
		for _, num := range draw.Numbers() {
			if draw.DrawNo > lastAppearMap[num] {
				lastAppearMap[num] = draw.DrawNo
			}
		}
	}

	// Beta-Binomial 모델로 Posterior 계산
	// Prior: Beta(α, β) where α = β = 1 (uniform)
	// Posterior: Beta(α + k, β + n - k)
	// Posterior mean = (α + k) / (α + β + n)
	alpha := 1.0
	beta := 1.0
	n := float64(totalTrials)

	stats := make([]BayesianNumberStat, 0, totalNumbers)
	var sumPosterior float64

	// 먼저 모든 번호의 posterior를 계산하고 합계를 구함
	posteriors := make([]float64, totalNumbers+1)
	for num := 1; num <= totalNumbers; num++ {
		k := float64(recentCountMap[num])
		// Posterior mean (Beta distribution)
		posteriors[num] = (alpha + k) / (alpha + beta + n)
		sumPosterior += posteriors[num]
	}

	// 정규화 및 통계 생성
	for num := 1; num <= totalNumbers; num++ {
		k := float64(recentCountMap[num])
		recentCount := recentCountMap[num]

		// Likelihood: 관측된 빈도 (최근 윈도우에서)
		likelihood := k / n

		// Normalized posterior (합이 1이 되도록)
		posterior := posteriors[num] / sumPosterior

		// 편차 계산
		deviation := k - expectedCount

		// Hot/Cold 상태 결정
		// 기대값 대비 +20% 이상이면 HOT, -20% 이하면 COLD
		var status string
		deviationRatio := deviation / expectedCount
		if deviationRatio > 0.2 {
			status = "HOT"
		} else if deviationRatio < -0.2 {
			status = "COLD"
		} else {
			status = "NEUTRAL"
		}

		// 마지막 출현 후 경과 회차
		gapSinceLast := latestDrawNo - lastAppearMap[num]

		stats = append(stats, BayesianNumberStat{
			Number:           num,
			Prior:            prior,
			Likelihood:       likelihood,
			Posterior:        posterior,
			RecentCount:      recentCount,
			ExpectedCount:    expectedCount,
			Deviation:        deviation,
			Status:           status,
			LastAppearDrawNo: lastAppearMap[num],
			GapSinceLastDraw: gapSinceLast,
		})
	}

	// Posterior 기준 내림차순 정렬
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Posterior > stats[j].Posterior
	})

	// Hot/Cold 번호 추출 (상위/하위 10개)
	topN := 10
	if topN > len(stats) {
		topN = len(stats)
	}

	hotNumbers := make([]BayesianNumberStat, 0, topN)
	coldNumbers := make([]BayesianNumberStat, 0, topN)

	for i := 0; i < topN; i++ {
		hotNumbers = append(hotNumbers, stats[i])
	}

	for i := len(stats) - 1; i >= len(stats)-topN && i >= 0; i-- {
		coldNumbers = append(coldNumbers, stats[i])
	}

	return &BayesianStatsResponse{
		Numbers:      stats,
		HotNumbers:   hotNumbers,
		ColdNumbers:  coldNumbers,
		WindowSize:   actualWindow,
		TotalDraws:   totalDraws,
		LatestDrawNo: latestDrawNo,
	}, nil
}

// CalculateIncrementalBayesianStats 점진적 베이지안 통계 계산
// - 이전 회차 통계가 있으면: 이전 posterior를 prior로 사용하여 업데이트
// - 이전 회차 통계가 없으면: 1회차부터 전체 계산 (초기화)
func (a *Analyzer) CalculateIncrementalBayesianStats(ctx context.Context) error {
	a.log.Infof("CalculateIncrementalBayesianStats: starting")

	// 현재 DB에 저장된 가장 최근 베이지안 통계 회차 확인
	latestBayesianDrawNo, err := a.repo.GetLatestBayesianDrawNo(ctx)
	if err != nil {
		a.log.Errorf("CalculateIncrementalBayesianStats: failed to get latest bayesian draw no: %v", err)
		return err
	}

	// 현재 DB에 저장된 가장 최근 당첨번호 회차 확인
	latestDrawNo, err := a.repo.GetLatestDrawNo(ctx)
	if err != nil {
		a.log.Errorf("CalculateIncrementalBayesianStats: failed to get latest draw no: %v", err)
		return err
	}

	if latestDrawNo == 0 {
		a.log.Infof("CalculateIncrementalBayesianStats: no draws found, skipping")
		return nil
	}

	// 베이지안 통계가 이미 최신이면 스킵
	if latestBayesianDrawNo >= latestDrawNo {
		a.log.Infof("CalculateIncrementalBayesianStats: already up to date (draw %d)", latestBayesianDrawNo)
		return nil
	}

	// 베이지안 통계가 없으면 전체 계산
	if latestBayesianDrawNo == 0 {
		a.log.Infof("CalculateIncrementalBayesianStats: no bayesian stats found, running full calculation")
		return a.CalculateFullBayesianStats(ctx)
	}

	// 점진적 업데이트: 마지막 계산 회차 이후부터 최신 회차까지 계산
	a.log.Infof("CalculateIncrementalBayesianStats: updating from draw %d to %d", latestBayesianDrawNo+1, latestDrawNo)

	// 이전 회차 통계 조회
	prevStats, err := a.repo.GetBayesianStatsByDrawNo(ctx, latestBayesianDrawNo)
	if err != nil {
		a.log.Errorf("CalculateIncrementalBayesianStats: failed to get prev stats: %v", err)
		return err
	}

	// 이전 통계를 맵으로 변환 (번호 -> 통계)
	prevStatsMap := make(map[int]BayesianStat)
	for _, stat := range prevStats {
		prevStatsMap[stat.Number] = stat
	}

	// 각 회차별로 업데이트
	for drawNo := latestBayesianDrawNo + 1; drawNo <= latestDrawNo; drawNo++ {
		draw, err := a.repo.GetDrawByNo(ctx, drawNo)
		if err != nil {
			a.log.Errorf("CalculateIncrementalBayesianStats: failed to get draw %d: %v", drawNo, err)
			return err
		}

		// 이번 회차 당첨번호
		newNumbers := draw.Numbers()
		newNumbersSet := make(map[int]bool)
		for _, num := range newNumbers {
			newNumbersSet[num] = true
		}

		// 새로운 통계 계산
		newStats := make([]BayesianStat, 0, 45)
		const totalNumbers = 45
		const alpha = 1.0
		const beta = 1.0

		for num := 1; num <= totalNumbers; num++ {
			prev := prevStatsMap[num]
			appeared := newNumbersSet[num]

			newCount := prev.TotalCount
			if appeared {
				newCount++
			}

			// Beta-Binomial 업데이트
			// posterior = (alpha + count) / (alpha + beta + totalTrials)
			totalTrials := float64(drawNo * 6) // 전체 시행 횟수
			posterior := (alpha + float64(newCount)) / (alpha + beta + totalTrials)

			newStat := BayesianStat{
				DrawNo:     drawNo,
				Number:     num,
				TotalCount: newCount,
				TotalDraws: drawNo,
				Prior:      prev.Posterior, // 이전 posterior가 새 prior
				Posterior:  posterior,
				Appeared:   appeared,
			}
			newStats = append(newStats, newStat)

			// 다음 회차를 위해 맵 업데이트
			prevStatsMap[num] = newStat
		}

		// DB에 저장
		if err := a.repo.UpsertBayesianStats(ctx, newStats); err != nil {
			a.log.Errorf("CalculateIncrementalBayesianStats: failed to upsert stats for draw %d: %v", drawNo, err)
			return err
		}
	}

	a.log.Infof("CalculateIncrementalBayesianStats: completed successfully")
	return nil
}

// CalculateFullBayesianStats 전체 베이지안 통계 계산 (초기화용)
// 1회차부터 현재까지 모든 회차에 대해 계산
func (a *Analyzer) CalculateFullBayesianStats(ctx context.Context) error {
	a.log.Infof("CalculateFullBayesianStats: starting full calculation")

	draws, err := a.repo.GetAllDraws(ctx)
	if err != nil {
		a.log.Errorf("CalculateFullBayesianStats: failed to get all draws: %v", err)
		return err
	}

	if len(draws) == 0 {
		a.log.Infof("CalculateFullBayesianStats: no draws found")
		return nil
	}

	// 회차순 정렬 (오름차순)
	sort.Slice(draws, func(i, j int) bool {
		return draws[i].DrawNo < draws[j].DrawNo
	})

	const totalNumbers = 45
	const alpha = 1.0
	const beta = 1.0
	prior := 1.0 / float64(totalNumbers) // 초기 사전확률

	// 누적 카운트 초기화
	countMap := make(map[int]int)
	for num := 1; num <= totalNumbers; num++ {
		countMap[num] = 0
	}

	// 각 회차별로 통계 계산 및 저장
	for _, draw := range draws {
		// 이번 회차 당첨번호
		newNumbers := draw.Numbers()
		newNumbersSet := make(map[int]bool)
		for _, num := range newNumbers {
			newNumbersSet[num] = true
		}

		// 카운트 업데이트
		for _, num := range newNumbers {
			countMap[num]++
		}

		// 통계 계산
		newStats := make([]BayesianStat, 0, totalNumbers)
		totalTrials := float64(draw.DrawNo * 6)

		for num := 1; num <= totalNumbers; num++ {
			appeared := newNumbersSet[num]
			count := countMap[num]

			// Beta-Binomial posterior
			posterior := (alpha + float64(count)) / (alpha + beta + totalTrials)

			newStats = append(newStats, BayesianStat{
				DrawNo:     draw.DrawNo,
				Number:     num,
				TotalCount: count,
				TotalDraws: draw.DrawNo,
				Prior:      prior, // 첫 회차는 균등 prior
				Posterior:  posterior,
				Appeared:   appeared,
			})

			// 다음 회차를 위해 prior 업데이트
			prior = posterior
		}

		// DB에 저장
		if err := a.repo.UpsertBayesianStats(ctx, newStats); err != nil {
			a.log.Errorf("CalculateFullBayesianStats: failed to upsert stats for draw %d: %v", draw.DrawNo, err)
			return err
		}

		// 다음 회차를 위해 prior 리셋 (각 번호별로 다르므로)
		prior = 1.0 / float64(totalNumbers)
	}

	a.log.Infof("CalculateFullBayesianStats: completed successfully (%d draws)", len(draws))
	return nil
}

// CalculateUnifiedStats 통합 분석 통계 계산 (점진적 업데이트)
// - 이전 회차 통계가 있으면: 점진적 업데이트
// - 이전 회차 통계가 없으면: 전체 계산
func (a *Analyzer) CalculateUnifiedStats(ctx context.Context) error {
	a.log.Infof("CalculateUnifiedStats: starting")

	// 현재 DB에 저장된 가장 최근 통합 분석 회차 확인
	latestAnalysisDrawNo, err := a.repo.GetLatestAnalysisDrawNo(ctx)
	if err != nil {
		a.log.Errorf("CalculateUnifiedStats: failed to get latest analysis draw no: %v", err)
		return err
	}

	// 현재 DB에 저장된 가장 최근 당첨번호 회차 확인
	latestDrawNo, err := a.repo.GetLatestDrawNo(ctx)
	if err != nil {
		a.log.Errorf("CalculateUnifiedStats: failed to get latest draw no: %v", err)
		return err
	}

	if latestDrawNo == 0 {
		a.log.Infof("CalculateUnifiedStats: no draws found, skipping")
		return nil
	}

	// 통합 분석이 이미 최신이면 스킵
	if latestAnalysisDrawNo >= latestDrawNo {
		a.log.Infof("CalculateUnifiedStats: already up to date (draw %d)", latestAnalysisDrawNo)
		return nil
	}

	// 통합 분석이 없으면 전체 계산
	if latestAnalysisDrawNo == 0 {
		a.log.Infof("CalculateUnifiedStats: no analysis stats found, running full calculation")
		return a.CalculateFullUnifiedStats(ctx)
	}

	// 점진적 업데이트: 마지막 계산 회차 이후부터 최신 회차까지
	a.log.Infof("CalculateUnifiedStats: updating from draw %d to %d", latestAnalysisDrawNo+1, latestDrawNo)

	// 이전 회차 통계 조회
	prevStats, err := a.repo.GetAnalysisStatsByDrawNo(ctx, latestAnalysisDrawNo)
	if err != nil {
		a.log.Errorf("CalculateUnifiedStats: failed to get prev stats: %v", err)
		return err
	}

	// 이전 통계를 맵으로 변환 (번호 -> 통계)
	prevStatsMap := make(map[int]AnalysisStat)
	for _, stat := range prevStats {
		prevStatsMap[stat.Number] = stat
	}

	// 모든 회차 데이터 조회 (재등장 확률 계산을 위해)
	draws, err := a.repo.GetAllDraws(ctx)
	if err != nil {
		a.log.Errorf("CalculateUnifiedStats: failed to get all draws: %v", err)
		return err
	}

	// 회차별로 맵 생성
	drawsMap := make(map[int]*LottoDraw)
	for _, draw := range draws {
		drawsMap[draw.DrawNo] = draw
	}

	// 각 회차별로 업데이트
	const totalNumbers = 45
	const alpha, beta = 1.0, 1.0

	for drawNo := latestAnalysisDrawNo + 1; drawNo <= latestDrawNo; drawNo++ {
		draw := drawsMap[drawNo]
		if draw == nil {
			continue
		}

		prevDraw := drawsMap[drawNo-1]

		// 이번 회차 당첨번호
		newNumbers := draw.Numbers()
		newNumbersSet := make(map[int]bool)
		for _, num := range newNumbers {
			newNumbersSet[num] = true
		}

		// 이전 회차 번호 (재등장 확률용)
		var prevNumbersSet map[int]bool
		if prevDraw != nil {
			prevNumbersSet = make(map[int]bool)
			for _, num := range prevDraw.Numbers() {
				prevNumbersSet[num] = true
			}
		}

		// 새로운 통계 계산
		newStats := make([]AnalysisStat, 0, totalNumbers)

		for num := 1; num <= totalNumbers; num++ {
			prev := prevStatsMap[num]
			appeared := newNumbersSet[num]

			// 기본 통계 업데이트
			newCount := prev.TotalCount
			newBonusCount := prev.BonusCount
			if appeared {
				newCount++
			}
			if draw.BonusNum == num {
				newBonusCount++
			}

			// 첫번째/마지막 위치 통계 업데이트
			newFirstCount := prev.FirstCount
			newLastCount := prev.LastCount
			if draw.Num1 == num {
				newFirstCount++
			}
			if draw.Num6 == num {
				newLastCount++
			}

			// 재등장 통계 업데이트
			newReappearTotal := prev.ReappearTotal
			newReappearCount := prev.ReappearCount
			if prevNumbersSet != nil && prevNumbersSet[num] {
				newReappearTotal++
				if appeared {
					newReappearCount++
				}
			}
			var reappearProb float64
			if newReappearTotal > 0 {
				reappearProb = float64(newReappearCount) / float64(newReappearTotal)
			}

			// 베이지안 업데이트
			totalTrials := float64(drawNo * 6)
			posterior := (alpha + float64(newCount)) / (alpha + beta + totalTrials)

			// 출현 확률 계산: total_count / (draw_no * 6)
			var totalProb float64
			if totalTrials > 0 {
				totalProb = float64(newCount) / totalTrials
			}

			// 보너스 출현 확률 계산: bonus_count / draw_no
			var bonusProb float64
			if drawNo > 0 {
				bonusProb = float64(newBonusCount) / float64(drawNo)
			}

			newStat := AnalysisStat{
				DrawNo:        drawNo,
				Number:        num,
				TotalCount:    newCount,
				TotalProb:     totalProb,
				BonusCount:    newBonusCount,
				BonusProb:     bonusProb,
				FirstCount:    newFirstCount,
				LastCount:     newLastCount,
				ReappearTotal: newReappearTotal,
				ReappearCount: newReappearCount,
				ReappearProb:  reappearProb,
				BayesianPrior: prev.BayesianPost,
				BayesianPost:  posterior,
				Appeared:      appeared,
			}
			newStats = append(newStats, newStat)

			// 다음 회차를 위해 맵 업데이트
			prevStatsMap[num] = newStat
		}

		// DB에 저장
		if err := a.repo.UpsertAnalysisStats(ctx, newStats); err != nil {
			a.log.Errorf("CalculateUnifiedStats: failed to upsert stats for draw %d: %v", drawNo, err)
			return err
		}
	}

	a.log.Infof("CalculateUnifiedStats: completed successfully")
	return nil
}

// CalculateFullUnifiedStats 전체 통합 분석 통계 계산 (초기화용)
func (a *Analyzer) CalculateFullUnifiedStats(ctx context.Context) error {
	a.log.Infof("CalculateFullUnifiedStats: starting full calculation")

	draws, err := a.repo.GetAllDraws(ctx)
	if err != nil {
		a.log.Errorf("CalculateFullUnifiedStats: failed to get all draws: %v", err)
		return err
	}

	if len(draws) == 0 {
		a.log.Infof("CalculateFullUnifiedStats: no draws found")
		return nil
	}

	// 회차순 정렬 (오름차순)
	sort.Slice(draws, func(i, j int) bool {
		return draws[i].DrawNo < draws[j].DrawNo
	})

	const totalNumbers = 45
	const alpha, beta = 1.0, 1.0

	// 누적 통계 초기화
	countMap := make(map[int]int)
	bonusCountMap := make(map[int]int)
	firstCountMap := make(map[int]int)
	lastCountMap := make(map[int]int)
	reappearTotalMap := make(map[int]int)
	reappearCountMap := make(map[int]int)
	for num := 1; num <= totalNumbers; num++ {
		countMap[num] = 0
		bonusCountMap[num] = 0
		firstCountMap[num] = 0
		lastCountMap[num] = 0
		reappearTotalMap[num] = 0
		reappearCountMap[num] = 0
	}

	var prevNumbersSet map[int]bool

	// 각 회차별로 통계 계산 및 저장
	for _, draw := range draws {
		// 이번 회차 당첨번호
		newNumbers := draw.Numbers()
		newNumbersSet := make(map[int]bool)
		for _, num := range newNumbers {
			newNumbersSet[num] = true
		}

		// 카운트 업데이트
		for _, num := range newNumbers {
			countMap[num]++
		}
		bonusCountMap[draw.BonusNum]++
		firstCountMap[draw.Num1]++
		lastCountMap[draw.Num6]++

		// 재등장 업데이트 (2회차부터)
		if prevNumbersSet != nil {
			for num := 1; num <= totalNumbers; num++ {
				if prevNumbersSet[num] {
					reappearTotalMap[num]++
					if newNumbersSet[num] {
						reappearCountMap[num]++
					}
				}
			}
		}

		// 통계 계산
		newStats := make([]AnalysisStat, 0, totalNumbers)
		totalTrials := float64(draw.DrawNo * 6)

		for num := 1; num <= totalNumbers; num++ {
			appeared := newNumbersSet[num]
			count := countMap[num]

			var reappearProb float64
			if reappearTotalMap[num] > 0 {
				reappearProb = float64(reappearCountMap[num]) / float64(reappearTotalMap[num])
			}

			posterior := (alpha + float64(count)) / (alpha + beta + totalTrials)

			// 출현 확률 계산: total_count / (draw_no * 6)
			var totalProb float64
			if totalTrials > 0 {
				totalProb = float64(count) / totalTrials
			}

			// 보너스 출현 확률 계산: bonus_count / draw_no
			var bonusProb float64
			if draw.DrawNo > 0 {
				bonusProb = float64(bonusCountMap[num]) / float64(draw.DrawNo)
			}

			newStats = append(newStats, AnalysisStat{
				DrawNo:        draw.DrawNo,
				Number:        num,
				TotalCount:    count,
				TotalProb:     totalProb,
				BonusCount:    bonusCountMap[num],
				BonusProb:     bonusProb,
				FirstCount:    firstCountMap[num],
				LastCount:     lastCountMap[num],
				ReappearTotal: reappearTotalMap[num],
				ReappearCount: reappearCountMap[num],
				ReappearProb:  reappearProb,
				BayesianPrior: 1.0 / float64(totalNumbers),
				BayesianPost:  posterior,
				Appeared:      appeared,
			})
		}

		// DB에 저장
		if err := a.repo.UpsertAnalysisStats(ctx, newStats); err != nil {
			a.log.Errorf("CalculateFullUnifiedStats: failed to upsert stats for draw %d: %v", draw.DrawNo, err)
			return err
		}

		// 다음 회차를 위해 현재 번호 저장
		prevNumbersSet = newNumbersSet
	}

	a.log.Infof("CalculateFullUnifiedStats: completed successfully (%d draws)", len(draws))
	return nil
}

// FixZeroProbabilityStats total_prob이 0인 행을 찾아서 수정
// 이미 total_count 값이 있는 행의 확률을 재계산하여 업데이트
func (a *Analyzer) FixZeroProbabilityStats(ctx context.Context) (int, error) {
	a.log.Infof("FixZeroProbabilityStats: starting")

	// total_prob이 0인 행 조회
	zeroStats, err := a.repo.GetAnalysisStatsWithZeroProb(ctx)
	if err != nil {
		a.log.Errorf("FixZeroProbabilityStats: failed to get zero prob stats: %v", err)
		return 0, err
	}

	if len(zeroStats) == 0 {
		a.log.Infof("FixZeroProbabilityStats: no rows with zero probability found")
		return 0, nil
	}

	a.log.Infof("FixZeroProbabilityStats: found %d rows with zero probability", len(zeroStats))

	// 확률 재계산
	updates := make([]AnalysisStat, 0, len(zeroStats))
	for _, stat := range zeroStats {
		totalTrials := float64(stat.DrawNo * 6)
		if totalTrials > 0 {
			stat.TotalProb = float64(stat.TotalCount) / totalTrials
		}
		updates = append(updates, stat)
	}

	// DB 업데이트
	if err := a.repo.UpdateAnalysisStatsTotalProb(ctx, updates); err != nil {
		a.log.Errorf("FixZeroProbabilityStats: failed to update stats: %v", err)
		return 0, err
	}

	a.log.Infof("FixZeroProbabilityStats: updated %d rows successfully", len(updates))
	return len(updates), nil
}

// FixZeroBonusProbabilityStats bonus_prob이 0인 행을 찾아서 수정
// 이미 bonus_count 값이 있는 행의 확률을 재계산하여 업데이트
func (a *Analyzer) FixZeroBonusProbabilityStats(ctx context.Context) (int, error) {
	a.log.Infof("FixZeroBonusProbabilityStats: starting")

	// bonus_prob이 0인 행 조회
	zeroStats, err := a.repo.GetAnalysisStatsWithZeroBonusProb(ctx)
	if err != nil {
		a.log.Errorf("FixZeroBonusProbabilityStats: failed to get zero bonus prob stats: %v", err)
		return 0, err
	}

	if len(zeroStats) == 0 {
		a.log.Infof("FixZeroBonusProbabilityStats: no rows with zero bonus probability found")
		return 0, nil
	}

	a.log.Infof("FixZeroBonusProbabilityStats: found %d rows with zero bonus probability", len(zeroStats))

	// 확률 재계산
	updates := make([]AnalysisStat, 0, len(zeroStats))
	for _, stat := range zeroStats {
		if stat.DrawNo > 0 {
			stat.BonusProb = float64(stat.BonusCount) / float64(stat.DrawNo)
		}
		updates = append(updates, stat)
	}

	// DB 업데이트
	if err := a.repo.UpdateAnalysisStatsBonusProb(ctx, updates); err != nil {
		a.log.Errorf("FixZeroBonusProbabilityStats: failed to update stats: %v", err)
		return 0, err
	}

	a.log.Infof("FixZeroBonusProbabilityStats: updated %d rows successfully", len(updates))
	return len(updates), nil
}
