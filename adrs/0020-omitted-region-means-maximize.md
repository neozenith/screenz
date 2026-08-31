---
type: Architecture Decision
title: An omitted region means maximize
description: Where is mandatory, how big is defaulted; the common rule is the short one
tags: [cli, rules, grammar, profiles]
status: accepted
accepted_on: 2026-08-31
provenance: "A live apply of three rules aborted on \"rule 1 (app=Code): missing --region\" when every rule wanted the whole screen anyway"
enforced_in:
  - internal/layout (DefaultRegion)
  - internal/rule (matchFlag.Set seeds it, List.Validate no longer demands it)
  - internal/profile (toRule treats an absent region key as the default)
generated: { by: human:neozenith, at: 2026-08-31T00:00:00Z }
---

> **Lens**: A rule must always say which display; it need not say how big.
> Omitting --region means maximize, so the commonest rule is also the shortest one to type.

## Relates to

- Extends [ADR-0011](0011-match-opens-a-rule.md) (narrows that record's completeness clause — display stays mandatory, region no longer is)
- See also [ADR-0022](0022-region-spellings-canonicalise-on-parse.md) (the default is stored in the catalogue spelling, as an explicitly typed region is)

## Problem

### Symptom

Every rule had to spell --region even when the intent was the whole display, so a three-rule context switch failed on the first rule with nothing moved.

### Pain point

The most common region was also the most typed, and the error arrived after the whole command line had been written rather than while writing it.

## Decision

### The lens

- **Given**: a rule already names its display, and maximize is both the commonest region and the only one that needs no further parameters
- **We prefer**: defaulting an omitted region to maximize, over keeping region mandatory or inventing a keep-current-size region
- **Because**: which display a window belongs on cannot be guessed, but how much of that display it should take can be, and a default frame keeps read-back verification exact
- **Unless**: a future region catalogue has no single obvious whole-frame member

### In practice

- The default is a constructed layout.DefaultRegion(), never the zero Region, which spans nothing.
- It is seeded when --match opens the rule, alongside the order and tolerance defaults, so a later --region simply overwrites it.
- The profile YAML grammar defaults identically; the save path still writes the key back explicitly, so a round-tripped profile is never ambiguous.

## Consequences

### Pros

- The whole-display context switch is one flag per rule shorter, and rules read as selector then destination.
- Verification is untouched — a defaulted region is a concrete frame, checked against read-back like any other.

### Cons

- A forgotten --region now maximizes instead of failing, so a typo in a region name is the only remaining loud error; --dry-run is the way to see the frames before they move.
