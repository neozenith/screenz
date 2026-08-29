# ADR 0015: Resolve the profile directory dotfiles-first

Plan ID: ADR5.2 | Date: 2026-08-28 | Status: accepted

## Decision

Profiles live in profiles/<name>.yaml under the first set entry of
$SCREENZ_HOME, $XDG_CONFIG_HOME/screenz, ~/.config/screenz. doctor prints
the resolved directory. os.UserConfigDir is deliberately not used.

## Why

The user versions tool config in a dotfiles repo; an XDG path is syncable
and space-free, while darwin's UserConfigDir ignores XDG and returns a path
containing a space under ~/Library. The env override keeps a wrong default
cheap for tests and client sites.
