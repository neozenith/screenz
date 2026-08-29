# ADR 0013: Use stdlib flag with a hand-rolled subcommand table

Plan ID: ADR4.3 | Date: 2026-08-28 | Status: accepted

## Decision

No CLI framework. cmd/screenz dispatches on the first positional to
handlers in internal/cli. Exit codes: 0 success, 1 runtime error, 2 usage
error. Help goes to stdout on request and to stderr on a usage mistake.

## Why

The house rule mandates stdlib parsing across languages, and the pattern
covers everything this tool needs.
