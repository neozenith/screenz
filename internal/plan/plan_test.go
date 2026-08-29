package plan

import (
	"flag"
	"io"
	"strings"
	"testing"

	"github.com/joshpeak/screenz/internal/discover"
	"github.com/joshpeak/screenz/internal/mac"
	"github.com/joshpeak/screenz/internal/rule"
)

func rect(x, y, w, h float64) mac.CGRect {
	return mac.CGRect{Origin: mac.CGPoint{X: x, Y: y}, Size: mac.CGSize{W: w, H: h}}
}

// The spike office: built-in plus two identically named panels.
func officeSnap() discover.Snapshot {
	return discover.Snapshot{
		Displays: []discover.Display{
			{Index: 1, Name: "Built-in Retina Display", UUID: "U1", BuiltIn: true, Main: true,
				Frame: rect(0, 0, 1512, 982), VisibleFrame: rect(59, 33, 1453, 949)},
			{Index: 2, Name: "LU28R55 (1)", UUID: "U2", Serial: 1129796439,
				Frame: rect(1512, 0, 1920, 1080), VisibleFrame: rect(1512, 25, 1920, 1055)},
			{Index: 3, Name: "LU28R55 (2)", UUID: "U3", Serial: 1129796439,
				Frame: rect(3432, 0, 1920, 1080), VisibleFrame: rect(3432, 25, 1920, 1055)},
		},
		Windows: []discover.Window{
			{ID: 1, PID: 500, Bundle: "com.microsoft.VSCode", App: "Code", Title: "beta — b.go", State: discover.StateNormal, DisplayIndex: 1, Frame: rect(100, 100, 800, 600)},
			{ID: 2, PID: 500, Bundle: "com.microsoft.VSCode", App: "Code", Title: "alpha — a.go", State: discover.StateNormal, DisplayIndex: 3, Frame: rect(3432, 100, 800, 600)},
			{ID: 3, PID: 500, Bundle: "com.microsoft.VSCode", App: "Code", Title: "gamma — g.go", State: discover.StateMinimized, DisplayIndex: 1, Frame: rect(0, 0, 800, 600)},
			{ID: 4, PID: 700, Bundle: "com.google.Chrome", App: "Google Chrome", Title: "Work — Inbox", State: discover.StateNormal, DisplayIndex: 1, Frame: rect(59, 33, 900, 900)},
			{ID: 5, PID: 700, Bundle: "com.google.Chrome", App: "Google Chrome", Title: "Personal — News", State: discover.StateNormal, DisplayIndex: 1, Frame: rect(200, 33, 900, 900)},
		},
	}
}

func rules(t *testing.T, args ...string) []*rule.Rule {
	t.Helper()
	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	l := &rule.List{}
	l.Register(fs)
	if err := fs.Parse(args); err != nil {
		t.Fatal(err)
	}
	if err := l.Validate(); err != nil {
		t.Fatal(err)
	}
	return l.Rules
}

func TestBuildOfficePlan(t *testing.T) {
	rs := rules(t,
		"--match", "bundle=com.microsoft.VSCode", "--display", "index=2", "--region", "maximize",
		"--match", `app="Google Chrome" title=/Work/`, "--display", "index=3", "--region", "left-half",
	)
	p, err := Build(rs, officeSnap())
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Actions) != 3 {
		t.Fatalf("got %d actions: %+v", len(p.Actions), p.Actions)
	}
	// Both normal VS Code windows to display 2 maximized.
	for _, a := range p.Actions[:2] {
		if a.Rule != 1 || a.To != 2 || a.Target != rect(1512, 25, 1920, 1055) {
			t.Errorf("VS Code action wrong: %+v", a)
		}
		if a.Change != "move" {
			t.Errorf("change = %q, want move", a.Change)
		}
	}
	// Work Chrome only; the Personal window matches no rule.
	chrome := p.Actions[2]
	if chrome.Rule != 2 || chrome.Window.ID != 4 || chrome.To != 3 {
		t.Errorf("chrome action wrong: %+v", chrome)
	}
	if chrome.Target != rect(3432, 25, 960, 1055) {
		t.Errorf("chrome target = %+v", chrome.Target)
	}
	// The minimized VS Code window is skipped with its reason (ADR2.2).
	if len(p.Skipped) != 1 || p.Skipped[0].Window.ID != 3 || p.Skipped[0].Reason != "minimized" {
		t.Errorf("skipped wrong: %+v", p.Skipped)
	}
	if p.Unmatched != 1 {
		t.Errorf("unmatched = %d, want 1 (Personal Chrome)", p.Unmatched)
	}
}

