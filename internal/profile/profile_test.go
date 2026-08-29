package profile

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joshpeak/screenz/internal/discover"
	"github.com/joshpeak/screenz/internal/rule"
)

func parseRules(t *testing.T, args ...string) []*rule.Rule {
	t.Helper()
	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	l := &rule.List{}
	l.Register(fs)
	if err := fs.Parse(args); err != nil {
		t.Fatal(err)
	}
	return l.Rules
}

func mkWin(bundle, app, title string) discover.Window {
	return discover.Window{Bundle: bundle, App: app, Title: title}
}

func TestTemplateParses(t *testing.T) {
	p, err := Parse([]byte(Template("office")))
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "office" || p.Version != 1 {
		t.Fatalf("header wrong: %+v", p)
	}
	if len(p.Displays) != 3 || len(p.Rules) != 3 {
		t.Fatalf("got %d displays, %d rules", len(p.Displays), len(p.Rules))
	}
	left := p.Displays["dell-left"]
	if left.Index != 1 || !left.Name.Match("LU28R55 (1)") {
		t.Errorf("dell-left spec wrong: %+v", left)
	}
	laptop := p.Displays["laptop"]
	if laptop.BuiltIn == nil || !*laptop.BuiltIn {
		t.Errorf("laptop spec wrong: %+v", laptop)
	}
	r1, r2, r3 := p.Rules[0], p.Rules[1], p.Rules[2]
	if r1.Display.Alias != "dell-left" || r1.Region.String() != "maximize" {
		t.Errorf("rule 1 wrong: %+v", r1)
	}
	if r2.Gap != 8 || !r2.Match.Matches(mkWin("com.google.Chrome", "Google Chrome", "Work — Inbox")) {
		t.Errorf("rule 2 wrong: %+v", r2)
	}
	if r3.Tolerance.String() != "5%" || r3.First || r3.Order != "title" {
		t.Errorf("rule 3 wrong: %+v", r3)
	}
}

