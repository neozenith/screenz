# Architecture Decision Records

Two numbering axes exist by history: files are NNNN ordered, and code
comments cite the planning-era IDs (ADR1.1 through ADR6.2) recorded here as
each file's Plan ID. The table maps both. Decisions are immutable; a new
ADR supersedes an old one with a cross-link.

| File | Plan ID | Decision |
|------|---------|----------|
| [0001](0001-distribute-via-github-releases.md) | ADR6.1, ADR6.2 | Distribute via CI-built GitHub Releases; the terminal app is the TCC client |
| [0002](0002-purego-bridge-cgo-free.md) | ADR1.1 | purego bridge with CGO_ENABLED=0 |
| [0003](0003-fail-loudly-without-grant.md) | ADR1.2 | Fail loudly without the Accessibility grant |
| [0004](0004-two-test-tiers.md) | ADR1.3 | Two test tiers; itest fails, never skips |
| [0005](0005-group-windows-by-bundle-id.md) | ADR2.1 | Group windows by bundle id |
| [0006](0006-report-unactionable-windows.md) | ADR2.2 | Report un-actionable windows |
| [0007](0007-display-index-y-then-minx.md) | ADR2.3 | Display index ordered by row then minX |
| [0008](0008-strict-tolerance-verification.md) | ADR3.1 | Strict read-back verification with numeric tolerance |
| [0009](0009-never-activate-during-apply.md) | ADR3.2 | Never activate applications during apply |
| [0010](0010-retry-three-times.md) | ADR3.3 | Retry a mismatched frame three times |
| [0011](0011-match-opens-a-rule.md) | ADR4.1 | Each --match opens a rule |
| [0012](0012-last-slash-regex.md) | ADR4.2 | Last-slash regex termination |
| [0013](0013-stdlib-flag-cli.md) | ADR4.3 | Stdlib flag, hand-rolled dispatch |
| [0014](0014-goccy-yaml-comments.md) | ADR5.1 | goccy/go-yaml comment preservation |
| [0015](0015-profile-dir-resolution.md) | ADR5.2 | Dotfiles-first profile directory |
| [0016](0016-no-capture-from-current.md) | ADR5.3 | No capture-from-current |
