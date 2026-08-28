//go:build darwin

// Package place moves one window to a target frame and verifies the result
// by reading it back. It is impure (talks to AX) and is exercised by
// `make itest` on real windows (ADR1.3); the tolerance judgment it applies
// is layout.Tolerance, covered purely.
package place

import (
	"time"

	"github.com/joshpeak/screenz/internal/layout"
	"github.com/joshpeak/screenz/internal/mac"
)

// Place moves one window to target and verifies every edge against the
// rule's tolerance. The recipe is settled prior art (Rectangle):
//
//  1. read AXEnhancedUserInterface on the app element and disable it if set
//     — Electron/Chromium windows otherwise "move in stages" or stop short;
//  2. set size, position, size — a window moved to another display keeps
//     its old size if size is set once, because macOS clamps the size to
//     the source display until the position lands;
//  3. read the frame back; on mismatch re-apply immediately, then once more
//     after 25 ms (ADR3.3);
//  4. restore AXEnhancedUserInterface.
//
// It never activates the application (ADR3.2) — activation exists for a
// single focused window and would strobe focus across a 16-window apply.
func Place(app, win mac.AXElement, target mac.CGRect, tol layout.Tolerance) Result {
	res := Result{Requested: target}
	res.Before.Origin, _ = win.Point("AXPosition")
	res.Before.Size, _ = win.Size("AXSize")

	if eui, ok := app.Bool("AXEnhancedUserInterface"); ok && eui {
		app.SetBool("AXEnhancedUserInterface", false)
		defer app.SetBool("AXEnhancedUserInterface", true)
	}

	for attempt := 1; attempt <= 3; attempt++ {
		res.Attempts = attempt
		if attempt == 3 {
			time.Sleep(25 * time.Millisecond)
		}
		if rc := win.SetSize("AXSize", target.Size); rc != 0 {
			res.Err = "set size: " + mac.AXErrorString(rc)
		}
		if rc := win.SetPoint("AXPosition", target.Origin); rc != 0 {
			res.Err = "set position: " + mac.AXErrorString(rc)
		}
		if rc := win.SetSize("AXSize", target.Size); rc != 0 {
			res.Err = "set size: " + mac.AXErrorString(rc)
		}
		res.Actual.Origin, _ = win.Point("AXPosition")
		res.Actual.Size, _ = win.Size("AXSize")
		if tol.Within(target, res.Actual) {
			res.OK = true
			return res
		}
	}
	return res
}