func TestParseErrors(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{"bad version", "version: 2\nname: x\nrules: []\n", "unsupported profile version"},
		{"unknown key", "version: 1\nname: x\ncolor: red\nrules: []\n", "unknown field"},
		{"unknown rule key", "version: 1\nname: x\nrules:\n  - match: {bundle: a}\n    display: {index: 1}\n    region: maximize\n    snap: true\n", "unknown field"},
		{"empty match", "version: 1\nname: x\nrules:\n  - match: {}\n    display: {index: 1}\n    region: maximize\n", "empty match"},
		{"missing display", "version: 1\nname: x\nrules:\n  - match: {bundle: a}\n    region: maximize\n", "missing display"},
		{"empty display spec", "version: 1\nname: x\nrules:\n  - match: {bundle: a}\n    display: {}\n    region: maximize\n", "empty display spec"},
		{"missing region", "version: 1\nname: x\nrules:\n  - match: {bundle: a}\n    display: {index: 1}\n", "missing region"},
		{"bad region", "version: 1\nname: x\nrules:\n  - match: {bundle: a}\n    display: {index: 1}\n    region: diagonal\n", "unknown region"},
		{"bad gap", "version: 1\nname: x\nrules:\n  - match: {bundle: a}\n    display: {index: 1}\n    region: maximize\n    gap: -3\n", "gap must be non-negative"},
		{"bad tolerance", "version: 1\nname: x\nrules:\n  - match: {bundle: a}\n    display: {index: 1}\n    region: maximize\n    tolerance: wide\n", "tolerance"},
		{"bad order", "version: 1\nname: x\nrules:\n  - match: {bundle: a}\n    display: {index: 1}\n    region: maximize\n    order: size\n", "order"},
		{"bad alias word", "version: 1\nname: x\nrules:\n  - match: {bundle: a}\n    display: \"dell left\"\n    region: maximize\n", "want key=value"},
		{"bad display key", "version: 1\nname: x\ndisplays:\n  d:\n    colour: red\nrules: []\n", "unknown field"},
		{"bad display value", "version: 1\nname: x\ndisplays:\n  d:\n    index: -1\nrules: []\n", "displays.d"},
		{"numeric alias reserved", "version: 1\nname: x\ndisplays:\n  \"2\":\n    built-in: true\nrules: []\n", "numeric alias names are reserved"},
		{"bad match regex", "version: 1\nname: x\nrules:\n  - match: {title: /(?=x)/}\n    display: {index: 1}\n    region: maximize\n", "regex"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse([]byte(tc.src))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestLoadAndResolve(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "office.yaml")
	if err := os.WriteFile(path, []byte(Template("office")), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	extra := parseRules(t, "--match", "bundle=com.spotify.client", "--display", "laptop", "--region", "bottom-half")
	rules, err := p.Resolved(extra)
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 4 {
		t.Fatalf("got %d rules", len(rules))
	}
	// Aliases substituted everywhere, including the inline extra rule.
	if rules[0].Display.Alias != "" || rules[0].Display.Index != 1 {
		t.Errorf("rule 1 alias not substituted: %+v", rules[0].Display)
	}
	if rules[3].Display.Alias != "" || rules[3].Display.BuiltIn == nil {
		t.Errorf("extra rule alias not substituted: %+v", rules[3].Display)
	}
	// The profile's own rules are untouched (Resolved copies).
	if p.Rules[0].Display.Alias != "dell-left" {
		t.Errorf("profile mutated: %+v", p.Rules[0].Display)
	}
}

func TestResolveUnknownAlias(t *testing.T) {
	p, err := Parse([]byte(Template("office")))
	if err != nil {
		t.Fatal(err)
	}
	extra := parseRules(t, "--match", "bundle=x", "--display", "cinema", "--region", "maximize")
	_, err = p.Resolved(extra)
	if err == nil || !strings.Contains(err.Error(), `alias "cinema"`) || !strings.Contains(err.Error(), "dell-left, dell-right, laptop") {
		t.Fatalf("err = %v", err)
	}
}

func TestLoadErrors(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "missing.yaml")); err == nil {
		t.Error("missing file accepted")
	}
	path := filepath.Join(t.TempDir(), "broken.yaml")
	os.WriteFile(path, []byte("version: [\n"), 0o644)
	if _, err := Load(path); err == nil || !strings.Contains(err.Error(), "parse "+path) {
		t.Errorf("err = %v", err)
	}
}

// The G5 proof shape: a save on a commented profile changes only the
// appended rule — every comment line present before is present after.
func TestAppendPreservesComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "office.yaml")
	if err := WriteTemplate(path, "office", false); err != nil {
		t.Fatal(err)
	}
	// Normalize once: the first save re-marshals (dropping blank lines,
	// goccy#285); after that, saves are append-only.
	if err := Append(path, nil); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(path)

	added := parseRules(t, "--match", "bundle=com.spotify.client", "--display", "laptop", "--region", "bottom-half", "--gap", "4", "--tolerance", "2%", "--first", "--order", "pid")
	if err := Append(path, added); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(path)

	for _, line := range strings.Split(strings.TrimRight(string(before), "\n"), "\n") {
		if !strings.Contains(string(after), line) {
			t.Errorf("line lost by save: %q", line)
		}
	}
	for _, want := range []string{
		"# added by screenz profile save",
		"bundle: com.spotify.client",
		"display: laptop",
		"region: bottom-half",
		"gap: 4",
		"tolerance: 2%",
		"each: false",
		"order: pid",
	} {
		if !strings.Contains(string(after), want) {
			t.Errorf("appended rule missing %q\n%s", want, after)
		}
	}
	// And the appended file still parses with the same semantics.
	p, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	last := p.Rules[len(p.Rules)-1]
	if !last.First || last.Gap != 4 || last.Tolerance.String() != "2%" || last.Order != "pid" {
		t.Errorf("appended rule reparsed wrong: %+v", last)
	}
}

