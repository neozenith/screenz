<!-- SPIKE: candidate root README.md if asciinema (+ agg) is chosen for animated docs.
     GIF paths are siblings here; promotion to root would move them to docs/demos/. -->

# screenz

Profile-driven bulk window layout CLI for macOS.
One command moves whole groups of application windows (by bundle id) onto named display regions, reads every frame back, and exits 0 only when everything landed where it was asked.
Rules compose from flags and save as commented YAML profiles per context (office, home, client).

![screenz apply: preview, move, verify every frame, exit 0](demo-apply.gif)

The demo is the office context switch on one display: VS Code takes the left half, Chrome the right.
Every `apply` is verified.
The plan is previewable with `--dry-run`, each moved window's frame is read back, and `TARGET` must equal `ACTUAL` within tolerance for exit 0.

## See your world first

`screenz doctor` names the app that must hold the Accessibility grant and checks every binding.
`screenz status` shows windows grouped by application and every connected display, filterable with the same `--match` grammar rules use.

![screenz doctor and status: trust check, then the world](demo-status.gif)

## Save the context, replay it forever

A rule that works once becomes a named profile.
Profiles are commented YAML, and your hand-written comments survive every save.

![screenz profile: save, inspect the YAML, replay](demo-profile.gif)

## Quickstart

Install the latest release into `~/.work/bin` (no App Store, no Homebrew; `gh` downloads never set the quarantine xattr):

```sh
mkdir -p ~/.work/bin
gh release download --repo neozenith/screenz \
  --pattern '*darwin_arm64.tar.gz' --output - | tar -xz -C ~/.work/bin
```

Intel Macs: use `--pattern '*darwin_amd64.tar.gz'`.
Pin a version by adding its tag: `gh release download v0.1.0 …`.
After that first install, `screenz update` self-updates in place (checksum-verified atomic swap; `--check` to just look).

Then grant Accessibility to **your terminal app** (not the binary; TCC attributes a shell-launched tool to the app hosting the shell) and verify:

```sh
screenz doctor                 # names the app to grant; exits 0 when trusted
screenz status                 # windows by app + displays
screenz profile init office    # commented template in ~/.config/screenz/profiles/
screenz apply --dry-run office # preview the context switch
screenz apply office           # do it, verified
```

Full install and grant runbook (browser downloads, macOS 26.1 caveats, `tccutil` reset): [docs/install.md](../../install.md).

## How these demos are made

Each demo is an [asciinema](https://asciinema.org) recording (`*.cast`), captured by the checked-in driver [`record.sh`](record.sh) and rendered to GIF with [`agg`](https://github.com/asciinema/agg):

```sh
brew install asciinema agg
cd docs/spikes/asciinema
asciinema rec --window-size 120x22 -c "bash record.sh apply" demo-apply.cast
agg --font-size 14 demo-apply.cast demo-apply.gif
```

The `.cast` files are plain text (asciicast), so they diff in review and replay in the terminal with `asciinema play`.
They could also be hosted on asciinema.org's shareable web player instead of committing GIFs.
The recordings are real Accessibility-granted runs against live VS Code and Chrome windows on this machine.

One known limitation: `agg` renders the raw terminal only and has no window-chrome option, so these GIFs carry no title bar.
A title bar needs either asciinema.org's hosted player (chrome comes from the embed) or an SVG pipeline such as `svg-term-cli --window`.
VHS draws one natively with `Set WindowBar`; see the sibling spike.

## Development

See [CONTRIBUTING.md](../../../CONTRIBUTING.md) for build, test tiers and house rules, [adrs/](../../../adrs/README.md) for design decisions, and [GLOSSARY.md](../../../GLOSSARY.md) for the project vocabulary.

## License

[MIT](../../../LICENSE)
