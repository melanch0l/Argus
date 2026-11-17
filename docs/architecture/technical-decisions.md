# Technical decisions and coordination register

## Status

These are Phase 1 design decisions for **Design and Implementation of an eBPF-Based Container Runtime Security Monitoring System Using Go**, governed by the canonical [PROJECT_SPECIFICATION.md, revision 2](../PROJECT_SPECIFICATION.md). “Selected” means selected for design, not implemented, benchmarked or proven compatible. Decisions needing concrete sizes, versions or wire formats remain explicitly deferred to their owners. Phase 2 portable foundation is explicitly approved (in progress, not accepted); PostgreSQL 17 is the current canonical database choice. Ubuntu target/kernel/access and compatibility evidence remain unknown on the macOS development host; Phase 3 and later phases remain unapproved.

Related documents: [system architecture](system-architecture.md), [component design](component-design.md), [data flow](data-flow.md), [threat model](threat-model.md). Links below refer to coordinated design contracts and their revision 2 updates, not implemented monitoring behavior. Research/experiment content remains separately owned. Reading another owner's contract is not a claim of their reciprocal review or approval.

## D01 — Separate host observation from management

**Selected:** a host Go agent owns privileged kernel lifecycle, collection, Docker correlation, local detection, health and delivery. A separate Go backend owns PostgreSQL 17, rule publication/distribution and REST. React/TypeScript consumes management REST via a frontend service layer. M1–M9 are agent capabilities, M10–M12 management capabilities, and M13 covers health across both applications.

**Why:** avoid making network-facing management code root-equivalent; keep observation independent of backend availability and make module ownership testable. Separate handlers, business services and repositories.

**Alternatives not selected:** one privileged web server; browser/Docker access; production backend-only detection. Backend-only detection would add round-trip dependency and change rule/evidence timing. **Cost:** agent rule distribution and bounded retry/state complexity. Docker inspection remains a root-equivalent agent risk; component separation alone does not eliminate it.

## D02 — Probe the actual Linux environment

**Selected:** initial target is Ubuntu 24.04, cgroup v2 and rootful Docker. Discover actual kernel/architecture, BTF, attach points, helpers, program/map types and effective permissions; load/attach tests determine capability status. Separate unsupported, denied and transient failure. Process execution capture is critical to readiness; optional paths may be explicitly degraded.

**Why:** kernel versions, BTF presence and distribution labels do not guarantee usable hooks or semantics. **Alternatives not selected:** version-number-only support matrix or silent fallback. **Cost:** per-environment compatibility evidence and probe maintenance. cgroup v1, rootless Docker and other runtimes require separate assessment, not implicit support. macOS editing validates no Linux functionality.

## D03 — Cilium eBPF, generated ABI and CO-RE where supported

**Selected:** separate eBPF C and Go, use Cilium eBPF v0.22.0 and its matching bpf2go Go tool from the [Go 1.25.0 module](../../go.mod), and BTF/CO-RE where supported. Validate record ABI version/size/type, bounded strings, byte order and generated layout. Keep raw record ABI separate from the versioned normalized event schema.

**Why:** reproducible bindings and explicit compatibility are preferable to handwritten unverified casts. **Alternatives not selected:** arbitrary runtime BPF uploads, undocumented layout parsing, or claiming CO-RE makes all kernels equivalent. **Cost:** actual Linux verifier/layout tests are mandatory. The Go/Cilium/bpf2go pins are present in the foundation; actual BPF compiler support, generated ABI and Ubuntu runtime evidence remain separate checks. A module/tool pin does not prove that the Go suite or Linux probes passed.

## D04 — Passive, semantically qualified coverage

**Selected:** use reviewed observational process/file/TCP/privilege hooks with explicit source and attempted/success/state meanings. No denial return, signal, kill, container termination, network blocking or workload reconfiguration. Publish path/truncation, namespace, syscall-pairing, nonblocking-connect and credential-transition limitations.

