package cli

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/neozenith/screenz/internal/discover"
	"github.com/neozenith/screenz/internal/mac"
)

// officeSnapshot mirrors the real office: three displays and windows for
// VS Code (one on another Space) and two Chrome profiles.
func officeSnapshot() discover.Snapshot {
	return discover.Snapshot{
		Displays: []discover.Display{
			{Index: 1, ID: 1, UUID: "37D8832A-1111", Name: "Built-in Retina Display", Main: true, BuiltIn: true, Scale: 2,
				Frame:        mac.CGRect{Origin: mac.CGPoint{X: 0, Y: 0}, Size: mac.CGSize{W: 1512, H: 982}},
				VisibleFrame: mac.CGRect{Origin: mac.CGPoint{X: 59, Y: 33}, Size: mac.CGSize{W: 1453, H: 949}},
				PixelW:       3024, PixelH: 1964},
			{Index: 2, ID: 2, UUID: "9A0E5FDD-2222", Name: "LU28R55 (1)", Scale: 2,
				Frame:        mac.CGRect{Origin: mac.CGPoint{X: 1512, Y: 0}, Size: mac.CGSize{W: 1920, H: 1080}},
				VisibleFrame: mac.CGRect{Origin: mac.CGPoint{X: 1512, Y: 25}, Size: mac.CGSize{W: 1920, H: 1055}},
				PixelW:       3840, PixelH: 2160},
		},
		Windows: []discover.Window{
			{ID: 9001, PID: 500, Bundle: "com.microsoft.VSCode", App: "Code", Title: "repo — main.go", State: discover.StateNormal, DisplayIndex: 2,
				Frame: mac.CGRect{Origin: mac.CGPoint{X: 1512, Y: 25}, Size: mac.CGSize{W: 1920, H: 1055}}},
			{ID: 0, PID: 500, Bundle: "com.microsoft.VSCode", App: "Code", Title: "notes — todo.md", State: discover.StateOffscreen, DisplayIndex: 2,
				Frame: mac.CGRect{Origin: mac.CGPoint{X: 1512, Y: 25}, Size: mac.CGSize{W: 1920, H: 1055}}},
			{ID: 9010, PID: 700, Bundle: "com.google.Chrome", App: "Google Chrome", Title: "Work — Inbox", State: discover.StateNormal, DisplayIndex: 1,
				Frame: mac.CGRect{Origin: mac.CGPoint{X: 59, Y: 33}, Size: mac.CGSize{W: 1200, H: 900}}},
		},
	}
}

func snapDeps(snap discover.Snapshot, err error) Deps {
	d := deps(officeSys(true))
	d.Snapshot = func() (discover.Snapshot, error) { return snap, err }
	return d
}

func TestStatusTable(t *testing.T) {
	code, out, errOut := run(t, []string{"status"}, snapDeps(officeSnapshot(), nil))
	if code != 0 {
		t.Fatalf("exit = %d; stderr: %s", code, errOut)
	}
	for _, want := range []string{
		"APP", "BUNDLE", "WID", "TITLE", "DISPLAY", "STATE", "X,Y,W,H",
		"com.google.Chrome", "Work — Inbox",
		"com.microsoft.VSCode", "notes — todo.md", "offscreen",
		"1512,25,1920,1055",
		"INDEX", "NAME", "UUID", "PX", "PT", "VISIBLE", "MAIN", "BUILTIN",
		"Built-in Retina Display", "3024x1964", "1512x982", "59,33,1453,949",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout missing %q\n%s", want, out)
		}
	}
	// Chrome sorts before VS Code by bundle id.
	if strings.Index(out, "com.google.Chrome") > strings.Index(out, "com.microsoft.VSCode") {
		t.Error("groups not sorted by bundle id")
	}
}

func TestStatusJSON(t *testing.T) {
	code, out, _ := run(t, []string{"status", "--json"}, snapDeps(officeSnapshot(), nil))
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	var rep struct {
		Schema   int                `json:"schema"`
		Displays []discover.Display `json:"displays"`
		Windows  []discover.Window  `json:"windows"`
	}
	if err := json.Unmarshal([]byte(out), &rep); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, out)
	}
	if rep.Schema != 1 || len(rep.Displays) != 2 || len(rep.Windows) != 3 {
		t.Errorf("unexpected shape: schema=%d displays=%d windows=%d", rep.Schema, len(rep.Displays), len(rep.Windows))
	}
	if rep.Windows[0].Title != "repo — main.go" || rep.Windows[0].DisplayIndex != 2 {
		t.Errorf("window 0: %+v", rep.Windows[0])
	}
	if rep.Displays[0].UUID == "" || rep.Displays[0].VisibleFrame.Size.H >= rep.Displays[0].Frame.Size.H {
		t.Errorf("display 0 visible frame must be smaller than frame: %+v", rep.Displays[0])
	}
}

