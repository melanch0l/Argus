# Normalized events, kernel records and alerts

**Status:** Phase 1 design, contract revision 1. No collector, hook, ABI layout, compatibility or capacity has been implemented or verified. The authority remains [PROJECT_SPECIFICATION.md](../PROJECT_SPECIFICATION.md). Related contracts: [database](database-design.md), [API](api-specification.md).

## 1. Boundaries and versioning

The kernel emits a small binary observation. The privileged Go agent validates it, resolves identity, estimates wall time, normalizes it and evaluates rules. The separate backend validates and persists submitted normalized events and alerts. Kernel memory layout is **not** the JSON or database schema; browsers never decode kernel records.

Three independent versions are required:

- `abi_version`: integer identifying a kernel record layout and interpretation.
- `schema_version`: normalized event or alert version, initially `1.0` (each document type has its own validator).
- Rule `revision` and `rule_schema_version`: immutable policy content revision and policy language version; neither is an event schema version.

API `/api/v1` is a fourth, independently managed compatibility boundary. A major event schema change requires explicit readers/migrations or rejection; do not silently reinterpret stored events. Minor versions may add optional fields only after validators and capability negotiation support them. Initial v1 validators accept exactly `1.0`, reject unknown fields and unknown enum values, and never discard unsupported fields. Rule distribution advertises supported schema versions. Database migration versions are independent of all of these.

## 2. Encoding and validation

- UTF-8 JSON objects; reject duplicate keys, invalid UTF-8, trailing documents, excessive nesting and non-finite numbers. Nullable means JSON `null`, not `0`, an empty string or an omitted error.
- Every unsigned 64-bit value, including kernel nanoseconds, cgroup IDs, counters, sequences, rule/inventory revisions and namespace inode IDs, is a **decimal JSON string** matching `0|[1-9][0-9]{0,19}`, range-checked through `18446744073709551615`. Store it as SQLite `TEXT`; never route it through JavaScript `Number`, floating point or a signed SQLite integer. No hexadecimal, sign or leading zeros. Signed 64-bit syscall results also use range-checked decimal strings (optional minus sign, no negative zero).
- UUIDs use canonical lowercase hyphenated form. Host enrollment assigns `host_id`; Linux boot UUID is `boot_id`. UUID uniqueness does not replace host-scoped authorization.
- Finite bounded 32-bit IDs, ports, enum values and small counts use JSON integers: PID/TID/PPID and UID/GID are unsigned 32-bit, ports 0–65535, CPU ID unsigned 32-bit. Unknown values are null. Do not interpret PID zero as missing.
- All wall timestamps use UTC RFC 3339 with `Z`, up to nine fractional digits. The backend canonicalizes them to exactly nine digits for storage and comparison; representable design range is 1970-01-01 through 9999-12-31. No local times or leap-second spelling. Monotonic values are not UTC timestamps.
- Strings have explicit byte limits after UTF-8 encoding. Linux paths/command names may not be UTF-8: use escaped display text plus `encoding_loss=true`, not fabricated exact bytes. Do not claim a display string is an exact canonical path. Missing and truncated observations remain explicit.

**Provisional design bounds, not measured capacities:** event 16 KiB canonical JSON; alert 32 KiB; source hook 128 bytes; command display 256 bytes; path display 4096 bytes; explanation 2048 bytes; at most 32 evidence references/alert; JSON nesting depth 12. The wire request bounds in the API also apply. The agent records and counts oversize rejection rather than silently truncating whole records. Specific observed strings can be truncated only with their corresponding flag set. File contents, network payloads, full environment and full command arguments are not collected by default.

## 3. Event contract v1.0

All fields below are required unless marked optional. Conditional fields use explicit nulls where applicable. An event is frozen before first transmission; late enrichment does not mutate its submitted identity or timestamps.

