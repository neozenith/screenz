# screenz

Profile-driven bulk window layout CLI for macOS.
One command moves whole groups of application windows (by bundle id) onto named display regions, reads every frame back, and exits 0 only when everything landed where it was asked.
Rules compose from flags and save as commented YAML profiles per context (office, home, client).

![screenz apply: preview, move, verify every frame, exit 0](docs/demos/demo-apply.gif)

The demo is the office context switch: VS Code maximises onto the built-in display, Chrome docks to an external's left half.
Every `apply` is verified.
The plan is previewable with `--dry-run`, each moved window's frame is read back, and `TARGET` must equal `ACTUAL` within tolerance for exit 0.

## See your world first

`screenz doctor` names the app that must hold the Accessibility grant and checks every binding.
`screenz status` shows windows grouped by application and every connected display, filterable with the same `--match` grammar rules use.

![screenz doctor and status: trust check, then the world](docs/demos/demo-status.gif)

## Save the context, replay it forever

A rule that works once becomes a named profile: add `--save-profile NAME` to the invocation that just worked.
Profiles are commented YAML, and your hand-written comments — and your display aliases — survive every save.

![screenz profile: save, inspect the YAML, replay](docs/demos/demo-profile.gif)

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
screenz doctor                  # names the app to grant; exits 0 when trusted
screenz status                  # windows by app + displays
screenz init --profile office   # commented template in ~/.config/screenz/profiles/
screenz list                    # which profiles fit the displays you have now
screenz apply -n -p office      # preview the context switch (-n is --dry-run)
screenz apply -p office         # do it, verified
```

Full install and grant runbook (browser downloads, macOS 26.1 caveats, `tccutil` reset): [docs/install.md](docs/install.md).

## How the demos are made

The demos are **code**: each GIF renders from a checked-in [VHS](https://github.com/charmbracelet/vhs) tape in [docs/demos/](docs/demos/), so they regenerate deterministically whenever the CLI changes.

```sh
brew install vhs
cd docs/demos && vhs demo-apply.tape   # re-render one demo
```

The tapes run in demo mode ([ADR-0018](adrs/0018-demo-mode-replays-a-recorded-world.md)): `SCREENZ_DEMO=demo-world.json` replays a recorded three-display world through the real pipeline with placement simulated.
Every line except the ACTUAL column is genuinely computed, `screenz doctor` discloses the mode, and no monitors are needed to regenerate.

## Development

See [CONTRIBUTING.md](CONTRIBUTING.md) for build, test tiers and house rules, [adrs/](adrs/index.md) for design decisions, and [GLOSSARY.md](GLOSSARY.md) for the project vocabulary.

## License

[MIT](LICENSE)
