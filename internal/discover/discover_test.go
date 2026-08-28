package discover

import (
	"reflect"
	"testing"

	"github.com/joshpeak/screenz/internal/mac"
)

// Raw values recorded by the 2026-08-22 spike on the real office setup:
// a built-in Retina at {0,0,1512,982} and two 4K panels that share a serial
// number and differ only by UUID. NSScreen origins are bottom-left; the
// externals sit at y=-98 because they are 1080 pt tall against a 982 pt
// primary.
func spikeRaw() ([]mac.DisplayRaw, []mac.ScreenRaw, float64) {
	displays := []mac.DisplayRaw{
		{ID: 2, Serial: 1129796439, UUID: "9A0E5FDD-2222-4444-8888-000000000002", PixelW: 3840, PixelH: 2160,
			Bounds: mac.CGRect{Origin: mac.CGPoint{X: 1512, Y: 0}, Size: mac.CGSize{W: 1920, H: 1080}}},
		{ID: 1, Main: true, BuiltIn: true, UUID: "37D8832A-1111-4444-8888-000000000001", PixelW: 3024, PixelH: 1964,
			Bounds: mac.CGRect{Origin: mac.CGPoint{X: 0, Y: 0}, Size: mac.CGSize{W: 1512, H: 982}}},
		{ID: 3, Serial: 1129796439, UUID: "B44FA153-3333-4444-8888-000000000003", PixelW: 3840, PixelH: 2160,
			Bounds: mac.CGRect{Origin: mac.CGPoint{X: 3432, Y: 0}, Size: mac.CGSize{W: 1920, H: 1080}}},
	}
	screens := []mac.ScreenRaw{
		{ID: 1, Name: "Built-in Retina Display", Scale: 2,
			Frame:        mac.CGRect{Origin: mac.CGPoint{X: 0, Y: 0}, Size: mac.CGSize{W: 1512, H: 982}},
			VisibleFrame: mac.CGRect{Origin: mac.CGPoint{X: 59, Y: 0}, Size: mac.CGSize{W: 1453, H: 949}}},
		{ID: 2, Name: "LU28R55 (1)", Scale: 2,
			Frame:        mac.CGRect{Origin: mac.CGPoint{X: 1512, Y: -98}, Size: mac.CGSize{W: 1920, H: 1080}},
			VisibleFrame: mac.CGRect{Origin: mac.CGPoint{X: 1512, Y: -98}, Size: mac.CGSize{W: 1920, H: 1055}}},
		{ID: 3, Name: "LU28R55 (2)", Scale: 2,
			Frame:        mac.CGRect{Origin: mac.CGPoint{X: 3432, Y: -98}, Size: mac.CGSize{W: 1920, H: 1080}},
			VisibleFrame: mac.CGRect{Origin: mac.CGPoint{X: 3432, Y: -98}, Size: mac.CGSize{W: 1920, H: 1055}}},
	}
	return displays, screens, 982
}

func TestBuildDisplaysJoinsAndOrders(t *testing.T) {
	displays, screens, primaryH := spikeRaw()
	got := BuildDisplays(displays, screens, primaryH, nil)
	if len(got) != 3 {
		t.Fatalf("got %d displays", len(got))
	}
	// y-then-minX (ADR2.3): all three tops sit at y=0, so order is by x.
	wantNames := []string{"Built-in Retina Display", "LU28R55 (1)", "LU28R55 (2)"}
	for i, d := range got {
		if d.Index != i+1 {
			t.Errorf("display %d: Index = %d, want %d", i, d.Index, i+1)
		}
		if d.Name != wantNames[i] {
			t.Errorf("display %d: Name = %q, want %q", i, d.Name, wantNames[i])
		}
		if d.UUID == "" {
			t.Errorf("display %d: empty UUID", i)
		}
	}
	// The NSScreen join flips the visible frame into AX coordinates: the
	// built-in's 33 pt menu bar lands at the top in y-down coordinates.
	builtin := got[0]
	wantVis := mac.CGRect{Origin: mac.CGPoint{X: 59, Y: 33}, Size: mac.CGSize{W: 1453, H: 949}}
	if builtin.VisibleFrame != wantVis {
		t.Errorf("builtin VisibleFrame = %+v, want %+v", builtin.VisibleFrame, wantVis)
	}
	if !builtin.BuiltIn || !builtin.Main || builtin.Scale != 2 || builtin.PixelW != 3024 {
		t.Errorf("builtin fields wrong: %+v", builtin)
	}
	// Identical panels share a serial (spike fact); UUID is the stable key.
	if got[1].Serial != got[2].Serial {
		t.Errorf("expected shared serial, got %d vs %d", got[1].Serial, got[2].Serial)
	}
	if got[1].UUID == got[2].UUID {
		t.Error("UUIDs must differ")
	}
}

