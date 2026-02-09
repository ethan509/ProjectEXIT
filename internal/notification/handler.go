package notification

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterDeviceToken POST /api/notifications/device-token
func (h *Handler) RegisterDeviceToken(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		h.errorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req RegisterTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Token == "" {
		h.errorResponse(w, http.StatusBadRequest, "token is required")
		return
	}
	if req.Platform == "" {
		req.Platform = "android"
	}

	if err := h.service.RegisterDeviceToken(r.Context(), userID, req.Token, req.Platform); err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "failed to register device token")
		return
	}

	h.jsonResponse(w, http.StatusOK, map[string]string{"message": "device token registered"})
}

// DeleteDeviceToken DELETE /api/notifications/device-token
func (h *Handler) DeleteDeviceToken(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		h.errorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req DeleteTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Token == "" {
		h.errorResponse(w, http.StatusBadRequest, "token is required")
		return
	}

	if err := h.service.DeactivateDeviceToken(r.Context(), userID, req.Token); err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "failed to delete device token")
		return
	}

	h.jsonResponse(w, http.StatusOK, map[string]string{"message": "device token deleted"})
}

// GetNotifications GET /api/notifications?limit=20&offset=0
func (h *Handler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		h.errorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit, offset := parsePagination(r)

	notifications, totalCount, err := h.service.GetNotifications(r.Context(), userID, limit, offset)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "failed to get notifications")
		return
	}

	h.jsonResponse(w, http.StatusOK, NotificationListResponse{
		Notifications: notifications,
		TotalCount:    totalCount,
	})
}

// GetWinnings GET /api/notifications/winnings?limit=20&offset=0
func (h *Handler) GetWinnings(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		h.errorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit, offset := parsePagination(r)

	winnings, totalCount, err := h.service.GetWinnings(r.Context(), userID, limit, offset)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "failed to get winnings")
		return
	}

	h.jsonResponse(w, http.StatusOK, WinningListResponse{
		Winnings:   winnings,
		TotalCount: totalCount,
	})
}

func (h *Handler) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) errorResponse(w http.ResponseWriter, status int, message string) {
	h.jsonResponse(w, status, map[string]string{"error": message})
}

func getUserID(r *http.Request) (int64, bool) {
	val := r.Context().Value("user_id")
	if val == nil {
		return 0, false
	}
	userID, ok := val.(int64)
	return userID, ok
}

func parsePagination(r *http.Request) (int, int) {
	limit := 20
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	return limit, offset
}
