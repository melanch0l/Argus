# Kernel source and reproducible generation boundary

Reserved for Phase 3 eBPF C sources. No kernel program, C ABI, generated object or loader is implemented in Phase 2. `make bpf-generate` intentionally fails; `go generate ./...` would currently generate nothing and must not be reported as an eBPF build.

## Pinned tooling

`go.mod` declares `github.com/cilium/ebpf/cmd/bpf2go` as a Go tool from **Cilium eBPF v0.22.0**. Resolve through `go tool bpf2go`, never `go install ...@latest`. `make bpf-tools` checks tool resolution/help only, not compiler/kernel compatibility. The tool and future runtime library share one module version and checksums in `go.sum`.

## Phase 3 generation convention (not executed)

After approval, put kernel C here, verifier/ABI fixtures under `tests/integration`, and generated bindings and embedded objects in a dedicated Linux-only loader package. Use explicit `//go:generate go tool bpf2go ...` directives in that package; give the generator an explicit Clang binary, target architectures, C flags, strip tool and source path. Do not invoke the generator from this directory without the Go package context.

For Linux amd64/arm64, select the actual `-target amd64,arm64` architecture modes after inspecting `go tool bpf2go -help`; the tool derives target-specific macros. Never substitute the macOS host architecture for the Linux target. Record endianness, explicit ABI offsets/sizes and generated layout tests. Keep generated source/object artifacts versioned once introduced and verify clean regeneration in CI. CO-RE inputs need provenance: obtain `vmlinux.h` from BTF on the documented kernel/reference baseline and record its hash, bpftool version and generation command. Do not generate a fictitious header or silently download unknown BTF.

Ubuntu 24.04's Clang/LLVM 18 series is a candidate, not a verified pin. An actual target report must record package revisions, BPF backend availability, BTF, helpers, program/map types, hook availability and effective permissions. Apple Clang alone is not a verified BPF toolchain. Phase 2 feature probing does not establish verifier acceptance or intended-program load/attach.

No monitoring attachment, map allocation, kernel compatibility claim or event transport is supplied by this README.
