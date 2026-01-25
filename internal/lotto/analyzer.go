package lotto

import (
	"context"

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
