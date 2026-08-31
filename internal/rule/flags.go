package rule

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/neozenith/screenz/internal/layout"
)

// List collects rules from repeated flags. Every --match opens a new rule
// and each sibling flag binds to the most recent one — the stdlib flag
// package documents that Value.Set is called in command-line order (ADR4.1).
type List struct{ Rules []*Rule }

// Register wires the rule flags onto a FlagSet. Every rule flag carries a
// one-letter alias bound to the same Value (ADR-0021); the long name stays
// canonical in help, in error text and as the profile YAML key.
func (l *List) Register(fs *flag.FlagSet) {
	bind := func(v flag.Value, long, short, usage string) {
		fs.Var(v, long, usage)
		fs.Var(v, short, usage)
	}
	bind(matchFlag{l}, "match", "m", "open a new rule: bundle=, app=, title= terms")
	bind(field{l, "display", setDisplay, false}, "display", "d", "bind the display for the last rule")
	bind(field{l, "region", setRegion, false}, "region", "r", "bind the region for the last rule")
	bind(field{l, "gap", setGap, false}, "gap", "g", "points between window and region edge")
	bind(field{l, "tolerance", setTolerance, false}, "tolerance", "t", "verification tolerance: points or N%")
	bind(field{l, "order", setOrder, false}, "order", "o", "window order within the rule: existing, title or pid")
	bind(field{l, "first", setFirst, true}, "first", "f", "place only the first matching window")
}

// Validate checks that every rule is complete after Parse. Only --display
// is mandatory; an omitted --region defaults to maximize (ADR-0020).
func (l *List) Validate() error {
	for i, r := range l.Rules {
		if !r.Display.IsSet() {
			return fmt.Errorf("rule %d (%s): missing --display", i+1, r.Match)
		}
	}
	return nil
}

type matchFlag struct{ l *List }

func (matchFlag) String() string { return "" }

func (m matchFlag) Set(s string) error {
	sel, err := ParseSelector(s)
	if err != nil {
		return err
	}
	m.l.Rules = append(m.l.Rules, &Rule{
		Match:     sel,
		Order:     "existing",
		Region:    layout.DefaultRegion(),
		Tolerance: layout.Tolerance{Value: layout.DefaultTolerance},
	})
	return nil
}

// field binds a flag value to the most recent rule; using it before any
// --match is a usage error (exit 2).
type field struct {
	l       *List
	name    string
	set     func(*Rule, string) error
	boolean bool
}

func (field) String() string     { return "" }
func (f field) IsBoolFlag() bool { return f.boolean }
func (f field) Set(s string) error {
	if len(f.l.Rules) == 0 {
		return fmt.Errorf("--%s given before any --match", f.name)
	}
	return f.set(f.l.Rules[len(f.l.Rules)-1], s)
}

func setDisplay(r *Rule, s string) error {
	spec, err := ParseDisplay(s)
	if err != nil {
		return err
	}
	r.Display = spec
	return nil
}

func setRegion(r *Rule, s string) error {
	reg, err := layout.ParseRegion(s)
	if err != nil {
		return err
	}
	r.Region = reg
	return nil
}

func setGap(r *Rule, s string) error {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || v < 0 {
		return fmt.Errorf("gap %q: want a non-negative number of points", s)
	}
	r.Gap = v
	return nil
}

func setTolerance(r *Rule, s string) error {
	t, err := layout.ParseTolerance(s)
	if err != nil {
		return err
	}
	r.Tolerance = t
	return nil
}

func setOrder(r *Rule, s string) error {
	if !Orders[s] {
		return fmt.Errorf("order %q: want existing, title or pid", s)
	}
	r.Order = s
	return nil
}

func setFirst(r *Rule, s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return fmt.Errorf("first %q: want true or false", s)
	}
	r.First = v
	return nil
}
