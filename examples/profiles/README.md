# Example profiles

Runnable profiles to copy into your own profile directory, linked from [the install guide](../../docs/install.md).
Both carry the header comments `screenz` itself writes, so they double as a reference for the file format.

| Profile | Shows |
|---------|-------|
| [office.yaml](office.yaml) | A `displays:` alias map with regex display names, and rules addressing displays by alias |
| [laptop-only.yaml](laptop-only.yaml) | The no-alias case: rules addressing the built-in display directly |

Copy one into your profile directory and run it:

```sh
cp office.yaml ~/.config/screenz/profiles/
screenz apply --dry-run --profile office   # preview
screenz apply --profile office             # apply, verified
```

`screenz list` reports whether a profile fits the displays attached right now.
See [GLOSSARY.md](../../GLOSSARY.md) for the terms these files use.
