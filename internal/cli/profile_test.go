package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joshpeak/screenz/internal/discover"
)

// profDeps points SCREENZ_HOME at a fresh directory and serves the office
// displays, so profile files are real files on disk.
func profDeps(t *testing.T) (Deps, string) {
	t.Helper()
	home := t.TempDir()
	d := applyDeps("ok")
	d.Getenv = func(k string) string {
		if k == "SCREENZ_HOME" {
			return home
		}
		return ""
	}
	d.Displays = func() ([]discover.Display, error) { return officeSnapshot().Displays, nil }
	return d, home
}

func writeProfile(t *testing.T, home, name, content string) string {
	t.Helper()
	path := filepath.Join(home, "profiles", name+".yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// officeProfileYAML targets the test displays: laptop resolves, cinema does
// not, and panel is ambiguous between the two office displays.
const officeProfileYAML = `version: 1
name: office
displays:
  laptop:
    built-in: true
  cinema:
    name: /XDR/
  panel:
    index: 1
rules:
  - match:
      bundle: com.microsoft.VSCode
    display: laptop
    region: maximize
`

func TestProfileDispatch(t *testing.T) {
	d, _ := profDeps(t)
	cases := []struct {
		name     string
		args     []string
		wantCode int
		wantOut  string
		wantErr  string
	}{
		{"no subcommand", []string{"profile"}, 2, "", "usage: screenz profile"},
		{"unknown subcommand", []string{"profile", "list"}, 2, "", `unknown command "list"`},
		{"help", []string{"profile", "--help"}, 0, "usage: screenz profile", ""},
		{"help word", []string{"profile", "help"}, 0, "usage: screenz profile", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, out, errOut := run(t, tc.args, d)
			if code != tc.wantCode {
				t.Fatalf("exit = %d, want %d", code, tc.wantCode)
			}
			if tc.wantOut != "" && !strings.Contains(out, tc.wantOut) {
				t.Errorf("stdout missing %q:\n%s", tc.wantOut, out)
			}
			if tc.wantErr != "" && !strings.Contains(errOut, tc.wantErr) {
				t.Errorf("stderr missing %q:\n%s", tc.wantErr, errOut)
			}
		})
	}
}

func TestProfileInit(t *testing.T) {
	d, home := profDeps(t)
	code, out, _ := run(t, []string{"profile", "init", "office"}, d)
	if code != 0 || !strings.Contains(out, "wrote ") {
		t.Fatalf("exit=%d out=%q", code, out)
	}
	src, err := os.ReadFile(filepath.Join(home, "profiles", "office.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(src), `# screenz profile "office"`) {
		t.Errorf("template header missing:\n%s", src)
	}
	// Second init refuses without --force, succeeds with it.
	code, _, errOut := run(t, []string{"profile", "init", "office"}, d)
	if code != 1 || !strings.Contains(errOut, "already exists") {
		t.Fatalf("exit=%d stderr=%q", code, errOut)
	}
	if code, _, _ = run(t, []string{"profile", "init", "office", "--force"}, d); code != 0 {
		t.Fatalf("force init exit=%d", code)
	}
}

func TestProfileInitUsage(t *testing.T) {
	d, _ := profDeps(t)
	if code, _, _ := run(t, []string{"profile", "init"}, d); code != 2 {
		t.Fatalf("no name: exit=%d", code)
	}
	if code, _, _ := run(t, []string{"profile", "init", "a", "b"}, d); code != 2 {
		t.Fatalf("two names: exit=%d", code)
	}
	if code, _, _ := run(t, []string{"profile", "init", "--nope", "a"}, d); code != 2 {
		t.Fatalf("bad flag: exit=%d", code)
	}
	if code, out, _ := run(t, []string{"profile", "init", "--help"}, d); code != 0 || !strings.Contains(out, "usage:") {
		t.Fatalf("help: exit=%d", code)
	}
}

func TestProfileSaveNewAppendAndForce(t *testing.T) {
	d, home := profDeps(t)
	save := []string{"profile", "save", "office", "--match", "bundle=com.microsoft.VSCode", "--display", "index=2", "--region", "maximize"}
	code, out, errOut := run(t, save, d)
	if code != 0 || !strings.Contains(out, "saved 1 rule(s)") {
		t.Fatalf("exit=%d out=%q err=%q", code, out, errOut)
	}
	path := filepath.Join(home, "profiles", "office.yaml")

	// Appending preserves what is already there and marks the new rule.
	code, _, _ = run(t, []string{"profile", "save", "office", "--match", "bundle=com.google.Chrome", "--display", "index=1", "--region", "left-half"}, d)
	if code != 0 {
		t.Fatalf("append exit=%d", code)
	}
	src, _ := os.ReadFile(path)
	for _, want := range []string{"com.microsoft.VSCode", "com.google.Chrome", "# added by screenz profile save"} {
		if !strings.Contains(string(src), want) {
			t.Errorf("saved profile missing %q:\n%s", want, src)
		}
	}

	// --force rewrites with only the given rules.
	code, _, _ = run(t, []string{"profile", "save", "office", "--force", "--match", "bundle=com.spotify.client", "--display", "index=1", "--region", "maximize"}, d)
	if code != 0 {
		t.Fatalf("force exit=%d", code)
	}
	src, _ = os.ReadFile(path)
	if strings.Contains(string(src), "VSCode") || !strings.Contains(string(src), "com.spotify.client") {
		t.Errorf("force did not rewrite:\n%s", src)
	}
}

func TestProfileSaveUsageAndIOErrors(t *testing.T) {
	d, home := profDeps(t)
	cases := []struct {
		name string
		args []string
	}{
		{"no name", []string{"profile", "save"}},
		{"flag before name", []string{"profile", "save", "--force", "office"}},
		{"no rules", []string{"profile", "save", "office"}},
		{"incomplete rule", []string{"profile", "save", "office", "--match", "bundle=a"}},
		{"trailing positional", []string{"profile", "save", "office", "--match", "bundle=a", "--display", "index=1", "--region", "maximize", "extra"}},
		{"bad flag", []string{"profile", "save", "office", "--nope"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if code, _, _ := run(t, tc.args, d); code != 2 {
				t.Fatalf("exit = %d, want 2", code)
			}
		})
	}
	if code, out, _ := run(t, []string{"profile", "save", "office", "--help"}, d); code != 0 || !strings.Contains(out, "usage:") {
		t.Fatalf("help exit=%d", code)
	}
	if code, out, _ := run(t, []string{"profile", "save", "--help"}, d); code != 0 || !strings.Contains(out, "usage:") {
		t.Fatalf("bare help exit=%d", code)
	}
	// Saving alias rules into a NEW profile is refused (no displays map).
	code, _, errOut := run(t, []string{"profile", "save", "fresh", "--match", "bundle=a", "--display", "laptop", "--region", "maximize"}, d)
	if code != 1 || !strings.Contains(errOut, `alias "laptop"`) {
		t.Fatalf("alias to new profile: exit=%d stderr=%q", code, errOut)
	}
	// Corrupt existing file: append fails with exit 1.
	writeProfile(t, home, "broken", "version: 1\nname: broken\ncolor: red\nrules: []\n")
	code, _, errOut = run(t, []string{"profile", "save", "broken", "--match", "bundle=a", "--display", "index=1", "--region", "maximize"}, d)
	if code != 1 || !strings.Contains(errOut, "parse") {
		t.Fatalf("exit=%d stderr=%q", code, errOut)
	}
}

func TestProfileStatus(t *testing.T) {
	d, home := profDeps(t)
	writeProfile(t, home, "office", officeProfileYAML)
	writeProfile(t, home, "broken", "version: [\n")

	code, out, _ := run(t, []string{"profile", "status"}, d)
	if code != 0 {
		t.Fatalf("exit=%d", code)
	}
	for _, want := range []string{
		"PROFILE", "RULES", "ALIAS", "RESOLVES", "DETAIL",
		"laptop", "true", "index=1 Built-in Retina Display",
		"cinema", "no connected display matches",
		"panel", "index=1 Built-in Retina Display",
		"broken", "error:",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout missing %q\n%s", want, out)
		}
	}
}

func TestProfileStatusAmbiguousAndNamed(t *testing.T) {
	d, home := profDeps(t)
	writeProfile(t, home, "twins", `version: 1
name: twins
displays:
  either:
    name: /[A-Z]/
rules: []
`)
	code, out, _ := run(t, []string{"profile", "status", "twins"}, d)
	if code != 0 {
		t.Fatalf("exit=%d", code)
	}
	if !strings.Contains(out, "ambiguous: matches 2 displays") {
		t.Errorf("stdout missing ambiguity:\n%s", out)
	}
}

func TestProfileStatusJSONAndEdgeCases(t *testing.T) {
	d, home := profDeps(t)
	code, out, _ := run(t, []string{"profile", "status"}, d)
	if code != 0 || !strings.Contains(out, "no profiles in") {
		t.Fatalf("empty dir: exit=%d out=%q", code, out)
	}
	writeProfile(t, home, "norules", "version: 1\nname: norules\nrules: []\n")
	code, out, _ = run(t, []string{"profile", "status", "--json"}, d)
	if code != 0 {
		t.Fatalf("json exit=%d", code)
	}
	var rep struct {
		Schema   int `json:"schema"`
		Profiles []struct {
			Name  string `json:"name"`
			Rules int    `json:"rules"`
		} `json:"profiles"`
	}
	if err := json.Unmarshal([]byte(out), &rep); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, out)
	}
	if rep.Schema != 1 || len(rep.Profiles) != 1 || rep.Profiles[0].Name != "norules" {
		t.Errorf("unexpected: %+v", rep)
	}
	// Table form for a profile with no aliases.
	code, out, _ = run(t, []string{"profile", "status"}, d)
	if code != 0 || !strings.Contains(out, "norules") {
		t.Fatalf("table exit=%d", code)
	}
	if code, _, _ := run(t, []string{"profile", "status", "a", "b"}, d); code != 2 {
		t.Fatal("two names accepted")
	}
	if code, _, _ := run(t, []string{"profile", "status", "--nope"}, d); code != 2 {
		t.Fatal("bad flag accepted")
	}
	if code, out, _ := run(t, []string{"profile", "status", "--help"}, d); code != 0 || !strings.Contains(out, "usage:") {
		t.Fatal("help failed")
	}
	d.Displays = func() ([]discover.Display, error) { return nil, errStub("no displays") }
	if code, _, errOut := run(t, []string{"profile", "status"}, d); code != 1 || !strings.Contains(errOut, "no displays") {
		t.Fatal("displays error not surfaced")
	}
}

