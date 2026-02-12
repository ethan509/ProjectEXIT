package lotto

import (
	"context"
	"math/rand"
	"sort"
	"time"

	"github.com/example/LottoSmash/internal/logger"
)

const (
	TotalNumbers    = 45
	NumbersPerDraw  = 6
	DefaultTopCount = 10
)

// Recommender 추천 엔진
type Recommender struct {
	repo     *Repository
	analyzer *Analyzer
	log      *logger.Logger
	rng      *rand.Rand
}

// NewRecommender 새 추천 엔진 생성
func NewRecommender(repo *Repository, analyzer *Analyzer, log *logger.Logger) *Recommender {
	return &Recommender{
		repo:     repo,
		analyzer: analyzer,
		log:      log,
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// numberScore 번호와 점수를 함께 저장하는 구조체
type numberScore struct {
	Number int
	Score  float64
}

// Recommend 추천 실행 (메인 진입점)
func (r *Recommender) Recommend(ctx context.Context, req RecommendRequest) (*RecommendResponse, error) {
	if req.Count <= 0 {
		req.Count = 1
	}
	if req.Count > 10 {
		req.Count = 10
	}
	if req.CombineCode == "" {
		req.CombineCode = CombineSimpleAvg
	}

	latestDrawNo, err := r.repo.GetLatestDrawNo(ctx)
	if err != nil {
		return nil, err
	}

	recommendations := make([]Recommendation, 0, req.Count)

	for i := 0; i < req.Count; i++ {
		rec, err := r.generateSingleRecommendation(ctx, req)
		if err != nil {
			return nil, err
		}
		recommendations = append(recommendations, *rec)
	}

	return &RecommendResponse{
		Recommendations: recommendations,
		GeneratedAt:     time.Now(),
		LatestDrawNo:    latestDrawNo,
	}, nil
}

// generateSingleRecommendation 단일 추천 생성
func (r *Recommender) generateSingleRecommendation(ctx context.Context, req RecommendRequest) (*Recommendation, error) {
	details := make(map[string]interface{})

	stats, err := r.repo.GetLatestAnalysisStats(ctx)
	if err != nil {
		return nil, err
	}

	// 확률 조합 방식으로 추천
	var scores map[int]float64

	if req.CombineCode != "" {
		// 각 분석기법별 확률 맵 수집
		probMaps := make([]map[int]float64, 0, len(req.MethodCodes))
		for _, code := range req.MethodCodes {
			probMap := r.getMethodProbabilities(code, stats)
			probMaps = append(probMaps, probMap)
			details[code] = map[string]interface{}{
				"method": code,
				"type":   "probability_based",
			}
		}

		// 조합 방법 적용
		switch req.CombineCode {
		case CombineSimpleAvg:
			scores = r.combineSimpleAverage(probMaps)
		case CombineWeightedAvg:
			scores = r.combineWeightedAverage(probMaps, req.MethodCodes, req.Weights)
		case CombineBayesian:
			scores = r.combineBayesian(probMaps)
		default:
			// 아직 미구현 조합방법은 단순평균으로 폴백
			scores = r.combineSimpleAverage(probMaps)
		}
	} else {
		// 기존 순위 기반 방식 (하위 호환)
		scores = make(map[int]float64)
		for _, code := range req.MethodCodes {
			candidates, methodDetails, err := r.getMethodCandidates(ctx, code, stats)
			if err != nil {
				r.log.Errorf("failed to get candidates for %s: %v", code, err)
				continue
			}
			for i, num := range candidates {
				weight := 1.0 / float64(i+1)
				scores[num] += weight
			}
			details[code] = methodDetails
		}
	}

	// 점수 기준 상위 6개 선택
	numbers := r.selectTopNumbers(scores, NumbersPerDraw)
	sort.Ints(numbers)

	// 보너스 번호 선택 (요청 시)
	var bonus *int
	if req.IncludeBonus {
		bonusNum := r.selectBonusNumber(stats, numbers)
		bonus = &bonusNum
	}

	// 신뢰도 계산
	confidence := r.calculateCombineConfidence(numbers, scores, len(req.MethodCodes))

	return &Recommendation{
		Numbers:       numbers,
		Bonus:         bonus,
		MethodsUsed:   req.MethodCodes,
		CombineMethod: req.CombineCode,
		Confidence:    confidence,
		Details:       details,
	}, nil
}

// getMethodCandidates 분석 방법별 후보 번호 추출
func (r *Recommender) getMethodCandidates(ctx context.Context, code string, stats []AnalysisStat) ([]int, map[string]interface{}, error) {
	switch code {
	case "NUMBER_FREQUENCY":
		return r.recommendByFrequency(stats)
	case "REAPPEAR_PROB":
		return r.recommendByReappear(stats)
	case "FIRST_POSITION":
		return r.recommendByFirstPosition(stats)
	case "LAST_POSITION":
		return r.recommendByLastPosition(stats)
	case "PAIR_FREQUENCY":
		return r.recommendByPairFrequency(ctx)
	case "CONSECUTIVE":
		return r.recommendByConsecutive(stats)
	case "ODD_EVEN_RATIO":
		return r.recommendByOddEven(stats)
	case "HIGH_LOW_RATIO":
		return r.recommendByHighLow(stats)
	case "BAYESIAN":
		return r.recommendByBayesian(stats)
	case "HOT_COLD":
		return r.recommendByHotCold(ctx)
	default:
		return r.recommendByBayesian(stats) // 기본값
	}
}

// recommendByFrequency 출현 빈도 기반 추천
func (r *Recommender) recommendByFrequency(stats []AnalysisStat) ([]int, map[string]interface{}, error) {
	sorted := make([]AnalysisStat, len(stats))
	copy(sorted, stats)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].TotalCount > sorted[j].TotalCount
	})

	candidates := make([]int, 0, DefaultTopCount)
	for i := 0; i < DefaultTopCount && i < len(sorted); i++ {
		candidates = append(candidates, sorted[i].Number)
	}

	details := map[string]interface{}{
		"top_numbers": candidates,
		"method":      "total_count 기준 상위 번호",
	}

	return candidates, details, nil
}

