package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/example/LottoSmash/internal/constants"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrAlreadyMember      = errors.New("user is already a member")
)

type EmailSender interface {
	SendVerificationEmail(email, code string) error
	SendPasswordResetEmail(email, code string) error
}

// ZamHistoryRecorder records Zam events for daily aggregation.
type ZamHistoryRecorder interface {
	Record(userID int64, amount int64, txType string)
}

type Service struct {
	repo               *Repository
	jwt                *JWTManager
	emailSender        EmailSender
	zamHistoryRecorder ZamHistoryRecorder
}

func NewService(repo *Repository, jwt *JWTManager, emailSender EmailSender) *Service {
	return &Service{
		repo:        repo,
		jwt:         jwt,
		emailSender: emailSender,
	}
}

// SetZamHistoryRecorder sets the optional Zam history recorder.
func (s *Service) SetZamHistoryRecorder(recorder ZamHistoryRecorder) {
	s.zamHistoryRecorder = recorder
}

func (s *Service) recordZamHistory(userID, amount int64, txType string) {
	if s.zamHistoryRecorder != nil {
		s.zamHistoryRecorder.Record(userID, amount, txType)
	}
}

// GuestLogin 비회원 로그인 (기기 ID 기반)
func (s *Service) GuestLogin(ctx context.Context, deviceID string) (*TokenResponse, error) {
	user, err := s.repo.GetUserByDeviceID(ctx, deviceID)
	isNewUser := false
	if errors.Is(err, ErrUserNotFound) {
		user, err = s.repo.CreateGuestUser(ctx, deviceID)
		if err != nil {
			return nil, fmt.Errorf("failed to create guest user: %w", err)
		}
		isNewUser = true
	} else if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 신규 가입 시 Zam 보너스 지급
	if isNewUser {
		reward := constants.TierGuest.GetZamReward()
		if err := s.repo.AddZam(ctx, user.ID, reward.RegisterBonus, string(constants.ZamTxRegisterBonus), "게스트 가입 보너스", nil); err == nil {
			s.recordZamHistory(user.ID, reward.RegisterBonus, string(constants.ZamTxRegisterBonus))
		}
	}

	// 일일 로그인 보상 지급
	s.grantDailyLoginReward(ctx, user)

	return s.generateTokens(ctx, user)
}

// EmailRegister 이메일 회원가입
func (s *Service) EmailRegister(ctx context.Context, email, password, code string, profile *UserProfileInput) (*TokenResponse, error) {
	// 이메일 중복 확인
	exists, err := s.repo.EmailExists(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailAlreadyExists
	}

	// 이메일 인증 확인
	if err := s.repo.VerifyEmail(ctx, email, code); err != nil {
		return nil, ErrEmailNotVerified
	}

	// 비밀번호 해시
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 사용자 생성
	user, err := s.repo.CreateMemberUser(ctx, email, string(hash), profile)
	if err != nil {
		return nil, err
	}

	// 회원가입 Zam 보너스 지급
	reward := constants.TierMember.GetZamReward()
	if err := s.repo.AddZam(ctx, user.ID, reward.RegisterBonus, string(constants.ZamTxRegisterBonus), "정회원 가입 보너스", nil); err == nil {
		s.recordZamHistory(user.ID, reward.RegisterBonus, string(constants.ZamTxRegisterBonus))
	}

	return s.generateTokens(ctx, user)
}

// EmailLogin 이메일 로그인
func (s *Service) EmailLogin(ctx context.Context, email, password string) (*TokenResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if errors.Is(err, ErrUserNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if user.PasswordHash == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// 일일 로그인 보상 지급
	s.grantDailyLoginReward(ctx, user)

	return s.generateTokens(ctx, user)
}

// LinkEmail 비회원 계정에 이메일 연동
func (s *Service) LinkEmail(ctx context.Context, userID int64, email, password, code string) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.IsMember() {
		return ErrAlreadyMember
	}

	// 이메일 중복 확인
	exists, err := s.repo.EmailExists(ctx, email)
	if err != nil {
		return err
	}
	if exists {
		return ErrEmailAlreadyExists
	}

	// 이메일 인증 확인
	if err := s.repo.VerifyEmail(ctx, email, code); err != nil {
		return ErrEmailNotVerified
	}

	// 비밀번호 해시
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repo.LinkEmail(ctx, userID, email, string(hash))
}

