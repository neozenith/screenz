# Installing screenz

`screenz` is a single darwin binary published on the GitHub Releases page by a tag-triggered CI workflow ([ADR-0001](../adrs/0001-distribute-via-github-releases.md)).
No App Store, no Homebrew, no Apple account. macOS 13 or newer.

## 1. Download

With the GitHub CLI (no quarantine xattr is set):

```sh
gh release download --repo neozenith/screenz \
  --pattern '*darwin_arm64*' --pattern checksums.txt
tar -xzf screenz_*_darwin_arm64.tar.gz
mkdir -p ~/.work/bin && mv screenz ~/.work/bin/   # any $PATH dir
screenz --version
```

Or with curl (also quarantine-free): take the URL from the Releases page (asset names embed the tag, so `/latest/download/` cannot be used):

```sh
TAG=v0.6.0   # the release you want; the tag appears twice in the URL
curl -LO https://github.com/neozenith/screenz/releases/download/$TAG/screenz_${TAG}_darwin_arm64.tar.gz
curl -LO https://github.com/neozenith/screenz/releases/download/$TAG/checksums.txt
```

Intel Macs use the `darwin_amd64` tarball; `uname -m` prints `arm64` on Apple Silicon and `x86_64` on Intel.
Verify against `checksums.txt` from the same release: `shasum -a 256 -c checksums.txt --ignore-missing`.

### Browser downloads are quarantined

Safari/Chrome downloads get the `com.apple.quarantine` xattr and Gatekeeper will refuse the unsigned binary.
Clear it once:

```sh
xattr -d com.apple.quarantine ~/Downloads/screenz
```

`screenz doctor` checks its own binary and prints this command with your binary's actual path when needed.

## 2. Grant Accessibility to your terminal app, not to screenz

TCC attributes a shell-launched tool to the app that hosts the shell ([ADR-0017](../adrs/0017-terminal-app-is-the-tcc-client.md)).
The grant must be given to **the terminal you run screenz from**:

| You run screenz in | Grant Accessibility to |
|--------------------|------------------------|
| Terminal.app       | Terminal               |
| iTerm2             | iTerm                  |
| Ghostty            | Ghostty                |
| VS Code integrated terminal | Visual Studio Code |

Steps:

1. Run `screenz doctor`.
   It names the responsible host app and (on first run) triggers the system prompt that deep-links the settings pane.
2. System Settings → Privacy & Security → Accessibility → enable (or add via **+**) that app.
3. Quit and reopen the terminal app, then `screenz doctor` again.
   It must print `accessibility: trusted` and exit 0.

Notes:

- The grant survives screenz upgrades: the terminal app is the TCC client, so the binary's ad-hoc signature never matters ([ADR-0017](../adrs/0017-terminal-app-is-the-tcc-client.md)).
- If the app is listed and enabled but doctor still says untrusted, reset and re-grant:

  ```sh
  tccutil reset Accessibility com.googlecode.iterm2   # your terminal's bundle id
  ```

- **macOS 26.1+**: a grant can be enforced while *not appearing* in System Settings.
  Trust `screenz doctor`'s answer over the Settings UI; the reset-and-regrant above clears the hidden state.

## 3. First run

```sh
screenz status                 # see windows and displays
screenz init --profile office  # commented template in ~/.config/screenz/profiles/
screenz apply -n -p office     # preview (-n is --dry-run)
screenz apply -p office        # the context switch
```

Profiles live in `$SCREENZ_HOME`, `$XDG_CONFIG_HOME/screenz` or `~/.config/screenz` ([ADR-0015](../adrs/0015-profile-dir-resolution.md)), which is dotfiles-friendly.
Example profiles are in [`examples/profiles/`](../examples/profiles/).

## 4. The short link and shell completions

An install is three things: the binary, a short name to type, and a completion script ([ADR-0029](../adrs/0029-update-owns-the-whole-install.md)).
`update` maintains all three:

```sh
screenz update --all           # the release, the sz link, this shell's completions
screenz update --link          # just the sz symlink beside the binary
screenz update --completions   # just this shell's script ($SHELL decides which)
screenz update --shell all     # write zsh, bash and fish scripts
screenz update --check --all   # report all three, change nothing
```

Naming a part does that part alone and needs no network.

**The short link.** `--link` creates `sz` as a symlink beside the screenz binary, so `sz apply -p office` works wherever `screenz` is already on `PATH`, with no alias to carry in a dotfiles repo.
The target is relative, so moving the install directory keeps the pair intact.
If something else already holds that name — `lrzsz` ships an `sz` — screenz refuses and says what it found rather than replacing it.

**Completions.** Scripts are generated from screenz's own flag definitions ([ADR-0030](../adrs/0030-completions-generated-from-the-parser.md)), so commands, their initials, every flag and its one-letter alias, the region names, and your profile names all complete.
They are written to `completions/` beside `profiles/` in the screenz config directory, and nowhere else.
Wiring them up is one line, which the command prints and you paste:

| Shell | Line to add | Where |
|-------|-------------|-------|
| zsh   | `fpath=(~/.config/screenz/completions $fpath)` | `~/.zshrc`, before `compinit` |
| bash  | `source ~/.config/screenz/completions/screenz.bash` | `~/.bashrc` |
| fish  | `source ~/.config/screenz/completions/screenz.fish` | `~/.config/fish/config.fish` |

Rerun `screenz update --all` after an upgrade to regenerate the scripts against the new binary.
`screenz doctor` reports the short link and which scripts are installed.

## Updating

An installed release updates itself:

```sh
screenz update --check   # report the latest release
screenz update           # download, verify checksums, atomic self-replace
screenz update --all     # and refresh the sz link and completions with it
```

The swap writes a sibling file and renames over the binary, so a failed update never leaves a truncated executable.
A source build (`screenz version` says `dev`) refuses to overwrite itself unless you pass `--force`.
A bare `screenz update` also names on stderr anything the rest of the install is still missing.
