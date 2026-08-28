//go:build integration && darwin

package mac

import (
	"fmt"
	"testing"
)

// The integration tier runs against the real window server on a trusted GUI
// session. A missing Accessibility grant FAILS the test with the grant
// instruction — it never skips (ADR1.3): a capability-gated skip would turn
// the suite green with zero coverage of the only seam that matters.

func TestSymbolsAllBind(t *testing.T) {
	if m := Missing(); len(m) != 0 {
		t.Fatalf("missing symbols: %v", m)
	}
}

func TestTrusted(t *testing.T) {
	if !Trusted() {
		name, bundle := HostApp()
		t.Fatalf("AXIsProcessTrusted() = false. Grant Accessibility to %s (%s) in "+
			"System Settings → Privacy & Security → Accessibility, then rerun 'make itest'.", name, bundle)
	}
}

func TestDisplaysMatchScreens(t *testing.T) {
	displays := Displays()
	screens, primaryH := Screens()
	if len(displays) == 0 {
		t.Fatal("no CoreGraphics displays")
	}
	if len(displays) != len(screens) {
		t.Fatalf("CGGetActiveDisplayList reports %d displays, NSScreen.screens %d", len(displays), len(screens))
	}
	byID := map[uint32]DisplayRaw{}
	for _, d := range displays {
		byID[d.ID] = d
		if d.UUID == "" {
			t.Errorf("display %d has empty UUID", d.ID)
		}
	}
	for _, s := range screens {
		d, ok := byID[s.ID]
		if !ok {
			t.Fatalf("screen %q (id %d) has no CG display", s.Name, s.ID)
		}
		if got := NSToAX(s.Frame, primaryH); got != d.Bounds {
			t.Errorf("screen %q: NSToAX(frame) = %+v, CGDisplayBounds = %+v", s.Name, got, d.Bounds)
		}
	}
}

func TestNSToAXRoundTrip(t *testing.T) {
	screens, primaryH := Screens()
	for _, s := range screens {
		if got := AXToNS(NSToAX(s.Frame, primaryH), primaryH); got != s.Frame {
			t.Errorf("screen %q: AXToNS(NSToAX(frame)) = %+v, want %+v", s.Name, got, s.Frame)
		}
	}
}

func TestWindowListAndBundleIDs(t *testing.T) {
	wins := WindowList()
	if len(wins) == 0 {
		t.Fatal("CGWindowListCopyWindowInfo returned no windows")
	}
	apps := RunningApps()
	if len(apps) == 0 {
		t.Fatal("NSWorkspace reports no regular applications")
	}
	for _, a := range apps {
		if a.Bundle == "" {
			continue // rare: regular app with no bundle id
		}
		if got := BundleID(a.PID); got != a.Bundle {
			t.Errorf("BundleID(%d) = %q, NSWorkspace says %q (%s)", a.PID, got, a.Bundle, a.Name)
		}
	}
}

func TestHostAppResolves(t *testing.T) {
	name, bundle := HostApp()
	if bundle == "" {
		t.Fatal("HostApp() found no ancestor application — the TCC client cannot be named")
	}
	fmt.Printf("host app: %s (%s)\n", name, bundle)
}
