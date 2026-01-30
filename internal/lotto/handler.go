package lotto

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetDraws GET /api/lotto/draws?limit=10&offset=0
func (h *Handler) GetDraws(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.service.GetDraws(r.Context(), limit, offset)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, resp)
}

// GetDraw GET /api/lotto/draws/{drawNo}
func (h *Handler) GetDraw(w http.ResponseWriter, r *http.Request) {
	drawNoStr := chi.URLParam(r, "drawNo")
	drawNo, err := strconv.Atoi(drawNoStr)
	if err != nil || drawNo < 1 {
		h.errorResponse(w, http.StatusBadRequest, "invalid draw number")
		return
	}

	draw, err := h.service.GetDrawByNo(r.Context(), drawNo)
	if errors.Is(err, ErrDrawNotFound) {
		h.errorResponse(w, http.StatusNotFound, "draw not found")
		return
	}
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, draw)
}

// GetStats GET /api/lotto/stats
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetStats(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, stats)
}

// GetNumberStats GET /api/lotto/stats/numbers
func (h *Handler) GetNumberStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetNumberStats(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"number_stats": stats,
	})
}

// GetReappearStats GET /api/lotto/stats/reappear
func (h *Handler) GetReappearStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetReappearStats(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"reappear_stats": stats,
	})
}

// GetFirstLastStats GET /api/lotto/stats/first-last
func (h *Handler) GetFirstLastStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetFirstLastStats(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, stats)
}

// GetPairStats GET /api/lotto/stats/pairs?top=20
func (h *Handler) GetPairStats(w http.ResponseWriter, r *http.Request) {
	topN := 20
	if t := r.URL.Query().Get("top"); t != "" {
		if v, err := strconv.Atoi(t); err == nil && v > 0 && v <= 100 {
			topN = v
		}
	}

	stats, err := h.service.GetPairStats(r.Context(), topN)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, stats)
}

// GetConsecutiveStats GET /api/lotto/stats/consecutive
func (h *Handler) GetConsecutiveStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetConsecutiveStats(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, stats)
}

// GetRatioStats GET /api/lotto/stats/ratio
func (h *Handler) GetRatioStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetRatioStats(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, stats)
}

// GetColorStats GET /api/lotto/stats/colors?top=20
func (h *Handler) GetColorStats(w http.ResponseWriter, r *http.Request) {
	topN := 20
	if t := r.URL.Query().Get("top"); t != "" {
		if v, err := strconv.Atoi(t); err == nil && v > 0 && v <= 100 {
			topN = v
		}
	}

	stats, err := h.service.GetColorStats(r.Context(), topN)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, stats)
}

// GetRowColStats GET /api/lotto/stats/grid?top=20
func (h *Handler) GetRowColStats(w http.ResponseWriter, r *http.Request) {
	topN := 20
	if t := r.URL.Query().Get("top"); t != "" {
		if v, err := strconv.Atoi(t); err == nil && v > 0 && v <= 100 {
			topN = v
		}
	}

	stats, err := h.service.GetRowColStats(r.Context(), topN)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, stats)
}

// TriggerSync POST /api/admin/lotto/sync
func (h *Handler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	if err := h.service.TriggerSync(r.Context()); err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, map[string]string{
		"message": "sync completed successfully",
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
