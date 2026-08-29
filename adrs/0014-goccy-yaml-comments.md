# ADR-0014: Use goccy/go-yaml so profile comments survive saves

| Field | Value |
|---|---|
| **Status** | Accepted, 2026-08-28 |
| **Plan ID** | ADR5.1 |
| **Provenance** | The user asked for commented profiles; library research found yaml.v3 archived and comment-dropping |
| **Relates to** | Serialises the grammar from [ADR-0011](0011-match-opens-a-rule.md) |
| **Enforced in** | internal/profile (strict load, comment map, append-only saves, atomic writes) |

> **Lens**: A hand-commented file is user data; every write must preserve the comments and survive interruption.

## Problem

### Symptom

Profiles are commented specs, and yaml.v3 drops comments on struct marshal while being archived upstream.

### Pain point

A save that strips the user's comments, or a crash that truncates the file, destroys hand-authored content that cannot be regenerated.

## Decision

### The lens

- **Given**: comments must round-trip and the library landscape offers goccy's comment map as the maintained option
- **We prefer**: goccy/go-yaml with Strict decoding and CommentToMap/WithComment, over yaml.v3 or a release-candidate v4
- **Because**: unknown keys error instead of vanishing, and comments survive the round trip
- **Unless**: a maintained yaml v4 stabilises with equivalent comment support, which would prompt a re-evaluation

### In practice

- Saves are append-only so existing comment paths stay valid, block style only, written via temp-file rename.
- Blank lines do not survive a save and flow-style comments are lost; the template avoids both.

## Consequences

### Pros

- profile save changes only the appended rule; every comment survives.

### Cons

- One third-party dependency, pinned and exercised by round-trip tests.
