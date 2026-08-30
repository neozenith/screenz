//go:build darwin

// Command screenz organises groups of application windows across displays.
// main only wires the impure internal/mac gatherers into internal/cli's pure
// command pipeline; all business logic lives under internal/.
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/neozenith/screenz/internal/cli"
	"github.com/neozenith/screenz/internal/demo"
	"github.com/neozenith/screenz/internal/discover"
	"github.com/neozenith/screenz/internal/mac"
	"github.com/neozenith/screenz/internal/place"
	"golang.org/x/sys/unix"
)

func main() { os.Exit(run()) }

func run() int {
	home, _ := os.UserHomeDir()
	exe, _ := os.Executable()
	deps := cli.Deps{
		Fetch:    fetch,
		ExePath:  exe,
		Getenv:   os.Getenv,
		Home:     home,
		Sys:      sysInfo,
		Snapshot: snapshot,
		Displays: displays,
		Place:    place.Place,
	}
	// Demo mode (ADR-0018): replay a recorded `status --json` world and
	// simulate placement so demonstration text renders without the
	// recorded hardware. Doctor discloses it; nothing else changes.
	if world := os.Getenv("SCREENZ_DEMO"); world != "" {
		snap, err := demo.Load(world)
		if err != nil {
			fmt.Fprintf(os.Stderr, "screenz: SCREENZ_DEMO: %v\n", err)
			return 1
		}
		deps.Snapshot = func() (discover.Snapshot, error) { return snap, nil }
		deps.Displays = func() ([]discover.Display, error) { return snap.Displays, nil }
		deps.Place = demo.Place
		realSys := deps.Sys
		deps.Sys = func(full bool) cli.SysInfo {
			info := realSys(full)
			info.DisplayNames = nil
			for _, d := range snap.Displays {
				info.DisplayNames = append(info.DisplayNames, d.Name)
			}
			return info
		}
	}
	return cli.Run(os.Args[1:], os.Stdout, os.Stderr, deps)
}

// fetch is the one HTTP dependency (screenz update); GitHub redirects
// release-asset downloads, which http.Client follows by default.
func fetch(url string) ([]byte, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// snapshot gathers and resolves the live window and display state. A 1 s
// AX messaging timeout keeps one hung app from stalling the sweep.
func snapshot() (discover.Snapshot, error) {
	if m := mac.Missing(); len(m) > 0 {
		return discover.Snapshot{}, fmt.Errorf("cannot bind macOS symbols: %s", strings.Join(m, ", "))
	}
	return discover.Build(mac.Snapshot(1.0)), nil
}

// displays resolves just the connected displays — no Accessibility needed,
// so profile status works before the grant exists.
func displays() ([]discover.Display, error) {
	if m := mac.Missing(); len(m) > 0 {
		return nil, fmt.Errorf("cannot bind macOS symbols: %s", strings.Join(m, ", "))
	}
	screens, primaryH := mac.Screens()
	return discover.BuildDisplays(mac.Displays(), screens, primaryH, mac.WindowList()), nil
}

// sysInfo gathers the machine report from the live bridge. When symbols
// are missing the bound functions may be nil, so it stops at the report.
// Only doctor (full=true, which also asks with the TCC prompt) needs the
// display names, OS version and quarantine state; status and apply only
// check the grant, so the host-app walk runs only when the grant is
// missing and its name is needed for the error message.
func sysInfo(full bool) cli.SysInfo {
	info := cli.SysInfo{MissingSymbols: mac.Missing()}
	if len(info.MissingSymbols) > 0 {
		return info
	}
	if full {
		info.Trusted = mac.TrustedWithPrompt()
	} else {
		info.Trusted = mac.Trusted()
	}
	if full || !info.Trusted {
		info.HostAppName, info.HostAppBundle = mac.HostApp()
	}
	if !full {
		return info
	}
	screens, _ := mac.Screens()
	for _, s := range screens {
		info.DisplayNames = append(info.DisplayNames, s.Name)
	}
	info.OSVersion, _ = unix.Sysctl("kern.osproductversion")
	if exe, err := os.Executable(); err == nil {
		info.ExePath = exe
		// Getxattr returns the attribute size when present; ENOATTR when not.
		if n, err := unix.Getxattr(exe, "com.apple.quarantine", nil); err == nil && n >= 0 {
			info.Quarantined = true
		}
	}
	return info
}
