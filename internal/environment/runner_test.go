package environment

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// The current test binary provides portable subprocesses without invoking a
// shell, relying on installed tools or emitting private host information.
func TestRunnerChild(t *testing.T) {
	if len(os.Args) < 2 {
		return
	}
	switch os.Args[len(os.Args)-1] {
	case "envcheck-child-fail":
		os.Stderr.WriteString("PRIVATE_STDERR\n")
		os.Exit(9)
	case "envcheck-child-flood":
		os.Stdout.WriteString(strings.Repeat("a", MaxOutputBytes*3))
		os.Stderr.WriteString(strings.Repeat("b", 32768))
		os.Exit(0)
	case "envcheck-child-sleep":
		time.Sleep(30 * time.Second)
		os.Exit(0)
	}
}

func TestCappedWriter(t *testing.T) {
	w := cappedWriter{limit: 4}
	for _, value := range []string{"abc", "defgh", "ij"} {
		n, err := w.Write([]byte(value))
		if n != len(value) || err != nil {
			t.Fatalf("short write: %d %v", n, err)
		}
	}
	if string(w.data) != "abcd" || !w.truncated {
		t.Fatalf("incorrect cap: data=%q truncated=%v", w.data, w.truncated)
	}
}

func TestExecRunner(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	runner := ExecRunner{}
	for _, tc := range []struct {
		mode      string
		code      int
		truncated bool
	}{
		{"envcheck-child-fail", 9, false}, {"envcheck-child-flood", 0, true},
	} {
		t.Run(tc.mode, func(t *testing.T) {
			c := runner.Run(context.Background(), exe, "-test.run=^TestRunnerChild$", "--", tc.mode)
			if c.ExitCode == nil || *c.ExitCode != tc.code {
				t.Fatalf("wrong status: %+v", c)
			}
			if len(c.Stdout) > MaxOutputBytes || c.StdoutTruncated != tc.truncated {
				t.Fatalf("output cap incorrect: %d", len(c.Stdout))
			}
			if !c.StderrPresent || c.StderrTruncated != tc.truncated {
				t.Fatalf("stderr metadata incorrect: %+v", c)
			}
			if strings.Contains(c.Stdout, "PRIVATE_STDERR") {
				t.Fatal("stderr persisted")
			}
		})
	}
	c := runner.Run(context.Background(), "envcheck-intentionally-nonexistent-command-42")
	if c.Error != "not_found" || c.ExitCode != nil {
		t.Fatalf("missing command incorrectly classified: %+v", c)
	}
}

func TestExecRunnerDeadline(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	c := (ExecRunner{}).Run(ctx, exe, "-test.run=^TestRunnerChild$", "--", "envcheck-child-sleep")
	if !c.TimedOut || c.Error != "deadline_exceeded" {
		t.Fatalf("missing timeout: %+v", c)
	}
	if time.Since(start) > 3*time.Second {
		t.Fatal("deadline did not bound subprocess")
	}
}

func TestOSFilesBound(t *testing.T) {
	path := t.TempDir() + "/input"
	if err := os.WriteFile(path, []byte("abcdef"), 0600); err != nil {
		t.Fatal(err)
	}
	b, truncated, err := (OSFiles{}).Read(path, 3)
	if err != nil || string(b) != "abc" || !truncated {
		t.Fatalf("read bound failed: %q %v %v", b, truncated, err)
	}
}
