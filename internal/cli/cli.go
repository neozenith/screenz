// Package cli parses arguments, formats output, and decides exit codes.
// Every OS-touching operation arrives through Deps so this package stays a
// pure pipeline testable on real captured values (ADR1.3); cmd/screenz wires
// the real implementations.
package cli

import (
	"fmt"
	"io"
)

// SysInfo is the machine report a doctor run needs, gathered impurely by
// cmd/screenz from internal/mac.
type SysInfo struct {
	Trusted        bool     `json:"trusted"`
	HostAppName    string   `json:"host_app_name"`
	HostAppBundle  string   `json:"host_app_bundle"`
	DisplayNames   []string `json:"display_names"`
	MissingSymbols []string `json:"missing_symbols,omitempty"`
}

// Deps carries every impure dependency a command may need.
type Deps struct {
	Getenv func(string) string
	Home   string
	Sys    func(prompt bool) SysInfo
}

// Run dispatches the first positional to its command handler and returns the
// process exit code: 0 success, 1 runtime failure, 2 usage error.
func Run(args []string, stdout, stderr io.Writer, d Deps) int {
	cmd := ""
	if len(args) > 0 {
		cmd = args[0]
	}
	switch cmd {
	case "doctor":
		return runDoctor(args[1:], stdout, stderr, d)
	case "-h", "--help", "help":
		usage(stdout)
		return 0
	default:
		if cmd == "" {
			fmt.Fprintln(stderr, "screenz: no command given")
		} else {
			fmt.Fprintf(stderr, "screenz: unknown command %q\n", cmd)
		}
		usage(stderr)
		return 2
	}
}

func usage(w io.Writer) {
	fmt.Fprint(w, `usage: screenz <command> [flags]

Commands:
  doctor    Check the Accessibility grant, displays, and symbol bindings.

Run 'screenz <command> --help' for command flags.
`)
}
