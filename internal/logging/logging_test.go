package logging

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"sync"
	"testing"

	"github.com/melanch0l/Argus/internal/config"
)

func TestLevelsAndFormats(t *testing.T) {
	for _, format := range []string{"json", "text"} {
		for _, level := range []string{"debug", "info", "warn", "error"} {
			t.Run(format+"/"+level, func(t *testing.T) {
				var out bytes.Buffer
				global := slog.Default()
				l, err := New(&out, "agent", config.Logging{Level: level, Format: format})
				if err != nil {
					t.Fatal(err)
				}
				if slog.Default() != global {
					t.Fatal("changed global logger")
				}
				l.Debug("debug-message")
				l.Info("info-message")
				l.Warn("warn-message")
				l.Error("error-message")
				text := out.String()
				if strings.Contains(text, "debug-message") != (level == "debug") {
					t.Fatal("incorrect debug filtering")
				}
				if strings.Contains(text, "info-message") != (level == "debug" || level == "info") {
					t.Fatal("incorrect info filtering")
				}
				if strings.Contains(text, "warn-message") != (level != "error") {
					t.Fatal("incorrect warn filtering")
				}
				if !strings.Contains(text, "error-message") || !strings.Contains(text, "component") || !strings.Contains(text, "agent") {
					t.Fatal("missing component/message")
				}
				if format == "json" {
					for _, line := range strings.Split(strings.TrimSpace(text), "\n") {
						var record map[string]any
						if err := json.Unmarshal([]byte(line), &record); err != nil {
							t.Fatal(err)
						}
						if record["component"] != "agent" {
							t.Fatal("bad component")
						}
					}
				}
			})
		}
	}
}
func TestRedaction(t *testing.T) {
	for _, format := range []string{"json", "text"} {
		var out bytes.Buffer
		l, err := New(&out, "backend", config.Logging{Level: "debug", Format: format})
		if err != nil {
			t.Fatal(err)
		}
		l = l.With("password", "canary-secret", slog.Group("credentials", "value", "canary-secret"))
		l.Info("safe message", "Authorization", "canary-secret", "token_file", "canary-secret", "dsn", "canary-secret", slog.Group("nested", "client_secret", "canary-secret", "safe", "visible"))
		l.WithGroup("credentials").Info("safe message", "value", "canary-secret")
		if strings.Contains(out.String(), "canary-secret") {
			t.Fatal("sensitive attribute leaked")
		}
		if !strings.Contains(out.String(), "[REDACTED]") || !strings.Contains(out.String(), "visible") {
			t.Fatal("incorrect redaction")
		}
	}
}
func TestInvalidOptions(t *testing.T) {
	var out bytes.Buffer
	for _, c := range []config.Logging{{Level: "trace", Format: "json"}, {Level: "info", Format: "yaml"}} {
		if _, err := New(&out, "agent", c); err == nil {
			t.Fatal("accepted invalid logging")
		}
	}
	if _, err := New(nil, "agent", config.Logging{Level: "info", Format: "json"}); err == nil {
		t.Fatal("accepted nil writer")
	}
	if _, err := New(&out, "other", config.Logging{Level: "info", Format: "json"}); err == nil {
		t.Fatal("accepted invalid component")
	}
}
func TestConcurrentGroupedLogging(t *testing.T) {
	var out bytes.Buffer
	l, err := New(&out, "agent", config.Logging{Level: "info", Format: "json"})
	if err != nil {
		t.Fatal(err)
	}
	l = l.WithGroup("outer").WithGroup("credentials")
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 25; j++ {
				l.Info("safe message", "value", "canary-secret")
			}
		}()
	}
	wg.Wait()
	if strings.Contains(out.String(), "canary-secret") {
		t.Fatal("concurrent redaction failed")
	}
}
