// Package constants 프로젝트 전역 상수 정의
package constants

import "time"

// 로또 번호 관련 상수
const (
	MinLottoNumber   = 1
	MaxLottoNumber   = 45
	TotalLottoNumber = MaxLottoNumber - MinLottoNumber + 1
	LottoNumberCount = 6 // 당첨번호 개수 (보너스 제외)
)

// 로또 회차 관련 상수
const (
	FirstDrawYear  = 2002
	FirstDrawMonth = 12
	FirstDrawDay   = 7
)

// 동행복권 URL 및 클라이언트 설정
const (
	DHLotteryBaseURL   = "https://www.dhlottery.co.kr/common.do?method=getLottoNumber&drwNo=%d"
	DHLotteryHTMLURL   = "https://www.dhlottery.co.kr/lt645/result?drwNo=%d"
	DHLotteryLatestURL = "https://www.dhlottery.co.kr/lt645/result"
	ClientTimeout      = 10 * time.Second
	RequestDelay       = 100 * time.Millisecond
	MaxRetries         = 3

	LottoCSVFileName = "lotto_draws.csv"
)
