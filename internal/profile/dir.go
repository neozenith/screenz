// Package profile owns profile storage: directory resolution, the YAML
// schema, and comment-preserving load/save.
package profile

import "path/filepath"

// Dir resolves the profile directory as $SCREENZ_HOME → $XDG_CONFIG_HOME/screenz
// → ~/.config/screenz (ADR5.2). os.UserConfigDir is deliberately not used: on
// darwin it ignores XDG_CONFIG_HOME and returns a path with a space, which is
// awkward to version in a dotfiles repo.
func Dir(getenv func(string) string, home string) string {
	if d := getenv("SCREENZ_HOME"); d != "" {
		return d
	}
	if x := getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "screenz")
	}
	return filepath.Join(home, ".config", "screenz")
}
