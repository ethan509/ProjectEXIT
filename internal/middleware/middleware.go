package middleware

import (
    "context"
    "log"
    "net/http"
    "time"

    "github.com/example/LottoSmash/internal/config"
    "github.com/example/LottoSmash/internal/logger"
    "github.com/example/LottoSmash/internal/metrics"
    "github.com/example/LottoSmash/internal/response"
    "github.com/example/LottoSmash/internal/txid"
)

type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, mws ...Middleware) http.Handler {
    for i := len(mws) - 1; i >= 0; i-- {
        h = mws[i](h)
    }
    return h
}

func TxID() Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            id := r.Header.Get("X-Request-Id")
            if id == "" {
                id = txid.NewID()
            }
            ctx := txid.WithTxID(r.Context(), id)
            r = r.WithContext(ctx)
            w.Header().Set("X-Request-Id", id)
            next.ServeHTTP(w, r)
        })
    }
}

type loggingResponseWriter struct {
    http.ResponseWriter
    status int
    bytes  int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
    lrw.status = code
    lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
    if lrw.status == 0 {
        lrw.status = http.StatusOK
    }
    n, err := lrw.ResponseWriter.Write(b)
    lrw.bytes += n
    return n, err
}

func Logging(l *logger.Logger) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            tx := txid.FromContext(r.Context())
            lrw := &loggingResponseWriter{ResponseWriter: w}
            next.ServeHTTP(lrw, r)
            dur := time.Since(start)
            l.Infof("tx=%s method=%s path=%s status=%d bytes=%d duration=%s",
                tx, r.Method, r.URL.Path, lrw.status, lrw.bytes, dur)
            metrics.ObserveRequest(dur)
        })
    }
}

func Recover(l *logger.Logger) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if rec := recover(); rec != nil {
                    tx := txid.FromContext(r.Context())
                    l.Criticalf("panic tx=%s: %v", tx, rec)
                    response.JSON(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error", nil)
                }
            }()
            next.ServeHTTP(w, r)
        })
    }
}

func ConcurrencyLimit(max int) Middleware {
    sem := make(chan struct{}, max)
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            select {
            case sem <- struct{}{}:
                defer func() { <-sem }()
                metrics.IncConcurrent()
                defer metrics.DecConcurrent()
                next.ServeHTTP(w, r)
            default:
                metrics.IncRejected()
                response.JSON(w, r, http.StatusServiceUnavailable,
                    "CONCURRENCY_LIMIT_EXCEEDED", "too many concurrent requests", nil)
            }
        })
    }
}

func Timeout(cfg config.Configger) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            hot := cfg.Hot()
            timeout := time.Duration(hot.RequestTimeoutSec) * time.Second
            ctx, cancel := context.WithTimeout(r.Context(), timeout)
            defer cancel()

            r = r.WithContext(ctx)

            ch := make(chan struct{})
            go func() {
                next.ServeHTTP(w, r)
                close(ch)
            }()

            select {
            case <-ch:
                return
            case <-ctx.Done():
                log.Printf("request timeout: %v", ctx.Err())
                response.JSON(w, r, http.StatusGatewayTimeout, "REQUEST_TIMEOUT", "request timeout", nil)
            }
        })
    }
}