**Why:** satisfy passive monitoring scope without mistaking hooks for a complete audit or prevention system. **Alternatives not selected:** treating syscall entry as successful operation, state transitions as task-attributed by default, or hook substitution without disclosure. **Cost:** some activity is unsupported or ambiguous; file mmap/io_uring, unselected write paths, non-TCP traffic and other privilege transitions require separate coverage work. See the candidate matrix in [components](component-design.md).

## D05 — Corroborated, lifecycle-aware Docker identity

**Selected:** correlate actual host cgroup membership/identity with Docker inspection and lifecycle/reconciliation evidence. Scope caches by host boot, cgroup/process generation and container execution interval; bound pending lookups, TTLs, history and concurrency. PID or ID-looking strings alone never establish container identity.

**Why:** PIDs/cgroups are reused and Docker stream/inspection observations race workload lifetimes. **Alternatives not selected:** `/proc/PID/cgroup` string parsing as authoritative identity, forever-valid cgroup-to-container mapping or marking failed Docker lookup as host. **Cost:** unresolved identity remains normal for short-lived containers, missing evidence and uncertain intervals. Agent/Docker restart triggers reconciliation; host reboot invalidates kernel keys. Host scope requires positive supported-environment evidence, not absence from a cache. Emitted events remain immutable.

## D06 — Versioned immutable event/evidence contracts

**Selected:** assign event/alert UUIDs once, preserve schema and source, host/boot/session context, monotonic occurrence time and uncertainty, separate backend ingestion time, typed details and explicit nullable identity. Use lossless encoding/storage for unsigned kernel IDs. Alert evidence includes rule ID/version, active set revision, facts, explanations and limitations.

**Why:** permit retry, traceability and schema evolution without conflating observation time or identity certainty. **Alternatives not selected:** PID as process identity, floating-point kernel IDs, mutable historical labels, or only unstructured alert text. **Cost:** event/API schema owners must agree on required versus unavailable fields, canonical content and compatible migrations. Concrete names and encodings are deferred to `event-schema.md` and `api-specification.md`, not independently frozen here.

## D07 — Agent-local contextual detection and immutable revisions

**Selected:** bounded declarative process/file/privilege/outbound/correlation rules with allowlists, evidence requirements, explanations and duplicate suppression. Rule edits and rollbacks create immutable assignment revisions. Agent polls its scoped rules, validates the entire snapshot and atomically activates it between evaluations. Alerts identify the captured active revision; heartbeat reports actual active revision separately from assigned/rejected revision.

**Why:** an assignment is not proof of activation; stable revisions make historical decisions reproducible. **Alternatives not selected:** executable rule plugins, partially applied updates, treating shells/admin tools as always malicious, or rerunning detection on transport retries. **Cost:** revision activation clears old correlation state and exposes warm-up; poll outage retains stale last-known-good rules. Persist the validated active manifest/revision and revalidate it on restart; no valid current or saved rules means explicit detection unavailable, not fabricated defaults. Required positive and negative tests are plans, not evidence of detection accuracy.

## D08 — Bounded memory and durable outbox with explicit loss

**Selected:** limit memory/disk records and bytes, workers, windows, cardinality, retry work/age and pending lookups. Drop newest complete work at declared admission points. Reserve bounded independent control capacity. Align with the event/API contracts: atomically persist frozen event/alert units, original epoch/sequence and dedup decisions in a bounded host-local durable outbox before transmission. Persist validated rules and lifecycle evidence for verified restart recovery; none of this is a second management database.

**Why:** hostile/high-volume workloads and backend outages must not produce unbounded memory or goroutines. **Alternatives not selected:** blocking kernel producers on REST, unbounded retries or silently discarding only alert evidence. **Cost:** overflow, pre-durable crash gaps, disk corruption/full-disk and explicit expiry lose data; correlation gaps reduce coverage. Report per-stage failure, suppression, unresolved and unknown outcomes separately. Numeric capacities/backoff/lateness limits must be chosen and recorded before load experiments; no throughput or loss-rate claims are made now.

