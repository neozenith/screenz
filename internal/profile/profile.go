package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/neozenith/screenz/internal/layout"
	"github.com/neozenith/screenz/internal/rule"
)

// Profile is one named rule set. The YAML keys map one-to-one onto the CLI
// rule grammar (ADR4.1) so `profile save` is lossless: match.bundle ↔
// --match bundle=, display alias ↔ --display ALIAS, region ↔ --region,
// gap, tolerance, each: false ↔ --first, order.
type Profile struct {
	Version  int
	Name     string
	Displays map[string]rule.DisplaySpec
	Rules    []*rule.Rule
}

// Path is the file for a profile name: <dir>/profiles/<name>.yaml (ADR5.2).
func Path(getenv func(string) string, home, name string) string {
	return filepath.Join(Dir(getenv, home), "profiles", name+".yaml")
}

// profileYAML is the on-disk shape, read strictly (an unknown key is an
// error, never dropped silently — ADR5.1) with a comment map so hand-written
// comments survive a save. Block style everywhere: goccy loses comments
// inside flow style (goccy/go-yaml#608).
type profileYAML struct {
	Version  int                 `yaml:"version"`
	Name     string              `yaml:"name"`
	Displays map[string]specYAML `yaml:"displays,omitempty"`
	Rules    []ruleYAML          `yaml:"rules"`
}

type specYAML struct {
	Name    string `yaml:"name,omitempty"`
	Index   int    `yaml:"index,omitempty"`
	UUID    string `yaml:"uuid,omitempty"`
	Serial  uint64 `yaml:"serial,omitempty"`
	BuiltIn *bool  `yaml:"built-in,omitempty"`
	Main    *bool  `yaml:"main,omitempty"`
}

type ruleYAML struct {
	Match     matchYAML   `yaml:"match"`
	Display   displayYAML `yaml:"display"`
	Region    string      `yaml:"region"`
	Gap       *float64    `yaml:"gap,omitempty"`
	Tolerance scalar      `yaml:"tolerance,omitempty"`
	Each      *bool       `yaml:"each,omitempty"`
	Order     string      `yaml:"order,omitempty"`
}

type matchYAML struct {
	Bundle string `yaml:"bundle,omitempty"`
	App    string `yaml:"app,omitempty"`
	Title  string `yaml:"title,omitempty"`
}

// displayYAML is either a bare alias string or an inline spec map.
type displayYAML struct {
	Alias string
	Spec  *specYAML
}

func (d *displayYAML) UnmarshalYAML(b []byte) error {
	var s string
	if err := yaml.Unmarshal(b, &s); err == nil {
		d.Alias = s
		return nil
	}
	var sp specYAML
	if err := yaml.UnmarshalWithOptions(b, &sp, yaml.Strict()); err != nil {
		return err
	}
	d.Spec = &sp
	return nil
}

func (d displayYAML) MarshalYAML() (any, error) {
	if d.Alias != "" {
		return d.Alias, nil
	}
	return d.Spec, nil
}

// scalar keeps a YAML scalar's literal spelling (tolerance: 5% and 0.5 are
// different YAML types but one CLI grammar). Unmarshal goes through a
// nested yaml parse because goccy hands custom unmarshalers the raw node
// bytes including any trailing line comment.
type scalar string

func (s *scalar) UnmarshalYAML(b []byte) error {
	var v string
	if err := yaml.Unmarshal(b, &v); err != nil {
		return err
	}
	*s = scalar(v)
	return nil
}

func (s scalar) MarshalYAML() (any, error) {
	if f, err := strconv.ParseFloat(string(s), 64); err == nil {
		return f, nil
	}
	return string(s), nil
}

// cliValue quotes a YAML literal for the CLI term grammar when needed. A
// value already carrying grammar quotes (a slash-leading literal saved by
// Matcher.Value) or spelled as a regex passes through untouched.
func cliValue(v string) string {
	if strings.HasPrefix(v, `"`) || strings.HasPrefix(v, "/") || !strings.Contains(v, " ") {
		return v
	}
	return `"` + v + `"`
}

func (s specYAML) toSpec() (rule.DisplaySpec, error) {
	var parts []string
	if s.Name != "" {
		parts = append(parts, "name="+cliValue(s.Name))
	}
	if s.Index != 0 {
		parts = append(parts, fmt.Sprintf("index=%d", s.Index))
	}
	if s.UUID != "" {
		parts = append(parts, "uuid="+cliValue(s.UUID))
	}
	if s.Serial != 0 {
		parts = append(parts, fmt.Sprintf("serial=%d", s.Serial))
	}
	if s.BuiltIn != nil {
		parts = append(parts, fmt.Sprintf("built-in=%t", *s.BuiltIn))
	}
	if s.Main != nil {
		parts = append(parts, fmt.Sprintf("main=%t", *s.Main))
	}
	if len(parts) == 0 {
		return rule.DisplaySpec{}, fmt.Errorf("empty display spec")
	}
	return rule.ParseDisplay(strings.Join(parts, " "))
}

