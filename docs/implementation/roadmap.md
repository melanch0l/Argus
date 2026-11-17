# Implementation roadmap and phase acceptance gates

**Status:** Phase 1 documentation baseline prepared; human architecture acceptance remains pending. The user **explicitly approved Phase 2 portable foundation**, now **in progress, not accepted**, with PostgreSQL 17 defined in the [canonical specification, revision 2](../PROJECT_SPECIFICATION.md). The [evaluation plan](../experiments/evaluation-plan.md) controls planned measurements. Current development host: macOS; actual Ubuntu target/access/kernel and compatibility evidence remain unknown. Portable foundation artifacts and selected checks are present below; they do not satisfy the pending Ubuntu target gate. **Phase 3 and later phases remain unapproved.**

## 1. Sequential phases

| Phase | Deliverable | Required exit evidence before next approval |
|---|---|---|
| 1 — Requirements and architecture | Research, requirements, scope, contracts, diagrams, threat model and evaluation plan | Documentation checks; reviewer accepts baseline and explicit remaining decisions |
| 2 — Environment and foundation (approved; in progress, not accepted) | Bootstrap environment/dependency reports, minimal portable layout/contracts/configuration/build conventions, justified verified foundation pins; Ubuntu capability evidence when target available | Actual OS/kernel/architecture, BTF, Docker/rootful/cgroup v2, hook/helper/map/program/permission probes; version manifest; clean build; documented unsupported features |
| 3 — Kernel monitoring | Incremental process, file, network and privilege programs with versioned raw records and failure counters | Real controlled workload observations and verifier/load/attach evidence for each supported capability; malformed-record and loss tests; candidate-to-actual hook matrix |
| 4 — Go agent | Production loading/decoding, bounded pipeline, normalization, health, durable outbox and shutdown | Real kernel events; unit/fuzz/race tests where supported; saturation, restart, disk-full and bounded-drain tests; counters and delivery-uncertainty evidence |
| 5 — Container identity | Docker events/inspection plus cgroup resolver and lifecycle history | Real Docker host/container controls; short-lived, stale, restart, reconnect, PID/cgroup reuse tests; incorrect and unresolved rates reported separately |
| 6 — Detection | Validated versioned YAML rule input, immutable revisions, contextual rules, allowlists, evidence and deduplication | Positive/benign cases, absent-field semantics, correlation limits and revision swaps; false-positive evaluation with declared units; no enforcement |
| 7 — Storage and REST (unapproved) | PostgreSQL 17 migrations/indexes/retention using the reviewed development image pin, pgx v5 exact pin/implementation, bounded pool/transactions; authenticated management/ingestion; inventory and rule distribution | Concurrent transaction/rollback/retry and idempotency races; native UUID/checked uint64 and exact-time round trips/keysets; retained evidence/epochs/lifecycles; role/TLS, durability, pg_dump/restore, WAL/vacuum/disk/pool, query plan and API tests |
| 8 — Dashboard | React/TypeScript pages and service-layer API integration | Real backend integration, loading/empty/error/stale states, filters, pagination, accessibility and authorization-aware frontend tests |
| 9 — Integration and evaluation | Reproducible repeated disabled/monitor/detection trials, reports, graphs and limitations | Independent ground truth; frozen protocol/thresholds; raw data, actual commands, uncertainty and repeatability evidence; requirement-to-test outcomes |
| 10 — Thesis | Six chapters grounded in implemented artifacts and actual results | Consistent citations, reproducible figures, traceability, limitations and clear distinction between original engineering and existing eBPF capabilities |

Phase 3 permits only the minimal Go loader/reader validation harness needed to test kernel programs. Phase 4 owns the reusable production agent; this avoids claiming kernel validation without userspace or prematurely implementing later phases. Before Phase 7 exists, test transport/storage boundaries using explicitly labelled test fixtures or harnesses, not a claimed production backend. Phase 6 can load reviewed local immutable policies; Phase 7 adds authenticated distribution of the same contract.