Before implementation, review the outbox format, atomic commit/sync and power-loss assumptions, quotas/expiry, protected files, integrity checks, bounded dead letters, evidence dependency retention, restart provenance and operator cleanup. Concrete storage format/dependency and encryption policy are not chosen here. Local durability does not mean lossless kernel capture or unlimited outage tolerance; see [data flow](data-flow.md).

## D09 — Atomic events-plus-alerts ingestion with idempotency

**Selected coordination baseline:** `POST /api/v1/agent/events` carries immutable batches of normalized events **and** derived alerts. Backend commits all new events, alerts, evidence links and per-object receipts in one transaction or none. Host-scoped event/alert IDs and canonical content provide idempotency; there is no required batch ID/ledger. Same IDs/content are no-ops; conflicting IDs/content, epoch sequences or alert dedup identity reject the request. Evidence must be included or already durably present. Stable original evidence can be repeated in later batches.

**Why:** separate alert delivery can lose explainability or create dangling references. **Alternatives not selected:** event-only ingestion with unspecified alerts, a success response before commit, or new IDs on retries. **Cost:** request construction must bound evidence closure, and body/transaction limits must align. Initial one-in-flight batch per session simplifies ordering; acknowledgments, not ordering, prove persistence. Objects may be regrouped without mutation for size handling. Keep receipts throughout open epochs; close historical epochs and retain tombstones before receipt cleanup. Expired-payload replay never resurrects evidence, so pending alerts require still-retained evidence.

The guarantee is idempotent persistence for accepted retries, not end-to-end exactly-once capture. Lost acknowledgment after commit leaves the local outcome unknown; do not report it as definite backend loss. Rejected/expired batches and queue drops stay visible.

## D10 — Heartbeat inventory snapshots, revision and health

**Selected coordination baseline:** heartbeat carries host/boot/session identity, sequence, capability/health counters, active rule revision and inventory snapshot/revision/completeness. `POST /api/v1/agent/heartbeat` follows the [API specification](api-specification.md), including registered producer epochs and active host leases. Inventory revisions are host-global current + 1 across boots/restarts, while heartbeat sequences are per epoch; backend receipt time governs status staleness. New sessions do not silently overwrite newer/current sessions through delayed requests.

**Why:** events alone cannot establish current inventory, stopped containers or agent readiness. **Alternatives not selected:** infer inventory solely from latest events; treat incomplete/failed listing as empty; accept stale heartbeat rollback. **Cost:** one complete snapshot must fit the bounded heartbeat; chunking/partial inventory is not supported. Only an accepted complete authoritative snapshot may mark omitted entries inventory-absent; this does not invent exact stop times. Oversize/incomplete collection sends health-only status and preserves previous inventory as stale. Snapshots never rewrite historical event identity. Docker restart/disconnection requires reconciliation.

## D11 — Scoped pull-based rule distribution

**Selected coordination baseline:** `GET /api/v1/agent/rules` authenticates the agent and returns only assigned immutable rule-set revisions and compatibility information. Bounded periodic polling/backoff/jitter, unchanged responses and last-known-good activation semantics are required. Management rule publication stays separate from agent rule retrieval.

**Why:** a simple outbound client avoids an agent inbound control listener and cross-host rule leakage. **Alternatives not selected:** unauthenticated global rules, arbitrary push commands or “assigned revision” substituted for active revision. **Cost:** update latency depends on polling and failures. ETag/conditional retrieval, immutable manifest digests, full bounded responses and active-lease checks follow the API owner's contract; polling interval is unvalidated configuration, not an achieved propagation bound.

## D12 — Backend-owned PostgreSQL and authorized REST

**Selected:** PostgreSQL **17**, as defined in canonical revision 2. [Development Compose](../../deploy/compose.yaml) pins `postgres:17.11@sha256:3b0cda93be7c8dba2ed484d19393d93845dc3dafb88973e248ba8e6d99d784e1`. The [deployment record](../../deploy/README.md) documents official tag/index metadata verification on 2026-11-02; Docker is absent and no image was pulled/run. pgx **v5** is the intended Go driver, with exact compatible pin, pool/repository and migration implementation deferred to approved Phase 7; no fake repository or driver dependency in Phase 2.

