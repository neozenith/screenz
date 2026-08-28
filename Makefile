PROJECT_ROOT := $(CURDIR)
TMP_ROOT := tmp
GO_VERSION := 1.27.0

# Detect the host so the same Makefile bootstraps a toolchain on macOS and Linux
# (the release workflow cross-compiles darwin binaries from ubuntu-latest).
MACHINE_OS := $(shell uname -s)
MACHINE_ARCH := $(shell uname -m)
GO_OS := $(if $(filter Darwin,$(MACHINE_OS)),darwin,$(if $(filter Linux,$(MACHINE_OS)),linux,unsupported))
GO_ARCH := $(if $(filter arm64 aarch64,$(MACHINE_ARCH)),arm64,$(if $(filter x86_64 amd64,$(MACHINE_ARCH)),amd64,unsupported))

# Pinned Go toolchain checksums, keyed by GOOS_GOARCH (verified against go.dev).
GO_SHA256_darwin_arm64 := 90493b3bbd5e10f91d12153198bf1994fd756399b4fec93b49b0c6e2acdeeb3e
GO_SHA256_darwin_amd64 := d3314e25496e4381d71a5c51d2907e7af655d199f6780b549f015bd85fef4986
GO_SHA256_linux_arm64 := 51798d2c42d0e1c6ed7fd9f48728b4193abac9e8aad6dbac2fe96a81f5909bda
GO_SHA256_linux_amd64 := 675c26c449cbb18fc24b74650de1eabbae6e16f64326fd85a283fb3b58280685

GO_ARCHIVE := go$(GO_VERSION).$(GO_OS)-$(GO_ARCH).tar.gz
GO_SHA256 := $(GO_SHA256_$(GO_OS)_$(GO_ARCH))

# macOS ships shasum; Linux ships sha256sum. Both print "<hash>  <file>".
SHA256SUM := $(if $(filter Darwin,$(MACHINE_OS)),shasum -a 256,sha256sum)

GO ?= $(TMP_ROOT)/.go-toolchain/bin/go

export GOCACHE := $(PROJECT_ROOT)/$(TMP_ROOT)/.go-build
export GOMODCACHE := $(PROJECT_ROOT)/$(TMP_ROOT)/.go-mod
export GOPATH := $(PROJECT_ROOT)/$(TMP_ROOT)/.go-path
export TMPDIR := $(PROJECT_ROOT)/$(TMP_ROOT)/.system-tmp
export GOTMPDIR := $(TMPDIR)

BINARY := $(PROJECT_ROOT)/bin/screenz
COVERAGE := $(TMP_ROOT)/coverage.out
DIST := dist

# Release identity, injected with the goreleaser variable names (G6).
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BUILT_BY ?= make
LDFLAGS := -s -w \
	-X github.com/joshpeak/screenz/internal/cli.version=$(VERSION) \
	-X github.com/joshpeak/screenz/internal/cli.commit=$(COMMIT) \
	-X github.com/joshpeak/screenz/internal/cli.date=$(DATE) \
	-X github.com/joshpeak/screenz/internal/cli.builtBy=$(BUILT_BY)

# cmd/screenz is darwin-only by build tag (G6); on a linux runner the check
# tiers cover the pure packages, which build and test everywhere.
PKGS := $(if $(filter darwin,$(GO_OS)),./...,./internal/...)

# Coverage gate scope: every package except cmd/screenz and the two packages
# that talk to the real window server (internal/mac, internal/place) — those
# are exercised for real by `make itest` on a trusted GUI session (ADR1.3).
COVEREXCLUDE := -e /cmd/screenz -e /internal/mac -e /internal/place

.DEFAULT_GOAL := help

.PHONY: all bootstrap build build-all check clean coverage dist fmt help install itest prepare race release test tidy vet

all: check build ## Run all confidence checks and build the local binary.

prepare: ## Create every project-local temporary and cache directory.
	mkdir -p $(GOCACHE) $(GOMODCACHE) $(GOPATH) $(TMPDIR)

