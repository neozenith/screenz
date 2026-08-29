# ADR 0010: Retry a mismatched frame three times with a 25 ms pause

Plan ID: ADR3.3 | Date: 2026-08-28 | Status: accepted

## Decision

Set size, position, size; read back; on mismatch re-apply immediately, then
once more after 25 ms; record the final frame.

## Why

macOS clamps the size to the source display until the position lands, so a
cross-display move needs the same three attempts Rectangle uses.
