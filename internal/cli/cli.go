// Package cli implements the portable, deliberately non-running Phase 2 commands.
package cli

import (
	"flag"
	"fmt"
	"io"

	"github.com/melanch0l/Argus/internal/buildinfo"
	"github.com/melanch0l/Argus/internal/config"
	"github.com/melanch0l/Argus/internal/logging"
)

const (
	ExitOK             = 0
	ExitNotImplemented = 1
	ExitUsage          = 2
)

type Component string

const (
	Agent   Component = "agent"
	Backend Component = "backend"
)

// Run opens only the explicitly selected config file for --check-config. It does
// not read environment overrides or secret files, create directories, or connect.
func Run(component Component, args []string, stdout, stderr io.Writer) int {
	if component != Agent && component != Backend {
		fmt.Fprintln(stderr, "unsupported component")
		return ExitUsage
	}
	fs := flag.NewFlagSet("argus-"+string(component), flag.ContinueOnError)
	// flag errors include caller-provided flag names and values; never echo them.
	fs.SetOutput(io.Discard)
	version := fs.Bool("version", false, "print deterministic version and phase")
	check := fs.Bool("check-config", false, "validate config syntax/settings only; no environment or credential checks")
	filename := fs.String("config", "", "JSON configuration file (required with --check-config)")
	showUsage := func() {
		fmt.Fprintf(stdout, "Usage: argus-%s [--version | --check-config --config PATH]\n", component)
		fmt.Fprintln(stdout, "  --version       print version and phase")
		fmt.Fprintln(stdout, "  --check-config  validate syntax/settings only, not environment or credentials")
		fmt.Fprintln(stdout, "  --config PATH   explicit JSON configuration file")
		fmt.Fprintln(stdout, "  -h, --help      show help")
		fmt.Fprintln(stdout, "Default startup is not_implemented in Phase 2; no service is started.")
	}
	fs.Usage = func() {} // flag invokes Usage on errors as well as explicit help.
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			showUsage()
			return ExitOK
		}
		fmt.Fprintln(stderr, "invalid command line; use --help")
		return ExitUsage
	}
	configSet := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "config" {
			configSet = true
		}
	})
	if fs.NArg() != 0 || (*version && (*check || configSet)) || (*check && *filename == "") || (configSet && !*check) {
		fmt.Fprintln(stderr, "invalid flag combination; use --help")
		return ExitUsage
	}
	if *version {
		fmt.Fprintf(stdout, "argus-%s %s %s\n", component, buildinfo.Version, buildinfo.Phase)
		return ExitOK
	}
	if *check {
		var err error
		if component == Agent {
			_, err = config.LoadAgent(*filename)
		} else {
			_, err = config.LoadBackend(*filename)
		}
		if err != nil {
			fmt.Fprintln(stderr, "configuration invalid:", err)
			return ExitUsage
		}
		fmt.Fprintln(stdout, "configuration valid (syntax/settings only; environment, credentials, and connectivity not checked)")
		return ExitOK
	}
	logger, err := logging.New(stderr, string(component), config.Logging{Level: "info", Format: "json"})
	if err != nil {
		fmt.Fprintln(stderr, "cannot initialize logging")
		return ExitUsage
	}
	logger.Error("service startup is not implemented in Phase 2", "status", "not_implemented", "phase", buildinfo.Phase)
	return ExitNotImplemented
}
