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

// PairStat 번호 쌍 동반 출현 통계
type PairStat struct {
	Number1     int     `json:"number1"`
	Number2     int     `json:"number2"`
	Count       int     `json:"count"`
	Probability float64 `json:"probability"`
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
