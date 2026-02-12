package lotto

import (
	"encoding/json"
	"errors"
	"fmt"
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

// GetBayesianStats GET /api/lotto/stats/bayesian?window=50
func (h *Handler) GetBayesianStats(w http.ResponseWriter, r *http.Request) {
	windowSize := 50
	if w := r.URL.Query().Get("window"); w != "" {
		if v, err := strconv.Atoi(w); err == nil && v > 0 && v <= 500 {
			windowSize = v
		}
	}

	stats, err := h.service.GetBayesianStats(r.Context(), windowSize)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, stats)
}

// GetBayesianStatsHistory GET /api/lotto/stats/bayesian/history?number=7&limit=50
func (h *Handler) GetBayesianStatsHistory(w http.ResponseWriter, r *http.Request) {
	// number 파라미터 (필수)
	numberStr := r.URL.Query().Get("number")
	if numberStr == "" {
		h.errorResponse(w, http.StatusBadRequest, "number parameter is required")
		return
	}
	number, err := strconv.Atoi(numberStr)
	if err != nil || number < 1 || number > 45 {
		h.errorResponse(w, http.StatusBadRequest, "number must be between 1 and 45")
		return
	}

	// limit 파라미터 (선택, 기본값 50)
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 1000 {
			limit = v
		}
	}

	stats, err := h.service.GetBayesianStatsHistory(r.Context(), number, limit)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, stats)
}

// GetBayesianStatsByDrawNo GET /api/lotto/stats/bayesian/{drawNo}
func (h *Handler) GetBayesianStatsByDrawNo(w http.ResponseWriter, r *http.Request) {
	drawNoStr := r.PathValue("drawNo")
	drawNo, err := strconv.Atoi(drawNoStr)
	if err != nil || drawNo < 1 {
		h.errorResponse(w, http.StatusBadRequest, "invalid draw number")
		return
	}

	stats, err := h.service.GetBayesianStatsByDrawNo(r.Context(), drawNo)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(stats) == 0 {
		h.errorResponse(w, http.StatusNotFound, "bayesian stats not found for this draw")
		return
	}

	h.jsonResponse(w, http.StatusOK, stats)
}

// GetAnalysisStats GET /api/lotto/stats/analysis
func (h *Handler) GetAnalysisStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetAnalysisStats(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, stats)
}

// GetAnalysisStatsByDrawNo GET /api/lotto/stats/analysis/{drawNo}
func (h *Handler) GetAnalysisStatsByDrawNo(w http.ResponseWriter, r *http.Request) {
	drawNoStr := r.PathValue("drawNo")
	drawNo, err := strconv.Atoi(drawNoStr)
	if err != nil || drawNo < 1 {
		h.errorResponse(w, http.StatusBadRequest, "invalid draw number")
		return
	}

	stats, err := h.service.GetAnalysisStatsByDrawNo(r.Context(), drawNo)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(stats) == 0 {
		h.errorResponse(w, http.StatusNotFound, "analysis stats not found for this draw")
		return
	}

	h.jsonResponse(w, http.StatusOK, stats)
}

// GetAnalysisStatsHistory GET /api/lotto/stats/analysis/history?number=7&limit=50
func (h *Handler) GetAnalysisStatsHistory(w http.ResponseWriter, r *http.Request) {
	// number 파라미터 (필수)
	numberStr := r.URL.Query().Get("number")
	if numberStr == "" {
		h.errorResponse(w, http.StatusBadRequest, "number parameter is required")
		return
	}
	number, err := strconv.Atoi(numberStr)
	if err != nil || number < 1 || number > 45 {
		h.errorResponse(w, http.StatusBadRequest, "number must be between 1 and 45")
		return
	}

	// limit 파라미터 (선택, 기본값 50)
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 1000 {
			limit = v
		}
	}

	stats, err := h.service.GetAnalysisStatsHistory(r.Context(), number, limit)
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

// ========================================
// 추천 기능 핸들러
// ========================================

// GetMethods GET /api/lotto/methods
func (h *Handler) GetMethods(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.GetAnalysisMethods(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, resp)
}

// RecommendNumbers POST /api/lotto/recommend
func (h *Handler) RecommendNumbers(w http.ResponseWriter, r *http.Request) {
	var req RecommendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// 기본값 설정
	if req.Count <= 0 {
		req.Count = 1
	}
	if req.Count > 10 {
		req.Count = 10
	}

	// 유효성 검사
	if len(req.MethodCodes) == 0 {
		h.errorResponse(w, http.StatusBadRequest, "at least one method_code is required")
		return
	}
	if len(req.MethodCodes) > MaxMethodCodes {
		h.errorResponse(w, http.StatusBadRequest, fmt.Sprintf("maximum %d method_codes allowed", MaxMethodCodes))
		return
	}

	// 가중 평균 선택 시 가중치 검증
	if req.CombineCode == CombineWeightedAvg {
		if len(req.Weights) == 0 {
			h.errorResponse(w, http.StatusBadRequest, "weights are required for WEIGHTED_AVG combine method")
			return
		}
		// method_codes에 대응하는 가중치가 있는지 확인
		methodSet := make(map[string]bool)
		for _, code := range req.MethodCodes {
			methodSet[code] = true
		}
		for key, val := range req.Weights {
			if !methodSet[key] {
				h.errorResponse(w, http.StatusBadRequest, fmt.Sprintf("weight key '%s' does not match any method_code", key))
				return
			}
			if val <= 0 {
				h.errorResponse(w, http.StatusBadRequest, fmt.Sprintf("weight for '%s' must be greater than 0", key))
				return
			}
		}
	}

	// TODO: 인증된 사용자인 경우 userID 추출
	var userID *int64 = nil

	resp, err := h.service.RecommendNumbers(r.Context(), req, userID)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, resp)
}

// GetCombineMethods GET /api/lotto/combine-methods
func (h *Handler) GetCombineMethods(w http.ResponseWriter, r *http.Request) {
	resp := h.service.GetCombineMethods()
	h.jsonResponse(w, http.StatusOK, resp)
}

func (h *Handler) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) errorResponse(w http.ResponseWriter, status int, message string) {
	h.jsonResponse(w, status, map[string]string{"error": message})
}
