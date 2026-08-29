# ADR-0012: Spell regexes as /pattern/, terminated at the last slash

| Field | Value |
|---|---|
| **Status** | Accepted, 2026-08-28 |
| **Plan ID** | ADR4.2 |
| **Provenance** | The planning session's grammar decision; the swallowed-term guard was added after review found the two-regex ambiguity |
| **Relates to** | Shares the term grammar with ADR-0011 |
| **Enforced in** | internal/rule (parseTerms, cutRegex, swallowedTerm) |

> **Lens**: The same regex literal must work unchanged in CLI and YAML, and slashes inside window
> titles need no escaping; any remaining ambiguity fails loudly, never silently.

## Problem

### Symptom

Window titles contain slashes (paths, URLs), and regex values need a delimiter that survives both
shell quoting and YAML scalars.

### Pain point

A conventionally-terminated delimiter would force escaping inside titles, and a second spelling for
YAML would break lossless saves.

## Decision

### The lens

- **Given**: titles contain slashes and the value must round-trip between CLI and YAML unchanged
- **We prefer**: /pattern/ terminated at the last slash outside quoted regions, with an optional
  trailing i, over tilde or prefix spellings
- **Because**: last-slash termination makes interior slashes free, and Go's RE2 rejects look-arounds
  as clean usage errors
- **Unless**: a real pattern proves ambiguous under last-slash parsing; the swallowed-term guard now
  turns that case into a loud usage error rather than a silent mis-match

### In practice

- The terminator scan ignores quoted regions, and a pattern that captures a following key=value term
  errors with guidance.

## Consequences

### Pros

- Natural title patterns; one spelling everywhere.

### Cons

- Two regex terms in one list must be reordered or reduced to one.
