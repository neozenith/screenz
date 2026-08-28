package discover

import "github.com/joshpeak/screenz/internal/mac"

// Window states (ADR2.2): windows the tool cannot act on are reported with
// a state instead of being hidden — "moved 9 of 10" must never look like
// success, and Spaces have no public API so `offscreen` is a permanent skip.
const (
	StateNormal    = "normal"
	StateMinimized = "minimized"
	StateHidden    = "hidden"
	StateSheet     = "sheet"
	StateDialog    = "dialog"
	StateOffscreen = "offscreen"
)

// Window is one AX window resolved against the CG window list and displays.
// Frame is in AX global points (top-left origin). ID is the CG window number
// (0 when the window has no on-screen CG row, e.g. another Space).
// DisplayIndex names the Display the window sits on (0 when none intersect).
type Window struct {
	ID           int64      `json:"id"`
	PID          int64      `json:"pid"`
	Bundle       string     `json:"bundle"`
	App          string     `json:"app"`
	Title        string     `json:"title"`
	Role         string     `json:"role"`
	Subrole      string     `json:"subrole"`
	State        string     `json:"state"`
	Frame        mac.CGRect `json:"frame"`
	DisplayIndex int        `json:"display_index"`

	el mac.AXElement
}

// El returns the live AX element for placement.
func (w Window) El() mac.AXElement { return w.el }

// Snapshot is the fully resolved discovery result.
type Snapshot struct {
	Displays []Display `json:"displays"`
	Windows  []Window  `json:"windows"`
	AppErrs  []AppErr  `json:"app_errors,omitempty"`

	appEls map[int64]mac.AXElement
}

// AppErr records an application whose AX enumeration failed.
type AppErr struct {
	PID    int64  `json:"pid"`
	Bundle string `json:"bundle"`
	App    string `json:"app"`
	Err    string `json:"err"`
}

// AppEl returns the AX application element for a pid (needed to toggle
// AXEnhancedUserInterface during placement).
func (s Snapshot) AppEl(pid int64) mac.AXElement { return s.appEls[pid] }

// Build resolves a raw mac snapshot into displays and windows.
func Build(raw mac.SnapshotRaw) Snapshot {
	displays := BuildDisplays(raw.Displays, raw.Screens, raw.PrimaryH)
	snap := Snapshot{Displays: displays, appEls: map[int64]mac.AXElement{}}

	// Greedily consume layer-0 CG rows so two identical AX frames of the
	// same app claim distinct window numbers.
	free := make([]bool, len(raw.CGWindows))
	for i, w := range raw.CGWindows {
		free[i] = w.Layer == 0
	}
	for _, app := range raw.Apps {
		snap.appEls[app.App.PID] = app.AppEl
		if app.Err != "" {
			snap.AppErrs = append(snap.AppErrs, AppErr{PID: app.App.PID, Bundle: app.App.Bundle, App: app.App.Name, Err: app.Err})
			continue
		}
		for _, w := range app.Windows {
			frame := mac.CGRect{Origin: w.Pos, Size: w.Size}
			win := Window{
				PID:          app.App.PID,
				Bundle:       app.App.Bundle,
				App:          app.App.Name,
				Title:        w.Title,
				Role:         w.Role,
				Subrole:      w.Subrole,
				Frame:        frame,
				DisplayIndex: assignDisplay(frame, displays),
				el:           w.El,
			}
			cgIdx := matchCG(raw.CGWindows, free, app.App.PID, frame)
			if cgIdx >= 0 {
				win.ID = raw.CGWindows[cgIdx].Number
				free[cgIdx] = false
			}
			win.State = classify(w, app.Hidden, cgIdx >= 0)
			snap.Windows = append(snap.Windows, win)
		}
	}
	return snap
}

// classify decides the window state (ADR2.2). Priority: minimized and
// hidden are definite AX facts; sheet and dialog come from the role and
// subrole; a window absent from the on-screen CG list sits on another
// Space (or is genuinely off screen).
func classify(w mac.WindowRaw, appHidden, onScreen bool) string {
	switch {
	case w.Minimized:
		return StateMinimized
	case appHidden:
		return StateHidden
	case w.Subrole == "AXSheet" || w.Role == "AXSheet":
		return StateSheet
	case w.Subrole == "AXSystemDialog" || w.Subrole == "AXDialog":
		return StateDialog
	case !onScreen:
		return StateOffscreen
	}
	return StateNormal
}

// matchCG finds an unclaimed layer-0 CG row with the same owner pid and
// (within a point) the same bounds; -1 when none matches.
func matchCG(rows []mac.CGWindowRaw, free []bool, pid int64, frame mac.CGRect) int {
	for i, row := range rows {
		if !free[i] || row.OwnerPID != pid {
			continue
		}
		if near(row.Bounds.Origin.X, frame.Origin.X) && near(row.Bounds.Origin.Y, frame.Origin.Y) &&
			near(row.Bounds.Size.W, frame.Size.W) && near(row.Bounds.Size.H, frame.Size.H) {
			return i
		}
	}
	return -1
}

func near(a, b float64) bool {
	d := a - b
	return d >= -1 && d <= 1
}

// assignDisplay picks the display fully containing the frame, else the one
// with the largest intersection (lower index wins ties); 0 when none
// intersect. Regions later use the display's usable frame, so a window
// straddling two panels follows its bigger half — Rectangle's rule.
func assignDisplay(frame mac.CGRect, displays []Display) int {
	best, bestArea := 0, 0.0
	for _, d := range displays {
		if contains(d.Frame, frame) {
			return d.Index
		}
		if a := intersectArea(d.Frame, frame); a > bestArea {
			best, bestArea = d.Index, a
		}
	}
	return best
}

func contains(outer, inner mac.CGRect) bool {
	return inner.Origin.X >= outer.Origin.X && inner.Origin.Y >= outer.Origin.Y &&
		inner.Origin.X+inner.Size.W <= outer.Origin.X+outer.Size.W &&
		inner.Origin.Y+inner.Size.H <= outer.Origin.Y+outer.Size.H
}

func intersectArea(a, b mac.CGRect) float64 {
	x1 := max(a.Origin.X, b.Origin.X)
	y1 := max(a.Origin.Y, b.Origin.Y)
	x2 := min(a.Origin.X+a.Size.W, b.Origin.X+b.Size.W)
	y2 := min(a.Origin.Y+a.Size.H, b.Origin.Y+b.Size.H)
	if x2 <= x1 || y2 <= y1 {
		return 0
	}
	return (x2 - x1) * (y2 - y1)
}
