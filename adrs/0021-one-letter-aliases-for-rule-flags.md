---
type: Architecture Decision
title: One-letter aliases for the rule grammar; destructive flags stay long
description: A letter for what you type per rule, the full word for what you cannot undo
tags: [cli, rules, grammar]
status: accepted
accepted_on: 2026-08-31
provenance: A three-rule context switch spells seven long flags; the same line in short form fits on one screen, and stdlib flag aliases cost one extra Var call per flag
enforced_in:
  - internal/rule (List.Register binds each rule flag under both names)
  - internal/cli (aliasBool for --json and --dry-run; status binds --match twice)
generated: { by: human:neozenith, at: 2026-08-31T00:00:00Z }
---

> **Lens**: Flags typed once per rule earn a one-letter alias; flags that overwrite a file or replace the binary are typed in full.
> The long name stays canonical everywhere a flag is named back to the user or written to disk.

## Relates to

- Extends [ADR-0011](0011-match-opens-a-rule.md) (adds a second spelling of each rule flag without changing what a rule is)
- See also [ADR-0013](0013-stdlib-flag-cli.md) (stdlib flag is why aliases are a second registration and why -mdr grouping does not exist)
- See also [ADR-0024](0024-commands-answer-to-their-initial.md) (the same one-letter treatment, applied to the command word rather than to its flags)

## Problem

### Symptom

Every rule in an apply line spells out --match, --display and --region, so a three-rule context switch runs past the width of a terminal before it says anything.

### Pain point

The flags typed most often were the longest, which pushed real use toward profiles even for one-off placements that profiles are not meant to absorb.

## Decision

### The lens

- **Given**: stdlib flag treats one and two dashes alike and lets one Value carry several names, but has no clustering and no separate short-flag namespace
- **We prefer**: a curated one-letter alias per rule flag (-m -d -r -g -t -f -o) plus -j and -n, over aliasing every flag or none
- **Because**: the alias set is small enough to hold in the head, and leaving --force and --check long makes the irreversible commands read deliberately
- **Unless**: a new rule flag has no free letter, in which case it ships long-only rather than displacing an existing letter

### In practice

- -f is --first, not --force; profile save registers both, and only one of them may own the letter.
- Errors, help text and profile YAML keys always name the long form, so a message never teaches a spelling that is not the canonical one.
- -h needs no registration; stdlib flag returns ErrHelp for an undefined -h and every subcommand already maps that to its help text.
- Clustering (-mdr) does not exist under stdlib flag and is not wanted here — every rule flag except --first takes a value, so a cluster could hold at most one of them and would only obscure which value bound to which flag.

## Consequences

### Pros

- A whole context switch fits on one line, which is what makes an inline apply competitive with a profile.
- Aliases are the same flag object, so a short and a long spelling can never disagree about a value.

### Cons

- Every future flag now carries a letter question, and the letters are nearly exhausted.
- Two spellings means a shell history or a script may read in either, so the alias table has to be known to read someone else's command line.
