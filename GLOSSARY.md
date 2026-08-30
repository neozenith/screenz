# Glossary

The ubiquitous language for screenz.
Code identifiers, docs and prose use these terms; add new domain terms here in the same change that introduces them.

| Term | Definition |
|------|------------|
| AX points | The global coordinate space all frames use: origin at the main display's top left, y down, in points. AppKit's bottom-left-origin rects are flipped on the primary screen height. |
| action | One planned window placement: the window, the matching rule (1-based) and the target frame; `change: none` when the window is already within tolerance. |
| alias | A profile-defined name for a display spec (`displays:` map). Purely numeric names are reserved for the index shorthand. |
| bundle id | The application identity windows are grouped by (for example `com.microsoft.VSCode`), never the PID. |
| clamped | A placement whose read-back frame fell outside the rule's tolerance; the run exits 1. |
| demo mode | Env-gated replay (`SCREENZ_DEMO=<world file>`) with simulated placement, for curating demonstration text (ADR-0018). Doctor disclosed, never in tests. |
| display index | The human-friendly display number, ordered by row top-to-bottom then left-to-right. Unstable across layout changes; UUID and name are the stable keys. |
| display spec | The terms addressing one display: a bare index, an alias, or `index=`, `name=`, `uuid=`, `serial=`, `built-in=`, `main=` ANDed together. |
| first | Rule setting that places only the first matching window instead of every match: `--first` on the CLI, `each: false` in YAML. |
| gap | Points inset between a window and its region edge. |
| grant | The macOS Accessibility permission, held by the terminal app that launched screenz (the TCC client), not by the binary. |
| group | Every window of one application, keyed by bundle id; the unit selectors match over. |
| host app | The first real application ancestor of the shell screenz ran from; the TCC client that must hold the grant. |
| order | The sort applied to a rule's matched windows before placement: `existing`, `title` or `pid`. |
| plan | The full resolution of rules against a snapshot: actions, skipped windows and the unmatched count, computed before any window moves. |
| profile | A named, commented YAML rule set under the profile directory (`profiles/<name>.yaml`). |
| region | Where windows land on a display's usable frame: a named region, `grid=CxR`, or `unit=x,y,w,h`. |
| rule | One selector, display spec and region (plus gap, tolerance, order, first) applied to all matching windows. A window is placed by the first rule it matches. |
| selector | The window-matching terms of a rule: `bundle=`, `app=`, `title=` with literal, quoted or `/regex/i` values, ANDed together. |
| skipped | A matched window the rule cannot act on (state not normal): claimed and reported with its state, never silently dropped. |
| snapshot | One fully resolved discovery pass: displays, windows and any per-application enumeration errors. |
| state | A window's actionability: normal, minimized, hidden, sheet, dialog, offscreen (another Space), or unknown (frame read failed). Only normal windows are placed. |
| tolerance | The per-rule verification width: points, or percent of the target size per axis. Default 0.5 pt; never infinite. |
| unmatched | The count of windows no rule matched; reported in the plan and left untouched. |
| usable frame | The display area regions are computed over: the visible frame with the menu bar strip carved out. |
| world file | A serialised `status --json` capture (schema 1: displays + windows) that demo mode replays through the real pipeline. |
