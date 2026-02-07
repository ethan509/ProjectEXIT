package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/example/LottoSmash/internal/constants"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	LottoTierKey contextKey = "lotto_tier"
	TierCodeKey  contextKey = "tier_code"
	TierLevelKey contextKey = "tier_level"
)

type Middleware struct {
	jwt *JWTManager
}

func NewMiddleware(jwt *JWTManager) *Middleware {
	return &Middleware{jwt: jwt}
}

// RequireAuth 인증 필수 미들웨어
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := m.extractClaims(r)
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		ctx := m.setClaimsToContext(r.Context(), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireMember 정회원 이상 전용 미들웨어
func (m *Middleware) RequireMember(next http.Handler) http.Handler {
	return m.RequireTierLevel(constants.MemberLevel, next)
}

// RequireGold 골드 이상 전용 미들웨어
func (m *Middleware) RequireGold(next http.Handler) http.Handler {
	return m.RequireTierLevel(constants.GoldLevel, next)
}

// RequireVIP VIP 전용 미들웨어
func (m *Middleware) RequireVIP(next http.Handler) http.Handler {
	return m.RequireTierLevel(constants.VIPLevel, next)
}

// RequireTierLevel 특정 등급 레벨 이상 필요 미들웨어
func (m *Middleware) RequireTierLevel(requiredLevel TierLevel, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := m.extractClaims(r)
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		if claims.TierLevel < int(requiredLevel) {
			http.Error(w, `{"error":"insufficient tier level"}`, http.StatusForbidden)
			return
		}

		ctx := m.setClaimsToContext(r.Context(), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth 인증 선택적 미들웨어 (비회원도 접근 가능하지만 토큰이 있으면 파싱)
func (m *Middleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := m.extractClaims(r)
		if err == nil {
			ctx := m.setClaimsToContext(r.Context(), claims)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) extractClaims(r *http.Request) (*Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, ErrInvalidToken
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, ErrInvalidToken
	}

	return m.jwt.ValidateAccessToken(parts[1])
}

func (m *Middleware) setClaimsToContext(ctx context.Context, claims *Claims) context.Context {
	ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
	ctx = context.WithValue(ctx, LottoTierKey, claims.LottoTier)
	ctx = context.WithValue(ctx, TierCodeKey, claims.TierCode)
	ctx = context.WithValue(ctx, TierLevelKey, claims.TierLevel)
	return ctx
}

// GetUserID 컨텍스트에서 사용자 ID 추출
func GetUserID(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok
}

// GetTierLevel 컨텍스트에서 등급 레벨 추출
func GetTierLevel(ctx context.Context) (int, bool) {
	level, ok := ctx.Value(TierLevelKey).(int)
	return level, ok
}

// GetTierCode 컨텍스트에서 등급 코드 추출
func GetTierCode(ctx context.Context) (string, bool) {
	code, ok := ctx.Value(TierCodeKey).(string)
	return code, ok
}

// IsMember 컨텍스트에서 정회원 이상 여부 추출 (하위 호환)
func IsMember(ctx context.Context) bool {
	level, ok := GetTierLevel(ctx)
	return ok && level >= 1
}