// Real layer-24 menu bar rows captured on this machine (2026-08-28): macOS
// 26 reserved a 30 pt strip on both externals while NSScreen.visibleFrame
// reported the full 1080 pt height — the carve restores the truth.
func TestCarveMenuBarFromWindowServerRows(t *testing.T) {
	frame := mac.CGRect{Origin: mac.CGPoint{X: -1920, Y: -98}, Size: mac.CGSize{W: 1920, H: 1080}}
	visLying := mac.CGRect{Origin: mac.CGPoint{X: -1857, Y: -98}, Size: mac.CGSize{W: 1857, H: 1080}}
	rows := []mac.CGWindowRaw{
		{Layer: 24, OwnerName: "Window Server", Bounds: mac.CGRect{Origin: mac.CGPoint{X: -1920, Y: -98}, Size: mac.CGSize{W: 1920, H: 30}}},
		{Layer: 25, OwnerName: "Control Centre", Bounds: mac.CGRect{Origin: mac.CGPoint{X: -145, Y: -98}, Size: mac.CGSize{W: 145, H: 30}}},
		{Layer: 0, OwnerName: "Code", Bounds: frame},
	}
	got := carveMenuBar(visLying, frame, rows)
	want := mac.CGRect{Origin: mac.CGPoint{X: -1857, Y: -68}, Size: mac.CGSize{W: 1857, H: 1050}}
	if got != want {
		t.Fatalf("carve = %+v, want %+v", got, want)
	}
	// A visibleFrame that already excludes the strip (the built-in) is
	// untouched, as is one with no strip row at all.
	builtinFrame := mac.CGRect{Origin: mac.CGPoint{}, Size: mac.CGSize{W: 1512, H: 982}}
	builtinVis := mac.CGRect{Origin: mac.CGPoint{X: 59, Y: 33}, Size: mac.CGSize{W: 1453, H: 949}}
	builtinRows := []mac.CGWindowRaw{
		{Layer: 24, OwnerName: "Window Server", Bounds: mac.CGRect{Origin: mac.CGPoint{}, Size: mac.CGSize{W: 1512, H: 33}}},
	}
	if got := carveMenuBar(builtinVis, builtinFrame, builtinRows); got != builtinVis {
		t.Fatalf("built-in carve changed a correct visibleFrame: %+v", got)
	}
	if got := carveMenuBar(visLying, frame, nil); got != visLying {
		t.Fatalf("no rows must be a no-op: %+v", got)
	}
	// A partial-width or tall layer-24 row is not a menu bar.
	oddRows := []mac.CGWindowRaw{
		{Layer: 24, Bounds: mac.CGRect{Origin: mac.CGPoint{X: -1920, Y: -98}, Size: mac.CGSize{W: 500, H: 30}}},
		{Layer: 24, Bounds: mac.CGRect{Origin: mac.CGPoint{X: -1920, Y: -98}, Size: mac.CGSize{W: 1920, H: 400}}},
	}
	if got := carveMenuBar(visLying, frame, oddRows); got != visLying {
		t.Fatalf("odd rows must be ignored: %+v", got)
	}
}

func TestBuildDisplaysWithoutScreenJoin(t *testing.T) {
	displays := []mac.DisplayRaw{{ID: 7, UUID: "u7",
		Bounds: mac.CGRect{Origin: mac.CGPoint{X: 0, Y: 0}, Size: mac.CGSize{W: 100, H: 100}}}}
	got := BuildDisplays(displays, nil, 100, nil)
	if got[0].Name != "" || got[0].Scale != 0 {
		t.Errorf("unjoined display should keep zero screen fields: %+v", got[0])
	}
}

