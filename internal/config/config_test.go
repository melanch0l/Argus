package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func validAgent() Agent {
	c := AgentDefaults()
	c.Version = Version
	c.BackendURL = "https://backend.example.invalid/api/v1"
	c.TokenFile = "/etc/argus/agent-token"
	c.StateDir = "/var/lib/argus/agent"
	c.DockerSocket = "/var/run/docker.sock"
	return c
}
func validBackend() Backend {
	c := BackendDefaults()
	c.Version = Version
	c.TokenFile = "/etc/argus/backend-token"
	c.PostgreSQL.Host = "postgres.example.invalid"
	c.PostgreSQL.Database = "argus"
	c.PostgreSQL.User = "argus_app"
	c.PostgreSQL.PasswordFile = "/etc/argus/postgres-password"
	return c
}
func encoded(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
func TestStrictJSON(t *testing.T) {
	good := encoded(t, validAgent())
	tests := map[string]string{
		"empty": "", "array": "[]", "null": "null", "truncated": "{", "malformed": "{secret=canary}",
		"unknown":            strings.Replace(good, "{", `{"canary-secret":"canary-secret",`, 1),
		"duplicate":          strings.Replace(good, `"version":1`, `"version":1,"version":1`, 1),
		"escaped duplicate":  strings.Replace(good, `"version":1`, `"version":1,"\u0076ersion":1`, 1),
		"nested duplicate":   strings.Replace(good, `"level":"info"`, `"level":"info","level":"debug"`, 1),
		"nested unknown":     strings.Replace(good, `"level":"info"`, `"canary-secret":"canary-secret"`, 1),
		"case alias":         strings.Replace(good, `"version"`, `"Version"`, 1),
		"null object":        strings.Replace(good, `{"level":"info","format":"json"}`, `null`, 1),
		"null scalar":        strings.Replace(good, `"version":1`, `"version":null`, 1),
		"null logging field": strings.Replace(good, `"level":"info"`, `"level":null`, 1),
		"wrong object type":  strings.Replace(good, `{"level":"info","format":"json"}`, `"canary-secret"`, 1),
		"wrong type":         strings.Replace(good, `"version":1`, `"version":"canary-secret"`, 1),
		"number overflow":    strings.Replace(good, `"version":1`, `"version":9999999999999999999999999`, 1),
		"fraction":           strings.Replace(good, `"version":1`, `"version":1.1`, 1),
		"version":            strings.Replace(good, `"version":1`, `"version":2`, 1),
		"missing version":    strings.Replace(good, `"version":1,`, ``, 1),
		"multiple":           good + good, "trailing null": good + " null", "trailing junk": good + " canary-secret",
		"oversize": strings.Repeat(" ", MaxBytes+1), "invalid utf8": good + string([]byte{255}),
		"deep": `{"logging":` + strings.Repeat("[", 10000),
	}
	for name, text := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := ParseAgent(strings.NewReader(text))
			if err == nil {
				t.Fatal("accepted invalid config")
			}
			if strings.Contains(err.Error(), "canary-secret") {
				t.Fatal("error leaked input")
			}
		})
	}
	padded := good + strings.Repeat(" ", MaxBytes-len(good))
	if _, err := ParseAgent(strings.NewReader(padded)); err != nil {
		t.Fatalf("exact bound: %v", err)
	}
	if _, err := ParseAgent(strings.NewReader(padded + " ")); err == nil {
		t.Fatal("accepted oversize")
	}
	if _, err := ParseAgent(strings.NewReader(good + "\n\t")); err != nil {
		t.Fatal(err)
	}
}

