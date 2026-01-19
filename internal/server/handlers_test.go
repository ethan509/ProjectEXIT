// ping 핸들러 테스트 예제
package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/ProjectEXIT/internal/config"
	"github.com/example/ProjectEXIT/internal/logger"
	"github.com/example/ProjectEXIT/internal/worker"
)

func newTestDeps(t *testing.T) Dependencies {
	t.Helper()
	cfg := config.Config{
		Server: config.ServerConfig{
			Address:             ":0",
			ReadTimeoutSec:      5,
			WriteTimeoutSec:     5,
			IdleTimeoutSec:      30,
			RequestTimeoutSec:   2,
			MaxRequestBodyBytes: 1048576,
		},
		Logging: config.LoggingConfig{
			Level: "debug",
			Dir:   "logs-test",
		},
		Concurrency: config.ConcurrencyConfig{
			MaxConcurrentRequests: 10,
			MainLogicWorkerCount:  1,
			DBWorkerCount:         1,
			ExternalWorkerCount:   1,
			InputChannelSize:      10,
			DBChannelSize:         10,
			ExternalChannelSize:   10,
		},
		Scheduler: config.SchedulerConfig{
			Timezone: "Asia/Seoul",
			Enabled:  false,
		},
		ConfigReload: config.ConfigReloadConfig{
			Enabled:         false,
			IntervalMinutes: 10,
		},
	}
	mgr := &config.ManagerMock{Cfg: cfg}
	lg, err := logger.New("logs-test", "debug")
	if err != nil {
		t.Fatalf("logger: %v", err)
	}
	pools := &worker.Pools{
		MainInput: make(chan worker.Job, 10),
		DBInput:   make(chan worker.Job, 10),
		ExtInput:  make(chan worker.Job, 10),
	}
	return Dependencies{
		ConfigMgr: mgr,
		Logger:    lg,
		Pools:     pools,
	}
}

func TestPingHandler(t *testing.T) {
	deps := newTestDeps(t)
	h := PingHandler(deps)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body["code"] != "OK" {
		t.Fatalf("expected code OK, got %v", body["code"])
	}
}
