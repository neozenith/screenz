package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neozenith/screenz/internal/discover"
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

// officeProfileYAML exercises every alias verdict against the test
// displays ("Built-in Retina Display" and "LU28R55 (1)"): laptop and panel
// resolve, cinema matches nothing, conflicted ANDs two terms that pick
// different displays, and panels matches both.
const officeProfileYAML = `version: 1
name: office
displays:
  laptop:
    built-in: true
  cinema:
    name: /XDR/
  panel:
    index: 1
  panels:
    name: /5|Display/
  conflicted:
    name: /LU28R55/
    index: 1
rules:
  - match:
      bundle: com.microsoft.VSCode
    display: laptop
    region: maximize
`

// fittingProfileYAML declares only aliases that resolve on the test
// displays, so the whole profile earns a tick.
const fittingProfileYAML = `version: 1
name: fits
displays:
  laptop:
    built-in: true
rules:
  - match:
      bundle: com.microsoft.VSCode
    display: laptop
    region: maximize
`

func TestListVerdictPerProfile(t *testing.T) {
	d, home := profDeps(t)
	writeProfile(t, home, "office", officeProfileYAML)
	writeProfile(t, home, "fits", fittingProfileYAML)

	code, out, errOut := run(t, []string{"list"}, d)
	if code != 0 {
		t.Fatalf("exit = %d; stderr: %s", code, errOut)
	}
	for _, want := range []string{
		"FITS", "PROFILE", "RULES", "PATH", "DETAIL",
		"office.yaml", "fits.yaml",
		"unresolved: cinema, conflicted, panels", // named, not explained
		"list --verbose",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout missing %q\n%s", want, out)
		}
	}
	// One verdict per profile, not one per alias: office declares four
	// aliases and must still occupy a single row.
	rows := 0
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "office") {
			rows++
		}
	}
	if rows != 1 {
		t.Errorf("want office on exactly one row, got %d\n%s", rows, out)
	}
	if !strings.Contains(out, "✓") || !strings.Contains(out, "✗") {
		t.Errorf("want both verdicts present\n%s", out)
	}
	// The long explanation is verbose-only: it would push PATH off-screen.
	if strings.Contains(out, "remove or fix the conflicting term") {
		t.Errorf("non-verbose list leaked the full explanation:\n%s", out)
	}
}

// A profile with no aliases is machine-independent, so it fits trivially.
func TestListNoAliasesFits(t *testing.T) {
	d, home := profDeps(t)
	writeProfile(t, home, "inline", "version: 1\nname: inline\nrules:\n  - match: {bundle: a}\n    display: {index: 1}\n")
	code, out, _ := run(t, []string{"list"}, d)
	if code != 0 || !strings.Contains(out, "✓") {
		t.Fatalf("exit=%d out=%q", code, out)
	}
	if strings.Contains(out, "list --verbose") {
		t.Errorf("the why-not hint appears when everything fits:\n%s", out)
	}
}

func TestListVerbose(t *testing.T) {
	d, home := profDeps(t)
	writeProfile(t, home, "office", officeProfileYAML)
	writeProfile(t, home, "inline", "version: 1\nname: inline\nrules:\n  - match: {bundle: a}\n    display: {index: 1}\n")

	code, out, _ := run(t, []string{"list", "--verbose"}, d)
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	for _, want := range []string{
		"ALIAS", "RESOLVES",
		"cinema", "no connected display matches",
		"conflicted", "remove or fix the conflicting term",
		"panels", "ambiguous",
		"laptop", "Built-in Retina Display",
		"no aliases; displays addressed inline",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout missing %q\n%s", want, out)
		}
	}
}

// A NAME may be followed by flags: stdlib flag stops at the first
// positional, so the name is peeled off before parsing.
func TestListOneProfileWithFlagsAfterTheName(t *testing.T) {
	d, home := profDeps(t)
	writeProfile(t, home, "office", officeProfileYAML)
	writeProfile(t, home, "fits", fittingProfileYAML)

	code, out, errOut := run(t, []string{"list", "office", "-v"}, d)
	if code != 0 {
		t.Fatalf("exit = %d; stderr: %s", code, errOut)
	}
	if !strings.Contains(out, "ALIAS") {
		t.Errorf("-v after the name was not read as a flag:\n%s", out)
	}
	if strings.Contains(out, "fits") {
		t.Errorf("naming one profile still listed the others:\n%s", out)
	}
}

func TestListJSON(t *testing.T) {
	d, home := profDeps(t)
	writeProfile(t, home, "office", officeProfileYAML)

	code, out, _ := run(t, []string{"list", "--json"}, d)
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	var got struct {
		Schema   int             `json:"schema"`
		Profiles []profileStatus `json:"profiles"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, out)
	}
	if got.Schema != 1 || len(got.Profiles) != 1 {
		t.Fatalf("got %+v", got)
	}
	p := got.Profiles[0]
	if p.Name != "office" || p.Rules != 1 || p.Fits {
		t.Errorf("profile summary wrong: %+v", p)
	}
	if len(p.Aliases) != 5 {
		t.Errorf("want 5 aliases, got %d", len(p.Aliases))
	}
}

func TestListEmptyAndErrors(t *testing.T) {
	d, home := profDeps(t)
	if code, out, _ := run(t, []string{"list"}, d); code != 0 || !strings.Contains(out, "no profiles in") {
		t.Fatalf("empty dir: exit=%d out=%q", code, out)
	}

	// A profile that will not parse is reported, not skipped, and never fits.
	writeProfile(t, home, "broken", "version: 9\nname: broken\nrules: []\n")
	code, out, _ := run(t, []string{"list"}, d)
	if code != 0 || !strings.Contains(out, "will not load") || !strings.Contains(out, "✗") {
		t.Fatalf("broken profile: exit=%d out=%q", code, out)
	}
	code, out, _ = run(t, []string{"list", "--verbose"}, d)
	if code != 0 || !strings.Contains(out, "error:") {
		t.Fatalf("broken profile verbose: exit=%d out=%q", code, out)
	}
}

func TestListUsageAndDisplayError(t *testing.T) {
	d, _ := profDeps(t)
	if code, _, errOut := run(t, []string{"list", "a", "b"}, d); code != 2 || !strings.Contains(errOut, "at most one NAME") {
		t.Fatalf("two names: exit=%d err=%q", code, errOut)
	}
	if code, _, _ := run(t, []string{"list", "--nope"}, d); code != 2 {
		t.Fatalf("bad flag: exit=%d", code)
	}
	if code, out, _ := run(t, []string{"list", "--help"}, d); code != 0 || !strings.Contains(out, "usage: screenz list") {
		t.Fatalf("help: exit=%d", code)
	}
	d.Displays = func() ([]discover.Display, error) { return nil, errStub("no displays") }
	if code, _, errOut := run(t, []string{"list"}, d); code != 1 || !strings.Contains(errOut, "no displays") {
		t.Fatalf("display error: exit=%d err=%q", code, errOut)
	}
}

// A profile directory that cannot be read is a real failure, not an empty
// list: reporting "no profiles" would hide it.
func TestListUnreadableDir(t *testing.T) {
	d, home := profDeps(t)
	dir := filepath.Join(home, "profiles")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
	if code, _, errOut := run(t, []string{"list"}, d); code != 1 || errOut == "" {
		t.Fatalf("exit=%d err=%q", code, errOut)
	}
}
