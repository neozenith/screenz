//go:build darwin

// Command screenz organises groups of application windows across displays.
// main only wires the impure internal/mac gatherers into internal/cli's pure
// command pipeline; all business logic lives under internal/.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/joshpeak/screenz/internal/cli"
	"github.com/joshpeak/screenz/internal/discover"
	"github.com/joshpeak/screenz/internal/mac"
	"github.com/joshpeak/screenz/internal/place"
)

func main() { os.Exit(run()) }

func run() int {
	home, _ := os.UserHomeDir()
	return cli.Run(os.Args[1:], os.Stdout, os.Stderr, cli.Deps{
		Getenv:   os.Getenv,
		Home:     home,
		Sys:      sysInfo,
		Snapshot: snapshot,
		Displays: displays,
		Place:    place.Place,
	})
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

// sysInfo gathers the doctor report from the live bridge. When symbols are
// missing the bound functions may be nil, so it stops at the report.
func sysInfo(prompt bool) cli.SysInfo {
	info := cli.SysInfo{MissingSymbols: mac.Missing()}
	if len(info.MissingSymbols) > 0 {
		return info
	}
	if prompt {
		info.Trusted = mac.TrustedWithPrompt()
	} else {
		info.Trusted = mac.Trusted()
	}
	info.HostAppName, info.HostAppBundle = mac.HostApp()
	screens, _ := mac.Screens()
	for _, s := range screens {
		info.DisplayNames = append(info.DisplayNames, s.Name)
	}
	return info
}
