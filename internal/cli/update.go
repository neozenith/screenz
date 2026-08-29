package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"runtime"

	"github.com/neozenith/screenz/internal/selfupdate"
)

const updateHelp = `usage: screenz update [--check] [--force]

Check the GitHub Releases page for a newer screenz and replace this binary
in place (checksum-verified, atomic swap; no quarantine xattr is set).

Flags:
  --check   Report whether an update exists without installing it.
  --force   Update even from a dev (source) build or the same version.
`

func runUpdate(args []string, stdout, stderr io.Writer, d Deps) int {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	check := fs.Bool("check", false, "report only")
	force := fs.Bool("force", false, "update even from a dev build or the same version")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(stdout, updateHelp)
			return 0
		}
		fmt.Fprintf(stderr, "screenz update: %v\n", err)
		fmt.Fprint(stderr, updateHelp)
		return 2
	}

	body, err := d.Fetch(selfupdate.LatestURL)
	if err != nil {
		fmt.Fprintf(stderr, "screenz update: %v\n", err)
		return 1
	}
	rel, err := selfupdate.ParseRelease(body)
	if err != nil {
		fmt.Fprintf(stderr, "screenz update: %v\n", err)
		return 1
	}

	if selfupdate.Same(version, rel.Tag) && !*force {
		fmt.Fprintf(stdout, "already up to date (%s)\n", rel.Tag)
		return 0
	}
	if *check {
		fmt.Fprintf(stdout, "update available: %s -> %s (run 'screenz update')\n", version, rel.Tag)
		return 0
	}
	// A dev build did not come from a release; overwriting it loses work
	// unless that is exactly what the user wants.
	if version == "dev" && !*force {
		fmt.Fprintf(stderr, "screenz update: this is a dev build, not a release install; use --force to replace it with %s\n", rel.Tag)
		return 1
	}

	binAsset, err := rel.Binary(runtime.GOARCH)
	if err != nil {
		fmt.Fprintf(stderr, "screenz update: %v\n", err)
		return 1
	}
	sumAsset, err := rel.Checksums()
	if err != nil {
		fmt.Fprintf(stderr, "screenz update: %v\n", err)
		return 1
	}
	tgz, err := d.Fetch(binAsset.URL)
	if err != nil {
		fmt.Fprintf(stderr, "screenz update: %v\n", err)
		return 1
	}
	sums, err := d.Fetch(sumAsset.URL)
	if err != nil {
		fmt.Fprintf(stderr, "screenz update: %v\n", err)
		return 1
	}
	if err := selfupdate.VerifyChecksum(sums, binAsset.Name, tgz); err != nil {
		fmt.Fprintf(stderr, "screenz update: %v\n", err)
		return 1
	}
	bin, err := selfupdate.ExtractBinary(tgz)
	if err != nil {
		fmt.Fprintf(stderr, "screenz update: %v\n", err)
		return 1
	}
	if err := selfupdate.Replace(d.ExePath, bin); err != nil {
		fmt.Fprintf(stderr, "screenz update: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "updated %s -> %s (%s)\n", version, rel.Tag, d.ExePath)
	return 0
}
