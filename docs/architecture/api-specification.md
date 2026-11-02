# REST API specification

**Status:** Phase 1 design revision 1. These endpoints and validation rules are proposed contracts, not implemented or security-tested services. The canonical scope is [PROJECT_SPECIFICATION.md](../PROJECT_SPECIFICATION.md); immutable payloads and storage semantics are defined in [event-schema.md](event-schema.md) and [database-design.md](database-design.md).

## 1. Transport, principals and scope

Every route is under `/api/v1`. JSON uses `application/json`; responses also set `Cache-Control: no-store` except that authenticated immutable rule documents may use private revalidation with ETags. No route exposes the Docker socket, arbitrary filesystem paths, SQL, eBPF control or workload enforcement.

Initial single-host deployment binds the API to loopback only. Authentication and authorization remain mandatory on loopback. Plain HTTP is permitted only on that local loopback connection with locally provisioned credentials and no forwarding to an untrusted network. Non-loopback/remote agent or browser traffic requires HTTPS with certificate/hostname verification (TLS 1.2 minimum, TLS 1.3 preferred). A TLS-terminating reverse proxy must authenticate its protected upstream boundary, reject direct remote access to any plaintext listener and not trust arbitrary forwarded headers. Never use `InsecureSkipVerify`, URL query tokens or a public unauthenticated development bind as a deployment shortcut.

Provision high-entropy revocable bearer credentials outside source control; store backend token digests, restrict local credential file permissions, rotate credentials, and redact Authorization, leases, secret configuration and sensitive body fields from logs. Browser authentication uses short-lived management tokens held in memory, not long-lived agent/admin credentials in bundled code or persistent browser storage. Initial bootstrap and credential issuance/revocation are local operator procedures, not public API endpoints. Restrict CORS to configured dashboard origins; do not use wildcard credentialed CORS. If cookie-based sessions are introduced, add Secure/HttpOnly/SameSite and CSRF protection before enabling them.

| Principal | Permission |
|---|---|
| `agent` | Submit events/alerts and heartbeat, read assigned rules; credential restricted to exactly one enrolled `host_id`. Cannot read management telemetry or edit rules. |
| `viewer` | Read health, hosts, containers, events, alerts, statistics and published rule history within its explicit host allowlist. Cannot ingest or edit. |
| `operator` | Viewer permissions plus rule updates; initial global ruleset updates require all-host policy-edit authority. A host-limited operator cannot change global policy. |

There is no implicit shared all-powerful token. Read handlers apply authorization scope before pagination/counting; explicit out-of-scope host requests return 403. Object details outside a principal's scope return 404 to avoid ID enumeration. Agent body `host_id` must equal credential scope; never trust a body field to select a principal. Host/epoch IDs are not credentials. Enrolled agent claims are trusted producer statements, not cryptographic proof of kernel observation or remote boot attestation.

## 2. Common representation, bounds and errors

All uint64 values are canonical **decimal strings**, all UUIDs are lowercase canonical UUIDs, wall timestamps are UTC RFC 3339 `Z`, and unknowns are null as specified in the event contract. No float encoding of kernel IDs or revisions. Reject duplicate JSON keys, unsupported schema versions, unknown fields/query parameters and invalid enum values. Uncompressed UTF-8 request bytes are counted before decoding; compressed request bodies are not supported initially. Multipart bodies/chunked inventory application protocols are not supported; HTTP transfer chunking does not bypass the body limit.

The following are **provisional design limits, not tested capacities**. Configuration may lower service bounds; increasing them requires contract/security review and compatibility documentation, not an unrecorded runtime change.

