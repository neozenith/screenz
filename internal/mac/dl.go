//go:build darwin

// Package mac binds the macOS Accessibility, CoreGraphics, ColorSync and
// AppKit primitives into a CGO_ENABLED=0 binary via purego (ADR1.1). It is
// the only package that talks to the window server; everything above it is a
// pure transform covered by `make check`, while this package is exercised by
// `make itest` on a real GUI session (ADR1.3).
package mac

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/ebitengine/purego"
)

// AppKit reads (NSScreen, NSRunningApplication, NSWorkspace) must happen on
// the main OS thread; package init runs on the main goroutine before main.
func init() { runtime.LockOSThread() }

const (
	rtldNow    = 0x2
	rtldGlobal = 0x8
)

var (
	loadOnce sync.Once
	missing  []string
)

// register wraps purego.RegisterLibFunc, which panics on a missing symbol.
// A missing symbol is recorded instead so `screenz doctor` can name it
// rather than crashing (G1 requires a loud, named failure).
func register(fptr any, lib uintptr, name string) {
	defer func() {
		if r := recover(); r != nil {
			missing = append(missing, name)
		}
	}()
	if lib == 0 {
		missing = append(missing, name)
		return
	}
	purego.RegisterLibFunc(fptr, lib, name)
}

func dlopen(path string) uintptr {
	h, err := purego.Dlopen(path, rtldNow|rtldGlobal)
	if err != nil {
		missing = append(missing, fmt.Sprintf("dlopen %s: %v", path, err))
		return 0
	}
	return h
}

// Load dlopens every framework once and registers the symbol tables.
// It never panics; Missing() reports anything that failed to bind.
func Load() {
	loadOnce.Do(func() {
		libCF := dlopen("/System/Library/Frameworks/CoreFoundation.framework/CoreFoundation")
		libCG := dlopen("/System/Library/Frameworks/CoreGraphics.framework/CoreGraphics")
		libAS := dlopen("/System/Library/Frameworks/ApplicationServices.framework/ApplicationServices")
		libCS := dlopen("/System/Library/Frameworks/ColorSync.framework/ColorSync")
		// AppKit must be loaded or the NSScreen / NSRunningApplication /
		// NSWorkspace classes do not exist for objc lookups.
		dlopen("/System/Library/Frameworks/AppKit.framework/AppKit")

		register(&cfRelease, libCF, "CFRelease")
		register(&cfRetain, libCF, "CFRetain")
		register(&cfGetTypeID, libCF, "CFGetTypeID")
		register(&cfBooleanGetTypeID, libCF, "CFBooleanGetTypeID")
		register(&cfStringCreateWithCString, libCF, "CFStringCreateWithCString")
		register(&cfStringGetCString, libCF, "CFStringGetCString")
		register(&cfStringGetLength, libCF, "CFStringGetLength")
		register(&cfArrayGetCount, libCF, "CFArrayGetCount")
		register(&cfArrayGetValueAtIndex, libCF, "CFArrayGetValueAtIndex")
		register(&cfDictionaryGetValue, libCF, "CFDictionaryGetValue")
		register(&cfNumberGetValue, libCF, "CFNumberGetValue")
		register(&cfBooleanGetValue, libCF, "CFBooleanGetValue")
		register(&cfUUIDCreateString, libCF, "CFUUIDCreateString")

		register(&cgMainDisplayID, libCG, "CGMainDisplayID")
		register(&cgGetActiveDisplayList, libCG, "CGGetActiveDisplayList")
		register(&cgDisplayBounds, libCG, "CGDisplayBounds")
		register(&cgDisplayIsBuiltin, libCG, "CGDisplayIsBuiltin")
		register(&cgDisplayIsMain, libCG, "CGDisplayIsMain")
		register(&cgDisplaySerialNumber, libCG, "CGDisplaySerialNumber")
		register(&cgDisplayVendorNumber, libCG, "CGDisplayVendorNumber")
		register(&cgDisplayModelNumber, libCG, "CGDisplayModelNumber")
		register(&cgDisplayCopyDisplayMode, libCG, "CGDisplayCopyDisplayMode")
		register(&cgDisplayModeGetPixelWidth, libCG, "CGDisplayModeGetPixelWidth")
		register(&cgDisplayModeGetPixelHeight, libCG, "CGDisplayModeGetPixelHeight")
		register(&cgWindowListCopyWindowInfo, libCG, "CGWindowListCopyWindowInfo")
		register(&cgRectMakeWithDictionaryRepresentation, libCG, "CGRectMakeWithDictionaryRepresentation")
		// CGDisplayCreateUUIDFromDisplayID lives in ColorSync.framework;
		// dlsym against CoreGraphics fails (verified in the spike).
		register(&cgDisplayCreateUUIDFromDisplayID, libCS, "CGDisplayCreateUUIDFromDisplayID")

		register(&axIsProcessTrusted, libAS, "AXIsProcessTrusted")
		register(&axIsProcessTrustedWithOptions, libAS, "AXIsProcessTrustedWithOptions")
		register(&axUIElementCreateApplication, libAS, "AXUIElementCreateApplication")
		register(&axUIElementCopyAttributeValue, libAS, "AXUIElementCopyAttributeValue")
		register(&axUIElementSetAttributeValue, libAS, "AXUIElementSetAttributeValue")
		register(&axUIElementSetMessagingTimeout, libAS, "AXUIElementSetMessagingTimeout")
		register(&axValueGetValue, libAS, "AXValueGetValue")
		register(&axValueCreate, libAS, "AXValueCreate")
	})
}

// Missing lists every framework or symbol that failed to bind; empty on a
// healthy machine. Doctor prints it, other commands refuse to run on it.
func Missing() []string {
	Load()
	return missing
}