// recommendByReappear 재등장 확률 기반 추천
func (r *Recommender) recommendByReappear(stats []AnalysisStat) ([]int, map[string]interface{}, error) {
	sorted := make([]AnalysisStat, len(stats))
	copy(sorted, stats)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ReappearProb > sorted[j].ReappearProb
	})

	candidates := make([]int, 0, DefaultTopCount)
	for i := 0; i < DefaultTopCount && i < len(sorted); i++ {
		candidates = append(candidates, sorted[i].Number)
	}

	details := map[string]interface{}{
		"top_numbers": candidates,
		"method":      "reappear_prob 기준 상위 번호",
	}

	return candidates, details, nil
}

// recommendByFirstPosition 첫번째 위치 기반 추천
func (r *Recommender) recommendByFirstPosition(stats []AnalysisStat) ([]int, map[string]interface{}, error) {
	// 첫번째 번호로 가능한 범위: 1~40 (실질적으로)
	filtered := make([]AnalysisStat, 0)
	for _, s := range stats {
		if s.Number <= 40 {
			filtered = append(filtered, s)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].FirstCount > filtered[j].FirstCount
	})

	candidates := make([]int, 0, DefaultTopCount)
	for i := 0; i < DefaultTopCount && i < len(filtered); i++ {
		candidates = append(candidates, filtered[i].Number)
	}

	details := map[string]interface{}{
		"top_numbers": candidates,
		"method":      "first_count 기준 상위 번호 (1~40 범위)",
	}

	return candidates, details, nil
}

