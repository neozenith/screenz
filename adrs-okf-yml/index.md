# Architecture Decision Records

* [ADR-0001: Distribute via CI-built artifacts on the GitHub Releases page](0001-distribute-via-github-releases.md) - Distribution machinery must match the audience that exists today
* [ADR-0002: Bridge to macOS through purego with CGO_ENABLED=0](0002-purego-bridge-cgo-free.md) - Bind OS primitives through purego, CGO off; prove each symbol on real hardware first
* [ADR-0003: Fail loudly when Accessibility is not granted](0003-fail-loudly-without-grant.md) - Fail with the responsible party named; never report an empty world as success
* [ADR-0004: Two test tiers; the integration tier fails, never skips](0004-two-test-tiers.md) - Pure coverage universal; the real seam tested on real hardware; a skip is a green lie
* [ADR-0005: Group windows by bundle id, not PID](0005-group-windows-by-bundle-id.md) - Group by product identity (bundle id), never process identity
* [ADR-0006: Report windows the tool cannot act on instead of hiding them](0006-report-unactionable-windows.md) - Never hide what the tool cannot act on; incomplete enumeration fails the run
* [ADR-0007: Order display indexes by row, then left to right](0007-display-index-y-then-minx.md) - Indexes follow the known ordering; stable identity is UUID or name
* [ADR-0008: Fail strictly on a mismatched frame, with a numeric tolerance](0008-strict-tolerance-verification.md) - Verification is never off; widen numerically and deliberately
* [ADR-0009: Never activate applications during a bulk apply](0009-never-activate-during-apply.md) - Bulk operations never fight the user for focus
* [ADR-0010: Retry a mismatched frame three times with a 25 ms pause](0010-retry-three-times.md) - Three attempts before judging a frame; retry, then once after 25 ms
* [ADR-0011: Each --match opens a rule; sibling flags bind to it](0011-match-opens-a-rule.md) - One quoting layer; the CLI grammar is the saved-profile grammar
* [ADR-0012: Spell regexes as /pattern/, terminated at the last slash](0012-last-slash-regex.md) - One regex spelling everywhere; remaining ambiguity fails loudly
* [ADR-0013: Use stdlib flag with a hand-rolled subcommand table](0013-stdlib-flag-cli.md) - Stdlib parsing, hand-rolled dispatch; no CLI framework
* [ADR-0014: Use goccy/go-yaml so profile comments survive saves](0014-goccy-yaml-comments.md) - Hand comments are user data; every write preserves them and survives interruption
* [ADR-0015: Resolve the profile directory dotfiles-first](0015-profile-dir-resolution.md) - Config is dotfiles-first: env override, XDG, ~/.config; never a path with a space
* [ADR-0016: Profiles are authored from flags or the template, not captured](0016-no-capture-from-current.md) - Automate the repeated task, not the rare one
* [ADR-0017: The terminal app is the TCC client; no daemon mode](0017-terminal-app-is-the-tcc-client.md) - Shell-launched only; the terminal app owns the grant (split from 0001)
# By plan area

## 1. macOS bridge

* [ADR-0002](0002-purego-bridge-cgo-free.md) - Bind OS primitives through purego, CGO off; prove each symbol on real hardware first
* [ADR-0003](0003-fail-loudly-without-grant.md) - Fail with the responsible party named; never report an empty world as success
* [ADR-0004](0004-two-test-tiers.md) - Pure coverage universal; the real seam tested on real hardware; a skip is a green lie
## 2. Discovery

* [ADR-0005](0005-group-windows-by-bundle-id.md) - Group by product identity (bundle id), never process identity
* [ADR-0006](0006-report-unactionable-windows.md) - Never hide what the tool cannot act on; incomplete enumeration fails the run
* [ADR-0007](0007-display-index-y-then-minx.md) - Indexes follow the known ordering; stable identity is UUID or name
## 3. Placement

* [ADR-0008](0008-strict-tolerance-verification.md) - Verification is never off; widen numerically and deliberately
* [ADR-0009](0009-never-activate-during-apply.md) - Bulk operations never fight the user for focus
* [ADR-0010](0010-retry-three-times.md) - Three attempts before judging a frame; retry, then once after 25 ms
## 4. CLI grammar

* [ADR-0011](0011-match-opens-a-rule.md) - One quoting layer; the CLI grammar is the saved-profile grammar
* [ADR-0012](0012-last-slash-regex.md) - One regex spelling everywhere; remaining ambiguity fails loudly
* [ADR-0013](0013-stdlib-flag-cli.md) - Stdlib parsing, hand-rolled dispatch; no CLI framework
## 5. Profiles

* [ADR-0014](0014-goccy-yaml-comments.md) - Hand comments are user data; every write preserves them and survives interruption
* [ADR-0015](0015-profile-dir-resolution.md) - Config is dotfiles-first: env override, XDG, ~/.config; never a path with a space
* [ADR-0016](0016-no-capture-from-current.md) - Automate the repeated task, not the rare one
## 6. Distribution

* [ADR-0001](0001-distribute-via-github-releases.md) - Distribution machinery must match the audience that exists today
* [ADR-0017](0017-terminal-app-is-the-tcc-client.md) - Shell-launched only; the terminal app owns the grant (split from 0001)
# Relationship graph

The typed edge set is rendered in [graph.md](graph.md), and generated as
[graph.json](graph.json) for any Cytoscape viewer.

* ADR-0001 --split_to--> ADR-0017
* ADR-0002 --tested_by--> ADR-0004
* ADR-0003 --extended_by--> ADR-0017
* ADR-0004 --tests--> ADR-0002
* ADR-0006 --see_also--> ADR-0008
* ADR-0008 --see_also--> ADR-0010
* ADR-0010 --see_also--> ADR-0008
* ADR-0011 --depended_on_by--> ADR-0014
* ADR-0012 --see_also--> ADR-0011
* ADR-0014 --depends_on--> ADR-0011
* ADR-0017 --split_from--> ADR-0001
* ADR-0017 --extends--> ADR-0003