func TestBuildDisplaysOrdersByRowThenX(t *testing.T) {
	displays := []mac.DisplayRaw{
		{ID: 1, UUID: "a", Bounds: mac.CGRect{Origin: mac.CGPoint{X: 500, Y: 900}, Size: mac.CGSize{W: 100, H: 100}}},
		{ID: 2, UUID: "b", Bounds: mac.CGRect{Origin: mac.CGPoint{X: 900, Y: 0}, Size: mac.CGSize{W: 100, H: 100}}},
		{ID: 3, UUID: "c", Bounds: mac.CGRect{Origin: mac.CGPoint{X: 0, Y: 0}, Size: mac.CGSize{W: 100, H: 100}}},
	}
	got := BuildDisplays(displays, nil, 100, nil)
	order := []uint32{3, 2, 1} // top row left-to-right, then the lower row
	for i, want := range order {
		if got[i].ID != want {
			t.Errorf("index %d: ID = %d, want %d", i+1, got[i].ID, want)
		}
	}
}

func window(pos mac.CGPoint, size mac.CGSize, title string) mac.WindowRaw {
	return mac.WindowRaw{Title: title, Role: "AXWindow", Subrole: "AXStandardWindow", Pos: pos, Size: size}
}

func TestBuildResolvesWindows(t *testing.T) {
	displays, screens, primaryH := spikeRaw()
	code := mac.AppRaw{PID: 500, Bundle: "com.microsoft.VSCode", Name: "Code"}
	// Two identical VS Code frames but only one on-screen CG row: the
	// second window sits on another Space and must be reported, not hidden.
	frame := mac.CGRect{Origin: mac.CGPoint{X: 1600, Y: 40}, Size: mac.CGSize{W: 1200, H: 900}}
	raw := mac.SnapshotRaw{
		Displays: displays, Screens: screens, PrimaryH: primaryH,
		Apps: []mac.AppWindows{{
			App: code,
			Windows: []mac.WindowRaw{
				window(frame.Origin, frame.Size, "repo — file.go"),
				window(frame.Origin, frame.Size, "other-space — notes.md"),
				func() mac.WindowRaw {
					w := window(mac.CGPoint{X: 100, Y: 100}, mac.CGSize{W: 800, H: 600}, "minimized")
					w.Minimized = true
					return w
				}(),
			},
		}},
		CGWindows: []mac.CGWindowRaw{
			{Number: 9001, OwnerPID: 500, Layer: 0, Bounds: frame},
			{Number: 9002, OwnerPID: 500, Layer: 25, Bounds: frame}, // non-zero layer never matches
		},
	}
	snap := Build(raw)
	if len(snap.Windows) != 3 {
		t.Fatalf("got %d windows", len(snap.Windows))
	}
	w0, w1, w2 := snap.Windows[0], snap.Windows[1], snap.Windows[2]
	if w0.ID != 9001 || w0.State != StateNormal {
		t.Errorf("first window: ID=%d state=%s, want 9001 normal", w0.ID, w0.State)
	}
	if w0.DisplayIndex != 2 {
		t.Errorf("first window DisplayIndex = %d, want 2 (fully inside LU28R55 (1))", w0.DisplayIndex)
	}
	if w1.ID != 0 || w1.State != StateOffscreen {
		t.Errorf("second window: ID=%d state=%s, want 0 offscreen (CG row already claimed)", w1.ID, w1.State)
	}
	if w2.State != StateMinimized {
		t.Errorf("minimized window state = %s", w2.State)
	}
	if w2.DisplayIndex != 1 {
		t.Errorf("minimized window DisplayIndex = %d, want 1", w2.DisplayIndex)
	}
}

func TestBuildRecordsAppErrors(t *testing.T) {
	displays, screens, primaryH := spikeRaw()
	raw := mac.SnapshotRaw{
		Displays: displays, Screens: screens, PrimaryH: primaryH,
		Apps: []mac.AppWindows{{App: mac.AppRaw{PID: 88, Bundle: "com.hung.app", Name: "Hung"},
			Err: "AXWindows: cannot complete (app not responding, or Window Server) (AXError -25204)"}},
	}
	snap := Build(raw)
	if len(snap.Windows) != 0 || len(snap.AppErrs) != 1 {
		t.Fatalf("windows=%d errs=%d", len(snap.Windows), len(snap.AppErrs))
	}
	if snap.AppErrs[0].PID != 88 || snap.AppErrs[0].Bundle != "com.hung.app" {
		t.Errorf("AppErr = %+v", snap.AppErrs[0])
	}
	if el := snap.AppEl(88); el != (mac.AXElement{}) {
		t.Errorf("AppEl(88) = %+v, want zero element", el)
	}
}

