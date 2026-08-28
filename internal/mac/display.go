//go:build darwin

package mac

var (
	cgMainDisplayID                        func() uint32
	cgGetActiveDisplayList                 func(max uint32, ids *uint32, count *uint32) int32
	cgDisplayBounds                        func(id uint32) CGRect
	cgDisplayIsBuiltin                     func(id uint32) uint32
	cgDisplayIsMain                        func(id uint32) uint32
	cgDisplaySerialNumber                  func(id uint32) uint32
	cgDisplayVendorNumber                  func(id uint32) uint32
	cgDisplayModelNumber                   func(id uint32) uint32
	cgDisplayCopyDisplayMode               func(id uint32) CFRef
	cgDisplayCreateUUIDFromDisplayID       func(id uint32) CFRef
	cgDisplayModeGetPixelWidth             func(m CFRef) uint64
	cgDisplayModeGetPixelHeight            func(m CFRef) uint64
	cgWindowListCopyWindowInfo             func(opts uint32, relativeTo uint32) CFRef
	cgRectMakeWithDictionaryRepresentation func(d CFRef, out *CGRect) bool
)

// Displays enumerates the active CoreGraphics displays.
func Displays() []DisplayRaw {
	Load()
	ids := make([]uint32, 16)
	var n uint32
	if rc := cgGetActiveDisplayList(uint32(len(ids)), &ids[0], &n); rc != 0 {
		return nil
	}
	main := cgMainDisplayID()
	out := make([]DisplayRaw, 0, n)
	for _, id := range ids[:n] {
		d := DisplayRaw{
			ID:      id,
			Main:    id == main,
			BuiltIn: cgDisplayIsBuiltin(id) != 0,
			Vendor:  cgDisplayVendorNumber(id),
			Model:   cgDisplayModelNumber(id),
			Serial:  cgDisplaySerialNumber(id),
			Bounds:  cgDisplayBounds(id),
		}
		if uuidRef := cgDisplayCreateUUIDFromDisplayID(id); uuidRef != 0 {
			s := cfUUIDCreateString(0, uuidRef)
			d.UUID = goString(s)
			Release(s)
			Release(uuidRef)
		}
		if mode := cgDisplayCopyDisplayMode(id); mode != 0 {
			d.PixelW = cgDisplayModeGetPixelWidth(mode)
			d.PixelH = cgDisplayModeGetPixelHeight(mode)
			Release(mode)
		}
		out = append(out, d)
	}
	return out
}
