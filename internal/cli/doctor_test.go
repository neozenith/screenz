package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// The real machine facts captured by the 2026-08-22 spike: three displays,
// two of them identically named Samsung panels, launched from a terminal app.
func officeSys(trusted bool) func(bool) SysInfo {
	return func(prompt bool) SysInfo {
		return SysInfo{
			Trusted:       trusted,
			HostAppName:   "iTerm2",
			HostAppBundle: "com.googlecode.iterm2",
			DisplayNames:  []string{"Built-in Retina Display", "LU28R55 (1)", "LU28R55 (2)"},
		}
	}
}

func deps(sys func(bool) SysInfo) Deps {
	return Deps{
		Getenv: func(string) string { return "" },
		Home:   "/Users/example",
		Sys:    sys,
	}
}

func run(t *testing.T, args []string, d Deps) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := Run(args, &stdout, &stderr, d)
	return code, stdout.String(), stderr.String()
}

func TestDoctorDisclosesDemoMode(t *testing.T) {
	d := deps(officeSys(true))
	d.Getenv = func(key string) string {
		if key == "SCREENZ_DEMO" {
			return "demo-world.json"
		}
		return ""
	}
	code, out, _ := run(t, []string{"doctor"}, d)
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if !strings.Contains(out, "demo mode: replaying demo-world.json; placement simulated") {
		t.Errorf("doctor must disclose demo mode\n%s", out)
	}
	_, jsonOut, _ := run(t, []string{"doctor", "--json"}, d)
	if !strings.Contains(jsonOut, `"demo_world": "demo-world.json"`) {
		t.Errorf("doctor --json must carry demo_world\n%s", jsonOut)
	}
}

func TestDoctorTrustedTable(t *testing.T) {
	code, out, errOut := run(t, []string{"doctor"}, deps(officeSys(true)))
	if code != 0 {
		t.Fatalf("exit = %d, want 0; stderr: %s", code, errOut)
	}
	for _, want := range []string{
		"accessibility: trusted",
		"host app: iTerm2 (com.googlecode.iterm2)",
		"displays: 3",
		"  - LU28R55 (2)",
		"profile dir: /Users/example/.config/screenz",
		"symbols: ok",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout missing %q\n%s", want, out)
		}
	}
	if errOut != "" {
		t.Errorf("stderr not empty: %s", errOut)
	}
}

func TestDoctorUntrustedExits1WithGrantInstruction(t *testing.T) {
	code, out, errOut := run(t, []string{"doctor"}, deps(officeSys(false)))
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(out, "accessibility: untrusted") {
		t.Errorf("stdout missing untrusted line\n%s", out)
	}
	for _, want := range []string{
		"Accessibility is NOT granted for iTerm2 (com.googlecode.iterm2)",
		"System Settings → Privacy & Security → Accessibility",
		"tccutil reset Accessibility com.googlecode.iterm2",
	} {
		if !strings.Contains(errOut, want) {
			t.Errorf("stderr missing %q\n%s", want, errOut)
		}
	}
}

func TestDoctorUntrustedUnknownHost(t *testing.T) {
	sys := func(bool) SysInfo { return SysInfo{} }
	code, out, errOut := run(t, []string{"doctor"}, deps(sys))
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(out, "host app: unknown") {
		t.Errorf("stdout missing unknown host\n%s", out)
	}
	for _, want := range []string{"your terminal app", "tccutil reset Accessibility the terminal app"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("stderr missing %q\n%s", want, errOut)
		}
	}
}

func TestDoctorJSON(t *testing.T) {
	code, out, _ := run(t, []string{"doctor", "--json"}, deps(officeSys(true)))
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	var rep struct {
		Schema        int      `json:"schema"`
		Accessibility string   `json:"accessibility"`
		HostAppBundle string   `json:"host_app_bundle"`
		Displays      []string `json:"displays"`
		ProfileDir    string   `json:"profile_dir"`
	}
	if err := json.Unmarshal([]byte(out), &rep); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, out)
	}
	if rep.Schema != 1 || rep.Accessibility != "trusted" || rep.HostAppBundle != "com.googlecode.iterm2" {
		t.Errorf("unexpected report: %+v", rep)
	}
	if len(rep.Displays) != 3 || rep.ProfileDir != "/Users/example/.config/screenz" {
		t.Errorf("unexpected report: %+v", rep)
	}
}

func TestDoctorMissingSymbols(t *testing.T) {
	sys := func(bool) SysInfo {
		return SysInfo{MissingSymbols: []string{"CGDisplayCreateUUIDFromDisplayID"}}
	}
	code, out, _ := run(t, []string{"doctor"}, deps(sys))
	if code != 1 {
		t.Fatalf("exit = %d, want 1 (untrusted because nothing could be checked)", code)
	}
	for _, want := range []string{"symbols: 1 missing", "  - CGDisplayCreateUUIDFromDisplayID"} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout missing %q\n%s", want, out)
		}
	}
}

func TestDoctorHelp(t *testing.T) {
	code, out, errOut := run(t, []string{"doctor", "--help"}, deps(officeSys(true)))
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if !strings.Contains(out, "usage: screenz doctor") {
		t.Errorf("help not on stdout:\n%s", out)
	}
	if errOut != "" {
		t.Errorf("stderr not empty: %s", errOut)
	}
}

func TestDoctorBadFlag(t *testing.T) {
	code, out, errOut := run(t, []string{"doctor", "--nope"}, deps(officeSys(true)))
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if out != "" {
		t.Errorf("usage error must not write stdout:\n%s", out)
	}
	if !strings.Contains(errOut, "screenz doctor:") || !strings.Contains(errOut, "usage: screenz doctor") {
		t.Errorf("stderr missing error + usage:\n%s", errOut)
	}
}

