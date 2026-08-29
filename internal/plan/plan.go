// Package plan resolves rules against a discovery snapshot into concrete
// actions — pure, so `--dry-run` and the executor share one truth. Display
// resolution happens for every rule BEFORE any window moves: a profile
// written for a display that is not connected fails loudly instead of
// silently falling back to the main display.
package plan

import (
	"fmt"
	"sort"

	"github.com/joshpeak/screenz/internal/discover"
	"github.com/joshpeak/screenz/internal/layout"
	"github.com/joshpeak/screenz/internal/mac"
	"github.com/joshpeak/screenz/internal/rule"
)

// Action is one planned window placement. Rule is the 1-based rule number.
type Action struct {
	Window    discover.Window  `json:"window"`
	Rule      int              `json:"rule"`
	From      int              `json:"from"`
	To        int              `json:"to"`
	Target    mac.CGRect       `json:"target"`
	Change    string           `json:"change"` // move | none
	Tolerance layout.Tolerance `json:"-"`
}

// Skipped is a matched window the rule cannot act on (ADR2.2): it is
// claimed and reported with its state, never silently dropped.
type Skipped struct {
	Window discover.Window `json:"window"`
	Rule   int             `json:"rule"`
	Reason string          `json:"reason"`
}

// Plan is the full resolution of rules against the live window set.
type Plan struct {
	Actions   []Action  `json:"actions"`
	Skipped   []Skipped `json:"skipped,omitempty"`
	Unmatched int       `json:"unmatched"`
}

// Build resolves every display selector first (zero or ambiguous matches
// error out, naming the rule and selector, before anything moves), then
// claims windows: a window is placed by the FIRST rule it matches.
func Build(rules []*rule.Rule, snap discover.Snapshot) (Plan, error) {
	targets := make([]discover.Display, len(rules))
	for i, r := range rules {
		if r.Display.Alias != "" {
			return Plan{}, fmt.Errorf("rule %d: display alias %q is not defined (aliases come from a profile's displays map)", i+1, r.Display.Alias)
		}
		var matches []discover.Display
		for _, d := range snap.Displays {
			if r.Display.Matches(d) {
				matches = append(matches, d)
			}
		}
		if len(matches) != 1 {
			msg := fmt.Sprintf("rule %d: display %q matches %d of %d connected displays; exactly one is needed",
				i+1, r.Display, len(matches), len(snap.Displays))
			if len(matches) == 0 {
				if explain := r.Display.Explain(snap.Displays); explain != "" {
					msg += " (" + explain + ")"
				}
			}
			return Plan{}, fmt.Errorf("%s", msg)
		}
		targets[i] = matches[0]
	}

	p := Plan{}
	claimed := make([]bool, len(snap.Windows))
	for ri, r := range rules {
		var actionable []int
		for wi, w := range snap.Windows {
			if claimed[wi] || !r.Match.Matches(w) {
				continue
			}
			if w.State != discover.StateNormal {
				if !r.First {
					claimed[wi] = true
					p.Skipped = append(p.Skipped, Skipped{Window: w, Rule: ri + 1, Reason: w.State})
				}
				continue
			}
			actionable = append(actionable, wi)
		}
		orderWindows(actionable, snap.Windows, r.Order)
		if r.First && len(actionable) > 1 {
			actionable = actionable[:1]
		}
		for i, wi := range actionable {
			w := snap.Windows[wi]
			claimed[wi] = true
			target := r.Region.Rect(targets[ri].VisibleFrame, i, len(actionable), r.Gap)
			change := "move"
			if r.Tolerance.Within(target, w.Frame) {
				change = "none"
			}
			p.Actions = append(p.Actions, Action{
				Window: w, Rule: ri + 1, From: w.DisplayIndex, To: targets[ri].Index,
				Target: target, Change: change, Tolerance: r.Tolerance,
			})
		}
	}
	p.Unmatched = len(snap.Windows) - (len(p.Actions) + len(p.Skipped))
	return p, nil
}

// orderWindows sorts the rule's window indices: existing keeps AX order,
// title and pid sort stably by that field.
func orderWindows(idx []int, windows []discover.Window, order string) {
	switch order {
	case "title":
		sort.SliceStable(idx, func(a, b int) bool { return windows[idx[a]].Title < windows[idx[b]].Title })
	case "pid":
		sort.SliceStable(idx, func(a, b int) bool { return windows[idx[a]].PID < windows[idx[b]].PID })
	}
}
