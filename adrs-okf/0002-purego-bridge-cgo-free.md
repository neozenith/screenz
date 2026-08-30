---
type: Architecture Decision
title: "Bridge to macOS through purego with CGO_ENABLED=0"
description: "Bind OS primitives through purego, CGO off; prove each symbol on real hardware first"
tags: [macos, ffi, build]
status: accepted
accepted_on: 2026-08-28
plan_id: ADR1.1
provenance: "The purego spike on the target machine (2026-08-22, Go 1.27, purego v0.10.2) proved every primitive"
enforced_in:
  - "internal/mac"
  - "Makefile (CGO_ENABLED=0 -trimpath builds)"
generated: { by: human:neozenith, at: 2026-08-28T00:00:00Z }
---

> **Lens**: Bind macOS primitives through purego and keep CGO off; prove each new symbol on real hardware before building anything on it.

## Relates to

- [ADR-0004](0004-two-test-tiers.md) covers how the bridge is tested

## Problem

### Symptom

screenz needs Accessibility, CoreGraphics and AppKit calls from Go, and purego has shipped ABI bugs for exactly the struct-by-value returns those APIs use.

### Pain point

cgo would require the Xcode command line tools on every build machine and client laptop, and would break the house CGO-free build convention.

## Decision

### The lens

- **Given**: the spike proved struct-returning CGDisplayBounds, objc.Send[CGRect] exact on arm64, and AX enumerate plus set-frame with read back, all without a crash or ABI fault
- **We prefer**: purego Dlopen + RegisterLibFunc bindings with objc.Send, over cgo
- **Because**: it keeps CGO_ENABLED=0 and needs no compiler toolchain anywhere
- **Unless**: a future symbol cannot be bound through purego; cgo then returns as the documented fallback for that symbol alone

### In practice

- All OS symbols live in internal/mac behind darwin build tags; missing symbols are recorded, never a panic.
- Every new binding is verified against the real window server before use.

## Consequences

### Pros

- CGO-free cross-compilation from any OS; no client toolchain.

### Cons

- The C ABI risk is owned here; each binding needs empirical proof.
