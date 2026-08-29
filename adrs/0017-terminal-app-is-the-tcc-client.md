# ADR-0017: The terminal app is the TCC client; no daemon mode

| Field | Value |
|---|---|
| **Status** | Accepted, 2026-08-28 (split from ADR-0001 on 2026-08-29) |
| **Plan ID** | ADR6.2 |
| **Provenance** | Distribution research: the Go linker ad-hoc signs darwin output, and TCC keys a self-responsible binary on its CDHash |
| **Relates to** | Split from ADR-0001; extends ADR-0003 |
| **Enforced in** | internal/mac (HostApp), internal/cli (grantInstruction, doctor), docs/install.md |

> **Lens**: Keep screenz shell-launched so the Accessibility grant belongs to the terminal app and survives every rebuild; a self-responsible daemon needs a Developer ID certificate first.

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