| Item | Initial bound |
|---|---|
| Event batch | 4 MiB body, at most 500 events and 100 alerts; at least one object total |
| Single event / alert | 16 KiB / 32 KiB canonical JSON, plus action-specific string/evidence limits |
| Heartbeat including inventory | 2 MiB body; 1000 containers maximum in one complete snapshot |
| Rule update / exact global ruleset | 64 KiB / 2 MiB uncompressed JSON; at most 256 rules in ruleset |
| List page | Default 50, maximum 200 items; bounded opaque cursor at most 2048 bytes; maximum encoded management JSON response 8 MiB |
| Summaries / published provenance | Host/container/rule list summary at most 8 KiB/item (no embedded full manifests); per heartbeat at most 16 capability manifests of 64 KiB each and 64 clock calibrations of 2 KiB each, within the total body bound |
| Time query | Default latest 24 hours, maximum 31 days per request |
| Statistics | At most 744 buckets; each at most 31 days total range; group cardinality at most 100 plus explicit `other` |
| Filter lists / JSON depth | At most 20 values per list filter; depth 12 |
| HTTP admission | 10 s body-read deadline, 15 s total ordinary handler deadline; 4 concurrent requests per agent principal, 8 per management principal |
| Rate admission | Per principal: agents 10 requests/s burst 20; management 20 requests/s burst 40; lower backend-wide bounded admission may also apply |
| Lease / heartbeat | 90 s lease, intended heartbeat interval 15 s; stale display after 45 s without backend contact |

Capacity exhaustion is visible as 429/503, not an invitation to unbounded queues. Clients honor `Retry-After` and use bounded exponential backoff with jitter. These limits must be evaluated on Linux before performance claims.

### Responses and errors

Success objects use `data`; paginated lists use `{data:[...], page:{limit,next_cursor,as_of}}`. `next_cursor=null` ends the traversal. Event and alert items are read envelopes with the immutable submitted document and backend `ingested_at`. Details may add separate linked context without rewriting that document.

Errors use this shape (illustrative **synthetic protocol example**, not collected telemetry):

```json
{
  "error": {
    "code": "id_conflict",
    "message": "An existing ID has different immutable content",
    "request_id": "bc46654c-1efb-42bd-ae19-88d3b1ee17c2",
    "details": [{"path": "events[0].event_id", "code": "content_mismatch"}]
  }
}
```

Details contain bounded paths/codes, not credential values, full rejected payloads or other hosts' metadata. Include `X-Request-ID` on every response (server generates/validates it) and UTC server time in `Date`.

| Status | Meaning |
|---|---|
| 200 | Successful read/update or committed ingestion/heartbeat, including exact replay |
| 304 | Conditional rule/manifest GET unchanged; no body |
| 400 | Malformed JSON/query, duplicate keys, invalid cursor or unknown field |
| 401 / 403 | Missing/invalid authentication / valid principal lacks operation or explicit host scope |
| 404 | Unknown or unauthorized detail object; list with no matches returns 200 and empty data |
| 409 | Immutable ID/sequence/dedup conflict, stale heartbeat/inventory revision, lost lease, or epoch not registered |
| 410 | Authorized ID known expired, closed ingestion epoch, or expired traversal cursor |
| 412 / 428 | Failed `If-Match` / required `If-Match` absent on rule update |
| 413 / 415 | Body/item/count limit exceeded / unsupported content type or content encoding |
| 422 | Well-formed but invalid schema/enum/range, evidence or rule semantics; unsupported schema version |
| 429 | Admission/rate limit, with `Retry-After` seconds |
| 500 / 503 | Unexpected internal failure / temporary writer, database or storage unavailability; no success acknowledgement, retryable where appropriate |

A 409 identifies the specific scoped conflict and **does not** authorize changing an existing object. 422 may contain up to 20 validation details. Internal errors do not disclose SQL, stack traces, filesystem layout or secrets. Do not use HTTP 207 or return partial batch success.

## 3. Management endpoints

All require viewer or operator and apply host scope. Rules are currently global, readable by those principals but editable only with global policy authority.

