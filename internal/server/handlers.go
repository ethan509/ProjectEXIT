// # /api/v1/ping, /api/v1/echo 예제
package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/example/LottoSmash/internal/response"
	"github.com/example/LottoSmash/internal/txid"
	"github.com/example/LottoSmash/internal/worker"
)

type pingResponse struct {
	Now time.Time `json:"now"`
}

func PingHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, r, http.StatusOK, "OK", "pong", pingResponse{Now: time.Now()})
	}
}

type echoRequest struct {
	Message string `json:"message"`
}

func EchoHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req echoRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.ErrorJSON(w, r, &response.AppError{
				Code:       "BAD_REQUEST",
				Message:    "invalid json",
				HTTPStatus: http.StatusBadRequest,
				Err:        err,
			})
			return
		}

		// enforce max body size at handler-level if needed; main limit is via server

		tx := txid.FromContext(r.Context())
		resCh := make(chan worker.Result, 1)
		job := worker.Job{
			Type:   worker.JobTypeExample,
			TxID:   tx,
			Ctx:    r.Context(),
			Input:  req,
			Result: resCh,
		}

		select {
		case deps.Pools.MainInput <- job:
		default:
			response.ErrorJSON(w, r, &response.AppError{
				Code:       "BACKPRESSURE",
				Message:    "server busy",
				HTTPStatus: http.StatusServiceUnavailable,
			})
			return
		}

		select {
		case res := <-resCh:
			if res.Err != nil {
				response.ErrorJSON(w, r, &response.AppError{
					Code:       "INTERNAL_ERROR",
					Message:    "internal error",
					HTTPStatus: http.StatusInternalServerError,
					Err:        res.Err,
				})
				return
			}
			response.JSON(w, r, http.StatusOK, "OK", "echo", res.Data)
		case <-r.Context().Done():
			response.ErrorJSON(w, r, &response.AppError{
				Code:       "REQUEST_TIMEOUT",
				Message:    "request timeout",
				HTTPStatus: http.StatusGatewayTimeout,
				Err:        r.Context().Err(),
			})
		}
	}
}
