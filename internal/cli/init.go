package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/neozenith/screenz/internal/profile"
)

const initHelp = `usage: screenz init --profile NAME [--force]

Write a commented template profile to the profile directory, ready to
hand-edit: display aliases, rules, and a note on every field.

Flags:
  -p, --profile NAME   The profile to create. Required.
      --force          Overwrite an existing profile. No short form: it
                       destroys a file you may have hand-edited.
  -h, --help           Show this help.

Inspect what exists with: screenz list
Apply the result with:    screenz apply --profile NAME
`

func runInit(args []string, stdout, stderr io.Writer, d Deps) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	name := fs.String("profile", "", "the profile to create")
	fs.StringVar(name, "p", "", "the profile to create")
	force := fs.Bool("force", false, "overwrite an existing profile")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(stdout, initHelp)
			return 0
		}
		fmt.Fprintf(stderr, "screenz init: %v\n", err)
		return 2
	}
	// The name is a flag, never a positional (ADR-0025), so `init office`
	// is caught here rather than silently writing nothing.
	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "screenz init: unexpected argument %q (name the profile with --profile %s)\n",
			fs.Arg(0), fs.Arg(0))
		return 2
	}
	if *name == "" {
		fmt.Fprintln(stderr, "screenz init: --profile NAME is required")
		fmt.Fprint(stderr, initHelp)
		return 2
	}
	path := profile.Path(d.Getenv, d.Home, *name)
	if err := profile.WriteTemplate(path, *name, *force); err != nil {
		fmt.Fprintf(stderr, "screenz init: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "wrote %s\n", path)
	return 0
}