| Method and route | Contract |
|---|---|
| `GET /api/v1/health` | Authenticated backend readiness/status, database/writer availability, backend UTC time and API version. No invented agent metrics; detailed per-host health comes from hosts. Returns 503 if not ready. Minimal liveness outside this API, if later needed, requires separate deployment review. |
| `GET /api/v1/hosts` | Paginated enrolled hosts, current boot/epoch, last heartbeat receipt, desired/applied ruleset revisions and freshness/degradation. Filters `host_id`, `health` (`healthy`, `degraded`, `stale`, `unknown`). |
| `GET /api/v1/hosts/{host_id}` | Host metadata, latest agent health/capability manifest, clock quality, inventory revision and capture/receipt times. Missing heartbeat data is unknown, not zero. No credentials. |
| `GET /api/v1/containers` | Paginated host-scoped container inventory projections. Filters `host_id`, `runtime`, full `container_id`, `presence` (`present`, `absent`, `unknown`), `runtime_state`; no substring-ID identity resolution. Include latest inventory freshness and metadata provenance. |
| `GET /api/v1/containers/{container_id}` | Requires `host_id` and `runtime=docker` (runtime defaults to docker); container key is not global. Metadata, current inventory presence, latest lifecycle and stale state; old lifecycles are separately paginated below. |
| `GET /api/v1/containers/{container_id}/lifecycles` | Requires host/runtime scope as above; paginated restart history with boot, lifecycle UUID, cgroup ID, observed Docker start/finish, uncertainty and source. No guessed finish timestamps. |
| `GET /api/v1/events` | Paginated immutable events; filters in §4. Default ingest-time ordering. |
| `GET /api/v1/events/{event_id}` | Requires `host_id`; immutable document and original ingest timestamp. Retained receipt without payload returns 410 `event_expired`. |
| `GET /api/v1/alerts` | Paginated immutable alerts; filters in §4. Default backend-ingest ordering. |
| `GET /api/v1/alerts/{alert_id}` | Requires `host_id`; alert plus ordered evidence references and immutable rule revision link. Evidence is retained while alert lives. A broken evidence invariant is an internal error, not a silently empty evidence array. Expired retained receipt returns 410 `alert_expired`. |
| `GET /api/v1/statistics` | Bounded aggregates: `metric=events` or `alerts`, `from`, `to`, `time_basis`, `bucket` (`1h`, `6h`, `1d`), optional `group_by` (`event_type`, `severity`, `identity_status`, `host_id`, `container`) appropriate to metric and §4 host/subject filters. Counts of actually retained data with queried interval, coverage/freshness, retention clipping and unknown-time exclusion count. Not lifetime totals or an unsupported loss/detection-rate estimate. |
| `GET /api/v1/rules` | Paginated current rule summaries including ID, current revision, enabled state, severity, schema version and global ruleset revision. Filters `enabled`, `severity`; initial rules provisioned as reviewed baseline policy, not browser-generated fixtures. |
| `GET /api/v1/rules/{id}` | Exact immutable rule revision if `revision` is supplied; otherwise current rule plus strong ETag and current global ruleset revision. |
| `PUT /api/v1/rules/{id}` | Operator/global policy authority only; full rule replacement with optimistic concurrency, §7. |

There are no rule create/delete, event mutation, alert evidence mutation, container-control or credential-issuance endpoints in initial v1. New rule IDs are introduced through reviewed baseline/provisioning changes; disabling a rule is a new revision, not deletion of history. A future creation API needs a separate contract.

**FR-17 report export is Medium priority and deferred.** No export endpoint, asynchronous export job or unlimited download is promised in v1. Bounded JSON list APIs are not a completed export feature. Later design must define formats, authorization, range/size limits, audit, sensitive-data handling and consistency.

## 4. Filters, time and pagination

Common list inputs are `limit` and `cursor`. Cursor and all filters must match the original traversal. Event/alert/statistics ranges are `[from,to)` in UTC, `from < to`; if omitted, use backend now minus 24 hours to now. If only one bound is supplied, reject 400 rather than infer an unexpectedly large range. Explicit ranges above 31 days return 422.

