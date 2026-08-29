package selfupdate

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// tgz builds a real release-shaped tarball holding one file.
func tgz(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(content)), Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	tw.Close()
	gz.Close()
	return buf.Bytes()
}

const releaseJSON = `{
  "tag_name": "v0.2.0",
  "assets": [
    {"name": "checksums.txt", "browser_download_url": "https://example/checksums.txt"},
    {"name": "screenz_v0.2.0_darwin_arm64.tar.gz", "browser_download_url": "https://example/arm64.tgz"},
    {"name": "screenz_v0.2.0_darwin_amd64.tar.gz", "browser_download_url": "https://example/amd64.tgz"}
  ]
}`

func TestParseReleaseAndAssetSelection(t *testing.T) {
	rel, err := ParseRelease([]byte(releaseJSON))
	if err != nil {
		t.Fatal(err)
	}
	if rel.Tag != "v0.2.0" {
		t.Fatalf("tag = %q", rel.Tag)
	}
	bin, err := rel.Binary("arm64")
	if err != nil || bin.URL != "https://example/arm64.tgz" {
		t.Fatalf("Binary(arm64) = %+v, %v", bin, err)
	}
	sums, err := rel.Checksums()
	if err != nil || sums.Name != "checksums.txt" {
		t.Fatalf("Checksums() = %+v, %v", sums, err)
	}
	if _, err := rel.Binary("riscv"); err == nil {
		t.Error("missing arch accepted")
	}
	if _, err := ParseRelease([]byte(`{"assets": []}`)); err == nil {
		t.Error("missing tag accepted")
	}
	if _, err := ParseRelease([]byte(`not json`)); err == nil {
		t.Error("bad json accepted")
	}
	empty := Release{Tag: "v1"}
	if _, err := empty.Checksums(); err == nil {
		t.Error("missing checksums accepted")
	}
}

func TestSame(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want bool
	}{
		{"v0.1.0", "v0.1.0", true},
		{"0.1.0", "v0.1.0", true},
		{"v0.1.0", "v0.2.0", false},
		{"dev", "v0.2.0", false},
	} {
		if got := Same(tc.a, tc.b); got != tc.want {
			t.Errorf("Same(%q,%q) = %v", tc.a, tc.b, got)
		}
	}
}

func TestVerifyChecksum(t *testing.T) {
	data := []byte("release bytes")
	sum := sha256.Sum256(data)
	name := "screenz_v0.2.0_darwin_arm64.tar.gz"
	sums := []byte(hex.EncodeToString(sum[:]) + "  " + name + "\nother  screenz_v0.2.0_darwin_amd64.tar.gz\n")
	if err := VerifyChecksum(sums, name, data); err != nil {
		t.Fatal(err)
	}
	if err := VerifyChecksum(sums, name, []byte("tampered")); err == nil || !strings.Contains(err.Error(), "mismatch") {
		t.Errorf("tampered data accepted: %v", err)
	}
	if err := VerifyChecksum(sums, "missing.tar.gz", data); err == nil || !strings.Contains(err.Error(), "no entry") {
		t.Errorf("missing entry accepted: %v", err)
	}
}

func TestExtractBinary(t *testing.T) {
	want := []byte("#!/fake screenz binary")
	bin, err := ExtractBinary(tgz(t, "screenz", want))
	if err != nil || !bytes.Equal(bin, want) {
		t.Fatalf("ExtractBinary = %q, %v", bin, err)
	}
	if _, err := ExtractBinary([]byte("not gzip")); err == nil {
		t.Error("bad gzip accepted")
	}
	if _, err := ExtractBinary(tgz(t, "otherfile", want)); err == nil {
		t.Error("archive without screenz accepted")
	}
	// A truncated stream fails the in-archive read.
	full := tgz(t, "screenz", bytes.Repeat([]byte("x"), 1<<16))
	if _, err := ExtractBinary(full[:len(full)/2]); err == nil {
		t.Error("truncated archive accepted")
	}
}

func TestReplace(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "screenz")
	if err := os.WriteFile(exe, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := Replace(exe, []byte("new")); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(exe)
	if string(got) != "new" {
		t.Fatalf("binary = %q", got)
	}
	info, _ := os.Stat(exe)
	if info.Mode().Perm() != 0o755 {
		t.Errorf("mode = %v", info.Mode())
	}
	if err := Replace(filepath.Join(dir, "missing", "screenz"), []byte("x")); err == nil {
		t.Error("write into missing dir accepted")
	}
	// Rename over a directory fails after a successful temp write.
	asDir := filepath.Join(dir, "target")
	os.MkdirAll(asDir, 0o755)
	if err := Replace(asDir, []byte("x")); err == nil {
		t.Error("rename over a directory accepted")
	}
}
