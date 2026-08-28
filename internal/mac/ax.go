//go:build darwin

package mac

import (
	"fmt"
	"unsafe"

	"github.com/ebitengine/purego/objc"
)

var (
	axIsProcessTrusted             func() bool
	axIsProcessTrustedWithOptions  func(opts CFRef) bool
	axUIElementCreateApplication   func(pid int32) CFRef
	axUIElementCopyAttributeValue  func(el CFRef, attr CFRef, out *CFRef) int32
	axUIElementSetAttributeValue   func(el CFRef, attr CFRef, val CFRef) int32
	axUIElementSetMessagingTimeout func(el CFRef, seconds float32) int32
	axValueGetValue                func(v CFRef, typ uint32, out unsafe.Pointer) bool
	axValueCreate                  func(typ uint32, in unsafe.Pointer) CFRef
)

const (
	kAXValueTypeCGPoint = 1
	kAXValueTypeCGSize  = 2
)

// axErrors maps the AXError codes this tool can meet; -25204 CannotComplete
// is expected from the Window Server pseudo-app and hung processes.
var axErrors = map[int32]string{
	-25200: "failure",
	-25201: "illegal argument",
	-25202: "invalid AXUIElement",
	-25204: "cannot complete (app not responding, or Window Server)",
	-25205: "attribute unsupported",
	-25211: "API disabled (Accessibility not granted)",
	-25212: "no value",
}

// AXErrorString names an AXError return code.
func AXErrorString(rc int32) string {
	if s, ok := axErrors[rc]; ok {
		return fmt.Sprintf("%s (AXError %d)", s, rc)
	}
	return fmt.Sprintf("AXError %d", rc)
}

// Trusted reports whether this process (via its terminal host app) holds the
// Accessibility grant. Every AX-using command checks it first (ADR1.2).
func Trusted() bool {
	Load()
	return axIsProcessTrusted()
}

// TrustedWithPrompt asks like Trusted but lets macOS show the one-time
// system prompt that deep-links to the Accessibility pane. The options
// dictionary is built through toll-free-bridged Foundation objects; NSNumber
// bool yields the shared CFBoolean and NSDictionary compares keys with
// isEqual, so a manually created key string matches the framework constant.
func TrustedWithPrompt() bool {
	Load()
	key := nsString("AXTrustedCheckOptionPrompt")
	yes := objc.Send[objc.ID](objc.ID(objc.GetClass("NSNumber")), objc.RegisterName("numberWithBool:"), 1)
	opts := objc.Send[objc.ID](objc.ID(objc.GetClass("NSDictionary")), objc.RegisterName("dictionaryWithObject:forKey:"), yes, key)
	return axIsProcessTrustedWithOptions(CFRef(opts))
}

// AXElement is one owned AX object reference (application or window).
type AXElement struct{ ref CFRef }

// Release frees the underlying AX reference.
func (e AXElement) Release() { Release(e.ref) }

// AXApp creates the AX application element for a pid with a messaging
// timeout, so one hung app cannot stall a 16-window apply for the 6 s
// default per call (negative measure; Rectangle sets the same timeout).
func AXApp(pid int32, timeoutSeconds float32) AXElement {
	Load()
	app := axUIElementCreateApplication(pid)
	if timeoutSeconds > 0 {
		axUIElementSetMessagingTimeout(app, timeoutSeconds)
	}
	return AXElement{app}
}

// Windows returns the owned AXWindow elements of an application element.
func (e AXElement) Windows() ([]AXElement, error) {
	arr, rc := e.attr("AXWindows")
	if rc != 0 || arr == 0 {
		return nil, fmt.Errorf("AXWindows: %s", AXErrorString(rc))
	}
	defer Release(arr)
	n := cfArrayGetCount(arr)
	out := make([]AXElement, 0, n)
	for i := int64(0); i < n; i++ {
		w := cfArrayGetValueAtIndex(arr, i)
		cfRetain(w) // the array owns its elements; retain before the array dies
		out = append(out, AXElement{w})
	}
	return out, nil
}

func (e AXElement) attr(name string) (CFRef, int32) {
	k := cfstr(name)
	defer Release(k)
	var v CFRef
	rc := axUIElementCopyAttributeValue(e.ref, k, &v)
	return v, rc
}

// String reads a string attribute such as AXTitle, AXRole, AXSubrole.
func (e AXElement) String(name string) string {
	v, _ := e.attr(name)
	if v == 0 {
		return ""
	}
	defer Release(v)
	return goString(v)
}

// Bool reads a boolean attribute such as AXMinimized or AXEnhancedUserInterface.
// ok is false when the attribute is absent or not a boolean.
func (e AXElement) Bool(name string) (value, ok bool) {
	v, _ := e.attr(name)
	if v == 0 {
		return false, false
	}
	defer Release(v)
	if cfGetTypeID(v) != cfBooleanGetTypeID() {
		return false, false
	}
	return cfBooleanGetValue(v), true
}

// Point reads an AXValue point attribute (AXPosition).
func (e AXElement) Point(name string) (CGPoint, bool) {
	v, _ := e.attr(name)
	if v == 0 {
		return CGPoint{}, false
	}
	defer Release(v)
	var p CGPoint
	ok := axValueGetValue(v, kAXValueTypeCGPoint, unsafe.Pointer(&p))
	return p, ok
}

// Size reads an AXValue size attribute (AXSize).
func (e AXElement) Size(name string) (CGSize, bool) {
	v, _ := e.attr(name)
	if v == 0 {
		return CGSize{}, false
	}
	defer Release(v)
	var s CGSize
	ok := axValueGetValue(v, kAXValueTypeCGSize, unsafe.Pointer(&s))
	return s, ok
}

// SetPoint writes an AXValue point attribute; returns the AXError code.
func (e AXElement) SetPoint(name string, p CGPoint) int32 {
	k := cfstr(name)
	defer Release(k)
	v := axValueCreate(kAXValueTypeCGPoint, unsafe.Pointer(&p))
	defer Release(v)
	return axUIElementSetAttributeValue(e.ref, k, v)
}

// SetSize writes an AXValue size attribute; returns the AXError code.
func (e AXElement) SetSize(name string, s CGSize) int32 {
	k := cfstr(name)
	defer Release(k)
	v := axValueCreate(kAXValueTypeCGSize, unsafe.Pointer(&s))
	defer Release(v)
	return axUIElementSetAttributeValue(e.ref, k, v)
}

// SetBool writes a boolean attribute (AXEnhancedUserInterface); the CFBoolean
// singletons come from toll-free-bridged NSNumber bools.
func (e AXElement) SetBool(name string, value bool) int32 {
	k := cfstr(name)
	defer Release(k)
	b := 0
	if value {
		b = 1
	}
	v := objc.Send[objc.ID](objc.ID(objc.GetClass("NSNumber")), objc.RegisterName("numberWithBool:"), b)
	return axUIElementSetAttributeValue(e.ref, k, CFRef(v))
}