Event filters: `host_id`, `boot_id`, `agent_epoch_id`, full `container_id`, `runtime`, `lifecycle_id`, `identity_status`, `event_type`, `action`, `pid`, `cgroup_id`, `outcome`, `from`, `to`, `time_basis=ingest|event` (default ingest). Container/lifecycle filters require a host; lifecycle also requires container/runtime. `identity_status=host|unresolved` with a container filter is contradictory and rejected. PID filtering finds observations, not a unique process. Cgroup equality is not an attribution claim. Repeated supported enum filters are OR within a field and AND across fields, at most 20 values; duplicate scalar parameters fail.

Alert filters: host/boot/producer epoch, full container/runtime/lifecycle, identity status, `severity`, `rule_id`, `rule_revision`, `from`, `to`, `time_basis=ingest|detected` (default ingest). Rule revision requires rule ID. No raw SQL/regex/full-text payload query is exposed. Invalid action/type and metric/filter combinations return 422, not silently empty results.

Occurrence-time queries exclude events with `observed_at=null`, expose `excluded_unknown_time_count` for the otherwise matching scope, and include events flagged uncertain without pretending their estimated time is exact. Statistics mark that distinction. Ingest-time queries include such events. Returned metadata describes the selected time basis and excludes no data merely because it is unresolved. A bucket zero means no retained matching records, not proof that no activity occurred.

Use keyset pagination: events/alerts descend by selected normalized wall time then UUID; inventories ascend by host/runtime/container key; hosts by host UUID; lifecycles by lifecycle UUID (not a claim of chronological order); rules by rule UUID. Cursors are authenticated opaque tokens bound to principal scope, filters, ordering, `as_of` backend receipt watermark and last key, with a **provisional 15-minute lifetime**. Expired tokens return 410 `cursor_expired`; restart the query. No offset pagination or exposed SQL cursor.

For immutable telemetry include only rows ingested by the first-page `as_of`; late arrival is visible on a new traversal even if event time is old. For mutable host/container/rule lists bind the cursor to the selected inventory/projection/rule generation; if it changed, return 409 `list_changed` and restart, never mix revisions silently. Inventory can be stale even if a page is consistent. Retention may remove rows mid-traversal; `as_of` is not an indefinite MVCC snapshot. Metadata states `consistency=bounded_live`, and available-data counts/retention warnings must not imply a frozen export. Total counts are omitted by default to avoid unbounded queries.

## 5. Agent event and alert ingestion

### `POST /api/v1/agent/events`

Requires host-scoped agent credential. The body is **exactly** `{ "events": [...], "alerts": [...] }`; both arrays required. Every object carries schema, host, boot and producer epoch. All objects in a request must share host, boot and agent epoch; evidence may reference retained events from another registered epoch on the same boot. No top-level host override or backend ingest timestamp is accepted. Agent epoch must already be registered and open. Active lease is not required for historical spool delivery; a historical epoch cannot thereby overwrite inventory/health. A credential's current host scope remains mandatory for old delivery.

Validate immutable event/alert schemas, known same-host/boot source capability and clock records, and known rule/ruleset revisions. Missing provenance returns 422 `unknown_manifest` or `unknown_calibration`; publish frozen records in heartbeat first rather than fabricating source/clock facts. An alert rule must be a member of the declared ruleset at exactly its revision, enabled in that manifest, and have compatible event/action semantics. Backend accepts a valid known historical ruleset from an offline agent; it does not rewrite to latest. Reject unknown rule content/revision with 422 `unknown_rule_revision` and require synchronization, not synthetic policy creation.

**Atomic behavior:** validate all objects, check scoped IDs/sequence/dedup uniqueness, insert new events, then alerts/evidence, and commit one transaction. A single invalid object, missing evidence or conflict means no new objects are committed. There is no per-item rejection/partial acknowledgement. Evidence may refer to a retained event or an event in this same batch; a payload-expired receipt is insufficient. Duplicate IDs inside an array fail 422 `duplicate_id_in_batch`.

