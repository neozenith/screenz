# ADR 0002: Bridge to macOS through purego with CGO_ENABLED=0

Plan ID: ADR1.1 | Date: 2026-08-28 | Status: accepted

## Decision

`internal/mac` binds the Accessibility, CoreGraphics, ColorSync and AppKit
symbols with `github.com/ebitengine/purego` (Dlopen + RegisterLibFunc,
objc.Send for NSScreen and NSRunningApplication). The build stays
`CGO_ENABLED=0`.

## Why

A spike on the target machine (2026-08-22, Go 1.27, purego v0.10.2) proved
every needed primitive: struct-returning CGDisplayBounds, objc.Send[CGRect]
exact on arm64, and AX enumerate plus set-frame with read back. It keeps the
house CGO-free convention and needs no Xcode CLT on client laptops.

## Rejected

- cgo: needs the command line tools on every build machine; kept only as a
  documented fallback if a future symbol cannot be bound.
- progrium/darwinkit: cgo based and dormant.