**Decision history:** Phase 1 selected SQLite; explicit user approval changed that choice to PostgreSQL, now incorporated into canonical revision 2. The previous single-writer/file-database design is not the active persistence contract.

Use migrations, explicit unique/FK constraints, parameterized SQL, bounded connection pool/admission/transactions, indexed keyset queries and retention. Ordinary isolation is `READ COMMITTED` with per-host row locks covering lease/epoch/ingest/inventory/retention transitions; lock the global publication head separately before rule heads. Recheck predicates after locks, handle unique insert races by immutable content comparison, and bound whole-transaction serialization/deadlock retries. REST handlers validate/authenticate/authorize and call services; only repositories own SQL. Retain alert/evidence dependencies and open-epoch receipts; closure tombstones precede receipt cleanup.

Use native `uuid`, uint64 `NUMERIC(20,0)` checked over 0..18446744073709551615 (wire decimal strings unchanged), and UTC `timestamptz(6)` projections plus checked nanosecond remainders for precise occurrence/detection sorting/filtering. Preserve exact event nanoseconds and monotonic numeric values. Canonical RFC 8785 bytes/text and SHA-256 are authoritative; optional JSONB is a separate query projection, not cryptographic canonicalization. Preserve original IDs, boot/epoch and lifecycle facts across replays/migrations.

Require `fsync`, `full_page_writes` and `synchronous_commit` on, separate least-privileged migration/app roles, remote `sslmode=verify-full`, tested `pg_dump`/restore, and explicit WAL/vacuum/disk/pool/query budgets. Physical base backup/WAL archiving/PITR is optional and separately tested. See [database design](database-design.md) for lock order, representations, recovery and provisional limits.

**Why:** user-approved relational backend with concurrent transactions while keeping history outside the privileged agent. **Alternatives not selected:** shared agent/browser database access, immutable-content upserts, ad-hoc SQL or per-record nontransactional batches. **Cost:** a separate database service adds deployment, role/TLS, pool contention, autovacuum/WAL, backup/recovery and capacity management. PostgreSQL is neither unlimited capacity nor tamper-proof evidence; no performance or compatibility is claimed verified.

## D13 — Least privilege, data minimization and explicit trust limits

**Selected:** backend/frontend receive no Docker socket or BPF privilege; scope agent credentials, use TLS, bound untrusted input, protect secrets and audit rule publication. Default collection excludes file contents, network payloads, full environments and full argv. Docker labels/names are untrusted display data. Minimum agent privileges are established by actual-kernel tests, not assumed capability lists.

**Why:** the privileged local agent is an unavoidable high-value component; keep its interfaces narrow. **Alternatives not selected:** exposing Docker over REST or claiming read-only socket usage removes root-equivalent risk. **Cost:** metadata remains sensitive and host/root compromise defeats monitoring. See the full [threat model](threat-model.md) for abuse cases and residual risk. Authentication mechanism, enrollment and rotation details require API/deployment agreement.

## D14 — Bounded drain and honest health

**Selected:** quiesce producers, detach, drain readers/pipeline, finalize pending identity, flush event/alert batches until a deadline, attempt final health, then close resources/join workers. Separate stopping, liveness, critical readiness, detection availability, capability degradation, resolver staleness and backend reachability. Counter generations prevent interpreting restart as zero loss. Preserve durably queued work under its original producer epoch when the delivery deadline expires; distinguish it from unsaved/discarded work and unknown in-flight results.

**Why:** canceling all workers immediately destroys drain work; keeping them alive indefinitely makes shutdown unsafe. **Alternatives not selected:** promise zero shutdown loss or call every expired unacknowledged batch definitely lost. **Cost:** forced exit/host crash can lose both data and final counters. Graceful-shutdown and race tests must validate actual cleanup and reporting later.

## D15 — Actual-data frontend and evidence-led evaluation

**Selected:** frontend service layer uses versioned REST only; views display actual backend records with loading/empty/stale/error/permission/degraded states. Filters/pagination are server-defined. Test fixtures are explicitly labelled and isolated from production displays. Evaluate monitoring-only versus detection-enabled conditions with ground truth and repeated workloads.

