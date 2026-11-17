package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/melanch0l/Argus/internal/environment"
)

type absentRunner struct{}

func (absentRunner) Run(_ context.Context, name string, args ...string) environment.Command {
	return environment.Command{Name: name, Args: args, Error: "not_found"}
}

type absentFiles struct{}

func (absentFiles) Read(string, int) ([]byte, bool, error) { return nil, false, os.ErrNotExist }
func inspector() environment.Inspector {
	return environment.Inspector{GOOS: "darwin", GOARCH: "arm64", EUID: 1000, Runner: absentRunner{}, Files: absentFiles{}}
}

func TestCLIJSONAndExitCodes(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		code int
		json bool
	}{
		{"default", nil, 0, true}, {"strict", []string{"--strict"}, 1, true},
		{"privileged_false", []string{"--privileged=false"}, 0, true},
		{"privileged_nonlinux", []string{"--privileged=true"}, 0, true},
		{"bad_flag", []string{"--unknown"}, 2, false}, {"positional", []string{"oops"}, 2, false},
		{"help", []string{"--help"}, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out, errout bytes.Buffer
			code := run(context.Background(), tc.args, &out, &errout, inspector())
			if code != tc.code {
				t.Fatalf("exit %d, want %d: %s", code, tc.code, errout.String())
			}
			if tc.json {
				if !json.Valid(out.Bytes()) {
					t.Fatalf("stdout is not JSON: %s", out.String())
				}
				if errout.Len() != 0 {
					t.Fatalf("report diagnostics leaked: %s", errout.String())
				}
			} else if out.Len() != 0 {
				t.Fatal("help/errors mixed into stdout")
			}
		})
	}
}

func TestPrivateOutputReplacement(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.json")
	if err := os.WriteFile(path, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), []string{"--output", path}, &stdout, &stderr, inspector()); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(b) || stdout.Len() != 0 {
		t.Fatal("file report invalid or duplicated to stdout")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("output mode %o", info.Mode().Perm())
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatal("temporary files leaked")
	}
}

func TestRefuseSymlinkAndMissingParent(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	link := filepath.Join(dir, "link")
	if err := os.WriteFile(target, []byte("preserve"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{link, filepath.Join(dir, "missing", "report.json"), dir} {
		var out, errout bytes.Buffer
		if code := run(context.Background(), []string{"--output", path}, &out, &errout, inspector()); code != 2 {
			t.Fatalf("expected IO failure, got %d", code)
		}
		if out.Len() != 0 || errout.Len() == 0 {
			t.Fatal("IO error handling mixed stdout/stderr")
		}
	}
	b, err := os.ReadFile(target)
	if err != nil || string(b) != "preserve" {
		t.Fatal("symlink target modified")
	}
}
