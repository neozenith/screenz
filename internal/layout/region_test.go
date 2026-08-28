package layout

import (
	"testing"

	"github.com/joshpeak/screenz/internal/mac"
)

// Real usable frames from the spike: the built-in Retina with a left Dock
// and 33 pt menu bar, and the first external panel.
var (
	builtinUsable = mac.CGRect{Origin: mac.CGPoint{X: 59, Y: 33}, Size: mac.CGSize{W: 1453, H: 949}}
	externalFrame = mac.CGRect{Origin: mac.CGPoint{X: 1512, Y: 25}, Size: mac.CGSize{W: 1920, H: 1055}}
)

func rect(x, y, w, h float64) mac.CGRect {
	return mac.CGRect{Origin: mac.CGPoint{X: x, Y: y}, Size: mac.CGSize{W: w, H: h}}
}

func mustRegion(t *testing.T, s string) Region {
	t.Helper()
	r, err := ParseRegion(s)
	if err != nil {
		t.Fatalf("ParseRegion(%q): %v", s, err)
	}
	return r
}

func TestNamedRegionsOnBuiltin(t *testing.T) {
	cases := []struct {
		region string
		want   mac.CGRect
	}{
		{"maximize", rect(59, 33, 1453, 949)},
		{"left-half", rect(59, 33, 726, 949)},
		{"right-half", rect(785, 33, 727, 949)}, // remainder: halves sum to 1453
		{"top-half", rect(59, 33, 1453, 474)},
		{"bottom-half", rect(59, 507, 1453, 475)},
		{"first-third", rect(59, 33, 484, 949)},
		{"center-third", rect(543, 33, 484, 949)},
		{"last-third", rect(1027, 33, 485, 949)},
		{"first-two-thirds", rect(59, 33, 968, 949)},
		{"last-two-thirds", rect(543, 33, 969, 949)},
		{"top-left", rect(59, 33, 726, 474)},
		{"top-right", rect(785, 33, 727, 474)},
		{"bottom-left", rect(59, 507, 726, 475)},
		{"bottom-right", rect(785, 507, 727, 475)},
	}
	for _, tc := range cases {
		t.Run(tc.region, func(t *testing.T) {
			got := mustRegion(t, tc.region).Rect(builtinUsable, 0, 1, 0)
			if got != tc.want {
				t.Fatalf("%s = %+v, want %+v", tc.region, got, tc.want)
			}
		})
	}
}

func TestHalvesAndThirdsTile(t *testing.T) {
	left := mustRegion(t, "left-half").Rect(builtinUsable, 0, 1, 0)
	right := mustRegion(t, "right-half").Rect(builtinUsable, 0, 1, 0)
	if left.Size.W+right.Size.W != builtinUsable.Size.W {
		t.Errorf("halves sum to %v, want %v", left.Size.W+right.Size.W, builtinUsable.Size.W)
	}
	if left.Origin.X+left.Size.W != right.Origin.X {
		t.Error("halves overlap or leave a seam")
	}
	var thirdsW float64
	prevRight := builtinUsable.Origin.X
	for _, name := range []string{"first-third", "center-third", "last-third"} {
		r := mustRegion(t, name).Rect(builtinUsable, 0, 1, 0)
		if r.Origin.X != prevRight {
			t.Errorf("%s starts at %v, want %v", name, r.Origin.X, prevRight)
		}
		prevRight = r.Origin.X + r.Size.W
		thirdsW += r.Size.W
	}
	if thirdsW != builtinUsable.Size.W {
		t.Errorf("thirds sum to %v, want %v", thirdsW, builtinUsable.Size.W)
	}
}

func TestGridTilesWithoutOverlap(t *testing.T) {
	g := mustRegion(t, "grid=3x2")
	var cells []mac.CGRect
	for i := 0; i < 6; i++ {
		cells = append(cells, g.Rect(externalFrame, i, 6, 0))
	}
	var area float64
	for i, a := range cells {
		area += a.Size.W * a.Size.H
		for j, b := range cells {
			if i == j {
				continue
			}
			if x := intersect(a, b); x > 0 {
				t.Errorf("cells %d and %d overlap by %v", i, j, x)
			}
		}
	}
	if want := externalFrame.Size.W * externalFrame.Size.H; area != want {
		t.Errorf("grid area %v, want the full usable %v", area, want)
	}
	// A seventh window wraps to the first cell (cell per window in group order).
	if got := g.Rect(externalFrame, 6, 7, 0); got != cells[0] {
		t.Errorf("wrap: window 6 got %+v, want cell 0 %+v", got, cells[0])
	}
}

