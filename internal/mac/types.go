package mac

// This file is deliberately unconstrained by build tags: the pure packages
// (discover, layout, plan, rule, profile, cli) compile and test on any OS —
// the release workflow runs `make check` on ubuntu-latest — while every
// function that talks to macOS stays behind //go:build darwin, and
// cmd/screenz itself refuses to build elsewhere (ADR6.1).

// CFRef is any CFTypeRef (also used for toll-free-bridged NSObject ids).
type CFRef uintptr

// AXElement is one owned AX object reference (application or window).
type AXElement struct{ ref CFRef }

// DisplayRaw is one CoreGraphics display as reported by the window server.
// Bounds is in global points with a top-left origin. PixelW/PixelH come from
// the current display mode because CGDisplayPixelsWide returns points on
// Retina panels (spike fact).
type DisplayRaw struct {
	ID      uint32
	Main    bool
	BuiltIn bool
	UUID    string
	Vendor  uint32
	Model   uint32
	Serial  uint32
	PixelW  uint64
	PixelH  uint64
	Bounds  CGRect
}

// ScreenRaw is one NSScreen. Frame and VisibleFrame keep AppKit's
// bottom-left origin; discover flips them with NSToAX. VisibleFrame is the
// usable area — the menu bar and Dock are already carved out.
type ScreenRaw struct {
	ID           uint32 // deviceDescription[NSScreenNumber], joins DisplayRaw.ID
	Name         string
	Scale        float64
	Frame        CGRect
	VisibleFrame CGRect
}

// CGWindowRaw is one row of CGWindowListCopyWindowInfo. Titles are omitted
// without the Screen Recording grant, so titles always come from AX instead.
type CGWindowRaw struct {
	Number    int64
	OwnerPID  int64
	OwnerName string
	Layer     int64
	Bounds    CGRect
}

// AppRaw is one running application with the regular activation policy.
type AppRaw struct {
	PID    int64
	Bundle string
	Name   string
}

// WindowRaw is one AX window with the attributes discovery needs. El stays
// alive for later placement in the same process. FrameOK is false when the
// AXPosition/AXSize read failed — Pos/Size are then zero values that must
// not be mistaken for a real frame.
type WindowRaw struct {
	El        AXElement
	Title     string
	Role      string
	Subrole   string
	Minimized bool
	FrameOK   bool
	Pos       CGPoint
	Size      CGSize
}

// AppWindows is one application's AX enumeration result.
type AppWindows struct {
	App     AppRaw
	AppEl   AXElement
	Hidden  bool
	Err     string
	Windows []WindowRaw
}

// SnapshotRaw is everything the pure discovery layer needs from the OS.
type SnapshotRaw struct {
	Displays  []DisplayRaw
	Screens   []ScreenRaw
	PrimaryH  float64
	Apps      []AppWindows
	CGWindows []CGWindowRaw
}
