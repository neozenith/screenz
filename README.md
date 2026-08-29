# screenz

Profile-driven bulk window layout CLI for macOS. One command moves whole
groups of application windows (by bundle id) onto named display regions,
reads every frame back, and exits 0 only when everything landed where it was
asked. Rules compose from flags and save as commented YAML profiles per
context (office, home, client).

## Quickstart

Install the latest release into `~/.work/bin` (no App Store, no Homebrew;
`gh` downloads never set the quarantine xattr):

```sh
mkdir -p ~/.work/bin
gh release download --repo neozenith/screenz \
  --pattern '*darwin_arm64.tar.gz' --output - | tar -xz -C ~/.work/bin
```

Intel Macs: use `--pattern '*darwin_amd64.tar.gz'`. Pin a version by adding
its tag: `gh release download v0.1.0 …`. After that first install,
`screenz update` self-updates in place (checksum-verified atomic swap;
`--check` to just look).

Then grant Accessibility to **your terminal app** (not the binary — TCC
attributes a shell-launched tool to the app hosting the shell) and verify:

```sh
screenz doctor                 # names the app to grant; exits 0 when trusted
screenz status                 # windows by app + displays
screenz profile init office    # commented template in ~/.config/screenz
screenz apply --dry-run office # preview the context switch
screenz apply office           # do it, verified
```

Full install and grant runbook (browser downloads, macOS 26.1 caveats,
`tccutil` reset): [docs/install.md](docs/install.md).

## Development

Everything runs through `make` (bootstraps its own pinned Go toolchain):
`make check` (pure tests, 100% coverage gate), `make itest` (real-window
integration tier, needs the Accessibility grant), `make build`,
`make release VERSION=vX.Y.Z`.

## License

[MIT](LICENSE)
