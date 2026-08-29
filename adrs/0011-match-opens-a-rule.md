# ADR 0011: Each --match opens a rule; sibling flags bind to it

Plan ID: ADR4.1 | Date: 2026-08-28 | Status: accepted

## Decision

Rules are repeated --match / --display / --region (plus --gap, --tolerance,
--first, --order) flags. Every --match opens a new rule and each sibling
flag binds to the most recent rule, in command-line order.

## Why

`profile save` must serialise flags losslessly, and the stdlib flag package
documents that Value.Set is called in command-line order. One quoting layer
and identical keys make the CLI grammar the saved-profile grammar.

## Rejected

- Repeated --rule 'match=... display=...' strings: a second parser and two
  or three nested quoting layers.
- Positional groups separated by --: fights the flag terminator semantics.
