package discover

import "sort"

// Group is every window of one application, keyed by bundle id (ADR2.1) —
// PID grouping splits multi-process apps, and profiles name bundle ids.
type Group struct {
	Bundle  string   `json:"bundle"`
	App     string   `json:"app"`
	Windows []Window `json:"windows"`
}

// GroupByBundle groups windows by bundle id, preserving AX order within a
// group. Groups are sorted by bundle id for stable output; the app name is
// a secondary display label.
func GroupByBundle(windows []Window) []Group {
	byBundle := map[string]*Group{}
	var order []string
	for _, w := range windows {
		g, ok := byBundle[w.Bundle]
		if !ok {
			g = &Group{Bundle: w.Bundle, App: w.App}
			byBundle[w.Bundle] = g
			order = append(order, w.Bundle)
		}
		g.Windows = append(g.Windows, w)
	}
	sort.Strings(order)
	out := make([]Group, 0, len(order))
	for _, b := range order {
		out = append(out, *byBundle[b])
	}
	return out
}
