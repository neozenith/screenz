---
type: Architecture Decision
title: Profiles are authored from flags or the template, not captured
description: Automate the repeated task, not the rare one
tags: [profiles, scope]
status: accepted
accepted_on: 2026-08-28
plan_id: ADR5.3
provenance: "Planning-session scope decision: the tool replaces the context switch, not profile authoring"
enforced_in:
  - internal/cli profile (init, save; no --from-current flag)
generated: { by: human:neozenith, at: 2026-08-28T00:00:00Z }
---

> **Lens**: Automate the repeated task, not the rare one; capture features earn their place only when authoring cost dominates.

## Problem

### Symptom

A capture-from-current command looks convenient: snapshot today's layout into a profile.

### Pain point

Captured selectors over-match (the title regex that splits browser profiles cannot be inferred), and captured regions describe today's frames rather than intent, so every capture needs hand-editing anyway.

## Decision

### The lens

- **Given**: profiles are authored a handful of times ever, while the context switch repeats daily
- **We prefer**: profile init (a commented template) and profile save from rule flags, over a --from-current capture
- **Because**: a capture mapper is surface without payoff and would force a snap-or-fail policy that contradicts strict verification
- **Unless**: the number of contexts or machines grows enough that authoring cost dominates

### In practice

- Authoring flows are init, hand-edit, and save; there is no capture path to maintain.

## Consequences

### Pros

- Less surface; profiles encode intent, not accidents of today's layout.

### Cons

- New contexts take a few minutes of hand-authoring.
