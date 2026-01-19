// # chi 라우터 + healthz/readyz/metrics + 예제 핸들러 연결
package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/example/ProjectEXIT/internal/config"
	"github.com/example/ProjectEXIT/internal/logger"
	"github.com/example/ProjectEXIT/internal/metrics"
	"github.com/example/ProjectEXIT/internal/middleware"
	"github.com/example/ProjectEXIT/internal/response"
	"github.com/example/ProjectEXIT/internal/worker"
)

type Dependencies struct {
	ConfigMgr config.Configger
	Logger    *logger.Logger
	Pools     *worker.Pools
}

func NewRouter(deps Dependencies) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.TxID())
	r.Use(middleware.Recover(deps.Logger))
	r.Use(middleware.Logging(deps.Logger))
	r.Use(middleware.ConcurrencyLimit(deps.ConfigMgr.Config().Concurrency.MaxConcurrentRequests))
	r.Use(middleware.Timeout(deps.ConfigMgr))

	// health
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, r, http.StatusOK, "OK", "alive", nil)
	})

	// readyz - simplified stub, always ready in template
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		// TODO: check DB/external dependencies
		response.JSON(w, r, http.StatusOK, "READY", "ready", nil)
	})

	r.Method(http.MethodGet, "/metrics", metrics.Handler())

	// example handlers
	r.Get("/api/v1/ping", PingHandler(deps))
	r.Post("/api/v1/echo", EchoHandler(deps))

	return r
}
