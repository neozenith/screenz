package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/neozenith/screenz/internal/discover"
	"github.com/neozenith/screenz/internal/mac"
	"github.com/neozenith/screenz/internal/plan"
	"github.com/neozenith/screenz/internal/profile"
	"github.com/neozenith/screenz/internal/rule"
)

const applyHelp = `usage: screenz apply [PROFILE] [--dry-run] [--json] [rule flags...]

Apply placement rules to the live window set in one invocation. With a
PROFILE name, the profile's rules run first and any inline rule flags are
appended after them; display aliases resolve through the profile. Every
--match opens a new rule; --display, --region, --gap, --tolerance, --first
and --order bind to the most recent rule (ADR4.1). A rule needs a
--display; an omitted --region means maximize (ADR-0020). A window is
placed by the FIRST rule it matches. Display selectors resolve against
connected displays before anything moves; zero or ambiguous matches
abort the run.

Every flag below has the one-letter alias shown beside it (ADR-0021).
Each takes its own dash: all but --first carry a value, so there is
nothing to cluster.

Example (the office context switch):
  screenz apply \
    --match bundle=com.microsoft.VSCode        --display index=1 --region maximize \
    --match 'app="Google Chrome" title=/Work/' --display index=2 --region left-half \
    --match bundle=com.microsoft.edgemac       --display index=2 --region right-half

  ...or in short form, leaning on the maximize default:
  screenz apply -m bundle=com.microsoft.VSCode -d 1 \
                -m 'app="Google Chrome" title=/Work/' -d 2 -r left-half \
                -m bundle=com.microsoft.edgemac -d 2 -r right-half

Rule flags:
  -m, --match TERMS      bundle=, app=, title= terms; "quoted" or /regex/i values
  -d, --display TERMS    a bare index number (e.g. 2), a profile alias, or terms:
                         index=N, name=, uuid=, serial=N, built-in=BOOL, main=BOOL
  -r, --region REGION    default maximize (max). Every name has a code:
                           max  maximize          tl  top-left
                           lh   left-half         tr  top-right
                           rh   right-half        bl  bottom-left
                           th   top-half          br  bottom-right
                           bh   bottom-half
                           f3   first-third       f23 first-two-thirds
                           c3   center-third      l23 last-two-thirds
                           l3   last-third
                         ...plus grid=CxR and unit=x,y,w,h. maximise and
                         centre-third are accepted spellings; a profile
                         always stores the long name.
  -g, --gap N            points between window and region edge
  -t, --tolerance T      verification width: points (default 0.5) or N%
  -f, --first            place only the first matching window
  -o, --order ORDER      existing (default), title, or pid

Flags:
  -n, --dry-run    Print the plan without moving anything.
  -j, --json       Emit machine-readable output.
  -h, --help       Show this help.
`

// actionResult joins a planned action with its execution result. The
// result is "ok" only when every edge of the read-back frame is within the
// rule's tolerance (ADR3.1); a mismatch is "clamped" and fails the run.
type actionResult struct {
	plan.Action
	Requested mac.CGRect `json:"requested"`
	Before    mac.CGRect `json:"before"`
	Actual    mac.CGRect `json:"actual"`
	Attempts  int        `json:"attempts"`
	Result    string     `json:"result"`
	Err       string     `json:"err,omitempty"`
}

