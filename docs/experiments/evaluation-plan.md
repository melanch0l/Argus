# Evaluation plan

## 1. Status, purpose and traceability

**Phase 1 protocol; no experiments have been executed and no numerical result or compatibility claim is made.** [The project specification](../PROJECT_SPECIFICATION.md) is authoritative. The [objectives](../research/objectives.md), [questions](../research/questions.md) and [scope](../research/scope.md) distinguish the execution MVP from the full process/file/network/privilege thesis. This plan authorizes no implementation or Phase 2 work.

| Study | Questions/objectives | Canonical requirements | Primary evidence |
|---|---|---|---|
| Capability and observation | RQ1; O1, O5 | FR-01–FR-07, FR-18 | Actual probes, semantic matrix, matched operations/events |
| Container attribution | RQ2; O2 | FR-08, FR-09 | Independent lifecycle/identity truth and event-level outcomes |
| Contextual detection | RQ3; O3 | FR-10, FR-11, FR-16 | Frozen labels/rules, matched alerts, benign controls |
| Management integrity | O4; supports RQ1–RQ3/RQ5 | FR-12–FR-17 | Durable ingestion/query/filter/export checks and negative API tests |
| Cost and scaling | RQ4, RQ5; O5, O6 | FR-01, FR-07, FR-10, FR-12, FR-18 | Repeated condition contrasts, stage counters, resource/time series |

NFR-01–NFR-10 additionally require architecture review, graceful shutdown, malformed-input/fuzz tests, retention and disk-pressure tests, security tests, pinned clean builds, unit/integration tests and actual compatibility records. Benchmarks alone do not establish these properties.

## 2. Pre-registration and reproducibility

Before confirmatory trials, freeze a dated, versioned protocol and retain its identifier with every run. It must contain:

- Host CPU/core count, RAM, storage/filesystem, virtualization, OS, actual kernel/build/architecture, BTF, cgroup mode, Docker version/configuration, privileges and available hook/helper/program-type probes. Record load/attach failures and reasons; kernel version alone is insufficient.
- Source/build identifiers, tool and dependency versions, event ABI/schema, enabled hooks, filters, rule versions, allowlists, deduplication window, resolver/retry policies, buffer/queue sizes, batch settings, SQLite/WAL/retention configuration and backend placement.
- Workload/image identifiers, operation manifests, seeds, concurrency, offered-load levels, trial durations, warm-up, drain limits, scenario labels, alert matching windows, resource sampling interval and clock calibration procedure.
- Numerical acceptance budgets and their rationale: per-family capture, incorrect/unresolved identity, precision/recall, false-positive units, p95/p99 latency, CPU/memory/workload overhead and stage loss. Record minimum sample sizes and uncertainty requirements. These budgets are **unresolved in Phase 1**, not assumed zero or silently chosen later.
- Fixed repetition count, randomized condition order/seed, run exclusion rules, handling of zero denominators/missing truth, confidence interval method and planned subgroup comparisons.

Pilots may establish safe load levels, sample size and feasible budgets. Label and retain pilot data, then freeze the confirmatory protocol before new trials; pilot data must not be reused as confirmatory evidence. The proposed starting design is five independent matched blocks per workload/load cell, each containing all three conditions. Pilot variance may justify a different count before freezing; five is not a power guarantee. No optional stopping after favorable results. Post-freeze changes require a deviation record and separately labelled exploratory results or a new protocol with fresh trials.

Retain raw independent workload logs, feature probes, counter snapshots, sampled resources, event/alert exports and analysis definitions, subject to privacy/retention constraints. No measurements or reference citations are supplied by this plan.

## 3. Ground truth and matching

### 3.1 Independent operation and identity records

Use a controlled workload driver that records run/scenario/operation identifiers, expected actions, measured attempt/completion times, return status and actual outcomes independently of the monitor. Define each expected event's cardinality and semantics before testing: one operation may produce zero, one or several expected hook observations. Do not assume one syscall equals one event.

