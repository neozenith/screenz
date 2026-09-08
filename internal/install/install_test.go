package install

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// binary writes a stand-in screenz executable and returns its path.
func binary(t *testing.T) string {
	t.Helper()
	exe := filepath.Join(t.TempDir(), "screenz")
	if err := os.WriteFile(exe, []byte("binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	return exe
}

// The short link is created relative, reported as linked afterwards, and
// creating it twice is a no-op rather than an error.
func TestCreateLinkIsIdempotent(t *testing.T) {
	exe := binary(t)
	if state, target := LinkState(exe); state != Absent || target != "" {
		t.Fatalf("before: state=%q target=%q", state, target)
	}
	path, created, err := CreateLink(exe)
	if err != nil || !created || path != LinkPath(exe) {
		t.Fatalf("create: path=%q created=%v err=%v", path, created, err)
	}
	// Relative, so moving the install directory keeps the pair intact.
	if target, _ := os.Readlink(path); target != "screenz" {
		t.Errorf("target = %q, want screenz", target)
	}
	if state, target := LinkState(exe); state != Linked || target != "screenz" {
		t.Errorf("after: state=%q target=%q", state, target)
	}
	if _, created, err := CreateLink(exe); err != nil || created {
		t.Errorf("second create: created=%v err=%v", created, err)
	}
	// An absolute target from an older screenz still reads as ours.
	os.Remove(path)
	if err := os.Symlink(exe, path); err != nil {
		t.Fatal(err)
	}
	if state, _ := LinkState(exe); state != Linked {
		t.Errorf("absolute target: state = %q, want linked", state)
	}
}

// A name something else holds is never taken: not a symlink elsewhere
// (lrzsz ships an sz), and not a regular file.
func TestCreateLinkRefusesForeignNames(t *testing.T) {
	t.Run("symlink elsewhere", func(t *testing.T) {
		exe := binary(t)
		if err := os.Symlink("/usr/bin/lrzsz", LinkPath(exe)); err != nil {
			t.Fatal(err)
		}
		state, target := LinkState(exe)
		if state != Foreign || target != "/usr/bin/lrzsz" {
			t.Fatalf("state=%q target=%q", state, target)
		}
		_, created, err := CreateLink(exe)
		if created || err == nil || !strings.Contains(err.Error(), "points at /usr/bin/lrzsz") {
			t.Fatalf("created=%v err=%v", created, err)
		}
	})
	t.Run("regular file", func(t *testing.T) {
		exe := binary(t)
		if err := os.WriteFile(LinkPath(exe), []byte("someone else"), 0o755); err != nil {
			t.Fatal(err)
		}
		state, target := LinkState(exe)
		if state != Foreign || target != "" {
			t.Fatalf("state=%q target=%q", state, target)
		}
		_, _, err := CreateLink(exe)
		if err == nil || strings.Contains(err.Error(), "points at") {
			t.Fatalf("err = %v", err)
		}
	})
}

// A binary that could not be located says so instead of naming a bare sz
// in whatever directory the shell happens to be in.
func TestLinkWithoutAKnownBinary(t *testing.T) {
	if state, _ := LinkState(""); state != Unknown {
		t.Errorf("state = %q, want unknown", state)
	}
	if path := LinkPath(""); path != "" {
		t.Errorf("path = %q, want empty", path)
	}
	if _, _, err := CreateLink(""); err == nil || !strings.Contains(err.Error(), "cannot locate") {
		t.Errorf("err = %v", err)
	}
}

func TestCreateLinkReportsAFailedSymlink(t *testing.T) {
	exe := filepath.Join(t.TempDir(), "missing", "screenz")
	if _, created, err := CreateLink(exe); created || err == nil {
		t.Fatalf("created=%v err=%v", created, err)
	}
}

func TestShellResolution(t *testing.T) {
	cases := []struct {
		name, sel, shellEnv string
		want                []string
		wantErr             string
	}{
		{name: "login shell", shellEnv: "/bin/zsh", want: []string{"zsh"}},
		{name: "every shell", sel: "all", want: Shells},
		{name: "named shell", sel: "fish", shellEnv: "/bin/zsh", want: []string{"fish"}},
		{name: "unset SHELL", wantErr: "cannot tell which shell"},
		{name: "unknown login shell", shellEnv: "/usr/bin/nu", wantErr: "no completion script for nu"},
		{name: "unknown named shell", sel: "csh", wantErr: "want bash, fish, zsh or all"},
		{name: "root SHELL", shellEnv: "/", wantErr: "cannot tell which shell"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ResolveShells(tc.sel, tc.shellEnv)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v, want %q", err, tc.wantErr)
				}
				return
			}
			if err != nil || strings.Join(got, ",") != strings.Join(tc.want, ",") {
				t.Fatalf("got %v err=%v", got, err)
			}
		})
	}
	if KnownShell("csh") {
		t.Error("csh must not be a known shell")
	}
}

// Scripts land in a completions directory beside profiles, and each shell
// gets the rc line that makes it load them.
func TestWriteScriptAndReport(t *testing.T) {
	dir := Dir(t.TempDir())
	if filepath.Base(dir) != "completions" {
		t.Fatalf("dir = %q", dir)
	}
	if got := InstalledShells(dir); len(got) != 0 {
		t.Fatalf("installed before writing: %v", got)
	}
	for _, shell := range Shells {
		path, err := WriteScript(dir, shell, "# "+shell+"\n")
		if err != nil || path != ScriptPath(dir, shell) {
			t.Fatalf("%s: path=%q err=%v", shell, path, err)
		}
		if !HasScript(dir, shell) {
			t.Errorf("%s: not reported as installed", shell)
		}
		if line := RCLine(dir, shell); !strings.Contains(line, dir) {
			t.Errorf("%s rc line does not name the directory: %q", shell, line)
		}
	}
	if got := InstalledShells(dir); strings.Join(got, ",") != "bash,fish,zsh" {
		t.Errorf("installed = %v", got)
	}
	// zsh is the one that needs fpath rather than a source line, because
	// its script is an autoloaded function file.
	if line := RCLine(dir, "zsh"); !strings.Contains(line, "fpath=(") {
		t.Errorf("zsh rc line = %q", line)
	}
	if line := RCLine(dir, "bash"); !strings.HasPrefix(line, "source ") {
		t.Errorf("bash rc line = %q", line)
	}
}

func TestWriteScriptReportsFilesystemFailures(t *testing.T) {
	t.Run("directory cannot be created", func(t *testing.T) {
		blocked := filepath.Join(t.TempDir(), "screenz")
		if err := os.WriteFile(blocked, []byte("a file, not a directory"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := WriteScript(Dir(blocked), "zsh", "x"); err == nil {
			t.Fatal("want an error writing under a regular file")
		}
	})
	t.Run("script cannot be written", func(t *testing.T) {
		dir := Dir(t.TempDir())
		if err := os.MkdirAll(ScriptPath(dir, "zsh"), 0o755); err != nil {
			t.Fatal(err)
		}
		if _, err := WriteScript(dir, "zsh", "x"); err == nil {
			t.Fatal("want an error writing over a directory")
		}
	})
}
