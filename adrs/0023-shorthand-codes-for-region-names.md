---
type: Architecture Decision
title: Every region name has a shorthand code; thirds use digits
description: Letters for halves and corners, digits for thirds, so no code is a transposition of another
tags: [cli, rules, grammar]
status: accepted
accepted_on: 2026-08-31
provenance: A drafted comparison of two schemes, chosen on which one makes a mistyped code fail rather than silently place a window somewhere else
enforced_in:
  - internal/layout (regionCodes, canonicalRegion, and the test that pairs every code with a region)
generated: { by: human:neozenith, at: 2026-08-31T00:00:00Z }
---

> **Lens**: A code is a second way to type a name, never a second name.
> Codes resolve at the parser door like any other alias, and the catalogue name is what help, errors and profiles say.

## Relates to

- Extends [ADR-0022](0022-region-spellings-canonicalise-on-parse.md) (reuses that record's canonicalise-at-the-door rule for a second class of alias)

## Problem

### Symptom

A rule spells out left-half or first-two-thirds in full, which is the longest token on a line that already carries a selector and a display.

### Pain point

A purely mechanical scheme of word initials puts lt (last-third) one transposition from tl (top-left) and ftt one character from ltt, so a slip parses cleanly and places a window in the wrong part of the screen.

## Decision

### The lens

- **Given**: fourteen named regions in three families — one whole-frame, halves and corners that split on position, thirds that split on ordinal
- **We prefer**: letters for the halves and corners with a digit suffix for the thirds (f3, c3, l3, f23, l23), over the initial-of-each-word scheme that would give ft, ct, lt, ftt, ltt
- **Because**: the digit puts the thirds in a visually separate family, so no code is a transposition or a one-character slip away from a different region
- **Unless**: a region joins the catalogue whose code would collide, which the record must then resolve before the region ships

### In practice

- Every named region has exactly one code, enforced by a test that walks the catalogue and the code table in both directions.
- max is the one code that is not built from initials; maximize is one word and m alone reads as nothing in particular.
- Codes are dialect-free — c3 is the centre third however it is spelled — so they need no entry in the spelling alias table.
- A code never reaches disk; --region l3 saves as region last-third.

## Consequences

### Pros

- The common rules shorten to -r lh and -r max, which is what makes a three-rule inline apply fit on one line.
- A mistyped code is an unknown region rather than a different region, because no code is one slip from another.

### Cons

- Twenty-eight accepted spellings now reach one catalogue of fourteen, so the help text has to carry a table it did not need before.
- max breaks the scheme's own shape, which is the price of not calling the commonest region m.