// recommendByLastPosition 마지막 위치 기반 추천
func (r *Recommender) recommendByLastPosition(stats []AnalysisStat) ([]int, map[string]interface{}, error) {
	// 마지막 번호로 가능한 범위: 6~45 (실질적으로)
	filtered := make([]AnalysisStat, 0)
	for _, s := range stats {
		if s.Number >= 6 {
			filtered = append(filtered, s)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].LastCount > filtered[j].LastCount
	})

	candidates := make([]int, 0, DefaultTopCount)
	for i := 0; i < DefaultTopCount && i < len(filtered); i++ {
		candidates = append(candidates, filtered[i].Number)
	}

	details := map[string]interface{}{
		"top_numbers": candidates,
		"method":      "last_count 기준 상위 번호 (6~45 범위)",
	}

	return candidates, details, nil
}

// recommendByPairFrequency 동반 출현 기반 추천
func (r *Recommender) recommendByPairFrequency(ctx context.Context) ([]int, map[string]interface{}, error) {
	// analyzer의 CalculatePairStats 활용
	pairStats, err := r.analyzer.CalculatePairStats(ctx, 10)
	if err != nil {
		return nil, nil, err
	}

	// 상위 3개 쌍에서 번호 추출 (6개)
	numberSet := make(map[int]bool)
	for i := 0; i < 3 && i < len(pairStats.TopPairs); i++ {
		numberSet[pairStats.TopPairs[i].Number1] = true
		numberSet[pairStats.TopPairs[i].Number2] = true
	}

	candidates := make([]int, 0, len(numberSet))
	for num := range numberSet {
		candidates = append(candidates, num)
	}
	sort.Ints(candidates)

	details := map[string]interface{}{
		"top_pairs":   pairStats.TopPairs[:min(3, len(pairStats.TopPairs))],
		"top_numbers": candidates,
		"method":      "동반 출현 빈도 상위 쌍 기반",
	}

	return candidates, details, nil
}

// recommendByConsecutive 연번 패턴 기반 추천
func (r *Recommender) recommendByConsecutive(stats []AnalysisStat) ([]int, map[string]interface{}, error) {
	// 출현 빈도 상위 번호에서 연번 2개 포함하여 선택
	sorted := make([]AnalysisStat, len(stats))
	copy(sorted, stats)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].TotalProb > sorted[j].TotalProb
	})

	// 연번 쌍 찾기
	var consecutivePair []int
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			n1, n2 := sorted[i].Number, sorted[j].Number
			if n1 > n2 {
				n1, n2 = n2, n1
			}
			if n2-n1 == 1 {
				consecutivePair = []int{n1, n2}
				break
			}
		}
		if len(consecutivePair) > 0 {
			break
		}
	}

	// 연번 쌍이 없으면 상위 2개 연속 번호 임의 선택
	if len(consecutivePair) == 0 {
		baseNum := r.rng.Intn(43) + 1 // 1~43
		consecutivePair = []int{baseNum, baseNum + 1}
	}

	// 나머지 4개는 상위 확률 번호에서 선택
	candidates := make([]int, 0, DefaultTopCount)
	candidates = append(candidates, consecutivePair...)

	added := map[int]bool{consecutivePair[0]: true, consecutivePair[1]: true}
	for _, s := range sorted {
		if len(candidates) >= DefaultTopCount {
			break
		}
		if !added[s.Number] {
			candidates = append(candidates, s.Number)
			added[s.Number] = true
		}
	}

	details := map[string]interface{}{
		"consecutive_pair": consecutivePair,
		"top_numbers":      candidates,
		"method":           "연번 2개 포함 + 상위 확률 번호",
	}

	return candidates, details, nil
}