# Download and verify the pinned host Go toolchain when it is absent.
$(GO): | prepare
	@test "$(GO_OS)" != "unsupported" || (echo "unsupported operating system: $(MACHINE_OS)" >&2; exit 1)
	@test "$(GO_ARCH)" != "unsupported" || (echo "unsupported architecture: $(MACHINE_ARCH)" >&2; exit 1)
	mkdir -p $(TMP_ROOT)/.downloads $(TMP_ROOT)/.go-toolchain
	curl -fsSL https://go.dev/dl/$(GO_ARCHIVE) -o $(TMP_ROOT)/.downloads/$(GO_ARCHIVE)
	@test "$$($(SHA256SUM) $(TMP_ROOT)/.downloads/$(GO_ARCHIVE) | awk '{print $$1}')" = "$(GO_SHA256)"
	tar -xzf $(TMP_ROOT)/.downloads/$(GO_ARCHIVE) -C $(TMP_ROOT)/.go-toolchain --strip-components 1

help: ## List the documented Make targets.
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z0-9_-]+:.*## / {printf "%-12s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

bootstrap: $(GO) | prepare ## Download and verify the host's pinned project-local Go toolchain.

fmt: | prepare $(GO) ## Format all Go source files.
	$(GO) fmt $(PKGS)

tidy: | prepare $(GO) ## Resolve dependencies and update go.mod and go.sum.
	$(GO) mod tidy

test: | prepare $(GO) ## Run the unit test suite.
	$(GO) test $(PKGS)

race: | prepare $(GO) ## Run the unit tests with the race detector.
	$(GO) test -race $(PKGS)

coverage: | prepare $(GO) ## Run tests and require 100 percent statement coverage.
	$(GO) test -covermode=atomic -coverprofile=$(COVERAGE) $$($(GO) list $(PKGS) | grep -v $(COVEREXCLUDE))
	$(GO) tool cover -func=$(COVERAGE)
	@test "$$($(GO) tool cover -func=$(COVERAGE) | awk '/^total:/ {print $$3}')" = "100.0%"

vet: | prepare $(GO) ## Run Go's static analysis checks.
	$(GO) vet $(PKGS)

check: fmt vet race coverage ## Format, vet, race-test, and verify full coverage.

itest: | prepare $(GO) ## Run the real-window integration tier (fails, never skips, without the Accessibility grant).
	$(GO) test -tags integration -count=1 ./internal/mac/... ./internal/place/... ./internal/discover/...

build: | prepare $(GO) ## Build a CGO-free executable in bin/.
	mkdir -p $(dir $(BINARY))
	CGO_ENABLED=0 $(GO) build -trimpath -o $(BINARY) ./cmd/screenz

build-all: | prepare $(GO) ## Cross-build CGO-free darwin binaries for both architectures.
	mkdir -p $(dir $(BINARY))
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY)-darwin-arm64 ./cmd/screenz
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY)-darwin-amd64 ./cmd/screenz

dist: build-all ## Tar both binaries and write dist/checksums.txt.
	rm -rf $(DIST)
	mkdir -p $(DIST)
	for arch in arm64 amd64; do \
		cp $(BINARY)-darwin-$$arch $(DIST)/screenz; \
		tar -czf $(DIST)/screenz_$(VERSION)_darwin_$$arch.tar.gz -C $(DIST) screenz; \
		rm $(DIST)/screenz; \
	done
	cd $(DIST) && $(SHA256SUM) *.tar.gz > checksums.txt

release: ## Tag VERSION=vX.Y.Z and push; the release workflow builds and uploads.
	@echo "$(VERSION)" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+' || (echo "usage: make release VERSION=vX.Y.Z" >&2; exit 1)
	git tag $(VERSION)
	git push origin $(VERSION)

install: | prepare $(GO) ## Install the CGO-free executable into the project-local GOPATH.
	CGO_ENABLED=0 $(GO) install ./cmd/screenz

clean: ## Remove the project-local tmp/ tree, dist/ and the local build.
	-chmod -R u+w $(TMP_ROOT) 2>/dev/null
	rm -rf $(TMP_ROOT) $(DIST)
	rm -f $(BINARY) $(BINARY)-darwin-arm64 $(BINARY)-darwin-amd64
