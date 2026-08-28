package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/joshpeak/screenz/internal/profile"
)

type doctorReport struct {
	Schema        int      `json:"schema"`
	Version       string   `json:"version"`
	Accessibility string   `json:"accessibility"`
	HostAppName   string   `json:"host_app_name"`
	HostAppBundle string   `json:"host_app_bundle"`
	OSVersion     string   `json:"os_version"`
	Executable    string   `json:"executable"`
	Quarantined   bool     `json:"quarantined"`
	Warnings      []string `json:"warnings,omitempty"`
	Displays      []string `json:"displays"`
	ProfileDir    string   `json:"profile_dir"`
	Missing       []string `json:"missing_symbols,omitempty"`
}

// osWarnings derives the platform caveats G6 documents: the macOS 13 floor,
// and the ≥ 26.1 behavior where a path-based TCC grant is enforced but
// hidden from System Settings, and a quarantined binary (browser download)
// that Gatekeeper will block.
func osWarnings(info SysInfo) []string {
	var out []string
	if major, minor, ok := osVersionParts(info.OSVersion); ok {
		if major < 13 {
			out = append(out, fmt.Sprintf("macOS %s is below the supported floor (13+)", info.OSVersion))
		}
		if major > 26 || (major == 26 && minor >= 1) {
			out = append(out, "macOS >= 26.1 can enforce the Accessibility grant while hiding it from System Settings; if doctor says untrusted after granting, run: tccutil reset Accessibility "+orWord(info.HostAppBundle, "<terminal-bundle-id>")+" and grant again")
		}
	}
	if info.Quarantined {
		out = append(out, "binary is quarantined (browser download); clear it with: xattr -d com.apple.quarantine "+info.ExePath)
	}
	return out
}

const doctorHelp = `usage: screenz doctor [--json]

Check that screenz can do its job on this machine: the Accessibility grant
(held by the terminal app that launched screenz, not the binary), the
connected displays, the resolved profile directory, and that every macOS
symbol bound. Exits 1 when Accessibility is not granted (ADR1.2).

Flags:
  --json    Emit the report as JSON.
`

func runDoctor(args []string, stdout, stderr io.Writer, d Deps) int {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	jsonOut := fs.Bool("json", false, "emit JSON")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(stdout, doctorHelp)
			return 0
		}
		fmt.Fprintf(stderr, "screenz doctor: %v\n", err)
		fmt.Fprint(stderr, doctorHelp)
		return 2
	}

	// Doctor asks with the prompt option so a first run deep-links the
	// System Settings Accessibility pane (G1).
	info := d.Sys(true)
	rep := doctorReport{
		Schema:        1,
		Version:       version,
		Accessibility: "untrusted",
		HostAppName:   info.HostAppName,
		HostAppBundle: info.HostAppBundle,
		OSVersion:     info.OSVersion,
		Executable:    info.ExePath,
		Quarantined:   info.Quarantined,
		Warnings:      osWarnings(info),
		Displays:      info.DisplayNames,
		ProfileDir:    profile.Dir(d.Getenv, d.Home),
		Missing:       info.MissingSymbols,
	}
	if info.Trusted {
		rep.Accessibility = "trusted"
	}

	if *jsonOut {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(rep)
	} else {
		fmt.Fprintf(stdout, "screenz: %s\n", rep.Version)
		fmt.Fprintf(stdout, "macos: %s\n", rep.OSVersion)
		fmt.Fprintf(stdout, "accessibility: %s\n", rep.Accessibility)
		host := "unknown"
		if rep.HostAppName != "" {
			host = fmt.Sprintf("%s (%s)", rep.HostAppName, rep.HostAppBundle)
		}
		fmt.Fprintf(stdout, "host app: %s\n", host)
		fmt.Fprintf(stdout, "binary: %s\n", orWord(rep.Executable, "unknown"))
		fmt.Fprintf(stdout, "displays: %d\n", len(rep.Displays))
		for _, name := range rep.Displays {
			fmt.Fprintf(stdout, "  - %s\n", name)
		}
		fmt.Fprintf(stdout, "profile dir: %s\n", rep.ProfileDir)
		if len(rep.Missing) == 0 {
			fmt.Fprintln(stdout, "symbols: ok")
		} else {
			fmt.Fprintf(stdout, "symbols: %d missing\n", len(rep.Missing))
			for _, s := range rep.Missing {
				fmt.Fprintf(stdout, "  - %s\n", s)
			}
		}
		for _, w := range rep.Warnings {
			fmt.Fprintf(stdout, "warning: %s\n", w)
		}
	}

	if !info.Trusted {
		fmt.Fprint(stderr, grantInstruction(info.HostAppName, info.HostAppBundle))
		return 1
	}
	return 0
}

// grantInstruction names the terminal host app that must be granted — TCC
// attributes a shell-launched tool to its terminal, which users do not
// expect, and on macOS 26.1+ the grant can be enforced yet hidden from
// System Settings.
func grantInstruction(hostName, hostBundle string) string {
	host := "your terminal app"
	if hostName != "" {
		host = fmt.Sprintf("%s (%s)", hostName, hostBundle)
	}
	return fmt.Sprintf(`Accessibility is NOT granted for %s.

screenz is attributed to the terminal app that launched it, so the grant
must be given to that app, not to the screenz binary:

  1. Open System Settings → Privacy & Security → Accessibility.
  2. Enable (or add, via +) %s.
  3. Quit and reopen the terminal app, then run 'screenz doctor' again.

If the app is already listed and enabled but this keeps failing, reset its
grant and re-add it:

  tccutil reset Accessibility %s
`, host, host, orWord(hostBundle, "the terminal app"))
}

func orWord(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
