---
type: Architecture Decision
title: status elides titles and takes a section
description: A table that holds its shape by default; the whole truth on request, and always in JSON
tags: [cli, discovery]
status: accepted
accepted_on: 2026-09-01
provenance: Real window titles on this machine run past 60 characters (an IDE window is "path (Working Tree) (file) — repo"), pushing every column after TITLE off the screen
enforced_in:
  - internal/cli (elideTitle, the section switch in runStatus)
generated: { by: human:neozenith, at: 2026-09-03T00:00:00Z }
---

> **Lens**: A human-facing table is allowed to shorten a field to stay readable; a machine-facing one never is.
> Elision is a display concern, so it lives in the renderer and is reversible with a flag — the underlying value is never truncated.

## Relates to

- See also [ADR-0006](0006-report-unactionable-windows.md) (elision shortens how a window is shown, never whether it is shown)

## Problem

### Symptom

One long window title wraps the status table, and a wrapped table loses the column alignment that makes it scannable — the display section below it becomes unreadable too.

### Pain point

Both halves of status are always printed, so reading about displays meant scrolling past every window, and the fix for either could not be a flag that hid information.

## Decision

### The lens

- **Given**: status serves two audiences — a human scanning a terminal, and a script reading --json
- **We prefer**: eliding the title to 8 runes either side of an ellipsis for the table only, with --verbose to restore it, over wrapping, over a width-detecting layout, or over dropping the column
- **Because**: a fixed elision is deterministic in a pipe and a recording, and confining it to the table keeps --json a faithful contract
- **Unless**: a field other than the title starts overflowing, which is a layout problem rather than a per-field one

### In practice

- --json is never elided, with or without --verbose; a machine reading it gets the real title.
- The cut is on rune boundaries, since titles carry em dashes and non-Latin scripts.
- A title no longer than the elided form is left alone — replacing 19 characters with 19 characters loses the ends and gains nothing.
- The apps and displays sections narrow the report, and narrow the JSON the same way, so a consumer asking for one section is not handed the other's rows. The other key stays present with a null value; the schema is one shape whichever section was asked for.
- An enumeration failure is not one of those rows. It states how much of the world was readable, so it survives every narrowing and appears under every section.
- An application that failed to enumerate is still reported even when only displays were asked for (ADR2.2).
- A bare word that is not a section is a usage error, never ignored — ignoring it would print the full report and look like the section was honoured.

## Consequences

### Pros

- The table survives a real window set, which is the only set it is ever used on.
- Reading about displays is one word, not a scroll.

### Cons

- Two titles that differ only in their middle now look identical in the table, and --verbose is the only way to tell them apart.
- Section is a positional on a CLI that otherwise names things with flags, which is a wrinkle justified only by how often it is typed.
