package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/joshpeak/screenz/internal/discover"
	"github.com/joshpeak/screenz/internal/mac"
	"github.com/joshpeak/screenz/internal/plan"
	"github.com/joshpeak/screenz/internal/rule"
)

const applyHelp = `usage: screenz apply [--dry-run] [--json] <rule flags...>

Apply placement rules to the live window set in one invocation. Every
--match opens a new rule; --display, --region, --gap, --tolerance, --first
and --order bind to the most recent rule (ADR4.1). A window is placed by
the FIRST rule it matches. Display selectors resolve against connected
displays before anything moves; zero or ambiguous matches abort the run.

Example (the office context switch):
  screenz apply \
    --match bundle=com.microsoft.VSCode        --display index=1 --region maximize \
    --match 'app="Google Chrome" title=/Work/' --display index=2 --region left-half \
    --match bundle=com.microsoft.edgemac       --display index=2 --region right-half

Rule flags:
  --match TERMS      bundle=, app=, title= terms; values may be "quoted" or /regex/i
  --display TERMS    index=N, name=, uuid=, serial=N, built-in=BOOL, main=BOOL
  --region REGION    maximize, halves, thirds, quarters, grid=CxR, unit=x,y,w,h
  --gap N            points between window and region edge
  --tolerance T      verification width: points (default 0.5) or N%
  --first            place only the first matching window
  --order ORDER      existing (default), title, or pid

Flags:
  --dry-run    Print the plan without moving anything.
  --json       Emit machine-readable output.
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
	fs := flag.NewFlagSet("apply", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	dryRun := fs.Bool("dry-run", false, "print the plan only")
	jsonOut := fs.Bool("json", false, "emit JSON")
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
		fmt.Fprintf(stderr, "screenz apply: unexpected argument %q\n", fs.Arg(0))
		fmt.Fprintln(stderr, "run 'screenz apply --help' for usage")
		return 2
	}
	if len(rules.Rules) == 0 {
		fmt.Fprintln(stderr, "screenz apply: no rules given (start one with --match)")
		fmt.Fprintln(stderr, "run 'screenz apply --help' for usage")
		return 2
	}
	if err := rules.Validate(); err != nil {
		fmt.Fprintf(stderr, "screenz apply: %v\n", err)
		return 2
	}

	info := d.Sys(false)
	if !info.Trusted {
		fmt.Fprint(stderr, grantInstruction(info.HostAppName, info.HostAppBundle))
		return 1
	}
	snap, err := d.Snapshot()
	if err != nil {
		fmt.Fprintf(stderr, "screenz apply: %v\n", err)
		return 1
	}
	p, err := plan.Build(rules.Rules, snap)
	if err != nil {
		fmt.Fprintf(stderr, "screenz apply: %v\n", err)
		return 1
	}

	if *dryRun {
		printPlan(p, snap, *jsonOut, stdout)
		return 0
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
		tw := tabwriter.NewWriter(stdout, 2, 4, 2, ' ', 0)
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
		fmt.Fprintf(stdout, "\n%d moved, %d skipped, %d windows matched no rule\n",
			len(results), len(p.Skipped), p.Unmatched)
	}

	if failed {
		return 1
	}
	return 0
}

// printPlan renders the dry-run plan; its JSON reuses the status shapes
// for displays and windows (G2/G4 contract).
func printPlan(p plan.Plan, snap discover.Snapshot, jsonOut bool, stdout io.Writer) {
	if jsonOut {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(struct {
			Schema   int                `json:"schema"`
			DryRun   bool               `json:"dry_run"`
			Displays []discover.Display `json:"displays"`
			Plan     plan.Plan          `json:"plan"`
		}{1, true, snap.Displays, p})
		return
	}
	tw := tabwriter.NewWriter(stdout, 2, 4, 2, ' ', 0)
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
