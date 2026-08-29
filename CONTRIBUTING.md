# Contributing

Everything runs through `make`; it bootstraps its own pinned Go toolchain
into `tmp/` on first use, so no local Go install is required.

## Build and test

```sh
make check   # fmt-check, vet, race tests, coverage gate
make itest   # real-window integration tier (see below)
make build   # CGO-free binary in bin/
make help    # every target
```

Both tiers must be green before a commit (ADR-0004, `adrs/`):

- **`make check`** is pure and runs anywhere, including the Linux release
  runner. It enforces 100% statement coverage on every package except
  `cmd/screenz`, `internal/mac` and `internal/place`.
- **`make itest`** runs `//go:build integration` tests against the real
  macOS window server. It needs the Accessibility grant on your terminal
  app (`screenz doctor` walks you through it) and fails, never skips,
  without it. It opens and closes a real TextEdit window.

## House rules

- Stdlib `flag` only; no CLI frameworks (ADR-0013).
- No mocks or fakes standing in for the OS bridge; pure tests feed real
  recorded values through injected `cli.Deps` functions, and the real seam
  is covered by `make itest` (ADR-0004).
- All OS calls live in `internal/mac` (darwin-only build tags); every other
  package is a pure transform that must compile on Linux.
- Diagnostics go to stderr so `--json` stdout stays parseable. Exit codes:
  0 success, 1 runtime error, 2 usage error.
- Check `adrs/` before re-litigating a design choice; record new binding
  decisions there.

## Releasing

```sh
make release VERSION=vX.Y.Z
```

Tags and pushes; the GitHub workflow runs `make check`, cross-compiles both
darwin architectures, and uploads the tarballs plus `checksums.txt` to the
release (ADR-0001). Proof-of-execution transcripts live in
`docs/evidence/` with ISO dates in the filenames.
