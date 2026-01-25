package lotto

import "time"

// LottoDraw 로또 당첨번호
type LottoDraw struct {
	DrawNo       int       `json:"draw_no"`
	DrawDate     string    `json:"draw_date"`
	Num1         int       `json:"num1"`
	Num2         int       `json:"num2"`
	Num3         int       `json:"num3"`
	Num4         int       `json:"num4"`
	Num5         int       `json:"num5"`
	Num6         int       `json:"num6"`
	BonusNum     int       `json:"bonus_num"`
	FirstPrize   int64     `json:"first_prize"`
	FirstWinners int       `json:"first_winners"`
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

	return &LottoDraw{
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
	}, nil
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
