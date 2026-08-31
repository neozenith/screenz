package rule

import (
	"flag"
	"io"
	"strings"
	"testing"

	"github.com/neozenith/screenz/internal/discover"
)

func win(bundle, app, title string) discover.Window {
	return discover.Window{Bundle: bundle, App: app, Title: title}
}

func TestParseSelectorAndMatch(t *testing.T) {
	cases := []struct {
		selector string
		window   discover.Window
		want     bool
	}{
		{"bundle=com.microsoft.VSCode", win("com.microsoft.VSCode", "Code", "main.go"), true},
		{"bundle=com.microsoft.VSCode", win("com.microsoft.edgemac", "Edge", "x"), false},
		{`app="Google Chrome" title=/Work/`, win("com.google.Chrome", "Google Chrome", "Work — Inbox"), true},
		{`app="Google Chrome" title=/Work/`, win("com.google.Chrome", "Google Chrome", "Personal — News"), false},
		{`title=/work/i`, win("b", "a", "My WORK window"), true},
		{`title=/work/`, win("b", "a", "My WORK window"), false},
		// Last-slash termination: slashes inside the pattern need no escaping.
		{`title=/repo/main/`, win("b", "a", "x repo/main y"), true},
		{`title=/repo/main/`, win("b", "a", "repo main"), false},
		// A later quoted value containing '/' cannot capture the terminator.
		{`title=/Work/ app="A/B"`, win("b", "A/B", "Work — Inbox"), true},
		{`title=/Work/ app="A/B"`, win("b", "Other", "Work — Inbox"), false},
		{`app=Code title=screenz`, win("b", "Code", "screenz"), true},
	}
	for _, tc := range cases {
		t.Run(tc.selector, func(t *testing.T) {
			sel, err := ParseSelector(tc.selector)
			if err != nil {
				t.Fatalf("ParseSelector(%q): %v", tc.selector, err)
			}
			if got := sel.Matches(tc.window); got != tc.want {
				t.Fatalf("Matches(%+v) = %v, want %v", tc.window, got, tc.want)
			}
			if sel.String() != tc.selector {
				t.Fatalf("String() = %q, not lossless", sel.String())
			}
		})
	}
	if (Selector{}).Matches(win("b", "a", "t")) {
		t.Error("empty selector must match nothing")
	}
}

func TestParseSelectorErrors(t *testing.T) {
	for _, bad := range []string{
		"", "   ", "bundle", "=x", "state=normal", "title=/unterminated",
		`title="unterminated`, "title=/a/x", "title=/a/ix", "title=",
		"a b=c", `title=/(?=lookahead)/`, "bundle=a title=/a/ x",
		// Two regex terms would silently merge under the last-slash rule;
		// the parser must fail loudly instead of matching wrong windows.
		`app=/Chrome/ title=/Work/`,
	} {
		if _, err := ParseSelector(bad); err == nil {
			t.Errorf("ParseSelector(%q) accepted", bad)
		}
	}
}

func TestParseMatcherDirect(t *testing.T) {
	if _, err := ParseMatcher(`"unterminated`); err == nil {
		t.Error("unterminated quote accepted")
	}
	if _, err := ParseMatcher(`/`); err == nil {
		t.Error("bare slash accepted")
	}
	if _, err := ParseMatcher(`/a/x`); err == nil {
		t.Error("bad regex flags accepted (only i is valid)")
	}
	m, err := ParseMatcher(`"quoted value"`)
	if err != nil || !m.Match("quoted value") || m.Match("other") {
		t.Errorf("quoted matcher wrong: %v %+v", err, m)
	}
	// Value(): the YAML spelling drops CLI quoting, regexes keep slashes.
	if m.Value() != "quoted value" {
		t.Errorf("Value() = %q", m.Value())
	}
	re, _ := ParseMatcher(`/Work/i`)
	if re.Value() != "/Work/i" {
		t.Errorf("regex Value() = %q", re.Value())
	}
	// A slash-leading literal keeps grammar quotes in its YAML spelling so
	// a reload cannot reinterpret it as a regex.
	slash, _ := ParseMatcher(`"/foo"`)
	if slash.Value() != `"/foo"` {
		t.Errorf("slash literal Value() = %q", slash.Value())
	}
	if !m.IsSet() || (Matcher{}).IsSet() {
		t.Error("IsSet wrong")
	}
}

func display(index int, name, uuid string, serial uint32, builtIn, main bool) discover.Display {
	return discover.Display{Index: index, Name: name, UUID: uuid, Serial: serial, BuiltIn: builtIn, Main: main}
}

