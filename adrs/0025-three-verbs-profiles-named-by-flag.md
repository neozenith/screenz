---
type: Architecture Decision
title: Three verbs, and a profile is named by flag
description: A command is what you are doing; a profile is which one, and that is a flag
tags: [cli, profiles, grammar]
status: accepted
accepted_on: 2026-09-01
provenance: Using the released v0.4.0 grammar in anger — `profile status` and `profile save` read as managing a noun, and the positional profile name meant two ways to say the same thing
enforced_in:
  - internal/cli (Run dispatch; apply --profile/--save-profile, init --profile, list)
  - internal/profile (Replace, the in-place rule swap --save-profile writes through)
generated: { by: human:neozenith, at: 2026-09-01T00:00:00Z }
---

> **Lens**: A subcommand names an action, never a noun to manage: apply, status, list, init.
> Which profile an action concerns is a flag, so the same name is read the same way everywhere and a bare word is never a silent guess.

## Relates to

- See also [ADR-0016](0016-no-capture-from-current.md) (its no-capture rule stands unchanged; only the commands its authoring flows live in have moved)
- Extends [ADR-0024](0024-commands-answer-to-their-initial.md) (changes which commands exist; the initial-as-short-name rule is unchanged and still holds)

## Problem

### Symptom

Profile work was split across `profile status`, `profile init` and `profile save`, while applying one was a bare positional on apply — four spellings of the same noun across two nesting levels.

### Pain point

A nested noun-command has to be learned as a place before anything can be done in it, and a positional name meant `apply office` and `apply --match ...` were different grammars that a typo could slide between.

## Decision

### The lens

- **Given**: every profile operation is really an action on the rule set — run it, save it, start one, see which ones fit
- **We prefer**: folding them into apply, init and list with the profile named by --profile or --save-profile, over a profile noun-command with subcommands
- **Because**: a flag-named profile reads identically in every command, and dropping the positional means a bare word is always an error rather than sometimes a profile
- **Unless**: an operation appears that is about the profile file itself rather than its rules, which would earn its own verb

### In practice

- apply --profile NAME runs it; apply FLAGS --save-profile NAME keeps what was just typed; init --profile NAME starts one from the template; list shows which fit.
- --save-profile replaces the profile's rules in place and leaves the rest of the file alone, so comments and the displays map survive (ADR-0014).
- What is saved is the unresolved rule set — an alias stays an alias, because saving the resolved form would bake today's display indexes into the file and destroy the indirection the profile exists for.
- The save runs before placement, so a run that fails to place still keeps the rules that were asked for.
- A bare word where a profile name used to go — apply office, init office — is a usage error naming the flag to use; neither guesses.
- list rolls one verdict per profile, with the reasons behind --verbose; a profile declaring no aliases fits trivially.

## Consequences

### Pros

- Four commands where there were six, and one grammar for naming a profile.
- A profile is now savable straight from the invocation that proved it, which was two commands and a retyped rule set before.

### Cons

- A breaking change one release after the flags were settled; every v0.4.0 invocation using `profile` or a positional name must be retyped.
- --save-profile replaces rather than appends, so building a profile up rule by rule now means passing the whole set each time.
