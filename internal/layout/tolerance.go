package layout

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/joshpeak/screenz/internal/mac"
)

// DefaultTolerance only absorbs AX's whole-point rounding (ADR3.1).
const DefaultTolerance = 0.5

// Tolerance is the per-rule verification width: a bare number is points, a
// % suffix is percent of the target width or height for that axis (ADR3.1).
// A numeric tolerance widens the check deliberately; there is no boolean
// that switches verification off.
type Tolerance struct {
	Value   float64
	Percent bool
}

// ParseTolerance parses "0.5" or "5%"; empty returns the default.
func ParseTolerance(s string) (Tolerance, error) {
	if s == "" {
		return Tolerance{Value: DefaultTolerance}, nil
	}
	t := Tolerance{}
	num := s
	if rest, ok := strings.CutSuffix(s, "%"); ok {
		t.Percent = true
		num = rest
	}
	v, err := strconv.ParseFloat(num, 64)
	if err != nil || v < 0 {
		return Tolerance{}, fmt.Errorf("tolerance %q: want a non-negative number of points or N%%", s)
	}
	t.Value = v
	return t, nil
}

// String renders the tolerance back to its literal.
func (t Tolerance) String() string {
	if t.Percent {
		return ftoa(t.Value) + "%"
	}
	return ftoa(t.Value)
}

// Within reports whether every edge of actual is within the tolerance of
// requested — the only honest success signal, because AX set-frame is a
// proposal that apps clamp silently (ADR3.1). Percent tolerances scale by
// the requested width for x edges and height for y edges.
func (t Tolerance) Within(requested, actual mac.CGRect) bool {
	tx, ty := t.Value, t.Value
	if t.Percent {
		tx = requested.Size.W * t.Value / 100
		ty = requested.Size.H * t.Value / 100
	}
	return math.Abs(actual.Origin.X-requested.Origin.X) <= tx &&
		math.Abs(actual.Origin.Y-requested.Origin.Y) <= ty &&
		math.Abs((actual.Origin.X+actual.Size.W)-(requested.Origin.X+requested.Size.W)) <= tx &&
		math.Abs((actual.Origin.Y+actual.Size.H)-(requested.Origin.Y+requested.Size.H)) <= ty
}
