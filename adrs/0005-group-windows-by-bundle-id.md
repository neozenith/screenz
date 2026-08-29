# ADR-0005: Group windows by bundle id, not PID

| Field | Value |
|---|---|
| **Status** | Accepted, 2026-08-28 |
| **Plan ID** | ADR2.1 |
| **Provenance** | Rectangle prior-art reading: its tile-active-app groups by PID and splits multi-process apps |
| **Relates to** | - |
| **Enforced in** | internal/discover (GroupByBundle, Window.Bundle), internal/rule selectors, profiles |

> **Lens**: Group and address windows by product identity (the bundle id), never by process identity.

## Problem

### Symptom

Chromium and Electron apps run many processes; PID grouping splits one app's windows into several groups or merges helpers into the wrong one.

### Pain point

A rule that says "all VS Code windows" cannot be expressed reliably against PIDs, and profiles would encode unstable numbers.

## Decision

### The lens

- **Given**: multi-process apps and profiles that must name applications durably
- **We prefer**: NSRunningApplication.bundleIdentifier resolved from the window's PID as the group key, over the PID itself
- **Because**: the bundle id is the stable product identity a profile can name, such as com.microsoft.VSCode
- **Unless**: never; this one is unconditional

### In practice

- Discovery resolves every window's PID to a bundle id; the app name is a secondary display label.

## Consequences

### Pros

- Rules and profiles address apps the way users think of them.

### Cons

- A rare bundle-less process groups under an empty key.
