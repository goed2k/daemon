package app

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/chenjia404/goed2kd/internal/config"
)

// NewLogger 根据配置构造slog（同时写文件与标准错误）。
func NewLogger(c *config.LoggingConfig) (*slog.Logger, func() error, error) {
	level := parseLevel(c.Level)
	opts := &slog.HandlerOptions{Level: level}

	var closers []io.Closer
	var writers []io.Writer
	writers = append(writers, os.Stderr)

	if path := strings.TrimSpace(c.Path); path != "" {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, nil, err
		}
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return nil, nil, err
		}
		closers = append(closers, f)
		writers = append(writers, f)
	}

	mw := io.MultiWriter(writers...)
	h := slog.NewJSONHandler(mw, opts)
	log := slog.New(h)

	closeFn := func() error {
		var err error
		for _, cl := range closers {
			if e := cl.Close(); e != nil && err == nil {
				err = e
			}
		}
		return err
	}
	return log, closeFn, nil
}

func parseLevel(s string) slog.Leveler {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
