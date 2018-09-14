package main

/*
#cgo pkg-config: ibus-1.0
#cgo CFLAGS: -std=c99
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <X11/Xatom.h>
#include <X11/keysym.h> //xproto-devel
#include <ibus.h> //ibus-devel

inline void ucharfree(unsigned char* uc) {
	XFree(uc);
}

inline void windowfree(Window* w) {
	XFree(w);
}

inline char* uchar2char(unsigned char* uc, unsigned long len) {
	for (int i=0; i<len; i++) {
		if (uc[i] == 0) {
			uc[i] = '\n';
		}
	}
	return (char*)uc;
}

inline unsigned long uchar2long(unsigned char* uc) {
	return *(unsigned long*)(uc);
}

static int ignore_x_error(Display *display, XErrorEvent *error) {
    return 0;
}

void setXIgnoreErrorHandler() {
	XSetErrorHandler(ignore_x_error);
}

*/
import "C"
import "strings"

const (
	MaxPropertyLen = 128

	WM_CLASS = "WM_CLASS"
)

type CDisplay C.Display

func init() {
	C.setXIgnoreErrorHandler()
}

func x11GetUCharProperty(display *C.Display, window C.Window, propName string) (*C.uchar, C.ulong) {
	var actualType C.Atom
	var actualFormat C.int
	var nItems, bytesAfter C.ulong
	var prop *C.uchar

	filterAtom := C.XInternAtom(display, C.CString(propName), C.True)

	status := C.XGetWindowProperty(display, window, filterAtom, 0, MaxPropertyLen, C.False, C.AnyPropertyType, &actualType, &actualFormat, &nItems, &bytesAfter, &prop)

	if status == C.Success {
		return prop, nItems
	}

	return nil, 0
}

func x11GetStringProperty(display *C.Display, window C.Window, propName string) string {
	prop, propLen := x11GetUCharProperty(display, window, propName)
	if prop != nil {
		defer C.ucharfree(prop)
		return C.GoString(C.uchar2char(prop, propLen))
	}

	return ""
}

func x11OpenDisplay() *CDisplay {
	return (*CDisplay)(C.XOpenDisplay(nil))
}

func x11GetInputFocus(display *CDisplay) C.Window {
	var window C.Window
	var revertTo C.int
	C.XGetInputFocus((*C.Display)(display), &window, &revertTo)

	return window
}

func x11GetParentWindow(display *CDisplay, w C.Window) (rootWindow, parentWindow C.Window) {
	var childrenWindows *C.Window
	var nChild C.uint
	C.XQueryTree((*C.Display)(display), w, &rootWindow, &parentWindow, &childrenWindows, &nChild)
	C.windowfree(childrenWindows)

	return
}

func x11CloseDisplay(d *CDisplay) {
	C.XCloseDisplay((*C.Display)(d))
}

func x11GetFocusWindowClass(display *CDisplay) []string {

	w := x11GetInputFocus(display)
	strClass := ""
	for {
		s := x11GetStringProperty((*C.Display)(display), w, WM_CLASS)

		rootWindow, parentWindow := x11GetParentWindow(display, w)

		if rootWindow == parentWindow {
			break
		}

		if len(s) > 0 {
			strClass += s + "\n"
		}

		w = parentWindow
	}

	return strings.Split(strClass, "\n")
}

func x11KeyvalToKeyCode(display *CDisplay, keyval uint32) uint32 {
	return uint32(C.XKeysymToKeycode((*C.Display)(display), (C.KeySym)(keyval)))
}

func ibusUnicodeToKeyval(r rune) uint32  {
	return uint32(C.ibus_unicode_to_keyval((C.gunichar)(r)))
}