func TestClassifyStates(t *testing.T) {
	base := window(mac.CGPoint{}, mac.CGSize{W: 10, H: 10}, "w")
	cases := []struct {
		name      string
		mutate    func(*mac.WindowRaw)
		appHidden bool
		onScreen  bool
		want      string
	}{
		{"minimized wins", func(w *mac.WindowRaw) { w.Minimized = true }, true, false, StateMinimized},
		{"hidden", func(w *mac.WindowRaw) {}, true, true, StateHidden},
		{"sheet by subrole", func(w *mac.WindowRaw) { w.Subrole = "AXSheet" }, false, true, StateSheet},
		{"sheet by role", func(w *mac.WindowRaw) { w.Role, w.Subrole = "AXSheet", "" }, false, true, StateSheet},
		{"system dialog", func(w *mac.WindowRaw) { w.Subrole = "AXSystemDialog" }, false, true, StateDialog},
		{"dialog", func(w *mac.WindowRaw) { w.Subrole = "AXDialog" }, false, true, StateDialog},
		{"offscreen", func(w *mac.WindowRaw) {}, false, false, StateOffscreen},
		{"normal", func(w *mac.WindowRaw) {}, false, true, StateNormal},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := base
			tc.mutate(&w)
			if got := classify(w, tc.appHidden, tc.onScreen); got != tc.want {
				t.Fatalf("classify = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestMatchCGRejectsWrongPIDAndFarFrames(t *testing.T) {
	rows := []mac.CGWindowRaw{
		{Number: 1, OwnerPID: 10, Layer: 0, Bounds: mac.CGRect{Origin: mac.CGPoint{X: 0, Y: 0}, Size: mac.CGSize{W: 100, H: 100}}},
		{Number: 2, OwnerPID: 20, Layer: 0, Bounds: mac.CGRect{Origin: mac.CGPoint{X: 500, Y: 0}, Size: mac.CGSize{W: 100, H: 100}}},
	}
	free := []bool{true, true}
	frame := mac.CGRect{Origin: mac.CGPoint{X: 0.5, Y: 0.5}, Size: mac.CGSize{W: 100, H: 100}}
	if got := matchCG(rows, free, 10, frame); got != 0 {
		t.Errorf("half-point offset should match: got %d", got)
	}
	if got := matchCG(rows, free, 30, frame); got != -1 {
		t.Errorf("unknown pid must not match: got %d", got)
	}
	far := mac.CGRect{Origin: mac.CGPoint{X: 300, Y: 0}, Size: mac.CGSize{W: 100, H: 100}}
	if got := matchCG(rows, free, 10, far); got != -1 {
		t.Errorf("far frame must not match: got %d", got)
	}
}

func TestAssignDisplayLargestIntersection(t *testing.T) {
	displays, screens, primaryH := spikeRaw()
	built := BuildDisplays(displays, screens, primaryH, nil)
	// Straddles the boundary at x=1512: 400 pt on the built-in, 800 pt on
	// LU28R55 (1) — the bigger half wins.
	straddle := mac.CGRect{Origin: mac.CGPoint{X: 1112, Y: 100}, Size: mac.CGSize{W: 1200, H: 500}}
	if got := assignDisplay(straddle, built); got != 2 {
		t.Errorf("assignDisplay(straddle) = %d, want 2", got)
	}
	nowhere := mac.CGRect{Origin: mac.CGPoint{X: 90000, Y: 0}, Size: mac.CGSize{W: 10, H: 10}}
	if got := assignDisplay(nowhere, built); got != 0 {
		t.Errorf("assignDisplay(nowhere) = %d, want 0", got)
	}
}

func TestGroupByBundle(t *testing.T) {
	windows := []Window{
		{Bundle: "com.microsoft.VSCode", App: "Code", Title: "one"},
		{Bundle: "com.google.Chrome", App: "Google Chrome", Title: "work"},
		{Bundle: "com.microsoft.VSCode", App: "Code", Title: "two"},
	}
	groups := GroupByBundle(windows)
	if len(groups) != 2 {
		t.Fatalf("got %d groups", len(groups))
	}
	if groups[0].Bundle != "com.google.Chrome" || groups[1].Bundle != "com.microsoft.VSCode" {
		t.Errorf("groups not sorted by bundle: %q, %q", groups[0].Bundle, groups[1].Bundle)
	}
	titles := []string{groups[1].Windows[0].Title, groups[1].Windows[1].Title}
	if !reflect.DeepEqual(titles, []string{"one", "two"}) {
		t.Errorf("AX order not preserved: %v", titles)
	}
}

func TestWindowElAccessor(t *testing.T) {
	var w Window
	if w.El() != (mac.AXElement{}) {
		t.Error("zero window should expose a zero element")
	}
}
