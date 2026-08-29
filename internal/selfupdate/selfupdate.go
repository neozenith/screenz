// Package selfupdate implements `screenz update`: resolve the latest
// GitHub release, verify the downloaded artifact against its checksums,
// and atomically replace the running binary. Everything here is pure over
// bytes and paths; the network lives behind cli.Deps.Fetch.
package selfupdate

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LatestURL is the GitHub API endpoint for the newest non-prerelease.
const LatestURL = "https://api.github.com/repos/neozenith/screenz/releases/latest"

// Release is the subset of the GitHub release response an update needs.
type Release struct {
	Tag    string  `json:"tag_name"`
	Assets []Asset `json:"assets"`
}

// Asset is one downloadable release file.
type Asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

// ParseRelease reads the GitHub API JSON.
func ParseRelease(src []byte) (Release, error) {
	var r Release
	if err := json.Unmarshal(src, &r); err != nil {
		return Release{}, fmt.Errorf("release response: %w", err)
	}
	if r.Tag == "" {
		return Release{}, fmt.Errorf("release response has no tag_name")
	}
	return r, nil
}

// Binary picks the darwin tarball for the given architecture.
func (r Release) Binary(goarch string) (Asset, error) {
	suffix := "_darwin_" + goarch + ".tar.gz"
	for _, a := range r.Assets {
		if strings.HasSuffix(a.Name, suffix) {
			return a, nil
		}
	}
	return Asset{}, fmt.Errorf("release %s has no %s asset", r.Tag, suffix)
}

// Checksums picks the checksums.txt asset.
func (r Release) Checksums() (Asset, error) {
	for _, a := range r.Assets {
		if a.Name == "checksums.txt" {
			return a, nil
		}
	}
	return Asset{}, fmt.Errorf("release %s has no checksums.txt asset", r.Tag)
}

// Same reports whether two version strings name the same release,
// tolerating a leading v on either side.
func Same(current, latest string) bool {
	return strings.TrimPrefix(current, "v") == strings.TrimPrefix(latest, "v")
}

// VerifyChecksum checks data against the release's checksums.txt entry for
// name — the release page and the running update use the same trust anchor.
func VerifyChecksum(checksums []byte, name string, data []byte) error {
	sum := sha256.Sum256(data)
	got := hex.EncodeToString(sum[:])
	for _, line := range strings.Split(string(checksums), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == name {
			if fields[0] != got {
				return fmt.Errorf("checksum mismatch for %s: want %s, got %s", name, fields[0], got)
			}
			return nil
		}
	}
	return fmt.Errorf("checksums.txt has no entry for %s", name)
}

// ExtractBinary pulls the screenz executable out of a release tarball.
func ExtractBinary(tgz []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(tgz))
	if err != nil {
		return nil, fmt.Errorf("gunzip: %w", err)
	}
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err != nil {
			return nil, fmt.Errorf("no screenz binary in archive: %w", err)
		}
		if hdr.Typeflag == tar.TypeReg && filepath.Base(hdr.Name) == "screenz" {
			bin, err := io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("read binary from archive: %w", err)
			}
			return bin, nil
		}
	}
}

// Replace atomically swaps the executable: write a sibling file, then
// rename over the original so a failed update can never leave a truncated
// binary on PATH.
func Replace(exe string, bin []byte) error {
	tmp := exe + ".new"
	if err := os.WriteFile(tmp, bin, 0o755); err != nil {
		return err
	}
	return os.Rename(tmp, exe)
}
