# ADR-0007: Order display indexes by row, then left to right

| Field | Value |
|---|---|
| **Status** | Accepted, 2026-08-28 |
| **Plan ID** | ADR2.3 |
| **Provenance** | Rectangle's default display ordering, which the user already lives with |
| **Relates to** | - |
| **Enforced in** | internal/discover (BuildDisplays), internal/rule (bare-integer index shorthand) |

> **Lens**: The human-friendly display number follows the ordering users already know; stable identity
> always comes from UUID or name, never the index.

## Problem

### Symptom

Displays need a short human handle, but CG display ids are small unstable ints and identical panels can
share a serial number.

### Pain point

Without a declared ordering, "display 2" means different things after any layout change, and profiles
written against numbers break silently.

## Decision

### The lens

- **Given**: the user's muscle memory comes from Rectangle's ordering, and identical panels are only
  distinguishable by UUID or the suffixed localized name
- **We prefer**: index numbering top-to-bottom by row then left-to-right, with a bare integer meaning
  index=N, over arbitrary or CG-id ordering
- **Because**: it matches the ordering the user already knows, while UUID and name stay the durable keys
- **Unless**: the reference convention changes; profiles that must survive layout changes already
  address panels by name or UUID instead

### In practice

- Purely numeric alias names are reserved so the integer shorthand stays unambiguous.
- Zero-match display specs explain per term what each term would have matched.

## Consequences

### Pros

- Quick one-digit tweaks for ad-hoc use; durable names for portable profiles.

### Cons

- Indexes shift when displays connect or move; that is documented, not hidden.
