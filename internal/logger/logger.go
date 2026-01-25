package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Level int

const (
	Debug Level = iota
	Info
	Warn
	Error
	Critical
)

func ParseLevel(s string) Level {
	switch s {
	case "debug":
		return Debug
	case "info":
		return Info
	case "warn", "warning":
		return Warn
	case "error":
		return Error
	case "critical":
		return Critical
	default:
		return Info
	}
}

type Logger struct {
	mu       sync.Mutex
	dir      string
	level    Level
	loggers  map[Level]*log.Logger
	date     string
	size     map[Level]int64
	maxBytes int64
}

func New(dir string, levelStr string) (*Logger, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	l := &Logger{
		dir:      dir,
		level:    ParseLevel(levelStr),
		loggers:  make(map[Level]*log.Logger),
		size:     make(map[Level]int64),
		maxBytes: 1 << 30, // 1GB
	}
	l.date = time.Now().Format("20060102")
	return l, nil
}

func (l *Logger) SetLevel(levelStr string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = ParseLevel(levelStr)
}

func (l *Logger) logf(level Level, format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	logger := l.ensureLogger(level)
	msg := fmt.Sprintf(format, args...)
	logger.Println(msg)
	l.size[level] += int64(len(msg)) + 1
	if l.size[level] >= l.maxBytes {
		l.rotate(level)
	}
}

func (l *Logger) ensureLogger(level Level) *log.Logger {
	today := time.Now().Format("20060102")
	if today != l.date {
		// day changed, reset
		l.closeAll()
		l.date = today
		l.size = make(map[Level]int64)
	}

	if lg, ok := l.loggers[level]; ok {
		return lg
	}

	fname := l.filename(level, 0)
	f, err := os.OpenFile(fname, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		// fallback to stdout
		return log.Default()
	}
	fi, _ := f.Stat()
	l.size[level] = fi.Size()
	lg := log.New(f, "", log.LstdFlags|log.Lmicroseconds)
	l.loggers[level] = lg
	return lg
}

func (l *Logger) rotate(level Level) {
	l.closeLevel(level)
	// find next index
	idx := 1
	for {
		fname := l.filename(level, idx)
		if _, err := os.Stat(fname); os.IsNotExist(err) {
			break
		}
		idx++
	}
	// rename current base to _idx
	base := l.filename(level, 0)
	_ = os.Rename(base, l.filename(level, idx))
	// reopen
	l.loggers[level] = nil
	l.size[level] = 0
	_ = l.ensureLogger(level)
}

func (l *Logger) filename(level Level, idx int) string {
	suffix := ""
	switch level {
	case Debug:
		suffix = "debug"
	case Info:
		suffix = "info"
	case Warn:
		suffix = "warn"
	case Error:
		suffix = "error"
	case Critical:
		suffix = "critical"
	}
	name := fmt.Sprintf("%s.%s.log", l.date, suffix)
	if idx > 0 {
		name = fmt.Sprintf("%s_%d", name, idx)
	}
	return filepath.Join(l.dir, name)
}

func (l *Logger) closeLevel(level Level) {
	if lg, ok := l.loggers[level]; ok {
		// underlying writer might be *os.File
		if out, ok := lg.Writer().(*os.File); ok {
			_ = out.Close()
		}
		delete(l.loggers, level)
	}
}

func (l *Logger) closeAll() {
	for lvl := range l.loggers {
		l.closeLevel(lvl)
	}
}

func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.closeAll()
}

func (l *Logger) Debugf(format string, args ...any)    { l.logf(Debug, format, args...) }
func (l *Logger) Infof(format string, args ...any)     { l.logf(Info, format, args...) }
func (l *Logger) Warnf(format string, args ...any)     { l.logf(Warn, format, args...) }
func (l *Logger) Errorf(format string, args ...any)    { l.logf(Error, format, args...) }
func (l *Logger) Criticalf(format string, args ...any) { l.logf(Critical, format, args...) }
