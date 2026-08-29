# ADR 0006: Report windows the tool cannot act on instead of hiding them

Plan ID: ADR2.2 | Date: 2026-08-28 | Status: accepted

## Decision

Minimized, hidden, sheet, dialog, other-Space (offscreen) and
unreadable-frame (unknown) windows appear in `status` with a state field.
`apply` lists them as skipped with the reason, and refuses to run when an
application could not be enumerated at all.

## Why

Silently dropping them makes "moved 9 of 10" look like success. Spaces have
no public API, so the offscreen skip is permanent and must be visible.
