SHELL := /bin/sh
GO ?= go
NPM ?= npm

.PHONY: help deps build test vet race fmt fmt-check check cross-build frontend-check envcheck envcheck-strict bpf-tools bpf-generate clean

help:
	@printf '%s\n' 'Phase 2 portable foundation (no monitoring functionality)' 'make deps            Download/verify pinned Go and frontend dependencies' 'make check           Format, test, vet, race, build, cross-build and frontend checks' 'make envcheck        Save actual local environment evidence to reports/environment.json' 'make envcheck-strict Require Ubuntu prerequisites; optional privileged probe must be run explicitly' 'make bpf-tools       Resolve and display help for the pinned bpf2go tool' 'make bpf-generate    Explain the Phase 3 generation gate (fails until approved programs exist)'

deps:
	$(GO) mod download
	$(GO) mod verify
	$(NPM) --prefix web ci

build:
	mkdir -p bin
	$(GO) build -trimpath -buildvcs=false -o bin/argus-agent ./cmd/agent
	$(GO) build -trimpath -buildvcs=false -o bin/argus-backend ./cmd/backend
	$(GO) build -trimpath -buildvcs=false -o bin/argus-envcheck ./cmd/envcheck

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

race:
	$(GO) test -race ./...

fmt:
	gofmt -w cmd internal

fmt-check:
	sh scripts/check-format.sh

cross-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build ./cmd/agent ./cmd/backend ./cmd/envcheck
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build ./cmd/agent ./cmd/backend ./cmd/envcheck

frontend-check:
	$(NPM) --prefix web test

check: fmt-check test vet race build cross-build frontend-check

envcheck:
	mkdir -p reports
	$(GO) run ./cmd/envcheck --output reports/environment.json

envcheck-strict:
	mkdir -p reports
	$(GO) run ./cmd/envcheck --strict --output reports/environment-strict.json

bpf-tools:
	$(GO) tool bpf2go -help

bpf-generate:
	@printf '%s\n' 'Not implemented: Phase 3 kernel programs and ABI require separate approval.' 'See bpf/README.md for pinned tool invocation and target prerequisites.' >&2
	@exit 2

clean:
	rm -rf bin web/dist web/.test-dist
