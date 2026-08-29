# ADR 0008: Fail strictly on a mismatched frame, with a numeric tolerance

Plan ID: ADR3.1 | Date: 2026-08-28 | Status: accepted

## Decision

After read back, every edge of the actual frame must be within the rule's
tolerance of the requested frame; otherwise the result is "clamped" and the
run exits 1. Tolerance is per rule: a bare number is points, a % suffix is
percent of the target size per axis. The default 0.5 pt only absorbs AX
whole-point rounding. Infinite and NaN tolerances are rejected: no setting
switches verification off.

## Why

AX set-frame is a proposal that apps clamp silently; a silently clamped
window is the canonical false success this tool exists to prevent. A
numeric tolerance widens the check deliberately for a known fixed-size app
without an all-or-nothing boolean.

## Rejected

- Always strict: one minimum-sized Finder window keeps a profile red forever.
- "Right display is enough": a window that refused to resize reads as success.
- Boolean tolerant flag: all or nothing.
