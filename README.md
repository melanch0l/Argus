# eBPF-Based Container Runtime Security Monitoring

**Thesis:** Design and Implementation of an eBPF-Based Container Runtime Security Monitoring System Using Go

**Current status:** Phase 1 documentation prepared for review. No monitoring software is implemented; no kernel compatibility, detection effectiveness or performance result is established. Phase 2 has not started.

## Start here

- [Canonical project specification](docs/PROJECT_SPECIFICATION.md)

- [Roadmap, Phase 1 acceptance checklist and unresolved decisions](docs/implementation/roadmap.md)
- [Phase 1 documentation record](docs/experiments/phase-01-record.md)

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

Ubuntu 24.04 LTS, actual kernel feature probes, eBPF C and Clang/LLVM, Cilium eBPF/bpf2go, Go host agent and REST backend, rootful Docker with cgroup v2, SQLite, React and TypeScript. The privileged agent runs on the host; management components may use Docker Compose. A macOS development session does not establish Linux support.

The system is passive: it observes selected operations and generates contextual alerts; it does not block operations or terminate workloads. Unresolved identity is never guessed or classified as host by default.

## Development roadmap

Implementation will proceed through environment verification, kernel monitoring, the Go agent, container identity resolution, behavioral detection, storage and REST APIs, the dashboard, experimental evaluation and thesis reporting. Deliverables and acceptance evidence for each stage are defined in the [implementation roadmap](docs/implementation/roadmap.md).

Phase 1 contains design documentation only. Subsequent phases require explicit approval. No source tree, Go module, build tool, dependency lockfile or deployment configuration has been created.
# Thesis
