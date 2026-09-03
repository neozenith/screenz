package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/neozenith/screenz/internal/discover"
	"github.com/neozenith/screenz/internal/profile"
)

const listHelp = `usage: screenz list [NAME] [--json] [--verbose]

List the profiles resolved on this machine: each one's name, the file it
came from, and whether it fits the displays currently connected. A profile
fits when every display alias it declares resolves to exactly one connected
display; anything else is a cross, and applying it would exit 1 before a
window moved.

With NAME, report only that profile.

Flags:
  -v, --verbose   Show every alias and what it resolved to, not just the verdict.
  -j, --json      Emit {"schema":1,"profiles":[…]} JSON.
      --jq QUERY  Filter that JSON through a jq query, as piping it to jq
                  would. Implies --json. Object keys come out sorted, as
                  with jq -S.
      --raw       Print string results from --jq unquoted (jq's -r).
  -h, --help      Show this help.

Author a profile with: screenz init --profile NAME
                  or:  screenz apply FLAGS --save-profile NAME
`

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
	Fits    bool          `json:"fits"`
	Err     string        `json:"err,omitempty"`
	Aliases []aliasStatus `json:"aliases,omitempty"`
}

// fits reports whether the profile can be applied on the connected
// displays: it must have loaded, and every alias must resolve to exactly
// one display. A profile that declares no aliases fits trivially — it
// addresses displays inline, so there is nothing machine-specific to miss.
func (p profileStatus) fits() bool {
	if p.Err != "" {
		return false
	}
	for _, a := range p.Aliases {
		if !a.Resolves {
			return false
		}
	}
	return true
}

func runList(args []string, stdout, stderr io.Writer, d Deps) int {
	// stdlib flag stops parsing at the first non-flag argument, so a
	// leading NAME is peeled off before the flags — otherwise `list work
	// -v` reads -v as a second positional rather than a flag.
	only := ""
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		only, args = args[0], args[1:]
	}
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	jsonOut := fs.Bool("json", false, "emit JSON")
	aliasBool(fs, jsonOut, "j", "emit JSON")
	verbose := fs.Bool("verbose", false, "show every alias")
	aliasBool(fs, verbose, "v", "show every alias")
	jq := &jqOpts{}
	jq.register(fs)
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(stdout, listHelp)
			return 0
		}
		fmt.Fprintf(stderr, "screenz list: %v\n", err)
		return 2
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "screenz list: at most one NAME, got %q too\n", fs.Arg(0))
		return 2
	}
	filter, ok := jq.resolve("list", jsonOut, stderr)
	if !ok {
		return 2
	}

	displays, err := d.Displays()
	if err != nil {
		fmt.Fprintf(stderr, "screenz list: %v\n", err)
		return 1
	}

	dir := filepath.Join(profile.Dir(d.Getenv, d.Home), "profiles")
	var paths []string
	if only != "" {
		paths = []string{filepath.Join(dir, only+".yaml")}
	} else {
		entries, err := os.ReadDir(dir)
		if err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(stderr, "screenz list: %v\n", err)
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
			// A profile that will not load cannot fit; the verdict is
			// computed the same way for it as for any other, so there is
			// one place that decides what a tick means.
			st.Err = err.Error()
			st.Fits = st.fits()
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
				if explain := spec.Explain(displays); explain != "" {
					as.Problem += " (" + explain + ")"
				}
			default:
				as.Problem = fmt.Sprintf("ambiguous: matches %d displays", len(matches))
			}
			st.Aliases = append(st.Aliases, as)
		}
		st.Fits = st.fits()
		statuses = append(statuses, st)
	}

	if *jsonOut {
		return emitJSON(stdout, stderr, "list", struct {
			Schema   int             `json:"schema"`
			Profiles []profileStatus `json:"profiles"`
		}{1, statuses}, filter)
	}
	if len(statuses) == 0 {
		fmt.Fprintf(stdout, "no profiles in %s\n", dir)
		return 0
	}
	if *verbose {
		printProfilesVerbose(statuses, stdout)
		return 0
	}
	tw := newTabwriter(stdout)
	fmt.Fprintln(tw, "FITS\tPROFILE\tRULES\tPATH\tDETAIL")
	for _, st := range statuses {
		// Which aliases failed, not why: the why runs to a paragraph and
		// would push PATH off the screen. --verbose is where it lives.
		detail := ""
		if st.Err != "" {
			detail = "will not load"
		} else if bad := unresolvedAliases(st); len(bad) > 0 {
			detail = "unresolved: " + strings.Join(bad, ", ")
		}
		fmt.Fprintf(tw, "%s\t%s\t%d\t%s\t%s\n", tick(st.Fits), st.Name, st.Rules, st.Path, detail)
	}
	tw.Flush()
	// Reached only when not verbose: the verbose branch returned above.
	for _, st := range statuses {
		if !st.Fits {
			fmt.Fprintln(stdout, "\nrun 'screenz list --verbose' for why a profile does not fit")
			break
		}
	}
	return 0
}

func unresolvedAliases(st profileStatus) []string {
	var out []string
	for _, as := range st.Aliases {
		if !as.Resolves {
			out = append(out, as.Alias)
		}
	}
	return out
}

func printProfilesVerbose(statuses []profileStatus, stdout io.Writer) {
	tw := newTabwriter(stdout)
	fmt.Fprintln(tw, "FITS\tPROFILE\tRULES\tALIAS\tRESOLVES\tDETAIL")
	for _, st := range statuses {
		if st.Err != "" {
			fmt.Fprintf(tw, "%s\t%s\t-\t-\t-\terror: %s\n", tick(false), st.Name, st.Err)
			continue
		}
		if len(st.Aliases) == 0 {
			fmt.Fprintf(tw, "%s\t%s\t%d\t-\t-\tno aliases; displays addressed inline\n",
				tick(st.Fits), st.Name, st.Rules)
			continue
		}
		for _, as := range st.Aliases {
			detail := as.Display
			if as.Problem != "" {
				detail = as.Problem
			}
			fmt.Fprintf(tw, "%s\t%s\t%d\t%s\t%s\t%s\n",
				tick(st.Fits), st.Name, st.Rules, as.Alias, tick(as.Resolves), detail)
		}
	}
	tw.Flush()
}

// tick renders a verdict as a mark. Safe in a table because text/tabwriter
// measures cell width in runes, not bytes — these are 3 bytes each and
// would otherwise skew every column after them.
func tick(ok bool) string {
	if ok {
		return "✓"
	}
	return "✗"
}
