package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/joshpeak/screenz/internal/discover"
	"github.com/joshpeak/screenz/internal/profile"
	"github.com/joshpeak/screenz/internal/rule"
)

const profileHelp = `usage: screenz profile <command>

Manage named rule-set profiles (commented YAML, one per context).
Profiles live in $SCREENZ_HOME, $XDG_CONFIG_HOME/screenz or
~/.config/screenz, under profiles/<name>.yaml.

Commands:
  status [--json] [NAME]      Show profiles and whether their display
                              aliases resolve on this machine.
  init NAME [--force]         Write a commented template profile.
  save NAME [--force] FLAGS   Append rules built from the usual apply
                              flags; --force rewrites the profile with
                              only the given rules. Comments survive.

Apply a profile with: screenz apply NAME [extra rule flags]
`

func runProfile(args []string, stdout, stderr io.Writer, d Deps) int {
	sub := ""
	if len(args) > 0 {
		sub = args[0]
	}
	switch sub {
	case "status":
		return runProfileStatus(args[1:], stdout, stderr, d)
	case "init":
		return runProfileInit(args[1:], stdout, stderr, d)
	case "save":
		return runProfileSave(args[1:], stdout, stderr, d)
	case "-h", "--help", "help", "":
		if sub == "" {
			fmt.Fprint(stderr, profileHelp)
			return 2
		}
		fmt.Fprint(stdout, profileHelp)
		return 0
	default:
		fmt.Fprintf(stderr, "screenz profile: unknown command %q\n", sub)
		fmt.Fprint(stderr, profileHelp)
		return 2
	}
}

type aliasStatus struct {
	Alias    string `json:"alias"`
	Spec     string `json:"spec"`
	Resolves bool   `json:"resolves"`
	Display  string `json:"display,omitempty"`
	Problem  string `json:"problem,omitempty"`
}

type profileStatus struct {
	Name    string        `json:"name"`
	Path    string        `json:"path"`
	Rules   int           `json:"rules"`
	Err     string        `json:"err,omitempty"`
	Aliases []aliasStatus `json:"aliases,omitempty"`
}

func runProfileStatus(args []string, stdout, stderr io.Writer, d Deps) int {
	fs := flag.NewFlagSet("profile status", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	jsonOut := fs.Bool("json", false, "emit JSON")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(stdout, profileHelp)
			return 0
		}
		fmt.Fprintf(stderr, "screenz profile status: %v\n", err)
		return 2
	}
	if fs.NArg() > 1 {
		fmt.Fprintf(stderr, "screenz profile status: at most one NAME, got %v\n", fs.Args())
		return 2
	}

	displays, err := d.Displays()
	if err != nil {
		fmt.Fprintf(stderr, "screenz profile status: %v\n", err)
		return 1
	}

	dir := filepath.Join(profile.Dir(d.Getenv, d.Home), "profiles")
	var paths []string
	if fs.NArg() == 1 {
		paths = []string{filepath.Join(dir, fs.Arg(0)+".yaml")}
	} else {
		entries, err := os.ReadDir(dir)
		if err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(stderr, "screenz profile status: %v\n", err)
			return 1
		}
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".yaml") {
				paths = append(paths, filepath.Join(dir, e.Name()))
			}
		}
		sort.Strings(paths)
	}

	var statuses []profileStatus
	for _, path := range paths {
		st := profileStatus{Name: strings.TrimSuffix(filepath.Base(path), ".yaml"), Path: path}
		p, err := profile.Load(path)
		if err != nil {
			st.Err = err.Error()
			statuses = append(statuses, st)
			continue
		}
		st.Rules = len(p.Rules)
		for _, alias := range p.Aliases() {
			spec := p.Displays[alias]
			as := aliasStatus{Alias: alias, Spec: spec.String()}
			var matches []discover.Display
			for _, disp := range displays {
				if spec.Matches(disp) {
					matches = append(matches, disp)
				}
			}
			switch len(matches) {
			case 1:
				as.Resolves = true
				as.Display = fmt.Sprintf("index=%d %s", matches[0].Index, matches[0].Name)
			case 0:
				as.Problem = "no connected display matches"
			default:
				as.Problem = fmt.Sprintf("ambiguous: matches %d displays", len(matches))
			}
			st.Aliases = append(st.Aliases, as)
		}
		statuses = append(statuses, st)
	}

	if *jsonOut {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(struct {
			Schema   int             `json:"schema"`
			Profiles []profileStatus `json:"profiles"`
		}{1, statuses})
		return 0
	}
	if len(statuses) == 0 {
		fmt.Fprintf(stdout, "no profiles in %s\n", dir)
		return 0
	}
	tw := newTabwriter(stdout)
	fmt.Fprintln(tw, "PROFILE\tRULES\tALIAS\tRESOLVES\tDETAIL")
	for _, st := range statuses {
		if st.Err != "" {
			fmt.Fprintf(tw, "%s\t-\t-\t-\terror: %s\n", st.Name, st.Err)
			continue
		}
		if len(st.Aliases) == 0 {
			fmt.Fprintf(tw, "%s\t%d\t-\t-\t-\n", st.Name, st.Rules)
			continue
		}
		for _, as := range st.Aliases {
			detail := as.Display
			if as.Problem != "" {
				detail = as.Problem
			}
			fmt.Fprintf(tw, "%s\t%d\t%s\t%v\t%s\n", st.Name, st.Rules, as.Alias, as.Resolves, detail)
		}
	}
	tw.Flush()
	return 0
}

