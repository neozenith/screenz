// Package discover turns the raw mac snapshot into displays with stable
// identity and windows grouped by application. It is a pure transform —
// every function here is covered at 100% by `make check`; the impure gather
// lives in internal/mac and is exercised by `make itest`.
package discover

import (
	"sort"

	"github.com/joshpeak/screenz/internal/mac"
)

// Display joins one CoreGraphics display with its NSScreen. Frames are in
// AX global points (top-left origin, y down). UUID is the only stable key —
// the CG display ID is a small unstable int, and two identical panels can
// share the same serial number (spike fact: both LU28R55s report 1129796439).
type Display struct {
	Index        int        `json:"index"`
	ID           uint32     `json:"id"`
	UUID         string     `json:"uuid"`
	Name         string     `json:"name"`
	Serial       uint32     `json:"serial"`
	Vendor       uint32     `json:"vendor"`
	Model        uint32     `json:"model"`
	Main         bool       `json:"main"`
	BuiltIn      bool       `json:"built_in"`
	Scale        float64    `json:"scale"`
	Frame        mac.CGRect `json:"frame"`
	VisibleFrame mac.CGRect `json:"visible_frame"`
	PixelW       uint64     `json:"pixel_w"`
	PixelH       uint64     `json:"pixel_h"`
}

// BuildDisplays joins CG displays and NSScreens on the display id and
// assigns Index by row (top to bottom) then minX (ADR2.3) — the ordering
// Rectangle users already live with. Index is 1-based.
func BuildDisplays(displays []mac.DisplayRaw, screens []mac.ScreenRaw, primaryH float64) []Display {
	byID := map[uint32]mac.ScreenRaw{}
	for _, s := range screens {
		byID[s.ID] = s
	}
	out := make([]Display, 0, len(displays))
	for _, d := range displays {
		disp := Display{
			ID:      d.ID,
			UUID:    d.UUID,
			Serial:  d.Serial,
			Vendor:  d.Vendor,
			Model:   d.Model,
			Main:    d.Main,
			BuiltIn: d.BuiltIn,
			Frame:   d.Bounds,
			PixelW:  d.PixelW,
			PixelH:  d.PixelH,
		}
		if s, ok := byID[d.ID]; ok {
			disp.Name = s.Name
			disp.Scale = s.Scale
			disp.VisibleFrame = mac.NSToAX(s.VisibleFrame, primaryH)
		}
		out = append(out, disp)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Frame.Origin.Y != out[j].Frame.Origin.Y {
			return out[i].Frame.Origin.Y < out[j].Frame.Origin.Y
		}
		return out[i].Frame.Origin.X < out[j].Frame.Origin.X
	})
	for i := range out {
		out[i].Index = i + 1
	}
	return out
}