func specFromRule(d rule.DisplaySpec) *specYAML {
	out := &specYAML{Index: d.Index, BuiltIn: d.BuiltIn, Main: d.Main}
	if d.Name.IsSet() {
		out.Name = d.Name.Value()
	}
	if d.UUID.IsSet() {
		out.UUID = d.UUID.Value()
	}
	if d.Serial != "" {
		out.Serial, _ = strconv.ParseUint(d.Serial, 10, 64)
	}
	return out
}

func (m matchYAML) toSelector() (rule.Selector, error) {
	var parts []string
	for _, kv := range []struct{ k, v string }{{"bundle", m.Bundle}, {"app", m.App}, {"title", m.Title}} {
		if kv.v != "" {
			parts = append(parts, kv.k+"="+cliValue(kv.v))
		}
	}
	if len(parts) == 0 {
		return rule.Selector{}, fmt.Errorf("empty match")
	}
	return rule.ParseSelector(strings.Join(parts, " "))
}

func matchFromSelector(s rule.Selector) matchYAML {
	var m matchYAML
	for _, t := range s.Terms {
		switch t.Key {
		case "bundle":
			m.Bundle = t.Value.Value()
		case "app":
			m.App = t.Value.Value()
		default:
			m.Title = t.Value.Value()
		}
	}
	return m
}

func (ry ruleYAML) toRule(i int) (*rule.Rule, error) {
	wrap := func(err error) error { return fmt.Errorf("rules[%d]: %w", i, err) }
	r := &rule.Rule{Order: "existing", Tolerance: layout.Tolerance{Value: layout.DefaultTolerance}}
	sel, err := ry.Match.toSelector()
	if err != nil {
		return nil, wrap(err)
	}
	r.Match = sel
	switch {
	case ry.Display.Alias != "":
		spec, err := rule.ParseDisplay(ry.Display.Alias)
		if err != nil {
			return nil, wrap(err)
		}
		r.Display = spec
	case ry.Display.Spec != nil:
		spec, err := ry.Display.Spec.toSpec()
		if err != nil {
			return nil, wrap(err)
		}
		r.Display = spec
	default:
		return nil, wrap(fmt.Errorf("missing display"))
	}
	if ry.Region == "" {
		return nil, wrap(fmt.Errorf("missing region"))
	}
	region, err := layout.ParseRegion(ry.Region)
	if err != nil {
		return nil, wrap(err)
	}
	r.Region, r.HasRegion = region, true
	if ry.Gap != nil {
		if *ry.Gap < 0 {
			return nil, wrap(fmt.Errorf("gap must be non-negative"))
		}
		r.Gap = *ry.Gap
	}
	if ry.Tolerance != "" {
		tol, err := layout.ParseTolerance(string(ry.Tolerance))
		if err != nil {
			return nil, wrap(err)
		}
		r.Tolerance = tol
	}
	if ry.Each != nil {
		r.First = !*ry.Each
	}
	if ry.Order != "" {
		if !rule.Orders[ry.Order] {
			return nil, wrap(fmt.Errorf("order %q: want existing, title or pid", ry.Order))
		}
		r.Order = ry.Order
	}
	return r, nil
}

func yamlFromRule(r *rule.Rule) ruleYAML {
	ry := ruleYAML{Match: matchFromSelector(r.Match), Region: r.Region.String()}
	if r.Display.Alias != "" {
		ry.Display = displayYAML{Alias: r.Display.Alias}
	} else {
		ry.Display = displayYAML{Spec: specFromRule(r.Display)}
	}
	if r.Gap != 0 {
		ry.Gap = &r.Gap
	}
	if r.Tolerance != (layout.Tolerance{Value: layout.DefaultTolerance}) {
		ry.Tolerance = scalar(r.Tolerance.String())
	}
	if r.First {
		f := false
		ry.Each = &f
	}
	if r.Order != "existing" {
		ry.Order = r.Order
	}
	return ry
}

// Parse reads a profile strictly from YAML bytes.
func Parse(src []byte) (*Profile, error) {
	var py profileYAML
	if err := yaml.UnmarshalWithOptions(src, &py, yaml.Strict()); err != nil {
		return nil, err
	}
	return convert(py)
}

