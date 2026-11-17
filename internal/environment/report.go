// Package environment performs bounded, passive Phase 2 environment inspection.
// Feature discovery is not proof that an intended monitoring program can attach.
package environment

import (
	"context"
	"io"
	"os"
	"runtime"
	"time"
)

type Status string

const (
	Pass           Status = "pass"
	Fail           Status = "fail"
	Blocked        Status = "blocked"
	NotTested      Status = "not_tested"
	CommandTimeout        = 5 * time.Second
	MaxOutputBytes        = 256 * 1024
	MaxFileBytes          = 64 * 1024
)

type Check struct {
	ID       string `json:"id"`
	Status   Status `json:"status"`
	Required bool   `json:"required_for_strict"`
	Summary  string `json:"summary"`
	Evidence any    `json:"evidence,omitempty"`
}

type Command struct {
	Name            string   `json:"name"`
	Args            []string `json:"args"`
	TimeoutMS       int64    `json:"timeout_ms"`
	ExitCode        *int     `json:"exit_code"`
	TimedOut        bool     `json:"timed_out"`
	Error           string   `json:"error,omitempty"`
	Stdout          string   `json:"stdout,omitempty"`
	StdoutTruncated bool     `json:"stdout_truncated"`
	StderrPresent   bool     `json:"stderr_present"`
	StderrTruncated bool     `json:"stderr_truncated"`
	OutputOmitted   bool     `json:"output_omitted"`
}

type Report struct {
	SchemaVersion       int       `json:"schema_version"`
	RuntimeOS           string    `json:"runtime_os"`
	RuntimeArch         string    `json:"runtime_arch"`
	PrivilegedRequested bool      `json:"privileged_requested"`
	StrictPassed        bool      `json:"strict_passed"`
	MonitoringReady     bool      `json:"monitoring_ready"`
	Checks              []Check   `json:"checks"`
	Commands            []Command `json:"commands"`
	Limitations         []string  `json:"limitations"`
}

// StrictPass requires positive evidence for every target prerequisite. Phase 3
// load/attach is deliberately separate: passing this gate never means ready.
func (r Report) StrictPass() bool {
	n := 0
	for _, c := range r.Checks {
		if c.Required {
			n++
			if c.Status != Pass {
				return false
			}
		}
	}
	return n > 0
}

type Runner interface {
	Run(context.Context, string, ...string) Command
}

// Files reads at most limit bytes and reports whether more bytes existed.
// It must not return the entire BTF object, mount table, or arbitrary tree.
type Files interface {
	Read(string, int) ([]byte, bool, error)
}

type OSFiles struct{}

func (OSFiles) Read(path string, limit int) ([]byte, bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, false, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, int64(limit)+1))
	if len(b) > limit {
		return b[:limit], true, err
	}
	return b, false, err
}

type Options struct{ Privileged bool }

// Inspector's dependencies make portable tests independent of host tools,
// filesystem contents and effective privileges. Use New for the real host.
type Inspector struct {
	GOOS   string
	GOARCH string
	EUID   int
	Runner Runner
	Files  Files
	// Getenv is limited to checking Docker endpoint override presence. Nil
	// means unknown, not an empty environment; tests need not read host env.
	Getenv func(string) string
}

func New() Inspector {
	return Inspector{
		GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, EUID: os.Geteuid(),
		Runner: ExecRunner{}, Files: OSFiles{}, Getenv: os.Getenv,
	}
}
