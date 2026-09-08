// Package install owns the parts of a screenz installation that are not
// the binary itself: the `sz` short link beside it, and the shell
// completion scripts under the screenz config directory (ADR-0029).
//
// Everything here is pure over paths except CreateLink and WriteScript,
// which are the two filesystem writes; `screenz update` drives them and
// `screenz doctor` reports what they left behind.
package install

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// LinkName is the short name screenz answers to. It is created only ever
// as a sibling of the binary, so it lands on PATH exactly when screenz
// already is, and never in a directory screenz does not own (ADR-0029).
const LinkName = "sz"

// The states the short link can be found in. Foreign covers everything
// screenz did not put there — a regular file, a directory, or a symlink to
// something else (lrzsz ships an `sz`) — and is never overwritten.
const (
	Absent  = "absent"
	Linked  = "linked"
	Foreign = "foreign"
	Unknown = "unknown"
)

// LinkPath is where the short link belongs for a given binary path, and
// empty when the binary itself could not be located — naming a bare `sz`
// relative to the working directory would be a path nobody meant.
func LinkPath(exe string) string {
	if exe == "" {
		return ""
	}
	return filepath.Join(filepath.Dir(exe), LinkName)
}

// LinkState reports what currently holds the short link's name, and what
// it points at when that is a symlink. An empty exe path (the binary could
// not be located) is Unknown: nothing can be said, and saying "absent"
// would invite a fix for a problem that may not exist.
func LinkState(exe string) (state, target string) {
	if exe == "" {
		return Unknown, ""
	}
	path := LinkPath(exe)
	// Readlink distinguishes all three cases on its own: it succeeds only
	// for a symlink, and fails differently for a real file than for a
	// missing one, which Lstat then tells apart.
	if target, err := os.Readlink(path); err == nil {
		if target == filepath.Base(exe) || target == exe {
			return Linked, target
		}
		return Foreign, target
	}
	if _, err := os.Lstat(path); err == nil {
		return Foreign, ""
	}
	return Absent, ""
}

// CreateLink makes the short link point at exe, and reports whether it had
// to create it. The target is relative (`sz -> screenz`), so moving or
// renaming the install directory keeps the pair intact. A name something
// else already holds is never taken: the error names the path and stops.
func CreateLink(exe string) (path string, created bool, err error) {
	path = LinkPath(exe)
	switch state, target := LinkState(exe); state {
	case Linked:
		return path, false, nil
	case Unknown:
		return path, false, fmt.Errorf("cannot locate the screenz binary to link beside")
	case Foreign:
		return path, false, fmt.Errorf("%s already exists%s; not replacing it", path, pointsAt(target))
	}
	if err := os.Symlink(filepath.Base(exe), path); err != nil {
		return path, false, err
	}
	return path, true, nil
}

func pointsAt(target string) string {
	if target == "" {
		return ""
	}
	return " and points at " + target
}

// Shells are the shells completion can be generated for, sorted. `all`
// selects every one of them.
var Shells = []string{"bash", "fish", "zsh"}

// scriptNames are the file names each shell looks for. zsh loads a
// function file named for the command it completes, off fpath; bash and
// fish source a plain script.
var scriptNames = map[string]string{
	"bash": "screenz.bash",
	"fish": "screenz.fish",
	"zsh":  "_screenz",
}

// Dir is where completion scripts are written: a `completions` directory
// beside `profiles` under the resolved screenz config directory (ADR-0015).
// screenz writes only inside the directory it owns and prints the one line
// to add to a shell's rc file, rather than writing into a directory that
// belongs to the shell (ADR-0029).
func Dir(configDir string) string {
	return filepath.Join(configDir, "completions")
}

// ScriptPath is the file one shell's completion script is written to.
func ScriptPath(dir, shell string) string {
	return filepath.Join(dir, scriptNames[shell])
}

// KnownShell reports whether completion can be generated for a shell.
func KnownShell(shell string) bool {
	_, ok := scriptNames[shell]
	return ok
}

// DetectShell reads the shell from a $SHELL value. A login shell screenz
// has no script for is an error naming the ones it has, never a silent
// fallback to bash: writing the wrong dialect looks installed and is not.
func DetectShell(shellEnv string) (string, error) {
	name := filepath.Base(shellEnv)
	if name == "" || name == "." || name == string(filepath.Separator) {
		return "", fmt.Errorf("cannot tell which shell to complete for ($SHELL is unset); name one with --shell (%s)", strings.Join(Shells, ", "))
	}
	if !KnownShell(name) {
		return "", fmt.Errorf("no completion script for %s; --shell takes %s or all", name, strings.Join(Shells, ", "))
	}
	return name, nil
}

// ResolveShells turns the --shell value into the shells to write for: the
// login shell when it is empty, every shell for `all`, otherwise the one
// named.
func ResolveShells(sel, shellEnv string) ([]string, error) {
	switch sel {
	case "":
		shell, err := DetectShell(shellEnv)
		if err != nil {
			return nil, err
		}
		return []string{shell}, nil
	case "all":
		return append([]string{}, Shells...), nil
	}
	if !KnownShell(sel) {
		return nil, fmt.Errorf("--shell %s: want %s or all", sel, strings.Join(Shells, ", "))
	}
	return []string{sel}, nil
}

// WriteScript writes one shell's completion script, creating the
// completions directory if it is not there yet.
func WriteScript(dir, shell, body string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := ScriptPath(dir, shell)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// HasScript reports whether one shell's completion script is already
// written. It says nothing about the script being current: an upgrade
// rewrites it, which is why `screenz update --all` always regenerates.
func HasScript(dir, shell string) bool {
	_, err := os.Stat(ScriptPath(dir, shell))
	return err == nil
}

// InstalledShells lists the shells that already have a script in dir,
// sorted, so doctor and update can report the install without regenerating
// anything.
func InstalledShells(dir string) []string {
	out := []string{}
	for _, shell := range Shells {
		if HasScript(dir, shell) {
			out = append(out, shell)
		}
	}
	sort.Strings(out)
	return out
}

// RCLine is the line to add to a shell's startup file so it finds the
// script. fish and bash source it directly; zsh needs the directory on
// fpath before compinit runs, which is why the script is a function file.
func RCLine(dir, shell string) string {
	switch shell {
	case "zsh":
		return fmt.Sprintf("fpath=(%s $fpath)   # in ~/.zshrc, before compinit", dir)
	case "bash":
		return fmt.Sprintf("source %s   # in ~/.bashrc", ScriptPath(dir, shell))
	}
	return fmt.Sprintf("source %s   # in ~/.config/fish/config.fish", ScriptPath(dir, shell))
}
