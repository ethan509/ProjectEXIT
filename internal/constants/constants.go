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

// TierLevel 회원 등급 (int와 string 동시 지원)
type TierLevel int

const (
	TierGuest  TierLevel = iota // 0 - 게스트
	TierMember                  // 1 - 정회원
	TierGold                    // 2 - 골드
	TierVIP                     // 3 - VIP
)

// String은 등급의 문자열 코드를 반환
func (t TierLevel) String() string {
	switch t {
	case TierGuest:
		return "GUEST"
	case TierMember:
		return "MEMBER"
	case TierGold:
		return "GOLD"
	case TierVIP:
		return "VIP"
	default:
		return "UNKNOWN"
	}
}

// Name은 한글 이름을 반환
func (t TierLevel) Name() string {
	switch t {
	case TierGuest:
		return "게스트"
	case TierMember:
		return "정회원"
	case TierGold:
		return "골드"
	case TierVIP:
		return "VIP"
	default:
		return "알수없음"
	}
}

// ParseTierLevel은 문자열을 TierLevel로 변환
func ParseTierLevel(s string) TierLevel {
	switch s {
	case "GUEST":
		return TierGuest
	case "MEMBER":
		return TierMember
	case "GOLD":
		return TierGold
	case "VIP":
		return TierVIP
	default:
		return TierGuest
	}
}

// 하위 호환용 별칭
const (
	GuestLevel  = TierGuest
	MemberLevel = TierMember
	GoldLevel   = TierGold
	VIPLevel    = TierVIP
)

// Zam 경제 시스템 상수
type ZamTxType string

const (
	ZamTxRegisterBonus ZamTxType = "REGISTER_BONUS" // 회원가입 보너스
	ZamTxDailyLogin    ZamTxType = "DAILY_LOGIN"    // 일일 로그인 보상
	ZamTxPurchase      ZamTxType = "PURCHASE"       // 구매 (차감)
	ZamTxRefund        ZamTxType = "REFUND"         // 환불
	ZamTxReward        ZamTxType = "REWARD"         // 보상
	ZamTxAdmin         ZamTxType = "ADMIN"          // 관리자 조정
)

// ZamReward 등급별 Zam 보상
type ZamReward struct {
	RegisterBonus int64 // 회원가입 보너스
	DailyLogin    int64 // 일일 로그인 보상
}

// 등급별 Zam 보상 설정
var ZamRewards = map[TierLevel]ZamReward{
	TierGuest:  {RegisterBonus: 100, DailyLogin: 10},
	TierMember: {RegisterBonus: 200, DailyLogin: 20},
	TierGold:   {RegisterBonus: 2000, DailyLogin: 200},
	TierVIP:    {RegisterBonus: 5000, DailyLogin: 500},
}

// GetZamReward 등급에 따른 Zam 보상 조회
func (t TierLevel) GetZamReward() ZamReward {
	if reward, ok := ZamRewards[t]; ok {
		return reward
	}
	return ZamRewards[TierGuest]
}