func TestStatusUntrustedExits1(t *testing.T) {
	d := snapDeps(officeSnapshot(), nil)
	d.Sys = officeSys(false)
	code, _, errOut := run(t, []string{"status"}, d)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(errOut, "Accessibility is NOT granted") {
		t.Errorf("stderr missing grant instruction:\n%s", errOut)
	}
}

// --match filters windows only (OR across repeats); displays stay complete.
func TestStatusMatchFilters(t *testing.T) {
	code, out, _ := run(t, []string{"status", "--match", "app=Code", "--match", `app="Google Chrome"`}, snapDeps(officeSnapshot(), nil))
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	for _, want := range []string{"repo — main.go", "notes — todo.md", "Work — Inbox", "LU28R55 (1)", "Built-in Retina Display"} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout missing %q\n%s", want, out)
		}
	}

	code, out, _ = run(t, []string{"status", "--json", "--match", "title=/Work/"}, snapDeps(officeSnapshot(), nil))
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	var rep struct {
		Displays []discover.Display `json:"displays"`
		Windows  []discover.Window  `json:"windows"`
	}
	if err := json.Unmarshal([]byte(out), &rep); err != nil {
		t.Fatal(err)
	}
	if len(rep.Windows) != 1 || rep.Windows[0].Title != "Work — Inbox" {
		t.Errorf("filtered windows wrong: %+v", rep.Windows)
	}
	if len(rep.Displays) != 2 {
		t.Errorf("displays must not be filtered: %d", len(rep.Displays))
	}

	if code, _, errOut := run(t, []string{"status", "--match", "color=red"}, snapDeps(officeSnapshot(), nil)); code != 2 || !strings.Contains(errOut, `selector key "color"`) {
		t.Fatalf("bad selector: exit=%d stderr=%q", code, errOut)
	}
}

func TestStatusMissingSymbolsIsNotAGrantProblem(t *testing.T) {
	d := snapDeps(officeSnapshot(), nil)
	d.Sys = func(bool) SysInfo { return SysInfo{MissingSymbols: []string{"CGDisplayBounds"}} }
	code, _, errOut := run(t, []string{"status"}, d)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(errOut, "cannot bind macOS symbols") || strings.Contains(errOut, "Accessibility is NOT granted") {
		t.Errorf("wrong diagnosis:\n%s", errOut)
	}
}

func TestStatusSnapshotError(t *testing.T) {
	code, _, errOut := run(t, []string{"status"}, snapDeps(discover.Snapshot{}, errors.New("cannot bind macOS symbols: X")))
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(errOut, "cannot bind macOS symbols: X") {
		t.Errorf("stderr missing cause:\n%s", errOut)
	}
}

func TestStatusAppErrWarnings(t *testing.T) {
	snap := officeSnapshot()
	snap.AppErrs = []discover.AppErr{{PID: 88, App: "Hung", Err: "AXWindows: cannot complete"}}
	code, _, errOut := run(t, []string{"status"}, snapDeps(snap, nil))
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	if !strings.Contains(errOut, "warning: Hung (pid 88): AXWindows: cannot complete") {
		t.Errorf("stderr missing warning:\n%s", errOut)
	}
}

func TestStatusHelpAndBadFlag(t *testing.T) {
	code, out, _ := run(t, []string{"status", "--help"}, snapDeps(officeSnapshot(), nil))
	if code != 0 || !strings.Contains(out, "usage: screenz status") {
		t.Fatalf("help: exit=%d out=%q", code, out)
	}
	code, out, errOut := run(t, []string{"status", "--nope"}, snapDeps(officeSnapshot(), nil))
	if code != 2 || out != "" || !strings.Contains(errOut, "usage: screenz status") {
		t.Fatalf("bad flag: exit=%d out=%q err=%q", code, out, errOut)
	}
}

