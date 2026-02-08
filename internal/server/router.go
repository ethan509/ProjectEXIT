// # chi 라우터 + healthz/readyz/metrics + 예제 핸들러 연결
package server

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/example/LottoSmash/internal/auth"
	"github.com/example/LottoSmash/internal/config"
	"github.com/example/LottoSmash/internal/logger"
	"github.com/example/LottoSmash/internal/lotto"
	"github.com/example/LottoSmash/internal/metrics"
	"github.com/example/LottoSmash/internal/middleware"
	"github.com/example/LottoSmash/internal/response"
	"github.com/example/LottoSmash/internal/worker"
	"github.com/example/LottoSmash/internal/zamhistory"
)

type Dependencies struct {
	ConfigMgr        config.Configger
	Logger           *logger.Logger
	Pools            *worker.Pools
	DB               *sql.DB
	LottoSvc         *lotto.Service
	ZamHistoryBuffer *zamhistory.Buffer
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

	// auth setup
	if deps.DB != nil {
		authHandler := setupAuth(deps)
		authMiddleware := setupAuthMiddleware(deps)

		// auth routes
		r.Route("/api/auth", func(r chi.Router) {
			// public routes
			r.Post("/guest", authHandler.GuestLogin)
			r.Post("/register", authHandler.EmailRegister)
			r.Post("/login", authHandler.EmailLogin)
			r.Post("/refresh", authHandler.RefreshToken)
			r.Post("/logout", authHandler.Logout)
			r.Post("/send-code", authHandler.SendVerificationCode)

			// protected routes
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.RequireAuth)
				r.Get("/me", authHandler.GetMe)
				r.Post("/link-email", authHandler.LinkEmail)
				r.Post("/change-password", authHandler.ChangePassword)
			})
		})

		// lotto routes
		if deps.LottoSvc != nil {
			lottoHandler := lotto.NewHandler(deps.LottoSvc)

			// public lotto routes
			r.Route("/api/lotto", func(r chi.Router) {
				r.Get("/draws", lottoHandler.GetDraws)
				r.Get("/draws/{drawNo}", lottoHandler.GetDraw)
				r.Get("/stats", lottoHandler.GetStats)
				r.Get("/stats/numbers", lottoHandler.GetNumberStats)
				r.Get("/stats/reappear", lottoHandler.GetReappearStats)
				r.Get("/stats/first-last", lottoHandler.GetFirstLastStats)
				r.Get("/stats/pairs", lottoHandler.GetPairStats)
				r.Get("/stats/consecutive", lottoHandler.GetConsecutiveStats)
				r.Get("/stats/ratio", lottoHandler.GetRatioStats)
				r.Get("/stats/colors", lottoHandler.GetColorStats)
				r.Get("/stats/grid", lottoHandler.GetRowColStats)
				r.Get("/stats/bayesian", lottoHandler.GetBayesianStats)
				r.Get("/stats/bayesian/history", lottoHandler.GetBayesianStatsHistory)
				r.Get("/stats/bayesian/{drawNo}", lottoHandler.GetBayesianStatsByDrawNo)
				r.Get("/stats/analysis", lottoHandler.GetAnalysisStats)
				r.Get("/stats/analysis/history", lottoHandler.GetAnalysisStatsHistory)
				r.Get("/stats/analysis/{drawNo}", lottoHandler.GetAnalysisStatsByDrawNo)

				// 추천 기능
				r.Get("/methods", lottoHandler.GetMethods)
				r.Post("/recommend", lottoHandler.RecommendNumbers)
			})

			// admin lotto routes (protected)
			r.Route("/api/admin/lotto", func(r chi.Router) {
				r.Use(authMiddleware.RequireAuth)
				r.Post("/sync", lottoHandler.TriggerSync)
			})
		}
	}

	return r
}

func setupAuth(deps Dependencies) *auth.Handler {
	cfg := deps.ConfigMgr.Config()

	jwtConfig := auth.JWTConfig{
		SecretKey:          cfg.JWT.SecretKey,
		AccessTokenExpiry:  time.Duration(cfg.JWT.AccessTokenExpiryMin) * time.Minute,
		RefreshTokenExpiry: time.Duration(cfg.JWT.RefreshTokenExpiryDays) * 24 * time.Hour,
	}
	jwtManager := auth.NewJWTManager(jwtConfig)

	repo := auth.NewRepository(deps.DB)

	var emailSender auth.EmailSender
	if cfg.SMTP.Enabled {
		emailSender = auth.NewSMTPEmailSender(auth.SMTPConfig{
			Host:     cfg.SMTP.Host,
			Port:     cfg.SMTP.Port,
			Username: cfg.SMTP.Username,
			Password: cfg.SMTP.Password,
			From:     cfg.SMTP.From,
		})
	} else {
		emailSender = auth.NewNoopEmailSender()
	}

	service := auth.NewService(repo, jwtManager, emailSender)
	if deps.ZamHistoryBuffer != nil {
		service.SetZamHistoryRecorder(deps.ZamHistoryBuffer)
	}
	return auth.NewHandler(service)
}

func setupAuthMiddleware(deps Dependencies) *auth.Middleware {
	cfg := deps.ConfigMgr.Config()

	jwtConfig := auth.JWTConfig{
		SecretKey:          cfg.JWT.SecretKey,
		AccessTokenExpiry:  time.Duration(cfg.JWT.AccessTokenExpiryMin) * time.Minute,
		RefreshTokenExpiry: time.Duration(cfg.JWT.RefreshTokenExpiryDays) * 24 * time.Hour,
	}
	jwtManager := auth.NewJWTManager(jwtConfig)

	return auth.NewMiddleware(jwtManager)
}