Idempotency is by `(host_id, event_id)` or `(host_id, alert_id)`, not by an ephemeral HTTP request ID. Freeze each document before initial delivery, use canonical JSON/digest rules from the database design, and retry it unchanged. Whitespace/object-key order variations canonicalize to the same content; changing identity, time, explanation or evidence does not. Array order is significant inside a document, so preserve evidence order on retries. Batch grouping/array order may change when splitting retries, but no individual object may change.

- Same ID/content: replay, original ingest timestamp retained, no second row.
- Same ID/different content: 409 `id_conflict` for whole batch.
- Different event ID using existing epoch sequence: 409 `sequence_conflict`.
- Different alert ID using an existing verified dedup tuple/key: 409 `dedup_conflict`. Agent recovers its durable original ID, not a newly generated replacement.
- Expired payload with retained receipt and same content: replay with `payload_state=expired`, no resurrection. New alert evidence requiring that expired event fails 422 `evidence_expired`.
- Closed epoch: 410 `epoch_closed`; unknown boot/epoch linkage: 409 `epoch_unregistered`. Never rewrite old records into the current epoch.

A committed request always returns 200, including all-new, mixed new/replayed and all-replayed batches. The following is an illustrative **synthetic acknowledgement only**:

```json
{
  "data": {
    "committed": true,
    "events": [
      {
        "event_id": "35393d10-7be6-423c-94bc-dc30221c3e58",
        "result": "created",
        "payload_state": "retained",
        "ingested_at": "2026-01-01T00:00:00.000000000Z"
      }
    ],
    "alerts": [],
    "counts": {"created": 1, "replayed": 0}
  }
}
```

Each result is `created` or `replayed` and includes its original backend timestamp; array order follows the submitted arrays. `counts` uses safe bounded integers. If a response is lost after commit, the agent resends immutable objects and obtains replays. 4xx validation/conflict requires operator-visible dead-letter handling; never retry unchanged invalid content forever. 413 may be handled by splitting arrays while retaining required same-request evidence dependencies and immutable IDs, or referencing already committed retained events. A single oversize object cannot be split; surface failure. 429/503 and ambiguous connection/timeouts retry with backoff. Never acknowledge or dequeue before a committed response or confirmed same-content replay.

## 6. Heartbeat, epochs and authoritative inventory

### `POST /api/v1/agent/heartbeat`

Requires the scoped agent credential. Required fields:

| Field | Contract |
|---|---|
| `schema_version` | `1.0` |
| `host_id`, `boot_id`, `agent_epoch_id` | Producer identity |
| `heartbeat_id`, `sequence` | UUID and uint64 string, increasing per epoch |
| `register_epoch` | Boolean; true only for first registration/retry of this epoch |
| `reported_at` | Agent UTC time, not freshness authority |
| `health` | Bounded object with state (`healthy`, `degraded`), explicit reasons, applied ruleset revision, actual capabilities/clock manifest and inventory collection state |
| `counters` | Named uint64-string stage counters, generation ID, observation interval; absent unsupported metrics are null/unavailable, not invented zeros |
| `queues` | Bounded-depth counts and capacities for userspace/durable queues; omitted unsupported subsystems explicitly marked unavailable |
| `inventory` | Null, or one full object described below |

Registration is a protocol operation within heartbeat, not an unprotected enrollment endpoint: host and agent credential must already be provisioned. The first heartbeat has `register_epoch=true`, sequence `"1"`, a fresh epoch UUID and `inventory=null`. It atomically registers the immutable boot linkage and obtains an active lease only if no unexpired host lease exists or the operator has explicitly revoked it. A previously historical/closed epoch cannot register again. A prior active epoch becomes historical on a successful takeover. At most one lease is active per host. No PID/hostname heuristic grants takeover; the credential is the authority for that enrolled host.