| Field | Type / meaning |
|---|---|
| `schema_version` | String `1.0` |
| `event_id` | Agent-generated UUID, assigned once before durable enqueue; reused across retries |
| `host_id`, `boot_id`, `agent_epoch_id` | UUIDs: enrolled host, observed Linux boot, and agent process instance respectively |
| `sequence` | uint64 string, strictly increasing within the agent epoch; gaps permitted, not a count of kernel loss |
| `event_type` | `process`, `file`, `network`, `privilege` |
| `action` | Typed action from §6, not a broad claim inferred from a program name |
| `source` | Object: `kind` (`tracepoint`, `fentry`, `fexit`, `kprobe`, `kretprobe`, `lsm`), actual `hook` string, `abi_version` integer, `observation_point` (`entry`, `exit`, `lifecycle`, `decision`), `coverage_id` (UUID of agent capability/attachment manifest) |
| `time` | Object specified in §5 |
| `process` | Object with host-namespace `pid`, `tid`, nullable `ppid`, `uid`, `gid`, nullable uint64-string `start_monotonic_ns`, nullable `comm` display string, `comm_truncated` and `encoding_loss` booleans. Optional namespace PID/UID values require explicit namespace identifiers; do not substitute them for host IDs. |
| `identity` | Object specified in §4 |
| `outcome` | Object: `status` (`attempted`, `succeeded`, `failed`, `unknown`), nullable signed-int64-string `return_value`, nullable positive 32-bit Linux `errno`, nullable bounded `reason` |
| `details` | Exactly the object permitted for `event_type` and `action`; no unbounded free-form metadata |

The submitted object has no trusted `ingested_at` field. Read responses wrap it as `{ "event": <immutable submitted object>, "ingested_at": <backend UTC time> }`. `ingested_at` is assigned once at the successful first backend commit and remains unchanged on retry. There is no claim of exactly-once collection: delivery is at least once with idempotent persistence, subject to observable loss and bounded buffers.

### Identifier lifecycle

`host_id` is enrollment identity, not a hostname, IP, Docker ID or image-cloned machine ID. Enrollment must detect duplicated registrations and provision a new ID when appropriate. `boot_id` separates kernel monotonic domains and PID/cgroup reuse across reboots. A fresh `agent_epoch_id` is registered at every agent process start, even within the same boot; a single active lease prevents competing agents from replacing inventory. Event replay from a registered historical epoch is distinct from permission to publish current inventory (see API).

Generate event IDs once, persist the frozen record and sequence before attempting delivery, and recover that same record after a crash. If a durable queue is unavailable/full, count the loss and report degraded health; do not invent records. `(host_id, agent_epoch_id, sequence)` must also be unique. Events retained in an old queue keep their old boot/epoch, never the new process's identifiers.

## 4. Identity and lifecycle

`identity` contains:

| Field | Meaning |
|---|---|
| `status` | Exactly `resolved`, `host` or `unresolved` |
| `cgroup_id` | Nullable uint64 string from the observation |
| `cgroup_version` | `2`, or null if unknown; v1 requires a separately assessed contract/coverage claim |
| `container` | Null unless resolved; otherwise `{runtime:"docker", container_id, lifecycle_id, name, image_id, image_reference}`; the last three metadata fields can be null |
| `resolution_method` | `docker_cgroup`, `verified_host`, `unresolved` |
| `reason` | Nullable bounded reason code, e.g. `metadata_unavailable`, `stale_mapping`, `lifecycle_ambiguous`, `unsupported_cgroup` |
| `mapping_revision` | Nullable uint64 string identifying the resolver snapshot used |
| `resolved_at` | Nullable UTC timestamp of the identity decision |

A resolved Docker ID is the full lowercase 64-hex ID. Its identity key is `(host_id, runtime, container_id)`. A running incarnation key is `(host_id, runtime, container_id, lifecycle_id)`: restart keeps a container ID but creates a new lifecycle UUID. Resolver lifecycle records bind Docker inspection start/finish metadata and cgroup observation to `boot_id`; retain historical intervals. Persist lifecycle UUIDs across agent restarts and revalidate the same incarnation before reusing them; an agent restart alone does not create a container restart. Do not use Docker's restart counter alone as a durable lifecycle identifier. Ambiguous inspection, missed transitions or uncertain interval boundaries mean unresolved rather than a guessed incarnation.

