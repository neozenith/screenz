#!/usr/bin/env -S uv run --quiet --no-project --with pyyaml --with jinja2 python3
"""Render the YAML records in this directory into their generated siblings.

    *.yml  ->  NNNN-slug.md   OKF-conformant markdown, one per record
               index.md       OKF reserved directory listing
               graph.md       richdocs companion with a ```cytoscape fence
               graph.json     Cytoscape elements payload

The YAML is the source of truth. Every .md and .json here is generated; edit
the .yml and re-run. Run from the repo root:

    uv run --no-project adrs-okf-yml/render.py
"""

from __future__ import annotations

import json
import re
from pathlib import Path

import yaml
from jinja2 import Environment, FileSystemLoader, StrictUndefined

HERE = Path(__file__).parent
TEMPLATES = HERE / "templates"

RESERVED_WORDS = {"y", "n", "yes", "no", "true", "false", "on", "off", "null", "~"}

INVERSE = {
    "extends": "extended_by", "extended_by": "extends",
    "split_from": "split_to", "split_to": "split_from",
    "supersedes": "superseded_by", "superseded_by": "supersedes",
    "tests": "tested_by", "tested_by": "tests",
    "depends_on": "depended_on_by", "depended_on_by": "depends_on",
    "see_also": "see_also",
}
PROSE = {
    "extends": "Extends", "extended_by": "Extended by",
    "split_from": "Split from", "split_to": "Split to",
    "supersedes": "Supersedes", "superseded_by": "Superseded by",
    "tests": "Tests", "tested_by": "Tested by",
    "depends_on": "Depends on", "depended_on_by": "Depended on by",
    "see_also": "See also",
}
# Plan areas come from the planning-era ADRn.n identifier, the second numbering
# axis the records carry. They become Cytoscape compound nodes.
AREAS = {
    "1": "macOS bridge",
    "2": "Discovery",
    "3": "Placement",
    "4": "CLI grammar",
    "5": "Profiles",
    "6": "Distribution",
}
# Data-encoding colours: they mean "plan area", so they stay fixed across
# themes and brands rather than coming from the brandpack.
AREA_COLOURS = {
    "1": "#6366f1", "2": "#0891b2", "3": "#059669",
    "4": "#d97706", "5": "#db2777", "6": "#7c3aed",
}


def yamlq(s: str) -> str:
    """Quote a YAML scalar only when the plain form would be mis-read."""
    s = str(s)
    if (
        s == ""
        or s[0] in ">|*&!%@`-?{}[],#'\"' "
        or s[-1] in ": "
        or ": " in s
        or " #" in s
        or s.lower() in RESERVED_WORDS
        or re.fullmatch(r"[-+]?[\d._eE:]+", s) is not None
    ):
        return '"' + s.replace("\\", "\\\\").replace('"', '\\"') + '"'
    return s


def load() -> list[dict]:
    recs = []
    for p in sorted(HERE.glob("0*.yml")):
        doc = yaml.safe_load(p.read_text())
        # YAML resolves an unquoted ISO date to datetime.date; the templates and
        # the JSON payload both want the ISO string back.
        for key in ("accepted_on", "last_changed_on"):
            doc[key] = str(doc[key])
        recs.append(doc)
    return recs


def asymmetries(recs: list[dict]) -> list[dict]:
    by_id = {r["id"]: r for r in recs}
    out = []
    for rec in recs:
        for e in rec["relates_to"]:
            want = INVERSE[e["relation"]]
            back = by_id[e["target"]]["relates_to"]
            if not any(b["target"] == rec["id"] and b["relation"] == want for b in back):
                out.append({"source": rec["id"], "relation": e["relation"],
                            "target": e["target"], "inverse": want})
    return out


def cytoscape(recs: list[dict]) -> dict:
    elements = []
    areas = sorted({r["plan_id"][3] for r in recs})
    for a in areas:
        elements.append({"data": {
            "id": f"area-{a}", "label": f"{a}. {AREAS[a]}", "colour": AREA_COLOURS[a],
        }})
    for r in recs:
        area = r["plan_id"][3]
        elements.append({"data": {
            "id": r["id"],
            "label": r["id"].replace("ADR-", ""),
            "parent": f"area-{area}",
            "colour": AREA_COLOURS[area],
            "title": r["title"],
            "description": r["description"],
            "plan_id": r["plan_id"],
            "status": r["status"],
            "tags": ", ".join(r["tags"]),
        }})
    seen = set()
    for r in recs:
        for e in r["relates_to"]:
            eid = f"{r['id']}-{e['relation']}-{e['target']}"
            if eid in seen:
                continue
            seen.add(eid)
            elements.append({"data": {
                "id": eid, "source": r["id"], "target": e["target"],
                "label": e["relation"].replace("_", " "),
                "relation": e["relation"],
                "note": e.get("note", ""),
            }})
    return {
        "elements": elements,
        "layout": {"name": "dagre", "rankDir": "LR"},
        "height": 620,
    }


def main() -> None:
    recs = load()
    by_id = {r["id"]: r for r in recs}

    env = Environment(
        loader=FileSystemLoader(TEMPLATES),
        undefined=StrictUndefined,
        keep_trailing_newline=True,
        trim_blocks=True,
        lstrip_blocks=True,
    )
    env.filters["yamlq"] = yamlq
    env.filters["blockquote"] = lambda s: s.replace("\n", "\n> ")
    env.filters["relation_prose"] = lambda r: PROSE[r]
    env.filters["adr_link"] = lambda i: f"{i[4:]}-{by_id[i]['slug']}.md"

    adr_t = env.get_template("adr.md.j2")
    for rec in recs:
        (HERE / f"{rec['id'][4:]}-{rec['slug']}.md").write_text(adr_t.render(rec=rec))

    edges = [
        {"source": r["id"], "relation": e["relation"], "target": e["target"]}
        for r in recs for e in r["relates_to"]
    ]
    areas = [
        (f"{a}. {AREAS[a]}", [r for r in recs if r["plan_id"][3] == a])
        for a in sorted({r["plan_id"][3] for r in recs})
    ]
    (HERE / "index.md").write_text(
        env.get_template("index.md.j2").render(records=recs, areas=areas, edges=edges)
    )

    graph = cytoscape(recs)
    (HERE / "graph.json").write_text(json.dumps(graph, indent=2) + "\n")
    (HERE / "graph.md").write_text(
        env.get_template("graph.md.j2").render(
            node_count=len(recs),
            edge_count=len(edges),
            area_count=len(areas),
            asymmetries=asymmetries(recs),
        )
    )

    print(f"rendered {len(recs)} records -> .md siblings, index.md, graph.md, graph.json")
    print(f"  {len(graph['elements'])} cytoscape elements ({len(edges)} edges)")
    print("\nview the graph:")
    print("  uv run --no-project .claude/skills/richdocs/scripts/md2html.py"
          " adrs-okf-yml/graph.md")
    print("  cp adrs-okf-yml/graph.json tmp/richdocs/")
    print("  uv run --no-project .claude/skills/richdocs/scripts/serve.py tmp/richdocs --open")


if __name__ == "__main__":
    main()