// recommendByOddEven 홀짝 비율 기반 추천
func (r *Recommender) recommendByOddEven(stats []AnalysisStat) ([]int, map[string]interface{}, error) {
	var oddNumbers, evenNumbers []AnalysisStat
	for _, s := range stats {
		if s.Number%2 == 1 {
			oddNumbers = append(oddNumbers, s)
		} else {
			evenNumbers = append(evenNumbers, s)
		}
	}

	// 각각 total_prob 기준 정렬
	sort.Slice(oddNumbers, func(i, j int) bool {
		return oddNumbers[i].TotalProb > oddNumbers[j].TotalProb
	})
	sort.Slice(evenNumbers, func(i, j int) bool {
		return evenNumbers[i].TotalProb > evenNumbers[j].TotalProb
	})

	// 홀수 5개 + 짝수 5개 (3:3 비율 추천을 위해 풀 확보)
	candidates := make([]int, 0, DefaultTopCount)
	for i := 0; i < 5 && i < len(oddNumbers); i++ {
		candidates = append(candidates, oddNumbers[i].Number)
	}
	for i := 0; i < 5 && i < len(evenNumbers); i++ {
		candidates = append(candidates, evenNumbers[i].Number)
	}

	details := map[string]interface{}{
		"odd_candidates":  candidates[:min(5, len(oddNumbers))],
		"even_candidates": candidates[min(5, len(oddNumbers)):],
		"method":          "홀수 3개 + 짝수 3개 추천",
	}

	return candidates, details, nil
}

// recommendByHighLow 고저 비율 기반 추천
func (r *Recommender) recommendByHighLow(stats []AnalysisStat) ([]int, map[string]interface{}, error) {
	var lowNumbers, highNumbers []AnalysisStat
	for _, s := range stats {
		if s.Number <= 22 {
			lowNumbers = append(lowNumbers, s)
		} else {
			highNumbers = append(highNumbers, s)
		}
	}

	// 각각 total_prob 기준 정렬
	sort.Slice(lowNumbers, func(i, j int) bool {
		return lowNumbers[i].TotalProb > lowNumbers[j].TotalProb
	})
	sort.Slice(highNumbers, func(i, j int) bool {
		return highNumbers[i].TotalProb > highNumbers[j].TotalProb
	})

	// 저번호 5개 + 고번호 5개
	candidates := make([]int, 0, DefaultTopCount)
	for i := 0; i < 5 && i < len(lowNumbers); i++ {
		candidates = append(candidates, lowNumbers[i].Number)
	}
	for i := 0; i < 5 && i < len(highNumbers); i++ {
		candidates = append(candidates, highNumbers[i].Number)
	}

	details := map[string]interface{}{
		"low_candidates":  candidates[:min(5, len(lowNumbers))],
		"high_candidates": candidates[min(5, len(lowNumbers)):],
		"method":          "저번호(1-22) 3개 + 고번호(23-45) 3개 추천",
	}

	return candidates, details, nil
}

// recommendByBayesian 베이지안 분석 기반 추천
func (r *Recommender) recommendByBayesian(stats []AnalysisStat) ([]int, map[string]interface{}, error) {
	sorted := make([]AnalysisStat, len(stats))
	copy(sorted, stats)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].BayesianPost > sorted[j].BayesianPost
	})

	candidates := make([]int, 0, DefaultTopCount)
	for i := 0; i < DefaultTopCount && i < len(sorted); i++ {
		candidates = append(candidates, sorted[i].Number)
	}

	details := map[string]interface{}{
		"top_numbers": candidates,
		"method":      "bayesian_post 사후확률 기준 상위 번호",
	}

	return candidates, details, nil
}

// recommendByHotCold HOT/COLD 조합 기반 추천
func (r *Recommender) recommendByHotCold(ctx context.Context) ([]int, map[string]interface{}, error) {
	bayesianStats, err := r.analyzer.CalculateBayesianStats(ctx, 50)
	if err != nil {
		return nil, nil, err
	}

	// HOT 5개
	hotNumbers := make([]int, 0, 5)
	for i := 0; i < 5 && i < len(bayesianStats.HotNumbers); i++ {
		hotNumbers = append(hotNumbers, bayesianStats.HotNumbers[i].Number)
	}

	// COLD 5개
	coldNumbers := make([]int, 0, 5)
	for i := 0; i < 5 && i < len(bayesianStats.ColdNumbers); i++ {
		coldNumbers = append(coldNumbers, bayesianStats.ColdNumbers[i].Number)
	}

	candidates := append(hotNumbers, coldNumbers...)

	details := map[string]interface{}{
		"hot_numbers":  hotNumbers,
		"cold_numbers": coldNumbers,
		"window_size":  bayesianStats.WindowSize,
		"method":       "HOT(최근 자주 출현) 3개 + COLD(출현 적음) 3개 조합",
	}

	return candidates, details, nil
}

