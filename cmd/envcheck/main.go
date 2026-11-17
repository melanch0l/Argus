// envcheck inspects Phase 2 prerequisites; it never claims monitoring readiness.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/melanch0l/Argus/internal/environment"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	os.Exit(run(ctx, os.Args[1:], os.Stdout, os.Stderr, environment.New()))
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer, inspector environment.Inspector) int {
	flags := flag.NewFlagSet("envcheck", flag.ContinueOnError)
	flags.SetOutput(stderr)
	output := flags.String("output", "", "write JSON to a new/existing regular file with mode 0600 (parent directory must exist); default stdout")
	privileged := flags.Bool("privileged", false, "explicitly authorize transient bpftool kernel feature probing (Linux/root only)")
	strict := flags.Bool("strict", false, "exit 1 unless all required Ubuntu 24.04 Phase 2 checks pass; does not establish Phase 3 readiness")
	flags.Usage = func() {
		fmt.Fprintln(stderr, "Usage: envcheck [--output path] [--privileged[=true|false]] [--strict[=true|false]]")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "envcheck: positional arguments are not supported")
		return 2
	}
	report := inspector.Inspect(ctx, environment.Options{Privileged: *privileged})
	if *output != "" {
		if err := writeReport(*output, report); err != nil {
			fmt.Fprintln(stderr, "envcheck: cannot write private JSON report (parent must exist; symlinks and non-regular files are refused)")
			return 2
		}
	} else if err := encodeReport(stdout, report); err != nil {
		fmt.Fprintln(stderr, "envcheck: cannot write JSON report")
		return 2
	}
	if *strict && !report.StrictPassed {
		return 1
	}
	return 0
}

func encodeReport(w io.Writer, report environment.Report) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

// Atomic replacement avoids following a final symlink or modifying another
// hard-link target. A temporary file is private from its first write. Parent
// directories must already exist and are trusted operator-controlled paths.
func writeReport(path string, report environment.Report) error {
	info, err := os.Lstat(path)
	if err == nil && !info.Mode().IsRegular() {
		return errors.New("output is not a regular file")
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".envcheck-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	err = f.Chmod(0600)
	if err == nil {
		err = encodeReport(f, report)
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