func runApply(args []string, stdout, stderr io.Writer, d Deps) int {
	name := ""
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		name = args[0]
		args = args[1:]
	}
	fs := flag.NewFlagSet("apply", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	dryRun := fs.Bool("dry-run", false, "print the plan only")
	aliasBool(fs, dryRun, "n", "print the plan only")
	jsonOut := fs.Bool("json", false, "emit JSON")
	aliasBool(fs, jsonOut, "j", "emit JSON")
	rules := &rule.List{}
	rules.Register(fs)
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(stdout, applyHelp)
			return 0
		}
		fmt.Fprintf(stderr, "screenz apply: %v\n", err)
		fmt.Fprintln(stderr, "run 'screenz apply --help' for usage")
		return 2
	}
	if fs.NArg() > 0 {
		if name != "" || fs.NArg() > 1 {
			fmt.Fprintf(stderr, "screenz apply: unexpected argument %q\n", fs.Arg(0))
			fmt.Fprintln(stderr, "run 'screenz apply --help' for usage")
			return 2
		}
		name = fs.Arg(0)
	}
	if err := rules.Validate(); err != nil {
		fmt.Fprintf(stderr, "screenz apply: %v\n", err)
		return 2
	}
	if name == "" && len(rules.Rules) == 0 {
		fmt.Fprintln(stderr, "screenz apply: no profile or rules given (start a rule with --match)")
		fmt.Fprintln(stderr, "run 'screenz apply --help' for usage")
		return 2
	}

	ruleSet := rules.Rules
	if name != "" {
		prof, err := profile.Load(profile.Path(d.Getenv, d.Home, name))
		if err != nil {
			fmt.Fprintf(stderr, "screenz apply: %v\n", err)
			return 1
		}
		// Profile rules first, inline rules appended after (ADR4.1); an
		// unresolved alias exits before any window moves.
		ruleSet, err = prof.Resolved(rules.Rules)
		if err != nil {
			fmt.Fprintf(stderr, "screenz apply: %v\n", err)
			return 1
		}
	}

	if !requireTrusted(d, stderr) {
		return 1
	}
	snap, err := d.Snapshot()
	if err != nil {
		fmt.Fprintf(stderr, "screenz apply: %v\n", err)
		return 1
	}
	// An application whose AX enumeration failed contributed no windows,
	// so matching ran against an incomplete world — "moved 9 of 10" must
	// never look like success (ADR2.2), and neither may "never saw 10".
	for _, e := range snap.AppErrs {
		fmt.Fprintf(stderr, "screenz apply: cannot enumerate %s (pid %d): %s\n", e.App, e.PID, e.Err)
	}
	p, err := plan.Build(ruleSet, snap)
	if err != nil {
		fmt.Fprintf(stderr, "screenz apply: %v\n", err)
		return 1
	}

	if *dryRun {
		printPlan(p, snap, *jsonOut, stdout)
		return 0
	}
	if len(snap.AppErrs) > 0 {
		fmt.Fprintln(stderr, "screenz apply: refusing to run against an incompletely enumerated window set")
		return 1
	}

	results := make([]actionResult, 0, len(p.Actions))
	failed := false
	for _, a := range p.Actions {
		res := d.Place(snap.AppEl(a.Window.PID), a.Window.El(), a.Target, a.Tolerance)
		ar := actionResult{Action: a, Requested: res.Requested, Before: res.Before,
			Actual: res.Actual, Attempts: res.Attempts, Err: res.Err}
		switch {
		case res.OK:
			ar.Result = "ok"
		case res.Err != "":
			ar.Result = "error"
			failed = true
		default:
			ar.Result = "clamped"
			failed = true
		}
		results = append(results, ar)
	}

	if *jsonOut {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(struct {
			Schema    int            `json:"schema"`
			Actions   []actionResult `json:"actions"`
			Skipped   []plan.Skipped `json:"skipped,omitempty"`
			Unmatched int            `json:"unmatched"`
		}{1, results, p.Skipped, p.Unmatched})
	} else {
		tw := newTabwriter(stdout)
		fmt.Fprintln(tw, "#\tRULE\tAPP\tTITLE\tWID\tFROM\tTO\tTARGET\tACTUAL\tRESULT")
		for i, r := range results {
			result := r.Result
			if r.Err != "" {
				result += " (" + r.Err + ")"
			}
			fmt.Fprintf(tw, "%d\t%d\t%s\t%s\t%d\t%d\t%d\t%s\t%s\t%s\n",
				i+1, r.Rule, r.Window.App, r.Window.Title, r.Window.ID, r.From, r.To,
				rectStr(r.Target), rectStr(r.Actual), result)
		}
		for _, s := range p.Skipped {
			fmt.Fprintf(tw, "-\t%d\t%s\t%s\t%d\t%d\t-\t-\t-\tskipped: %s\n",
				s.Rule, s.Window.App, s.Window.Title, s.Window.ID, s.Window.DisplayIndex, s.Reason)
		}
		tw.Flush()
		moved := 0
		for _, r := range results {
			if r.Result == "ok" {
				moved++
			}
		}
		fmt.Fprintf(stdout, "\n%d moved, %d failed, %d skipped, %d windows matched no rule\n",
			moved, len(results)-moved, len(p.Skipped), p.Unmatched)
	}

	if failed {
		return 1
	}
	return 0
}

// printPlan renders the dry-run plan; its JSON reuses the status shapes
// for displays and windows (the status/apply JSON contract).
func printPlan(p plan.Plan, snap discover.Snapshot, jsonOut bool, stdout io.Writer) {
	if jsonOut {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(struct {
			Schema   int                `json:"schema"`
			DryRun   bool               `json:"dry_run"`
			Displays []discover.Display `json:"displays"`
			Plan     plan.Plan          `json:"plan"`
			AppErrs  []discover.AppErr  `json:"app_errors,omitempty"`
		}{1, true, snap.Displays, p, snap.AppErrs})
		return
	}
	tw := newTabwriter(stdout)
	fmt.Fprintln(tw, "#\tRULE\tAPP\tTITLE\tWID\tFROM\tTO\tTARGET\tCHANGE")
	for i, a := range p.Actions {
		fmt.Fprintf(tw, "%d\t%d\t%s\t%s\t%d\t%d\t%d\t%s\t%s\n",
			i+1, a.Rule, a.Window.App, a.Window.Title, a.Window.ID, a.From, a.To,
			rectStr(a.Target), a.Change)
	}
	for _, s := range p.Skipped {
		fmt.Fprintf(tw, "-\t%d\t%s\t%s\t%d\t%d\t-\t-\tskipped: %s\n",
			s.Rule, s.Window.App, s.Window.Title, s.Window.ID, s.Window.DisplayIndex, s.Reason)
	}
	tw.Flush()
	fmt.Fprintf(stdout, "\n%d to move, %d skipped, %d windows matched no rule\n",
		len(p.Actions), len(p.Skipped), p.Unmatched)
}

func rectStr(r mac.CGRect) string {
	return fmt.Sprintf("%.0f,%.0f,%.0f,%.0f", r.Origin.X, r.Origin.Y, r.Size.W, r.Size.H)
}
