# ADR 0001: Distribute via CI-built artifacts on the GitHub Releases page

Date: 2026-08-28
Status: accepted

## Decision

A tag-triggered GitHub Actions workflow (`.github/workflows/release.yml`)
runs `make check` and `make dist` — `CGO_ENABLED=0`, darwin arm64 + amd64,
cross-compiled on `ubuntu-latest` — and uploads the tarballs and
`checksums.txt` to a GitHub Release. Installation is downloading the
artifact from the Releases page (`gh release download` or `curl`, neither
of which sets the quarantine xattr). No Homebrew tap, no committed `bin/`
binaries, no notarisation.

## Why

`screenz` is a tool for one user. A Homebrew tap's machinery (a second
repo, formula sha256 rewrites on every release) serves an audience that
does not exist, while the Releases page already gives versioned,
downloadable artifacts straight from CI. Gatekeeper keys on the quarantine
xattr, which CLI downloads never set, so no Apple account is needed.
`CGO_ENABLED=0` (ADR1.1) means both darwin architectures cross-compile
from any runner OS — no macOS runner.

## Rejected

- **Prebuilt tap formula + `go install`** — the multi-user answer; its
  machinery outweighs a single-user audience. The upgrade path if the tool
  ever grows one.
- **Source-build tap** — requires a Go toolchain on every client laptop.
- **Committed per-arch `bin/` binaries** — the house convention elsewhere,
  superseded here by CI-published artifacts; the repo stays binary-free.

## Deferred: daemon / LaunchAgent mode

v1 is a shell-launched tool only: it inherits the terminal app's
Accessibility grant, so the Go linker's ad-hoc signature never matters
(ADR6.2). A self-responsible daemon would be keyed on its own CDHash and
lose the grant on every rebuild — only a Developer ID certificate fixes
that, so daemon mode waits until that cost is justified.
