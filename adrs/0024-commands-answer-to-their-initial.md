---
type: Architecture Decision
title: Every command answers to its initial
description: The first letter is the command's short name, so a new command must claim a free one
tags: [cli, grammar]
status: accepted
accepted_on: 2026-08-31
provenance: The rule flags had just taken one-letter aliases, and the commands they belong to had not; the six existing commands happen to have six distinct initials
enforced_in:
  - internal/cli (Run's dispatch switch, one case per command carrying both spellings)
generated: { by: human:neozenith, at: 2026-08-31T00:00:00Z }
---

> **Lens**: A command's short name is its first letter — not a chosen abbreviation, so there is nothing to remember and nothing to look up.
> That makes the initial part of naming a command: a new command with a taken letter is a naming problem, not an alias problem.

## Relates to

- Extends [ADR-0013](0013-stdlib-flag-cli.md) (adds a second case label per command to the hand-rolled dispatch table, with no parser change)
- See also [ADR-0021](0021-one-letter-aliases-for-rule-flags.md) (the same one-letter treatment, applied to the command word rather than to its flags)

## Problem

### Symptom

The rule flags shortened to -m, -d and -r while the command in front of them stayed a whole word, so the shortest useful invocation still opened with apply.

### Pain point

An abbreviation table for commands would be one more thing to memorise, and a prefix matcher would make screenz s ambiguous the day a second s command lands, silently rebinding a command someone had already learned.

## Decision

### The lens

- **Given**: six commands with six distinct initials, dispatched by a switch on the first positional
- **We prefer**: the initial as a fixed second case label per command, over prefix matching or a chosen-abbreviation table
- **Because**: a fixed label can never become ambiguous later, and prefix matching would change what an existing short form means the moment a sibling command is added
- **Unless**: a command that must exist has no free initial, which is a reason to rename the command rather than to skip its letter

### In practice

- a apply, d doctor, s status, p profile, u update, v version, h help; --version and -h keep working as before.
- The full name is what usage text, error messages and the ADRs say; the letter is only ever an input.
- No prefix matching — st is not status, and an unknown command is still an unknown command.

## Consequences

### Pros

- The whole invocation shortens at once, so screenz a -m app=Code -d 2 is the short form end to end.
- Adding a command now asks the letter question at naming time, when renaming is still cheap.

### Cons

- Seven of twenty-six letters are spent, and the obvious next commands (save, snapshot, swap) all want s.
- A letter is silent about what it does, so a command line in someone else's history reads worse than the long form.
