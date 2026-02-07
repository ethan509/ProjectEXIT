package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

type JWTConfig struct {
	SecretKey          string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	Issuer             string
}

type JWTManager struct {
	config JWTConfig
}

type Claims struct {
	UserID    int64  `json:"user_id"`
	LottoTier int    `json:"lotto_tier"`
	TierCode  string `json:"tier_code"`  // string으로 저장 (GUEST, MEMBER, GOLD, VIP)
	TierLevel int    `json:"tier_level"` // int로 저장 (0, 1, 2, 3)
	jwt.RegisteredClaims
}

func NewJWTManager(config JWTConfig) *JWTManager {
	if config.AccessTokenExpiry == 0 {
		config.AccessTokenExpiry = 15 * time.Minute
	}
	if config.RefreshTokenExpiry == 0 {
		config.RefreshTokenExpiry = 7 * 24 * time.Hour
	}
	if config.Issuer == "" {
		config.Issuer = "LottoSmash"
	}
	return &JWTManager{config: config}
}

func (j *JWTManager) GenerateAccessToken(user *User) (string, error) {
	now := time.Now()

	lottoTier := user.LottoTier
	tierCode := TierGuest.String()
	tierLevel := 0
	if user.Tier != nil {
		tierCode = user.Tier.Code
		tierLevel = user.Tier.Level
	}

	claims := Claims{
		UserID:    user.ID,
		LottoTier: lottoTier,
		TierCode:  tierCode,
		TierLevel: tierLevel,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.config.Issuer,
			Subject:   string(rune(user.ID)),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.config.AccessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.SecretKey))
}

func (j *JWTManager) GenerateRefreshToken() (string, time.Time, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", time.Time{}, err
	}
	token := hex.EncodeToString(bytes)
	expiresAt := time.Now().Add(j.config.RefreshTokenExpiry)
	return token, expiresAt, nil
}

func (j *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(j.config.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

func (j *JWTManager) GetAccessTokenExpiry() time.Duration {
	return j.config.AccessTokenExpiry
}

func (j *JWTManager) GetRefreshTokenExpiry() time.Duration {
	return j.config.RefreshTokenExpiry
}
