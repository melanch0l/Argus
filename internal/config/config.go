// Package config validates portable settings only. It never reads secret files,
// probes the environment, or establishes connections.
package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/netip"
	"net/url"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	Version  = 1
	MaxBytes = 64 * 1024
)

type Logging struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

func (c Logging) Validate() error {
	switch c.Level {
	case "debug", "info", "warn", "error":
	default:
		return errors.New("logging.level must be debug, info, warn, or error")
	}
	switch c.Format {
	case "json", "text":
	default:
		return errors.New("logging.format must be json or text")
	}
	return nil
}

type Agent struct {
	Version               int     `json:"version"`
	Logging               Logging `json:"logging"`
	BackendURL            string  `json:"backend_url"`
	TokenFile             string  `json:"token_file"`
	StateDir              string  `json:"state_dir"`
	DockerSocket          string  `json:"docker_socket"`
	RequestTimeoutSeconds int     `json:"request_timeout_seconds"`
}

type Backend struct {
	Version               int        `json:"version"`
	Logging               Logging    `json:"logging"`
	Listen                string     `json:"listen"`
	TokenFile             string     `json:"token_file"`
	PostgreSQL            PostgreSQL `json:"postgresql"`
	RequestTimeoutSeconds int        `json:"request_timeout_seconds"`
}

type PostgreSQL struct {
	Host                    string `json:"host"`
	Port                    int    `json:"port"`
	Database                string `json:"database"`
	User                    string `json:"user"`
	PasswordFile            string `json:"password_file"`
	SSLMode                 string `json:"sslmode"`
	AllowLocalInsecure      bool   `json:"allow_local_insecure"`
	PoolMax                 int    `json:"pool_max"`
	TotalConnectionBudget   int    `json:"total_connection_budget"`
	ConnectTimeoutSeconds   int    `json:"connect_timeout_seconds"`
	StatementTimeoutSeconds int    `json:"statement_timeout_seconds"`
}

// Defaults intentionally omit version and required deployment-specific settings.
// A version must be present in every configuration document.
func AgentDefaults() Agent {
	return Agent{Logging: Logging{Level: "info", Format: "json"}, RequestTimeoutSeconds: 30}
}

func BackendDefaults() Backend {
	return Backend{Logging: Logging{Level: "info", Format: "json"}, Listen: "127.0.0.1:8080", RequestTimeoutSeconds: 30,
		PostgreSQL: PostgreSQL{Port: 5432, SSLMode: "verify-full", PoolMax: 16, TotalConnectionBudget: 30, ConnectTimeoutSeconds: 5, StatementTimeoutSeconds: 30}}
}

func LoadAgent(filename string) (Agent, error) {
	f, err := os.Open(filename)
	if err != nil {
		return Agent{}, errors.New("cannot open configuration file")
	}
	defer f.Close()
	return ParseAgent(f)
}

func LoadBackend(filename string) (Backend, error) {
	f, err := os.Open(filename)
	if err != nil {
		return Backend{}, errors.New("cannot open configuration file")
	}
	defer f.Close()
	return ParseBackend(f)
}

func ParseAgent(r io.Reader) (Agent, error) {
	c := AgentDefaults()
	if err := decode(r, &c); err != nil {
		return Agent{}, err
	}
	if err := c.Validate(); err != nil {
		return Agent{}, err
	}
	return c, nil
}

func ParseBackend(r io.Reader) (Backend, error) {
	c := BackendDefaults()
	if err := decode(r, &c); err != nil {
		return Backend{}, err
	}
	if err := c.Validate(); err != nil {
		return Backend{}, err
	}
	return c, nil
}

func decode(r io.Reader, dst any) error {
	b, err := io.ReadAll(io.LimitReader(r, MaxBytes+1))
	if err != nil {
		return errors.New("cannot read configuration")
	}
	if len(b) > MaxBytes {
		return errors.New("configuration exceeds 65536 bytes")
	}
	if !utf8.Valid(b) {
		return errors.New("configuration must be UTF-8 JSON")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	if err := checkValue(d, reflect.TypeOf(dst).Elem()); err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return errors.New("configuration must contain exactly one JSON object")
	}
	// Discard decoder errors: they can contain attacker-controlled keys/values.
	if err := json.Unmarshal(b, dst); err != nil {
		return errors.New("configuration contains an invalid field type or value")
	}
	return nil
}