func runProfileInit(args []string, stdout, stderr io.Writer, d Deps) int {
	name := ""
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		name = args[0]
		args = args[1:]
	}
	fs := flag.NewFlagSet("profile init", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	force := fs.Bool("force", false, "overwrite an existing profile")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(stdout, profileHelp)
			return 0
		}
		fmt.Fprintf(stderr, "screenz profile init: %v\n", err)
		return 2
	}
	if name == "" || fs.NArg() > 0 {
		fmt.Fprintln(stderr, "screenz profile init: exactly one NAME, given first")
		return 2
	}
	path := profile.Path(d.Getenv, d.Home, name)
	if err := profile.WriteTemplate(path, name, *force); err != nil {
		fmt.Fprintf(stderr, "screenz profile init: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "wrote %s\n", path)
	return 0
}

func runProfileSave(args []string, stdout, stderr io.Writer, d Deps) int {
	fs := flag.NewFlagSet("profile save", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	force := fs.Bool("force", false, "rewrite the profile with only the given rules")
	rules := &rule.List{}
	rules.Register(fs)
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
			fmt.Fprint(stdout, profileHelp)
			return 0
		}
		fmt.Fprintln(stderr, "screenz profile save: NAME must come first")
		return 2
	}
	name := args[0]
	if err := fs.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(stdout, profileHelp)
			return 0
		}
		fmt.Fprintf(stderr, "screenz profile save: %v\n", err)
		return 2
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "screenz profile save: unexpected argument %q\n", fs.Arg(0))
		return 2
	}
	if len(rules.Rules) == 0 {
		fmt.Fprintln(stderr, "screenz profile save: no rules given (start one with --match)")
		return 2
	}
	if err := rules.Validate(); err != nil {
		fmt.Fprintf(stderr, "screenz profile save: %v\n", err)
		return 2
	}

	path := profile.Path(d.Getenv, d.Home, name)
	var err error
	if *force {
		err = profile.WriteNew(path, name, rules.Rules)
	} else {
		// Append reads the file itself; a missing profile falls through to
		// a fresh write instead of a separate exists pre-check.
		err = profile.Append(path, rules.Rules)
		if errors.Is(err, os.ErrNotExist) {
			err = profile.WriteNew(path, name, rules.Rules)
		}
	}
	if err != nil {
		fmt.Fprintf(stderr, "screenz profile save: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "saved %d rule(s) to %s\n", len(rules.Rules), path)
	return 0
}