func TestFirstRuleWinsPerWindow(t *testing.T) {
	rs := rules(t,
		"--match", "title=/alpha/", "--display", "index=1", "--region", "left-half",
		"--match", "bundle=com.microsoft.VSCode", "--display", "index=2", "--region", "maximize",
	)
	p, err := Build(rs, officeSnap())
	if err != nil {
		t.Fatal(err)
	}
	byID := map[int64]Action{}
	for _, a := range p.Actions {
		byID[a.Window.ID] = a
	}
	if byID[2].Rule != 1 || byID[2].To != 1 {
		t.Errorf("alpha window should be claimed by rule 1: %+v", byID[2])
	}
	if byID[1].Rule != 2 {
		t.Errorf("beta window should fall to rule 2: %+v", byID[1])
	}
}

func TestGridAssignsCellsInOrder(t *testing.T) {
	rs := rules(t,
		"--match", "bundle=com.microsoft.VSCode", "--display", "index=2", "--region", "grid=2x1", "--order", "title",
	)
	p, err := Build(rs, officeSnap())
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Actions) != 2 {
		t.Fatalf("got %d actions", len(p.Actions))
	}
	// order=title: "alpha — a.go" gets cell 0 (left), "beta — b.go" cell 1.
	if p.Actions[0].Window.Title != "alpha — a.go" || p.Actions[0].Target != rect(1512, 25, 960, 1055) {
		t.Errorf("cell 0 wrong: %+v", p.Actions[0])
	}
	if p.Actions[1].Window.Title != "beta — b.go" || p.Actions[1].Target != rect(2472, 25, 960, 1055) {
		t.Errorf("cell 1 wrong: %+v", p.Actions[1])
	}
}

func TestOrderPidAndExisting(t *testing.T) {
	rs := rules(t, "--match", "bundle=com.microsoft.VSCode", "--display", "index=2", "--region", "grid=2x1", "--order", "pid")
	p, err := Build(rs, officeSnap())
	if err != nil {
		t.Fatal(err)
	}
	// Same pid: stable sort keeps AX order (beta first).
	if p.Actions[0].Window.Title != "beta — b.go" {
		t.Errorf("pid order changed AX order: %+v", p.Actions[0].Window.Title)
	}
	rs = rules(t, "--match", "bundle=com.microsoft.VSCode", "--display", "index=2", "--region", "grid=2x1")
	p, err = Build(rs, officeSnap())
	if err != nil {
		t.Fatal(err)
	}
	if p.Actions[0].Window.Title != "beta — b.go" {
		t.Errorf("existing order wrong: %+v", p.Actions[0].Window.Title)
	}
}

func TestFirstPlacesOneWindowOnly(t *testing.T) {
	rs := rules(t,
		"--match", "bundle=com.microsoft.VSCode", "--display", "index=2", "--region", "maximize", "--first", "--order", "title",
	)
	p, err := Build(rs, officeSnap())
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Actions) != 1 || p.Actions[0].Window.Title != "alpha — a.go" {
		t.Fatalf("first should place only alpha: %+v", p.Actions)
	}
	// With --first the un-placed windows stay free for later rules, and the
	// minimized one is not claimed as skipped.
	if len(p.Skipped) != 0 {
		t.Errorf("first must not claim skips: %+v", p.Skipped)
	}
	if p.Unmatched != 4 {
		t.Errorf("unmatched = %d, want 4", p.Unmatched)
	}
}

func TestChangeNoneWhenAlreadyPlaced(t *testing.T) {
	snap := officeSnap()
	snap.Windows = []discover.Window{
		{ID: 9, PID: 1, Bundle: "b", App: "A", Title: "t", State: discover.StateNormal, DisplayIndex: 2,
			Frame: rect(1512, 25, 1920, 1055)},
	}
	rs := rules(t, "--match", "bundle=b", "--display", "index=2", "--region", "maximize")
	p, err := Build(rs, snap)
	if err != nil {
		t.Fatal(err)
	}
	if p.Actions[0].Change != "none" {
		t.Errorf("change = %q, want none", p.Actions[0].Change)
	}
}

func TestDisplayResolutionFailsBeforeAnyMove(t *testing.T) {
	cases := []struct {
		name    string
		display string
		wantErr string
	}{
		{"zero matches", "index=9", "matches 0 of 3"},
		{"ambiguous", "name=/LU28R55/", "matches 2 of 3"},
		{"conflicting terms explained", `name="LU28R55 (1)" index=1`, "terms AND together"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rs := rules(t, "--match", "bundle=com.microsoft.VSCode", "--display", tc.display, "--region", "maximize")
			_, err := Build(rs, officeSnap())
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("err = %v, want %q", err, tc.wantErr)
			}
			if !strings.Contains(err.Error(), "rule 1") {
				t.Errorf("error must name the rule: %v", err)
			}
		})
	}
}

func TestUnresolvedAliasErrors(t *testing.T) {
	rs := rules(t, "--match", "bundle=b", "--display", "dell-right", "--region", "maximize")
	_, err := Build(rs, officeSnap())
	if err == nil || !strings.Contains(err.Error(), `alias "dell-right"`) {
		t.Fatalf("err = %v", err)
	}
}
