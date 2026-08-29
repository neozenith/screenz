// Package rule is the selector / display / region grammar shared by CLI
// flags and YAML profiles (ADR4.1): one quoting layer, identical keys, so
// `profile save` serialises flags losslessly. Pure parsing and matching.
package rule

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/joshpeak/screenz/internal/discover"
	"github.com/joshpeak/screenz/internal/layout"
)

// Matcher matches a string field either literally or with a /regex/
// (ADR4.2: RE2, spelled /pattern/ with an optional trailing i). The raw
// literal is preserved so the same spelling round-trips through YAML.
type Matcher struct {
	raw     string
	literal string
	re      *regexp.Regexp
}

// IsSet reports whether the matcher was provided at all.
func (m Matcher) IsSet() bool { return m.raw != "" }

// String returns the original literal (lossless for profile save).
func (m Matcher) String() string { return m.raw }

// Value returns the YAML spelling of the matcher: a regex keeps its
// /slashes/, a literal drops CLI quoting (YAML has its own quoting layer)
// — except a slash-leading literal, which keeps grammar quotes so a
// reload cannot reinterpret it as a regex.
func (m Matcher) Value() string {
	if m.re != nil {
		return m.raw
	}
	if strings.HasPrefix(m.literal, "/") {
		return `"` + m.literal + `"`
	}
	return m.literal
}

// Match reports whether the field satisfies the matcher.
func (m Matcher) Match(s string) bool {
	if m.re != nil {
		return m.re.MatchString(s)
	}
	return s == m.literal
}

// ParseMatcher parses a value literal: /regex/ (optional trailing i),
// "quoted", or a bare word.
func ParseMatcher(raw string) (Matcher, error) {
	m := Matcher{raw: raw}
	switch {
	case strings.HasPrefix(raw, "/"):
		body, flags, ok := cutRegex(raw)
		if !ok {
			return Matcher{}, fmt.Errorf("regex %q: want /pattern/ with an optional trailing i", raw)
		}
		if flags == "i" {
			body = "(?i)" + body
		}
		re, err := regexp.Compile(body)
		if err != nil {
			return Matcher{}, fmt.Errorf("regex %q: %v", raw, err)
		}
		m.re = re
	case strings.HasPrefix(raw, `"`):
		if len(raw) < 2 || !strings.HasSuffix(raw, `"`) {
			return Matcher{}, fmt.Errorf("unterminated quote in %q", raw)
		}
		m.literal = raw[1 : len(raw)-1]
	default:
		m.literal = raw
	}
	return m, nil
}

// cutRegex splits /pattern/i into pattern and flags. The pattern ends at
// the LAST slash (ADR4.2), so slashes inside window titles need no escaping.
func cutRegex(raw string) (body, flags string, ok bool) {
	last := strings.LastIndexByte(raw, '/')
	if last <= 0 {
		return "", "", false
	}
	body, flags = raw[1:last], raw[last+1:]
	if flags != "" && flags != "i" {
		return "", "", false
	}
	return body, flags, true
}

// Term is one key=value pair of a term list.
type Term struct {
	Key   string
	Value Matcher
}

// swallowedTerm spots a known grammar key opening a new term inside what
// the last-slash rule captured as one regex body.
var swallowedTerm = regexp.MustCompile(` (bundle|app|title|index|name|uuid|serial|built-in|main)=`)

