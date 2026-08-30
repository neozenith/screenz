---
type: Architecture Decision
title: Fail loudly when Accessibility is not granted
description: Fail with the responsible party named; never report an empty world as success
tags: [accessibility, errors]
status: accepted
accepted_on: 2026-08-28
plan_id: ADR1.2
provenance: "TCC research during planning: grants attach to the terminal host app, which users do not expect"
enforced_in:
  - internal/cli (requireTrusted, grantInstruction, doctor)
  - internal/mac (Missing)
generated: { by: human:neozenith, at: 2026-08-28T00:00:00Z }
---

> **Lens**: An environment that cannot do the job fails with the responsible party named; it never reports an empty world as success.

## Relates to

- Extended by [ADR-0017](0017-terminal-app-is-the-tcc-client.md) (who the TCC client is)

## Problem

### Symptom

Without the grant, AX enumeration returns nothing and an unguarded run reports "moved 0 windows" with exit 0.

### Pain point

That is a silent false success, and the fix is invisible: the grant belongs to the terminal app, not to screenz, so users looking for "screenz" in System Settings find nothing.

## Decision

### The lens

- **Given**: TCC attributes a shell-launched tool to its terminal host app
- **We prefer**: checking the grant first in every AX command and exiting 1 with the host app named, over proceeding and reporting whatever AX returns
- **Because**: a run against an ungranted or unbindable machine can only produce misleading output
- **Unless**: never; this one is unconditional

### In practice

- requireTrusted gates status and apply; missing dlopen symbols get their own named error, not the grant instruction.
- doctor asks with the system prompt option and prints the exact remediation, including `tccutil reset`.

## Consequences

### Pros

- No silent empty runs; the error names the fix.

### Cons

- Commands need the injected trust check even in pure tests.
