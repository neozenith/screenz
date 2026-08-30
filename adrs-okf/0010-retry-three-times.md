---
type: Architecture Decision
title: "Retry a mismatched frame three times with a 25 ms pause"
description: "Three attempts before judging a frame; retry, then once after 25 ms"
tags: [placement, verification]
status: accepted
accepted_on: 2026-08-28
plan_id: ADR3.3
provenance: "Rectangle's WindowManager retry loop, adopted as settled prior art"
enforced_in:
  - "internal/place (the attempt loop)"
generated: { by: human:neozenith, at: 2026-08-28T00:00:00Z }
---

> **Lens**: Give macOS the settled three attempts before judging a frame; retry immediately, then once more after 25 ms.

## Relates to

- Feeds the verdict in [ADR-0008](0008-strict-tolerance-verification.md)

## Problem

### Symptom

A window moved to another display keeps its old size on the first attempt.

### Pain point

macOS clamps the size to the source display until the position lands, so a single set produces an honest-but-avoidable clamped verdict.

## Decision

### The lens

- **Given**: the size clamp resolves once the position has landed, usually within milliseconds
- **We prefer**: re-applying immediately on mismatch and once more after 25 ms, over a single attempt or an open-ended retry loop
- **Because**: three attempts is the proven prior art and bounds the worst case per window
- **Unless**: macOS changes the clamp behaviour this recipe compensates for

### In practice

- Each attempt runs set size, set position, set size, then reads back; the loop exits early once within tolerance.

## Consequences

### Pros

- Cross-display moves converge without user-visible flicker.

### Cons

- A genuinely refusing window costs three attempts before its clamped verdict.