// parseTerms tokenizes the shared term-list grammar: space-separated
// key=value terms where a value is a bare word, a "quoted string", or a
// /regex/ terminated at the last slash of the input (ADR4.2).
func parseTerms(s string) ([]Term, error) {
	var out []Term
	i, n := 0, len(s)
	for i < n {
		for i < n && s[i] == ' ' {
			i++
		}
		if i >= n {
			break
		}
		eq := strings.IndexByte(s[i:], '=')
		if eq <= 0 {
			return nil, fmt.Errorf("term %q: want key=value", strings.TrimSpace(s[i:]))
		}
		key := s[i : i+eq]
		if strings.ContainsAny(key, ` "/`) {
			return nil, fmt.Errorf("term %q: want key=value", key)
		}
		i += eq + 1
		var raw string
		switch {
		case i < n && s[i] == '"':
			j := strings.IndexByte(s[i+1:], '"')
			if j < 0 {
				return nil, fmt.Errorf("unterminated quote in %q", s)
			}
			raw = s[i : i+j+2]
			i += j + 2
		case i < n && s[i] == '/':
			// The pattern ends at the last slash (ADR4.2) — but only a
			// slash outside quoted regions can terminate it, so a later
			// "quoted/value" cannot capture the terminator.
			last := -1
			inQuote := false
			for k := i + 1; k < n; k++ {
				switch s[k] {
				case '"':
					inQuote = !inQuote
				case '/':
					if !inQuote {
						last = k
					}
				}
			}
			if last < 0 {
				return nil, fmt.Errorf("unterminated regex in %q", s)
			}
			end := last + 1
			if end < n && s[end] == 'i' {
				end++
			}
			if end < n && s[end] != ' ' {
				return nil, fmt.Errorf("regex in %q must end its term", s)
			}
			raw = s[i:end]
			// Two regex terms in one list would silently merge under the
			// last-slash rule; detect a swallowed sibling term and fail
			// loudly instead of matching the wrong windows.
			if swallowedTerm.MatchString(raw) {
				return nil, fmt.Errorf("ambiguous regex termination in %q: the pattern swallows a later term (put the regex term last, or use only one regex per list)", raw)
			}
			i = end
		default:
			j := strings.IndexByte(s[i:], ' ')
			if j < 0 {
				j = n - i
			}
			raw = s[i : i+j]
			i += j
			if raw == "" {
				return nil, fmt.Errorf("empty value for key %q", key)
			}
		}
		m, err := ParseMatcher(raw)
		if err != nil {
			return nil, err
		}
		out = append(out, Term{Key: key, Value: m})
	}
	return out, nil
}

// Selector matches windows: bundle=, app=, title= terms ANDed together.
type Selector struct {
	Terms []Term
	raw   string
}

// String returns the original selector literal.
func (s Selector) String() string { return s.raw }

// ParseSelector parses a --match value.
func ParseSelector(s string) (Selector, error) {
	terms, err := parseTerms(s)
	if err != nil {
		return Selector{}, err
	}
	if len(terms) == 0 {
		return Selector{}, fmt.Errorf("empty selector")
	}
	for _, t := range terms {
		switch t.Key {
		case "bundle", "app", "title":
		default:
			return Selector{}, fmt.Errorf("selector key %q: want bundle, app or title", t.Key)
		}
	}
	return Selector{Terms: terms, raw: s}, nil
}

// Matches reports whether every term matches the window.
func (s Selector) Matches(w discover.Window) bool {
	for _, t := range s.Terms {
		var field string
		switch t.Key {
		case "bundle":
			field = w.Bundle
		case "app":
			field = w.App
		default:
			field = w.Title
		}
		if !t.Value.Match(field) {
			return false
		}
	}
	return len(s.Terms) > 0
}

// DisplaySpec addresses one connected display, either by alias (resolved
// through a profile's displays map) or by terms: index=N, name=, uuid=,
// serial=N, built-in=BOOL, main=BOOL.
type DisplaySpec struct {
	Alias   string
	Index   int
	Serial  string
	Name    Matcher
	UUID    Matcher
	BuiltIn *bool
	Main    *bool
	raw     string
}

// String returns the original display literal.
func (d DisplaySpec) String() string { return d.raw }

// IsSet reports whether any addressing was provided.
func (d DisplaySpec) IsSet() bool { return d.raw != "" }

// ParseDisplay parses a --display value: a bare integer (shorthand for
// index=N — numeric alias names are reserved), a bare alias word, or a
// term list.
func ParseDisplay(s string) (DisplaySpec, error) {
	spec := DisplaySpec{raw: s}
	if s == "" {
		return DisplaySpec{}, fmt.Errorf("empty display")
	}
	if isDigits(s) {
		v, err := strconv.Atoi(s)
		if err != nil || v < 1 {
			return DisplaySpec{}, fmt.Errorf("display index %q: want an integer >= 1", s)
		}
		spec.Index = v
		return spec, nil
	}
	if !strings.ContainsAny(s, "= ") {
		spec.Alias = s
		return spec, nil
	}
	terms, err := parseTerms(s)
	if err != nil {
		return DisplaySpec{}, err
	}
	for _, t := range terms {
		raw := t.Value.String()
		switch t.Key {
		case "index":
			v, err := strconv.Atoi(raw)
			if err != nil || v < 1 {
				return DisplaySpec{}, fmt.Errorf("display index %q: want an integer >= 1", raw)
			}
			spec.Index = v
		case "serial":
			if _, err := strconv.ParseUint(raw, 10, 32); err != nil {
				return DisplaySpec{}, fmt.Errorf("display serial %q: want an integer", raw)
			}
			spec.Serial = raw
		case "name":
			spec.Name = t.Value
		case "uuid":
			spec.UUID = t.Value
		case "built-in", "main":
			v, err := strconv.ParseBool(raw)
			if err != nil {
				return DisplaySpec{}, fmt.Errorf("display %s=%q: want true or false", t.Key, raw)
			}
			if t.Key == "built-in" {
				spec.BuiltIn = &v
			} else {
				spec.Main = &v
			}
		default:
			return DisplaySpec{}, fmt.Errorf("display key %q: want index, name, uuid, serial, built-in or main", t.Key)
		}
	}
	return spec, nil
}

