# adrs-okf-yml — structured records, generated markdown

A prototype. `adrs/` remains the source of truth for the repo and is untouched.

Here the **YAML is the source and the markdown is a build artifact**. Each record is one `.yml` file where every prose field of the Lens template is a named key, validated by a JSON Schema; Jinja templates render each record into an OKF-conformant `.md` sibling, an index, and a Cytoscape graph of the decision network.

```
adrs-okf-yml/
├── adr.schema.json        # source   — the record contract (JSON Schema 2020-12)
├── NNNN-slug.yml          # source   — one record each (17)
├── templates/*.j2         # source   — Jinja templates
├── render.py              # source   — the build step
│
├── NNNN-slug.md           # GENERATED — OKF-conformant markdown, human-readable
├── index.md               # GENERATED — OKF reserved directory listing
├── graph.md               # GENERATED — richdocs companion with a cytoscape fence
└── graph.json             # GENERATED — Cytoscape elements payload
```

Never edit a generated file. Edit the `.yml` and re-run:

```sh
uv run --no-project adrs-okf-yml/render.py
```

**This directory is not an OKF bundle** — OKF is Markdown-only, and a folder of `.yml` files is a different format. The *generated* `.md` files are OKF-conformant; the YAML beside them is this project's own structure. Name it accordingly if it is adopted.

## What the structure buys

Three things the markdown forms cannot do:

1. **Typed graph edges.** `Relates to` was untyped prose; it is now `{relation, target, note}` over a closed vocabulary of eleven relations paired with inverses. A validator can then check inverse symmetry, which immediately found two one-way edges: `ADR-0006 --see_also--> ADR-0008` and `ADR-0012 --see_also--> ADR-0011`, neither with a back-edge. Those are gaps in the record set, not migration artifacts.
2. **Field-level linting.** `adr.schema.json` sets `additionalProperties: false`, so a typo in a key is an error rather than a silently ignored field. `consequences.cons` has `minItems: 1`, because a record with no cons has not been thought about. `description` must not end in a full stop, because it is a label, not a sentence.
3. **Derived views for free.** `graph.json` and `graph.md` are generated from the same `relates_to` blocks that render the markdown, so the diagram cannot drift from the records.

## Viewing the graph

`graph.json` is a Cytoscape payload: 23 nodes (17 records plus 6 compound plan-area parents), 12 typed edges. Nodes carry `title`, `description`, `plan_id`, `status` and `tags` in `data`, so a viewer can show detail on tap. Colour is on `data.colour` and encodes plan area, which means it survives a theme flip unchanged.

```sh
uv run --no-project .claude/skills/richdocs/scripts/md2html.py adrs-okf-yml/graph.md
cp adrs-okf-yml/graph.json tmp/richdocs/
uv run --no-project .claude/skills/richdocs/scripts/serve.py tmp/richdocs --open
```

The `cp` is needed because richdocs resolves a data pointer relative to the **served** directory, and `md2html.py` copies only the markdown.

## Fidelity

`tmp/okf_yml/verify.py` is the gate. It checks four things, and currently all four pass:

| Check | Result |
|---|---|
| Every record validates against `adr.schema.json` | 17/17 |
| Every `relates_to` target resolves | 12 edges |
| Generated argument byte-identical to `adrs/` | **17/17** |
| `graph.json` ids, endpoints and parents resolve | 23 nodes, 12 edges |

"Argument" means everything from `## Problem` onward — symptom, pain point, given/we prefer/because/unless, in practice, pros, cons. All of it survives the YAML round trip exactly, including the repo's sentence-per-line prose discipline, because multi-line prose uses YAML literal block scalars (`|-`).

The two fields that do **not** round-trip to the original wording are deliberate:

- **`Relates to`** — `Feeds the verdict in [ADR-0008]` becomes `See also [ADR-0008] (feeds the verdict)`. Turning prose into typed edges is one-way; the nuance survives in `note`, the phrasing does not.
- **`Status` on ADR-0017** — `Accepted, 2026-08-28 (split from ADR-0001 on 2026-08-29)` splits into `accepted_on`, `last_changed_on`, and a `split_from` edge.

## Two footguns found building this

**YAML implicit typing.** `accepted_on: 2026-08-28` unquoted parses as a `datetime.date`, not a string, so a JSON Schema `"type": "string"` rejects it. Any validator must check the JSON projection (`json.dumps(doc, default=str)`), not the raw YAML load. Same family as the Norway problem. `render.py` re-stringifies both date fields on load for the same reason.

**Over-blunt schema rules.** The first `description` rule banned the `.` character to enforce "no trailing full stop", which failed ADR-0015 because its description contains `~/.config`. The rule is now anchored (`\.$`). A schema is code; it needs the same care about edge cases.

## If this is adopted

`render.py` is Python in a Go repo, which is a smell. The build step should become a `make` target, and probably a Go generator using `goccy/go-yaml` (already a dependency per ADR-0014) with `text/template` instead of Jinja. The templates translate almost directly.

Scripts for the one-shot migration live in `tmp/okf_yml/` (gitignored):

```sh
uv run --no-project tmp/okf_yml/extract.py     # adrs/*.md   -> adrs-okf-yml/*.yml
uv run --no-project --with pyyaml --with jsonschema python3 tmp/okf_yml/verify.py
```