Subsequent requests set `register_epoch=false`, carry `X-Agent-Lease` with the returned lease generation UUID and renew the 90-second lease on accepted heartbeat. A different agent restart may have to wait for lease expiry or local operator revocation. Historical epochs retain event-delivery permission while open, not heartbeat authority. Check lease and epoch before replay handling; even an old exact heartbeat cannot renew a replaced lease. An exact retry of the first registration is allowed only while that same epoch remains active and returns the original lease generation; replay does not extend the lease. If it has expired, renewal requires a new higher-sequence heartbeat with that lease and no intervening takeover; backend checks atomically. After any newer heartbeat is accepted, an older registration replay is stale and cannot recover an obsolete acknowledgement.

An exact retry of the latest accepted heartbeat ID/sequence/content returns its original acknowledgement and does not refresh backend freshness; older heartbeats are stale even if formerly accepted. Lower sequence, reused sequence/ID with different content, or lost lease returns 409 (`stale_heartbeat`, `heartbeat_conflict`, `lease_lost`). Higher sequence may skip numbers. Latest `agent_status` and optional inventory commit atomically. Gaps are telemetry, not measured ring-buffer loss.

Counter names separate ring reservation failures, ABI decode rejection, userspace queue loss, resolver uncertainty, rule suppression, durable spool loss and backend retry/rejection. A counter reset requires a new explicit counter-generation UUID; report deltas only for the same generation/domain. Include actual kernel/architecture/BTF/cgroup/runtime and hook probe results in a bounded capability manifest referenced by event `coverage_id`. Advertise missing coverage as degraded/unsupported, not a successful test. `health.capability_manifests` and `health.clock_calibrations` are bounded arrays of immutable records, each with its own host/boot/registered producer epoch, UUID and canonical content digest; they may be empty when already published. Records are persisted independently of short-lived heartbeat history and cannot change under an existing ID. An authorized current agent may republish durable historical records for its host before draining that registered epoch's spool, without changing historical lease/status or current inventory. Missing unregistered historical epochs require explicit local operator reconciliation, not relabeling old events. Reject health manifests that exceed the heartbeat limit; large debug dumps are not API data.

### Complete snapshot, never chunks

`inventory` is `{snapshot_id, revision, complete:true, capture_started_at, captured_at, containers:[...]}`; the two UTC times bound collection and `captured_at` is its end. Completeness means full successful reconciled Docker enumeration, not a transactional Docker/kernel snapshot or proof of seeing every transient container. `revision` is a host-global uint64 string; `snapshot_id` UUID. Every item contains full Docker ID, runtime, nullable current lifecycle UUID, boot ID, observed runtime state (`running`, `paused`, `restarting`, `exited`, `created`, `dead`, `unknown`), bounded nullable name/image ID/reference, inspected started/finished UTC timestamps, nullable cgroup ID, nullable mapping revision and `mapping_status=resolved|unresolved`. Lifecycle-bound fields must agree with that boot/lifecycle. An empty array is authoritative only after a successful full Docker enumeration/inspection confirming no containers.

The agent must successfully list **all** Docker containers, not only running ones, reconcile lifecycle events/inspection around capture, and establish a complete bounded snapshot. Concurrent changes that prevent a coherent reconciliation, daemon errors, any missing inspection or oversize make collection fail; send `inventory=null` and an explicit degraded/stale reason. Do not send the first 1000 containers, multiple chunks or a partial list with `complete=true`. `complete=false`, chunk IDs, continuation tokens and incremental deletion lists are rejected. More than 1000 items or a body over 2 MiB returns 413 **without applying any inventory or heartbeat changes**. The agent then sends a new health-only heartbeat with a new ID/sequence and inventory-too-large reason. Supporting larger hosts requires a reviewed contract change, not automatic partial deletion.

On a new snapshot, `revision` must equal backend current revision + 1 (new host starts at `"1"`), even across boots/agent epochs. The backend returns current revision at registration so the agent can construct its next snapshot. Same current revision/snapshot ID/content can be resent without replacing data. Stale, skipped or same-revision/different-content snapshots return 409 `inventory_revision_conflict` with the current revision; the entire heartbeat rolls back. A new epoch may not replay a snapshot document owned by a different epoch; collect a new one at the next revision.

