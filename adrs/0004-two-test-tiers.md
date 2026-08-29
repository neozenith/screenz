# ADR 0004: Two test tiers; the integration tier fails, never skips

Plan ID: ADR1.3 | Date: 2026-08-28 | Status: accepted

## Decision

`make check` runs pure tests (geometry, parsing, YAML round trips, CLI
pipelines on real recorded values) at 100% statement coverage for every
package except internal/mac, internal/place and cmd/screenz. `make itest`
runs `//go:build integration` tests against the real window server: it
opens a real TextEdit window, places it through the shipped code path,
asserts the read-back frame, and closes it. An integration test that finds
the grant missing fails with the grant instruction; it never calls t.Skip.

## Why

House rules forbid mocks, hand-rolled fakes and capability-gated skips.
Hosted macOS runners accept TCC rows yet still time out on AX calls, so the
real seam is exercised locally before every commit instead.

## Rejected

- One tier with real windows inside `go test`: no machine without a GUI
  session could run `make check`.
- Hosted CI with TCC.db inserts: unreliable by evidence.
