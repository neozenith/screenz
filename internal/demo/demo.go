// Package demo replays a recorded world so demonstration text can be
// curated without the recorded hardware attached (ADR-0018). The world
// file is the `screenz status --json` schema, so a real machine is
// captured with that command and replayed via SCREENZ_DEMO=<file>.
// Placement is simulated: the read back is fabricated as the target, so
// demo output must never be presented as a verification record. Doctor
// discloses demo mode whenever the variable is set.
package demo

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/neozenith/screenz/internal/discover"
	"github.com/neozenith/screenz/internal/layout"
	"github.com/neozenith/screenz/internal/mac"
	"github.com/neozenith/screenz/internal/place"
)

// World mirrors the `screenz status --json` payload (schema 1).
type World struct {
	Schema   int                `json:"schema"`
	Displays []discover.Display `json:"displays"`
	Windows  []discover.Window  `json:"windows"`
}

// Load reads a world file and returns it as a discovery snapshot. AX
// elements do not survive serialisation, so the snapshot carries zero
// elements; only the simulated Place below may consume them.
func Load(path string) (discover.Snapshot, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return discover.Snapshot{}, err
	}
	var w World
	if err := json.Unmarshal(raw, &w); err != nil {
		return discover.Snapshot{}, fmt.Errorf("%s: %w", path, err)
	}
	if w.Schema != 1 {
		return discover.Snapshot{}, fmt.Errorf("%s: unsupported world schema %d (want 1)", path, w.Schema)
	}
	return discover.Snapshot{Displays: w.Displays, Windows: w.Windows}, nil
}

// Place simulates a placement by reporting the target as the read-back
// frame. It exists solely for demo capture and must never be wired
// outside SCREENZ_DEMO (ADR-0018).
func Place(_, _ mac.AXElement, target mac.CGRect, _ layout.Tolerance) place.Result {
	return place.Result{Requested: target, Before: target, Actual: target, Attempts: 1, OK: true}
}
