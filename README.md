# eBPF-Based Container Runtime Security Monitoring

**Thesis:** Design and Implementation of an eBPF-Based Container Runtime Security Monitoring System Using Go

**Current status:** Phase 2 portable foundation implemented and locally tested; acceptance is pending actual Ubuntu 24.04 environment evidence. PostgreSQL 17 replaces the initial SQLite design. Monitoring, collection, detection, persistence, REST and dashboard features are not implemented. Phase 3 is not approved.

## Start here

- [Canonical project specification](docs/PROJECT_SPECIFICATION.md)

- [Roadmap, Phase 1 acceptance checklist and unresolved decisions](docs/implementation/roadmap.md)
- [Phase 1 documentation record](docs/experiments/phase-01-record.md)
- [Phase 2 environment verification and remaining target checks](docs/implementation/environment.md)
- [Pinned dependencies and toolchain limitations](docs/implementation/dependency-matrix.md)
- [Phase 2 validation record](docs/experiments/phase-02-record.md)

## Documentation

### Research

- [Problem statement](docs/research/problem-statement.md)
- [Objectives and requirement mapping](docs/research/objectives.md)
- [Research questions](docs/research/questions.md)
- [MVP, full-thesis scope and exclusions](docs/research/scope.md)
- [Evaluation methodology](docs/experiments/evaluation-plan.md)

### Architecture

- [System architecture and deployment boundaries](docs/architecture/system-architecture.md)
- [Components, interfaces and identity resolution](docs/architecture/component-design.md)
- [Event processing and detection flows](docs/architecture/data-flow.md)
- [Normalized event, alert and kernel-record contracts](docs/architecture/event-schema.md)
- [Database entities and retention](docs/architecture/database-design.md)
- [REST API specification](docs/architecture/api-specification.md)
- [Threat model](docs/architecture/threat-model.md)
- [Technical decisions](docs/architecture/technical-decisions.md)

## Planned technology

Ubuntu 24.04 LTS, actual kernel feature probes, eBPF C and Clang/LLVM, Cilium eBPF/bpf2go, Go host agent and REST backend, rootful Docker with cgroup v2, PostgreSQL 17, React and TypeScript. The privileged agent runs on the host; management components may use Docker Compose. A macOS development session does not establish Linux support.

The system is passive: it observes selected operations and generates contextual alerts; it does not block operations or terminate workloads. Unresolved identity is never guessed or classified as host by default.

## Development roadmap

Implementation will proceed through environment verification, kernel monitoring, the Go agent, container identity resolution, behavioral detection, storage and REST APIs, the dashboard, experimental evaluation and thesis reporting. Deliverables and acceptance evidence for each stage are defined in the [implementation roadmap](docs/implementation/roadmap.md).

Phase 2 adds separate Go command boundaries, strict configuration validation, structured logging, a bounded environment inspector, locked frontend build scaffolding, CI and a PostgreSQL-only development deployment. These are foundation artifacts, not operational monitoring. Subsequent phases require explicit approval.

## Foundation quick start

Prerequisites for portable checks: the pinned Go 1.27.3 toolchain (Go language minimum 1.25.0), Node 22.18.x and npm 10.9.3. A compatible bootstrap Go installation can download the declared toolchain when `GOTOOLCHAIN=auto`. See the dependency matrix before target deployment: version pinning is not a vulnerability scan or kernel compatibility claim.

```sh
make deps
make check
make bpf-tools
make envcheck
bin/argus-agent --check-config --config configs/agent.example.json
bin/argus-backend --check-config --config configs/backend.example.json
```

`make envcheck` writes actual host evidence under ignored `reports/`; on macOS it records blocked Linux checks without claiming readiness. `--strict` fails on missing target prerequisites. [Ubuntu verification instructions](docs/implementation/environment.md) explain the separately authorized privileged feature probe. Cross-builds do not run the Linux binaries or attach BPF.

Agent/backend commands support `--help`, `--version` and settings-only `--check-config`. Running without flags exits nonzero with `not_implemented`, starts no service and opens no network/database connection. The [frontend scaffold](web/README.md) displays only an implementation-status notice; its tests use labelled fixtures, not live metrics.

[PostgreSQL development setup](deploy/README.md) describes the exact tag/digest, protected password-file requirement, localhost-only port and bootstrap administrator. Docker is not installed on the current host; Compose has not been run. Application roles, pgx, schema and repositories remain Phase 7. No SQLite data or runtime implementation exists to migrate.

## Layout

```text
cmd/agent, cmd/backend    Separate non-running command foundations
cmd/envcheck             Bounded environment inspection and JSON report
internal/                Config, logging, CLI and environment helpers
bpf/                     Phase 3 kernel source/generation boundary
configs/                 Credential-free JSON examples
web/                     React/TypeScript build and test foundation
deploy/                  PostgreSQL-only development Compose
migrations/              Phase 7 boundary; no SQL schema yet
tests/integration/       Future explicitly gated integration tests
docs/                    Canonical specification, design and actual evidence
```

CI checks portable Go/frontend builds only. It does not establish Ubuntu host eBPF compatibility, PostgreSQL functionality or thesis experimental results.