// Matches reports whether the spec addresses this display. An alias never
// matches directly — it must be resolved through a profile first.
func (d DisplaySpec) Matches(disp discover.Display) bool {
	if d.Alias != "" {
		return false
	}
	if d.Index != 0 && disp.Index != d.Index {
		return false
	}
	if d.Serial != "" && strconv.FormatUint(uint64(disp.Serial), 10) != d.Serial {
		return false
	}
	if d.Name.IsSet() && !d.Name.Match(disp.Name) {
		return false
	}
	if d.UUID.IsSet() && !d.UUID.Match(disp.UUID) {
		return false
	}
	if d.BuiltIn != nil && disp.BuiltIn != *d.BuiltIn {
		return false
	}
	if d.Main != nil && disp.Main != *d.Main {
		return false
	}
	return true
}

// isDigits reports whether s is a non-empty run of ASCII digits.
func isDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return len(s) > 0
}

// term-by-term diagnosis for a spec that matches nothing: each entry is one
// constraining field of the spec as its own single-term spec.
func (d DisplaySpec) terms() []DisplaySpec {
	var out []DisplaySpec
	if d.Index != 0 {
		out = append(out, DisplaySpec{Index: d.Index, raw: fmt.Sprintf("index=%d", d.Index)})
	}
	if d.Serial != "" {
		out = append(out, DisplaySpec{Serial: d.Serial, raw: "serial=" + d.Serial})
	}
	if d.Name.IsSet() {
		out = append(out, DisplaySpec{Name: d.Name, raw: "name=" + d.Name.String()})
	}
	if d.UUID.IsSet() {
		out = append(out, DisplaySpec{UUID: d.UUID, raw: "uuid=" + d.UUID.String()})
	}
	if d.BuiltIn != nil {
		out = append(out, DisplaySpec{BuiltIn: d.BuiltIn, raw: fmt.Sprintf("built-in=%t", *d.BuiltIn)})
	}
	if d.Main != nil {
		out = append(out, DisplaySpec{Main: d.Main, raw: fmt.Sprintf("main=%t", *d.Main)})
	}
	return out
}

// Explain diagnoses a multi-term spec that matched no display: terms AND
// together, and the usual cause is two individually correct terms that
// contradict each other (name says one panel, index says another). Returns
// "" when the spec has fewer than two terms — there is nothing to untangle.
func (d DisplaySpec) Explain(displays []discover.Display) string {
	terms := d.terms()
	if len(terms) < 2 {
		return ""
	}
	parts := make([]string, 0, len(terms))
	for _, t := range terms {
		var hit []string
		for _, disp := range displays {
			if t.Matches(disp) {
				hit = append(hit, fmt.Sprintf("index=%d %s", disp.Index, disp.Name))
			}
		}
		if len(hit) == 0 {
			parts = append(parts, t.raw+" matches nothing")
		} else {
			parts = append(parts, t.raw+" matches ["+strings.Join(hit, "; ")+"]")
		}
	}
	return "terms AND together: " + strings.Join(parts, ", ") + " — remove or fix the conflicting term"
}

// Rule is one selector → display → region placement rule.
type Rule struct {
	Match     Selector
	Display   DisplaySpec
	Region    layout.Region
	HasRegion bool
	Gap       float64
	Tolerance layout.Tolerance
	First     bool   // place only the first matching window (YAML each: false)
	Order     string // existing | title | pid
}

// Orders are the valid --order values: existing keeps AX order.
var Orders = map[string]bool{"existing": true, "title": true, "pid": true}
