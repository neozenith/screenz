---
type: Architecture Decision
title: Use goccy/go-yaml so profile comments survive saves
description: Hand comments are user data; every write preserves them and survives interruption
tags: [profiles, yaml, dependencies]
status: accepted
accepted_on: 2026-08-28
plan_id: ADR5.1
provenance: The user asked for commented profiles; library research found yaml.v3 archived and comment-dropping
enforced_in:
  - internal/profile (strict load, comment map, whole-rule-list Replace, atomic writes)
generated: { by: human:neozenith, at: 2026-09-03T00:00:00Z }
---

> **Lens**: A hand-commented file is user data; every write must preserve the comments and survive interruption.

## Relates to

- Depends on [ADR-0011](0011-match-opens-a-rule.md) (serialises the grammar)
- Extended by [ADR-0025](0025-three-verbs-profiles-named-by-flag.md) (turned the save from an append into a whole-rule-list replace, which narrowed what survives)

## Problem

### Symptom

Profiles are commented specs, and yaml.v3 drops comments on struct marshal while being archived upstream.

### Pain point

A save that strips the user's comments, or a crash that truncates the file, destroys hand-authored content that cannot be regenerated.

## Decision

### The lens

- **Given**: comments must round-trip and the library landscape offers goccy's comment map as the maintained option
- **We prefer**: goccy/go-yaml with Strict decoding and CommentToMap/WithComment, over yaml.v3 or a release-candidate v4
- **Because**: unknown keys error instead of vanishing, and comments survive the round trip
- **Unless**: a maintained yaml v4 stabilises with equivalent comment support, which would prompt a re-evaluation

### In practice

- A save rewrites the whole rules list in place, block style only, written via temp-file rename (ADR-0025).
- Comment-map keys are positional, so a comment on a rule that no longer exists would reattach to whatever now sits at that index. Every per-rule comment is dropped on a save; every comment outside the rules list survives.
- The comment on the rules key itself survives, because that key carries no index and documents the section rather than any one rule.
- Blank lines do not survive a save and flow-style comments are lost; the template avoids both.

## Consequences

### Pros

- The header, the displays alias map and their comments survive every save, which is where hand-authored context actually lives.

### Cons

- One third-party dependency, pinned and exercised by round-trip tests.
- A comment written against an individual rule is lost the next time that profile is saved, and only the positional-key hazard makes that defensible.
