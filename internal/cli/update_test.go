package cli

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/neozenith/screenz/internal/selfupdate"
)

// releaseFixture is a complete fake release: real tarball, real sha256.
func releaseFixture(t *testing.T) (urls map[string][]byte, binary []byte) {
	t.Helper()
	binary = []byte("#!/fake updated screenz")
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	tw.WriteHeader(&tar.Header{Name: "screenz", Mode: 0o755, Size: int64(len(binary)), Typeflag: tar.TypeReg})
	tw.Write(binary)
	tw.Close()
	gz.Close()
	tarball := buf.Bytes()
	assetName := fmt.Sprintf("screenz_v9.9.9_darwin_%s.tar.gz", runtime.GOARCH)
	sum := sha256.Sum256(tarball)
	urls = map[string][]byte{
		selfupdate.LatestURL: []byte(fmt.Sprintf(`{"tag_name":"v9.9.9","assets":[
			{"name":"checksums.txt","browser_download_url":"https://dl/checksums.txt"},
			{"name":%q,"browser_download_url":"https://dl/bin.tgz"}]}`, assetName)),
		"https://dl/bin.tgz":       tarball,
		"https://dl/checksums.txt": []byte(hex.EncodeToString(sum[:]) + "  " + assetName + "\n"),
	}
	return urls, binary
}