`resolved` requires corroborated cgroup membership and Docker inspection/lifecycle evidence valid at observation time, not PID alone, a matching ID-looking substring or the current owner of a reused cgroup. `host` requires positive evidence of host scope under a documented resolver policy; failure to resolve a Docker container is never host evidence. `unresolved` preserves available raw cgroup/process facts and sets `container=null` and `resolution_method=unresolved`. Resolved records require a non-null cgroup ID, lifecycle ID and matching mapping revision; host records require `container=null` and `verified_host`.

Identity mapping is boot-scoped, time-bounded and invalidated on lifecycle changes, daemon disconnect, stale cache, cgroup reuse or agent restart until revalidated. Short-lived containers may remain unresolved. Events contain historical metadata snapshots; a later container rename or inventory snapshot cannot rewrite them. A future enrichment feature would be a separately versioned annotation, not mutation of the original event. No unverified hostname/PID heuristic substitutes for identity.

## 5. Time and uncertainty

`time` is:

- `monotonic_ns`: uint64 string observed in kernel time, or null only for an explicitly supported non-kernel source (none in initial v1).
- `clock`: initially `bpf_ktime_get_ns`, the boot-relative monotonic domain excluding suspend; do not equate it to `CLOCK_BOOTTIME`. A different helper/clock requires a declared compatible mapping.
- `observed_at`: UTC estimate of occurrence, or null if no valid conversion exists.
- `agent_received_at`: agent UTC time when userspace decoded the observation; not event occurrence and not backend ingest.
- `calibration_id`: nullable UUID of an agent clock calibration record.
- `uncertainty_ns`: nullable uint64 string bounding the estimated wall conversion error under stated assumptions; null means unbounded/unknown, not zero.
- `quality`: `calibrated`, `uncertain` or `unavailable`.

The agent brackets monotonic/realtime samples and records calibration bounds, sample interval, clock/helper domain and wall-clock steps in its capability/clock manifest. A wall-clock step, suspend/resume or stale calibration invalidates the previous conversion bound until recalibration. `calibrated` requires non-null calibration, observed time and uncertainty. `uncertain` may provide an estimate but cannot assert a finite bound without evidence. `unavailable` requires null occurrence estimate and uncertainty. A conversion bound is relative to the host realtime clock, not proof of its absolute UTC accuracy; record synchronization status and absolute offset uncertainty separately, with unknown as the default. Agent/backend wall-clock skew can make naive latency negative; report unavailable/uncertain latency, never clamp it into a supposedly valid measurement. Immutable clock/capability records referenced by retained events must survive heartbeat-history pruning; publish those records through authenticated heartbeat before their events.

Use `(host_id, boot_id, monotonic_ns)` for relative ordering/durations only within the same clock domain, with CPU/concurrency caveats; sequence is agent processing order, not a total kernel order. Cross-host correlation uses wall-time intervals plus uncertainty, not raw monotonic values. Filters explicitly select event occurrence time or backend ingestion time. Unknown occurrence times do not silently become receive/ingest times. Backend ingest measures first persistence, not a kernel timestamp.

## 6. Hook-dependent actions and typed details

These are candidate observation semantics, **not verified attachments or promised coverage**. The agent must report the actual hook, runtime probe result and attachment manifest. Hook absence disables that capability; fallback observations retain their own semantics. Never relabel a syscall entry as a successful operation.

