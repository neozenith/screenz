# ADR 0009: Never activate applications during a bulk apply

Plan ID: ADR3.2 | Date: 2026-08-28 | Status: accepted

## Decision

`apply` sets frames through AX without activating any application; the
frontmost app is unchanged after the run.

## Why

Rectangle's post-move activation exists for a single focused window. Across
16 windows it would strobe focus and reorder the window stack.
