# ADR 0016: Profiles are authored from flags or the template, not captured

Plan ID: ADR5.3 | Date: 2026-08-28 | Status: accepted

## Decision

There is no --from-current capture. Profiles are authored with
`profile init` (a commented template) and `profile save` from rule flags.

## Why

The tool replaces the repeated context switch, not the three-times-ever
profile authoring. Captured selectors over-match (a title regex that splits
browser profiles cannot be inferred) and captured regions describe today's
frames rather than intent.
