//go:build darwin

// Command screenz organises groups of application windows across displays.
// main only wires the impure internal/mac gatherers into internal/cli's pure
// command pipeline; all business logic lives under internal/.
package main

import (
	"os"

	"github.com/joshpeak/screenz/internal/cli"
	"github.com/joshpeak/screenz/internal/mac"
)

func main() { os.Exit(run()) }

func run() int {
	home, _ := os.UserHomeDir()
	return cli.Run(os.Args[1:], os.Stdout, os.Stderr, cli.Deps{
		Getenv: os.Getenv,
		Home:   home,
		Sys:    sysInfo,
	})
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
