# ADR 0012: Spell regexes as /pattern/, terminated at the last slash

Plan ID: ADR4.2 | Date: 2026-08-28 | Status: accepted

## Decision

Regex values are written key=/pattern/ with an optional trailing i for
case-insensitive matching, terminated at the last slash outside quoted
regions. A pattern that would swallow a following term is a loud usage
error. Go regexp is RE2, so look-arounds are usage errors.

## Why

The same literal works unchanged in CLI flags and YAML values, and
last-slash termination means slashes inside window titles need no escaping.
The swallowed-term error keeps the rare two-regex ambiguity loud instead of
silently matching the wrong windows.

## Rejected

- key~pattern: tilde expands at word start when unquoted.
- key=re:pattern: no closing delimiter makes trailing spaces ambiguous.
