---
type: Architecture Decision
title: Demo mode replays a recorded world and simulates placement
description: "Fabricated output is a documentation prop: env-gated, cmd-wired, doctor-disclosed, never in tests"
tags: [demo, documentation, testing]
status: accepted
accepted_on: 2026-08-30
provenance: "Demo curation session: the animated-README bakeoff could not render multi-display demos on an undocked laptop, and the maintainer chose fabricated apply output over hardware-gated docs"
enforced_in:
  - internal/demo; cmd/screenz wiring (env-gated); internal/cli doctor (disclosure)
generated: { by: human:neozenith, at: 2026-08-30T00:00:00Z }
---

> **Lens**: Fabricated output is a documentation prop, never a product path: gate it on an explicit env var, wire it only in cmd, disclose it in doctor, and keep it out of every test.

## Relates to

- Scoped exception to [ADR-0008](0008-strict-tolerance-verification.md) (demo mode simulates read-back; the rule stands everywhere else)
- Depends on [ADR-0004](0004-two-test-tiers.md) (replay rides the injected Deps seams)

## Problem

### Symptom

Curating demonstration text for the README requires the demoed hardware: undocked, `apply` aborts at display resolution, and the verified table cannot exist without real windows landing.

### Pain point

Demo GIFs could only be re-rendered when docked, so documentation freshness was hostage to physical setup.

## Decision

### The lens

- **Given**: the CLI is a pure pipeline over injected `Deps`, and `status --json` already serialises a complete world
- **We prefer**: `SCREENZ_DEMO=<world.json>` replaying a recorded world through the real pipeline, with placement simulated as read-back-equals-target
- **Because**: every demo line except the ACTUAL column is then genuinely computed, and demos render anywhere
- **Unless**: the fabricated output starts being mistaken for verification evidence; then demo output must carry an inline marker, not just the doctor disclosure

### In practice

- `internal/demo` owns `Load` (the `status --json` schema) and the simulated `Place`; `cmd/screenz` wires them only when `SCREENZ_DEMO` is set.
- Doctor always prints `demo mode: replaying <path>; placement simulated` (and `demo_world` in JSON) when the variable is set.
- The read-back rule of ADR-0008 stands everywhere else: tests never wire demo mode, and `docs/evidence/` transcripts must never be captured under it.

## Consequences

### Pros

- Demos regenerate on any machine, including CI without displays; demo worlds are curated fixtures with clean titles.

### Cons

- An apply table in a demo GIF is staged output; the honesty of the README now rests on this ADR's discipline rather than on the run itself.
