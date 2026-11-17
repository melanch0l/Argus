# Phase 2 foundation implementation and verification record

**Date:** 2026-11-02

**Authorization:** User explicitly requested Phase 2 implementation and replacement of SQLite with PostgreSQL. No commit, push, branch, Phase 3 or later feature implementation is authorized by that request.

**Status:** Portable foundation implemented and locally validated. Full Phase 2 acceptance **blocked on actual Ubuntu target evidence**; not accepted. No monitoring data or research benchmark was generated.

## Changes and decisions

- Canonical specification revision 2 and architecture/research/evaluation/roadmap documents now use PostgreSQL 17, not operational SQLite semantics. Historical Phase 1 record is preserved, not rewritten as if PostgreSQL had originally been selected. No stored SQLite data/schema exists to migrate.
- Root Go module with pinned Cilium eBPF/bpf2go tool, separate non-running agent/backend entry points, bounded environment inspector, strict JSON config, structured component logging and meaningful parser/CLI/logging/inspection tests. The normalized production event model remains documented until ABI/collector phases; no fake database repository or kernel loader added.
- Credential-free examples, Makefile and formatting script; compile-only Linux cross-builds; minimal React/TypeScript build/service-boundary scaffold, lockfile and five frontend tests. No real HTTP requests or synthetic telemetry display.
- Database-only PostgreSQL 17.11 Compose with verified tag/index digest, file-backed external password secret, loopback port, explicit durability and conservative resource settings. Runtime and SQL/role validation remain unperformed. Application driver, migrations, schema and REST remain Phase 7.
- SHA-pinned CI definitions, dependency/environment reports and a reviewed sanitized actual host report. Local coding-assistant instructions/prompts remain ignored, as previously requested; no such files are added to the published source set.

## Actual environment and evidence

Development is macOS/Darwin 27.2.0 arm64, installed bootstrap Go 1.25.0 with the project-selected Go **1.27.3** toolchain, Apple Clang 21.0.0, Node 22.18.0, npm 10.9.3 and PostgreSQL clients 17.6. Docker, bpftool, llvm-config and llc are absent on PATH. Ubuntu BTF/cgroup/tracefs/kernel/helper/privilege checks are blocked or not tested. The [machine-readable report](phase-02-environment.json) records exact command arguments/exit statuses and selected bounded output; [environment.md](../implementation/environment.md) explains limitations and target procedure.

Metadata checks resolved Cilium 0.22.0 (Go requirement 1.25.0), exact frontend versions/engine requirements and official PostgreSQL 17.11 tag/index digest. These are dependency provenance, not runtime functionality. Initial go.dev metadata lookup using curl with five-second connect timeout failed with exit 28. A subsequent official GitHub `golang/go` matching-tag query resolved stable 1.27.3; `GOTOOLCHAIN=go1.27.3 go version` successfully downloaded it. The project toolchain/.go-version and CI were updated to 1.27.3 while preserving the 1.25 language minimum, then the suite was rerun. No vulnerability scan is implied. A previous broad metadata command exceeded its 60-second bound after returning Go/npm metadata; later commands were separately bounded. No dependency installation was performed through system package managers.

## Commands and outcomes

| Actual command | Outcome |
|---|---|
| `go mod tidy` | Passed; generated a coherent tool dependency graph and go.sum |
| `go mod verify` | Passed: all modules verified |
| `go mod tidy -diff` | Passed: no graph changes |
| `go tool bpf2go -help` | Passed; tool version resolution/help only, no C compilation |
| `make check` | Passed: format, Go tests/vet/race, native build, Linux amd64/arm64 cross-build and frontend build/tests |
| `go test ./...` / `go vet ./...` / `go test -race ./...` | Passed within make check |
| `go test -cover ./...` | Passed; envcheck CLI 85.7%, shared CLI 95.8%, config 95.3%, environment 94.3%, logging 100%; trivial main wrappers 0% instrumentation, no overall coverage claim |
| `bin/argus-agent --version`, `bin/argus-backend --version` | Passed: deterministic 0.1.0-dev / phase2 |
| Both commands with `--check-config --config configs/<component>.example.json` | Passed syntax/settings checks only; no secret reads or connections |
| Both commands without flags | Expected exit 1 with structured `not_implemented`; no service starts |
| `npm ci && npm test && npm audit --audit-level=low` in web | Passed; build/typecheck, 5/5 tests, zero audit findings |
| `make envcheck` | Passed report generation; strict_passed=false, monitoring_ready=false |
| `bin/argus-envcheck --strict --output reports/environment-strict.json` | Expected exit 1 on macOS; the combined shell validation command ended nonzero only because of this deliberate strict check |
| `pg_isready -h 127.0.0.1 -p 5432 -t 2` | Exit 2: no response; no PostgreSQL runtime claim |
| Pinned `actionlint v1.7.12` on both workflow files | Passed with no diagnostics |
| Ruby YAML source parsing and Compose source invariants | Passed source checks; not Docker Compose rendering/runtime |
| `sh -n scripts/check-format.sh` | Passed |
| `git diff --check` | Passed at validation checkpoint; no commit or push performed |
| Python documentation/link/JSON checks | Passed for 29 published Markdown documents, 150 local file links, seven Mermaid blocks and two JSON examples at the documentation checkpoint; no rendering verification |

An initial Go suite exposed that the standard flag parser invokes Usage on malformed flags, printing help to stdout. The CLI now prints usage only for explicit help and uses fixed safe stderr errors otherwise; existing invalid-flag regression tests and the complete suite pass. Environment strict checks were also tightened: selected Docker must use the standard local Unix socket without context/host overrides and match local kernel/architecture; Node/npm engine checks and helper-per-program negative tests prevent false-ready reports. Fixtures remain labelled tests, not host observations.

## Skipped/blocked validation

No Ubuntu kernel feature/load/attach/verifier tests, Docker/Compose pull/render/start/health/authentication checks, PostgreSQL role/schema/durability/backup/restore tests, REST integration, browser integration, Go vulnerability scan, container image scan or monitoring/detection/overhead benchmark ran. CI definitions were linted locally; no GitHub Actions execution is claimed. Apple Clang is not a verified BPF compiler. Go/Node advisory status and target Linux tool package pins require review before deployment.

## Phase 2 checkpoint and next actions

Portable artifact/tests are ready for review; full acceptance is not. Obtain the actual Ubuntu 24.04 target, provision a reviewed BPF toolchain and rootful Docker, run the documented read-only/explicit privileged verification procedure and archive its results. Render and validate the PostgreSQL development deployment when Docker is available. Resolve security patch/tooling and service privilege questions, then conduct the Phase 2 acceptance review. Do not proceed to Phase 3 without separate approval. A Git checkpoint may be requested after review; none was created by this task.
