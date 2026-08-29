# ADR-0009: Never activate applications during a bulk apply

| Field | Value |
|---|---|
| **Status** | Accepted, 2026-08-28 |
| **Plan ID** | ADR3.2 |
| **Provenance** | Rectangle prior art: its cross-display move force-activates the app afterwards |
| **Relates to** | - |
| **Enforced in** | internal/place (Place never calls activate) |

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