| Type / action | Candidate hook and evidence | Details and limitations |
|---|---|---|
| process / `exec` | `sched:sched_process_exec` lifecycle observation; syscall exec entry/exit are separate alternatives | Nullable executable path display with truncation/encoding flags; successful exec tracepoint means image transition observed. Entry means attempted; failure requires exit result. No default argv collection. |
| process / `fork` | `sched:sched_process_fork` | Parent/child host PID and nullable start times; creation observed, not proof of successful later execution. Child thread/process semantics must be recorded as `creation_kind` (`process`, `thread`, `unknown`). |
| process / `exit` | `sched:sched_process_exit` | Nullable exit code/signal only if decoded from verified kernel semantics; `exit_scope` (`thread`, `process`, `unknown`). Observation of thread exit is not necessarily container/process-group termination. |
| file / `open`, `write`, `unlink`, `rename` | Selected syscall tracepoint entry/exit pairs, or separately assessed file hooks | Nullable requested path(s), `path_origin` (`user_argument`, `resolved_kernel`, `unknown`), nullable dirfd/flags; bounded write byte count only, no contents. Relative paths, mounts, symlinks, truncation and TOCTOU prevent claiming a canonical file identity. An open-for-write flag is not an observed write. |
| network / `connect` | `syscalls:sys_enter_connect` / `sys_exit_connect` if available | Nullable address family, transport, source/destination address/port and socket identity. Entry observes attempt; return zero observes syscall success, not an established application session. `EINPROGRESS` is `unknown` for connection completion with its actual errno retained. No packet/payload claim. |
| privilege / `credential_change` | Selected setuid/setgid/capability syscall entry/exit, separately assessed | Operation name, nullable requested and observed before/after UID/GID or capability-mask strings. Successful syscall alone does not prove every requested credential became effective; actual observed state and missing state remain distinct. |
| privilege / `namespace_change` | Selected setns/unshare entry/exit | Requested namespace flags/identifier and nullable observed before/after namespace IDs; attempt is not successful escape. |

Details validators are per action: prohibit fields that imply data not supplied by the selected hook. Unknown values are null, address strings are validated IPv4/IPv6 (never reverse-DNS resolved during ingestion), and all uint64 flags/masks follow §2. Additional file/network/privilege actions require an explicit schema and coverage update. LSM decisions, if assessed later, are passive observations of a decision point, not proof of a later operation; this system does not block operations. An executable named `sh` or an administration command is not inherently malicious.

Outcome constraints: entry has `attempted`, null return/errno; a verified return failure has `failed` and errno when Linux syscall semantics apply. Success is allowed only for the action actually established by that hook. Lifecycle observations can use `succeeded` for the observed transition, but no fabricated syscall return. `unknown` and a reason are preferable to inference. Entry/exit correlation must include boot, task lifetime and bounded pending state; missing entry/exit or reused PID cannot be paired speculatively.

## 7. Alert contract v1.0

Alerts are immutable agent detection results, not backend-edited incident objects.

| Field | Meaning |
|---|---|
| `schema_version`, `alert_id` | `1.0`, UUID assigned once before enqueue |
| `host_id`, `boot_id`, `agent_epoch_id` | Producer scope; all evidence must be from this host and boot; historical evidence from another registered epoch on that boot is permitted |
| `rule_id`, `rule_revision`, `rule_schema_version` | UUID, uint64 revision string, initial rule language `1.0` |
| `ruleset_revision` | uint64 string of the validated manifest actually active at detection |
| `severity` | `info`, `low`, `medium`, `high`, `critical` |
| `detected_at` | Agent UTC timestamp, not backend ingest or evidence occurrence |
| `window` | `{start_monotonic_ns, end_monotonic_ns}` uint64 strings on this boot, inclusive start/exclusive end; start less than end |
| `subject` | Identity snapshot using §4, no stronger identity assertion than the evidence supports |
| `evidence` | 1–32 unique `{host_id, event_id}` references; host must match this alert |
| `explanation` | Bounded factual explanation of matched conditions, relevant allowlist/context decisions and uncertainty; do not include secrets |
| `dedup` | `{version:"1", key:<lowercase SHA-256 hex>, scope:<bounded canonical subject key>}` |

