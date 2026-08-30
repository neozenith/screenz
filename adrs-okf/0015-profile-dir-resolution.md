---
type: Architecture Decision
title: "Resolve the profile directory dotfiles-first"
description: "Config is dotfiles-first: env override, XDG, ~/.config; never a path with a space"
tags: [profiles, config]
status: accepted
accepted_on: 2026-08-28
plan_id: ADR5.2
provenance: "The user versions tool config in a dotfiles repo; darwin's UserConfigDir ignores XDG"
enforced_in:
  - "internal/profile (Dir, Path)"
  - "doctor's profile dir line"
generated: { by: human:neozenith, at: 2026-08-28T00:00:00Z }
---

> **Lens**: Config paths are dotfiles-first: an explicit env override, then XDG, then ~/.config;
> never a path with a space.

## Problem

### Symptom

Go's os.UserConfigDir on darwin ignores XDG_CONFIG_HOME and returns ~/Library/Application Support, a path containing a space.

### Pain point

That path is awkward to version in a dotfiles repo and unoverridable in tests and client-site setups.

## Decision

### The lens

- **Given**: the user syncs tool config through dotfiles and needs a cheap override
- **We prefer**: $SCREENZ_HOME, then $XDG_CONFIG_HOME/screenz, then ~/.config/screenz, over os.UserConfigDir
- **Because**: the XDG path is syncable and space-free, and the env override keeps a wrong default cheap everywhere
- **Unless**: a GUI or daemon mode ships, where Application Support conventions would apply

### In practice

- Profiles live at profiles/<name>.yaml under the resolved directory; doctor prints the resolution.

## Consequences

### Pros

- Dotfiles-friendly, trivially overridable in tests.

### Cons

- Diverges from the platform's default app-config location, deliberately.