In one transaction insert the snapshot, update present container projections, retain/create lifecycle history, and mark previously present containers omitted from this snapshot as **inventory absent**. Never hard-delete historical containers, close old lifecycle intervals with invented times or turn absence into a process-exit claim. Names/images change projections only; event snapshots remain immutable. Running/restarting items whose lifecycle cannot be established carry null lifecycle/cgroup and an explicit unresolved mapping state rather than an invented UUID. Lifecycle history observed independently by Docker events is kept even if absent from this latest snapshot.

200 acknowledgement contains `heartbeat_id`, `sequence`, `result=accepted|replayed`, original `received_at`, `lease_generation`, `lease_expires_at`, `current_inventory_revision` (initial `"0"`), `inventory_result=unchanged|applied|replayed`, and `desired_ruleset_revision`. These values acknowledge the accepted heartbeat; a replay returns that original acknowledgement, not necessarily the latest desired state. Fetch fresh rules separately. Lease generation is protocol state, not a substitute for the agent credential. Health freshness is based on backend receive time, inventory freshness on capture **and** receipt with clock uncertainty disclosed; a fresh health heartbeat does not make an old inventory fresh.

## 7. Immutable rule distribution and optimistic updates

### `GET /api/v1/agent/rules`

This route is required: a heartbeat revision number is not rule distribution. Requires agent credential, explicit `host_id`, `boot_id`, `agent_epoch_id` query and current `X-Agent-Lease`; validate against active enrollment/lease. It cannot expose other hosts' policy or accept a body-supplied principal. Initial policy is global but distribution remains host-authorized for future scope changes. Provisioning publishes validated ruleset revision `"1"` (an explicitly empty set is permitted); desired revision is never an invented default. Before any set is applied, heartbeat `applied_ruleset_revision=null` and detection state is unavailable, not revision zero.

Without `revision`, return the latest published assigned global ruleset. With a uint64-string `revision`, return exactly that retained immutable manifest, not a silently substituted latest version; unknown is 404. Response `data` includes `schema_version:"1.0"`, `ruleset_revision`, `published_at`, `manifest_digest`, the sorted complete list of `{rule_id, revision, content_digest, rule}` and supported event/rule schema versions. Manifest digest is SHA-256 of RFC 8785 canonical `{schema_version,ruleset_revision,members}`; members are sorted by rule UUID and contain the full rule plus its digest. Publication time and digest field itself are excluded. Each rule digest covers its entire immutable rule document, including its rule ID/revision/schema version. Backend and agent verify all bounds and digest inputs identically; hashes protect consistency, HTTPS/authentication provide origin trust.

Use strong ETag `"ruleset-<revision>-<digest>"`. `If-None-Match` for the selected revision returns 304 when unchanged. This is one bounded complete response, **not a paginated policy delta**. Over 256 rules or 2 MiB prevents publication/distribution with an explicit error; never activate a prefix of the set.

The agent downloads into staging; checks schema compatibility, unique rule IDs, exact revision fields/digests, known typed fields/operators, bounded conditions/windows/allowlists, and all required capabilities. Compile/validate the **whole** set before atomic replacement. Any invalid/unsupported member prevents activation, retains last-known-good set, and reports desired/applied mismatch and reason in heartbeat. A freshly enrolled agent without a usable baseline reports detection unavailable; collection may continue, not falsely claim detection coverage. Persist applied manifest and revision for restart. Offline operation uses a validated last-known-good revision and reports staleness; no automatic invented defaults. Rollback is publication of a new ruleset revision referencing reviewed policy revisions, not mutation/reuse of an old manifest number.

### Rule schema v1.0 and `PUT /api/v1/rules/{id}`