// RefreshToken 토큰 재발급
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	rt, err := s.repo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.GetUserByID(ctx, rt.UserID)
	if err != nil {
		return nil, err
	}

	// 기존 토큰 삭제
	_ = s.repo.DeleteRefreshToken(ctx, refreshToken)

	return s.generateTokens(ctx, user)
}

// Logout 로그아웃
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	return s.repo.DeleteRefreshToken(ctx, refreshToken)
}

// SendVerificationCode 이메일 인증 코드 발송
func (s *Service) SendVerificationCode(ctx context.Context, email string) error {
	code := generateCode()
	expiresAt := time.Now().Add(10 * time.Minute)

	_, err := s.repo.CreateEmailVerification(ctx, email, code, expiresAt)
	if err != nil {
		return err
	}

	if s.emailSender != nil {
		return s.emailSender.SendVerificationEmail(email, code)
	}
	return nil
}

// ChangePassword 비밀번호 변경
func (s *Service) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.PasswordHash == nil {
		return ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrInvalidCredentials
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(ctx, userID, string(hash))
}

// GetUser 사용자 정보 조회
func (s *Service) GetUser(ctx context.Context, userID int64) (*UserResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	tierResp := TierResponse{
		Code:  TierGuest.String(),
		Name:  TierGuest.Name(),
		Level: int(TierGuest),
	}
	if user.Tier != nil {
		tierResp = TierResponse{
			Code:  user.Tier.Code,
			Name:  user.Tier.Name,
			Level: user.Tier.Level,
		}
	}

	resp := &UserResponse{
		ID:         user.ID,
		Email:      user.Email,
		Tier:       tierResp,
		ZamBalance: user.ZamBalance,
		Gender:     user.Gender,
		Region:     user.Region,
		Nickname:   user.Nickname,
	}

	// birth_date 포맷팅
	if user.BirthDate != nil {
		bd := user.BirthDate.Format("2006-01-02")
		resp.BirthDate = &bd
	}

	// purchase_frequency
	if user.PurchaseFrequency != nil {
		pf := string(*user.PurchaseFrequency)
		resp.PurchaseFrequency = &pf
	}

	return resp, nil
}

// GetAllTiers 모든 등급 조회
func (s *Service) GetAllTiers(ctx context.Context) ([]MembershipTier, error) {
	return s.repo.GetAllTiers(ctx)
}

// UpdateUserTier 사용자 등급 변경
func (s *Service) UpdateUserTier(ctx context.Context, userID int64, tierLevel TierLevel) error {
	tier, err := s.repo.GetTierByCode(ctx, tierLevel.String())
	if err != nil {
		return err
	}
	return s.repo.UpdateUserTier(ctx, userID, tier.ID)
}

func (s *Service) generateTokens(ctx context.Context, user *User) (*TokenResponse, error) {
	accessToken, err := s.jwt.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, expiresAt, err := s.jwt.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	_, err = s.repo.CreateRefreshToken(ctx, user.ID, refreshToken, expiresAt)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwt.GetAccessTokenExpiry().Seconds()),
		TokenType:    "Bearer",
	}, nil
}

func generateCode() string {
	const digits = "0123456789"
	code := make([]byte, 6)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		code[i] = digits[n.Int64()]
	}
	return string(code)
}

// grantDailyLoginReward 일일 로그인 보상 지급
func (s *Service) grantDailyLoginReward(ctx context.Context, user *User) {
	// 오늘 이미 보상을 받았는지 확인
	if user.LastDailyRewardAt != nil {
		now := time.Now()
		lastReward := *user.LastDailyRewardAt
		// 같은 날이면 보상 지급하지 않음
		if lastReward.Year() == now.Year() && lastReward.YearDay() == now.YearDay() {
			return
		}
	}

	// 등급에 따른 일일 보상 지급
	tierLevel := constants.TierGuest
	if user.Tier != nil {
		tierLevel = constants.TierLevel(user.Tier.Level)
	}
	reward := tierLevel.GetZamReward()

	if err := s.repo.AddZam(ctx, user.ID, reward.DailyLogin, string(constants.ZamTxDailyLogin), "일일 로그인 보상", nil); err == nil {
		s.recordZamHistory(user.ID, reward.DailyLogin, string(constants.ZamTxDailyLogin))
	}
	_ = s.repo.UpdateLastDailyReward(ctx, user.ID)
}
