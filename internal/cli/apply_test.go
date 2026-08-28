package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/joshpeak/screenz/internal/discover"
	"github.com/joshpeak/screenz/internal/layout"
	"github.com/joshpeak/screenz/internal/mac"
	"github.com/joshpeak/screenz/internal/place"
)

func rect(x, y, w, h float64) mac.CGRect {
	return mac.CGRect{Origin: mac.CGPoint{X: x, Y: y}, Size: mac.CGSize{W: w, H: h}}
}

// placeOutcome replays a recorded placement outcome through the pipeline:
// "ok" (landed exactly), "clamped" (the app refused the size — the real
// Finder-minimum-size case), or "error" (AX set failed).
func placeOutcome(kind string) func(app, win mac.AXElement, target mac.CGRect, tol layout.Tolerance) place.Result {
	return func(_, _ mac.AXElement, target mac.CGRect, tol layout.Tolerance) place.Result {
		res := place.Result{Requested: target, Before: rect(100, 100, 800, 600), Attempts: 1}
		switch kind {
		case "ok":
			res.Actual, res.OK = target, true
		case "clamped":
			clamped := target
			clamped.Size.W = target.Size.W - 200 // AX accepted, app clamped
			res.Actual, res.Attempts = clamped, 3
		default:
			res.Err = "set position: cannot complete (app not responding, or Window Server) (AXError -25204)"
			res.Actual, res.Attempts = res.Before, 3
		}
		return res
	}
}

func applyDeps(outcome string) Deps {
	d := snapDeps(officeSnapshot(), nil)
	d.Place = placeOutcome(outcome)
	return d
}

var vscodeRule = []string{"--match", "bundle=com.microsoft.VSCode", "--display", "index=2", "--region", "left-half"}

func TestApplyExecutesAndReports(t *testing.T) {
	code, out, errOut := run(t, append([]string{"apply"}, vscodeRule...), applyDeps("ok"))
	if code != 0 {
		t.Fatalf("exit = %d; stderr: %s", code, errOut)
	}
	for _, want := range []string{
		"RULE", "APP", "TITLE", "WID", "FROM", "TO", "TARGET", "ACTUAL", "RESULT",
		"repo — main.go", "1512,25,960,1055", "ok",
		"skipped: offscreen", // the other-Space window is reported, not hidden
		"1 moved, 1 skipped, 1 windows matched no rule",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout missing %q\n%s", want, out)
		}
	}
}

func TestApplyClampedExits1(t *testing.T) {
	code, out, _ := run(t, append([]string{"apply"}, vscodeRule...), applyDeps("clamped"))
	if code != 1 {
		t.Fatalf("exit = %d, want 1 (a silently clamped window is the canonical false success)", code)
	}
	if !strings.Contains(out, "clamped") {
		t.Errorf("stdout missing clamped row:\n%s", out)
	}
}

func TestApplyAXErrorExits1(t *testing.T) {
	code, out, _ := run(t, append([]string{"apply"}, vscodeRule...), applyDeps("error"))
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(out, "error (set position:") {
		t.Errorf("stdout missing error detail:\n%s", out)
	}
}

func TestApplyJSON(t *testing.T) {
	code, out, _ := run(t, append([]string{"apply", "--json"}, vscodeRule...), applyDeps("ok"))
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	var rep struct {
		Schema  int `json:"schema"`
		Actions []struct {
			Window    discover.Window `json:"window"`
			Rule      int             `json:"rule"`
			Requested mac.CGRect      `json:"requested"`
			Actual    mac.CGRect      `json:"actual"`
			Result    string          `json:"result"`
			Attempts  int             `json:"attempts"`
		} `json:"actions"`
		Skipped   []struct{ Reason string } `json:"skipped"`
		Unmatched int                       `json:"unmatched"`
	}
	if err := json.Unmarshal([]byte(out), &rep); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, out)
	}
	if rep.Schema != 1 || len(rep.Actions) != 1 || rep.Actions[0].Result != "ok" {
		t.Errorf("unexpected report: %+v", rep)
	}
	if rep.Actions[0].Requested != rep.Actions[0].Actual {
		t.Errorf("ok action must have actual == requested: %+v", rep.Actions[0])
	}
	if len(rep.Skipped) != 1 || rep.Skipped[0].Reason != "offscreen" || rep.Unmatched != 1 {
		t.Errorf("skipped/unmatched wrong: %+v", rep)
	}
}

