# Future API service boundary (Phase 8)

This directory is reserved for browser-to-management API integration. Phase 2
provides only `apiPath`, a pure same-origin `/api/v1` path builder. It does not
send requests, implement REST endpoints, authenticate, or return telemetry. The
static application deliberately does not import or call it. Tests use synthetic
path strings only, never claimed monitoring observations.

The helper accepts conservative ASCII route segments (letters, digits,
underscores and hyphens), preserving UUIDs and decimal strings; it rejects full
URLs, traversal, separators, queries and fragments. It is not an authorization
boundary, request client, full URL builder or event-schema validator. Query
construction and any wider segment grammar need separate Phase 8 design/tests.

Before implementing integration in Phase 8:

- Follow `docs/architecture/api-specification.md` and `event-schema.md`, relative
  to the repository root. Confirm the Phase 7 implementation contract first.
- Keep requests/DTO decoding here, not in components or eBPF-related packages.
- Require authentication/authorization; no agent/admin credentials in browser
  bundles or persistent browser storage. All `VITE_*` values are public.
- Preserve uint64 decimal strings and exact timestamp representations. Do not
  parse identifiers through JavaScript `number`.
- Implement bounded pagination/requests, cancellation, safe errors and explicit
  loading, empty, stale and error states using actual backend responses.
- Never call Docker/kernel interfaces, infer coverage from zero alerts, or use
  generated data as a production fallback. No monitoring fixtures ship here.

No API origin, development proxy, fetch wrapper, tokens or mock REST server is
configured in this foundation.
