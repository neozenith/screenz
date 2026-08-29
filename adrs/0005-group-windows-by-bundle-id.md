# ADR 0005: Group windows by bundle id, not PID

Plan ID: ADR2.1 | Date: 2026-08-28 | Status: accepted

## Decision

The window group key is NSRunningApplication.bundleIdentifier resolved from
the window's PID; the app name is a secondary display field.

## Why

PID grouping splits multi-process apps (Chromium, Electron). Bundle id is
what a profile names, such as com.microsoft.VSCode.
