//go:build integration && darwin

package place

import (
	"os/exec"
	"testing"
	"time"

	"github.com/neozenith/screenz/internal/discover"
	"github.com/neozenith/screenz/internal/layout"
	"github.com/neozenith/screenz/internal/mac"
)

// Opens a real TextEdit document, places it left-half on the built-in
// display through the shipped code path, asserts the read-back frame, and
// closes it (saving no). Fails, never skips, without the grant (ADR1.3).
func TestPlaceRealWindow(t *testing.T) {
	if !mac.Trusted() {
		name, bundle := mac.HostApp()
		t.Fatalf("Accessibility not granted; grant it to %s (%s) and rerun 'make itest'", name, bundle)
	}

	if out, err := exec.Command("osascript",
		"-e", `tell application "TextEdit" to activate`,
		"-e", `tell application "TextEdit" to make new document`).CombinedOutput(); err != nil {
		t.Fatalf("osascript could not open TextEdit: %v\n%s", err, out)
	}
	defer exec.Command("osascript",
		"-e", `tell application "TextEdit" to close front document saving no`).Run()

	// The fresh window can take a moment to appear in AX enumeration.
	var win discover.Window
	var snap discover.Snapshot
	found := false
	for i := 0; i < 20 && !found; i++ {
		time.Sleep(150 * time.Millisecond)
		snap = discover.Build(mac.Snapshot(2.0))
		for _, w := range snap.Windows {
			if w.Bundle == "com.apple.TextEdit" && w.State == discover.StateNormal {
				win, found = w, true
				break
			}
		}
	}
	if !found {
		t.Fatal("no normal TextEdit window appeared within 3s")
	}

	var builtin discover.Display
	for _, d := range snap.Displays {
		if d.BuiltIn {
			builtin = d
		}
	}
	if builtin.UUID == "" {
		t.Fatal("no built-in display found")
	}

	region, err := layout.ParseRegion("left-half")
	if err != nil {
		t.Fatal(err)
	}
	tol := layout.Tolerance{Value: layout.DefaultTolerance}
	target := region.Rect(builtin.VisibleFrame, 0, 1, 0)

	res := Place(snap.AppEl(win.PID), win.El(), target, tol)
	if res.Err != "" {
		t.Fatalf("placement error: %s (result %+v)", res.Err, res)
	}
	if !res.OK {
		t.Fatalf("frame not within tolerance after %d attempts: requested %+v, actual %+v",
			res.Attempts, res.Requested, res.Actual)
	}

	// Independent read back through a fresh AX query, not the returned struct.
	pos, _ := win.El().Point("AXPosition")
	size, _ := win.El().Size("AXSize")
	if got := (mac.CGRect{Origin: pos, Size: size}); !tol.Within(target, got) {
		t.Fatalf("fresh read back %+v not within tolerance of %+v", got, target)
	}
}