func TestApplyDryRunMovesNothing(t *testing.T) {
	d := applyDeps("ok")
	d.Place = func(_, _ mac.AXElement, _ mac.CGRect, _ layout.Tolerance) place.Result {
		t.Fatal("dry-run must not place windows")
		return place.Result{}
	}
	code, out, _ := run(t, append([]string{"apply", "--dry-run"}, vscodeRule...), d)
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	for _, want := range []string{
		"CHANGE", "move", "repo — main.go", "1512,25,960,1055",
		"skipped: offscreen", "1 to move, 1 skipped, 1 windows matched no rule",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout missing %q\n%s", want, out)
		}
	}
}

func TestApplyDryRunJSONSharesStatusShapes(t *testing.T) {
	code, out, _ := run(t, append([]string{"apply", "--dry-run", "--json"}, vscodeRule...), applyDeps("ok"))
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	var rep struct {
		Schema   int                `json:"schema"`
		DryRun   bool               `json:"dry_run"`
		Displays []discover.Display `json:"displays"`
		Plan     struct {
			Actions []struct {
				Window discover.Window `json:"window"`
				Target mac.CGRect      `json:"target"`
				Change string          `json:"change"`
			} `json:"actions"`
			Unmatched int `json:"unmatched"`
		} `json:"plan"`
	}
	if err := json.Unmarshal([]byte(out), &rep); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, out)
	}
	if !rep.DryRun || len(rep.Displays) != 2 || len(rep.Plan.Actions) != 1 {
		t.Errorf("unexpected plan: %+v", rep)
	}
	if rep.Plan.Actions[0].Window.Bundle != "com.microsoft.VSCode" {
		t.Errorf("plan window shape wrong: %+v", rep.Plan.Actions[0])
	}
}

func TestApplyUsageErrors(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"bad flag", []string{"apply", "--nope"}, "flag provided but not defined"},
		{"positional", []string{"apply", "office", "--match", "bundle=a"}, "unexpected argument"},
		{"no rules", []string{"apply"}, "no rules given"},
		{"missing region", []string{"apply", "--match", "bundle=a", "--display", "index=1"}, "missing --region"},
		{"sibling before match", []string{"apply", "--display", "index=1"}, "--display given before any --match"},
		{"bad selector", []string{"apply", "--match", "state=weird"}, `selector key "state"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, out, errOut := run(t, tc.args, applyDeps("ok"))
			if code != 2 {
				t.Fatalf("exit = %d, want 2", code)
			}
			if out != "" {
				t.Errorf("usage error must not write stdout: %q", out)
			}
			if !strings.Contains(errOut, tc.want) {
				t.Errorf("stderr missing %q:\n%s", tc.want, errOut)
			}
		})
	}
}

func TestApplyRuntimeFailures(t *testing.T) {
	t.Run("untrusted", func(t *testing.T) {
		d := applyDeps("ok")
		d.Sys = officeSys(false)
		code, _, errOut := run(t, append([]string{"apply"}, vscodeRule...), d)
		if code != 1 || !strings.Contains(errOut, "Accessibility is NOT granted") {
			t.Fatalf("exit=%d stderr=%s", code, errOut)
		}
	})
	t.Run("snapshot error", func(t *testing.T) {
		d := applyDeps("ok")
		d.Snapshot = func() (discover.Snapshot, error) { return discover.Snapshot{}, errStub("boom") }
		code, _, errOut := run(t, append([]string{"apply"}, vscodeRule...), d)
		if code != 1 || !strings.Contains(errOut, "boom") {
			t.Fatalf("exit=%d stderr=%s", code, errOut)
		}
	})
	t.Run("display resolution", func(t *testing.T) {
		code, _, errOut := run(t, []string{"apply", "--match", "bundle=a", "--display", "index=9", "--region", "maximize"}, applyDeps("ok"))
		if code != 1 || !strings.Contains(errOut, "matches 0 of 2") {
			t.Fatalf("exit=%d stderr=%s", code, errOut)
		}
	})
}

type errStub string

func (e errStub) Error() string { return string(e) }

func TestApplyHelp(t *testing.T) {
	code, out, _ := run(t, []string{"apply", "--help"}, applyDeps("ok"))
	if code != 0 || !strings.Contains(out, "usage: screenz apply") {
		t.Fatalf("help: exit=%d\n%s", code, out)
	}
}
