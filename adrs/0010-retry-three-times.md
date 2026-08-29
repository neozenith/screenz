# ADR-0010: Retry a mismatched frame three times with a 25 ms pause

| Field | Value |
|---|---|
| **Status** | Accepted, 2026-08-28 |
| **Plan ID** | ADR3.3 |
| **Provenance** | Rectangle's WindowManager retry loop, adopted as settled prior art |
| **Relates to** | Feeds the verdict in ADR-0008 |
| **Enforced in** | internal/place (the attempt loop) |

> **Lens**: Give macOS the settled three attempts before judging a frame; retry immediately, then once
> more after 25 ms.

## Problem

### Symptom

A window moved to another display keeps its old size on the first attempt.

### Pain point

macOS clamps the size to the source display until the position lands, so a single set produces an
honest-but-avoidable clamped verdict.

## Decision

### The lens

- **Given**: the size clamp resolves once the position has landed, usually within milliseconds
- **We prefer**: re-applying immediately on mismatch and once more after 25 ms, over a single attempt
  or an open-ended retry loop
- **Because**: three attempts is the proven prior art and bounds the worst case per window
- **Unless**: macOS changes the clamp behaviour this recipe compensates for

### In practice

- Each attempt runs set size, set position, set size, then reads back; the loop exits early once within
  tolerance.

## Consequences

### Pros

- Cross-display moves converge without user-visible flicker.

### Cons

- A genuinely refusing window costs three attempts before its clamped verdict.
