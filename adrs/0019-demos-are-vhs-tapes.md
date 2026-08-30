---
type: Architecture Decision
title: Animated demos are VHS tapes rendered through demo mode
description: A demo must be regenerable by script from checked-in sources; a human-at-the-desk recording is a screenshot with extra steps
tags: [demo, documentation, ci]
status: accepted
accepted_on: 2026-08-30
provenance: Bakeoff of VHS against asciinema + agg (docs/spikes, removed after reconciliation; both variants live in git history before this commit)
enforced_in:
  - docs/demos (tapes + GIFs + world file); README embeds the GIFs
generated: { by: human:neozenith, at: 2026-08-30T00:00:00Z }
---

> **Lens**: A demo is documentation, and documentation drifts: it must be regenerable by script from checked-in sources.
> A recording that needs a human at the right desk is a screenshot with extra steps.

## Relates to

- Depends on [ADR-0018](0018-demo-mode-replays-a-recorded-world.md) (renders through demo mode)

## Problem

### Symptom

The README needed animated demos, and two toolchains could produce them: VHS (scripted tapes) and asciinema + agg (recorded casts).

### Pain point

Recorded casts are point-in-time captures: re-recording needs the same machine state, and agg cannot draw window chrome.
The choice had to be settled once so demos would not fork by tool.

## Decision

### The lens

- **Given**: demos must stay true to a moving CLI, and the repo already treats docs drift as a continuous process
- **We prefer**: VHS tapes checked in beside their GIFs under docs/demos, rendered through `SCREENZ_DEMO`
- **Because**: the tape is the demo (deterministic re-render, CI-capable via vhs-action), and `Set WindowBar` draws the title bar agg cannot
- **Unless**: demos need terminal replay or a hosted player; then an asciinema cast can be added alongside, not instead

### In practice

- Three tapes cover the narrative: trust + status, verified apply, profile round-trip.
- Tapes export `SCREENZ_DEMO=demo-world.json`, so rendering needs no displays and no window moves.

## Consequences

### Pros

- Demos regenerate with one command on any machine; demo changes are reviewable text diffs (tape + world file).

### Cons

- GIF binaries live in git and are re-committed on every re-render; VHS sizes in pixels, so column budgets need width arithmetic per tape.
