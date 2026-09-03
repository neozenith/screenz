---
type: Architecture Decision
title: An incomplete enumeration blocks only the runs it could have changed
description: A gap in the world stops the run when a rule could have matched into it, not merely because it exists
tags: [cli, discovery]
status: accepted
accepted_on: 2026-09-02
provenance: Slack held a stale AX server for four days and returned AXError -25211 for AXWindows; every apply refused, including "screenz a -m app=Code -d 2", whose selector could not have matched a Slack window under any circumstances
enforced_in:
  - internal/rule (Selector.CouldMatchApp)
  - internal/cli (apply.go - blockingAppErrs and the refusal gate in runApply)
generated: { by: human:neozenith, at: 2026-09-02T00:00:00Z }
---

> **Lens**: A safety gate must ask whether the defect could have changed this answer, not merely whether a defect exists - a gate that fires on runs it can prove are unaffected trains people to route around it.
> Narrow such a gate by proof, never by preference: refuse unless the tool can demonstrate the gap was irrelevant, and let everything unknowable count as relevant.

## Relates to

- Extends [ADR-0006](0006-report-unactionable-windows.md) (keeps the guarantee and narrows its trigger to the runs the gap could actually have changed)
- See also [ADR-0008](0008-strict-tolerance-verification.md) (both refuse to call a run successful on evidence the tool does not have)

## Problem

### Symptom

One application with a stale AX server made every apply exit 1, whatever the rules asked for, because the gate counted enumeration failures rather than weighing them.

### Pain point

The refusal was unactionable from inside the tool - no flag, no narrowing and no rule change could clear it - so the only route to a working apply was to quit the unrelated application, and the obvious next step for a user is to stop trusting the gate.

## Decision

### The lens

- **Given**: an application whose windows could not be read might have hidden a window some rule would have matched, and might equally have hidden nothing that could ever match
- **We prefer**: refusing only when a rule could have matched the unread application, judged from its name and bundle id, over refusing on the mere existence of a failure, over an --allow-incomplete override that moves the judgement to the user on every run, and over dropping the gate
- **Because**: the selector grammar already decides this - app= and bundle= terms are matched against an identity the failure report carries, so irrelevance is provable rather than assumed
- **Unless**: nothing about the application can be established, in which case it is treated as relevant and the run refuses

### In practice

- Selector.CouldMatchApp is one-sided. False is a proof that no window of the application can satisfy the selector; true is only the absence of that proof.
- A title= term never rules an application out, because a title is exactly what an unread window has not got.
- An empty app name or bundle id never rules an application out either; an application that could not be named cannot be dismissed.
- Every enumeration failure is still reported on stderr whether or not it blocks, so the gap is never hidden (ADR-0006).
- A refusal now names the rule and the application that caused it, so the reason is legible instead of a bare verdict.
- The check runs over the resolved rule set, so a profile's rules are weighed exactly like inline ones.
- Dry-run is unchanged - it warns and previews, because it moves nothing and can promise nothing.

## Consequences

### Pros

- The common case works - one stale application no longer holds every unrelated layout hostage.
- The gate keeps its meaning, so it is worth obeying when it does fire; the alternative was a gate people learn to bypass.
- A refusal is now diagnosable, naming the rule and the application rather than only the verdict.

### Cons

- The gate is more code and more subtlety than a count, and its correctness now rests on the selector semantics staying exact-match.
- A rule that reaches an application only through a title= term still blocks, which is correct but will read as arbitrary next to a bundle= rule that does not.
- A new selector key that addresses an application would have to be taught to CouldMatchApp, or it would silently stop ruling anything out.
