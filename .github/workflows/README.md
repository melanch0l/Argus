# Portable foundation CI

Workflows assume this repository's root Go module is
`github.com/melanch0l/Argus`, with `go.mod`, `cmd/agent`, `cmd/backend`, and `web/`
at the root (not under an extra `Thesis/` or `Argus/` directory).

- `frontend.yml`: Node 22.18.0 from `web/.nvmrc`, npm 10.9.3, `npm ci`, production
  build/typecheck and Node tests via `npm test`, then full `npm audit` with a
  failure threshold of low. No silently ignored advisory failures.
- `go.yml`: Go 1.27.3 selected by `setup-go`'s `go-version-file: .go-version`, module
  download/verify, gofmt enforcement, unit tests, vet, race tests, native command
  builds and CGO-disabled Linux amd64/arm64 cross-builds of **only** agent and
  backend. No dependency cache until a stable `go.sum` exists; a dependency-free
  foundation may have no `go.sum`. The workflow needs the separately owned root
  Go module and command files; it does not create or silently skip missing ones.

All jobs use `ubuntu-24.04`, read-only repository permissions, no persisted Git
credentials, no deploy step, no privileged containers, no environment/secret
uploads and no artifact uploads. Only npm's package cache is enabled. Pull
requests use `pull_request`, not `pull_request_target`.

The GitHub runner is a portable build/test environment, **not** the actual
monitored Ubuntu target. These workflows deliberately do not load eBPF, call
Docker, run target feature probes, or claim hook/helper/BTF/cgroup/rootful-Docker
support. Cross-compilation is not runtime validation. Race tests use the hosted
runner's C toolchain; cross-builds explicitly disable CGO. Future kernel and
Docker tests need a separately approved target suite and captured evidence.

## Local validation record — 2026-11-02

- `go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12 frontend.yml go.yml`
  from this directory: passed with no workflow diagnostics. This validation tool
  version was confirmed from upstream GitHub release metadata; it adds no project
  Go dependency. Go downloaded its validation-tool dependencies to the tool cache.
- Ruby YAML parsing of both workflows and `deploy/compose.yaml`: passed.
- Compose source checks for a single database service, loopback-only port,
  file-backed password, PostgreSQL 17 data mount and no privileged flag: passed.
  These are source checks, not Docker Compose rendering/runtime validation.
- Frontend clean install/build/typecheck/tests and full npm audit: passed locally;
  see `web/README.md` for exact commands and results.

No GitHub Actions run was triggered. The separately owned root `go.mod` was not
present at validation time, so module verification, application Go tests/vet/race
and native/Linux cross-builds were not run here. Docker was absent, so Compose
render/start/health/authentication/durability/backup checks were not executed.
No actual target-kernel compatibility evidence is claimed.

## Action provenance

Exact immutable pins resolved from authoritative GitHub REST tag/commit metadata
on **2026-11-02**, with `action.yml` fetched at each commit:

| Action | Release tag | Commit |
|---|---|---|
| `actions/checkout` | `v7.0.1` | `3d3c42e5aac5ba805825da76410c181273ba90b1` |
| `actions/setup-node` | `v7.1.0` | `949feb2413d6458794dcd2491c4babbbce0c15c1` |
| `actions/setup-go` | `v7.0.0` | `b7ad1dad31e06c5925ef5d2fc7ad053ef454303e` |

Metadata endpoints:

- <https://api.github.com/repos/actions/checkout/tags>
- <https://api.github.com/repos/actions/setup-node/tags>
- <https://api.github.com/repos/actions/setup-go/tags>

These actions use the runner-provided Node 24 action runtime, separate from the
application's Node 22 toolchain; use current GitHub-hosted runners. Hosted runner
images are moving infrastructure, not pinned VM digests. When updating, verify
new upstream release tags, commit objects, runtime requirements and security
notes, replace the full SHA and matching comment together, and rerun CI. Metadata
verification and local tests do not mean a GitHub Actions run has executed.
