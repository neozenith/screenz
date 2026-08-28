//go:build darwin

package mac

import "github.com/ebitengine/purego/objc"

var (
	selScreens            = objc.RegisterName("screens")
	selCount              = objc.RegisterName("count")
	selObjectAtIndex      = objc.RegisterName("objectAtIndex:")
	selFrame              = objc.RegisterName("frame")
	selVisibleFrame       = objc.RegisterName("visibleFrame")
	selLocalizedName      = objc.RegisterName("localizedName")
	selBackingScaleFactor = objc.RegisterName("backingScaleFactor")
	selDeviceDescription  = objc.RegisterName("deviceDescription")
	selObjectForKey       = objc.RegisterName("objectForKey:")
	selStringWithUTF8     = objc.RegisterName("stringWithUTF8String:")
	selUnsignedIntValue   = objc.RegisterName("unsignedIntValue")
)

func nsString(s string) objc.ID {
	return objc.Send[objc.ID](objc.ID(objc.GetClass("NSString")), selStringWithUTF8, s)
}

// nsStringToGo reads an NSString via its toll-free CFString bridge, avoiding
// any raw char* handling (keeps `go vet`'s unsafeptr check clean).
func nsStringToGo(id objc.ID) string { return goString(CFRef(id)) }

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

// Screens enumerates NSScreen.screens and returns the primary screen height
// (screens[0].frame height) — the anchor for every NS→AX coordinate flip.
func Screens() ([]ScreenRaw, float64) {
	Load()
	cls := objc.GetClass("NSScreen")
	arr := objc.Send[objc.ID](objc.ID(cls), selScreens)
	n := objc.Send[uint64](arr, selCount)
	keyScreenNumber := nsString("NSScreenNumber")
	var out []ScreenRaw
	var primaryH float64
	for i := uint64(0); i < n; i++ {
		scr := objc.Send[objc.ID](arr, selObjectAtIndex, i)
		frame := objc.Send[CGRect](scr, selFrame)
		if i == 0 {
			primaryH = frame.Size.H
		}
		desc := objc.Send[objc.ID](scr, selDeviceDescription)
		num := objc.Send[objc.ID](desc, selObjectForKey, keyScreenNumber)
		out = append(out, ScreenRaw{
			ID:           objc.Send[uint32](num, selUnsignedIntValue),
			Name:         nsStringToGo(objc.Send[objc.ID](scr, selLocalizedName)),
			Scale:        objc.Send[float64](scr, selBackingScaleFactor),
			Frame:        frame,
			VisibleFrame: objc.Send[CGRect](scr, selVisibleFrame),
		})
	}
	return out, primaryH
}
