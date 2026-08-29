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