// selectTopNumbers 점수 기준 상위 N개 번호 선택
func (r *Recommender) selectTopNumbers(scores map[int]float64, count int) []int {
	scoreSlice := make([]numberScore, 0, len(scores))
	for num, score := range scores {
		scoreSlice = append(scoreSlice, numberScore{Number: num, Score: score})
	}

	sort.Slice(scoreSlice, func(i, j int) bool {
		if scoreSlice[i].Score == scoreSlice[j].Score {
			// 동점 시 랜덤
			return r.rng.Intn(2) == 0
		}
		return scoreSlice[i].Score > scoreSlice[j].Score
	})

	numbers := make([]int, 0, count)
	for i := 0; i < count && i < len(scoreSlice); i++ {
		numbers = append(numbers, scoreSlice[i].Number)
	}

	// 부족하면 랜덤으로 채움
	if len(numbers) < count {
		added := make(map[int]bool)
		for _, n := range numbers {
			added[n] = true
		}
		for len(numbers) < count {
			n := r.rng.Intn(TotalNumbers) + 1
			if !added[n] {
				numbers = append(numbers, n)
				added[n] = true
			}
		}
	}

	return numbers
}

// selectBonusNumber 보너스 번호 선택
func (r *Recommender) selectBonusNumber(stats []AnalysisStat, excludeNumbers []int) int {
	excludeSet := make(map[int]bool)
	for _, n := range excludeNumbers {
		excludeSet[n] = true
	}

	// 보너스 확률 기준 정렬
	sorted := make([]AnalysisStat, 0)
	for _, s := range stats {
		if !excludeSet[s.Number] {
			sorted = append(sorted, s)
		}
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].BonusProb > sorted[j].BonusProb
	})

	if len(sorted) > 0 {
		return sorted[0].Number
	}

	// 폴백: 랜덤 선택
	for {
		n := r.rng.Intn(TotalNumbers) + 1
		if !excludeSet[n] {
			return n
		}
	}
}

// calculateConfidence 신뢰도 계산
func (r *Recommender) calculateConfidence(numbers []int, scores map[int]float64, methodCount int) float64 {
	if methodCount == 0 {
		return 0.5
	}

	// 선택된 번호들의 점수 합계
	totalScore := 0.0
	for _, n := range numbers {
		totalScore += scores[n]
	}

	// 이론적 최대값: 각 방법당 상위 6개 점수 (1 + 0.5 + 0.33 + 0.25 + 0.2 + 0.166 ≈ 2.45)
	maxScorePerMethod := 2.45
	maxTotal := maxScorePerMethod * float64(methodCount)

	confidence := totalScore / maxTotal
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}

// ========================================
// 확률 조합 관련 메서드
// ========================================

// getMethodProbabilities 분석기법별 번호(1~45) 확률 맵 반환
func (r *Recommender) getMethodProbabilities(code string, stats []AnalysisStat) map[int]float64 {
	probMap := make(map[int]float64, TotalNumbers)

	for _, s := range stats {
		switch code {
		case "NUMBER_FREQUENCY":
			probMap[s.Number] = s.TotalProb
		case "REAPPEAR_PROB":
			probMap[s.Number] = s.ReappearProb
		case "FIRST_POSITION":
			probMap[s.Number] = s.FirstProb
		case "LAST_POSITION":
			probMap[s.Number] = s.LastProb
		case "PAIR_FREQUENCY":
			probMap[s.Number] = s.TotalProb // 쌍 기반이지만 개별 번호 확률로 대체
		case "CONSECUTIVE":
			probMap[s.Number] = s.TotalProb
		case "ODD_EVEN_RATIO":
			probMap[s.Number] = s.TotalProb
		case "HIGH_LOW_RATIO":
			probMap[s.Number] = s.TotalProb
		case "BAYESIAN":
			probMap[s.Number] = s.BayesianPost
		case "HOT_COLD":
			probMap[s.Number] = s.BayesianPost
		default:
			probMap[s.Number] = s.TotalProb
		}
	}

	return probMap
}

