//go:build darwin

package mac

import "unsafe"

const (
	kCFStringEncodingUTF8 = 0x08000100
	kCFNumberSInt64Type   = 4
)

var (
	cfRelease                 func(CFRef)
	cfRetain                  func(CFRef) CFRef
	cfGetTypeID               func(CFRef) uintptr
	cfBooleanGetTypeID        func() uintptr
	cfStringCreateWithCString func(alloc uintptr, s string, enc uint32) CFRef
	cfStringGetCString        func(s CFRef, buf *byte, bufLen int64, enc uint32) bool
	cfStringGetLength         func(s CFRef) int64
	cfArrayGetCount           func(a CFRef) int64
	cfArrayGetValueAtIndex    func(a CFRef, i int64) CFRef
	cfDictionaryGetValue      func(d CFRef, key CFRef) CFRef
	cfNumberGetValue          func(n CFRef, typ int64, out unsafe.Pointer) bool
	cfBooleanGetValue         func(b CFRef) bool
	cfUUIDCreateString        func(alloc uintptr, uuid CFRef) CFRef
)

func cfstr(s string) CFRef { return cfStringCreateWithCString(0, s, kCFStringEncodingUTF8) }

// Release releases any owned CF reference; safe on zero.
func Release(v CFRef) {
	if v != 0 {
		cfRelease(v)
	}
}

func goString(s CFRef) string {
	if s == 0 {
		return ""
	}
	n := cfStringGetLength(s)*4 + 1
	buf := make([]byte, n)
	if !cfStringGetCString(s, &buf[0], n, kCFStringEncodingUTF8) {
		return ""
	}
	for i, b := range buf {
		if b == 0 {
			return string(buf[:i])
		}
	}
	return string(buf)
}

func dictInt64(d CFRef, key string) (int64, bool) {
	k := cfstr(key)
	defer Release(k)
	v := cfDictionaryGetValue(d, k)
	if v == 0 {
		return 0, false
	}
	var out int64
	ok := cfNumberGetValue(v, kCFNumberSInt64Type, unsafe.Pointer(&out))
	return out, ok
}

func dictString(d CFRef, key string) string {
	k := cfstr(key)
	defer Release(k)
	return goString(cfDictionaryGetValue(d, k))
}
