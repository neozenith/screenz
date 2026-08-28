//go:build darwin

package mac

// CGPoint, CGSize and CGRect are all-float structs: on darwin/arm64 a CGRect
// is a homogeneous floating-point aggregate returned in d0–d3, which purego's
// struct-return path handles (verified exact in the spike, ADR1.1).
type CGPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type CGSize struct {
	W float64 `json:"w"`
	H float64 `json:"h"`
}

type CGRect struct {
	Origin CGPoint `json:"origin"`
	Size   CGSize  `json:"size"`
}

// NSToAX converts a bottom-left-origin AppKit rect into top-left-origin
// CG/AX coordinates. primaryHeight is the height of NSScreen.screens[0].frame
// — the flip pivots on the primary screen, never the rect's own screen
// (Rectangle's CGExtension bug is the negative measure here).
func NSToAX(r CGRect, primaryHeight float64) CGRect {
	return CGRect{CGPoint{r.Origin.X, primaryHeight - (r.Origin.Y + r.Size.H)}, r.Size}
}

// AXToNS is the inverse flip.
func AXToNS(r CGRect, primaryHeight float64) CGRect {
	return CGRect{CGPoint{r.Origin.X, primaryHeight - (r.Origin.Y + r.Size.H)}, r.Size}
}
