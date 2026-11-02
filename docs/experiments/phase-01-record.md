# Phase 1 documentation record

**Date:** 2026-11-02

**Phase:** Requirements and architecture

**Status:** Documentation prepared for review; implementation not started; user acceptance and Phase 2 approval pending.

## Starting state and scope

The project directory was empty when inspected. The supplied project rules and thesis baseline were used to create `AGENTS.md` and the sole canonical specification at `docs/PROJECT_SPECIFICATION.md`, then read before elaborating the design. Work was documentation-only, in a macOS editor session. No Ubuntu host was inspected.

## Artifacts

- Root README and persistent agent instructions.
- Canonical specification with FR-01–FR-18 and NFR-01–NFR-10.
- Four research documents, eight architecture documents and implementation roadmap.
- Evaluation plan and this documentation record; no experimental observations or result tables.
- Ten phase prompts, stored but not executed.

The [roadmap](../implementation/roadmap.md) contains the twelve-item acceptance checklist and unresolved decisions.

## Decisions recorded

- Separate privileged host agent from management backend and dashboard; agent owns detection.
- Initial rootful Docker/cgroup v2 target, actual-kernel support probes and passive monitoring.
- Versioned immutable events, explicit unresolved identity, lifecycle-aware attribution, separate clock domains and lossless unsigned identifier encoding.
- Bounded pipeline/outbox, idempotent atomic events-plus-alerts ingestion and evidence-aware retention.
- Complete revisioned inventory snapshots and scoped immutable rule distribution; stale/unknown states remain visible.
- MVP is an execution-event vertical slice; full thesis still requires selected file/network/privilege coverage and report export.
- Dependencies, raw ABI layout, operational budgets and numerical performance/detection targets require later verification or pre-registration.

## Validation and limitations

Executed a repository-wide Python 3 documentation check (`python3 - <<'PY' ... PY`) from the project root. Outcome: PASS. All 28 required Markdown documents were present, 109 local file links resolved, seven Mermaid blocks were present with balanced fences, and both JSON examples parsed. UTF-8, final newlines, absence of tabs, requirement IDs FR-01–FR-18/NFR-01–NFR-10 and all twelve documented checklist entries passed. Only the expected documentation artifacts were present; no duplicate root specification, source tree, Go module, frontend package manifest or Git directory existed. The checker was an inline validation command, not an added project script. These checks do not establish semantic correctness of every contract or successful Mermaid rendering.

No Go tests, vet, race tests, frontend builds, eBPF compilation, verifier/load/attach tests, Docker tests, API/database tests or benchmarks were run: no implementation exists, and the current session is not the Ubuntu target. Primary reference links are supplied entry points, not verified citations or proof of compatibility. No dependencies were installed, and no Git repository, commit or branch was created.

## Next gate

Human review of the [Phase 1 checklist](../implementation/roadmap.md#3-phase-1-acceptance-checklist) and unresolved decisions, followed by explicit Phase 2 authorization. Phase 2 must verify the actual Ubuntu environment and choose reproducible compatible pins before claiming support.
