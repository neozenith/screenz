package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/neozenith/screenz/internal/install"
	"github.com/neozenith/screenz/internal/profile"
	"github.com/neozenith/screenz/internal/selfupdate"
)

const updateHelp = `usage: screenz update [--check] [--force] [--all] [--link] [--completions] [--shell SHELL]

Keep the whole installation current, not just the binary (ADR-0029):

  the release      Check the GitHub Releases page for a newer screenz and
                   replace this binary in place (checksum-verified, atomic
                   swap; no quarantine xattr is set).
  the sz link      A symlink named sz beside the binary, so the same tool
                   answers to two letters wherever screenz is on PATH.
  completions      A completion script for zsh, bash or fish, written to
                   the screenz config directory with the one line to add
                   to your shell's startup file.

Bare 'screenz update' does the release only, and says on stderr when the
link or the completions are missing. Naming a part does that part alone
and needs no network; --all does everything.

Flags:
      --all           The release, the sz link and this shell's completions.
      --link          Create the sz symlink beside the binary.
      --completions   Write this shell's completion script.
      --shell SHELL   Which shell to complete for: zsh, bash, fish or all.
                      Implies --completions; the default is $SHELL.
      --check         Report what each selected part would do, and change
                      nothing.
      --force         Update even from a dev (source) build or the same version.
  -h, --help          Show this help.

No flag here has a one-letter alias: this command replaces the running
binary and writes to your shell configuration, so each is typed in full
(ADR-0021).

An sz that is already something else is never replaced; the run says so
and exits 1. Completion scripts are written only inside the screenz config
directory — nothing edits your shell's startup file for you.
`

// updateFlags are the parsed values of update's flags; registerUpdate is
// what both the parser and the completion generator read (ADR-0030).
type updateFlags struct {
	check       *bool
	force       *bool
	all         *bool
	link        *bool
	completions *bool
	shell       *string
}

func registerUpdate(fs *flag.FlagSet) updateFlags {
	return updateFlags{
		check:       fs.Bool("check", false, "report only"),
		force:       fs.Bool("force", false, "update even from a dev build or the same version"),
		all:         fs.Bool("all", false, "the release, the sz link and this shell's completions"),
		link:        fs.Bool("link", false, "create the sz symlink beside the binary"),
		completions: fs.Bool("completions", false, "write this shell's completion script"),
		shell:       fs.String("shell", "", "which shell to complete for: zsh, bash, fish or all"),
	}
}

func runUpdate(args []string, stdout, stderr io.Writer, d Deps) int {
	fs := newFlagSet("update")
	f := registerUpdate(fs)
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(stdout, updateHelp)
			return 0
		}
		fmt.Fprintf(stderr, "screenz update: %v\n", err)
		fmt.Fprint(stderr, updateHelp)
		return 2
	}

	// Naming a part narrows the run to the parts named: --link must not
	// wait on the network to do something local. --shell implies
	// --completions, as --jq implies --json (ADR-0027).
	doLink := *f.all || *f.link
	doComp := *f.all || *f.completions || *f.shell != ""
	doRelease := *f.all || (!doLink && !doComp)

	code := 0
	if doRelease {
		if c := updateRelease(*f.check, *f.force, stdout, stderr, d); c != 0 {
			return c
		}
	}
	if doLink {
		if c := updateLink(*f.check, stdout, stderr, d); c != 0 {
			code = c
		}
	}
	if doComp {
		if c := updateCompletions(*f.check, *f.shell, stdout, stderr, d); c != 0 {
			code = c
		}
	}
	// A release-only run reports what the rest of the install is missing,
	// on stderr so it never lands in the middle of the update line.
	if !doLink && !doComp {
		reportMissing(stderr, d)
	}
	return code
}

// updateRelease is the original self-update: resolve the latest release,
// verify the artifact against its checksums and swap the binary.
func updateRelease(check, force bool, stdout, stderr io.Writer, d Deps) int {
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

	if selfupdate.Same(version, rel.Tag) && !force {
		fmt.Fprintf(stdout, "already up to date (%s)\n", rel.Tag)
		return 0
	}
	if check {
		fmt.Fprintf(stdout, "update available: %s -> %s (run 'screenz update')\n", version, rel.Tag)
		return 0
	}
	// A dev build did not come from a release; overwriting it loses work
	// unless that is exactly what the user wants.
	if version == "dev" && !force {
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

// updateLink creates the sz short link, or reports its state under
// --check. A name something else holds is never taken (ADR-0029).
func updateLink(check bool, stdout, stderr io.Writer, d Deps) int {
	if check {
		state, _ := install.LinkState(d.ExePath)
		fmt.Fprintf(stdout, "sz short link: %s (%s)\n", state, install.LinkPath(d.ExePath))
		return 0
	}
	path, created, err := install.CreateLink(d.ExePath)
	if err != nil {
		fmt.Fprintf(stderr, "screenz update: %v\n", err)
		return 1
	}
	if !created {
		fmt.Fprintf(stdout, "sz short link already installed (%s)\n", path)
		return 0
	}
	fmt.Fprintf(stdout, "linked %s -> %s\n", path, filepath.Base(d.ExePath))
	return 0
}

// updateCompletions writes the completion script for each selected shell,
// or reports what is installed under --check. The script is generated from
// the command table (ADR-0030), so it is current by construction.
func updateCompletions(check bool, shellSel string, stdout, stderr io.Writer, d Deps) int {
	shells, err := install.ResolveShells(shellSel, d.Getenv("SHELL"))
	if err != nil {
		fmt.Fprintf(stderr, "screenz update: %v\n", err)
		// A --shell value screenz does not know is a usage error; a
		// $SHELL it cannot read is the machine's state, not the command's.
		if shellSel != "" {
			return 2
		}
		return 1
	}
	dir := install.Dir(profile.Dir(d.Getenv, d.Home))
	if check {
		for _, shell := range shells {
			state := "not installed"
			if install.HasScript(dir, shell) {
				state = "installed"
			}
			fmt.Fprintf(stdout, "%s completions: %s (%s)\n", shell, state, install.ScriptPath(dir, shell))
		}
		return 0
	}
	for _, shell := range shells {
		path, err := install.WriteScript(dir, shell, CompletionScript(shell))
		if err != nil {
			fmt.Fprintf(stderr, "screenz update: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "wrote %s\n", path)
		fmt.Fprintf(stdout, "  add to your shell startup file: %s\n", install.RCLine(dir, shell))
	}
	return 0
}

// reportMissing names the parts of the install that are not there yet.
// Only what can be fixed by rerunning is named: an sz that belongs to
// something else is not a missing link, and a shell screenz has no script
// for is not a missing script.
func reportMissing(stderr io.Writer, d Deps) {
	var missing []string
	if state, _ := install.LinkState(d.ExePath); state == install.Absent {
		missing = append(missing, "the sz short link")
	}
	if shell, err := install.DetectShell(d.Getenv("SHELL")); err == nil {
		if !install.HasScript(install.Dir(profile.Dir(d.Getenv, d.Home)), shell) {
			missing = append(missing, shell+" completions")
		}
	}
	if len(missing) == 0 {
		return
	}
	fmt.Fprintf(stderr, "screenz update: %s not installed; run 'screenz update --all' to set up\n",
		strings.Join(missing, " and "))
}
