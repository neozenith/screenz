package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neozenith/screenz/internal/discover"
	"github.com/neozenith/screenz/internal/layout"
	"github.com/neozenith/screenz/internal/mac"
	"github.com/neozenith/screenz/internal/place"
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
		"1 moved, 0 failed, 1 skipped, 1 windows matched no rule",
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
	if !strings.Contains(out, "0 moved, 1 failed") {
		t.Errorf("summary must not count a failed placement as moved:\n%s", out)
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

// The --json output is the machine-readable contract automation uses to
// detect failed placements — assert the failure shapes, not just success.
func TestApplyJSONFailureShapes(t *testing.T) {
	var rep struct {
		Actions []struct {
			Result    string     `json:"result"`
			Err       string     `json:"err"`
			Requested mac.CGRect `json:"requested"`
			Actual    mac.CGRect `json:"actual"`
			Attempts  int        `json:"attempts"`
		} `json:"actions"`
	}
	code, out, _ := run(t, append([]string{"apply", "--json"}, vscodeRule...), applyDeps("clamped"))
	if code != 1 {
		t.Fatalf("clamped: exit = %d, want 1", code)
	}
	if err := json.Unmarshal([]byte(out), &rep); err != nil {
		t.Fatalf("clamped: not JSON: %v\n%s", err, out)
	}
	if rep.Actions[0].Result != "clamped" || rep.Actions[0].Attempts != 3 {
		t.Errorf("clamped action: %+v", rep.Actions[0])
	}
	if rep.Actions[0].Actual == rep.Actions[0].Requested {
		t.Error("clamped action must show the divergent actual frame")
	}

	code, out, _ = run(t, append([]string{"apply", "--json"}, vscodeRule...), applyDeps("error"))
	if code != 1 {
		t.Fatalf("error: exit = %d, want 1", code)
	}
	if err := json.Unmarshal([]byte(out), &rep); err != nil {
		t.Fatalf("error: not JSON: %v\n%s", err, out)
	}
	if rep.Actions[0].Result != "error" || !strings.Contains(rep.Actions[0].Err, "AXError -25204") {
		t.Errorf("error action: %+v", rep.Actions[0])
	}
}

// An app whose AX enumeration failed means matching ran against an
// incomplete world: apply must refuse, and the dry-run must warn.
func TestApplyRefusesOnAppErrs(t *testing.T) {
	snap := officeSnapshot()
	snap.AppErrs = []discover.AppErr{{PID: 88, App: "Hung", Bundle: "com.hung.app", Err: "AXWindows: cannot complete"}}
	d := applyDeps("ok")
	d.Snapshot = func() (discover.Snapshot, error) { return snap, nil }

	code, _, errOut := run(t, append([]string{"apply"}, vscodeRule...), d)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	for _, want := range []string{"cannot enumerate Hung (pid 88)", "incompletely enumerated"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("stderr missing %q:\n%s", want, errOut)
		}
	}

	// Dry-run still previews but warns and reports the app errors in JSON.
	code, out, errOut := run(t, append([]string{"apply", "--dry-run", "--json"}, vscodeRule...), d)
	if code != 0 {
		t.Fatalf("dry-run exit = %d, want 0", code)
	}
	if !strings.Contains(errOut, "cannot enumerate Hung") {
		t.Errorf("dry-run stderr missing warning:\n%s", errOut)
	}
	var plan struct {
		AppErrs []discover.AppErr `json:"app_errors"`
	}
	if err := json.Unmarshal([]byte(out), &plan); err != nil || len(plan.AppErrs) != 1 {
		t.Errorf("dry-run JSON missing app_errors: %v %s", err, out)
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

// -n and -j reach the same code paths as --dry-run and --json (ADR-0021),
// including the promise that a dry run places nothing.
func TestApplyShortFlags(t *testing.T) {
	d := applyDeps("ok")
	d.Place = func(_, _ mac.AXElement, _ mac.CGRect, _ layout.Tolerance) place.Result {
		t.Fatal("dry-run must not place windows")
		return place.Result{}
	}
	args := []string{"apply", "-n", "-j", "-m", "bundle=com.microsoft.VSCode", "-d", "index=2", "-r", "left-half"}
	code, out, _ := run(t, args, d)
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	if !strings.Contains(out, `"schema": 1`) || !strings.Contains(out, `"actions"`) {
		t.Errorf("short flags did not produce dry-run JSON:\n%s", out)
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
		// The old grammar took the profile positionally. It must say what
		// to type instead, never guess (ADR-0025).
		{"positional profile", []string{"apply", "office"}, "name a profile with --profile office"},
		{"no rules", []string{"apply"}, "no profile or rules given"},
		{"missing display", []string{"apply", "--match", "bundle=a", "--region", "maximize"}, "missing --display"},
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

// --profile names the profile; its rules run before any inline ones, and
// the two are one merged plan (ADR4.1, ADR-0025).
func TestApplyProfileFlagMergesWithInlineRules(t *testing.T) {
	d, home := profDeps(t)
	writeProfile(t, home, "office", fittingProfileYAML)

	code, out, errOut := run(t, []string{
		"apply", "--dry-run", "--profile", "office",
		"-m", "app=Slack", "-d", "index=2", "-r", "rh",
	}, d)
	if code != 0 {
		t.Fatalf("exit = %d; stderr: %s", code, errOut)
	}
	// The profile's VS Code rule is rule 1, the inline Slack rule is rule 2.
	if !strings.Contains(out, "Code") {
		t.Errorf("profile rule did not run:\n%s", out)
	}
	if code, _, _ := run(t, []string{"apply", "-p", "office", "--dry-run"}, d); code != 0 {
		t.Fatalf("-p short form: exit=%d", code)
	}
}

func TestApplyProfileFlagErrors(t *testing.T) {
	d, home := profDeps(t)
	if code, _, errOut := run(t, []string{"apply", "--profile", "missing"}, d); code != 1 || errOut == "" {
		t.Fatalf("missing profile: exit=%d err=%q", code, errOut)
	}
	// An alias the profile never declared fails before anything moves.
	writeProfile(t, home, "office", fittingProfileYAML+"  - match: {bundle: b}\n    display: nosuch\n    region: maximize\n")
	if code, _, errOut := run(t, []string{"apply", "--profile", "office"}, d); code != 1 ||
		!strings.Contains(errOut, `alias "nosuch" is not defined`) {
		t.Fatalf("undeclared alias: exit=%d err=%q", code, errOut)
	}
	// An alias that no connected display satisfies fails at plan time,
	// still before any window moves.
	writeProfile(t, home, "office", officeProfileYAML+"  - match: {bundle: b}\n    display: cinema\n    region: maximize\n")
	if code, _, errOut := run(t, []string{"apply", "--profile", "office"}, d); code != 1 ||
		!strings.Contains(errOut, "matches 0 of 2 connected displays") {
		t.Fatalf("unresolvable alias: exit=%d err=%q", code, errOut)
	}
}

// --save-profile creates the file when there is none.
func TestSaveProfileCreates(t *testing.T) {
	d, home := profDeps(t)
	code, out, errOut := run(t, []string{
		"apply", "--dry-run", "--save-profile", "office",
		"-m", "bundle=com.microsoft.VSCode", "-d", "index=2", "-r", "lh",
	}, d)
	if code != 0 || !strings.Contains(out, "saved 1 rule(s)") {
		t.Fatalf("exit=%d out=%q err=%q", code, out, errOut)
	}
	src, err := os.ReadFile(filepath.Join(home, "profiles", "office.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"region: left-half", "screenz apply --profile office"} {
		if !strings.Contains(string(src), want) {
			t.Errorf("saved profile missing %q:\n%s", want, src)
		}
	}
}

// Saving into an existing profile replaces its rules and keeps everything
// else — comments, and the displays map an alias rule depends on (ADR-0025).
func TestSaveProfileReplacesRulesAndKeepsTheFile(t *testing.T) {
	d, home := profDeps(t)
	writeProfile(t, home, "office", "# hand-written header\n"+fittingProfileYAML+
		"  - match: {bundle: com.google.Chrome}\n    display: laptop\n    region: right-half\n")

	code, out, errOut := run(t, []string{
		"apply", "--dry-run", "--save-profile", "office",
		"-m", "app=Slack", "-d", "index=1", "-r", "th",
	}, d)
	if code != 0 || !strings.Contains(out, "saved 1 rule(s)") {
		t.Fatalf("exit=%d out=%q err=%q", code, out, errOut)
	}
	src := string(mustRead(t, filepath.Join(home, "profiles", "office.yaml")))
	for _, want := range []string{
		"# hand-written header", // comments survive
		"laptop:",               // the displays map survives even though
		"built-in: true",        // no rule uses it any more
		"app: Slack",            // and the rules are exactly what was asked for
		"region: top-half",
	} {
		if !strings.Contains(src, want) {
			t.Errorf("saved profile missing %q:\n%s", want, src)
		}
	}
	for _, gone := range []string{"com.microsoft.VSCode", "com.google.Chrome", "right-half"} {
		if strings.Contains(src, gone) {
			t.Errorf("replaced rule %q is still there:\n%s", gone, src)
		}
	}
}

// Replaying a profile and saving it back must not bake today's display
// indexes into the file: the alias is the point of the profile.
func TestSaveProfileKeepsAliasesWhenReplayingOne(t *testing.T) {
	d, home := profDeps(t)
	writeProfile(t, home, "office", fittingProfileYAML)

	code, _, errOut := run(t, []string{
		"apply", "--dry-run", "--profile", "office", "--save-profile", "office",
		"-m", "app=Slack", "-d", "laptop",
	}, d)
	if code != 0 {
		t.Fatalf("exit = %d; stderr: %s", code, errOut)
	}
	src := string(mustRead(t, filepath.Join(home, "profiles", "office.yaml")))
	if !strings.Contains(src, "display: laptop") {
		t.Errorf("alias was resolved away on save:\n%s", src)
	}
	if strings.Contains(src, "built-in: true\n    region") || strings.Contains(src, "display:\n      built-in") {
		t.Errorf("a rule's display was flattened to a spec:\n%s", src)
	}
}

func TestSaveProfileErrors(t *testing.T) {
	d, home := profDeps(t)
	// A brand-new profile has no displays map, so an alias rule cannot be
	// saved into one — it would never resolve.
	code, _, errOut := run(t, []string{
		"apply", "--dry-run", "--save-profile", "fresh", "-m", "app=Slack", "-d", "laptop",
	}, d)
	if code != 1 || !strings.Contains(errOut, "display alias") {
		t.Fatalf("alias into new profile: exit=%d err=%q", code, errOut)
	}
	// A profile that will not parse is not silently overwritten.
	writeProfile(t, home, "broken", "version: 9\n")
	code, _, errOut = run(t, []string{
		"apply", "--dry-run", "--save-profile", "broken", "-m", "app=Slack", "-d", "index=1",
	}, d)
	if code != 1 || !strings.Contains(errOut, "--save-profile") {
		t.Fatalf("broken profile: exit=%d err=%q", code, errOut)
	}
}

// Saving an alias rule into a profile this run did not load: the save is
// what was asked for and succeeds, but the placement has no displays map
// to resolve the alias through, so the run still fails and says why.
func TestSaveProfileAliasWithoutLoadingItStillFailsTheApply(t *testing.T) {
	d, home := profDeps(t)
	writeProfile(t, home, "office", fittingProfileYAML)
	code, out, errOut := run(t, []string{
		"apply", "--dry-run", "--save-profile", "office", "-m", "app=Slack", "-d", "laptop",
	}, d)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(out, "saved 1 rule(s)") {
		t.Errorf("the save should still have happened:\n%s", out)
	}
	if !strings.Contains(errOut, "aliases come from a profile's displays map") {
		t.Errorf("stderr should name the cause:\n%s", errOut)
	}
	if !strings.Contains(string(mustRead(t, filepath.Join(home, "profiles", "office.yaml"))), "display: laptop") {
		t.Error("the alias rule was not written")
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
