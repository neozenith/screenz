// Package discover turns the raw mac snapshot into displays with stable
// identity and windows grouped by application. It is a pure transform —
// every function here is covered at 100% by `make check`; the impure gather
// lives in internal/mac and is exercised by `make itest`.
package discover

import (
	"sort"

	"github.com/neozenith/screenz/internal/mac"
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

// menuBarLayer is kCGMainMenuWindowLevel: the window server keeps one
// width-spanning window at this level per display for the menu bar.
const menuBarLayer = 24

// BuildDisplays joins CG displays and NSScreens on the display id and
// assigns Index by row (top to bottom) then minX (ADR2.3) — the ordering
// Rectangle users already live with. Index is 1-based.
//
// cgRows further carves the usable frame: on macOS 26 NSScreen.visibleFrame
// can report the full display height on external panels while the window
// server still reserves the menu bar strip (verified on this machine — the
// read-back clamped every window 30 pt short until the strip was carved).
// The strip is the layer-24 window anchored at the display's top-left.
func BuildDisplays(displays []mac.DisplayRaw, screens []mac.ScreenRaw, primaryH float64, cgRows []mac.CGWindowRaw) []Display {
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
			disp.VisibleFrame = carveMenuBar(mac.NSToAX(s.VisibleFrame, primaryH), d.Bounds, cgRows)
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

// carveMenuBar lowers the visible frame's top edge below any menu bar
// strip: a layer-24 row anchored at the display's top-left, spanning its
// width, no taller than 100 pt. When visibleFrame already excludes the
// strip (the built-in Retina does) this is a no-op.
func carveMenuBar(vis, frame mac.CGRect, cgRows []mac.CGWindowRaw) mac.CGRect {
	for _, r := range cgRows {
		if r.Layer != menuBarLayer ||
			r.Bounds.Origin.X != frame.Origin.X || r.Bounds.Origin.Y != frame.Origin.Y ||
			r.Bounds.Size.W < frame.Size.W || r.Bounds.Size.H > 100 {
			continue
		}
		if bottom := r.Bounds.Origin.Y + r.Bounds.Size.H; bottom > vis.Origin.Y {
			vis.Size.H -= bottom - vis.Origin.Y
			vis.Origin.Y = bottom
		}
	}
	return vis
}