func TestParseDisplayAndMatch(t *testing.T) {
	panels := []discover.Display{
		display(1, "Built-in Retina Display", "UUID-BUILTIN", 0, true, true),
		display(2, "LU28R55 (1)", "UUID-LEFT", 1129796439, false, false),
		display(3, "LU28R55 (2)", "UUID-RIGHT", 1129796439, false, false),
	}
	cases := []struct {
		spec string
		want []int // matching display indexes
	}{
		{"index=2", []int{2}},
		{"name=/LU28R55/", []int{2, 3}},
		{"name=/LU28R55/ index=3", []int{3}},
		{`name="LU28R55 (1)"`, []int{2}},
		{"built-in=true", []int{1}},
		{"main=false", []int{2, 3}},
		{"serial=1129796439", []int{2, 3}},
		{"uuid=UUID-RIGHT", []int{3}},
		{"index=9", nil},
	}
	for _, tc := range cases {
		t.Run(tc.spec, func(t *testing.T) {
			spec, err := ParseDisplay(tc.spec)
			if err != nil {
				t.Fatalf("ParseDisplay(%q): %v", tc.spec, err)
			}
			var got []int
			for _, d := range panels {
				if spec.Matches(d) {
					got = append(got, d.Index)
				}
			}
			if len(got) != len(tc.want) {
				t.Fatalf("matched %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("matched %v, want %v", got, tc.want)
				}
			}
			if spec.String() != tc.spec {
				t.Fatalf("String() = %q, not lossless", spec.String())
			}
		})
	}
}

func TestParseDisplayBareIndexShorthand(t *testing.T) {
	spec, err := ParseDisplay("2")
	if err != nil || spec.Index != 2 || spec.Alias != "" {
		t.Fatalf("ParseDisplay(2) = %+v, %v", spec, err)
	}
	if spec.String() != "2" {
		t.Errorf("String() = %q, not lossless", spec.String())
	}
	if !spec.Matches(display(2, "LU28R55 (1)", "u", 0, false, false)) {
		t.Error("bare index must match by index")
	}
	if _, err := ParseDisplay("0"); err == nil {
		t.Error("index 0 accepted")
	}
}

// The identical-monitors trap: two individually correct terms that AND to
// nothing must be explained per term, not just counted.
func TestDisplayExplainDiagnosesConflictingTerms(t *testing.T) {
	panels := []discover.Display{
		display(1, "Built-in Retina Display", "UUID-BUILTIN", 0, true, true),
		display(2, "LU28R55 (1)", "UUID-LEFT", 1129796439, false, false),
		display(3, "LU28R55 (2)", "UUID-RIGHT", 1129796439, false, false),
	}
	spec, err := ParseDisplay(`name="LU28R55 (1)" index=1`)
	if err != nil {
		t.Fatal(err)
	}
	got := spec.Explain(panels)
	for _, want := range []string{
		"terms AND together",
		`name="LU28R55 (1)" matches [index=2 LU28R55 (1)]`,
		"index=1 matches [index=1 Built-in Retina Display]",
		"remove or fix the conflicting term",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Explain missing %q:\n%s", want, got)
		}
	}
	// A term matching nothing at all says so.
	spec, _ = ParseDisplay("name=/XDR/ index=1")
	if got := spec.Explain(panels); !strings.Contains(got, "name=/XDR/ matches nothing") {
		t.Errorf("Explain = %q", got)
	}
	// Single-term specs have nothing to untangle.
	spec, _ = ParseDisplay("index=9")
	if got := spec.Explain(panels); got != "" {
		t.Errorf("single-term Explain = %q, want empty", got)
	}
	// Every term kind renders.
	spec, _ = ParseDisplay("uuid=UUID-LEFT serial=1129796439 built-in=false main=false")
	got = spec.Explain(panels)
	for _, want := range []string{"uuid=UUID-LEFT", "serial=1129796439", "built-in=false", "main=false"} {
		if !strings.Contains(got, want) {
			t.Errorf("Explain missing %q:\n%s", want, got)
		}
	}
}

func TestParseDisplayAlias(t *testing.T) {
	spec, err := ParseDisplay("dell-left")
	if err != nil || spec.Alias != "dell-left" || !spec.IsSet() {
		t.Fatalf("alias parse: %+v, %v", spec, err)
	}
	if spec.Matches(display(1, "any", "u", 0, false, false)) {
		t.Error("an unresolved alias must match nothing")
	}
}

func TestParseDisplayErrors(t *testing.T) {
	for _, bad := range []string{
		"", "index=0", "index=abc", "serial=abc", "built-in=maybe",
		"main=x", "color=red", "name=/bad(/", "index=1 extra",
	} {
		if _, err := ParseDisplay(bad); err == nil {
			t.Errorf("ParseDisplay(%q) accepted", bad)
		}
	}
}