// Walking the schema before decoding also rejects case aliases (encoding/json
// otherwise accepts them), duplicate escaped keys, and null overriding defaults.
// Recursion is bounded by the fixed, nonrecursive configuration schema.
func checkValue(d *json.Decoder, typ reflect.Type) error {
	tok, err := d.Token()
	if err != nil {
		return errors.New("configuration is not valid JSON")
	}
	if tok == nil {
		return errors.New("configuration must not contain null")
	}
	if typ.Kind() != reflect.Struct {
		if _, ok := tok.(json.Delim); ok {
			return errors.New("configuration contains an invalid field type")
		}
		return nil
	}
	if tok != json.Delim('{') {
		return errors.New("configuration requires JSON objects")
	}
	fields := make(map[string]reflect.Type, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		fields[f.Tag.Get("json")] = f.Type
	}
	seen := make(map[string]bool, len(fields))
	for d.More() {
		key, err := d.Token()
		if err != nil {
			return errors.New("configuration is not valid JSON")
		}
		name, ok := key.(string)
		if !ok {
			return errors.New("configuration contains an invalid key")
		}
		fieldType, ok := fields[name]
		if !ok {
			return errors.New("configuration contains an unknown field")
		}
		if seen[name] {
			return errors.New("configuration contains a duplicate key")
		}
		seen[name] = true
		if err := checkValue(d, fieldType); err != nil {
			return err
		}
	}
	if tok, err := d.Token(); err != nil || tok != json.Delim('}') {
		return errors.New("configuration is not valid JSON")
	}
	return nil
}

func (c Agent) Validate() error {
	if c.Version != Version {
		return errors.New("unsupported or missing configuration version")
	}
	if err := c.Logging.Validate(); err != nil {
		return err
	}
	if !validBaseURL(c.BackendURL) {
		return errors.New("backend_url requires HTTPS (HTTP only on literal loopback), optional valid port, and exactly /api/v1 without credentials, query, or fragment")
	}
	if !absolutePath(c.TokenFile) {
		return errors.New("token_file must be an absolute Unix file path")
	}
	if !absolutePath(c.StateDir) {
		return errors.New("state_dir must be an absolute Unix directory path")
	}
	if !absolutePath(c.DockerSocket) {
		return errors.New("docker_socket must be an absolute Unix socket path")
	}
	if !bounded(c.RequestTimeoutSeconds, 1, 120) {
		return errors.New("request_timeout_seconds must be within 1..120")
	}
	return nil
}

func (c Backend) Validate() error {
	if c.Version != Version {
		return errors.New("unsupported or missing configuration version")
	}
	if err := c.Logging.Validate(); err != nil {
		return err
	}
	host, port, err := net.SplitHostPort(c.Listen)
	if err != nil || !loopback(host) || !validPort(port) {
		return errors.New("listen must be a literal loopback host:port with port within 1..65535")
	}
	if !absolutePath(c.TokenFile) {
		return errors.New("token_file must be an absolute Unix file path")
	}
	if !bounded(c.RequestTimeoutSeconds, 1, 120) {
		return errors.New("request_timeout_seconds must be within 1..120")
	}
	return c.PostgreSQL.Validate()
}