Rule content remains available by immutable revision even after retirement. Evidence must already exist or appear in the same atomic batch. An alert may refer to unresolved events but cannot label them as resolved/host without supporting identity evidence. Correlation windows and allowlists are bounded and versioned with policy.

Dedup v1 is lowercase SHA-256 over the RFC 8785 canonical JSON array `["1", host_id, boot_id, rule_id, rule_revision, scope, start_monotonic_ns, end_monotonic_ns, sorted_event_ids]`; UUIDs sort lexically, all uint64 elements remain decimal strings, and the last element is the sorted unique evidence-ID array. This encodes the canonical tuple of host, boot, rule ID/revision, subject scope, fixed monotonic window and evidence without ambiguous string concatenation. Subject scope is a lifecycle key when resolved, a verified host task-lifetime key when available, otherwise `unresolved` plus cgroup/task facts (never a guessed container). The policy defines the fixed window width/alignment and maximum suppression interval. Equal tuple means the same immutable detection; keep the original alert ID and content. Additional evidence after a window closes produces a new detection/window, not mutation. Per-condition suppression of overlapping windows is separate bounded agent state; count suppressed matches in health. No mutable occurrence counter is hidden inside immutable alert retries.

The backend verifies digest derivation and enforces uniqueness of `(host_id, dedup.version, dedup.key)`. A different ID claiming an existing dedup key is a conflict, not a second alert or a silent ID rewrite. Detection must persist dedup decisions together with its queue records to survive retries/restarts. Alert read envelopes add immutable backend `ingested_at` just as events do.

## 8. Conceptual kernel wire header (not a C ABI declaration)

The intended common header carries:

| Conceptual member | Purpose |
|---|---|
| magic / record kind | Distinguish this record family and typed payload |
| ABI version | Select an explicitly supported decoder |
| total length / header length | Bound parsing and payload extent |
| flags | Declared truncation/availability flags; reserved bits must be zero |
| monotonic nanoseconds | Kernel occurrence time with documented helper/domain |
| cgroup ID | Raw kernel identifier, not container identity |
| host-namespace PID / TID, UID / GID | Observation subject |
| CPU ID and optional per-CPU sequence | Diagnostic ordering/loss clues, not a global sequence |
| bounded command name / declared length | Optional bounded display source |

The payload is record-kind-specific. UUIDs, Docker metadata, wall conversion, rules and backend timestamps belong in userspace, not this header. **No byte offsets, total sizes, packing directives, host-endian assumptions or struct casts are specified here.** Phase 2/3 must choose fixed-width C types, explicit byte order, padding/alignment and reserved bytes; pin generator/compiler dependencies, inspect generated layouts, and assert sizes/offsets in C and Go for supported architectures. CO-RE relocations do not by themselves stabilize a userspace record ABI.

The reader first checks the minimum bytes needed to discover version/length according to an approved bootstrap layout, then version, exact allowed header/record lengths, type, flags, bounded string lengths and payload. It never reads beyond the received buffer, uses unsafe unchecked casts, or treats a missing NUL as a valid unbounded string. Unknown versions/types, short/oversized records and inconsistent lengths are rejected and counted separately. A future extension must explicitly define skip/compatibility behavior; initial decoding fails closed on unsupported records.

## 9. Reliability and planned validation

Health separates ring-buffer reservation failures, transport/decode rejection, userspace queue drops, resolver failures, detection suppression, spool overflow and backend validation/retry failures. Counters carry epoch and counter-generation IDs; gaps in event sequence are not substituted for measured kernel loss. Unsupported probes are coverage gaps, not zero-event evidence of safety.

Planned tests (not executed): golden binary fixtures per approved ABI/architecture; malformed length/version/flag and decoder fuzz tests; maximum uint64 JSON/SQLite round trips; schema version and unknown-field rejection; restart/cgroup/PID reuse identity tests; suspend/wall-clock-step calibration tests; benign and suspicious rule fixtures; immutable retry and dedup-conflict tests; retention/evidence referential tests. Synthetic fixtures must be explicitly labelled and never displayed as collected telemetry.
