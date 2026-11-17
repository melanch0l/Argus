# Phase 2 dependency and reproducibility matrix

**Recorded:** 2026-11-02. See the [environment record](environment.md) and [experiment log](../experiments/phase-02-record.md). Pins below are observed/resolved versions, not assertions of current security support or Ubuntu/kernel compatibility. No new dependency implements a later-phase feature.

## Introduced foundation dependencies

| Component | Exact baseline / pin | Purpose and observed evidence |
|---|---|---|
| Go module | `github.com/melanch0l/Argus`; Go language minimum `1.25.0`; toolchain `1.27.3` in go.mod / .go-version | One module separates `cmd/agent`, `cmd/backend`, `cmd/envcheck` and internal boundaries. Local native/race tests and Linux amd64/arm64 cross-builds passed. Runtime CLI/config/logging/environment code uses standard library only. |
| Cilium eBPF / bpf2go | `github.com/cilium/ebpf v0.22.0`; Go tool declaration | Selected as the verified upstream module (requires Go 1.25.0) for future loading/generation. `go mod tidy`, `go mod verify` and `go tool bpf2go -help` passed. No BPF C/object/runtime loader exists yet. |
| Go transitive module | `golang.org/x/sys v0.43.0` | Pinned by the generator's dependency graph in go.mod/go.sum. No independent fake syscall layer added. Test-only upstream dependencies also have recorded checksums. |
| Node / npm | `22.18.0` / `10.9.3`; root/web .nvmrc and packageManager | Actual Node/npm host versions. `web/package.json` requires stable Node >=22.18.0 <23 and npm >=10.9.3 <11; exact CI baseline and npm version are pinned. |
| React / React DOM | `19.3.1` each | Static foundation notice only; no dashboard or telemetry. Clean install/build/static-render tests passed. |
| TypeScript | `7.0.2` | Strict frontend and test typechecking; local compiler checks passed. |
| Vite | `8.3.5` | Production asset build; upstream engine requirement ^20.19.0 or >=22.12.0 is compatible with observed Node 22.18.0. No dev server was started. |
| Frontend type packages | `@types/node 22.20.5`, `@types/react 19.3.0`, `@types/react-dom 19.3.0` | Exact manifest entries and package-lock.json. No separate third-party test runner required: Node's built-in runner runs compiled tests. |
| PostgreSQL dev image | `postgres:17.11@sha256:3b0cda93be7c8dba2ed484d19393d93845dc3dafb88973e248ba8e6d99d784e1` | Official Hub/Registry metadata verified including Linux amd64/arm64 index; no Docker pull/run. [Deployment provenance](../../deploy/README.md#verified-image-pin). |
| GitHub actions | Full commit SHA pins | Checkout v7.0.1, setup-go v7.0.0 and setup-node v7.1.0 source metadata checked. [SHA/provenance table](../../.github/workflows/README.md#action-provenance). Actions have not run remotely. |

Exact frontend direct versions and all transitive integrity records live in [package.json](../../web/package.json) and [package-lock.json](../../web/package-lock.json). The lockfile—not a prose copy—is authoritative for transitive pins. Source license files/notices remain with upstream packages; review licenses and any redistribution obligations before packaging the final system. No dependencies are vendored or relicensed here.

## Deferred or unverified dependencies

| Item | Current status and phase boundary |
|---|---|
| Clang/LLVM / strip / bpftool | Apple Clang 21.0.0 is present; llvm-config, llc and bpftool are absent. Ubuntu Clang/LLVM 18 is a candidate only. Exact target distro package revisions and BPF-target checks require Ubuntu access, not guessed package pins. |
| Go patch/toolchain provenance | Bootstrap Go on the host is 1.25.0; project toolchain is now **1.27.3**. An initial go.dev request timed out, but official `golang/go` matching-tag metadata subsequently identified 1.27.3, and `GOTOOLCHAIN=go1.27.3 go version` downloaded/resolved it successfully. go.mod keeps the Cilium-compatible 1.25 language minimum and declares toolchain go1.27.3; CI reads .go-version=1.27.3. Official metadata source: `https://api.github.com/repos/golang/go/git/matching-refs/tags/go1.`. This fixes the tested compiler baseline rather than silently trusting the older bootstrap compiler. Vulnerability scanning/advisory review remains separate and unperformed. |
| Docker Engine / Compose v2 | Not installed on current PATH. Do not infer availability from valid YAML or Registry metadata. Record exact client/server/API/Compose versions and host match on target. |
| pgx v5 / migration runner | Intended for Phase 7; **no driver dependency, repository, pool or SQL migrations are added in Phase 2**. The observed upstream pgx metadata is not an approved pin or compatibility test. |
| PostgreSQL client vs server | Host psql/pg_config 17.6; no responding server on 127.0.0.1:5432. Dev image patch is independently pinned to 17.11; a client version is not server evidence. |
| YAML parser / Docker Go SDK / HTTP libraries | Deferred until their feature phases justify exact compatible versions. Phase 2 config uses bounded strict JSON to avoid unnecessary dependencies. No production policy parser/resolver/REST framework introduced. |
| Test/lint tooling | Go test/vet/race and Node built-in tests have no extra project dependency. `actionlint v1.7.12` was invoked as a pinned external validation command; it is not imported into application go.mod. Ruby's installed YAML parser was used for source syntax checks only, not Compose runtime validation. |

## Reproducible workflow

```sh
go mod download
go mod verify
go mod tidy -diff
make bpf-tools
npm --prefix web ci
make check
npm --prefix web audit --audit-level=low
```

`go.sum` and the npm lockfile authenticate expected downloads; they do not themselves establish trustworthy upstream behavior or absence of vulnerabilities. `go tool bpf2go` resolves the declared tool from the same Cilium version; do not install an unpinned global generator. Builds use `-trimpath -buildvcs=false` and avoid injecting wall-clock timestamps; Go build information is deterministic version/phase metadata. Linux cross-builds disable CGO, build all three commands, and never run them or invoke kernel tools. `make bpf-generate` intentionally fails until Phase 3 programs are approved and present.

A clean npm install and full audit reported **0 vulnerabilities** in the current graph on 2026-11-02. This is time-specific registry evidence, not a security guarantee. Go vulnerability scanning, container image scanning and OS advisory review were not performed; track these separately before target acceptance/deployment. Never describe an old local patch pin as the newest secure release without verification.

## Update policy

Change version manifests, lockfiles, matching tool declarations and this record together; use clean installs, graph checks, portable tests, cross-builds and upstream license/security review. Re-verify action commits and container index digests from authoritative metadata. A digest prevents silent image changes but also prevents automatic security updates. Review PostgreSQL 17 patch/base-image updates with retained-volume and recovery tests before rollout. Keep runtime/kernel compatibility and benchmark claims tied to actual target evidence, not dependency metadata.
