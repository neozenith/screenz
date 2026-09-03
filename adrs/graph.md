# Decision relationship graph

28 decision records, 44 typed edges, grouped into
9 groups. Every edge comes from a record's `relates_to`
block, so this view cannot drift from the records.

```cytoscape
{ "data": "graph.json", "height": 620 }
```

## Edge vocabulary

| Relation | Inverse | Meaning |
|---|---|---|
| `extends` | `extended_by` | Adds a clause to an existing decision without replacing it |
| `split_from` | `split_to` | The record was carved out of another when one grew two decisions |
| `supersedes` | `superseded_by` | Replaces the earlier decision; the earlier record stays, marked superseded |
| `depends_on` | `depended_on_by` | The decision is only implementable because the other one holds |
| `tests` | `tested_by` | Names how the other decision is verified |
| `see_also` | `see_also` | Related reasoning, no dependency |

## Asymmetries

An edge whose inverse is missing on the far record. These are gaps in the
record set, not rendering artifacts.

* `ADR-0006` declares `see_also ADR-0008`, but `ADR-0008` has no `see_also ADR-0006`
* `ADR-0012` declares `see_also ADR-0011`, but `ADR-0011` has no `see_also ADR-0012`
* `ADR-0018` declares `excepts ADR-0008`, but `ADR-0008` has no `excepted_by ADR-0018`
* `ADR-0018` declares `depends_on ADR-0004`, but `ADR-0004` has no `depended_on_by ADR-0018`
* `ADR-0019` declares `depends_on ADR-0018`, but `ADR-0018` has no `depended_on_by ADR-0019`
* `ADR-0027` declares `see_also ADR-0021`, but `ADR-0021` has no `see_also ADR-0027`
* `ADR-0027` declares `see_also ADR-0002`, but `ADR-0002` has no `see_also ADR-0027`
* `ADR-0028` declares `see_also ADR-0008`, but `ADR-0008` has no `see_also ADR-0028`
