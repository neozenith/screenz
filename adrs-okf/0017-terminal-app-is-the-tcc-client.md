---
type: Architecture Decision
title: "The terminal app is the TCC client; no daemon mode"
description: "Shell-launched only; the terminal app owns the grant (split from 0001)"
tags: [macos, accessibility, distribution]
status: accepted
accepted_on: 2026-08-28
plan_id: ADR6.2
provenance: "Distribution research: the Go linker ad-hoc signs darwin output, and TCC keys a self-responsible binary on its CDHash"
enforced_in:
  - "internal/mac (HostApp)"
  - "internal/cli (grantInstruction, doctor)"
  - "docs/install.md"
generated: { by: human:neozenith, at: 2026-08-29T00:00:00Z }
---

> **Lens**: Keep screenz shell-launched so the Accessibility grant belongs to the terminal app and survives every rebuild; a self-responsible daemon needs a Developer ID certificate first.

## Relates to

- Split from [ADR-0001](0001-distribute-via-github-releases.md); extends [ADR-0003](0003-fail-loudly-without-grant.md)

## Problem

### Symptom

An ad-hoc signed binary that owns its own TCC grant loses it on every rebuild, because TCC keys it on the binary's CDHash.

### Pain point

Users would re-grant Accessibility after every update, and would not know why, since System Settings shows the app list, not the hash bookkeeping.

## Decision

### The lens

- **Given**: a shell-launched tool inherits its terminal host app's grant, so the Go linker's ad-hoc signature never matters
- **We prefer**: documenting and testing screenz only as a shell-launched tool, over a LaunchAgent or daemon mode
- **Because**: the grant then keys on the terminal app and survives rebuilds and updates
- **Unless**: daemon mode is ever justified and a Developer ID certificate is acquired to give the binary a stable identity

### In practice

- doctor and every error name the responsible terminal app, never the binary.
- The install runbook grants Accessibility per terminal app and documents `tccutil reset`.

## Consequences

### Pros

- Grants survive updates; no signing infrastructure.

### Cons

- No background or hotkey mode; every action is a shell invocation.
