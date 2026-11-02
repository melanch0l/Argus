# Research questions

These operationalizations preserve RQ1–RQ5 from [the canonical specification](../PROJECT_SPECIFICATION.md). They are questions, not claims of achieved functionality. Detailed formulas, ground truth and trial procedures are in the [evaluation plan](../experiments/evaluation-plan.md).

| ID and canonical question | Operational investigation | Canonical FR mapping | Objectives |
|---|---|---|---|
| RQ1: How effectively can eBPF observe selected runtime activities in Docker containers? | For each selected process, file, network and privilege action, establish hook availability and attempt/success semantics. Compare unique matched observations with independently recorded expected observations. Report supported-scope capture recall, total tested-scope coverage, duplicates and unexplained misses separately. | FR-01–FR-07, FR-18; FR-12–FR-15 provide inspection | O1, O4, O5 |
| RQ2: How accurately can observations be associated with originating containers? | Match collected observations to independent Docker/lifecycle/cgroup ground truth across host controls, concurrent containers, restarts, short lifetimes and reuse. Measure correct attribution, incorrect attribution and unresolved identity without excluding unknowns from the main denominator. | FR-08, FR-09; FR-12, FR-13, FR-15 provide inspection | O2, O4 |
| RQ3: Can contextual rules detect suspicious behavior with acceptable false positives? | Use policy-labelled suspicious episodes and benign lookalikes. Measure episode-level precision, recall and false-positive rate, operational false alerts per container-hour, explanation completeness and duplicate suppression. Examine both detection conditional on usable evidence and end-to-end detection including collection/identity failures. | FR-10, FR-11, FR-16; FR-12–FR-15, FR-17 provide inspection/reporting | O3, O4 |
| RQ4: What CPU, memory and latency overhead does monitoring introduce? | Compare matched disabled, monitoring-only and detection-enabled runs. Separate component CPU/memory cost from workload slowdown; measure kernel-to-agent and end-to-end latency only across compatible or uncertainty-bounded clock domains. | FR-01, FR-10, FR-12, FR-18 | O6 |
| RQ5: How does event volume affect reliability and pipeline performance? | Increase controlled offered load and concurrency; distinguish offered operations, expected hook observations and delivered events. Locate saturation through stage counters, queue occupancy, throughput, latency and attributable loss; assess recovery after overload. | FR-01–FR-07, FR-10, FR-12, FR-18; FR-13–FR-15 provide inspection | O1, O4, O5, O6 |

## Interpretation rules

1. **RQ1 is not complete coverage of Linux.** Report the support matrix and tested semantics before any aggregate. Missing support is a coverage limitation, not a ring-buffer drop.
2. **RQ2 is not conditional accuracy alone.** High accuracy among resolved events can coexist with poor identity resolution. Publish the all-eligible denominator, resolved-only accuracy and resolution rate together; unresolved never means host.
3. **RQ3 is policy-relative.** Ground-truth labels express the frozen scenario and policy, not universal maliciousness. A permitted shell is a negative case. False-positive units and alert-matching windows must be fixed before trials.
4. **RQ4 is environment-specific.** Disabled establishes workload behavior without the agent; monitoring-only isolates collection/persistence from rules. Detection cost is the incremental difference, not an assumption that all overhead comes from eBPF.
5. **RQ5 is stage-specific.** A low received-event count does not identify where loss occurred. Counter resets, retries, deduplication and unaccounted discrepancies must be disclosed.

## Planned comparisons and decision policy

The study will estimate paired condition differences with run-level uncertainty, stratify by event family and load, and retain failed or degraded trials with reasons. It will not treat individual events from one run as independent experimental repetitions. Proposed budgets for acceptable false alarms, attribution error, overhead, latency and loss will be frozen before confirmatory trials, using explicitly separate pilot evidence or an external rationale. Until then, “acceptable” is unresolved; thresholds may not be selected retrospectively to make results pass.

The MVP can answer only execution-slice versions of these questions. Full-thesis answers require all four activity families and the full management path; advanced compatibility findings must remain separately qualified.
