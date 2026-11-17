package main

import (
	"os"

	"github.com/melanch0l/Argus/internal/cli"
)

func main() { os.Exit(cli.Run(cli.Backend, os.Args[1:], os.Stdout, os.Stderr)) }
