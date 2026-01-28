package lotto

import (
	"context"
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

	a.log.Infof("RunFullAnalysis: completed successfully")
	return nil
}