func (c PostgreSQL) Validate() error {
	local := loopback(c.Host) || absolutePath(c.Host)
	if !validHost(c.Host) && !absolutePath(c.Host) {
		return errors.New("postgresql.host must be a hostname, literal IP, or absolute Unix socket directory")
	}
	if !bounded(c.Port, 1, 65535) {
		return errors.New("postgresql.port must be within 1..65535")
	}
	if !identifier(c.Database) || !identifier(c.User) {
		return errors.New("postgresql.database and user must be simple identifiers of 1..63 characters")
	}
	switch strings.ToLower(c.User) {
	case "postgres", "root", "admin", "administrator", "rdsadmin", "azure_pg_admin":
		return errors.New("postgresql.user must be a dedicated application role, not a bootstrap or admin role")
	}
	if !absolutePath(c.PasswordFile) {
		return errors.New("postgresql.password_file must be an absolute Unix file path")
	}
	if c.SSLMode != "verify-full" && !(c.SSLMode == "disable" && local && c.AllowLocalInsecure) {
		return errors.New("postgresql.sslmode requires verify-full; disable requires a literal loopback or Unix socket host and allow_local_insecure=true")
	}
	if c.AllowLocalInsecure && (!local || c.SSLMode != "disable") {
		return errors.New("postgresql.allow_local_insecure is only valid for explicit local sslmode=disable")
	}
	// Unix sockets do not provide TLS peer verification.
	if absolutePath(c.Host) && c.SSLMode != "disable" {
		return errors.New("PostgreSQL Unix sockets require explicit local insecure settings")
	}
	if !bounded(c.TotalConnectionBudget, 1, 30) || !bounded(c.PoolMax, 1, c.TotalConnectionBudget) {
		return errors.New("postgresql pool_max must be within 1..total_connection_budget, with total_connection_budget within 1..30")
	}
	if !bounded(c.ConnectTimeoutSeconds, 1, 30) {
		return errors.New("postgresql.connect_timeout_seconds must be within 1..30")
	}
	if !bounded(c.StatementTimeoutSeconds, 1, 120) {
		return errors.New("postgresql.statement_timeout_seconds must be within 1..120")
	}
	return nil
}

func bounded(n, min, max int) bool { return n >= min && n <= max }
func loopback(s string) bool {
	ip, err := netip.ParseAddr(s)
	return err == nil && ip.Zone() == "" && ip.Unmap().IsLoopback()
}
func validPort(s string) bool {
	if s == "" {
		return false
	}
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	n, err := strconv.Atoi(s)
	return err == nil && bounded(n, 1, 65535)
}
func absolutePath(s string) bool {
	return strings.HasPrefix(s, "/") && s != "/" && path.Clean(s) == s && !strings.ContainsAny(s, "\\\x00") && !strings.ContainsFunc(s, unicode.IsControl)
}
func identifier(s string) bool {
	if len(s) < 1 || len(s) > 63 {
		return false
	}
	for i, ch := range s {
		if ch != '_' && !(ch >= 'a' && ch <= 'z') && !(ch >= 'A' && ch <= 'Z') && !(i > 0 && ch >= '0' && ch <= '9') {
			return false
		}
	}
	return true
}
func validHost(s string) bool {
	if ip, err := netip.ParseAddr(s); err == nil {
		return ip.Zone() == ""
	}
	if len(s) == 0 || len(s) > 253 {
		return false
	}
	for _, label := range strings.Split(s, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, ch := range label {
			if !(ch >= 'a' && ch <= 'z') && !(ch >= 'A' && ch <= 'Z') && !(ch >= '0' && ch <= '9') && ch != '-' {
				return false
			}
		}
	}
	return true
}
func validBaseURL(s string) bool {
	u, err := url.Parse(s)
	if err != nil || u.Opaque != "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || strings.Contains(s, "#") || u.Path != "/api/v1" || u.RawPath != "" {
		return false
	}
	if !validHost(u.Hostname()) || (u.Port() != "" && !validPort(u.Port())) || strings.HasSuffix(u.Host, ":") {
		return false
	}
	// Require an unambiguous authority: IPv6 literals must be bracketed, and
	// brackets must not turn a DNS name into a purported IP literal.
	expectedHost := u.Hostname()
	if strings.Contains(expectedHost, ":") {
		expectedHost = "[" + expectedHost + "]"
	}
	if u.Port() != "" {
		expectedHost = net.JoinHostPort(u.Hostname(), u.Port())
	}
	if u.Host != expectedHost {
		return false
	}
	return u.Scheme == "https" || (u.Scheme == "http" && loopback(u.Hostname()))
}
