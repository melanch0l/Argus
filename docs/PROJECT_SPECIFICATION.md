# Project specification

**Title:** Design and Implementation of an eBPF-Based Container Runtime Security Monitoring System Using Go

**Type:** University thesis and practical software implementation

**Baseline:** Revision 2 — PostgreSQL 17 replaces the initial SQLite design by explicit user authorization. Phase 2 portable foundation is approved and implemented for local validation; Phase 2 acceptance remains pending actual Ubuntu environment evidence. Phase 3 and later phases are not approved.

## 1. Overview and research problem

Containers share the Linux kernel. Namespaces and cgroups isolate resources but do not by themselves provide comprehensive container-aware security visibility. This project investigates whether selected eBPF observations, Go processing and validated Docker/cgroup correlation can support explainable behavioral alerts with measurable collection reliability and overhead. Claims about existing approaches require literature evidence, not assumed superiority.

## 2. Objectives and research questions

Design and evaluate a modular architecture; collect process, file, network and privilege observations; correlate Docker identities; implement configurable behavioral detection; persist events and alerts; expose a Go REST API and React dashboard; measure detection effectiveness, attribution accuracy, latency, loss and overhead.

- RQ1: How effectively can eBPF observe selected runtime activities in Docker containers?
- RQ2: How accurately can observations be associated with originating containers?
- RQ3: Can contextual rules detect suspicious behavior with acceptable false positives?
- RQ4: What CPU, memory and latency overhead does monitoring introduce?
- RQ5: How does event volume affect reliability and pipeline performance?

See `research/` for operational definitions and scope.

## 3. Architecture and technology

Two applications: a host monitoring agent and a separate management service. The agent owns kernel loading, collection, validation, normalization, identity resolution, detection, health and authenticated transmission. The management service owns persistence, inventory, rule distribution, REST and the dashboard. The browser never accesses the Docker socket or kernel interfaces.

Target: Ubuntu 24.04 LTS; discover the running kernel, architecture, BTF, cgroups, hook/helper support and permissions. Initial compatibility target is rootful Docker with cgroup v2; cgroup v1, rootless Docker and other runtimes are advanced assessments, not promised support.

Stack: eBPF C, Clang/LLVM, Cilium eBPF, bpf2go, BTF/CO-RE where supported, Go, Docker Engine API, PostgreSQL 17, React/TypeScript, REST and Docker Compose for management components where appropriate. Run the agent as a host service. Phase 2 pins Cilium eBPF/bpf2go v0.22.0, the Go 1.27.3 toolchain (language minimum 1.25.0) and exact frontend dependencies/lockfile. PostgreSQL development image tag and digest are pinned in `../deploy/compose.yaml`; metadata verification is not database runtime evidence. See `implementation/dependency-matrix.md` for verified versions, purposes, security-update caveats and deferred Linux toolchain checks. pgx v5 is the intended driver; its exact pin and implementation remain Phase 7.

## 4. Functional requirements

| ID | Requirement | Priority |
|---|---|---|
| FR-01 | Load and attach eBPF programs | Critical |
| FR-02 | Capture process execution | Critical |
| FR-03 | Capture process creation and exit | High |
| FR-04 | Capture selected file activity | High |
| FR-05 | Capture selected network connection events | High |
| FR-06 | Capture selected privilege-related events | High |
| FR-07 | Decode kernel events in Go | Critical |
| FR-08 | Identify Docker containers using validated metadata | Critical |
| FR-09 | Enrich events with container metadata | Critical |
| FR-10 | Analyze events using configurable rules | Critical |
| FR-11 | Generate explainable alerts | Critical |
| FR-12 | Store events and alerts | Critical |
| FR-13 | Expose versioned REST endpoints | High |
| FR-14 | Display an actual-data dashboard | High |
| FR-15 | Filter by container and time | High |
| FR-16 | Configure detection rules | Medium |
| FR-17 | Export security reports | Medium |
| FR-18 | Account for measurable loss and pipeline health | High |

MVP and full-thesis coverage are distinguished in `research/scope.md`; an early vertical slice does not satisfy every requirement.

## 5. Non-functional requirements

| ID | Requirement | Verification intent |
|---|---|---|
| NFR-01 | Modular, maintainable code | Boundary review and unit tests |
| NFR-02 | Low overhead | Repeated baseline/monitor/detection trials; budget set before evaluation |
| NFR-03 | Reliable collection | Controlled load and stage-specific loss accounting |
| NFR-04 | Graceful shutdown | Cancellation and bounded drain tests |
| NFR-05 | Malformed-event safety | Invalid length/version and fuzz tests |
| NFR-06 | Configurable retention | Expiry, disk-pressure and referential-integrity tests |
| NFR-07 | Secure API communication | Authentication, authorization, transport and negative tests |
| NFR-08 | Reproducible builds | Pinned tools, lockfiles and clean-build checks |
| NFR-09 | Unit/integration testing | Portable tests plus privileged Linux integration suite |
| NFR-10 | Documented compatibility | Actual feature probes and per-kernel support report |

## 6. Modules and event contract

M1 kernel monitoring; M2 collector; M3 bounded pipeline; M4 Docker resolver; M5 process, M6 file, M7 network and M8 privilege capabilities; M9 detection; M10 storage; M11 REST; M12 dashboard; M13 health. The agent maintains a bounded durable outbox and saved validated rule/lifecycle state for restart recovery; this is not the backend's authoritative history. Capacity exhaustion, corruption and unsaved observations remain explicit failure/loss cases.

Normalized events contain UUID, schema version, occurrence and ingestion timestamps, event type/action/source, host and boot identity, PID/TID, optional PPID, UID, cgroup ID, nullable container metadata, explicit identity status, command name and typed event details. Preserve kernel monotonic time and conversion uncertainty. Do not guess unresolved identity. Detailed contracts are in `architecture/event-schema.md`.