// Titles are elided so the table holds its shape; --verbose prints them
// whole (ADR-0026).
func TestStatusElidesLongTitles(t *testing.T) {
	long := "a-very-long-window-title-that-runs-off-the-screen.go — screenz"
	snap := officeSnapshot()
	snap.Windows[0].Title = long

	code, out, _ := run(t, []string{"status"}, snapDeps(snap, nil))
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	if strings.Contains(out, long) {
		t.Errorf("long title was not elided:\n%s", out)
	}
	if !strings.Contains(out, "a-very-l...z — screenz"[:11]) || !strings.Contains(out, "...") {
		t.Errorf("elided form missing:\n%s", out)
	}

	for _, flag := range []string{"--verbose", "-v"} {
		code, out, _ := run(t, []string{"status", flag}, snapDeps(snap, nil))
		if code != 0 || !strings.Contains(out, long) {
			t.Errorf("%s did not print the full title: exit=%d\n%s", flag, code, out)
		}
	}

	// JSON is a machine contract: it is never elided, verbose or not.
	code, out, _ = run(t, []string{"status", "--json"}, snapDeps(snap, nil))
	if code != 0 || !strings.Contains(out, long) {
		t.Errorf("JSON elided a title: exit=%d\n%s", code, out)
	}
}

// A title at or below the elided width is left alone: swapping 19
// characters for 19 characters loses the ends and gains nothing.
func TestStatusShortTitlesAreUntouched(t *testing.T) {
	for _, title := range []string{"", "short", strings.Repeat("x", titleKeep*2+3)} {
		if got := elideTitle(title); got != title {
			t.Errorf("elideTitle(%q) = %q, want it unchanged", title, got)
		}
	}
	if got := elideTitle(strings.Repeat("x", titleKeep*2+4)); !strings.Contains(got, "...") {
		t.Errorf("one rune over the limit was not elided: %q", got)
	}
}

// Titles carry em dashes and non-Latin scripts; slicing bytes would print
// a replacement character mid-rune.
func TestStatusElidesOnRuneBoundaries(t *testing.T) {
	got := elideTitle("日本語のウィンドウタイトルはとても長いのです — screenz")
	if strings.Contains(got, "\uFFFD") {
		t.Errorf("elide split a rune: %q", got)
	}
	if n := len([]rune(got)); n != titleKeep*2+3 {
		t.Errorf("elided to %d runes, want %d: %q", n, titleKeep*2+3, got)
	}
}

func TestStatusSections(t *testing.T) {
	d := snapDeps(officeSnapshot(), nil)
	cases := []struct {
		section  string
		wantSeen string
		wantGone string
	}{
		{"apps", "com.microsoft.VSCode", "BUILTIN"},
		{"displays", "BUILTIN", "com.microsoft.VSCode"},
	}
	for _, tc := range cases {
		t.Run(tc.section, func(t *testing.T) {
			code, out, _ := run(t, []string{"status", tc.section}, d)
			if code != 0 {
				t.Fatalf("exit = %d", code)
			}
			if !strings.Contains(out, tc.wantSeen) {
				t.Errorf("stdout missing %q:\n%s", tc.wantSeen, out)
			}
			if strings.Contains(out, tc.wantGone) {
				t.Errorf("stdout leaked the other section (%q):\n%s", tc.wantGone, out)
			}
		})
	}
	// A section may still take flags after it.
	if code, out, _ := run(t, []string{"status", "apps", "-m", "app=Code"}, d); code != 0 ||
		strings.Contains(out, "Google Chrome") {
		t.Errorf("section + --match: exit=%d\n%s", code, out)
	}
	// A bare word that is not a section stays an argument error.
	if code, _, _ := run(t, []string{"status", "windows"}, d); code != 2 {
		t.Errorf("unknown section: exit=%d, want 2", code)
	}
}

// A section narrows the JSON exactly as it narrows the tables.
func TestStatusSectionJSON(t *testing.T) {
	d := snapDeps(officeSnapshot(), nil)
	var got statusJSON

	code, out, _ := run(t, []string{"status", "apps", "--json"}, d)
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("not JSON: %v", err)
	}
	if len(got.Windows) == 0 || got.Displays != nil {
		t.Errorf("apps JSON should carry windows only: %+v", got)
	}

	got = statusJSON{}
	code, out, _ = run(t, []string{"status", "displays", "--json"}, d)
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("not JSON: %v", err)
	}
	if len(got.Displays) == 0 || got.Windows != nil {
		t.Errorf("displays JSON should carry displays only: %+v", got)
	}
}

// An app that could not be enumerated is reported even when only the
// display section was asked for: an incomplete world is never silent.
func TestStatusSectionStillWarnsOnAppErrs(t *testing.T) {
	snap := officeSnapshot()
	snap.AppErrs = []discover.AppErr{{PID: 88, App: "Hung", Err: "AXWindows: cannot complete"}}
	code, _, errOut := run(t, []string{"status", "displays"}, snapDeps(snap, nil))
	if code != 0 || !strings.Contains(errOut, "Hung") {
		t.Errorf("exit=%d stderr=%q", code, errOut)
	}
}
