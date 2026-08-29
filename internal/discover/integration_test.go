//go:build integration && darwin

package discover

import (
	"testing"

	"github.com/neozenith/screenz/internal/mac"
)

// Real-machine invariants over the live window set (ADR1.3: fails, never
// skips, without the Accessibility grant — mac's own integration tests name
// the grant instruction).

func TestSnapshotInvariants(t *testing.T) {
	if !mac.Trusted() {
		name, bundle := mac.HostApp()
		t.Fatalf("Accessibility not granted; grant it to %s (%s) and rerun", name, bundle)
	}
	snap := Build(mac.Snapshot(1.0))

	if len(snap.Displays) == 0 {
		t.Fatal("no displays discovered")
	}
	byIndex := map[int]Display{}
	for _, d := range snap.Displays {
		if d.UUID == "" {
			t.Errorf("display %q (index %d) has empty UUID", d.Name, d.Index)
		}
		byIndex[d.Index] = d
	}

	for _, w := range snap.Windows {
		if w.DisplayIndex == 0 {
			continue // no display intersects (e.g. minimized with a stale frame)
		}
		d, ok := byIndex[w.DisplayIndex]
		if !ok {
			t.Errorf("window %q names display %d which does not exist", w.Title, w.DisplayIndex)
			continue
		}
		if intersectArea(d.Frame, w.Frame) <= 0 {
			t.Errorf("window %q (frame %+v) does not intersect display %d (%+v)", w.Title, w.Frame, d.Index, d.Frame)
		}
	}
}

func TestWindowCountsMatchAX(t *testing.T) {
	if !mac.Trusted() {
		name, bundle := mac.HostApp()
		t.Fatalf("Accessibility not granted; grant it to %s (%s) and rerun", name, bundle)
	}
	raw := mac.Snapshot(1.0)
	snap := Build(raw)
	counts := map[int64]int{}
	for _, w := range snap.Windows {
		counts[w.PID]++
	}
	for _, app := range raw.Apps {
		if app.Err != "" {
			continue
		}
		if counts[app.App.PID] != len(app.Windows) {
			t.Errorf("pid %d (%s): discover kept %d windows, AX enumerated %d",
				app.App.PID, app.App.Bundle, counts[app.App.PID], len(app.Windows))
		}
	}
}
