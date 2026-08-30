# Documentation conventions

The declared dialect for this repository.
Where this file states a rule, it is the oracle; anything it leaves unstated falls back to whatever the repo consistently does.

## Dialect

- **ADR surface**: `okf-yaml` — decision records are authored as YAML in `adrs/NNNN-slug.yml`, and their markdown is generated.
- **ADR record shape**: OKF frontmatter, then a Lens blockquote (the reusable rule, first), then Problem, Decision (Given / We prefer / Because / Unless), and Consequences.
- **Generated paths**: `adrs/*.md`, `adrs/index.md`, `adrs/graph.md`, `adrs/graph.json`.
  These are build artifacts. Never hand-edit one; edit the `.yml` and regenerate.
- **Regenerate**: `make adrs`.
- **Record contract**: `adrs/record.schema.json`.
  Validation runs before rendering, so an invalid record stops the build rather than producing markdown nothing checked.
- **Agent file**: `AGENTS.md`, with `CLAUDE.md` including it.
- **Prose**: one sentence per line.

## Layout

| Path | Charter | Audience |
|---|---|---|
| `README.md` | What screenz is, and the shortest path to using it | A new user |
| `CONTRIBUTING.md` | Build, test tiers, house rules | A contributor |
| `AGENTS.md` | Operating rules for an agent working in this repo | An agent |
| `GLOSSARY.md` | The ubiquitous language: one canonical term per concept | Everyone |
| `docs/CONVENTIONS.md` | This file: how documentation itself is organised | An agent or maintainer filing a doc |
| `docs/install.md` | Installation and the Accessibility grant runbook | A new user |
| `docs/demos/` | The VHS tapes and recorded world behind the README GIFs | A maintainer regenerating demos |
| `adrs/*.yml` | One decision each, authored as data | A maintainer recording a decision |
| `adrs/index.md` | Generated listing of every record with its Lens | Anyone routing to a decision |
| `adrs/graph.md` | Generated view of the typed relation graph | Anyone tracing how decisions relate |

## How records are written

Records follow the template shape: OKF frontmatter carrying the metadata, a Lens blockquote (the reusable rule, first), then Problem, Decision (Given / We prefer / Because / Unless), and Consequences.
Read a Lens to apply a decision; open the record only for its argument.

Decisions are immutable in substance; a change of mind is a new ADR that supersedes the old one with links both ways, and supersession may retire a single clause (recorded in the earlier record's `status` and `relates_to`, never by rewriting the clause).

Two numbering axes exist by history: files are NNNN ordered, and code comments cite the planning-era IDs (ADR1.1 through ADR6.2) carried in each record's `plan_id` field.
That axis ended: records accepted after it carry `plan_id: null`, and ADR-0018 is the first.

## Typed relations

`relates_to` holds typed edges over a closed vocabulary, each paired with an inverse so symmetry is checkable:

| Relation | Inverse | Means |
|---|---|---|
| `extends` | `extended_by` | Adds a clause without replacing the earlier decision |
| `split_from` | `split_to` | Carved out when one record grew two decisions |
| `supersedes` | `superseded_by` | Replaces it; the earlier record stays, marked superseded |
| `depends_on` | `depended_on_by` | Only implementable because the other holds |
| `tests` | `tested_by` | Names how the other decision is verified |
| `excepts` | `excepted_by` | A scoped carve-out; the rule stands everywhere else |
| `see_also` | `see_also` | Related reasoning, no dependency |

`make adrs` reports any edge whose inverse is missing on the far record.
Those are gaps in the record set, not build failures: closing one is a decision, so it is reported and left alone.
