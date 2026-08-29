# screenz agent guide

macOS window-organiser CLI in Go.
Single binary, no daemon.

## Commands

- `make check` before every commit: fmt-check, vet, race, and a 100% statement-coverage gate on every package except `cmd/screenz`, `internal/mac` and `internal/place`.
- `make itest` for the real-window tier: `internal/mac`, `internal/place`, and `internal/discover`'s integration test.
  It needs the Accessibility grant and fails (never skips) without it.
- `make build` produces `bin/screenz`.
  Never invoke `go` directly; the Makefile pins the toolchain.

## Layout

- `internal/mac`: the only package that binds macOS frameworks (purego, darwin build tags); `internal/place` and `cmd/screenz` are darwin-tagged too.
  Impure, covered by itest.
- `internal/{discover,layout,rule,plan,profile,selfupdate}`: pure transforms, 100% coverage, must compile on Linux.
- `internal/place`: impure placement engine, covered by itest.
- `internal/cli`: pure command pipeline over injected `Deps`; `cmd/screenz` only wires real implementations.

## Hard rules

- Never add a CLI framework, a mock of the OS bridge, or a capability-gated `t.Skip`.
- Never print diagnostics to stdout; `--json` output must stay parseable.
- Never weaken the read-back verification: apply exits 0 only when every frame lands within tolerance.

## Decisions and language

- Check `adrs/README.md` before asking about or revisiting a design choice; self-answer from existing decisions, and record new binding decisions as a new ADR (never rewrite an accepted one).
- Use the canonical terms in `GLOSSARY.md` for all naming in code, docs and prose; when a new domain term appears, add it to the glossary in the same change.