For each container, independently retain Docker create/start/stop/destroy evidence, immutable container ID, image identity, host/boot identity, lifetime intervals and verified cgroup membership. Observe workload process membership while alive where possible; include namespace-to-host PID mapping and process start/lifetime evidence rather than PID alone. Ground-truth attribution must not simply copy the resolver under test or accept a container-ID-looking string. Short-lived cases need lifecycle capture established before workload launch. Missing independent truth is a ground-truth failure, reported explicitly rather than scored as correct or silently removed.

Match expected and observed events using the predeclared tuple of run, host/boot, verified identity/lifetime, action, safe operation-specific markers and a bounded time interval. Markers must be observable under the selected hook semantics; do not assume a scenario UUID appears in a kernel record. Prefer serialized uniquely distinguishable operations for correctness trials. For high-rate trials where unique matching is impossible, use count/counter analysis and disclose that event-level recall or attribution cannot be established there. Ambiguous matches remain ambiguous and are counted separately. One observation cannot satisfy multiple expected records; duplicates are tracked, not extra true positives.

### 3.2 Scenario matrix

| Family | Positive/coverage trials | Benign and boundary controls |
|---|---|---|
| Process | Known executions, process creation/exit, concurrent and short-lived processes | Allowed shells/admin tools, failed exec attempts, repeated executable names, host processes |
| File | Known selected reads/writes or other frozen actions on disposable test paths | Allowed maintenance; denied operations; rename, relative/long paths and namespace cases |
| Network | Controlled connections to isolated test endpoints with known outcomes | Allowed destinations, failed connections, each claimed protocol/address family; no payload dependence |
| Privilege | Safe credential/capability-related requests and verified outcomes in disposable containers | Legitimate administration, denied/no-change requests; distinguish attempted from successful change |
| Identity | Concurrent containers, restart/recreate, rapid exit, stale-cache opportunities and reuse stress | Independent host activity; unresolved lookup; Docker API interruption and recovery |

“Positive” in coverage trials means an expected observation, not maliciousness. Detection labels instead come from a frozen policy/scenario manifest: suspicious episode, benign episode, required evidence, eligible rule, expected alert and allowed matching window. Include benign lookalikes for every rule, isolated and bounded process/network sequences, allowlist cases and bursts inside/outside the deduplication window. Use a held-out confirmatory scenario/seed set after rule tuning. These are policy-labelled tests, not proof of real-attack prevalence or universal malicious intent.

For each scenario, preserve both expected intent and independently confirmed outcome. A failed privilege request must not become ground truth for successful escalation. Do not conduct destructive exploits or contact unapproved external destinations.

## 4. Coverage and loss: separate accounting

Publish a per-kernel, per-event-family matrix: selected action, hook/program type, attempt/success semantics, probe result, attachment result, permission needs, path/field limitations and test outcome. Distinguish unsupported, misconfigured/permission-denied, attach-failed, supported-but-not-tested and tested outcomes. An optional fallback is a different observation path and must be evaluated with its own semantics.

Let `E_scope` be expected records across all declared tested-scope operations, and `E_supported` the subset whose required observation path is supported and successfully attached. Let `M_supported` be uniquely matched captured records from that subset.

- Supported-scope capture recall = `M_supported / E_supported`.
- Tested-scope delivery coverage = `M_scope / E_scope`, where `M_scope` counts unique matches across all declared operations, including evaluated fallback paths. Also report unsupported expected records and each other ineligible category separately.
- Duplicate fraction = `duplicate captured records / all captured records` under the frozen matching rule.

Unsupported records do not enter a ring-buffer-loss denominator. Keeping them visible in tested-scope coverage prevents conditional success from hiding missing functionality. A supported expected record with no match is an **observation discrepancy** until evidence locates its cause; it is not automatically a buffer drop. Required but unsupported families remain a full-thesis limitation.

Instrument and report interval deltas for at least the following stages when implemented:

| Stage | Numerator / denominator for a stage-specific rate |
|---|---|
| Kernel reservation | Failed reservations / reservation attempts, where attempts = successful + failed reservations |
| Decoder | Rejected records / records read; break down invalid length, ABI and other reasons |
| Agent queue | Dropped new records / enqueue attempts; separately count eviction of already accepted records if that policy exists |
| Delivery | Unique terminally failed logical records / unique logical records submitted; retries are attempts, not new records |
| Persistence | Unique rejected logical records / unique logical records offered; separately report durable acknowledgments, duplicates and pending records |

Report policy filters and suppression separately from involuntary loss. Measure relevant pre-reservation filters where feasible; uninstrumented gaps remain unknown. Record queue high-water marks and occupancy, retry counts, accepted/pending records at boundaries and unique logical event IDs. Reconcile each boundary using `opening backlog + accepted inputs = forwarded outputs + terminal discards + closing backlog`, with retry/deduplication semantics declared. Snapshot/reset counters by boot and agent session; include late drain or bounded undelivered records. A durable acknowledgment must mean a committed transaction, not merely queue acceptance.

Do not sum stage percentages or overlapping failure counters. End-to-end missing fraction, when unique matching is valid, is `(E_supported - M_durable) / E_supported`, with `M_durable` the uniquely matched expected records durably stored. Report stage explanations and the remaining discrepancy separately; unavailable counters mean unknown loss, not zero. Missing ground truth and ambiguous matches must be disclosed alongside denominators.

## 5. Attribution and detection metrics

### 5.1 Identity

For captured records with independently known origin, classify the attribution at the detection deadline as `C` correct, `I` incorrect or `U` unresolved; `N = C + I + U`. A correctly identified host record belongs in `C` only with affirmative host ground truth. An unresolved container record is not host; a resolved wrong-container or wrong-host assignment belongs in `I`.

- All-eligible attribution accuracy = `C / N`.
- Incorrect attribution rate = `I / N`; unresolved rate = `U / N`.
- Resolution rate = `(C + I) / N`; resolved-only accuracy = `C / (C + I)`.

Report counts and rates by family, host/container, churn condition and load. Report records lacking independent truth separately. Missing collection is measured in coverage, not credited as identity success; optionally report correctly attributed durable records divided by supported expected records as a joint end-to-end measure. For eventual re-enrichment, repeat metrics at a separately frozen deadline and report resolution delay; later correction must not erase an incorrect identity used by detection.

### 5.2 Detection units and formulas

The primary classification unit is one predeclared scenario episode for one eligible rule in a known container context. Freeze episode boundaries, evidence requirements, matching windows and whether one or several alerts are expected. Match one alert to at most one labelled positive unit. A positive episode with a matching alert is `TP`; an unmatched positive is `FN`; a benign episode with at least one incorrect alert is `FP_episode`; a benign episode with none is `TN_episode`.

- Episode recall / detection rate = `TP / (TP + FN)`; these are the same measure, not two independent achievements.
- Episode precision = `TP / (TP + FP_episode)` under this collapsed one-decision-per-unit design.
- Episode false-positive rate = `FP_episode / (FP_episode + TN_episode)`; the unit is benign episodes, **not events or hours**.
- Alert-level precision = `matched true alerts / all scored emitted alerts`, using frozen one-to-one matching. Extra alerts on a true episode are duplicates/excess, not extra true positives; report their count separately from benign false alerts.
- Operational false-alert density = `false alerts on independently benign activity / monitored benign container-hours`; optionally also report per host-hour and per 1,000 benign operations, with those denominators stated. Density is not a probability or a substitute for episode FPR.
- Duplicate suppression effectiveness = `suppressed duplicate candidates / duplicate candidates eligible under the frozen suppression policy`; separately check missed distinct episodes so suppression cannot hide recall loss.

Report micro totals and per-rule/scenario results; use a predeclared macro summary only if justified. A conditional detection score uses only episodes with supported, captured, correctly attributed evidence; publish excluded counts and reasons. Also report end-to-end recall over all labelled positive episodes in declared scope, including unsupported coverage, collection failures and unresolved/wrong identity as missed alerts when no correct match exists. An alert against the wrong container does not satisfy the true-container episode. Match by rule/version, evidence, identity and time, not timestamp alone.