A full rule document contains `rule_schema_version:"1.0"`, UUID `rule_id`, immutable uint64-string `revision`, `name` (1–128 UTF-8 bytes), `description` (at most 2048), boolean `enabled`, severity enum, supported `event_types`/`actions`, `match`, `allowlist`, `correlation`, `dedup` and `explanation_template`. Shell/tool names alone do not constitute a maliciousness claim. Conditions must include contextual evidence and explicit allowlist behavior.

The initial language is a bounded typed condition tree: `all`, `any`, `not` and leaf `{field,op,value}`; max depth 8, 64 leaves, max 100 total allowlist entries. Fields come only from the documented normalized event/action schema; operators are typed `eq`, `in`, `prefix`, `cidr_contains`, `exists`. No arbitrary scripts, SQL, network lookups, regex engine or templated execution. `in` is bounded to 20 values; `prefix` applies to bounded strings (not a canonical-path guarantee); CIDR operations require parsed IP data. An absent/null field does not match an ordinary comparison, including under `not`: use three-valued evaluation (true/false/unknown), `not unknown=unknown`, with only final true matching; `exists` explicitly tests availability. Allowlist true suppresses a match; unknown is not a proven exemption and is disclosed in explanation.

`correlation` is null for single-event rules or declares a bounded same-host/boot subject key, required action set, minimum/maximum event count (2–32), window width (1–300 seconds) and maximum pending subjects (1–10000). Resolved-container correlation uses lifecycle keys; task correlation requires task lifetime identity. Null/ambiguous keys cannot merge unrelated unresolved tasks. Exhaustion is counted/degraded, never unbounded state. `dedup` declares version `"1"`, boot-monotonic fixed-window width/alignment (1–300 seconds) and bounded suppression interval (0–300 seconds); derive the alert tuple exactly as the event contract, canonical JSON array before SHA-256 hashing. Templates interpolate only approved bounded evidence fields, with no executable expressions. This is a design language to validate/test later, not a claim that a detector exists.

Management updates require `Content-Type: application/json` and `If-Match` from the current `GET /rules/{id}` response: strong ETag `"rule-<id>-<revision>-<digest>"`. The PUT body is the full editable rule content above **without `revision`**; the backend assigns current+1. Body/path rule IDs must match. Missing precondition returns 428; stale ETag 412 with no mutation. Validate all policy/capability-independent semantics before publication, reject unknown schemas/fields/invalid combinations as 422. Host capability-dependent activation remains agent-side and visible in health.

A successful transaction appends the rule revision and a new exact global ruleset manifest, updates current pointers, and returns 200 with the new immutable rule, ETag, ruleset revision and publication timestamp. Disabled rules remain in manifests with `enabled=false` and generate no new alerts, while old revisions remain explainable. Every successful replacement, even identical editable content, advances revisions. PUT retries use the original `If-Match`: after a lost response a repeat gets 412; fetch the current rule and audit revision to determine whether the intended edit committed, rather than blindly replacing a concurrent change. No undocumented last-write-wins behavior.

## 8. Planned contract validation and limitations

Planned only: principal/role/host isolation; loopback/remote TLS negative checks; malformed/oversized JSON and slow body deadlines; decimal uint64 extremes in a TypeScript client; cursor/filter/time-bound tests; atomic mixed batch rollback and dropped-response replay; payload expiry/closed epoch behavior; rule evidence and dedup conflict tests; lease competition and stale heartbeat; complete/empty/failed/oversized inventories with restart histories; concurrent `If-Match` rule edits; immutable manifest digest/304 handling; unsupported rule activation preserving last-known-good; bounded backpressure/disk-full and truthful freshness states.

No HTTP server, OpenAPI generator, authentication library, migration, frontend code, kernel test or Phase 2 dependency is introduced here. Linux runtime compatibility and all numerical capacity assumptions require subsequent approved implementation and executed evaluation. Coverage gaps, stale inventory, unresolved identity, clock uncertainty and retained-data clipping must remain visible in API responses and the future dashboard.
