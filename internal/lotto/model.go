package lotto

import "time"

// LottoDraw 로또 당첨번호
type LottoDraw struct {
	DrawNo   int    `json:"draw_no"`
	DrawDate string `json:"draw_date"`
	Num1     int    `json:"num1"`
	Num2     int    `json:"num2"`
	Num3     int    `json:"num3"`
	Num4     int    `json:"num4"`
	Num5     int    `json:"num5"`
	Num6     int    `json:"num6"`
	BonusNum int    `json:"bonus_num"`
	// 1등 정보
	FirstPrize   int64 `json:"first_prize"`
	FirstWinners int   `json:"first_winners"`
	FirstPerGame int64 `json:"first_per_game"`
	// 2등 정보
	SecondPrize   int64 `json:"second_prize"`
	SecondWinners int   `json:"second_winners"`
	SecondPerGame int64 `json:"second_per_game"`
	// 3등 정보
	ThirdPrize   int64 `json:"third_prize"`
	ThirdWinners int   `json:"third_winners"`
	ThirdPerGame int64 `json:"third_per_game"`
	// 4등 정보
	FourthPrize   int64 `json:"fourth_prize"`
	FourthWinners int   `json:"fourth_winners"`
	FourthPerGame int64 `json:"fourth_per_game"`
	// 5등 정보
	FifthPrize   int64     `json:"fifth_prize"`
	FifthWinners int       `json:"fifth_winners"`
	FifthPerGame int64     `json:"fifth_per_game"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Numbers 당첨번호 배열 반환
func (d *LottoDraw) Numbers() []int {
	return []int{d.Num1, d.Num2, d.Num3, d.Num4, d.Num5, d.Num6}
}

// NumberStat 번호별 통계
type NumberStat struct {
	ID           int64     `json:"id"`
	Number       int       `json:"number"`
	TotalCount   int       `json:"total_count"`
	BonusCount   int       `json:"bonus_count"`
	LastDrawNo   int       `json:"last_draw_no"`
	CalculatedAt time.Time `json:"calculated_at"`
}

// ReappearStat 번호 재등장 통계
type ReappearStat struct {
	Number        int     `json:"number"`
	TotalAppear   int     `json:"total_appear"`
	ReappearCount int     `json:"reappear_count"`
	Probability   float64 `json:"probability"`
}

// PositionStat 포지션별 번호 통계 (첫번째/마지막 번호)
type PositionStat struct {
	Number      int     `json:"number"`
	Count       int     `json:"count"`
	Probability float64 `json:"probability"`
}

// FirstLastStatsResponse 첫번째/마지막 번호 확률 응답
type FirstLastStatsResponse struct {
	FirstStats   []PositionStat `json:"first_stats"`
	LastStats    []PositionStat `json:"last_stats"`
	TotalDraws   int            `json:"total_draws"`
	LatestDrawNo int            `json:"latest_draw_no"`
}

// PairStat 번호 쌍 동반 출현 통계 (API 응답용)
type PairStat struct {
	Number1     int     `json:"number1"`
	Number2     int     `json:"number2"`
	Count       int     `json:"count"`
	Probability float64 `json:"probability"`
}

// PairStatDB 번호 쌍 동시출현 통계 (DB 저장용)
// 회차별로 모든 번호 쌍(45C2=990개)의 누적 통계를 저장
type PairStatDB struct {
	DrawNo       int       `json:"draw_no"`       // 회차 번호
	Number1      int       `json:"number1"`       // 작은 번호 (1~44)
	Number2      int       `json:"number2"`       // 큰 번호 (2~45), number1 < number2
	Count        int       `json:"count"`         // 누적 동시출현 횟수
	Prob         float64   `json:"prob"`          // 동시출현 확률 (count / draw_no)
	CalculatedAt time.Time `json:"calculated_at"`
}

// PairStatsResponse 동반 출현 통계 응답
type PairStatsResponse struct {
	TopPairs     []PairStat `json:"top_pairs"`
	BottomPairs  []PairStat `json:"bottom_pairs"`
	TotalDraws   int        `json:"total_draws"`
	LatestDrawNo int        `json:"latest_draw_no"`
}

// ConsecutiveCountStat 연번 개수별 통계
type ConsecutiveCountStat struct {
	ConsecutiveCount int     `json:"consecutive_count"` // 연번 개수 (0, 2, 3, 4, 5, 6)
	DrawCount        int     `json:"draw_count"`        // 해당 연번 개수가 나온 회차 수
	Probability      float64 `json:"probability"`       // 확률
}

// ConsecutiveExample 연번 포함 회차 예시
type ConsecutiveExample struct {
	DrawNo           int   `json:"draw_no"`
	Numbers          []int `json:"numbers"`
	ConsecutiveCount int   `json:"consecutive_count"`
}

// ConsecutiveStatsResponse 연번 패턴 통계 응답
type ConsecutiveStatsResponse struct {
	CountStats     []ConsecutiveCountStat `json:"count_stats"`
	RecentExamples []ConsecutiveExample   `json:"recent_examples"`
	TotalDraws     int                    `json:"total_draws"`
	LatestDrawNo   int                    `json:"latest_draw_no"`
}

// ConsecutiveStatDB 연번 통계 (DB 저장용)
// 회차별 연번 개수 확률 추이를 저장
type ConsecutiveStatDB struct {
	DrawNo       int       `json:"draw_no"`       // 회차 번호
	ActualCount  int       `json:"actual_count"`  // 해당 회차의 실제 연번 개수
	Count0       int       `json:"count_0"`       // 연번 0개 누적 횟수
	Count2       int       `json:"count_2"`       // 연번 2개 누적 횟수
	Count3       int       `json:"count_3"`       // 연번 3개 누적 횟수
	Count4       int       `json:"count_4"`       // 연번 4개 누적 횟수
	Count5       int       `json:"count_5"`       // 연번 5개 누적 횟수
	Count6       int       `json:"count_6"`       // 연번 6개 누적 횟수
	Prob0        float64   `json:"prob_0"`        // 연번 0개 확률
	Prob2        float64   `json:"prob_2"`        // 연번 2개 확률
	Prob3        float64   `json:"prob_3"`        // 연번 3개 확률
	Prob4        float64   `json:"prob_4"`        // 연번 4개 확률
	Prob5        float64   `json:"prob_5"`        // 연번 5개 확률
	Prob6        float64   `json:"prob_6"`        // 연번 6개 확률
	CalculatedAt time.Time `json:"calculated_at"`
}

// OddEvenStatDB 홀짝 비율 통계 (DB 저장용)
// 회차별 홀수:짝수 비율 확률 추이를 저장
type OddEvenStatDB struct {
	DrawNo       int       `json:"draw_no"`       // 회차 번호
	ActualRatio  string    `json:"actual_ratio"`  // 해당 회차의 실제 비율 (예: "3:3")
	Count0_6     int       `json:"count_0_6"`     // 홀0:짝6 누적 횟수
	Count1_5     int       `json:"count_1_5"`     // 홀1:짝5 누적 횟수
	Count2_4     int       `json:"count_2_4"`     // 홀2:짝4 누적 횟수
	Count3_3     int       `json:"count_3_3"`     // 홀3:짝3 누적 횟수
	Count4_2     int       `json:"count_4_2"`     // 홀4:짝2 누적 횟수
	Count5_1     int       `json:"count_5_1"`     // 홀5:짝1 누적 횟수
	Count6_0     int       `json:"count_6_0"`     // 홀6:짝0 누적 횟수
	Prob0_6      float64   `json:"prob_0_6"`      // 홀0:짝6 확률
	Prob1_5      float64   `json:"prob_1_5"`      // 홀1:짝5 확률
	Prob2_4      float64   `json:"prob_2_4"`      // 홀2:짝4 확률
	Prob3_3      float64   `json:"prob_3_3"`      // 홀3:짝3 확률
	Prob4_2      float64   `json:"prob_4_2"`      // 홀4:짝2 확률
	Prob5_1      float64   `json:"prob_5_1"`      // 홀5:짝1 확률
	Prob6_0      float64   `json:"prob_6_0"`      // 홀6:짝0 확률
	CalculatedAt time.Time `json:"calculated_at"`
}

// HighLowStatDB 고저 비율 통계 (DB 저장용)
// 회차별 고번호(23~45):저번호(1~22) 비율 확률 추이를 저장
type HighLowStatDB struct {
	DrawNo       int       `json:"draw_no"`       // 회차 번호
	ActualRatio  string    `json:"actual_ratio"`  // 해당 회차의 실제 비율 (예: "3:3")
	Count0_6     int       `json:"count_0_6"`     // 고0:저6 누적 횟수
	Count1_5     int       `json:"count_1_5"`     // 고1:저5 누적 횟수
	Count2_4     int       `json:"count_2_4"`     // 고2:저4 누적 횟수
	Count3_3     int       `json:"count_3_3"`     // 고3:저3 누적 횟수
	Count4_2     int       `json:"count_4_2"`     // 고4:저2 누적 횟수
	Count5_1     int       `json:"count_5_1"`     // 고5:저1 누적 횟수
	Count6_0     int       `json:"count_6_0"`     // 고6:저0 누적 횟수
	Prob0_6      float64   `json:"prob_0_6"`      // 고0:저6 확률
	Prob1_5      float64   `json:"prob_1_5"`      // 고1:저5 확률
	Prob2_4      float64   `json:"prob_2_4"`      // 고2:저4 확률
	Prob3_3      float64   `json:"prob_3_3"`      // 고3:저3 확률
	Prob4_2      float64   `json:"prob_4_2"`      // 고4:저2 확률
	Prob5_1      float64   `json:"prob_5_1"`      // 고5:저1 확률
	Prob6_0      float64   `json:"prob_6_0"`      // 고6:저0 확률
	CalculatedAt time.Time `json:"calculated_at"`
}

// RatioStat 비율별 통계 (홀짝, 고저)
type RatioStat struct {
	Ratio       string  `json:"ratio"`       // 비율 표현 (예: "3:3", "4:2")
	Count       int     `json:"count"`       // 해당 비율이 나온 회차 수
	Probability float64 `json:"probability"` // 확률
}

// RatioStatsResponse 홀짝/고저 비율 통계 응답
type RatioStatsResponse struct {
	OddEvenStats []RatioStat `json:"odd_even_stats"` // 홀:짝 비율
	HighLowStats []RatioStat `json:"high_low_stats"` // 고:저 비율 (23~45:1~22)
	TotalDraws   int         `json:"total_draws"`
	LatestDrawNo int         `json:"latest_draw_no"`
}

// ColorPatternStat 색상 패턴별 통계
// 한국 로또 공식 색상: 1~10(노랑Y), 11~20(파랑B), 21~30(빨강R), 31~40(회색G), 41~45(초록E)
type ColorPatternStat struct {
	Pattern     string  `json:"pattern"`     // 색상 패턴 (예: "YBRRGR")
	Count       int     `json:"count"`       // 출현 횟수
	Probability float64 `json:"probability"` // 확률
}

// ColorStatsResponse 색상 패턴 통계 응답
type ColorStatsResponse struct {
	TopPatterns  []ColorPatternStat `json:"top_patterns"` // 가장 많이 나온 패턴
	ColorCounts  map[string]int     `json:"color_counts"` // 각 색상별 총 출현 횟수
	TotalDraws   int                `json:"total_draws"`
	LatestDrawNo int                `json:"latest_draw_no"`
}

// LineStat 행/열별 통계
type LineStat struct {
	Line        int     `json:"line"`        // 행 또는 열 번호 (1~7)
	Count       int     `json:"count"`       // 해당 라인에 번호가 포함된 총 횟수
	Probability float64 `json:"probability"` // 확률
}

// LineDistStat 행/열 분포별 통계 (예: "2:1:1:1:1:0:0" = 각 행에 몇 개)
type LineDistStat struct {
	Distribution string  `json:"distribution"` // 분포 패턴
	Count        int     `json:"count"`        // 출현 횟수
	Probability  float64 `json:"probability"`  // 확률
}

// RowColStatsResponse 행/열 분포 통계 응답 (7x7 격자 기준)
type RowColStatsResponse struct {
	RowStats       []LineStat     `json:"row_stats"`        // 각 행(1~7)별 통계
	ColStats       []LineStat     `json:"col_stats"`        // 각 열(1~7)별 통계
	TopRowPatterns []LineDistStat `json:"top_row_patterns"` // 가장 많이 나온 행 분포 패턴
	TopColPatterns []LineDistStat `json:"top_col_patterns"` // 가장 많이 나온 열 분포 패턴
	TotalDraws     int            `json:"total_draws"`
	LatestDrawNo   int            `json:"latest_draw_no"`
}

// DhlotteryResponse 동행복권 API 응답
type DhlotteryResponse struct {
	ReturnValue    string `json:"returnValue"`
	DrwNo          int    `json:"drwNo"`
	DrwNoDate      string `json:"drwNoDate"`
	DrwtNo1        int    `json:"drwtNo1"`
	DrwtNo2        int    `json:"drwtNo2"`
	DrwtNo3        int    `json:"drwtNo3"`
	DrwtNo4        int    `json:"drwtNo4"`
	DrwtNo5        int    `json:"drwtNo5"`
	DrwtNo6        int    `json:"drwtNo6"`
	BnusNo         int    `json:"bnusNo"`
	FirstAccumamnt int64  `json:"firstAccumamnt"`
	FirstPrzwnerCo int    `json:"firstPrzwnerCo"`
}

// ToLottoDraw API 응답을 LottoDraw로 변환
func (r *DhlotteryResponse) ToLottoDraw() (*LottoDraw, error) {
	drawDate, err := time.Parse("2006-01-02", r.DrwNoDate)
	if err != nil {
		return nil, err
	}

	draw := &LottoDraw{
		DrawNo:       r.DrwNo,
		DrawDate:     drawDate.Format("2006.01.02"),
		Num1:         r.DrwtNo1,
		Num2:         r.DrwtNo2,
		Num3:         r.DrwtNo3,
		Num4:         r.DrwtNo4,
		Num5:         r.DrwtNo5,
		Num6:         r.DrwtNo6,
		BonusNum:     r.BnusNo,
		FirstPrize:   r.FirstAccumamnt,
		FirstWinners: r.FirstPrzwnerCo,
	}

	if draw.FirstWinners > 0 {
		draw.FirstPerGame = draw.FirstPrize / int64(draw.FirstWinners)
	}

	return draw, nil
}

// DrawListResponse 당첨번호 목록 응답
type DrawListResponse struct {
	Draws      []LottoDraw `json:"draws"`
	TotalCount int         `json:"total_count"`
	LatestDraw int         `json:"latest_draw"`
}

// StatsResponse 통계 응답
type StatsResponse struct {
	NumberStats   []NumberStat   `json:"number_stats"`
	ReappearStats []ReappearStat `json:"reappear_stats"`
	LatestDrawNo  int            `json:"latest_draw_no"`
	CalculatedAt  time.Time      `json:"calculated_at"`
}

// BayesianNumberStat 베이지안 추론 기반 번호별 통계
// P(θ|D) ∝ P(D|θ)P(θ) - 사후확률은 우도와 사전확률의 곱에 비례
type BayesianNumberStat struct {
	Number           int     `json:"number"`           // 번호 (1~45)
	Prior            float64 `json:"prior"`            // 사전 확률 P(θ) = 1/45
	Likelihood       float64 `json:"likelihood"`       // 우도 P(D|θ) - 최근 N회차 출현 빈도
	Posterior        float64 `json:"posterior"`        // 사후 확률 P(θ|D)
	RecentCount      int     `json:"recent_count"`     // 최근 N회차 출현 횟수
	ExpectedCount    float64 `json:"expected_count"`   // 기대 출현 횟수
	Deviation        float64 `json:"deviation"`        // 편차 (실제 - 기대)
	Status           string  `json:"status"`           // HOT, COLD, NEUTRAL
	LastAppearDrawNo int     `json:"last_appear_draw"` // 마지막 출현 회차
	GapSinceLastDraw int     `json:"gap_since_last"`   // 마지막 출현 후 경과 회차
}

// BayesianStatsResponse 베이지안 분석 응답
type BayesianStatsResponse struct {
	Numbers      []BayesianNumberStat `json:"numbers"`        // 전체 번호 통계
	HotNumbers   []BayesianNumberStat `json:"hot_numbers"`    // HOT 번호 (상위)
	ColdNumbers  []BayesianNumberStat `json:"cold_numbers"`   // COLD 번호 (하위)
	WindowSize   int                  `json:"window_size"`    // 분석 윈도우 크기 (최근 N회차)
	TotalDraws   int                  `json:"total_draws"`    // 전체 회차 수
	LatestDrawNo int                  `json:"latest_draw_no"` // 최신 회차 번호
}

// BayesianStat DB 저장용 누적 베이지안 통계
// 각 회차별, 각 번호별 확률을 누적하여 저장
type BayesianStat struct {
	ID           int64     `json:"id"`
	DrawNo       int       `json:"draw_no"`       // 회차 번호
	Number       int       `json:"number"`        // 번호 (1~45)
	TotalCount   int       `json:"total_count"`   // 1회차부터 해당 회차까지 누적 출현 횟수
	TotalDraws   int       `json:"total_draws"`   // 총 회차 수
	Prior        float64   `json:"prior"`         // 사전 확률 (이전 회차의 posterior)
	Posterior    float64   `json:"posterior"`     // 사후 확률
	Appeared     bool      `json:"appeared"`      // 해당 회차에 출현했는지
	CalculatedAt time.Time `json:"calculated_at"`
}

// AnalysisStat 통합 분석 통계 (회차별, 번호별)
// 모든 분석 결과를 하나의 테이블에 저장
type AnalysisStat struct {
	DrawNo        int       `json:"draw_no"`        // 회차 번호
	Number        int       `json:"number"`         // 번호 (1~45)
	TotalCount    int       `json:"total_count"`    // 누적 출현 횟수
	TotalProb     float64   `json:"total_prob"`     // 출현 확률 (total_count / (draw_no * 6))
	BonusCount    int       `json:"bonus_count"`    // 보너스 출현 횟수
	BonusProb     float64   `json:"bonus_prob"`     // 보너스 출현 확률 (bonus_count / draw_no)
	FirstCount    int       `json:"first_count"`    // 첫번째 위치(Num1)로 나온 누적 횟수
	FirstProb     float64   `json:"first_prob"`     // 첫번째 위치 확률 (first_count / draw_no)
	LastCount     int       `json:"last_count"`     // 마지막 위치(Num6)로 나온 누적 횟수
	LastProb      float64   `json:"last_prob"`      // 마지막 위치 확률 (last_count / draw_no)
	ReappearTotal int       `json:"reappear_total"` // 재등장 기준 총 출현
	ReappearCount int       `json:"reappear_count"` // 재등장 횟수
	ReappearProb  float64   `json:"reappear_prob"`  // 재등장 확률
	BayesianPrior float64   `json:"bayesian_prior"` // 베이지안 사전 확률
	BayesianPost  float64   `json:"bayesian_post"`  // 베이지안 사후 확률
	ColorCount    int       `json:"color_count"`    // 해당 번호 색상의 총 출현 횟수
	ColorProb     float64   `json:"color_prob"`     // 색상 출현 확률
	RowCount      int       `json:"row_count"`      // 해당 번호 행의 총 출현 횟수
	RowProb       float64   `json:"row_prob"`       // 행 출현 확률
	ColCount      int       `json:"col_count"`      // 해당 번호 열의 총 출현 횟수
	ColProb       float64   `json:"col_prob"`       // 열 출현 확률
	Appeared      bool      `json:"appeared"`       // 해당 회차 출현 여부
	CalculatedAt  time.Time `json:"calculated_at"`
}

// UnclaimedPrize 미수령 당첨금
type UnclaimedPrize struct {
	ID             int64     `json:"id"`
	DrawNo         int       `json:"draw_no"`
	PrizeRank      int       `json:"prize_rank"`      // 1등 또는 2등
	Amount         int64     `json:"amount"`          // 당첨금
	WinnerName     *string   `json:"winner_name"`     // 당첨자명 (마스킹됨, 예: 김**)
	WinningDate    string    `json:"winning_date"`    // 당첨일
	ExpirationDate string    `json:"expiration_date"` // 만기일
	Status         string    `json:"status"`          // unclaimed, claimed 등
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ========================================
// 추천 기능 관련 모델
// ========================================

// AnalysisMethod 분석 방법 메타데이터
type AnalysisMethod struct {
	ID          int       `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Category    string    `json:"category"`
	IsActive    bool      `json:"is_active"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}

// LottoRecommendation 추천 기록
type LottoRecommendation struct {
	ID            int64     `json:"id"`
	UserID        *int64    `json:"user_id,omitempty"`
	MethodCodes   []string  `json:"method_codes"`
	CombineMethod string    `json:"combine_method"`
	Numbers       []int     `json:"numbers"`
	BonusNumber   *int      `json:"bonus_number,omitempty"`
	Confidence    float64   `json:"confidence"`
	CreatedAt     time.Time `json:"created_at"`
}

// RecommendRequest 추천 요청
type RecommendRequest struct {
	MethodCodes  []string           `json:"method_codes"`
	CombineCode  string             `json:"combine_code"`            // 조합 방법 코드 (기본값: SIMPLE_AVG)
	Weights      map[string]float64 `json:"weights,omitempty"`       // 가중 평균 시 기법별 가중치 (예: {"BAYESIAN": 0.5, "NUMBER_FREQUENCY": 0.3})
	IncludeBonus bool               `json:"include_bonus"`
	Count        int                `json:"count"`                   // 추천 세트 개수 (기본값: 1, 최대: 10)
}

// Recommendation 단일 추천 결과
type Recommendation struct {
	Numbers       []int                  `json:"numbers"`
	Bonus         *int                   `json:"bonus,omitempty"`
	MethodsUsed   []string               `json:"methods_used"`
	CombineMethod string                 `json:"combine_method"`
	Confidence    float64                `json:"confidence"`
	Details       map[string]interface{} `json:"details,omitempty"`
}

// RecommendResponse 추천 응답
type RecommendResponse struct {
	Recommendations []Recommendation `json:"recommendations"`
	GeneratedAt     time.Time        `json:"generated_at"`
	LatestDrawNo    int              `json:"latest_draw_no"`
}

// MethodListResponse 분석 방법 목록 응답
type MethodListResponse struct {
	Methods    []AnalysisMethod `json:"methods"`
	TotalCount int              `json:"total_count"`
}

// ========================================
// 확률 조합 방법 관련 모델
// ========================================

// 조합 방법 코드 상수
const (
	CombineSimpleAvg     = "SIMPLE_AVG"
	CombineWeightedAvg   = "WEIGHTED_AVG"
	CombineBayesian      = "BAYESIAN_COMBINE"
	CombineGeometricMean = "GEOMETRIC_MEAN"
	CombineMinMax        = "MIN_MAX"

	MaxMethodCodes = 3 // 최대 선택 가능한 분석기법 수
)

// CombineMethod 확률 조합 방법 메타데이터
type CombineMethod struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
	SortOrder   int    `json:"sort_order"`
}

// CombineMethodListResponse 조합 방법 목록 응답
type CombineMethodListResponse struct {
	Methods    []CombineMethod `json:"methods"`
	TotalCount int             `json:"total_count"`
}

// AllCombineMethods 전체 조합 방법 목록 (하드코딩)
var AllCombineMethods = []CombineMethod{
	{Code: CombineSimpleAvg, Name: "단순 평균", Description: "선택한 기법들의 확률을 단순 평균하여 조합", IsActive: true, SortOrder: 1},
	{Code: CombineWeightedAvg, Name: "가중 평균", Description: "각 기법별 가중치를 직접 지정하여 평균", IsActive: true, SortOrder: 2},
	{Code: CombineBayesian, Name: "베이지안 결합", Description: "베이지안 확률 결합으로 두 확률을 보수적으로 조합", IsActive: true, SortOrder: 3},
	{Code: CombineGeometricMean, Name: "기하 평균", Description: "확률의 기하 평균으로 낮은 확률에 더 민감하게 반응", IsActive: false, SortOrder: 4},
	{Code: CombineMinMax, Name: "최대/최소 기반", Description: "낙관적(최대) 또는 보수적(최소) 확률 선택", IsActive: false, SortOrder: 5},
}
