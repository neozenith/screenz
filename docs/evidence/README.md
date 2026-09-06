# Evidence transcripts

Proof-of-execution transcripts from real runs on the maintainer's machine (see [CONTRIBUTING.md](../../CONTRIBUTING.md)).
Point-in-time artifacts: each filename carries the ISO date of the run and is never updated in place.

| File | Produced by | As of |
|------|-------------|-------|
| [apply-2026-08-28.json](apply-2026-08-28.json) | `screenz apply --json` | 2026-08-28 |
| [apply-dry-run-2026-08-28.txt](apply-dry-run-2026-08-28.txt) | `bin/screenz apply --dry-run` | 2026-08-28 |
| [doctor-2026-08-28.txt](doctor-2026-08-28.txt) | `screenz doctor` | 2026-08-28 |
| [install-2026-08-29.txt](install-2026-08-29.txt) | `gh release view` / release-install walkthrough | 2026-08-29 |
| [profile-roundtrip-2026-08-28.diff](profile-roundtrip-2026-08-28.diff) | `screenz profile save` round-trip diff | 2026-08-28 |
| [status-2026-08-28.json](status-2026-08-28.json) | `screenz status --json` | 2026-08-28 |
| [status-after-apply-2026-08-28.json](status-after-apply-2026-08-28.json) | `screenz status --json` after apply | 2026-08-28 |

The commands above are the ones that ran on the day, not today's spelling.
Transcripts dated before 2026-09-01 predate ADR-0025: `profile save NAME …` is now `apply … --save-profile NAME`, and `apply NAME` is now `apply --profile NAME`.
Nothing here is re-run to match a rename, because a transcript that is edited stops being evidence.
