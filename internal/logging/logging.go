// Package logging creates component-scoped structured loggers. Callers must not
// log configurations, credentials, arbitrary input, or secret-file contents.
package logging

import (
	"errors"
	"io"
	"log/slog"
	"strings"

	"github.com/melanch0l/Argus/internal/config"
)

// New does not modify the process-global default logger. Redaction is a
// defense-in-depth key filter, not a way to make arbitrary messages safe.
func New(w io.Writer, component string, c config.Logging) (*slog.Logger, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	if w == nil {
		return nil, errors.New("logging output is required")
	}
	if component != "agent" && component != "backend" {
		return nil, errors.New("unsupported logging component")
	}
	level := slog.LevelInfo
	switch c.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	opts := &slog.HandlerOptions{Level: level, ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		if sensitive(a.Key) {
			return slog.String(a.Key, "[REDACTED]")
		}
		for _, key := range groups {
			if sensitive(key) {
				return slog.String(a.Key, "[REDACTED]")
			}
		}
		return a
	}}
	var h slog.Handler
	if c.Format == "json" {
		h = slog.NewJSONHandler(w, opts)
	} else {
		h = slog.NewTextHandler(w, opts)
	}
	return slog.New(h).With("component", component), nil
}

func sensitive(key string) bool {
	key = strings.ToLower(key)
	for _, part := range []string{"password", "passwd", "secret", "token", "authorization", "credential", "cookie", "dsn", "private_key", "privatekey"} {
		if strings.Contains(key, part) {
			return true
		}
	}
	return false
}