MVP is an early execution-event end-to-end slice across phases, not permission to skip phase gates. Full thesis scope includes selected process lifecycle, file, network and privilege coverage, rule management and FR-17 report export. The initial API deliberately defers an export endpoint: define bounded authorized export/report format, retention clipping, privacy controls and tests during Phase 7 design review, implement by Phase 9, or obtain an explicit scope amendment. Do not mark FR-17 complete merely because events can be listed.

## 2. Compatibility and testing strategy

1. Record actual target OS, `uname`/architecture, kernel configuration where available, BTF availability, security lockdown/LSM settings, effective privileges, Docker version/API, cgroup mount/driver and tool versions on Ubuntu. A version number alone is insufficient.
2. For each intended hook, record required program/map types, helper availability, attach point and observed permissions. Minimal load/attach plus controlled operations establish support more strongly than symbol discovery. Distinguish unavailable, permission denied, policy denied, failed and not tested.
3. Initially require process execution and ring-buffer support; absence fails readiness. Optional-family absence produces explicit degraded capability status, never a fabricated successful observation. Full-thesis acceptance still tracks missing FRs.
4. Record hook semantics and action/outcome guarantees: successful exec versus attempt, thread versus process lifecycle, requested versus canonical path, syscall connect versus established TCP session, requested versus effective credentials. Test IPv4/IPv6 and selected syscall variants only where claimed.
5. Separate portable unit/contract tests from privileged Linux tests. Run `go test ./...` and `go vet ./...` once Go code exists; run race/fuzz tests where supported. Test TypeScript uint64 string handling and database round trips at boundary values.
6. Test kernel reservation failure, ABI rejection, queue loss, spool failure, missing correlation state, retry and backend rejection independently. Unsupported coverage and unresolved identity are not ring-buffer loss. Do not sum overlapping counters into a fictitious end-to-end loss rate.
7. Execute the ground-truth and controlled-load methodology in the evaluation plan. Set budgets and numerical acceptance thresholds before confirmatory trials; publish failures and unknowns.

## 3. Phase 1 acceptance checklist

Checked means the **document exists and addresses the design criterion**, not reviewer approval, implemented functionality or an executed Linux test.

