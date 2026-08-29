# ADR 0007: Order display indexes by row, then left to right

Plan ID: ADR2.3 | Date: 2026-08-28 | Status: accepted

## Decision

`index=N` numbers displays top to bottom by row, then by minX within a row.
A bare integer display value means index=N; purely numeric alias names are
reserved. UUID and name remain the stable keys for profiles.

## Why

It matches Rectangle's default ordering that the user already lives with.
Indexes shift when displays change, so profiles that must survive layout
changes address panels by name or UUID instead.
