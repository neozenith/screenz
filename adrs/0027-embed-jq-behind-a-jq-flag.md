---
type: Architecture Decision
title: Embed a jq engine behind --jq rather than shell out
description: The filter travels in the binary, so --json stays useful where jq is not installed
tags: [cli]
status: accepted
accepted_on: 2026-09-02
provenance: Every --json output on this CLI is read by piping it to jq, so the pipe is the actual interface; the machines that most need `screenz status --json` (a fresh laptop, a CI runner, a colleague's shell) are the ones least likely to have jq on PATH
enforced_in:
  - internal/cli (jq.go — jqOpts.register, jqOpts.resolve, emitJSON)
  - internal/cli (apply.go, doctor.go, list.go, status.go — the four commands that emit JSON)
generated: { by: human:neozenith, at: 2026-09-02T00:00:00Z }
---

> **Lens**: A convenience that only works when a second tool is installed is not a feature, it is a shorter way to type a pipe.
> When the CLI absorbs another tool's job, it absorbs that tool's grammar and defaults too — deviating only where this CLI's own grammar already claimed the spelling, and saying so where it cannot comply.

## Relates to

- Extends [ADR-0013](0013-stdlib-flag-cli.md) (a jq engine is a library behind one flag, not the CLI framework ADR-0013 refuses)
- See also [ADR-0021](0021-one-letter-aliases-for-rule-flags.md) (-r is --region, so the raw-string switch keeps only its long name)
- See also [ADR-0002](0002-purego-bridge-cgo-free.md) (gojq is pure Go, so the CGO-free single binary survives the dependency)

## Problem

### Symptom

Reading one field out of a --json report costs a pipe to jq, and on a machine without jq the report has to be read by eye or copied elsewhere to be parsed at all.

### Pain point

The CLI's machine-readable surface is only as portable as the binary it depends on, which defeats the point of shipping a single self-contained executable.

## Decision

### The lens

- **Given**: every command that emits JSON is read through jq in practice, and the tool ships as one CGO-free binary
- **We prefer**: embedding github.com/itchyny/gojq and exposing it as --jq QUERY, over executing a jq binary found on PATH, over inventing a smaller field-selector syntax, and over leaving the pipe as the only option
- **Because**: a query that runs inside the binary works wherever the binary works, and reusing jq's grammar means there is nothing new to learn or document
- **Unless**: a query needs a jq feature gojq does not implement, in which case the unfiltered --json and a real jq pipe are still there and still the contract

### In practice

- --jq implies --json, so neither ever has to be typed with the other.
- The query is compiled before the command reads the world, so a typo costs no snapshot and moves no window.
- A query that will not parse or compile is a usage error (exit 2); one that compiles and then fails on the data is a runtime error (exit 1).
- Output follows jq's defaults - one result per line, two-space indent, strings quoted - and --raw is jq's -r. The letter is not offered, because -r is --region on apply (ADR-0021) and -q reads as "quiet" everywhere else in Unix.
- --raw without --jq is a usage error rather than a no-op, so a script never believes it asked for unquoted strings and silently got quoted ones (ADR2.2).
- Emitted object keys are sorted, matching jq -S rather than plain jq. gojq holds objects in Go maps, so key order cannot survive; the help text says so instead of the CLI pretending otherwise.
- A halt in the query ends the stream cleanly; halt_error is reported like any other runtime error.
- Filtered output still goes to stdout alone, and diagnostics still go to stderr, so --jq output stays as parseable as --json.

## Consequences

### Pros

- The JSON surface is useful on any machine the binary runs on, which is the whole reason the binary is self-contained.
- One quoting layer instead of two - the query is an argument, not a shell pipeline, so the shell only quotes it once.
- Every JSON-emitting command gained the flag from one helper, so the four cannot drift apart.

### Cons

- The binary carries a jq implementation and its module, for a feature that a user with jq installed can already do with a pipe.
- gojq is not bit-identical to jq - emitted object keys sort, so a consumer diffing whole objects against a real jq pipe will see a difference that is real but harmless.
- A jq version skew now exists, because the embedded engine advances when the dependency is bumped rather than when the user upgrades their own jq.
