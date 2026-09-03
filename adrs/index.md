# Decision Records

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
* [ADR-0018: Demo mode replays a recorded world and simulates placement](0018-demo-mode-replays-a-recorded-world.md) - Fabricated output is a documentation prop: env-gated, cmd-wired, doctor-disclosed, never in tests
* [ADR-0019: Animated demos are VHS tapes rendered through demo mode](0019-demos-are-vhs-tapes.md) - A demo must be regenerable by script from checked-in sources; a human-at-the-desk recording is a screenshot with extra steps
* [ADR-0020: An omitted region means maximize](0020-omitted-region-means-maximize.md) - Where is mandatory, how big is defaulted; the common rule is the short one
* [ADR-0021: One-letter aliases for the rule grammar; destructive flags stay long](0021-one-letter-aliases-for-rule-flags.md) - A letter for what you type per rule, the full word for what you cannot undo
* [ADR-0022: Accept en-GB region spellings, canonicalise them on parse](0022-region-spellings-canonicalise-on-parse.md) - Type it either way, store it one way; an alias never reaches disk
* [ADR-0023: Every region name has a shorthand code; thirds use digits](0023-shorthand-codes-for-region-names.md) - Letters for halves and corners, digits for thirds, so no code is a transposition of another
* [ADR-0024: Every command answers to its initial](0024-commands-answer-to-their-initial.md) - The first letter is the command's short name, so a new command must claim a free one
* [ADR-0025: Three verbs, and a profile is named by flag](0025-three-verbs-profiles-named-by-flag.md) - A command is what you are doing; a profile is which one, and that is a flag
* [ADR-0026: status elides titles and takes a section](0026-status-elides-titles-and-takes-sections.md) - A table that holds its shape by default; the whole truth on request, and always in JSON
* [ADR-0027: Embed a jq engine behind --jq rather than shell out](0027-embed-jq-behind-a-jq-flag.md) - The filter travels in the binary, so --json stays useful where jq is not installed
* [ADR-0028: An incomplete enumeration blocks only the runs it could have changed](0028-incomplete-enumeration-blocks-only-what-it-could-change.md) - A gap in the world stops the run when a rule could have matched into it, not merely because it exists
# By group

## accessibility

* [ADR-0003](0003-fail-loudly-without-grant.md) - Fail with the responsible party named; never report an empty world as success
## cli

* [ADR-0011](0011-match-opens-a-rule.md) - One quoting layer; the CLI grammar is the saved-profile grammar
* [ADR-0012](0012-last-slash-regex.md) - One regex spelling everywhere; remaining ambiguity fails loudly
* [ADR-0013](0013-stdlib-flag-cli.md) - Stdlib parsing, hand-rolled dispatch; no CLI framework
* [ADR-0020](0020-omitted-region-means-maximize.md) - Where is mandatory, how big is defaulted; the common rule is the short one
* [ADR-0021](0021-one-letter-aliases-for-rule-flags.md) - A letter for what you type per rule, the full word for what you cannot undo
* [ADR-0022](0022-region-spellings-canonicalise-on-parse.md) - Type it either way, store it one way; an alias never reaches disk
* [ADR-0023](0023-shorthand-codes-for-region-names.md) - Letters for halves and corners, digits for thirds, so no code is a transposition of another
* [ADR-0024](0024-commands-answer-to-their-initial.md) - The first letter is the command's short name, so a new command must claim a free one
* [ADR-0025](0025-three-verbs-profiles-named-by-flag.md) - A command is what you are doing; a profile is which one, and that is a flag
* [ADR-0026](0026-status-elides-titles-and-takes-sections.md) - A table that holds its shape by default; the whole truth on request, and always in JSON
* [ADR-0027](0027-embed-jq-behind-a-jq-flag.md) - The filter travels in the binary, so --json stays useful where jq is not installed
* [ADR-0028](0028-incomplete-enumeration-blocks-only-what-it-could-change.md) - A gap in the world stops the run when a rule could have matched into it, not merely because it exists
## demo

* [ADR-0018](0018-demo-mode-replays-a-recorded-world.md) - Fabricated output is a documentation prop: env-gated, cmd-wired, doctor-disclosed, never in tests
* [ADR-0019](0019-demos-are-vhs-tapes.md) - A demo must be regenerable by script from checked-in sources; a human-at-the-desk recording is a screenshot with extra steps
## discovery