| # | Documentation criterion | State | Evidence |
|---|---|---|---|
| 1 | Research problem and objectives | [x] Documented | [Problem](../research/problem-statement.md), [objectives](../research/objectives.md) |
| 2 | Research questions finalized as baseline | [x] Documented | [Questions and operational mapping](../research/questions.md) |
| 3 | Functional requirements | [x] Documented | [Specification §4](../PROJECT_SPECIFICATION.md#4-functional-requirements), FR-01–FR-18 |
| 4 | Non-functional requirements | [x] Documented | [Specification §5](../PROJECT_SPECIFICATION.md#5-non-functional-requirements), NFR-01–NFR-10 |
| 5 | System architecture diagram | [x] Documented | [System architecture](../architecture/system-architecture.md) |
| 6 | Agent/management boundaries | [x] Documented | [Component contracts](../architecture/component-design.md) |
| 7 | Common event schema | [x] Documented | [Event, alert and raw-record contracts](../architecture/event-schema.md) |
| 8 | Database design | [x] Documented | [Entities, ER diagram and retention](../architecture/database-design.md) |
| 9 | REST API specification | [x] Documented | [Authentication, routes and wire behavior](../architecture/api-specification.md) |
| 10 | Threat model and security boundaries | [x] Documented | [Threat model](../architecture/threat-model.md) |
| 11 | MVP versus advanced scope | [x] Documented | [Scope and full-thesis requirements](../research/scope.md) |
| 12 | Kernel compatibility/testing strategy | [x] Documented | §2 above, [technical decisions](../architecture/technical-decisions.md), [evaluation](../experiments/evaluation-plan.md) |

**Documentation coverage: 12/12.** Human architecture acceptance: pending. Mermaid syntax/rendering requires visual review; fenced diagram presence alone is not rendering verification. Phase 2 approval: **explicitly granted by the user**, independently of this historical documentation checklist. Approval is not Phase 1 reviewer acceptance, a passed Ubuntu check or Phase 2 exit acceptance.

## 4. Phase 2 acceptance checklist — portable progress, target gate pending

Read the canonical revision 2 specification and existing project rules, architecture and scope first. The [environment report](environment.md), [dependency matrix](dependency-matrix.md) and Phase 2 experiment record are maintained Phase 2 outputs: use their actual command outcomes and distinguish observed, blocked, deferred and untested results. Bootstrap or updates to those records must not block independent portable foundation work. Missing target access blocks dependent tests/acceptance, not all scaffolding. Never substitute a macOS Docker client/VM, Linux cross-build or stub build for actual Ubuntu kernel evidence.

Present artifacts include separate `cmd/agent`, `cmd/backend` and `cmd/envcheck` entry points; portable configuration/CLI/environment helpers; example configuration; formatting/build/CI definitions; a static frontend scaffold and locked toolchain; and PostgreSQL-only development Compose. Their presence is not production collection, REST, persistence or dashboard functionality. Detailed evidence is owned by the environment/dependency reports and Phase 2 experiment record; this checklist summarizes without independently rerunning or pre-approving their checks.

| Criterion | Current status / evidence needed |
|---|---|
| Approval and phase scope | Explicit Phase 2 approval recorded; in progress, not accepted. Phase 3 not approved. |
| Development versus target environment | macOS host known; Ubuntu host/VM/access, architecture, kernel, Docker/cgroup/BTF and effective privileges **unknown/pending target**. |
| Phase 2 records | Environment/dependency reports and Phase 2 experiment record document foundation work and checks; keep command-level evidence and blockers there, not an assumption that reports must predate bootstrap. |
| Go/tool foundation | [go.mod](../../go.mod) declares Go language minimum **1.25.0** and project toolchain **1.27.3**, Cilium eBPF **v0.22.0** and matching bpf2go as a Go tool. Pins are present; the portable Go suite and tool help passed as recorded below. C generation, Clang BPF-target support and actual Linux probes remain unverified. |
| Frontend reproducibility | [Frontend record](../../web/README.md#local-validation-record--2026-11-02): Node **22.18.0**, npm **10.9.3**, exact dependency manifest and npm lock present. On macOS arm64, clean install/build/typecheck and **5 tests passed**; audit reported **0 vulnerabilities** on 2026-11-02. These recorded results are not browser/API/Ubuntu validation or a security certification. |
| Portable structure and tooling | Separate entry points, portable helpers, config examples, Makefile, format script and CI workflow definitions present. No collector/resolver/detection/REST/dashboard or fake DB repository. File presence is not a passed local suite or executed remote CI run. |
| PostgreSQL development foundation | [Compose](../../deploy/compose.yaml) pins `postgres:17.11@sha256:3b0cda93be7c8dba2ed484d19393d93845dc3dafb88973e248ba8e6d99d784e1`. [Deployment record](../../deploy/README.md) verifies tag/index metadata only: **no Docker pull/run**, Compose runtime, database authentication, durability or restore test. pgx v5 exact pin, pool/repositories/migrations and DB integration remain Phase 7 work. |
| Ubuntu feature readiness | Actual OS/kernel/architecture, BTF, cgroup v2/rootful Docker, hook/helper/program/map/permission report **pending target**. Phase 2 inspects capabilities; Phase 3 load/attach observation evidence remains separately gated. |
| Go validation | **Passed locally:** `make check`, `go test ./...`, `go vet ./...`, `go test -race ./...`, module verification/graph consistency, native builds and pinned bpf2go help. Invalid-flag Usage regression fixed and tested. See the [Phase 2 record](../experiments/phase-02-record.md); no BPF program generation/load was performed. |
| Linux cross-build | **Passed compile-only:** CGO-disabled Linux amd64 and arm64 builds of all three commands. Not BPF load/attach, Docker integration or Ubuntu runtime evidence. |
| Exit review | **Not passed**. Record unresolved blockers and actual evidence; request separate Phase 3 approval only after required foundation acceptance. |

## 5. Unresolved design and validation decisions

These are explicit follow-up gates, not permission to guess during coding.

| Decision | Resolve by | Required evidence/decision |
|---|---|---|
| Actual Ubuntu host, CPU architecture and kernel support matrix | Phase 2 exit | Reproducible probe report from the deployment host; no Linux inference from macOS |
| Go/Cilium/bpf2go and frontend pin execution evidence; remaining Clang/LLVM and Linux tools | Phase 2 exit | Go toolchain 1.27.3 (language minimum 1.25.0)/Cilium 0.22.0/matching bpf2go and frontend pins/lock present; frontend local results recorded. Go validation passed locally as recorded; actual BPF compiler/kernel checks and target security patch review remain pending |
| PostgreSQL runtime compatibility, pgx v5 exact pin and migration tools | Phase 7 implementation | Use the existing metadata-verified PostgreSQL 17.11 tag/index digest; review security re-pins via deployment procedure. Runtime server/client/module, schema and recovery tests remain pending; no DB driver/repository added in Phase 2 |
| PostgreSQL transaction, capacity and recovery budgets | Phase 7 exit | READ COMMITTED + host/global publication locks and unique/FK tests; bounded retries/pool/admission; durability/roles/remote verify-full; pg_dump/restore and optional PITR separately; WAL/autovacuum/disk/query-plan evidence |
| Minimal service capabilities and Docker socket access model | Phase 2 exit | Measured load/attach/inspection requirements; explicit root-equivalent Docker risk; service-hardening plan |
| Exact C ABI sizes/offsets/alignment, byte order, string bounds and hook-specific details schemas | Before each Phase 3 capability | Generated layout assertions, decoder fixtures and verifier/runtime evidence; conceptual record header is not final ABI |
| Required syscall variants, TCP hooks and fallback coverage | Phase 3 acceptance | Actual-kernel semantics/ground truth including file paths, asynchronous network outcomes and credential changes |
| Resolver mapping algorithm limits and verified-host classification policy | Phase 5 acceptance | Lifecycle interval/reuse/reconciliation evidence; unresolved is default when proof is insufficient |
| Queue, map, cache, outbox, state persistence and shutdown budgets | Phases 3–7, frozen before evaluation | Explicit record/byte/time limits, disk failure behavior and controlled saturation/recovery tests |
| Baseline policy contents, capability compatibility and acceptable FP/detection thresholds | Phase 6, before confirmatory evaluation | Reviewed benign/suspicious scenarios, units, revision fixtures and pre-registered targets |
| Credential bootstrap/rotation/revocation, browser token issuance and deployment TLS procedure | Before Phase 7 exposure | Operator runbook and negative security tests; no API-supplied agent enrollment trust |
| Retention/retry epoch closure and evidence pinning limits | Phase 7 exit | Bounded storage and operator-visible expiry; no silent evidence deletion or replay resurrection |
| Report export contract for FR-17 | Phase 7 design review | Approved bounded output format/interface, authorization/redaction and completion tests by Phase 9 |
| API provisional body/rate/lease/response bounds | Phase 7 exit and Phase 9 | Capacity/security tests; limits currently design assumptions, not benchmarked capacities |
| Trial sizes, workload rates, latency clocks/uncertainty, resource and reliability budgets | Before Phase 9 confirmatory trials | Versioned pre-registration; pilot and confirmatory runs separated |

Future changes must keep the canonical revision 2 specification, architecture contracts and evidence records consistent, with explicit approval for behavior or scope changes. Keep unresolved support visible rather than lowering success criteria after seeing results.

## 6. Recommended phase checkpoint

Review these checklists and open decisions; Phase 2 authorization is already explicit, but record portable work, actual Ubuntu evidence and acceptance separately. Request separate approval before Phase 3; do not infer it from ongoing foundation work. Keep a short [phase record](../experiments/phase-01-record.md). A Git checkpoint is recommended only after the user requests one; this task does not initialize Git, create a branch or commit. Subsequent logs must include actual commands, outcomes, artifacts and skipped checks.