func TestVersionCommand(t *testing.T) {
	for _, args := range [][]string{{"version"}, {"--version"}} {
		code, out, errOut := run(t, args, deps(officeSys(true)))
		if code != 0 || !strings.Contains(out, "screenz dev (commit none, built unknown by source)") {
			t.Errorf("%v: exit=%d out=%q err=%q", args, code, out, errOut)
		}
	}
}

func TestOSVersionParts(t *testing.T) {
	cases := []struct {
		in           string
		major, minor int
		ok           bool
	}{
		{"26.6.2", 26, 6, true},
		{"13.0", 13, 0, true},
		{"15", 15, 0, true},
		{"", 0, 0, false},
		{"beta.1", 0, 0, false},
		{"26.x", 0, 0, false},
	}
	for _, tc := range cases {
		major, minor, ok := osVersionParts(tc.in)
		if major != tc.major || minor != tc.minor || ok != tc.ok {
			t.Errorf("osVersionParts(%q) = %d,%d,%v", tc.in, major, minor, ok)
		}
	}
}

// The install-runbook doctor caveats on the real machine shapes: the macOS 13 floor, the
// 26.1+ hidden-grant behavior, and a quarantined browser download.
func TestDoctorPlatformWarnings(t *testing.T) {
	base := officeSys(true)(true)
	base.OSVersion = "26.6.2"
	base.ExePath = "/Users/example/Downloads/screenz"

	t.Run("26.1 hidden grant caveat", func(t *testing.T) {
		info := base
		sys := func(bool) SysInfo { return info }
		code, out, _ := run(t, []string{"doctor"}, deps(sys))
		if code != 0 {
			t.Fatalf("exit=%d", code)
		}
		for _, want := range []string{
			"macos: 26.6.2",
			"binary: /Users/example/Downloads/screenz",
			"macOS >= 26.1 can enforce the Accessibility grant while hiding it",
			"tccutil reset Accessibility com.googlecode.iterm2",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("stdout missing %q\n%s", want, out)
			}
		}
	})
	t.Run("quarantine names the exact command", func(t *testing.T) {
		info := base
		info.Quarantined = true
		sys := func(bool) SysInfo { return info }
		_, out, _ := run(t, []string{"doctor"}, deps(sys))
		if !strings.Contains(out, "xattr -d com.apple.quarantine /Users/example/Downloads/screenz") {
			t.Errorf("stdout missing xattr command\n%s", out)
		}
	})
	t.Run("macOS 13 floor", func(t *testing.T) {
		info := base
		info.OSVersion = "12.7.4"
		sys := func(bool) SysInfo { return info }
		_, out, _ := run(t, []string{"doctor"}, deps(sys))
		if !strings.Contains(out, "below the supported floor (13+)") {
			t.Errorf("stdout missing floor warning\n%s", out)
		}
	})
	t.Run("json carries the warnings", func(t *testing.T) {
		info := base
		info.Quarantined = true
		sys := func(bool) SysInfo { return info }
		_, out, _ := run(t, []string{"doctor", "--json"}, deps(sys))
		var rep struct {
			OSVersion   string   `json:"os_version"`
			Quarantined bool     `json:"quarantined"`
			Warnings    []string `json:"warnings"`
		}
		if err := json.Unmarshal([]byte(out), &rep); err != nil {
			t.Fatalf("not JSON: %v", err)
		}
		if rep.OSVersion != "26.6.2" || !rep.Quarantined || len(rep.Warnings) != 2 {
			t.Errorf("unexpected: %+v", rep)
		}
	})
}

func TestRunDispatch(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		wantCode int
		wantOut  string // substring of stdout
		wantErr  string // substring of stderr
	}{
		{"no args", nil, 2, "", "no command given"},
		{"unknown command", []string{"list"}, 2, "", `unknown command "list"`},
		{"help flag", []string{"--help"}, 0, "usage: screenz", ""},
		{"help word", []string{"help"}, 0, "usage: screenz", ""},
		{"dash h", []string{"-h"}, 0, "usage: screenz", ""},
		// Each command answers to its initial too (ADR-0024); -h on the
		// short form proves it reached the same handler, not a near miss.
		{"a is apply", []string{"a", "-h"}, 0, "usage: screenz apply", ""},
		{"d is doctor", []string{"d", "-h"}, 0, "usage: screenz doctor", ""},
		{"s is status", []string{"s", "-h"}, 0, "usage: screenz status", ""},
		{"p is profile", []string{"p", "-h"}, 0, "usage: screenz profile", ""},
		{"u is update", []string{"u", "-h"}, 0, "usage: screenz update", ""},
		{"v is version", []string{"v"}, 0, "screenz", ""},
		{"h is help", []string{"h"}, 0, "usage: screenz", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, out, errOut := run(t, tc.args, deps(officeSys(true)))
			if code != tc.wantCode {
				t.Fatalf("exit = %d, want %d", code, tc.wantCode)
			}
			if tc.wantOut != "" && !strings.Contains(out, tc.wantOut) {
				t.Errorf("stdout missing %q:\n%s", tc.wantOut, out)
			}
			if tc.wantErr != "" && !strings.Contains(errOut, tc.wantErr) {
				t.Errorf("stderr missing %q:\n%s", tc.wantErr, errOut)
			}
			if tc.wantErr == "" && errOut != "" {
				t.Errorf("stderr not empty: %s", errOut)
			}
		})
	}
}
