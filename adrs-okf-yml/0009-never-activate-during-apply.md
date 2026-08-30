---
type: Architecture Decision
title: Never activate applications during a bulk apply
description: Bulk operations never fight the user for focus
tags: [placement, focus]
status: accepted
accepted_on: 2026-08-28
plan_id: ADR3.2
provenance: "Rectangle prior art: its cross-display move force-activates the app afterwards"
enforced_in:
  - internal/place (Place never calls activate)
generated: { by: human:neozenith, at: 2026-08-28T00:00:00Z }
---

> **Lens**: A bulk operation must not fight the user for focus; place frames through AX and leave the frontmost app alone.

## Problem

### Symptom

Rectangle activates an app after a cross-display move, which is fine for one focused window.

### Pain point

Across a 16-window context switch, per-window activation would strobe focus, reorder the window stack, and interrupt whatever the user is doing.

## Decision

### The lens

- **Given**: applies touch many windows across many apps in one run
- **We prefer**: setting frames through AX without any activation, over Rectangle's activate-after-move
- **Because**: the frontmost app must be unchanged after the run
- **Unless**: a single-window interactive mode ever ships, where activation may be the point

### In practice

- place.Place performs the set-size, set-position, set-size recipe with no NSRunningApplication activation anywhere.

## Consequences

### Pros

- Applies are silent; focus never moves.

### Cons

- None observed; cross-display placement works without activation.
