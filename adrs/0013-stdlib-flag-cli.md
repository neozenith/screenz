---
type: Architecture Decision
title: Use stdlib flag with a hand-rolled subcommand table
description: Stdlib parsing, hand-rolled dispatch; no CLI framework
tags: [cli, dependencies]
status: accepted
accepted_on: 2026-08-28
plan_id: ADR4.3
provenance: House CLI rules, applied across languages
enforced_in:
  - cmd/screenz
  - internal/cli (Run dispatch, per-command FlagSets, exit codes)
generated: { by: human:neozenith, at: 2026-09-02T00:00:00Z }
---

> **Lens**: Stdlib parsing with a hand-rolled dispatch table; no CLI framework enters this codebase.

## Relates to

- See also [ADR-0021](0021-one-letter-aliases-for-rule-flags.md) (its alias mechanism and the absence of flag clustering both follow from stdlib flag)
- Extended by [ADR-0024](0024-commands-answer-to-their-initial.md) (gives each command in the dispatch table a second case label, its initial)
- Extended by [ADR-0027](0027-embed-jq-behind-a-jq-flag.md) (settles where the no-framework rule stops - a library behind one flag is not a framework)

## Problem

### Symptom

The CLI needs subcommands, repeated flags and exit-code discipline, which frameworks advertise.

### Pain point

A framework buys convenience with a dependency, its own conventions, and drift from the house pattern used across every other tool.

## Decision

### The lens

- **Given**: the house rule mandates stdlib parsing and the tool's needs are modest
- **We prefer**: stdlib flag plus a switch on the first positional, over cobra or urfave/cli
- **Because**: the pattern covers everything needed and stays dependency-free
- **Unless**: never; this one is unconditional

### In practice

- Exit codes: 0 success, 1 runtime error, 2 usage error; help to stdout on request, errors and usage to stderr.
- Read-only verbs are named status.

## Consequences

### Pros

- No framework dependency; predictable behaviour.

### Cons

- Repeated FlagSet boilerplate per command, kept small by shared helpers.