For alerts, check explanation completeness against required rule/version, evidence links, severity/reason, time and explicit identity status. A complete explanation does not prove correct detection. Zero-denominator metrics are `N/A` with counts, never silently 0% or 100%. Small samples and no observed false alerts do not prove a zero population false-positive rate.

## 6. Clock domains and latency

Record host/boot identity, clock source and units for every timestamp. Establish whether the selected kernel source is monotonic time excluding suspend or boot time including suspend; use the matching userspace clock. Do not directly subtract kernel monotonic time from wall-clock ingestion time. Record suspend, clock adjustments and calibration discontinuities.

Measure separate stages where instrumentation permits: kernel occurrence → agent read, read → identity/detection decision, decision → management receipt, receipt → durable commit, and commit → queried/displayed visibility. For two timestamps in the same host/boot and compatible monotonic domain, latency is `end - start`. Different clock sources require measured mappings with before/after calibration samples, interpolation assumptions and an uncertainty bound; refresh mappings during long runs. Wall-clock timestamps support human-facing time filters but do not automatically support accurate duration measurement.

For a remote backend, retain synchronization method, measured offset/drift bounds and network uncertainty. End-to-end latency is an estimate or interval incorporating those bounds, not an exact subtraction of two wall clocks. If reliable mapping is unavailable, report local-stage durations and round-trip measurements separately; mark one-way/end-to-end latency unmeasurable. Do not clamp negative values to zero or discard them without diagnosing clock/matching errors.

For valid matched records, report sample count, median, p95, p99 and maximum by stage/family/load using a frozen quantile method. Report missing and censored observations separately: unmatched or undelivered records are not zero-latency samples, and latency among survivors must accompany delivery coverage. Define detection latency from the final required evidence event to alert decision; report first-event-to-alert separately for multi-event rules. Define dashboard visibility via an actual query/render acknowledgment and include its polling interval if measured.

## 7. Repeated disabled/monitor/detection experiments

Use identical workload manifests and matched seeds within independent blocks:

| Condition | Active path |
|---|---|
| D — Disabled | No agent/eBPF attachments. Workload and independent truth/resource instrumentation remain active. Keep management idle configuration and placement consistent. |
| M — Monitoring only | Selected hooks, validation, attribution, health, transmission, durable storage and query path active; detection/rules disabled. |
| T — Monitoring with detection | Same M settings plus frozen rules and alert persistence/delivery. |

Randomize/counterbalance D/M/T order within each block and record the seed. Use the same host allocation, CPU/power policy, images, storage state policy, measurement duration and background-load controls. Warm up consistently, then measure steady state and retain startup/recovery measurements separately. Bound drain time and preserve pending work. Restart/reset the system under test between runs according to the frozen policy; avoid carrying caches, stale resolver state, retained data or rule state across conditions unintentionally. Cold versus warm cache tests, if included, are separate strata.

Use at least idle, representative mixed activity and increasing fixed offered-load levels, with concurrency/container counts recorded. Include isolated family workloads for attribution of costs and a mixed workload for interaction. Distinguish offered operations/s, expected observations/s, captured records/s and durable records/s; count multiplicity can differ. For closed-loop workloads report achieved rate and slowdown; for open-loop loads record scheduled rate, achieved rate and backlog so generator saturation is not mistaken for monitor capacity.

At each pre-registered load, repeat the full matched blocks rather than treating thousands of events in one run as independent replications. Include a bounded overload/recovery trial: increase above sustainable capacity safely, return to a lower rate, measure queue drain/recovery and account for loss. Define sustainable capacity before trials as the highest tested offered rate satisfying frozen coverage/loss/latency budgets with no continuing backlog growth over the observation window. If none qualifies, report none; do not infer untested rates.

## 8. Resource and performance formulas

