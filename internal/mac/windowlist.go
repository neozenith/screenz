//go:build darwin

package mac

import (
	"os"

	"github.com/ebitengine/purego/objc"
	"golang.org/x/sys/unix"
)

const (
	kCGWindowListOptionOnScreenOnly     = 1 << 0
	kCGWindowListExcludeDesktopElements = 1 << 4
)

// CGWindowRaw is one row of CGWindowListCopyWindowInfo. Titles are omitted
// without the Screen Recording grant, so titles always come from AX instead.
type CGWindowRaw struct {
	Number    int64
	OwnerPID  int64
	OwnerName string
	Layer     int64
	Bounds    CGRect
}

// WindowList returns the on-screen CG window rows (all layers; callers
// filter on Layer == 0 for ordinary windows).
func WindowList() []CGWindowRaw {
	Load()
	arr := cgWindowListCopyWindowInfo(kCGWindowListOptionOnScreenOnly|kCGWindowListExcludeDesktopElements, 0)
	if arr == 0 {
		return nil
	}
	defer Release(arr)
	kBounds := cfstr("kCGWindowBounds")
	defer Release(kBounds)
	n := cfArrayGetCount(arr)
	out := make([]CGWindowRaw, 0, n)
	for i := int64(0); i < n; i++ {
		d := cfArrayGetValueAtIndex(arr, i)
		w := CGWindowRaw{OwnerName: dictString(d, "kCGWindowOwnerName")}
		w.Number, _ = dictInt64(d, "kCGWindowNumber")
		w.OwnerPID, _ = dictInt64(d, "kCGWindowOwnerPID")
		w.Layer, _ = dictInt64(d, "kCGWindowLayer")
		if bd := cfDictionaryGetValue(d, kBounds); bd != 0 {
			cgRectMakeWithDictionaryRepresentation(bd, &w.Bounds)
		}
		out = append(out, w)
	}
	return out
}

var (
	selRunningAppWithPID   = objc.RegisterName("runningApplicationWithProcessIdentifier:")
	selBundleIdentifier    = objc.RegisterName("bundleIdentifier")
	selLocalizedNameApp    = objc.RegisterName("localizedName")
	selSharedWorkspace     = objc.RegisterName("sharedWorkspace")
	selRunningApplications = objc.RegisterName("runningApplications")
	selActivationPolicy    = objc.RegisterName("activationPolicy")
	selProcessIdentifier   = objc.RegisterName("processIdentifier")
)

func runningApp(pid int64) objc.ID {
	cls := objc.GetClass("NSRunningApplication")
	return objc.Send[objc.ID](objc.ID(cls), selRunningAppWithPID, int32(pid))
}

// BundleID resolves a pid to its bundle identifier ("" for non-app processes).
func BundleID(pid int64) string {
	Load()
	app := runningApp(pid)
	if app == 0 {
		return ""
	}
	return nsStringToGo(objc.Send[objc.ID](app, selBundleIdentifier))
}

// AppRaw is one running application with the regular activation policy.
type AppRaw struct {
	PID    int64
	Bundle string
	Name   string
}

// RunningApps lists NSWorkspace's regular-activation-policy applications, so
// hidden apps and apps whose windows sit on another Space are still seen
// (CGWindowList alone omits them — spike fact).
func RunningApps() []AppRaw {
	Load()
	ws := objc.Send[objc.ID](objc.ID(objc.GetClass("NSWorkspace")), selSharedWorkspace)
	arr := objc.Send[objc.ID](ws, selRunningApplications)
	n := objc.Send[uint64](arr, selCount)
	var out []AppRaw
	for i := uint64(0); i < n; i++ {
		app := objc.Send[objc.ID](arr, selObjectAtIndex, i)
		if objc.Send[int64](app, selActivationPolicy) != 0 { // 0 == NSApplicationActivationPolicyRegular
			continue
		}
		out = append(out, AppRaw{
			PID:    int64(objc.Send[int32](app, selProcessIdentifier)),
			Bundle: nsStringToGo(objc.Send[objc.ID](app, selBundleIdentifier)),
			Name:   nsStringToGo(objc.Send[objc.ID](app, selLocalizedNameApp)),
		})
	}
	return out
}

// HostApp walks the parent-process chain to the first ancestor that is a real
// application and returns its name and bundle id. That app — not the screenz
// binary — is the TCC client that must hold the Accessibility grant (ADR6.2).
func HostApp() (name, bundle string) {
	Load()
	pid := os.Getpid()
	for i := 0; i < 20 && pid > 1; i++ {
		kp, err := unix.SysctlKinfoProc("kern.proc.pid", pid)
		if err != nil {
			return "", ""
		}
		ppid := int(kp.Eproc.Ppid)
		if app := runningApp(int64(ppid)); app != 0 {
			if b := nsStringToGo(objc.Send[objc.ID](app, selBundleIdentifier)); b != "" {
				return nsStringToGo(objc.Send[objc.ID](app, selLocalizedNameApp)), b
			}
		}
		pid = ppid
	}
	return "", ""
}
