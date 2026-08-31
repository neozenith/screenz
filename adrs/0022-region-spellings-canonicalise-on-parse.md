---
type: Architecture Decision
title: Accept en-GB region spellings, canonicalise them on parse
description: Type it either way, store it one way; an alias never reaches disk
tags: [cli, rules, grammar, profiles]
status: accepted
accepted_on: 2026-08-31
provenance: An Australian author typing maximise against a catalogue written in American English, on a tool whose only region-name error is "unknown region"
enforced_in:
  - internal/layout (regionAliases, resolved by canonicalRegion at the top of ParseRegion)
generated: { by: human:neozenith, at: 2026-08-31T00:00:00Z }
---

> **Lens**: A vocabulary word with a second correct spelling accepts both at the door and stores one.
> Aliases live in the parser, never in the catalogue, so nothing downstream learns there are two names for one thing.

## Relates to

- See also [ADR-0020](0020-omitted-region-means-maximize.md) (the defaulted region is stored in canonical spelling like any other)
- Extended by [ADR-0023](0023-shorthand-codes-for-region-names.md) (shorthand codes reuse this canonicalise-at-the-door rule for a second class of alias)

## Problem

### Symptom

The catalogue spells maximize and center-third, so maximise and centre-third are rejected as unknown regions by a user who spelled them correctly for their own dialect.

### Pain point

Adding the variants to the catalogue itself would double the vocabulary — two names for one region in help text, in error listings and in saved profiles, with no way to tell which is the real one.

## Decision

### The lens

- **Given**: the region catalogue is a closed set of words, two of which have an en-GB/en-AU variant, and every region already round-trips through String into profile YAML
- **We prefer**: a parser-level alias table that rewrites the input before lookup, over adding variant keys to the catalogue or normalising every s/z in sight
- **Because**: canonicalising at the single door keeps one spelling in help, in errors and on disk, while costing the user nothing at the keyboard
- **Unless**: a variant is not a spelling of the same word but a different region, which the catalogue must then name in full

### In practice

- maximise resolves to maximize and centre-third to center-third; those are the only two catalogue words with a variant.
- The alias is not preserved — a profile saved from --region maximise stores region maximize, which is semantically lossless but not byte-for-byte.
- Unknown input keeps its original text in the error, since canonicalisation only happens on a hit.

## Consequences

### Pros

- The tool reads as correctly spelled to both dialects without either spelling becoming a second official name.
- One door — profile YAML, inline flags and the demo transcripts all validate through the same ParseRegion.

### Cons

- A user who types maximise sees maximize come back in their saved profile, which is a small surprise the help text has to carry.
- Every future catalogue word now needs the variant question asked at the time it is added.
