# ADR-0013: Use stdlib flag with a hand-rolled subcommand table

| Field | Value |
|---|---|
| **Status** | Accepted, 2026-08-28 |
| **Plan ID** | ADR4.3 |
| **Provenance** | House CLI rules, applied across languages |
| **Relates to** | - |
| **Enforced in** | cmd/screenz, internal/cli (Run dispatch, per-command FlagSets, exit codes) |

> **Lens**: Stdlib parsing with a hand-rolled dispatch table; no CLI framework enters this codebase.

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
