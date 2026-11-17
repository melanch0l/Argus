package environment

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"sync"
	"time"
)

// cappedWriter keeps consuming after its storage cap so a verbose child cannot
// deadlock on a full pipe. Stdout and stderr each have an independent bound.
type cappedWriter struct {
	mu        sync.Mutex
	data      []byte
	limit     int
	truncated bool
	present   bool
}

func (w *cappedWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	n := len(p)
	w.present = w.present || n > 0
	remaining := w.limit - len(w.data)
	if len(p) > remaining {
		w.truncated = true
		p = p[:remaining]
	}
	w.data = append(w.data, p...)
	return n, nil
}

type ExecRunner struct{}

func (ExecRunner) Run(parent context.Context, name string, args ...string) Command {
	ctx, cancel := context.WithTimeout(parent, CommandTimeout)
	defer cancel()
	r := Command{Name: name, Args: append([]string{}, args...), TimeoutMS: CommandTimeout.Milliseconds()}
	out := &cappedWriter{limit: MaxOutputBytes}
	errout := &cappedWriter{limit: 8192}
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout, cmd.Stderr = out, errout
	// Descendants retaining inherited pipes must not hold Wait indefinitely.
	cmd.WaitDelay = 100 * time.Millisecond
	err := cmd.Run()
	if cmd.ProcessState != nil {
		code := cmd.ProcessState.ExitCode()
		r.ExitCode = &code
	}
	r.Stdout, r.StdoutTruncated = string(out.data), out.truncated
	r.StderrPresent, r.StderrTruncated = errout.present, errout.truncated
	switch {
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		r.TimedOut, r.Error = true, "deadline_exceeded"
	case ctx.Err() != nil:
		r.Error = "canceled"
	case errors.Is(err, exec.ErrNotFound), errors.Is(err, os.ErrNotExist):
		r.Error = "not_found"
	case errors.Is(err, os.ErrPermission):
		r.Error = "permission_denied"
	case err != nil:
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			r.Error = "exit_nonzero"
		} else {
			r.Error = "execution_error"
		}
	}
	// Raw stderr can contain endpoints, paths or credentials. Only presence and
	// truncation are recorded; errors are classified, not copied verbatim.
	return r
}

func commandStatus(c Command) Status {
	if c.Error == "not_found" || c.Error == "permission_denied" || c.Error == "canceled" {
		return Blocked
	}
	if c.Error != "" || c.TimedOut || c.ExitCode == nil || *c.ExitCode != 0 || c.StdoutTruncated {
		return Fail
	}
	return Pass
}