func convert(py profileYAML) (*Profile, error) {
	if py.Version != 1 {
		return nil, fmt.Errorf("unsupported profile version %d (want 1)", py.Version)
	}
	p := &Profile{Version: py.Version, Name: py.Name}
	if len(py.Displays) > 0 {
		p.Displays = map[string]rule.DisplaySpec{}
		for alias, sy := range py.Displays {
			if _, err := strconv.Atoi(alias); err == nil {
				return nil, fmt.Errorf("displays.%s: purely numeric alias names are reserved (a bare number in display: means index=N)", alias)
			}
			spec, err := sy.toSpec()
			if err != nil {
				return nil, fmt.Errorf("displays.%s: %w", alias, err)
			}
			p.Displays[alias] = spec
		}
	}
	for i, ry := range py.Rules {
		r, err := ry.toRule(i)
		if err != nil {
			return nil, err
		}
		p.Rules = append(p.Rules, r)
	}
	return p, nil
}

// Load reads a profile file.
func Load(path string) (*Profile, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	p, err := Parse(src)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return p, nil
}

// Resolved returns the profile's rules with display aliases substituted
// from the displays map. An unknown alias errors here — before any window
// moves — and the substituted spec still resolves through the plan phase,
// so an alias that matches zero or two connected displays exits 1 (ADR5.1).
func (p *Profile) Resolved(extra []*rule.Rule) ([]*rule.Rule, error) {
	all := append(append([]*rule.Rule{}, p.Rules...), extra...)
	out := make([]*rule.Rule, 0, len(all))
	for i, r := range all {
		if r.Display.Alias != "" {
			spec, ok := p.Displays[r.Display.Alias]
			if !ok {
				return nil, fmt.Errorf("rule %d: display alias %q is not defined in profile %q (known: %s)",
					i+1, r.Display.Alias, p.Name, strings.Join(p.Aliases(), ", "))
			}
			r2 := *r
			r2.Display = spec
			r = &r2
		}
		out = append(out, r)
	}
	return out, nil
}

// Aliases lists the profile's display aliases, sorted.
func (p *Profile) Aliases() []string {
	out := make([]string, 0, len(p.Displays))
	for a := range p.Displays {
		out = append(out, a)
	}
	sort.Strings(out)
	return out
}

// Append parses an existing profile with its comment map, appends rules,
// and writes it back: append-only keeps every existing $.rules[i] comment
// path valid, so hand-written comments survive (ADR5.1).
func Append(path string, rules []*rule.Rule) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var py profileYAML
	cm := yaml.CommentMap{}
	if err := yaml.UnmarshalWithOptions(src, &py, yaml.Strict(), yaml.CommentToMap(cm)); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	if _, err := convert(py); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	for _, r := range rules {
		py.Rules = append(py.Rules, yamlFromRule(r))
		cm[fmt.Sprintf("$.rules[%d]", len(py.Rules)-1)] = []*yaml.Comment{yaml.HeadComment(" added by screenz profile save")}
	}
	// Marshal cannot fail for these types (every custom marshaler is
	// error-free), so the error is not a reachable branch.
	out, _ := yaml.MarshalWithOptions(py, yaml.WithComment(cm), yaml.IndentSequence(true))
	return writeFileAtomic(path, out)
}

// writeFileAtomic writes via a sibling temp file and rename so an
// interrupted save can never leave a truncated profile — profiles carry
// hand-written comments, so losing one is expensive.
func writeFileAtomic(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// WriteNew creates a profile file holding exactly the given rules. It
// writes no displays map, so a rule addressing an alias would produce a
// profile that always fails to apply — reject it here instead.
func WriteNew(path, name string, rules []*rule.Rule) error {
	for i, r := range rules {
		if r.Display.Alias != "" {
			return fmt.Errorf("rule %d: display alias %q cannot be saved into a new profile (it has no displays map); use inline display terms, or add the alias under displays: in the profile file first", i+1, r.Display.Alias)
		}
	}
	py := profileYAML{Version: 1, Name: name}
	for _, r := range rules {
		py.Rules = append(py.Rules, yamlFromRule(r))
	}
	cm := yaml.CommentMap{
		"$.version": []*yaml.Comment{yaml.HeadComment(
			fmt.Sprintf(" screenz profile %q — written by: screenz profile save %s", name, name),
			fmt.Sprintf(" Apply with:   screenz apply %s", name),
			fmt.Sprintf(" Preview with: screenz apply --dry-run %s", name))},
	}
	out, _ := yaml.MarshalWithOptions(py, yaml.WithComment(cm), yaml.IndentSequence(true))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return writeFileAtomic(path, out)
}
