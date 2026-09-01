package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitWritesTemplate(t *testing.T) {
	d, home := profDeps(t)
	code, out, errOut := run(t, []string{"init", "--profile", "office"}, d)
	if code != 0 || !strings.Contains(out, "wrote ") {
		t.Fatalf("exit=%d out=%q err=%q", code, out, errOut)
	}
	src, err := os.ReadFile(filepath.Join(home, "profiles", "office.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	// The template teaches the current grammar, or it teaches a CLI that
	// no longer exists (ADR-0025).
	for _, want := range []string{
		`# screenz profile "office"`,
		"screenz init --profile office",
		"screenz apply --profile office",
		"screenz apply --dry-run --profile office",
	} {
		if !strings.Contains(string(src), want) {
			t.Errorf("template missing %q:\n%s", want, src)
		}
	}
}

func TestInitRefusesToClobberWithoutForce(t *testing.T) {
	d, _ := profDeps(t)
	if code, _, _ := run(t, []string{"init", "-p", "office"}, d); code != 0 {
		t.Fatalf("first init exit=%d", code)
	}
	code, _, errOut := run(t, []string{"init", "-p", "office"}, d)
	if code != 1 || !strings.Contains(errOut, "already exists") {
		t.Fatalf("second init: exit=%d stderr=%q", code, errOut)
	}
	if code, _, _ = run(t, []string{"init", "-p", "office", "--force"}, d); code != 0 {
		t.Fatalf("force init exit=%d", code)
	}
}

func TestInitUsage(t *testing.T) {
	d, _ := profDeps(t)
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"no name", []string{"init"}, "--profile NAME is required"},
		// The old grammar took the name positionally; it must fail loudly
		// and say what to type, not write a profile called "".
		{"positional name", []string{"init", "office"}, "name the profile with --profile office"},
		{"bad flag", []string{"init", "--nope"}, "flag provided but not defined"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, out, errOut := run(t, tc.args, d)
			if code != 2 {
				t.Fatalf("exit = %d, want 2", code)
			}
			if out != "" {
				t.Errorf("usage error wrote stdout: %q", out)
			}
			if !strings.Contains(errOut, tc.want) {
				t.Errorf("stderr missing %q:\n%s", tc.want, errOut)
			}
		})
	}
	if code, out, _ := run(t, []string{"init", "--help"}, d); code != 0 || !strings.Contains(out, "usage: screenz init") {
		t.Fatalf("help: exit=%d out=%q", code, out)
	}
}

// The template directory is created on demand; a path that cannot be
// written is a runtime failure, not a usage error.
func TestInitWriteFailure(t *testing.T) {
	d, home := profDeps(t)
	if err := os.WriteFile(filepath.Join(home, "profiles"), []byte("not a dir"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, errOut := run(t, []string{"init", "-p", "office"}, d); code != 1 || errOut == "" {
		t.Fatalf("exit=%d err=%q", code, errOut)
	}
}