func TestAppendIsStableAfterNormalization(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "p.yaml")
	if err := WriteTemplate(path, "p", false); err != nil {
		t.Fatal(err)
	}
	if err := Append(path, nil); err != nil {
		t.Fatal(err)
	}
	first, _ := os.ReadFile(path)
	if err := Append(path, nil); err != nil {
		t.Fatal(err)
	}
	second, _ := os.ReadFile(path)
	if string(first) != string(second) {
		t.Errorf("empty append not idempotent:\n--- first\n%s\n--- second\n%s", first, second)
	}
}

func TestAppendErrors(t *testing.T) {
	if err := Append(filepath.Join(t.TempDir(), "missing.yaml"), nil); err == nil {
		t.Error("missing file accepted")
	}
	path := filepath.Join(t.TempDir(), "bad.yaml")
	os.WriteFile(path, []byte("version: 1\nname: x\ncolor: red\nrules: []\n"), 0o644)
	if err := Append(path, nil); err == nil || !strings.Contains(err.Error(), "parse") {
		t.Errorf("err = %v", err)
	}
	os.WriteFile(path, []byte("version: 3\nname: x\nrules: []\n"), 0o644)
	if err := Append(path, nil); err == nil || !strings.Contains(err.Error(), "unsupported profile version") {
		t.Errorf("err = %v", err)
	}
}

func TestWriteNewRoundTrips(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profiles", "laptop.yaml")
	rules := parseRules(t,
		"--match", `app="Google Chrome" title=/Work/`, "--display", "built-in=true", "--region", "left-half",
		"--match", "bundle=com.microsoft.VSCode", "--display", "name=/LU28R55/ index=2", "--region", "grid=2x2",
	)
	if err := WriteNew(path, "laptop", rules); err != nil {
		t.Fatal(err)
	}
	src, _ := os.ReadFile(path)
	if !strings.Contains(string(src), `# screenz profile "laptop"`) {
		t.Errorf("head comment missing:\n%s", src)
	}
	p, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "laptop" || len(p.Rules) != 2 {
		t.Fatalf("reparse wrong: %+v", p)
	}
	r1, r2 := p.Rules[0], p.Rules[1]
	if !r1.Match.Matches(mkWin("com.google.Chrome", "Google Chrome", "My Work tab")) {
		t.Errorf("quoted app + regex title lost: %+v", r1.Match)
	}
	if r1.Display.BuiltIn == nil || !*r1.Display.BuiltIn {
		t.Errorf("built-in lost: %+v", r1.Display)
	}
	if r2.Display.Index != 2 || !r2.Display.Name.Match("LU28R55 (2)") || r2.Region.String() != "grid=2x2" {
		t.Errorf("rule 2 lost: %+v", r2)
	}
}

func TestWriteNewAndTemplateErrors(t *testing.T) {
	dir := t.TempDir()
	blocking := filepath.Join(dir, "file")
	os.WriteFile(blocking, []byte("x"), 0o644)
	if err := WriteNew(filepath.Join(blocking, "sub", "p.yaml"), "p", nil); err == nil {
		t.Error("mkdir through a file accepted")
	}
	asDir := filepath.Join(dir, "target.yaml")
	os.MkdirAll(asDir, 0o755)
	if err := WriteNew(asDir, "p", nil); err == nil {
		t.Error("writing over a directory accepted")
	}
	if err := WriteTemplate(filepath.Join(blocking, "sub", "p.yaml"), "p", false); err == nil {
		t.Error("template mkdir through a file accepted")
	}
	path := filepath.Join(dir, "exists.yaml")
	os.WriteFile(path, []byte("x"), 0o644)
	if err := WriteTemplate(path, "p", false); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Errorf("err = %v", err)
	}
	if err := WriteTemplate(path, "p", true); err != nil {
		t.Errorf("force overwrite failed: %v", err)
	}
	// A non-exists open failure (target is a directory) surfaces as-is.
	if err := WriteTemplate(asDir, "p", true); err == nil || strings.Contains(err.Error(), "already exists") {
		t.Errorf("directory target: err = %v", err)
	}
}

