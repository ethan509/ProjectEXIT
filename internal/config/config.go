package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type ServerConfig struct {
	Address             string `json:"address"`
	ReadTimeoutSec      int    `json:"readTimeoutSec"`
	WriteTimeoutSec     int    `json:"writeTimeoutSec"`
	IdleTimeoutSec      int    `json:"idleTimeoutSec"`
	RequestTimeoutSec   int    `json:"requestTimeoutSec"`
	MaxRequestBodyBytes int64  `json:"maxRequestBodyBytes"`
}

type LoggingConfig struct {
	Level string `json:"level"`
	Dir   string `json:"dir"`
}

type ConcurrencyConfig struct {
	MaxConcurrentRequests int `json:"maxConcurrentRequests"`
	MainLogicWorkerCount  int `json:"mainLogicWorkerCount"`
	DBWorkerCount         int `json:"dbWorkerCount"`
	ExternalWorkerCount   int `json:"externalWorkerCount"`
	InputChannelSize      int `json:"inputChannelSize"`
	DBChannelSize         int `json:"dbChannelSize"`
	ExternalChannelSize   int `json:"externalChannelSize"`
}

type SchedulerConfig struct {
	Timezone string `json:"timezone"`
	Enabled  bool   `json:"enabled"`
}

type ConfigReloadConfig struct {
	Enabled         bool `json:"enabled"`
	IntervalMinutes int  `json:"intervalMinutes"`
}

type DatabaseConfig struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	User        string `json:"user"`
	Password    string `json:"password"`
	DBName      string `json:"dbName"`
	SSLMode     string `json:"sslMode"`
	AutoMigrate bool   `json:"autoMigrate"`
}

type JWTConfig struct {
	SecretKey               string `json:"secretKey"`
	AccessTokenExpiryMin    int    `json:"accessTokenExpiryMin"`
	RefreshTokenExpiryDays  int    `json:"refreshTokenExpiryDays"`
}

