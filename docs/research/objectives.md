# Research objectives

Phase 1 proposed objectives; achievement requires executed evidence. Canonical requirement meanings and priorities remain in [the specification](../PROJECT_SPECIFICATION.md). RQ numbers below preserve its five research questions; no new FR identifiers are introduced.

## Overall objective

Design, implement in later approved phases, and evaluate a passive container runtime monitoring system that connects bounded eBPF observations to validated Docker identity, explainable detection and inspectable management data, while quantifying effectiveness, uncertainty, reliability and overhead.

## Specific objectives and traceability

| Objective | Planned outcome and evidence | Canonical FR mapping | RQ mapping |
|---|---|---|---|
| O1 — Bounded kernel visibility | Probe and document runtime capabilities; collect execution and lifecycle events plus selected file, connection and privilege events. Validate record decoding and hook semantics against operation ground truth; publish unsupported cases and path limitations. | FR-01, FR-02, FR-03, FR-04, FR-05, FR-06, FR-07 | RQ1, RQ5 |
| O2 — Defensible attribution | Correlate cgroups with Docker inspection and lifecycle evidence; enrich normalized events without guessing unknown identities. Evaluate stable, restarting, short-lived and identifier-reuse cases, reporting correct, incorrect and unresolved outcomes separately. | FR-08, FR-09 | RQ2 |
| O3 — Contextual detection | Apply configurable, versioned rules and allowlists; produce evidence-linked explanations and bounded duplicate suppression. Evaluate benign lookalikes and suspicious scenarios with explicit alert matching. | FR-10, FR-11, FR-16 | RQ3 |
| O4 — Inspectable management path | Persist events and alerts, expose authenticated versioned endpoints, display actual backend data, filter by container/time and export bounded security reports. Validate ingestion, retrieval, filtering, authorization and export consistency. | FR-12, FR-13, FR-14, FR-15, FR-17 | RQ1, RQ2, RQ3, RQ5 (evidence access, not independent proof of detection) |
| O5 — Measurable reliability | Account for buffer failures, decode rejection, queue drops and delivery/persistence outcomes; test bounded queues, restart/recovery and controlled load. Separate unsupported coverage from measurable loss and unexplained discrepancies. | FR-07, FR-12, FR-18 | RQ1, RQ5 |
| O6 — Reproducible cost assessment | Compare disabled, monitoring-only and detection-enabled conditions under repeated matched workloads. Measure system-component CPU/memory, workload slowdown, stage latency and sustained throughput with a frozen environment and protocol. | FR-01, FR-10, FR-12, FR-18 (evaluated execution path) | RQ4, RQ5 |

All FR-01 through FR-18 are represented. NFR-01 through NFR-10 remain cross-cutting acceptance constraints: modular boundaries, bounded resource use, graceful shutdown, malformed-input safety, retention, API security, reproducible builds, testing and verified compatibility are not replaced by research metrics.

## Milestone interpretation

- **MVP:** O1 restricted to process execution; O2 and O3 exercised for that event family; O4 provides a minimal durable API/dashboard/filtering path; O5 and O6 measure that slice. This demonstrates integration, not all requirements.
- **Full thesis:** all six objectives across the full selected process/file/network/privilege scope, with lifecycle, rule management and report export included. Each FR must have executed acceptance evidence or an explicit unmet/partial status; unsupported mandatory coverage is a limitation, not a pass.
- **Advanced:** optional portability or deeper-analysis extensions listed in [scope](scope.md), evaluated separately rather than substituted for unfinished full-thesis requirements.

## Completion rule

Objective completion requires traceable artifacts, executed tests and measurements with denominators, environment details and limitations. Numerical success budgets are uncertain in Phase 1: they must be justified and pre-registered before confirmatory trials under the [evaluation plan](../experiments/evaluation-plan.md). Neither a design document nor an attractive dashboard is evidence of collection or detection effectiveness.