// combineSimpleAverage 단순 평균 조합: 각 번호별 확률을 산술 평균
func (r *Recommender) combineSimpleAverage(probMaps []map[int]float64) map[int]float64 {
	if len(probMaps) == 0 {
		return make(map[int]float64)
	}

	combined := make(map[int]float64, TotalNumbers)
	count := float64(len(probMaps))

	for num := 1; num <= TotalNumbers; num++ {
		sum := 0.0
		for _, pm := range probMaps {
			sum += pm[num]
		}
		combined[num] = sum / count
	}

	return combined
}

// combineWeightedAverage 가중 평균 조합: 유저 지정 가중치로 확률을 가중 평균
func (r *Recommender) combineWeightedAverage(probMaps []map[int]float64, methodCodes []string, weights map[string]float64) map[int]float64 {
	if len(probMaps) == 0 {
		return make(map[int]float64)
	}

	// 가중치 합계 계산 (정규화용)
	totalWeight := 0.0
	for _, code := range methodCodes {
		totalWeight += weights[code]
	}
	if totalWeight == 0 {
		// 가중치 합이 0이면 단순 평균으로 폴백
		return r.combineSimpleAverage(probMaps)
	}

	combined := make(map[int]float64, TotalNumbers)
	for num := 1; num <= TotalNumbers; num++ {
		weightedSum := 0.0
		for i, pm := range probMaps {
			w := weights[methodCodes[i]]
			weightedSum += w * pm[num]
		}
		combined[num] = weightedSum / totalWeight
	}

	return combined
}

// combineBayesian 베이지안 결합: 각 확률을 독립 증거로 취급하여 결합
// P_combined(n) = ∏P_i(n) / (∏P_i(n) + ∏(1-P_i(n)))
func (r *Recommender) combineBayesian(probMaps []map[int]float64) map[int]float64 {
	if len(probMaps) == 0 {
		return make(map[int]float64)
	}

	const epsilon = 1e-10 // 0/1 클램핑용

	combined := make(map[int]float64, TotalNumbers)
	for num := 1; num <= TotalNumbers; num++ {
		prodP := 1.0    // ∏P_i(n)
		prodNotP := 1.0 // ∏(1-P_i(n))
		for _, pm := range probMaps {
			p := pm[num]
			// 0과 1 클램핑 (log 안전)
			if p < epsilon {
				p = epsilon
			}
			if p > 1-epsilon {
				p = 1 - epsilon
			}
			prodP *= p
			prodNotP *= (1 - p)
		}
		combined[num] = prodP / (prodP + prodNotP)
	}

	return combined
}

// calculateCombineConfidence 확률 조합 기반 신뢰도 계산
func (r *Recommender) calculateCombineConfidence(numbers []int, scores map[int]float64, methodCount int) float64 {
	if methodCount == 0 || len(numbers) == 0 {
		return 0.0
	}

	// 선택된 번호들의 평균 확률
	totalProb := 0.0
	for _, n := range numbers {
		totalProb += scores[n]
	}
	avgProb := totalProb / float64(len(numbers))

	// 기대 확률(1/45 ≈ 0.0222)과 비교하여 신뢰도 산출
	// 기대 확률보다 얼마나 높은지를 0~1 범위로 정규화
	expectedProb := 1.0 / float64(TotalNumbers)
	confidence := avgProb / (expectedProb * 3) // 기대치의 3배를 1.0으로 설정

	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