func TestAgentValidation(t *testing.T) {
	tests := map[string]func(*Agent){
		"remote http":      func(c *Agent) { c.BackendURL = "http://backend.example.invalid/api/v1" },
		"localhost http":   func(c *Agent) { c.BackendURL = "http://localhost/api/v1" },
		"credentials":      func(c *Agent) { c.BackendURL = "https://canary-secret:canary-secret@host/api/v1" },
		"query":            func(c *Agent) { c.BackendURL += "?token=canary-secret" },
		"empty query":      func(c *Agent) { c.BackendURL += "?" },
		"fragment":         func(c *Agent) { c.BackendURL += "#canary-secret" },
		"empty fragment":   func(c *Agent) { c.BackendURL += "#" },
		"path":             func(c *Agent) { c.BackendURL += "/events" },
		"encoded path":     func(c *Agent) { c.BackendURL = "https://host/api/%76%31" },
		"port zero":        func(c *Agent) { c.BackendURL = "https://host:0/api/v1" },
		"port high":        func(c *Agent) { c.BackendURL = "https://host:65536/api/v1" },
		"port empty":       func(c *Agent) { c.BackendURL = "https://host:/api/v1" },
		"unbracketed IPv6": func(c *Agent) { c.BackendURL = "https://2001:db8::1:443/api/v1" },
		"bracketed DNS":    func(c *Agent) { c.BackendURL = "https://[host]/api/v1" },
		"token relative":   func(c *Agent) { c.TokenFile = "canary-secret" },
		"state traversal":  func(c *Agent) { c.StateDir = "/var/../canary-secret" },
		"state root":       func(c *Agent) { c.StateDir = "/" },
		"socket URI":       func(c *Agent) { c.DockerSocket = "unix:///var/run/docker.sock" },
		"socket control":   func(c *Agent) { c.DockerSocket = "/canary-secret\n" },
		"level":            func(c *Agent) { c.Logging.Level = "canary-secret" },
		"format":           func(c *Agent) { c.Logging.Format = "canary-secret" },
		"timeout low":      func(c *Agent) { c.RequestTimeoutSeconds = 0 },
		"timeout high":     func(c *Agent) { c.RequestTimeoutSeconds = 121 },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			c := validAgent()
			mutate(&c)
			_, err := ParseAgent(strings.NewReader(encoded(t, c)))
			if err == nil {
				t.Fatal("accepted invalid settings")
			}
			if strings.Contains(err.Error(), "canary-secret") {
				t.Fatal("leaked value")
			}
		})
	}
	for _, u := range []string{"https://host/api/v1", "http://127.0.0.1:8080/api/v1", "http://[::1]:8080/api/v1", "http://[::ffff:127.0.0.1]/api/v1"} {
		c := validAgent()
		c.BackendURL = u
		if err := c.Validate(); err != nil {
			t.Fatalf("%s: %v", u, err)
		}
	}
	for _, level := range []string{"debug", "info", "warn", "error"} {
		for _, format := range []string{"json", "text"} {
			if err := (Logging{Level: level, Format: format}).Validate(); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestBackendValidation(t *testing.T) {
	tests := map[string]func(*Backend){
		"version":              func(c *Backend) { c.Version = 0 },
		"public listen":        func(c *Backend) { c.Listen = "0.0.0.0:8080" },
		"dns listen":           func(c *Backend) { c.Listen = "localhost:8080" },
		"listen zero":          func(c *Backend) { c.Listen = "127.0.0.1:0" },
		"listen high":          func(c *Backend) { c.Listen = "[::1]:65536" },
		"token":                func(c *Backend) { c.TokenFile = "canary-secret" },
		"request timeout":      func(c *Backend) { c.RequestTimeoutSeconds = 121 },
		"remote disable":       func(c *Backend) { c.PostgreSQL.SSLMode = "disable"; c.PostgreSQL.AllowLocalInsecure = true },
		"remote prefer":        func(c *Backend) { c.PostgreSQL.SSLMode = "prefer" },
		"remote require":       func(c *Backend) { c.PostgreSQL.SSLMode = "require" },
		"remote verify-ca":     func(c *Backend) { c.PostgreSQL.SSLMode = "verify-ca" },
		"remote local flag":    func(c *Backend) { c.PostgreSQL.AllowLocalInsecure = true },
		"local missing opt-in": func(c *Backend) { c.PostgreSQL.Host = "127.0.0.1"; c.PostgreSQL.SSLMode = "disable" },
		"localhost not literal": func(c *Backend) {
			c.PostgreSQL.Host = "localhost"
			c.PostgreSQL.SSLMode = "disable"
			c.PostgreSQL.AllowLocalInsecure = true
		},
		"unix verify-full": func(c *Backend) { c.PostgreSQL.Host = "/var/run/postgresql" },
		"host DSN":         func(c *Backend) { c.PostgreSQL.Host = "postgres://canary-secret@host/db" },
		"host parameters":  func(c *Backend) { c.PostgreSQL.Host = "host sslmode=disable" },
		"host empty":       func(c *Backend) { c.PostgreSQL.Host = "" },
		"port low":         func(c *Backend) { c.PostgreSQL.Port = 0 },
		"port high":        func(c *Backend) { c.PostgreSQL.Port = 65536 },
		"db":               func(c *Backend) { c.PostgreSQL.Database = "canary-secret?" },
		"user":             func(c *Backend) { c.PostgreSQL.User = "canary-secret?" },
		"admin":            func(c *Backend) { c.PostgreSQL.User = "postgres" },
		"password path":    func(c *Backend) { c.PostgreSQL.PasswordFile = "canary-secret" },
		"pool low":         func(c *Backend) { c.PostgreSQL.PoolMax = 0 },
		"pool budget":      func(c *Backend) { c.PostgreSQL.PoolMax = 31 },
		"budget high":      func(c *Backend) { c.PostgreSQL.TotalConnectionBudget = 31 },
		"budget low":       func(c *Backend) { c.PostgreSQL.TotalConnectionBudget = 0 },
		"connect low":      func(c *Backend) { c.PostgreSQL.ConnectTimeoutSeconds = 0 },
		"connect high":     func(c *Backend) { c.PostgreSQL.ConnectTimeoutSeconds = 31 },
		"statement low":    func(c *Backend) { c.PostgreSQL.StatementTimeoutSeconds = 0 },
		"statement high":   func(c *Backend) { c.PostgreSQL.StatementTimeoutSeconds = 121 },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			c := validBackend()
			mutate(&c)
			_, err := ParseBackend(strings.NewReader(encoded(t, c)))
			if err == nil {
				t.Fatal("accepted invalid settings")
			}
			if strings.Contains(err.Error(), "canary-secret") {
				t.Fatal("leaked value")
			}
		})
	}
	for _, host := range []string{"127.0.0.1", "::1", "/var/run/postgresql"} {
		c := validBackend()
		c.PostgreSQL.Host = host
		c.PostgreSQL.SSLMode = "disable"
		c.PostgreSQL.AllowLocalInsecure = true
		c.Listen = "[::1]:8080"
		if err := c.Validate(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDefaultsAndExamples(t *testing.T) {
	agent := `{"version":1,"backend_url":"https://host/api/v1","token_file":"/etc/argus/token","state_dir":"/var/lib/argus","docker_socket":"/var/run/docker.sock"}`
	a, err := ParseAgent(strings.NewReader(agent))
	if err != nil {
		t.Fatal(err)
	}
	if a.Logging.Level != "info" || a.Logging.Format != "json" || a.RequestTimeoutSeconds != 30 {
		t.Fatal("incorrect agent defaults")
	}
	backend := `{"version":1,"token_file":"/etc/argus/token","postgresql":{"host":"db.example.invalid","database":"argus","user":"argus_app","password_file":"/etc/argus/password"}}`
	b, err := ParseBackend(strings.NewReader(backend))
	if err != nil {
		t.Fatal(err)
	}
	if b.PostgreSQL.PoolMax != 16 || b.PostgreSQL.TotalConnectionBudget != 30 || b.PostgreSQL.SSLMode != "verify-full" || b.PostgreSQL.Port != 5432 {
		t.Fatal("incorrect backend defaults")
	}
	for _, component := range []string{"agent", "backend"} {
		src, err := os.ReadFile(filepath.Join("..", "..", "configs", component+".example.json"))
		if err != nil {
			t.Fatal(err)
		}
		filename := filepath.Join(t.TempDir(), "config.json")
		if err := os.WriteFile(filename, src, 0600); err != nil {
			t.Fatal(err)
		}
		if component == "agent" {
			_, err = LoadAgent(filename)
		} else {
			_, err = LoadBackend(filename)
		}
		if err != nil {
			t.Fatal(err)
		}
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("canary-secret") }
func TestSafeIOErrors(t *testing.T) {
	_, err := ParseAgent(failingReader{})
	if err == nil || strings.Contains(err.Error(), "canary-secret") {
		t.Fatal("unsafe read error")
	}
	for _, load := range []func(string) error{
		func(s string) error { _, err := LoadAgent(s); return err },
		func(s string) error { _, err := LoadBackend(s); return err },
	} {
		err := load(filepath.Join(t.TempDir(), "canary-secret"))
		if err == nil || strings.Contains(err.Error(), "canary-secret") {
			t.Fatal("unsafe file error")
		}
	}
}

func FuzzParseConfig(f *testing.F) {
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"version":1,"logging":{"level":"info"}}`))
	f.Add([]byte(`{"version":1,"version":1}`))
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > MaxBytes+1 {
			return
		}
		_, _ = ParseAgent(bytes.NewReader(b))
		_, _ = ParseBackend(bytes.NewReader(b))
	})
}

var _ io.Reader = failingReader{}