**Why:** a plausible dashboard is not collection evidence, and planned tests are not results. **Alternatives not selected:** generated “live” metrics, fabricated alerts or benchmark values. **Cost:** unavailable data produces an honest empty/error display. Phase 1 does not choose performance budgets or claim precision/recall, attribution accuracy, latency, overhead or Linux compatibility.

## Cross-owner contract handoff

Contracts consulted: [API specification](api-specification.md), [event schema](event-schema.md), [database design](database-design.md). Reconciliation adopted per-object idempotency rather than a batch ledger, host-global inventory revisions and no chunking, registered epochs/leases, and bounded durable outbox/rule/lifecycle state rather than memory-only recovery. These are planned requirements, not new code or dependencies.

This register records the shared documentation baseline. Repository-level review checked the delivery, inventory and rule contracts for alignment; human acceptance and all implementation validation remain pending. Resolve disagreements before coding; do not create competing root specifications.

| Owner / coordination target | Agreement required before implementation |
|---|---|
| API specification | Keep `/api/v1`; `POST /agent/events` atomic events+alerts and idempotent IDs; heartbeat inventory snapshots/revision/health and authorized sessions; `GET /agent/rules` scoped immutable revision retrieval; authenticated management health/hosts/containers/events/alerts/statistics/rules/agent status and bounded export where required. Define wire/error/auth/pagination/body limits without weakening semantics. |
| Event schema | Raw versus normalized versioning, typed event details/hook result semantics, lossless IDs, boot/session/process-instance evidence, nullable identity/provenance, timestamps/uncertainty and immutable alert evidence/revision fields. |
| API + persistence | Per-object canonical-content receipts/commit responses, ID/sequence/dedup conflict policy, retained evidence validation, open-epoch receipts, closed-epoch replay rejection and alert-pinned evidence retention; no batch ledger or expired-payload resurrection. |
| API + resolver | Active epoch/lease authorization, per-epoch heartbeat sequence versus host-global inventory revision, one bounded complete snapshot without chunks, incomplete/oversize behavior, saved lifecycle UUID revalidation, Docker freshness and historical containers. |
| Rule management + agent | Immutable set and individual rule versions, scope, compatibility and budgets, last-known-good validation, rollback-as-new-revision, warm-up and active-vs-assigned reporting. |
| Implementation/experiments | Actual-kernel probe reports, minimum privileges, hook ground truth, queue capacities, reorder/retry/drain bounds, saturation/restart/security tests and recorded results. Version pins are selected only after compatibility review. |

## Validation status and phase checkpoint

**Historical Phase 1 outcome:** documentation only; that phase created no Go/eBPF/React code, dependency, database migration, API endpoint or benchmark. Documentation checks of links, diagram blocks and contract alignment did not establish rendered diagrams or working functionality.

**Phase 2 progress:** portable source/tooling and a static frontend scaffold now exist. The [frontend record](../../web/README.md#local-validation-record--2026-11-02) reports a successful macOS build/typecheck, five tests and zero audit vulnerabilities on 2026-11-02. The [deployment record](../../deploy/README.md) verifies the development image pin through registry metadata only, with no Docker run. These records do not imply passing Go-suite, privileged Linux, database, API or monitoring tests; record each actual outcome separately. See the [roadmap](../implementation/roadmap.md) for acceptance and unresolved decisions.

Recommended checkpoint: review the Phase 1 architecture with the API/event-schema and research/experiment owners, capture remaining contract decisions in their approved documents, and track the explicitly approved Phase 2 portable foundation to its still-pending acceptance gate. Environment/dependency reports and the Phase 2 experiment record consolidate foundation artifacts and actual command outcomes; report bootstrap is not a prerequisite stopping independent portable work. Distinguish present pins and recorded portable checks from still-pending Go-suite and actual-Ubuntu validation; obtain separate explicit approval before Phase 3. Later experiment logs should record actual environment, tool versions, commands, outcomes and deviations, not restate planned requirements as achieved results.
