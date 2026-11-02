# Research and delivery scope

Phase 1 documentation only. This scope refines milestones under [the canonical specification](../PROJECT_SPECIFICATION.md); it neither authorizes implementation nor changes FR priorities. All capabilities below are planned, not verified.

## Deployment and architectural boundary

Initial evaluation target: Ubuntu 24.04 LTS, the actual probed kernel/architecture, rootful Docker and cgroup v2. A privileged host agent owns eBPF loading, collection, normalization, identity resolution, detection and health. A separate management backend owns persistence, inventory, rule distribution and `/api/v1`; a React/TypeScript dashboard consumes its authenticated API. The browser and API must not expose the Docker socket. Kernel C remains separate from Go userspace; a shared versioned event model preserves identity status and timing uncertainty.

## Milestones

| Area | MVP: process-execution vertical slice | Full thesis | Advanced, optional |
|---|---|---|---|
| Observation | Runtime feature probes, one documented execution observation path, validated Go decoding (FR-01, FR-02, FR-07). State whether the path represents an attempt or successful execution. | Add process creation/exit and selected file, network-connection and privilege observations (FR-03–FR-06). Document each hook, semantics, limitations and support evidence. | Broader hook coverage or richer correlation; never assume complete syscall or packet visibility. |
| Identity | Validate Docker/cgroup attribution and metadata for execution events; retain unresolved status (FR-08, FR-09). Include host and short-lived-container controls. | Exercise lifecycle churn, stale caches, restart/reuse and all selected event families; quantify incorrect and unresolved attribution. | Assess cgroup v1, rootless Docker or other runtimes separately. |
| Detection | A configurable execution-context rule, benign allowlist cases, evidence-linked alert and duplicate suppression (initial FR-10, FR-11). | Versioned policies for unexpected processes, sensitive file actions, privilege events, outbound destinations and bounded process/network correlation; rule management (FR-10, FR-11, FR-16). | More complex correlation or separately justified detection approaches; not a required machine-learning contribution. |
| Management | Durable execution events/alerts, minimal authenticated ingestion/query API, actual-data view and container/time filters (initial FR-12–FR-15). | Full scoped persistence, versioned REST, inventory, health, event-family views, rule management and report export (FR-12–FR-17). | Additional visualization or integrations. |
| Reliability/evaluation | Measurable kernel/pipeline loss, bounded processing and repeated three-condition execution trials (FR-18). | All-family ground truth, attribution and detection assessment, overhead/load/recovery experiments, secure API and retention tests. | Expanded multi-environment or scale study, explicitly separated from the primary results. |

MVP is an early end-to-end integration milestone, **not** completion of the thesis. FR-03–FR-06, full FR-16/FR-17 and comprehensive family-specific coverage are not silently deferred to “advanced.” Full-thesis completion requires executed evidence for FR-01–FR-18 and applicable NFRs. If required support cannot be established, record partial/unmet status and seek an explicit scope decision rather than redefining success after evaluation.

## Activity boundaries

- **Process:** execution plus creation/exit in the full thesis; document thread/process distinctions, host versus namespace identifiers and what failed executions are observable.
- **File:** selected access/modification actions and available path metadata. No complete filesystem audit claim; distinguish attempted actions from confirmed outcomes and explain truncation, relative paths, rename and namespace limitations.
- **Network:** selected connection lifecycle/actions and endpoint metadata, not payload capture or complete traffic accounting. State protocol, address-family and outbound/inbound coverage explicitly.
- **Privilege:** selected credential/capability or related security-relevant actions supported by the tested hooks. A request is not proof of a successful privilege change; no universal escape-detection claim.

The exact hook/action matrix is a later compatibility-validation deliverable and must be frozen before confirmatory evaluation. Unsupported cells remain visible in the report.

## Exclusions and safety

No active enforcement, process killing, container termination, exploit deployment, kernel modification or claim to replace host isolation. No file contents, packet payloads or full environments by default. Use isolated, authorized test workloads and reversible policy-labelled scenarios, not destructive attacks. Collect only evidence needed for the selected rules; protect test metadata, credentials and exported records through access control and retention.

Multi-host orchestration, Kubernetes, cgroup v1, rootless support and cross-kernel generalization are not initial compatibility promises. Development on macOS supplies documentation or portable-test evidence only, not Linux/eBPF validation. No monitoring data, benchmark or citation is invented to fill a gap.

## Acceptance evidence

Maintain a requirement-to-test status with supported, partial, unsupported, failed and not-tested outcomes. Use the [objectives](objectives.md) and [evaluation plan](../experiments/evaluation-plan.md) to distinguish functionality, scientific measurements and operational limitations. Phase 1 delivers these plans only; later phases require explicit approval. A phase checkpoint and experiment log are recommended when authorized, with no implicit commit or branch creation.
