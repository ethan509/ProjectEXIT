package auth

import (
	"time"
)

// TierCode 회원 등급 코드 상수
type TierCode string

const (
	TierGuest  TierCode = "GUEST"  // 게스트 (회원가입 안함)
	TierMember TierCode = "MEMBER" // 정회원 (회원가입 완료)
	TierGold   TierCode = "GOLD"   // 골드 (월정액 구독)
	TierVIP    TierCode = "VIP"    // VIP (특별 등급)
)

// Gender 성별 코드
type Gender string

const (
	GenderMale   Gender = "M" // 남성
	GenderFemale Gender = "F" // 여성
	GenderOther  Gender = "O" // 기타
)

// PurchaseFrequency 로또 구매 빈도
type PurchaseFrequency string

const (
	FreqWeekly    PurchaseFrequency = "WEEKLY"    // 주1회
	FreqMonthly   PurchaseFrequency = "MONTHLY"   // 월1회
	FreqBimonthly PurchaseFrequency = "BIMONTHLY" // 월2~3회
	FreqIrregular PurchaseFrequency = "IRREGULAR" // 비정기
)

// MembershipTier 회원 등급 메타 정보
type MembershipTier struct {
	ID          int       `json:"id"`
	Code        TierCode  `json:"code"`
	Name        string    `json:"name"`
	Level       int       `json:"level"`
	Description *string   `json:"description,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type User struct {
	ID                int64              `json:"id"`
	DeviceID          *string            `json:"device_id,omitempty"`
	Email             *string            `json:"email,omitempty"`
	PasswordHash      *string            `json:"-"`
	LottoTier         int                `json:"lotto_tier"`
	Tier              *MembershipTier    `json:"tier,omitempty"`
	Gender            *Gender            `json:"gender,omitempty"`
	BirthDate         *time.Time         `json:"birth_date,omitempty"`
	Region            *string            `json:"region,omitempty"`
	Nickname          *string            `json:"nickname,omitempty"`
	PurchaseFrequency *PurchaseFrequency `json:"purchase_frequency,omitempty"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

// IsMember 정회원 이상인지 확인 (하위 호환용)
func (u *User) IsMember() bool {
	return u.LottoTier >= 2 // MEMBER 이상
}

// HasTier 특정 등급 이상인지 확인
func (u *User) HasTier(tierLevel int) bool {
	if u.Tier != nil {
		return u.Tier.Level >= tierLevel
	}
	return u.LottoTier >= tierLevel
}

type RefreshToken struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type EmailVerification struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
	Verified  bool      `json:"verified"`
	CreatedAt time.Time `json:"created_at"`
}

// Request/Response DTOs

type GuestLoginRequest struct {
	DeviceID string `json:"device_id"`
}

type EmailRegisterRequest struct {
	Email             string  `json:"email"`
	Password          string  `json:"password"`
	Code              string  `json:"code"`
	Gender            string  `json:"gender"`              // 필수: M, F, O
	BirthDate         string  `json:"birth_date"`          // 필수: YYYY-MM-DD
	Region            *string `json:"region,omitempty"`    // 옵션: 거주지역
	Nickname          *string `json:"nickname,omitempty"`  // 옵션: 닉네임 (max 20자)
	PurchaseFrequency *string `json:"purchase_frequency,omitempty"` // 옵션: WEEKLY, MONTHLY, BIMONTHLY, IRREGULAR
}

type EmailLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LinkEmailRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type SendVerificationRequest struct {
	Email string `json:"email"`
}

type VerifyEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type PasswordResetRequest struct {
	Email string `json:"email"`
}

type PasswordChangeRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type UserResponse struct {
	ID                int64        `json:"id"`
	Email             *string      `json:"email,omitempty"`
	Tier              TierResponse `json:"tier"`
	Gender            *Gender      `json:"gender,omitempty"`
	BirthDate         *string      `json:"birth_date,omitempty"`
	Region            *string      `json:"region,omitempty"`
	Nickname          *string      `json:"nickname,omitempty"`
	PurchaseFrequency *string      `json:"purchase_frequency,omitempty"`
}

type TierResponse struct {
	Code  TierCode `json:"code"`
	Name  string   `json:"name"`
	Level int      `json:"level"`
}
