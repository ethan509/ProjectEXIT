package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/example/LottoSmash/internal/config"
	"github.com/example/LottoSmash/internal/database"
	"github.com/example/LottoSmash/internal/logger"
	"github.com/example/LottoSmash/internal/lotto"
	"github.com/example/LottoSmash/internal/scheduler"
	"github.com/example/LottoSmash/internal/server"
	"github.com/example/LottoSmash/internal/worker"
)

func main() {
	cfgPath := os.Getenv("LOTTOMASH_CONFIG")
	if cfgPath == "" {
		cfgPath = "config/config.json"
	}

	cfgMgr, err := config.NewManager(cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if err := cfgMgr.EnsureLogDir(); err != nil {
		log.Fatalf("failed to create log dir: %v", err)
	}

	lg, err := logger.New(cfgMgr.Config().Logging.Dir, cfgMgr.Config().Logging.Level)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer lg.Close()

	lg.Infof("LottoSmash starting with config %s", cfgPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// config reload goroutine
	if cfgMgr.Config().ConfigReload.Enabled {
		go func() {
			interval := time.Duration(cfgMgr.Config().ConfigReload.IntervalMinutes) * time.Minute
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					cfgMgr.ReloadIfNeeded(func(err error) {
						lg.Errorf("config reload failed: %v", err)
					})
					lg.SetLevel(cfgMgr.Hot().LogLevel)
				}
			}
		}()
	}

	// worker pools
	pools := &worker.Pools{
		MainInput: make(chan worker.Job, cfgMgr.Config().Concurrency.InputChannelSize),
		DBInput:   make(chan worker.Job, cfgMgr.Config().Concurrency.DBChannelSize),
		ExtInput:  make(chan worker.Job, cfgMgr.Config().Concurrency.ExternalChannelSize),
	}
	worker.StartMainWorkers(ctx, cfgMgr.Config().Concurrency.MainLogicWorkerCount, pools, lg)
	worker.StartDBWorkers(ctx, cfgMgr.Config().Concurrency.DBWorkerCount, pools, lg)
	worker.StartExternalWorkers(ctx, cfgMgr.Config().Concurrency.ExternalWorkerCount, pools, lg)

	// database connection
	dbCfg := cfgMgr.Config().Database

	// Docker 환경변수 오버라이드 (설정 파일보다 우선)
	if host := os.Getenv("DB_HOST"); host != "" {
		dbCfg.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			dbCfg.Port = p
		}
	}
	if user := os.Getenv("DB_USER"); user != "" {
		dbCfg.User = user
	}
	if pass := os.Getenv("DB_PASSWORD"); pass != "" {
		dbCfg.Password = pass
	}

	db, err := database.New(database.Config{
		Host:     dbCfg.Host,
		Port:     dbCfg.Port,
		User:     dbCfg.User,
		Password: dbCfg.Password,
		DBName:   dbCfg.DBName,
		SSLMode:  dbCfg.SSLMode,
	})
	if err != nil {
		lg.Errorf("failed to connect to database: %v", err)
	} else {
		defer db.Close()
		lg.Infof("connected to database")
	}

	// lotto service initialization
	var lottoSvc *lotto.Service
	if db != nil {
		lottoRepo := lotto.NewRepository(db)
		lottoClient := lotto.NewClient()
		lottoAnalyzer := lotto.NewAnalyzer(lottoRepo, lg)
		lottoSvc = lotto.NewService(lottoRepo, lottoClient, lottoAnalyzer, lg)

		// 서버 시작 시 당첨번호 초기화
		docsPath := cfgMgr.ResolvePath("docs")
		if err := lottoSvc.InitializeDraws(ctx, docsPath); err != nil {
			lg.Errorf("failed to initialize lotto draws: %v", err)
		}
	}

	// scheduler (DB 연결 후 시작)
	if cfgMgr.Config().Scheduler.Enabled {
		sched, err := scheduler.New(cfgMgr.Config(), lg, lottoSvc)
		if err != nil {
			lg.Errorf("failed to init scheduler: %v", err)
		} else {
			sched.Start(ctx)
		}
	}

	deps := server.Dependencies{
		ConfigMgr: cfgMgr,
		Logger:    lg,
		Pools:     pools,
		DB:        db,
		LottoSvc:  lottoSvc,
	}
	router := server.NewRouter(deps)

	hot := cfgMgr.Hot()
	srv := &http.Server{
		Addr:         cfgMgr.Config().Server.Address,
		Handler:      http.MaxBytesHandler(router, hot.MaxBodyBytes),
		ReadTimeout:  time.Duration(hot.ReadTimeoutSec) * time.Second,
		WriteTimeout: time.Duration(hot.WriteTimeoutSec) * time.Second,
		IdleTimeout:  time.Duration(hot.IdleTimeoutSec) * time.Second,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	// signal handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		lg.Infof("HTTP server listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			lg.Criticalf("HTTP server error: %v", err)
			cancel()
		}
	}()

	<-stop
	lg.Infof("shutdown signal received")
	cancel()

	// graceful shutdown (max 1 minute)
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Minute)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		lg.Errorf("server shutdown error: %v", err)
		_ = srv.Close()
	}
	lg.Infof("LottoSmash stopped")
}