func TestProfileStatusReadDirError(t *testing.T) {
	d, home := profDeps(t)
	// Make the profiles path a file: ReadDir fails with ENOTDIR, which
	// must surface instead of masquerading as "no profiles".
	if err := os.WriteFile(filepath.Join(home, "profiles"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, errOut := run(t, []string{"profile", "status"}, d)
	if code != 1 || !strings.Contains(errOut, "profile status:") {
		t.Fatalf("exit=%d stderr=%q", code, errOut)
	}
}

func TestApplyWithProfile(t *testing.T) {
	d, home := profDeps(t)
	writeProfile(t, home, "office", officeProfileYAML)
	// The laptop alias resolves to the built-in (index 1 in the office
	// snapshot); the VS Code repo window moves there.
	code, out, _ := run(t, []string{"apply", "office", "--dry-run"}, d)
	if code != 0 {
		t.Fatalf("exit=%d\n%s", code, out)
	}
	if !strings.Contains(out, "repo — main.go") || !strings.Contains(out, "skipped: offscreen") {
		t.Errorf("plan wrong:\n%s", out)
	}
	// Inline rules append after profile rules and may use profile aliases.
	code, out, _ = run(t, []string{"apply", "office", "--dry-run",
		"--match", "bundle=com.google.Chrome", "--display", "laptop", "--region", "right-half"}, d)
	if code != 0 {
		t.Fatalf("exit=%d\n%s", code, out)
	}
	if !strings.Contains(out, "Work — Inbox") {
		t.Errorf("inline rule missing from plan:\n%s", out)
	}
	// Name may follow the flags too: apply --dry-run office.
	code, _, _ = run(t, []string{"apply", "--dry-run", "office"}, d)
	if code != 0 {
		t.Fatalf("trailing name exit=%d", code)
	}
}

func TestApplyProfileErrors(t *testing.T) {
	d, home := profDeps(t)
	if code, _, errOut := run(t, []string{"apply", "nowhere"}, d); code != 1 || !strings.Contains(errOut, "no such file") {
		t.Fatalf("missing profile: exit=%d stderr=%q", code, errOut)
	}
	// An alias that is not defined fails before any window is touched
	// (Place would t.Fatal via the dry-run guard if it ran).
	writeProfile(t, home, "office", officeProfileYAML)
	code, _, errOut := run(t, []string{"apply", "office",
		"--match", "bundle=x", "--display", "dell-right", "--region", "maximize"}, d)
	if code != 1 || !strings.Contains(errOut, `alias "dell-right"`) {
		t.Fatalf("unresolved alias: exit=%d stderr=%q", code, errOut)
	}
	// Two positionals are a usage error.
	if code, _, _ := run(t, []string{"apply", "office", "second"}, d); code != 2 {
		t.Fatal("two names accepted")
	}
}