* [ADR-0005](0005-group-windows-by-bundle-id.md) - Group by product identity (bundle id), never process identity
* [ADR-0006](0006-report-unactionable-windows.md) - Never hide what the tool cannot act on; incomplete enumeration fails the run
* [ADR-0007](0007-display-index-y-then-minx.md) - Indexes follow the known ordering; stable identity is UUID or name
## distribution

* [ADR-0001](0001-distribute-via-github-releases.md) - Distribution machinery must match the audience that exists today
## macos

* [ADR-0002](0002-purego-bridge-cgo-free.md) - Bind OS primitives through purego, CGO off; prove each symbol on real hardware first
* [ADR-0017](0017-terminal-app-is-the-tcc-client.md) - Shell-launched only; the terminal app owns the grant (split from 0001)
## placement

* [ADR-0008](0008-strict-tolerance-verification.md) - Verification is never off; widen numerically and deliberately
* [ADR-0009](0009-never-activate-during-apply.md) - Bulk operations never fight the user for focus
* [ADR-0010](0010-retry-three-times.md) - Three attempts before judging a frame; retry, then once after 25 ms
## profiles

* [ADR-0014](0014-goccy-yaml-comments.md) - Hand comments are user data; every write preserves them and survives interruption
* [ADR-0015](0015-profile-dir-resolution.md) - Config is dotfiles-first: env override, XDG, ~/.config; never a path with a space
* [ADR-0016](0016-no-capture-from-current.md) - Automate the repeated task, not the rare one
## testing

* [ADR-0004](0004-two-test-tiers.md) - Pure coverage universal; the real seam tested on real hardware; a skip is a green lie
# Relationship graph

The typed edge set is rendered in [graph.md](graph.md), and generated as
[graph.json](graph.json) for any Cytoscape viewer.

* ADR-0001 --split_to--> ADR-0017
* ADR-0002 --tested_by--> ADR-0004
* ADR-0003 --extended_by--> ADR-0017
* ADR-0004 --tests--> ADR-0002
* ADR-0006 --see_also--> ADR-0008
* ADR-0006 --see_also--> ADR-0026
* ADR-0006 --extended_by--> ADR-0028
* ADR-0008 --see_also--> ADR-0010
* ADR-0010 --see_also--> ADR-0008
* ADR-0011 --depended_on_by--> ADR-0014
* ADR-0011 --extended_by--> ADR-0020
* ADR-0011 --extended_by--> ADR-0021
* ADR-0012 --see_also--> ADR-0011
* ADR-0013 --see_also--> ADR-0021
* ADR-0013 --extended_by--> ADR-0024
* ADR-0013 --extended_by--> ADR-0027
* ADR-0014 --depends_on--> ADR-0011
* ADR-0016 --see_also--> ADR-0025
* ADR-0017 --split_from--> ADR-0001
* ADR-0017 --extends--> ADR-0003
* ADR-0018 --excepts--> ADR-0008
* ADR-0018 --depends_on--> ADR-0004
* ADR-0019 --depends_on--> ADR-0018
* ADR-0020 --extends--> ADR-0011
* ADR-0020 --see_also--> ADR-0022
* ADR-0021 --extends--> ADR-0011
* ADR-0021 --see_also--> ADR-0013
* ADR-0021 --see_also--> ADR-0024
* ADR-0022 --see_also--> ADR-0020
* ADR-0022 --extended_by--> ADR-0023
* ADR-0023 --extends--> ADR-0022
* ADR-0024 --extends--> ADR-0013
* ADR-0024 --see_also--> ADR-0021
* ADR-0024 --extended_by--> ADR-0025
* ADR-0025 --see_also--> ADR-0016
* ADR-0025 --extends--> ADR-0024
* ADR-0026 --see_also--> ADR-0006
* ADR-0027 --extends--> ADR-0013
* ADR-0027 --see_also--> ADR-0021
* ADR-0027 --see_also--> ADR-0002
* ADR-0028 --extends--> ADR-0006
* ADR-0028 --see_also--> ADR-0008
