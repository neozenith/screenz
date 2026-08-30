---
type: Architecture Decision
title: Each --match opens a rule; sibling flags bind to it
description: One quoting layer; the CLI grammar is the saved-profile grammar
tags: [cli, rules, grammar]
status: accepted
accepted_on: 2026-08-28
plan_id: ADR4.1
provenance: The planning session's grammar decision, settled against the documented stdlib flag ordering guarantee
enforced_in:
  - internal/rule (List, matchFlag, field)
  - internal/profile (YAML key mapping)
generated: { by: human:neozenith, at: 2026-08-28T00:00:00Z }
---

> **Lens**: One quoting layer and identical keys across CLI and YAML; the CLI grammar is the saved-profile grammar.

## Relates to

- Depended on by [ADR-0014](0014-goccy-yaml-comments.md) (relies on this for lossless profile saves)

## Problem

### Symptom

Stdlib flag has no native repeated groups, but one invocation must carry many selector, display, region rules.

### Pain point

A nested rule string would need its own parser and two or three quoting layers, and profiles would store opaque strings.

## Decision

### The lens

- **Given**: profile save must serialise flags losslessly, and the flag package documents that Value.Set is called in command-line order
- **We prefer**: repeated --match opening a rule with sibling flags binding to the most recent one, over --rule strings or positional groups
- **Because**: ordering makes the binding deterministic, and identical keys map one-to-one into YAML
- **Unless**: a future selector cannot be expressed in a single flag value

### In practice

- A sibling flag before any --match is a usage error (exit 2); every rule requires a display and a region after parsing.

## Consequences

### Pros

- Lossless save; one grammar to learn.

### Cons

- Long command lines for big contexts, which profiles exist to absorb.