func updateDeps(t *testing.T, urls map[string][]byte) Deps {
	t.Helper()
	d := deps(officeSys(true))
	d.Fetch = func(url string) ([]byte, error) {
		body, ok := urls[url]
		if !ok {
			return nil, fmt.Errorf("GET %s: 404 Not Found", url)
		}
		return body, nil
	}
	exe := filepath.Join(t.TempDir(), "screenz")
	if err := os.WriteFile(exe, []byte("old binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	d.ExePath = exe
	return d
}

// The dev-build guard: a source build refuses to overwrite itself unless
// forced; --force performs the full verified swap.
func TestUpdateDevGuardAndForce(t *testing.T) {
	urls, binary := releaseFixture(t)
	d := updateDeps(t, urls)

	code, _, errOut := run(t, []string{"update"}, d)
	if code != 1 || !strings.Contains(errOut, "dev build") {
		t.Fatalf("dev guard: exit=%d stderr=%q", code, errOut)
	}

	code, out, errOut := run(t, []string{"update", "--force"}, d)
	if code != 0 {
		t.Fatalf("force: exit=%d stderr=%q", code, errOut)
	}
	if !strings.Contains(out, "updated dev -> v9.9.9") {
		t.Errorf("stdout missing update line:\n%s", out)
	}
	got, _ := os.ReadFile(d.ExePath)
	if !bytes.Equal(got, binary) {
		t.Fatalf("binary not replaced: %q", got)
	}
}

func TestUpdateCheckAndUpToDate(t *testing.T) {
	urls, _ := releaseFixture(t)
	d := updateDeps(t, urls)
	code, out, _ := run(t, []string{"update", "--check"}, d)
	if code != 0 || !strings.Contains(out, "update available: dev -> v9.9.9") {
		t.Fatalf("check: exit=%d out=%q", code, out)
	}
	// Same version reports up to date and touches nothing.
	urls[selfupdate.LatestURL] = []byte(`{"tag_name":"dev","assets":[]}`)
	code, out, _ = run(t, []string{"update"}, d)
	if code != 0 || !strings.Contains(out, "already up to date (dev)") {
		t.Fatalf("up to date: exit=%d out=%q", code, out)
	}
}

func TestUpdateFailurePaths(t *testing.T) {
	urls, _ := releaseFixture(t)
	corrupt := func(mutate func(map[string][]byte)) Deps {
		u := map[string][]byte{}
		for k, v := range urls {
			u[k] = v
		}
		mutate(u)
		return updateDeps(t, u)
	}
	cases := []struct {
		name string
		deps Deps
		args []string
		want string
	}{
		{"api unreachable", corrupt(func(u map[string][]byte) { delete(u, selfupdate.LatestURL) }), []string{"update"}, "404"},
		{"bad release json", corrupt(func(u map[string][]byte) { u[selfupdate.LatestURL] = []byte("garbage") }), []string{"update"}, "release response"},
		{"missing arch asset", corrupt(func(u map[string][]byte) {
			u[selfupdate.LatestURL] = []byte(`{"tag_name":"v9.9.9","assets":[{"name":"checksums.txt","browser_download_url":"https://dl/checksums.txt"}]}`)
		}), []string{"update", "--force"}, "has no _darwin_"},
		{"missing checksums asset", corrupt(func(u map[string][]byte) {
			u[selfupdate.LatestURL] = bytes.Replace(u[selfupdate.LatestURL], []byte("checksums.txt"), []byte("nochecks.txt"), -1)
		}), []string{"update", "--force"}, "no checksums.txt"},
		{"tarball unreachable", corrupt(func(u map[string][]byte) { delete(u, "https://dl/bin.tgz") }), []string{"update", "--force"}, "404"},
		{"checksums unreachable", corrupt(func(u map[string][]byte) { delete(u, "https://dl/checksums.txt") }), []string{"update", "--force"}, "404"},
		{"tampered tarball", corrupt(func(u map[string][]byte) { u["https://dl/bin.tgz"] = []byte("evil bytes") }), []string{"update", "--force"}, "checksum mismatch"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, _, errOut := run(t, tc.args, tc.deps)
			if code != 1 || !strings.Contains(errOut, tc.want) {
				t.Fatalf("exit=%d stderr=%q (want %q)", code, errOut, tc.want)
			}
		})
	}

	// A tarball that passes checksum but holds no binary, and a bad exe path.
	t.Run("no binary in archive", func(t *testing.T) {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		tw := tar.NewWriter(gz)
		tw.WriteHeader(&tar.Header{Name: "README", Mode: 0o644, Size: 2, Typeflag: tar.TypeReg})
		tw.Write([]byte("hi"))
		tw.Close()
		gz.Close()
		sum := sha256.Sum256(buf.Bytes())
		name := fmt.Sprintf("screenz_v9.9.9_darwin_%s.tar.gz", runtime.GOARCH)
		d := corrupt(func(u map[string][]byte) {
			u["https://dl/bin.tgz"] = buf.Bytes()
			u["https://dl/checksums.txt"] = []byte(hex.EncodeToString(sum[:]) + "  " + name + "\n")
		})
		code, _, errOut := run(t, []string{"update", "--force"}, d)
		if code != 1 || !strings.Contains(errOut, "no screenz binary") {
			t.Fatalf("exit=%d stderr=%q", code, errOut)
		}
	})
	t.Run("replace fails", func(t *testing.T) {
		d := updateDeps(t, urls)
		d.ExePath = filepath.Join(t.TempDir(), "missing", "screenz")
		code, _, errOut := run(t, []string{"update", "--force"}, d)
		if code != 1 || errOut == "" {
			t.Fatalf("exit=%d stderr=%q", code, errOut)
		}
	})
}

func TestUpdateHelpAndBadFlag(t *testing.T) {
	urls, _ := releaseFixture(t)
	d := updateDeps(t, urls)
	if code, out, _ := run(t, []string{"update", "--help"}, d); code != 0 || !strings.Contains(out, "usage: screenz update") {
		t.Fatalf("help: exit=%d", code)
	}
	if code, out, errOut := run(t, []string{"update", "--nope"}, d); code != 2 || out != "" || !strings.Contains(errOut, "usage: screenz update") {
		t.Fatalf("bad flag: exit=%d out=%q err=%q", code, out, errOut)
	}
}

// upToDate makes the release check a no-op, so a test can exercise the
// install parts of update without a version bump getting in the way.
func upToDate(urls map[string][]byte) map[string][]byte {
	u := map[string][]byte{}
	for k, v := range urls {
		u[k] = v
	}
	u[selfupdate.LatestURL] = []byte(`{"tag_name":"dev","assets":[]}`)
	return u
}

// installDeps is updateDeps with an environment: $SHELL decides which
// completion script is written, $SCREENZ_HOME where it goes (ADR-0015).
func installDeps(t *testing.T, urls map[string][]byte, env map[string]string) Deps {
	t.Helper()
	d := updateDeps(t, urls)
	if _, ok := env["SCREENZ_HOME"]; !ok {
		env["SCREENZ_HOME"] = t.TempDir()
	}
	d.Getenv = func(key string) string { return env[key] }
	return d
}

// The sz short link is created beside the binary, is idempotent, and is
// never taken from something else that holds the name (ADR-0029).
func TestUpdateLink(t *testing.T) {
	urls, _ := releaseFixture(t)
	d := installDeps(t, urls, map[string]string{})

	code, out, errOut := run(t, []string{"update", "--link"}, d)
	if code != 0 || !strings.Contains(out, "linked ") || !strings.HasSuffix(strings.TrimSpace(out), "-> screenz") {
		t.Fatalf("link: exit=%d out=%q err=%q", code, out, errOut)
	}
	link := filepath.Join(filepath.Dir(d.ExePath), "sz")
	if target, err := os.Readlink(link); err != nil || target != "screenz" {
		t.Fatalf("readlink: %q %v", target, err)
	}
	if code, out, _ := run(t, []string{"update", "--link"}, d); code != 0 || !strings.Contains(out, "already installed") {
		t.Errorf("second link: exit=%d out=%q", code, out)
	}
	if code, out, _ := run(t, []string{"update", "--check", "--link"}, d); code != 0 || !strings.Contains(out, "sz short link: linked") {
		t.Errorf("check: exit=%d out=%q", code, out)
	}
}

func TestUpdateLinkRefusesAForeignSZ(t *testing.T) {
	urls, _ := releaseFixture(t)
	d := installDeps(t, urls, map[string]string{})
	if err := os.WriteFile(filepath.Join(filepath.Dir(d.ExePath), "sz"), []byte("lrzsz"), 0o755); err != nil {
		t.Fatal(err)
	}
	code, _, errOut := run(t, []string{"update", "--link"}, d)
	if code != 1 || !strings.Contains(errOut, "not replacing it") {
		t.Fatalf("exit=%d err=%q", code, errOut)
	}
	if code, out, _ := run(t, []string{"update", "--check", "--link"}, d); code != 0 || !strings.Contains(out, "sz short link: foreign") {
		t.Errorf("check: exit=%d out=%q", code, out)
	}
}

// Completions are written for the login shell by default, for the shell
// --shell names otherwise, and the run prints the one line that makes the
// shell load them — nothing edits a startup file (ADR-0029).
func TestUpdateCompletions(t *testing.T) {
	urls, _ := releaseFixture(t)
	home := t.TempDir()
	d := installDeps(t, urls, map[string]string{"SHELL": "/bin/zsh", "SCREENZ_HOME": home})

	code, out, errOut := run(t, []string{"update", "--completions"}, d)
	if code != 0 {
		t.Fatalf("exit=%d err=%q", code, errOut)
	}
	script := filepath.Join(home, "completions", "_screenz")
	if !strings.Contains(out, "wrote "+script) || !strings.Contains(out, "fpath=(") {
		t.Fatalf("out = %q", out)
	}
	body, err := os.ReadFile(script)
	if err != nil || !strings.HasPrefix(string(body), "#compdef screenz sz") {
		t.Fatalf("script = %q err=%v", body, err)
	}

	// --shell implies --completions, and `all` writes every dialect.
	code, out, _ = run(t, []string{"update", "--shell", "all"}, d)
	if code != 0 {
		t.Fatalf("--shell all: exit=%d", code)
	}
	for _, name := range []string{"_screenz", "screenz.bash", "screenz.fish"} {
		if !strings.Contains(out, name) {
			t.Errorf("--shell all did not write %s\n%s", name, out)
		}
	}
	code, out, _ = run(t, []string{"update", "--check", "--completions"}, d)
	if code != 0 || !strings.Contains(out, "zsh completions: installed") {
		t.Errorf("check: exit=%d out=%q", code, out)
	}
}

func TestUpdateCompletionsRefusesAShellItCannotWrite(t *testing.T) {
	urls, _ := releaseFixture(t)
	// A --shell value screenz has no script for is a usage error…
	d := installDeps(t, urls, map[string]string{"SHELL": "/bin/zsh"})
	if code, _, errOut := run(t, []string{"update", "--shell", "csh"}, d); code != 2 || !strings.Contains(errOut, "want bash, fish, zsh or all") {
		t.Errorf("bad --shell: exit=%d err=%q", code, errOut)
	}
	// …an unreadable $SHELL is the machine's state, not the command's.
	d = installDeps(t, urls, map[string]string{})
	if code, _, errOut := run(t, []string{"update", "--completions"}, d); code != 1 || !strings.Contains(errOut, "cannot tell which shell") {
		t.Errorf("no $SHELL: exit=%d err=%q", code, errOut)
	}
	// A directory that cannot be written reports the failure.
	blocked := filepath.Join(t.TempDir(), "home")
	if err := os.WriteFile(blocked, []byte("a file, not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	d = installDeps(t, urls, map[string]string{"SHELL": "/bin/zsh", "SCREENZ_HOME": blocked})
	if code, _, errOut := run(t, []string{"update", "--completions"}, d); code != 1 || errOut == "" {
		t.Errorf("unwritable home: exit=%d err=%q", code, errOut)
	}
}

// --all does the release, the link and this shell's completions in one go;
// a bare update does the release and says what the rest of the install is
// still missing.
func TestUpdateAllAndTheMissingReport(t *testing.T) {
	urls, _ := releaseFixture(t)
	home := t.TempDir()
	d := installDeps(t, upToDate(urls), map[string]string{"SHELL": "/bin/bash", "SCREENZ_HOME": home})

	code, out, errOut := run(t, []string{"update"}, d)
	if code != 0 || !strings.Contains(out, "already up to date") {
		t.Fatalf("bare: exit=%d out=%q", code, out)
	}
	if !strings.Contains(errOut, "the sz short link and bash completions not installed") ||
		!strings.Contains(errOut, "screenz update --all") {
		t.Fatalf("bare update must name what is missing: %q", errOut)
	}

	code, out, errOut = run(t, []string{"update", "--all"}, d)
	if code != 0 {
		t.Fatalf("--all: exit=%d err=%q", code, errOut)
	}
	for _, want := range []string{"already up to date", "linked ", "screenz.bash"} {
		if !strings.Contains(out, want) {
			t.Errorf("--all did not report %q\n%s", want, out)
		}
	}
	// With the install complete there is nothing left to nudge about.
	if _, _, errOut = run(t, []string{"update"}, d); errOut != "" {
		t.Errorf("stderr = %q, want silence", errOut)
	}
}

// A part named on its own is done on its own: no network, and the release
// is left alone.
func TestUpdateNarrowsToTheNamedPart(t *testing.T) {
	d := installDeps(t, map[string][]byte{}, map[string]string{"SHELL": "/bin/zsh"})
	d.Fetch = func(url string) ([]byte, error) { return nil, fmt.Errorf("network used for %s", url) }
	if code, out, errOut := run(t, []string{"update", "--link"}, d); code != 0 || !strings.Contains(out, "linked ") || errOut != "" {
		t.Fatalf("--link: exit=%d out=%q err=%q", code, out, errOut)
	}
	if code, _, errOut := run(t, []string{"update", "--completions"}, d); code != 0 || errOut != "" {
		t.Fatalf("--completions: exit=%d err=%q", code, errOut)
	}
}
