# Architecture Decision Records

Records follow the template shape: a metadata table, a Lens blockquote (the
reusable rule, first), then Problem, Decision (Given / We prefer / Because /
Unless), and Consequences. Read a Lens to apply a decision; open the file
only for its argument. Decisions are immutable in substance; a change of
mind is a new ADR that supersedes the old one with links both ways, and
supersession may retire a single clause (recorded in the earlier record's
Status and Relates to rows, never by rewriting the clause).

Two numbering axes exist by history: files are NNNN ordered, and code
comments cite the planning-era IDs (ADR1.1 through ADR6.2) carried in each
record's Plan ID row.

| File | Plan ID | Lens |
|------|---------|------|
| [0001](0001-distribute-via-github-releases.md) | ADR6.1 | Distribution machinery must match the audience that exists today |
| [0002](0002-purego-bridge-cgo-free.md) | ADR1.1 | Bind OS primitives through purego, CGO off; prove each symbol on real hardware first |
| [0003](0003-fail-loudly-without-grant.md) | ADR1.2 | Fail with the responsible party named; never report an empty world as success |
| [0004](0004-two-test-tiers.md) | ADR1.3 | Pure coverage universal; the real seam tested on real hardware; a skip is a green lie |
| [0005](0005-group-windows-by-bundle-id.md) | ADR2.1 | Group by product identity (bundle id), never process identity |
| [0006](0006-report-unactionable-windows.md) | ADR2.2 | Never hide what the tool cannot act on; incomplete sweeps fail the run |
| [0007](0007-display-index-y-then-minx.md) | ADR2.3 | Indexes follow the known ordering; stable identity is UUID or name |
| [0008](0008-strict-tolerance-verification.md) | ADR3.1 | Verification is never off; widen numerically and deliberately |
| [0009](0009-never-activate-during-apply.md) | ADR3.2 | Bulk operations never fight the user for focus |
| [0010](0010-retry-three-times.md) | ADR3.3 | Three attempts before judging a frame; retry, then once after 25 ms |
| [0011](0011-match-opens-a-rule.md) | ADR4.1 | One quoting layer; the CLI grammar is the saved-profile grammar |
| [0012](0012-last-slash-regex.md) | ADR4.2 | One regex spelling everywhere; remaining ambiguity fails loudly |
| [0013](0013-stdlib-flag-cli.md) | ADR4.3 | Stdlib parsing, hand-rolled dispatch; no CLI framework |
| [0014](0014-goccy-yaml-comments.md) | ADR5.1 | Hand comments are user data; every write preserves them and survives interruption |
| [0015](0015-profile-dir-resolution.md) | ADR5.2 | Config is dotfiles-first: env override, XDG, ~/.config; never a path with a space |
| [0016](0016-no-capture-from-current.md) | ADR5.3 | Automate the repeated task, not the rare one |
| [0017](0017-terminal-app-is-the-tcc-client.md) | ADR6.2 | Shell-launched only; the terminal app owns the grant (split from 0001) |
