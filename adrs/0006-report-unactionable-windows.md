# ADR-0006: Report windows the tool cannot act on instead of hiding them

| Field | Value |
|---|---|
| **Status** | Accepted, 2026-08-28 |
| **Plan ID** | ADR2.2 |
| **Provenance** | The planning session's negative measures: "moved 9 of 10" must never look like success |
| **Relates to** | [ADR-0008](0008-strict-tolerance-verification.md) applies the same honesty to frames |
| **Enforced in** | internal/discover (states), internal/plan (Skipped), internal/cli status and apply |

> **Lens**: Never hide what the tool cannot act on; every un-actionable window carries a state and every incomplete sweep fails the run.

## Problem

### Symptom

Minimized, hidden and other-Space windows cannot be placed, and an app whose AX enumeration failed contributes no windows at all.

### Pain point

Silently dropping any of them makes a partial result look like a complete one, which is the exact failure this tool exists to prevent.

## Decision

### The lens

- **Given**: Spaces have no public API (the offscreen skip is permanent) and AX enumeration can fail per app
- **We prefer**: reporting a state (minimized, hidden, sheet, dialog, offscreen, unknown) and listing skips with reasons, over filtering un-actionable windows out
- **Because**: the user must see what was not done and why, not infer it from absence
- **Unless**: never; this one is unconditional

### In practice

- status shows every window with its state; a failed frame read is state unknown, never a zero-valued frame passed off as real.
- apply lists matched-but-skipped windows with reasons and refuses to run when an application could not be enumerated.

## Consequences

### Pros

- Partial results are visibly partial.

### Cons

- apply is stricter than fire-and-forget tools; a hung app blocks the run until addressed.
