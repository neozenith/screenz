# ADR-0008: Fail strictly on a mismatched frame, with a numeric tolerance

| Field | Value |
|---|---|
| **Status** | Accepted, 2026-08-28 |
| **Plan ID** | ADR3.1 |
| **Provenance** | Rectangle bug history (no read back) plus the spike fact that AX rounds sizes to whole points |
| **Relates to** | ADR-0010 covers the retry before the verdict |
| **Enforced in** | internal/layout (Tolerance), internal/place (read back), internal/cli apply exit codes |

> **Lens**: Verification is never off. A rule may widen its tolerance numerically and deliberately;
> nothing may disable the check.

## Problem

### Symptom

AX set-frame is a proposal: apps clamp to minimum sizes and aspect ratios silently, and the call still returns success.

### Pain point

A silently clamped window reported as placed is the canonical false success; without read back the tool cannot be trusted at all.

## Decision

### The lens

- **Given**: apps clamp frames silently and AX rounds requested sizes to whole points
- **We prefer**: comparing every edge of the read-back frame against a per-rule numeric tolerance (points, or percent of the target size per axis; default 0.5 pt), over always-strict or a boolean tolerant switch
- **Because**: the default only absorbs AX rounding, while a known fixed-size app can widen its own rule without turning verification off for everything
- **Unless**: never; the check can be widened but not disabled, so infinite and NaN tolerances are rejected at parse time

### In practice

- Any edge outside tolerance marks the window clamped and the run exits 1 with requested versus actual.
- The stale error from a failed attempt is cleared so the verdict describes the final attempt only.

## Consequences

### Pros

- Exit 0 is a real guarantee about where every window sits.

### Cons

- Minimum-sized windows need a deliberate per-rule tolerance to go green.
