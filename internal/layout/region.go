// Package layout is pure rect arithmetic: named regions, grids and unit
// rects over a display's usable frame, plus the tolerance model used to
// verify placements (ADR3.1). Everything is in AX points (top-left origin).
package layout

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/neozenith/screenz/internal/mac"
)

// Region describes where windows land on a display's usable frame.
type Region struct {
	Name       string     // named region, mutually exclusive with the below
	Cols, Rows int        // grid=CxR: one cell per window in group order
	Unit       [4]float64 // unit=x,y,w,h: fractions of the usable frame
	kind       byte       // 'n' named, 'g' grid, 'u' unit
}

// span is a region expressed as fractions of the usable frame.
type span struct{ x1, y1, x2, y2 float64 }

var namedRegions = map[string]span{
	"maximize":         {0, 0, 1, 1},
	"left-half":        {0, 0, 0.5, 1},
	"right-half":       {0.5, 0, 1, 1},
	"top-half":         {0, 0, 1, 0.5},
	"bottom-half":      {0, 0.5, 1, 1},
	"first-third":      {0, 0, 1.0 / 3, 1},
	"center-third":     {1.0 / 3, 0, 2.0 / 3, 1},
	"last-third":       {2.0 / 3, 0, 1, 1},
	"first-two-thirds": {0, 0, 2.0 / 3, 1},
	"last-two-thirds":  {1.0 / 3, 0, 1, 1},
	"top-left":         {0, 0, 0.5, 0.5},
	"top-right":        {0.5, 0, 1, 0.5},
	"bottom-left":      {0, 0.5, 0.5, 1},
	"bottom-right":     {0.5, 0.5, 1, 1},
}

// ParseRegion parses a region literal: a name from the catalogue,
// grid=CxR, or unit=x,y,w,h.
func ParseRegion(s string) (Region, error) {
	if _, ok := namedRegions[s]; ok {
		return Region{Name: s, kind: 'n'}, nil
	}
	if rest, ok := strings.CutPrefix(s, "grid="); ok {
		c, r, found := strings.Cut(rest, "x")
		cols, err1 := strconv.Atoi(c)
		rows, err2 := strconv.Atoi(r)
		if !found || err1 != nil || err2 != nil || cols < 1 || rows < 1 {
			return Region{}, fmt.Errorf("region %q: grid wants grid=CxR with C,R >= 1", s)
		}
		return Region{Cols: cols, Rows: rows, kind: 'g'}, nil
	}
	if rest, ok := strings.CutPrefix(s, "unit="); ok {
		parts := strings.Split(rest, ",")
		if len(parts) != 4 {
			return Region{}, fmt.Errorf("region %q: unit wants unit=x,y,w,h fractions", s)
		}
		var u [4]float64
		for i, p := range parts {
			v, err := strconv.ParseFloat(p, 64)
			if err != nil || v < 0 || v > 1 {
				return Region{}, fmt.Errorf("region %q: unit fractions must be within 0..1", s)
			}
			u[i] = v
		}
		if u[0]+u[2] > 1 || u[1]+u[3] > 1 || u[2] == 0 || u[3] == 0 {
			return Region{}, fmt.Errorf("region %q: unit rect must fit inside the display", s)
		}
		return Region{Unit: u, kind: 'u'}, nil
	}
	return Region{}, fmt.Errorf("unknown region %q", s)
}

// String renders the region back to its literal (lossless for profile save).
func (r Region) String() string {
	switch r.kind {
	case 'g':
		return fmt.Sprintf("grid=%dx%d", r.Cols, r.Rows)
	case 'u':
		return fmt.Sprintf("unit=%s,%s,%s,%s", ftoa(r.Unit[0]), ftoa(r.Unit[1]), ftoa(r.Unit[2]), ftoa(r.Unit[3]))
	}
	return r.Name
}

func ftoa(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }

// Rect computes the target rect for window i of n in this region over a
// display's usable frame. Boundaries are floored like Rectangle's
// floor(v + 0.0001) at each fraction line, so halves sum exactly to the
// usable width and grid cells tile without overlap. gap insets the final
// rect on all sides.
func (r Region) Rect(usable mac.CGRect, i, n int, gap float64) mac.CGRect {
	s := r.span(i)
	x1 := edge(usable.Origin.X, usable.Size.W, s.x1)
	x2 := edge(usable.Origin.X, usable.Size.W, s.x2)
	y1 := edge(usable.Origin.Y, usable.Size.H, s.y1)
	y2 := edge(usable.Origin.Y, usable.Size.H, s.y2)
	out := mac.CGRect{Origin: mac.CGPoint{X: x1 + gap, Y: y1 + gap},
		Size: mac.CGSize{W: x2 - x1 - 2*gap, H: y2 - y1 - 2*gap}}
	if out.Size.W < 1 || out.Size.H < 1 {
		// A gap larger than the cell would invert the rect; drop the gap
		// rather than propose a degenerate frame.
		return mac.CGRect{Origin: mac.CGPoint{X: x1, Y: y1}, Size: mac.CGSize{W: x2 - x1, H: y2 - y1}}
	}
	return out
}

func (r Region) span(i int) span {
	switch r.kind {
	case 'g':
		cell := i % (r.Cols * r.Rows)
		col, row := cell%r.Cols, cell/r.Cols
		return span{
			float64(col) / float64(r.Cols), float64(row) / float64(r.Rows),
			float64(col+1) / float64(r.Cols), float64(row+1) / float64(r.Rows),
		}
	case 'u':
		return span{r.Unit[0], r.Unit[1], r.Unit[0] + r.Unit[2], r.Unit[1] + r.Unit[3]}
	}
	return namedRegions[r.Name]
}

// edge floors a fraction line onto whole points — AX rounds requested sizes
// to whole points (spike fact), so targets are pre-rounded to compare
// like-for-like on read back.
func edge(origin, size, fraction float64) float64 {
	return origin + math.Floor(size*fraction+0.0001)
}