func intersect(a, b mac.CGRect) float64 {
	x1 := max(a.Origin.X, b.Origin.X)
	y1 := max(a.Origin.Y, b.Origin.Y)
	x2 := min(a.Origin.X+a.Size.W, b.Origin.X+b.Size.W)
	y2 := min(a.Origin.Y+a.Size.H, b.Origin.Y+b.Size.H)
	if x2 <= x1 || y2 <= y1 {
		return 0
	}
	return (x2 - x1) * (y2 - y1)
}

func TestUnitRegion(t *testing.T) {
	got := mustRegion(t, "unit=0.25,0,0.5,1").Rect(externalFrame, 0, 1, 0)
	want := rect(1512+480, 25, 960, 1055)
	if got != want {
		t.Fatalf("unit = %+v, want %+v", got, want)
	}
}

func TestGapInsets(t *testing.T) {
	got := mustRegion(t, "left-half").Rect(builtinUsable, 0, 1, 8)
	want := rect(67, 41, 710, 933)
	if got != want {
		t.Fatalf("gap rect = %+v, want %+v", got, want)
	}
	// A gap larger than the cell is dropped instead of inverting the rect.
	tiny := mac.CGRect{Origin: mac.CGPoint{}, Size: mac.CGSize{W: 20, H: 20}}
	got = mustRegion(t, "left-half").Rect(tiny, 0, 1, 50)
	if got.Size.W <= 0 || got.Size.H <= 0 {
		t.Fatalf("degenerate rect: %+v", got)
	}
}

func TestParseRegionErrors(t *testing.T) {
	for _, bad := range []string{
		"maximise", "grid=0x2", "grid=3", "grid=ax2", "grid=2xb",
		"unit=0,0,1", "unit=0,0,2,1", "unit=a,0,1,1", "unit=0.6,0,0.6,1",
		"unit=0,0,0,1", "",
	} {
		if _, err := ParseRegion(bad); err == nil {
			t.Errorf("ParseRegion(%q) accepted", bad)
		}
	}
}

func TestRegionString(t *testing.T) {
	for _, s := range []string{"maximize", "grid=3x2", "unit=0.25,0,0.5,1"} {
		if got := mustRegion(t, s).String(); got != s {
			t.Errorf("String() = %q, want %q", got, s)
		}
	}
}

func TestParseTolerance(t *testing.T) {
	cases := []struct {
		in      string
		want    Tolerance
		wantErr bool
	}{
		{"", Tolerance{Value: 0.5}, false},
		{"0.5", Tolerance{Value: 0.5}, false},
		{"5%", Tolerance{Value: 5, Percent: true}, false},
		{"12", Tolerance{Value: 12}, false},
		{"-1", Tolerance{}, true},
		{"abc", Tolerance{}, true},
		{"%", Tolerance{}, true},
		// Inf/NaN parse as floats but would switch verification off.
		{"Inf", Tolerance{}, true},
		{"+Inf", Tolerance{}, true},
		{"Inf%", Tolerance{}, true},
		{"NaN", Tolerance{}, true},
	}
	for _, tc := range cases {
		got, err := ParseTolerance(tc.in)
		if (err != nil) != tc.wantErr {
			t.Errorf("ParseTolerance(%q) err = %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseTolerance(%q) = %+v, want %+v", tc.in, got, tc.want)
		}
	}
	if got := (Tolerance{Value: 5, Percent: true}).String(); got != "5%" {
		t.Errorf("String() = %q", got)
	}
	if got := (Tolerance{Value: 0.5}).String(); got != "0.5" {
		t.Errorf("String() = %q", got)
	}
}

func TestWithin(t *testing.T) {
	req := rect(59, 33, 726, 949)
	cases := []struct {
		name string
		act  mac.CGRect
		tol  Tolerance
		want bool
	}{
		{"exact", req, Tolerance{Value: 0.5}, true},
		{"half-point rounding", rect(59, 33, 726.5, 949), Tolerance{Value: 0.5}, true},
		{"clamped width", rect(59, 33, 700, 949), Tolerance{Value: 0.5}, false},
		{"shifted origin", rect(61, 33, 726, 949), Tolerance{Value: 0.5}, false},
		{"clamped height caught", rect(59, 33, 726, 900), Tolerance{Value: 0.5}, false},
		{"percent absorbs clamp", rect(59, 33, 700, 949), Tolerance{Value: 5, Percent: true}, true},
		{"percent still strict", rect(59, 33, 600, 949), Tolerance{Value: 5, Percent: true}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.tol.Within(req, tc.act); got != tc.want {
				t.Fatalf("Within = %v, want %v", got, tc.want)
			}
		})
	}
}
