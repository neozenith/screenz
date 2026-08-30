# Demos

Animated documentation for the root README, rendered as code ([ADR-0019](../../adrs/0019-demos-are-vhs-tapes.md)).
Each GIF regenerates from its tape.
The tapes run in demo mode ([ADR-0018](../../adrs/0018-demo-mode-replays-a-recorded-world.md)) against the recorded world in [demo-world.json](demo-world.json), so no monitors are needed.

| Tape | GIF | Shows |
|------|-----|-------|
| [demo-status.tape](demo-status.tape) | [demo-status.gif](demo-status.gif) | `doctor` trust check, then `status` with two `--match` filters |
| [demo-apply.tape](demo-apply.tape) | [demo-apply.gif](demo-apply.gif) | `apply --dry-run` preview, verified apply, `exit=0` |
| [demo-profile.tape](demo-profile.tape) | [demo-profile.gif](demo-profile.gif) | `profile save`, the commented YAML, `apply --dry-run office` |

Re-render after a CLI change:

```sh
brew install vhs
cd docs/demos && for t in demo-*.tape; do vhs "$t"; done
```

The world file is a `screenz status --json` capture (recorded displays from 2026-08-28 plus two curated demo windows).
Capture a new one on real hardware with `screenz status --json > demo-world.json`, then hand-trim the window list.