// parse runs the office three-rule command through a real FlagSet, exactly
// as apply does.
func parse(t *testing.T, args ...string) (*List, error) {
	t.Helper()
	fs := flag.NewFlagSet("apply", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	l := &List{}
	l.Register(fs)
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return l, l.Validate()
}

func TestFlagsBuildRulesInOrder(t *testing.T) {
	l, err := parse(t,
		"--match", "bundle=com.microsoft.VSCode", "--display", "index=1", "--region", "maximize",
		"--match", `app="Google Chrome" title=/Work/`, "--display", "index=2", "--region", "left-half", "--gap", "8",
		"--match", "bundle=com.microsoft.edgemac", "--display", "index=2", "--region", "right-half",
		"--tolerance", "5%", "--first", "--order", "title",
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(l.Rules) != 3 {
		t.Fatalf("got %d rules", len(l.Rules))
	}
	r1, r2, r3 := l.Rules[0], l.Rules[1], l.Rules[2]
	if r1.Region.String() != "maximize" || r1.Display.Index != 1 || r1.Order != "existing" {
		t.Errorf("rule 1 wrong: %+v", r1)
	}
	if r2.Gap != 8 || r2.Region.String() != "left-half" {
		t.Errorf("rule 2 wrong: %+v", r2)
	}
	// Sibling flags bound to the MOST RECENT rule (ADR4.1).
	if !r3.First || r3.Order != "title" || r3.Tolerance.String() != "5%" {
		t.Errorf("rule 3 wrong: %+v", r3)
	}
	if r1.First || r2.First || r1.Tolerance.String() != "0.5" {
		t.Error("sibling flags leaked onto earlier rules")
	}
}

// The one-letter aliases are the same flags, not a parallel set: they open
// and bind rules identically, and may be mixed with the long names in one
// command line (ADR-0021).
func TestShortFlagsAreTheSameFlags(t *testing.T) {
	l, err := parse(t,
		"-m", "bundle=com.microsoft.VSCode", "-d", "index=1",
		"-m", `app="Google Chrome"`, "--display", "index=2", "-r", "left-half", "-g", "8",
		"-m", "bundle=com.microsoft.edgemac", "-d", "2", "-r", "right-half",
		"-t", "5%", "-f", "-o", "title",
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(l.Rules) != 3 {
		t.Fatalf("got %d rules, want 3", len(l.Rules))
	}
	r1, r2, r3 := l.Rules[0], l.Rules[1], l.Rules[2]
	if r1.Display.Index != 1 || r1.Region.String() != "maximize" || r1.First {
		t.Errorf("rule 1 wrong: %+v", r1)
	}
	if r2.Display.Index != 2 || r2.Region.String() != "left-half" || r2.Gap != 8 {
		t.Errorf("rule 2 wrong: %+v", r2)
	}
	if !r3.First || r3.Order != "title" || r3.Tolerance.String() != "5%" {
		t.Errorf("rule 3 wrong: %+v", r3)
	}
	// A short flag before any --match is the same usage error, named by the
	// long form so the message routes to the documented flag.
	_, err = parse(t, "-d", "index=1")
	if err == nil || !strings.Contains(err.Error(), "--display given before any --match") {
		t.Errorf("err = %v, want the --display usage error", err)
	}
}

func TestFlagsSiblingBeforeMatchIsUsageError(t *testing.T) {
	for _, args := range [][]string{
		{"--display", "index=1"},
		{"--region", "maximize"},
		{"--first"},
	} {
		if _, err := parse(t, args...); err == nil {
			t.Errorf("%v accepted before any --match", args)
		}
	}
}

func TestFlagsValueErrorsPropagate(t *testing.T) {
	for _, args := range [][]string{
		{"--match", "="},
		{"--match", "bundle=a", "--display", "index=zero"},
		{"--match", "bundle=a", "--region", "diagonal"},
		{"--match", "bundle=a", "--gap", "-2"},
		{"--match", "bundle=a", "--tolerance", "-1"},
		{"--match", "bundle=a", "--order", "size"},
		{"--match", "bundle=a", "--first=maybe"},
	} {
		if _, err := parse(t, args...); err == nil {
			t.Errorf("%v accepted", args)
		}
	}
}

func TestValidateIncompleteRules(t *testing.T) {
	_, err := parse(t, "--match", "bundle=a", "--region", "maximize")
	if err == nil || !strings.Contains(err.Error(), "missing --display") {
		t.Errorf("missing display not caught: %v", err)
	}
}

// An omitted --region is the whole usable frame (ADR-0020), not the
// degenerate zero Region.
func TestOmittedRegionDefaultsToMaximize(t *testing.T) {
	l, err := parse(t, "--match", "bundle=a", "--display", "index=1")
	if err != nil {
		t.Fatalf("rule without --region rejected: %v", err)
	}
	if got := l.Rules[0].Region.String(); got != "maximize" {
		t.Errorf("default region = %q, want maximize", got)
	}
	// An explicit --region still wins, whichever side of --display it lands.
	l, err = parse(t, "--match", "bundle=a", "--region", "left-half", "--display", "index=1")
	if err != nil {
		t.Fatalf("explicit region rejected: %v", err)
	}
	if got := l.Rules[0].Region.String(); got != "left-half" {
		t.Errorf("explicit region = %q, want left-half", got)
	}
}

func TestFlagValueStringers(t *testing.T) {
	// flag.Value.String is called by the flag package for defaults.
	if (matchFlag{}).String() != "" || (field{}).String() != "" {
		t.Error("zero flag values must render empty")
	}
	if !(field{boolean: true}).IsBoolFlag() {
		t.Error("--first must be a boolean flag")
	}
}
