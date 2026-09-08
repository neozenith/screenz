package cli

import (
	"flag"
	"io"

	"github.com/neozenith/screenz/internal/install"
	"github.com/neozenith/screenz/internal/layout"
	"github.com/neozenith/screenz/internal/rule"
)

// candidates are the words a shell offers for a flag value or for a
// command's positional argument. words is a fixed set known at build time;
// dyn is a command the shell runs to print one candidate per line, for a
// set that is whatever this machine has right now. A flag with neither
// takes a value no shell can guess, and completion stops rather than
// offering flag names where a value belongs.
type candidates struct {
	words []string
	dyn   string
}

func (c candidates) empty() bool { return len(c.words) == 0 && c.dyn == "" }

// profileNames completes profile names through screenz itself, filtered by
// the embedded jq (ADR-0027) so the script needs no jq on PATH. Double
// quotes, not single: the same string is pasted into zsh, bash and fish,
// and fish wraps a command substitution in single quotes.
var profileNames = candidates{dyn: `screenz list --jq ".profiles[].name" --raw 2>/dev/null`}

// flagSpec names one flag of one command: its long name, its one-letter
// alias if it has one (ADR-0021), and what completes after it. Whether the
// flag is boolean and what its help string says are read from the parser's
// own registration, not restated here.
type flagSpec struct {
	long  string
	short string
	cand  candidates
}

// command is one verb of the CLI: how it is spelled, what it does, the
// flags it parses and the handler it dispatches to. Dispatch, `screenz
// --help` and every generated completion script read this one table, so a
// command cannot exist in one of them and not the others (ADR-0030).
type command struct {
	name     string
	initial  string
	summary  string
	flags    []flagSpec
	args     candidates
	register func(*flag.FlagSet)
	run      func(args []string, stdout, stderr io.Writer, d Deps) int
}

// jqFlags are the two filtering flags every JSON-emitting command carries
// (ADR-0027). A jq query completes to nothing; --raw is boolean.
var jqFlags = []flagSpec{{long: "jq"}, {long: "raw"}}

// commandTable is the command set, in the order help prints it. Every
// initial is distinct, which is what makes it a second spelling of the
// command (ADR-0024); a summary must stay free of the colon and the
// apostrophe, which zsh and fish read as syntax inside a description.
//
// It is built per call rather than held in a package variable because a
// command handler reads it back — update generates completion scripts from
// it — and Go will not initialise a variable that refers to itself.
func commandTable() []command {
	return []command{
		{
			name: "apply", initial: "a",
			summary:  "Move groups of windows by rules or a profile, verifying every frame.",
			register: func(fs *flag.FlagSet) { registerApply(fs) },
			run:      runApply,
			flags: append([]flagSpec{
				{long: "profile", short: "p", cand: profileNames},
				{long: "save-profile", cand: profileNames},
				{long: "dry-run", short: "n"},
				{long: "json", short: "j"},
				{long: "match", short: "m"},
				{long: "display", short: "d"},
				{long: "region", short: "r", cand: candidates{words: layout.RegionWords()}},
				{long: "gap", short: "g"},
				{long: "tolerance", short: "t"},
				{long: "order", short: "o", cand: candidates{words: rule.OrderWords()}},
				{long: "first", short: "f"},
			}, jqFlags...),
		},
		{
			name: "status", initial: "s",
			summary:  "Show windows grouped by application and connected displays.",
			register: func(fs *flag.FlagSet) { registerStatus(fs) },
			run:      runStatus,
			args:     candidates{words: []string{"apps", "displays"}},
			flags: append([]flagSpec{
				{long: "match", short: "m"},
				{long: "verbose", short: "v"},
				{long: "json", short: "j"},
			}, jqFlags...),
		},
		{
			name: "list", initial: "l",
			summary:  "List profiles and whether each fits the connected displays.",
			register: func(fs *flag.FlagSet) { registerList(fs) },
			run:      runList,
			args:     profileNames,
			flags: append([]flagSpec{
				{long: "verbose", short: "v"},
				{long: "json", short: "j"},
			}, jqFlags...),
		},
		{
			name: "init", initial: "i",
			summary:  "Write a commented template profile to hand-edit.",
			register: func(fs *flag.FlagSet) { registerInit(fs) },
			run:      runInit,
			flags: []flagSpec{
				{long: "profile", short: "p", cand: profileNames},
				{long: "force"},
			},
		},
		{
			name: "doctor", initial: "d",
			summary:  "Check the Accessibility grant, displays, and symbol bindings.",
			register: func(fs *flag.FlagSet) { registerDoctor(fs) },
			run:      runDoctor,
			flags:    append([]flagSpec{{long: "json", short: "j"}}, jqFlags...),
		},
		{
			name: "update", initial: "u",
			summary:  "Self-update, and set up the sz short link and shell completions.",
			register: func(fs *flag.FlagSet) { registerUpdate(fs) },
			run:      runUpdate,
			flags: []flagSpec{
				{long: "check"},
				{long: "force"},
				{long: "all"},
				{long: "link"},
				{long: "completions"},
				{long: "shell", cand: candidates{words: append([]string{"all"}, install.Shells...)}},
			},
		},
		{
			name: "version", initial: "v",
			summary: "Print the release version (also --version).",
			run: func(_ []string, stdout, _ io.Writer, _ Deps) int {
				printVersion(stdout)
				return 0
			},
		},
	}
}

// lookup resolves a spelling to its command: the full name or its initial,
// and nothing else.
func lookup(name string) *command {
	table := commandTable()
	for i := range table {
		if name == table[i].name || name == table[i].initial {
			return &table[i]
		}
	}
	return nil
}

// newFlagSet builds the FlagSet every command parses with: errors are
// returned rather than printed, so each command formats its own usage.
func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}
