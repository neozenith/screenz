---
type: Reference
title: "How these records are written"
description: Template shape, immutability policy, and the two numbering axes
tags: [meta, conventions]
status: accepted
generated: { by: human:neozenith, at: 2026-08-30T00:00:00Z }
---

> **Lens**: Read a Lens to apply a decision; open the record only for its argument.

## Template shape

Every record carries its metadata as OKF frontmatter, then a Lens blockquote (the reusable rule, first), then Problem, Decision (Given / We prefer / Because / Unless), and Consequences.

The Lens is duplicated as the frontmatter `description`, so an agent reading only the index or only the frontmatter still gets the rule without opening the body.

## Immutability

Decisions are immutable in substance.
A change of mind is a new record that supersedes the old one with links both ways, and supersession may retire a single clause (recorded in the earlier record's `status` and its Relates to section, never by rewriting the clause).

## Two numbering axes

Two numbering axes exist by history.
Files are `NNNN` ordered, and code comments cite the planning-era IDs (ADR1.1 through ADR6.2) carried in each record's `plan_id` frontmatter field.

## OKF extensions used here

The specification lets producers add custom frontmatter keys, and requires consumers to preserve unknown keys rather than reject them.
This bundle uses four:

| Key | Carries |
|---|---|
| `plan_id` | The planning-era identifier cited in code comments |
| `provenance` | The evidence the decision rests on, in prose |
| `enforced_in` | The packages and files where the decision is actually binding |
| `accepted_on` | The acceptance date, kept distinct from `generated.at` (last significant change) |
