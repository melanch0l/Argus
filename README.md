# eBPF-Based Container Runtime Security Monitoring

**Thesis:** Design and Implementation of an eBPF-Based Container Runtime Security Monitoring System Using Go

**Current status:** Phase 1 documentation prepared for review. No monitoring software is implemented; no kernel compatibility, detection effectiveness or performance result is established. Phase 2 has not started.

## Start here

- [Canonical project specification](docs/PROJECT_SPECIFICATION.md)
- [Agent development rules](AGENTS.md)
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

## Sequential work prompts

1. [Requirements](prompts/01-requirements.md)
2. [Environment and foundation](prompts/02-foundation.md)
3. [Kernel monitoring](prompts/03-ebpf-monitoring.md)
4. [Go agent](prompts/04-go-agent.md)
5. [Container identity](prompts/05-container-identity.md)
6. [Detection](prompts/06-detection.md)
7. [Storage and API](prompts/07-backend.md)
8. [Dashboard](prompts/08-dashboard.md)
9. [Evaluation](prompts/09-evaluation.md)
10. [Thesis](prompts/10-thesis.md)

Prompts are instructions for later approved work, not evidence those phases ran. Do not execute the next phase without explicit approval. No source tree, Go module, build tool, dependency lockfile or deployment configuration is created in Phase 1.
# Thesis
