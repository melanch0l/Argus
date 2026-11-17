# Integration test boundary

Portable foundation unit/CLI/parser/environment tests live with their Go packages and run with `go test ./...`. Those tests do not exercise the kernel or PostgreSQL.

This directory is reserved for explicitly gated integration tests introduced with their approved feature phases. Kernel tests require the real Ubuntu target, effective privileges, a verified toolchain and controlled Docker workloads. Storage tests require PostgreSQL 17, a disposable isolated database and Phase 7 schema. Never connect to an arbitrary database, load BPF or modify host state merely because a developer runs the portable unit suite.

Record commands, actual environment, outcomes, skipped checks and cleanup. Synthetic fixtures must be labelled; real monitoring and experimental claims require real observations.
