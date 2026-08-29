# ADR 0003: Fail loudly when Accessibility is not granted

Plan ID: ADR1.2 | Date: 2026-08-28 | Status: accepted

## Decision

Every command that needs AX checks the grant first and exits 1 with an
instruction naming the terminal host app. `doctor` additionally asks with
the system prompt option. Missing dlopen symbols are recorded and named,
never a panic.

## Why

An unguarded run reports "moved 0 windows" and exits 0: a silent failure.
TCC attributes the grant to the terminal host app, which users do not
expect, so the error must name that app.
