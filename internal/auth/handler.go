package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GuestLogin POST /api/auth/guest
func (h *Handler) GuestLogin(w http.ResponseWriter, r *http.Request) {
	var req GuestLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.DeviceID == "" {
		h.errorResponse(w, http.StatusBadRequest, "device_id is required")
		return
	}

	tokens, err := h.service.GuestLogin(r.Context(), req.DeviceID)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, tokens)
}

// EmailRegister POST /api/auth/register
func (h *Handler) EmailRegister(w http.ResponseWriter, r *http.Request) {
	var req EmailRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// 필수 필드 검증
	if req.Email == "" || req.Password == "" || req.Code == "" {
		h.errorResponse(w, http.StatusBadRequest, "email, password and code are required")
		return
	}

	// 필수 프로필 필드 검증
	if req.Gender == "" {
		h.errorResponse(w, http.StatusBadRequest, "gender is required")
		return
	}
	if req.BirthDate == "" {
		h.errorResponse(w, http.StatusBadRequest, "birth_date is required")
		return
	}

	// 성별 검증
	gender := Gender(req.Gender)
	if gender != GenderMale && gender != GenderFemale && gender != GenderOther {
		h.errorResponse(w, http.StatusBadRequest, "gender must be M, F, or O")
		return
	}

	// 생년월일 파싱
	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "birth_date must be in YYYY-MM-DD format")
		return
	}

	// 구매빈도 검증 (옵션)
	var purchaseFreq *PurchaseFrequency
	if req.PurchaseFrequency != nil && *req.PurchaseFrequency != "" {
		pf := PurchaseFrequency(*req.PurchaseFrequency)
		if pf != FreqWeekly && pf != FreqMonthly && pf != FreqBimonthly && pf != FreqIrregular {
			h.errorResponse(w, http.StatusBadRequest, "purchase_frequency must be WEEKLY, MONTHLY, BIMONTHLY, or IRREGULAR")
			return
		}
		purchaseFreq = &pf
	}

	// 닉네임 길이 검증 (옵션)
	if req.Nickname != nil && len(*req.Nickname) > 20 {
		h.errorResponse(w, http.StatusBadRequest, "nickname must be 20 characters or less")
		return
	}

	profile := &UserProfileInput{
		Gender:            gender,
		BirthDate:         birthDate,
		Region:            req.Region,
		Nickname:          req.Nickname,
		PurchaseFrequency: purchaseFreq,
	}

	tokens, err := h.service.EmailRegister(r.Context(), req.Email, req.Password, req.Code, profile)
	if errors.Is(err, ErrEmailAlreadyExists) {
		h.errorResponse(w, http.StatusConflict, "email already exists")
		return
	}
	if errors.Is(err, ErrEmailNotVerified) {
		h.errorResponse(w, http.StatusBadRequest, "email not verified")
		return
	}
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusCreated, tokens)
}

// EmailLogin POST /api/auth/login
func (h *Handler) EmailLogin(w http.ResponseWriter, r *http.Request) {
	var req EmailLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		h.errorResponse(w, http.StatusBadRequest, "email and password are required")
		return
	}

	tokens, err := h.service.EmailLogin(r.Context(), req.Email, req.Password)
	if errors.Is(err, ErrInvalidCredentials) {
		h.errorResponse(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, tokens)
}

// LinkEmail POST /api/auth/link-email
func (h *Handler) LinkEmail(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)

	var req LinkEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.Code == "" {
		h.errorResponse(w, http.StatusBadRequest, "email, password and code are required")
		return
	}

	err := h.service.LinkEmail(r.Context(), userID, req.Email, req.Password, req.Code)
	if errors.Is(err, ErrAlreadyMember) {
		h.errorResponse(w, http.StatusBadRequest, "user is already a member")
		return
	}
	if errors.Is(err, ErrEmailAlreadyExists) {
		h.errorResponse(w, http.StatusConflict, "email already exists")
		return
	}
	if errors.Is(err, ErrEmailNotVerified) {
		h.errorResponse(w, http.StatusBadRequest, "email not verified")
		return
	}
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, map[string]string{"message": "email linked successfully"})
}

// RefreshToken POST /api/auth/refresh
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		h.errorResponse(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	tokens, err := h.service.RefreshToken(r.Context(), req.RefreshToken)
	if errors.Is(err, ErrTokenNotFound) || errors.Is(err, ErrTokenExpired) {
		h.errorResponse(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, tokens)
}

// Logout POST /api/auth/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		h.errorResponse(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	_ = h.service.Logout(r.Context(), req.RefreshToken)
	h.jsonResponse(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

// SendVerificationCode POST /api/auth/send-code
func (h *Handler) SendVerificationCode(w http.ResponseWriter, r *http.Request) {
	var req SendVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" {
		h.errorResponse(w, http.StatusBadRequest, "email is required")
		return
	}

	err := h.service.SendVerificationCode(r.Context(), req.Email)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, map[string]string{"message": "verification code sent"})
}

// ChangePassword POST /api/auth/change-password
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)

	var req PasswordChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		h.errorResponse(w, http.StatusBadRequest, "old_password and new_password are required")
		return
	}

	err := h.service.ChangePassword(r.Context(), userID, req.OldPassword, req.NewPassword)
	if errors.Is(err, ErrInvalidCredentials) {
		h.errorResponse(w, http.StatusUnauthorized, "invalid old password")
		return
	}
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, map[string]string{"message": "password changed successfully"})
}

// GetMe GET /api/auth/me
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)

	user, err := h.service.GetUser(r.Context(), userID)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, user)
}

func (h *Handler) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) errorResponse(w http.ResponseWriter, status int, message string) {
	h.jsonResponse(w, status, map[string]string{"error": message})
}
