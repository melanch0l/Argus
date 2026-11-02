# Problem statement

## Status and authority

Phase 1 research design only; no implementation, compatibility or evaluation result is claimed. [The project specification](../PROJECT_SPECIFICATION.md) is authoritative. This document motivates its requirements rather than introducing a competing specification.

## Research problem

Containers share a host kernel. Application logs alone do not necessarily expose the selected process executions, file operations, network connections and privilege-related actions needed to investigate runtime behavior. Kernel observations can add evidence, but an observation is not automatically a complete activity record, a correct container attribution or proof of malicious intent.

The research problem is to determine whether a passive eBPF-based host agent, implemented with Go userspace processing, can turn explicitly bounded kernel observations into container-aware, explainable alerts and durable, inspectable records at a measured resource cost. The study must distinguish what can be observed on the tested kernel from what is actually delivered, attributed and detected.

## Sources of uncertainty

- **Observation semantics:** an attempted operation differs from a successful one. A connection event is not a packet stream; a file hook may not reconstruct a complete path. Hook availability, helper support, permissions and attachment success require actual probes.
- **Identity:** a PID or a container-ID-looking path is insufficient evidence. Docker lifecycle metadata and cgroup association can become stale during restarts, rapid termination or identifier reuse. Unknown identity must remain unresolved, not silently become host activity.
- **Detection:** shells, administration utilities and sensitive-file access can be legitimate. Context, allowlists, rule versions and explicit scenario labels are needed to assess false alarms as well as suspicious activity.
- **Reliability:** unsupported operations, hook-semantic gaps, kernel buffer failures, userspace drops and delivery failures have different causes. A single missing-event percentage cannot diagnose them.
- **Cost and timing:** workload slowdown differs from monitoring CPU consumption. Kernel occurrence time and management ingestion time may use different clock domains, making naive timestamp subtraction misleading.

## Intended contribution

The intended contribution is an integrated, reproducibly evaluated design: validated container correlation; a versioned event pipeline with explicit uncertainty and health accounting; contextual rules with supporting evidence; and an authenticated management API and actual-data dashboard. Its value will be assessed through controlled ground-truth workloads and repeated disabled, monitoring-only and detection-enabled trials, not by asserting superiority over untested alternatives.

The MVP establishes a process-execution vertical slice. The full thesis additionally studies process lifecycle, selected file, network and privilege observations; the MVP alone cannot establish full-thesis success. See [scope](scope.md), [objectives](objectives.md), [questions](questions.md) and the [evaluation plan](../experiments/evaluation-plan.md).

## Claim boundaries

Results will apply to the recorded Ubuntu 24.04 LTS host, actual kernel, rootful Docker/cgroup v2 configuration, selected hooks and tested workloads. They will not establish complete intrusion detection, exploit prevention, universal Linux support or capture of every container action. Monitoring never blocks operations or terminates workloads. No file contents, network payloads or full environment variables are collected by default.

This problem framing is a research rationale, not a literature-review finding. The canonical specification supplies technical reference entry points; exact versions, access dates and citations must be recorded when sources are actually consulted. No external reference or comparative result is asserted here.
