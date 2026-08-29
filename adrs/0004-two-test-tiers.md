# ADR-0004: Two test tiers; the integration tier fails, never skips

| Field | Value |
|---|---|
| **Status** | Accepted, 2026-08-28 |
| **Plan ID** | ADR1.3 |
| **Provenance** | House test rules (no mocks, no capability skips) plus evidence that hosted macOS runners time out on AX |
| **Relates to** | Tests the bridge from ADR-0002 |
| **Enforced in** | Makefile (check, coverage, itest), internal/mac and internal/place integration tests, internal/cli Deps seam |

> **Lens**: Keep pure coverage universal and exercise the real OS seam on real hardware; a
> capability-gated skip is a green lie.

## Problem

### Symptom

The only seam that matters (real AX behaviour) cannot run on hosted CI, yet the suite must stay
runnable on any machine.

### Pain point

The easy outs both lie: mocks prove the fake, and `t.Skip` when the grant is missing turns the suite
green with zero coverage of the seam.

## Decision

### The lens

- **Given**: house rules forbid mocks, hand-rolled fakes and capability-gated skips, and hosted macOS
  runners accept TCC rows yet still time out on AX calls
- **We prefer**: two tiers, over one merged tier or hosted-CI AX testing
- **Because**: `make check` stays pure and runs anywhere at 100% coverage, while `make itest` opens a
  real window, places it through the shipped path, and asserts the read back
- **Unless**: hosted runners ever exercise AX reliably, which the evidence says they do not

### In practice

- check covers every package except cmd/screenz, internal/mac and internal/place; those three are
  itest's job.
- An integration test that finds the grant missing fails with the grant instruction; it never skips.
- Pure CLI tests feed real recorded machine values through injected Deps functions.

## Consequences

### Pros

- check runs on the Linux release runner; the real seam still gets exercised before every commit.

### Cons

- itest requires a local trusted GUI session and briefly opens a TextEdit window.