type SMTPConfig struct {
	Enabled  bool   `json:"enabled"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
}

type CrawlerConfig struct {
	BatchSize    int `json:"batchSize"`    // 배치당 크롤링할 회차 수
	BatchDelayMs int `json:"batchDelayMs"` // 배치 간 딜레이 (밀리초)
}

type Config struct {
	Server       ServerConfig       `json:"server"`
	Logging      LoggingConfig      `json:"logging"`
	Concurrency  ConcurrencyConfig  `json:"concurrency"`
	Crawler      CrawlerConfig      `json:"crawler"`
	Scheduler    SchedulerConfig    `json:"scheduler"`
	ConfigReload ConfigReloadConfig `json:"configReload"`
	Database     DatabaseConfig     `json:"database"`
	JWT          JWTConfig          `json:"jwt"`
	SMTP         SMTPConfig         `json:"smtp"`
}

type HotConfig struct {
	// hot-reloadable fields
	ReadTimeoutSec    int
	WriteTimeoutSec   int
	IdleTimeoutSec    int
	RequestTimeoutSec int
	MaxBodyBytes      int64
	LogLevel          string
}

type Configger interface {
	Config() Config
	Hot() HotConfig
	Path() string
	ReloadIfNeeded(onError func(error))
	EnsureLogDir() error
	ResolvePath(p string) string
}

type Manager struct {
	mu          sync.RWMutex
	cfg         Config
	hot         HotConfig
	path        string
	lastModTime time.Time
}

func NewManager(path string) (*Manager, error) {
	cfg, modTime, err := load(path)
	if err != nil {
		return nil, err
	}
	m := &Manager{
		cfg:         cfg,
		path:        path,
		lastModTime: modTime,
	}
	m.hot = extractHot(cfg)
	return m, nil
}

func load(path string) (Config, time.Time, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, time.Time{}, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return Config{}, time.Time{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, time.Time{}, fmt.Errorf("parse config: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return Config{}, time.Time{}, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		return cfg, time.Time{}, nil
	}
	return cfg, fi.ModTime(), nil
}

func validate(c *Config) error {
	if c.Server.Address == "" {
		return errors.New("server.address required")
	}
	if c.Server.ReadTimeoutSec <= 0 || c.Server.WriteTimeoutSec <= 0 || c.Server.IdleTimeoutSec <= 0 {
		return errors.New("server timeouts must be > 0")
	}
	if c.Server.RequestTimeoutSec <= 0 {
		return errors.New("server.requestTimeoutSec must be > 0")
	}
	if c.Server.MaxRequestBodyBytes <= 0 {
		return errors.New("server.maxRequestBodyBytes must be > 0")
	}
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Dir == "" {
		c.Logging.Dir = "logs"
	}
	if c.Concurrency.MaxConcurrentRequests < 1 {
		return errors.New("concurrency.maxConcurrentRequests must be >= 1")
	}
	if c.Concurrency.MainLogicWorkerCount < 1 ||
		c.Concurrency.DBWorkerCount < 1 ||
		c.Concurrency.ExternalWorkerCount < 1 {
		return errors.New("worker counts must be >= 1")
	}
	if c.Concurrency.InputChannelSize <= 0 {
		c.Concurrency.InputChannelSize = 1024
	}
	if c.Concurrency.DBChannelSize <= 0 {
		c.Concurrency.DBChannelSize = 256
	}
	if c.Concurrency.ExternalChannelSize <= 0 {
		c.Concurrency.ExternalChannelSize = 256
	}
	if c.Crawler.BatchSize <= 0 {
		c.Crawler.BatchSize = 10
	}
	if c.Crawler.BatchDelayMs <= 0 {
		c.Crawler.BatchDelayMs = 2000
	}
	if c.Scheduler.Timezone == "" {
		c.Scheduler.Timezone = "Asia/Seoul"
	}
	if c.ConfigReload.IntervalMinutes <= 0 {
		c.ConfigReload.IntervalMinutes = 10
	}
	return nil
}

func extractHot(c Config) HotConfig {
	return HotConfig{
		ReadTimeoutSec:    c.Server.ReadTimeoutSec,
		WriteTimeoutSec:   c.Server.WriteTimeoutSec,
		IdleTimeoutSec:    c.Server.IdleTimeoutSec,
		RequestTimeoutSec: c.Server.RequestTimeoutSec,
		MaxBodyBytes:      c.Server.MaxRequestBodyBytes,
		LogLevel:          c.Logging.Level,
	}
}

func (m *Manager) Config() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cfg
}

func (m *Manager) Hot() HotConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.hot
}

func (m *Manager) Path() string {
	return m.path
}

func (m *Manager) ReloadIfNeeded(onError func(error)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	fi, err := os.Stat(m.path)
	if err != nil {
		onError(fmt.Errorf("stat config: %w", err))
		return
	}
	if !fi.ModTime().After(m.lastModTime) {
		return
	}

	cfg, modTime, err := load(m.path)
	if err != nil {
		onError(err)
		return
	}

	// Only hot fields updated
	m.cfg.Server.ReadTimeoutSec = cfg.Server.ReadTimeoutSec
	m.cfg.Server.WriteTimeoutSec = cfg.Server.WriteTimeoutSec
	m.cfg.Server.IdleTimeoutSec = cfg.Server.IdleTimeoutSec
	m.cfg.Server.RequestTimeoutSec = cfg.Server.RequestTimeoutSec
	m.cfg.Server.MaxRequestBodyBytes = cfg.Server.MaxRequestBodyBytes
	m.cfg.Logging.Level = cfg.Logging.Level

	m.hot = extractHot(m.cfg)
	m.lastModTime = modTime
}

// Ensure log dir exists.
func (m *Manager) EnsureLogDir() error {
	dir := m.cfg.Logging.Dir
	return os.MkdirAll(dir, 0o755)
}

// ResolvePath resolves config-relative paths.
func (m *Manager) ResolvePath(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	base := filepath.Dir(m.path)
	return filepath.Join(base, "..", p)
}