## 7. Security detection

Configurable contextual rules cover unexpected shells/processes, sensitive file access/modification, privilege changes, outbound destinations and bounded process/network correlation. Alerts include rule ID/version, severity, event evidence, explanation, timestamps and known container identity. Use allowlists, deduplication and positive/negative tests. Monitoring is passive; there is no automatic workload termination or enforcement.

## 8. Persistence and API

PostgreSQL 17 entities: hosts, containers, optional processes, events, alerts, detection_rules and agent_status, with explicit epoch/lifecycle/evidence/rule-revision/receipt records. Phase 7 will supply migrations, parameterized SQL, bounded pools/admission/transactions, indexed keyset queries and configurable retention; Phase 2 does not implement storage repositories or SQL migrations. Use `READ COMMITTED` with per-host row locks and separately locked global rule publication, unique/FK constraints and race-safe immutable idempotency. Bound complete-transaction deadlock/serialization retries under the original deadline.

Use native `uuid`, checked `NUMERIC(20,0)` for uint64 values 0..18446744073709551615 (wire decimal strings unchanged), and UTC microsecond `timestamptz` projections plus exact nanosecond remainders for event/detection time. Preserve canonical JSON bytes/text and SHA-256 separately from optional JSONB query projections. Never round identifiers through floating point or hash a JSONB serialization. Keep event time and backend ingestion time distinct.

Require `fsync`, `full_page_writes` and `synchronous_commit` on, separate least-privileged migration/application roles, remote `sslmode=verify-full`, operational pool/WAL/vacuum/disk budgets and tested `pg_dump`/restore; optional physical PITR needs separate validation. Development Compose starts only a database with a bootstrap administrator, not the future application role. No PostgreSQL runtime or backup test has passed merely because this configuration exists. Database design and retention/evidence semantics are normative in `architecture/database-design.md`.

Every endpoint is under `/api/v1`: health; hosts; containers; events and event detail; alerts and alert detail; statistics; rules, rule detail and rule updates; agent event ingestion and heartbeat; scoped agent rule retrieval (`GET /agent/rules`). Protect management and ingestion, paginate, validate and bound inputs. `POST /agent/events` atomically ingests immutable events and alerts with idempotent per-object IDs. Heartbeat synchronizes bounded complete inventory snapshots and health using registered epochs/leases. Agents retrieve whole immutable rule revisions, validate before activation and report applied versus desired state. Concrete contracts are in `architecture/api-specification.md`. FR-17 export is not part of the initial endpoint baseline; its bounded authorized contract must be approved during Phase 7 and completed by Phase 9 or explicitly removed through scope review.

## 9. Dashboard

Overview, container inventory, process, file, network, alerts, detection rules and agent health. Use actual backend data only, with filters, pagination, loading/error/empty/stale states. Never display fabricated metrics as collected data.

## 10. Evaluation

Use ground-truth workloads and repeated experiments to measure precision, recall, detection and false-positive rates, container attribution accuracy, unresolved rate, CPU/memory overhead, collection latency and event loss. Compare disabled, monitoring-only and monitoring-with-detection conditions. Record environment, versions, seeds, workloads and deviations; publish actual measurements and limitations, not invented results. See `experiments/evaluation-plan.md`.

## 11. Phases and success

1. Requirements and architecture (documentation only).
2. Environment and foundation.
3. Kernel monitoring, with only the minimal loader/reader harness needed for validation.
4. Production Go collector and bounded agent pipeline.
5. Container identity resolution.
6. Detection engine.
7. Storage and REST API.
8. React dashboard.
9. Integration and evaluation.
10. Thesis documentation.

Success requires real kernel collection, validated attribution, documented process/file/network/privilege coverage, explainable alerts, durable records, a real-data dashboard, completed integration/performance experiments and reproducible results. All implementation claims require executed tests.

Phase gates and acceptance evidence: `implementation/roadmap.md`. The user explicitly approved Phase 2 portable foundation and PostgreSQL. Actual Ubuntu BTF/hooks/helpers/permissions/Docker validation is still required for Phase 2 acceptance; macOS tests and Linux cross-builds do not establish runtime compatibility. Do not start Phase 3 without separate approval.

### Current Phase 2 implementation boundary

Separate `cmd/agent` and `cmd/backend` provide help, version and bounded strict JSON configuration checking only. Default startup explicitly reports not implemented; there is no monitoring, collector, resolver, detector, database connection, REST server or production pipeline. `cmd/envcheck` produces bounded actual environment reports, with optional explicit privileged bpftool feature probing on Linux. The React/TypeScript foundation displays a static implementation-status notice, not a dashboard or synthetic monitoring data. Build/CI tooling, example configuration, PostgreSQL-only Compose and environment/dependency records support later approved phases. The normalized event contract remains documentation until the appropriate ABI/collector phases; no incomplete production model is substituted.

## 12. Primary technical references

These are reference entry points, not a claim that their current contents or compatibility have been verified during this documentation phase.

- [Cilium eBPF getting started](https://github.com/cilium/ebpf/blob/main/docs/ebpf/guides/getting-started.md)
- [Linux BPF documentation](https://kernel.org/doc/html/latest/bpf/)
- [Linux BPF ring buffer](https://kernel.org/doc/html/latest/bpf/ringbuf.html)
- [Ubuntu 24.04 release notes](https://documentation.ubuntu.com/release-notes/24.04/)
- [Docker Engine API](https://docs.docker.com/reference/api/engine/)
- [PostgreSQL 17 documentation](https://www.postgresql.org/docs/17/)
- [pgx v5](https://github.com/jackc/pgx)

Record exact consulted versions, access dates and stable citations when conducting the literature review and implementation validation.
