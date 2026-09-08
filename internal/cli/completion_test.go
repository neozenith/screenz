package cli

import (
	"flag"
	"sort"
	"strings"
	"testing"
)

// The gate that makes generated completions trustworthy (ADR-0030): every
// flag the parser registers is described in the table, and every flag the
// table describes is one the parser registers. Rename a flag and this
// fails in the same commit, instead of leaving a script that offers a name
// screenz no longer accepts.
func TestSpecMatchesTheParser(t *testing.T) {
	for _, c := range commandTable() {
		t.Run(c.name, func(t *testing.T) {
			fs := newFlagSet(c.name)
			if c.register != nil {
				c.register(fs)
			}
			var registered []string
			fs.VisitAll(func(f *flag.Flag) { registered = append(registered, f.Name) })
			var described []string
			for _, s := range c.flags {
				described = append(described, s.long)
				if s.short != "" {
					described = append(described, s.short)
				}
			}
			sort.Strings(registered)
			sort.Strings(described)
			if strings.Join(registered, " ") != strings.Join(described, " ") {
				t.Errorf("flags drift\n parser: %v\n  table: %v", registered, described)
			}
		})
	}
}

// ADR-0024: the initial is the command's first letter, and no two commands
// share one, so `screenz a` can only ever mean apply.
func TestEveryCommandAnswersToItsInitial(t *testing.T) {
	seen := map[string]string{}
	for _, c := range commandTable() {
		if c.initial != c.name[:1] {
			t.Errorf("%s: initial %q is not its first letter", c.name, c.initial)
		}
		if other, dup := seen[c.initial]; dup {
			t.Errorf("%s and %s both answer to %q", other, c.name, c.initial)
		}
		seen[c.initial] = c.name
		if got := lookup(c.initial); got == nil || got.name != c.name {
			t.Errorf("lookup(%q) did not resolve to %s", c.initial, c.name)
		}
		// zsh and fish read these two characters as syntax inside a
		// completion description.
		if strings.ContainsAny(c.summary, ":'") {
			t.Errorf("%s: summary must not contain ':' or an apostrophe: %q", c.name, c.summary)
		}
	}
	if lookup("nonesuch") != nil {
		t.Error("lookup resolved a command that does not exist")
	}
}

// Help is rendered from the same table, so a command cannot be listed
// without being dispatchable.
func TestUsageListsEveryCommand(t *testing.T) {
	_, out, _ := run(t, []string{"--help"}, deps(officeSys(true)))
	for _, c := range commandTable() {
		if !strings.Contains(out, c.initial+", "+c.name) || !strings.Contains(out, c.summary) {
			t.Errorf("usage does not list %s\n%s", c.name, out)
		}
	}
}

func TestCompletionScriptsCoverTheWholeCLI(t *testing.T) {
	scripts := map[string]string{
		"zsh":  CompletionScript("zsh"),
		"bash": CompletionScript("bash"),
		"fish": CompletionScript("fish"),
	}
	for shell, script := range scripts {
		for _, c := range commandTable() {
			if !strings.Contains(script, c.name) {
				t.Errorf("%s: command %s missing", shell, c.name)
			}
		}
		// Flags the parser registers, and the values that come from the
		// domain rather than from a hand-written list. Only the long names
		// are checked across all three: fish spells a flag `-l region -s r`
		// where zsh and bash spell it `--region -r`.
		for _, want := range []string{"save-profile", "dry-run", "tolerance",
			"completions", "shell", "centre-third", "l23", "existing", "sz"} {
			if !strings.Contains(script, want) {
				t.Errorf("%s: %q missing", shell, want)
			}
		}
		// Profile names come from screenz itself through the embedded jq,
		// so the script needs no jq on PATH (ADR-0027).
		if !strings.Contains(script, `screenz list --jq ".profiles[].name" --raw`) {
			t.Errorf("%s: profile names are not completed dynamically", shell)
		}
	}
	// The one-letter aliases are offered wherever the parser accepts them
	// (ADR-0021).
	for _, shell := range []string{"zsh", "bash"} {
		for _, want := range []string{"--dry-run -n", "--region -r", "--first -f"} {
			if !strings.Contains(scripts[shell], want) {
				t.Errorf("%s: %q missing", shell, want)
			}
		}
	}
	if !strings.HasPrefix(scripts["zsh"], "#compdef screenz sz\n") {
		t.Error("the zsh script must complete both names")
	}
	if !strings.Contains(scripts["bash"], "complete -F _screenz screenz sz") {
		t.Error("the bash script must complete both names")
	}
	if !strings.Contains(scripts["fish"], "for __screenz_bin in screenz sz") {
		t.Error("the fish script must complete both names")
	}
	// A value flag offers no candidates rather than offering flag names.
	if !strings.Contains(scripts["zsh"], "--display|--gap|--jq|--match|--tolerance|-d|-g|-m|-t) return ;;") {
		t.Errorf("zsh script does not stop on opaque values:\n%s", scripts["zsh"])
	}
	// fish marks a switch as a switch and a value flag as exclusive.
	if !strings.Contains(scripts["fish"], "-l dry-run -s n -d") {
		t.Error("fish script treats --dry-run as taking a value")
	}
	if !strings.Contains(scripts["fish"], "-l jq -x -d") {
		t.Error("fish script treats --jq as a switch")
	}
}
