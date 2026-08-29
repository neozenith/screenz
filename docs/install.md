# Installing screenz

`screenz` is a single darwin binary published on the GitHub Releases page by
a tag-triggered CI workflow (ADR6.1). No App Store, no Homebrew, no Apple
account. macOS 13 or newer.

## 1. Download

With the GitHub CLI (no quarantine xattr is set):

```sh
gh release download --repo neozenith/screenz --pattern '*darwin_arm64*'
tar -xzf screenz_*_darwin_arm64.tar.gz
mkdir -p ~/.local/bin && mv screenz ~/.local/bin/   # any $PATH dir
screenz --version
```

Or with curl (also quarantine-free) — take the URL from the Releases page:

```sh
curl -LO https://github.com/neozenith/screenz/releases/latest/download/screenz_<version>_darwin_arm64.tar.gz
```

Intel Macs use the `darwin_amd64` tarball. Verify against `checksums.txt`
from the same release: `shasum -a 256 -c checksums.txt --ignore-missing`.

### Browser downloads are quarantined

Safari/Chrome downloads get the `com.apple.quarantine` xattr and Gatekeeper
will refuse the unsigned binary. Clear it once:

```sh
xattr -d com.apple.quarantine ~/Downloads/screenz
```

`screenz doctor` checks its own binary and prints this exact command when
needed.

## 2. Grant Accessibility — to your terminal app, not to screenz

TCC attributes a shell-launched tool to the app that hosts the shell
(ADR6.2). The grant must be given to **the terminal you run screenz from**:

| You run screenz in | Grant Accessibility to |
|--------------------|------------------------|
| Terminal.app       | Terminal               |
| iTerm2             | iTerm                  |
| Ghostty            | Ghostty                |
| VS Code integrated terminal | Visual Studio Code |

Steps:

1. Run `screenz doctor`. It names the responsible host app and (on first
   run) triggers the system prompt that deep-links the settings pane.
2. System Settings → Privacy & Security → Accessibility → enable (or add
   via **+**) that app.
3. Quit and reopen the terminal app, then `screenz doctor` again — it must
   print `accessibility: trusted` and exit 0.

Notes:

- The grant survives screenz upgrades: the terminal app is the TCC client,
  so the binary's ad-hoc signature never matters (ADR6.2).
- If the app is listed and enabled but doctor still says untrusted, reset
  and re-grant:

  ```sh
  tccutil reset Accessibility com.googlecode.iterm2   # your terminal's bundle id
  ```

- **macOS 26.1+**: a grant can be enforced while *not appearing* in System
  Settings. Trust `screenz doctor`'s answer over the Settings UI; the
  reset-and-regrant above clears the hidden state.

## 3. First run

```sh
screenz status                 # see windows and displays
screenz profile init office    # commented template in ~/.config/screenz
screenz apply --dry-run office # preview
screenz apply office           # the context switch
```

Profiles live in `$SCREENZ_HOME`, `$XDG_CONFIG_HOME/screenz` or
`~/.config/screenz` (ADR5.2) — dotfiles-friendly. Example profiles are in
[`examples/profiles/`](../examples/profiles/).

## Updating

An installed release updates itself:

```sh
screenz update --check   # report the latest release
screenz update           # download, verify checksums, atomic self-replace
```

The swap writes a sibling file and renames over the binary, so a failed
update never leaves a truncated executable. A source build (`screenz
version` says `dev`) refuses to overwrite itself unless you pass `--force`.
