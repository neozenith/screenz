package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/neozenith/screenz/internal/discover"
	"github.com/neozenith/screenz/internal/rule"
)

const statusHelp = `usage: screenz status [--json] [--match TERMS ...]

Show every on-screen window grouped by application (bundle id) and every
connected display with stable identity. Coordinates are AX global points
(origin at the top-left of the main display, y down).

Flags:
  -m, --match TERMS  Show only windows matching the selector (same grammar
                     as apply: bundle=, app=, title=; values may be "quoted"
                     or /regex/i). Repeat to OR several selectors, e.g.
                     -m app=Code -m 'app="Microsoft Edge"'.
                     Displays are always listed in full.
  -j, --json         Emit {"schema":1,"displays":[…],"windows":[…]} JSON.
  -h, --help         Show this help.
`

// matchList collects repeated --match selectors; a window is shown when ANY
// selector matches (OR), unlike apply where each --match opens its own rule.
type matchList struct{ sels []rule.Selector }

func (m *matchList) String() string { return "" }

func (m *matchList) Set(s string) error {
	sel, err := rule.ParseSelector(s)
	if err != nil {
		return err
	}
	m.sels = append(m.sels, sel)
	return nil
}

func (m *matchList) keep(w discover.Window) bool {
	if len(m.sels) == 0 {
		return true
	}
	for _, sel := range m.sels {
		if sel.Matches(w) {
			return true
		}
	}
	return false
}

// statusJSON is the machine-readable status shape; apply's plan JSON reuses
// the same displays/windows object shapes (the status/apply JSON contract).
type statusJSON struct {
	Schema   int                `json:"schema"`
	Displays []discover.Display `json:"displays"`
	Windows  []discover.Window  `json:"windows"`
	AppErrs  []discover.AppErr  `json:"app_errors,omitempty"`
}

func runStatus(args []string, stdout, stderr io.Writer, d Deps) int {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	jsonOut := fs.Bool("json", false, "emit JSON")
	aliasBool(fs, jsonOut, "j", "emit JSON")
	matches := &matchList{}
	fs.Var(matches, "match", "show only windows matching this selector (repeatable)")
	fs.Var(matches, "m", "show only windows matching this selector (repeatable)")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(stdout, statusHelp)
			return 0
		}
		fmt.Fprintf(stderr, "screenz status: %v\n", err)
		fmt.Fprint(stderr, statusHelp)
		return 2
	}

	if !requireTrusted(d, stderr) {
		return 1
	}
	snap, err := d.Snapshot()
	if err != nil {
		fmt.Fprintf(stderr, "screenz status: %v\n", err)
		return 1
	}
	kept := snap.Windows[:0:0]
	for _, w := range snap.Windows {
		if matches.keep(w) {
			kept = append(kept, w)
		}
	}
	snap.Windows = kept

	if *jsonOut {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(statusJSON{Schema: 1, Displays: snap.Displays, Windows: snap.Windows, AppErrs: snap.AppErrs})
		return 0
	}

	tw := newTabwriter(stdout)
	fmt.Fprintln(tw, "APP\tBUNDLE\tWID\tTITLE\tDISPLAY\tSTATE\tX,Y,W,H")
	for _, g := range discover.GroupByBundle(snap.Windows) {
		for _, w := range g.Windows {
			fmt.Fprintf(tw, "%s\t%s\t%d\t%s\t%d\t%s\t%.0f,%.0f,%.0f,%.0f\n",
				g.App, g.Bundle, w.ID, w.Title, w.DisplayIndex, w.State,
				w.Frame.Origin.X, w.Frame.Origin.Y, w.Frame.Size.W, w.Frame.Size.H)
		}
	}
	tw.Flush()

	fmt.Fprintln(stdout)
	tw = newTabwriter(stdout)
	fmt.Fprintln(tw, "INDEX\tNAME\tUUID\tPX\tPT\tVISIBLE\tMAIN\tBUILTIN")
	for _, disp := range snap.Displays {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%dx%d\t%.0fx%.0f\t%.0f,%.0f,%.0f,%.0f\t%v\t%v\n",
			disp.Index, disp.Name, disp.UUID, disp.PixelW, disp.PixelH,
			disp.Frame.Size.W, disp.Frame.Size.H,
			disp.VisibleFrame.Origin.X, disp.VisibleFrame.Origin.Y,
			disp.VisibleFrame.Size.W, disp.VisibleFrame.Size.H,
			disp.Main, disp.BuiltIn)
	}
	tw.Flush()

	for _, e := range snap.AppErrs {
		fmt.Fprintf(stderr, "warning: %s (pid %d): %s\n", e.App, e.PID, e.Err)
	}
	return 0
}