For each condition X over measured wall duration `t`:

- Component CPU in core-equivalents = `delta process-group CPU seconds / t`; percent of one core = `100 × delta CPU seconds / t`. If reporting percent of total host capacity, additionally divide by logical core count and label it. Include agent/backend/frontend or load generator separately; record host system CPU and kernel/eBPF CPU where measurable rather than assuming it is in agent process CPU.
- Component memory = sampled RSS and, where available, PSS plus kernel map/ring-buffer allocations; report median/peak and avoid blindly summing shared RSS. Define the sampled process/cgroup set and interval. Separate management memory, kernel allocations and workload memory.
- Workload runtime overhead for X = `100 × (runtime_X - runtime_D) / runtime_D` for a fixed-work trial.
- Throughput degradation for X = `100 × (throughput_D - throughput_X) / throughput_D` for a fixed-duration comparable service workload. Report request latency separately where applicable.
- Absolute incremental resource cost = `resource_M - resource_D` for monitoring, and `resource_T - resource_M` for detection; also report total `resource_T - resource_D`. Near-zero baseline resource values make percentage resource overhead unstable, so publish absolute differences.

Compare matched runs, not unmatched aggregate averages. Report block-level differences, central tendency, dispersion and 95% confidence intervals using the predeclared run/block-level method (for example, paired block bootstrap when sample size permits). Show individual runs for small samples and disclose weak interval precision. Do not bootstrap individual events as independent evidence of system-level repeatability. Negative measured overhead is possible due to noise and must not be rewritten as zero. Any claim of meeting budgets must identify uncertainty and the frozen decision rule, not only a favorable mean.

## 9. Management, resilience and safety checks

In later approved implementation/evaluation phases, verify durable restart recovery, idempotent retries, duplicate event IDs, schema/length rejection, queue saturation, backend outage, Docker API outage, graceful bounded shutdown and counter reset behavior. Test retention/expiry and disk-pressure handling on disposable data; preserve the ability to explain discarded records. Do not mix intentional fault-injection outcomes into normal-load summaries without separate labels.

Verify `/api/v1` authentication/authorization for ingestion and management, bounded input sizes, validation, pagination, rule-version distribution and transport protection. Confirm no credentials in logs and no Docker socket exposure. Compare stored ground-truth-linked records with API results, container/time filters, exports and actual dashboard rows; test loading, empty, stale and error states rather than replacing failures with synthetic data. Confirm unsigned identifiers, occurrence/ingestion time separation and time-filter boundary behavior. API/dashboard correctness enables inspection but is not independent evidence of kernel capture.

Run portable tests where applicable and privileged Linux integration tests on the declared target; record actual commands, outcomes and skips. macOS documentation work cannot validate eBPF loading, Docker/cgroup attribution or Linux overhead. Use authorized isolated workloads, disposable paths/endpoints and passive rules; do not collect contents/payloads/full environments by default or perform blocking/termination.

## 10. Reporting and limitations

Each run record must identify protocol, condition, environment, seed, load, duration, support matrix, truth quality, raw-artifact locations, counters, metrics and deviations. Preserve failed runs and reasons. Exclusions must follow frozen rules; replacement trials keep a link to the originals. Publish numerator/denominator counts, `N/A` metrics, unsupported coverage, unresolved and incorrect identity, clock uncertainty and censored outcomes alongside averages.

Present MVP execution results separately from full-thesis results. Map FR-01–FR-18 and applicable NFRs to passed, partial, unsupported, failed or not-tested evidence; no tested subset may silently stand for the whole thesis. Discuss controlled-workload bias, imperfect ground truth, observer overhead, tuning leakage, short runs, kernel specificity and the inability of passive rules to establish all malicious intent. Compare to other systems only with an actually executed comparable protocol and properly consulted sources.

No experiment log or results file is created in this documentation-only task. At the authorized evaluation phase, retain a dated checkpoint and reproducible experiment log. Until executed evidence exists, conclusions remain planned investigations and all numerical budgets remain unverified.
