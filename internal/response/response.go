package response

import (
    "encoding/json"
    "net/http"

    "github.com/example/ProjectEXIT/internal/txid"
)

type AppError struct {
    Code       string `json:"code"`
    Message    string `json:"message"`
    HTTPStatus int    `json:"-"`
    Err        error  `json:"-"`
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return e.Err.Error()
    }
    return e.Message
}

type Envelope struct {
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    TxID    string      `json:"txId"`
}

func JSON(w http.ResponseWriter, r *http.Request, status int, code string, msg string, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    tx := txid.FromContext(r.Context())
    _ = json.NewEncoder(w).Encode(Envelope{
        Code:    code,
        Message: msg,
        Data:    data,
        TxID:    tx,
    })
}

func ErrorJSON(w http.ResponseWriter, r *http.Request, appErr *AppError) {
    if appErr == nil {
        JSON(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error", nil)
        return
    }
    JSON(w, r, appErr.HTTPStatus, appErr.Code, appErr.Message, nil)
}