// Pins every scalar and display encoding branch: numeric tolerance, uuid /
// serial / main display terms, inline display maps with unknown keys, and
// values that are not scalars at all.
func TestYAMLEncodingBranches(t *testing.T) {
	src := `version: 1
name: x
displays:
  exact:
    uuid: 37D8832A-2D66
    serial: 1129796439
    main: true
rules:
  - match: {bundle: a}
    display: exact
    region: maximize
    tolerance: 0.75
`
	p, err := Parse([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	spec := p.Displays["exact"]
	if !spec.UUID.Match("37D8832A-2D66") || spec.Serial != "1129796439" || spec.Main == nil || !*spec.Main {
		t.Errorf("spec lost terms: %+v", spec)
	}
	if p.Rules[0].Tolerance.String() != "0.75" {
		t.Errorf("numeric tolerance = %q", p.Rules[0].Tolerance)
	}

	if _, err := Parse([]byte("version: 1\nname: x\nrules:\n  - match: {bundle: a}\n    display: {colour: red}\n    region: maximize\n")); err == nil {
		t.Error("inline display with unknown key accepted")
	}
	if _, err := Parse([]byte("version: 1\nname: x\nrules:\n  - match: {bundle: a}\n    display: {index: 1}\n    region: maximize\n    tolerance: [1, 2]\n")); err == nil {
		t.Error("sequence tolerance accepted")
	}

	// Saving a uuid/serial/main spec round-trips it (specFromRule).
	dir := t.TempDir()
	path := filepath.Join(dir, "exact.yaml")
	rules := parseRules(t, "--match", "bundle=a", "--display", "uuid=37D8832A-2D66 serial=1129796439 main=true", "--region", "maximize", "--tolerance", "2")
	if err := WriteNew(path, "exact", rules); err != nil {
		t.Fatal(err)
	}
	p2, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	got := p2.Rules[0].Display
	if !got.UUID.Match("37D8832A-2D66") || got.Serial != "1129796439" || got.Main == nil || !*got.Main {
		t.Errorf("saved spec lost terms: %+v", got)
	}
	if p2.Rules[0].Tolerance.String() != "2" {
		t.Errorf("saved tolerance = %q", p2.Rules[0].Tolerance)
	}
}

// A slash-leading literal must survive the save/load round trip as a
// literal, not come back as a regex.
func TestSlashLeadingLiteralRoundTrips(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "p.yaml")
	rules := parseRules(t, "--match", `title="/foo"`, "--display", "index=1", "--region", "maximize")
	if err := WriteNew(path, "p", rules); err != nil {
		t.Fatal(err)
	}
	p, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	m := p.Rules[0].Match
	if !m.Matches(mkWin("b", "a", "/foo")) {
		t.Fatalf("literal /foo lost: %+v", m)
	}
	if m.Matches(mkWin("b", "a", "xx/fooyy")) {
		t.Fatal("reloaded as regex instead of exact literal")
	}
}

// WriteNew writes no displays map, so alias rules must be rejected loudly
// instead of producing a profile that can never apply.
func TestWriteNewRejectsAliasRules(t *testing.T) {
	rules := parseRules(t, "--match", "bundle=a", "--display", "laptop", "--region", "maximize")
	err := WriteNew(filepath.Join(t.TempDir(), "p.yaml"), "p", rules)
	if err == nil || !strings.Contains(err.Error(), `alias "laptop"`) {
		t.Fatalf("err = %v", err)
	}
}

func TestWriteFileAtomicErrors(t *testing.T) {
	dir := t.TempDir()
	locked := filepath.Join(dir, "locked")
	if err := os.Mkdir(locked, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(locked, 0o755) })
	rules := parseRules(t, "--match", "bundle=a", "--display", "index=1", "--region", "maximize")
	if err := WriteNew(filepath.Join(locked, "p.yaml"), "p", rules); err == nil {
		t.Fatal("write into a read-only directory accepted")
	}
}

func TestPath(t *testing.T) {
	env := func(k string) string {
		if k == "SCREENZ_HOME" {
			return "/dot/screenz"
		}
		return ""
	}
	if got := Path(env, "/Users/example", "office"); got != "/dot/screenz/profiles/office.yaml" {
		t.Fatalf("Path = %q", got)
	}
}
