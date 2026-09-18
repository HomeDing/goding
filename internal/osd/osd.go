//go:build windows

// osd.go - On-screen display window implementation for Windows
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// Package osd provides a lightweight Windows OSD overlay window for showing text and a progress bar.
package osd

import (
	"syscall"
	"unsafe"

	"github.com/deploymenttheory/go-bindings-win32/bindings/runtime/win32"
	"github.com/deploymenttheory/go-bindings-win32/bindings/win32/foundation"
	"github.com/deploymenttheory/go-bindings-win32/bindings/win32/graphics/dwm"
	"github.com/deploymenttheory/go-bindings-win32/bindings/win32/graphics/gdi"
	"github.com/deploymenttheory/go-bindings-win32/bindings/win32/system/libraryloader"
	wm "github.com/deploymenttheory/go-bindings-win32/bindings/win32/ui/windowsandmessaging"
)

// some constants for the OSD window
const (
	windowClass  = "GoOSDWindow"
	windowFont   = "Segoe UI"
	windowWidth  = 284
	windowHeight = 72
	windowMargin = 12
	windowGap    = 8
	timerID      = uintptr(1)
)

// OSD represents an overlay window for transient information.
type OSD struct {
	hwnd        foundation.HWND
	instance    foundation.HINSTANCE
	proc        uintptr
	isOpen      bool
	enforceQuit bool

	// UI settings and pre-loaded objects
	uiTitleRect   foundation.RECT
	uiMessageRect foundation.RECT

	headingFont gdi.HFONT
	headingText string

	textFont gdi.HFONT
	text     string
	progress int32

	color foundation.COLORREF

	workingarea foundation.RECT

	timeoutSecs   int
	activeTimerID uintptr
}

// ===== OSD Package functions =====

// Create a new OSD window.
func New() (*OSD, error) {
	instance, err := libraryloader.GetModuleHandle(nil)
	if err != nil {
		return nil, err
	}

	// OSD to be returned
	this := &OSD{
		instance:    foundation.HINSTANCE(instance),
		headingText: "<heading>",
		text:        "<text>",
	}
	if err := this.defineLayout(); err != nil {
		return nil, err
	}

	this.proc = syscall.NewCallback(this.windowProc)

	className := win32.UTF16Ptr(windowClass)
	class := wm.WNDCLASSW{
		LpfnWndProc:   wm.WNDPROC(this.proc),
		HInstance:     this.instance,
		LpszClassName: className,
	}
	if _, err := wm.RegisterClass(&class); err != nil {
		return nil, err
	}

	classNameString := windowClass
	this.hwnd, err = wm.CreateWindowEx(
		wm.WS_EX_TOPMOST|wm.WS_EX_LAYERED|wm.WS_EX_TRANSPARENT|wm.WS_EX_NOACTIVATE,
		&classNameString,
		nil,
		wm.WS_POPUP,
		this.workingarea.Right-windowWidth-windowMargin,
		this.workingarea.Bottom-windowHeight-windowMargin,
		windowWidth,
		windowHeight,
		0,
		0,
		this.instance,
		nil,
	)
	if err != nil {
		return nil, err
	}

	if err := wm.SetLayeredWindowAttributes(this.hwnd, 0, 235, wm.LWA_ALPHA); err != nil {
		return nil, err
	}

	_ = dwm.DwmSetWindowAttribute(
		this.hwnd,
		uint32(dwm.DWMWA_WINDOW_CORNER_PREFERENCE),
		[]byte{byte(dwm.DWMWCP_ROUNDSMALL), 0, 0, 0},
	)

	this.isOpen = false
	return this, nil
} // New()

// ===== Private OSD functions =====

// defineLayout loads the desktop metrics used to position the OSD.
func (this *OSD) defineLayout() error {

	// this.uiMargin = windowMargin
	const lhTitle = 20
	const pxTitle = 14
	const lhMessage = 20
	const pxMessage = 14

	fontFace := windowFont

	this.uiTitleRect = foundation.RECT{Left: windowMargin, Top: windowMargin, Right: windowWidth - windowMargin, Bottom: windowMargin + lhTitle}
	this.uiMessageRect = foundation.RECT{Left: this.uiTitleRect.Left, Top: this.uiTitleRect.Bottom + windowGap, Right: this.uiTitleRect.Right}
	this.uiMessageRect.Bottom = this.uiMessageRect.Top + lhMessage

	this.headingFont = gdi.CreateFont(-pxTitle,
		0, 0, 0, int32(gdi.FW_DEMIBOLD),
		0, 0, 0,
		uint32(gdi.DEFAULT_CHARSET),
		uint32(gdi.OUT_DEFAULT_PRECIS),
		uint32(gdi.CLIP_DEFAULT_PRECIS),
		uint32(gdi.DEFAULT_QUALITY),
		uint32(gdi.DEFAULT_PITCH)|uint32(gdi.FF_DONTCARE),
		&fontFace,
	)

	this.textFont = gdi.CreateFont(-pxMessage,
		0, 0, 0, int32(gdi.FW_NORMAL),
		0, 0, 0,
		uint32(gdi.DEFAULT_CHARSET),
		uint32(gdi.OUT_DEFAULT_PRECIS),
		uint32(gdi.CLIP_DEFAULT_PRECIS),
		uint32(gdi.DEFAULT_QUALITY),
		uint32(gdi.DEFAULT_PITCH)|uint32(gdi.FF_DONTCARE),
		&fontFace,
	)

	this.color = foundation.COLORREF(gdi.GetSysColor(gdi.COLOR_WINDOWTEXT))
	wm.SystemParametersInfo(wm.SPI_GETWORKAREA, 0, unsafe.Pointer(&this.workingarea), 0)

	return nil
} // defineLayout()

// startTimer starts or restarts the timer for hiding the osd window.
func (this *OSD) startTimer() {
	if this.hwnd != 0 {
		if this.activeTimerID != 0 {
			wm.KillTimer(this.hwnd, this.activeTimerID)
			this.activeTimerID = 0
		}

		if this.timeoutSecs > 0 {
			this.activeTimerID, _ = wm.SetTimer(this.hwnd, timerID, uint32(this.timeoutSecs)*1000, 0)
		}
	}
} // startTimer()

// stopTimer disables the hide timer.
func (this *OSD) stopTimer() {
	if this.hwnd != 0 && this.activeTimerID != 0 {
		wm.KillTimer(this.hwnd, this.activeTimerID)
		this.activeTimerID = 0
	}
} // stopTimer()

func (this *OSD) invalidate() {
	if this.hwnd != 0 {
		gdi.InvalidateRect(this.hwnd, nil, true)
	}
} // invalidate()

// ===== Public OSD functions =====

// SetHeading updates the heading and shows the window.
func (this *OSD) SetHeading(title string) *OSD {
	this.headingText = title
	if !this.isOpen {
		wm.ShowWindow(this.hwnd, wm.SW_SHOW)
	}
	this.invalidate()
	this.startTimer()

	return this
} // SetHeading()

// SetMessage updates the message and shows the window.
func (this *OSD) SetMessage(text string) *OSD {
	this.text = text
	if !this.isOpen {
		wm.ShowWindow(this.hwnd, wm.SW_SHOW)
	}
	this.invalidate()
	this.startTimer()

	return this
} // SetMessage()

// SetProgress updates the progress value and shows the window.
func (this *OSD) SetProgress(n int32) *OSD {
	this.progress = n
	if !this.isOpen {
		wm.ShowWindow(this.hwnd, wm.SW_SHOW)
	}
	this.invalidate()
	this.startTimer()

	return this
} // SetProgress()

// SetTimeout sets how long the OSD stays visible.
func (this *OSD) SetTimeout(secs int) *OSD {
	this.timeoutSecs = secs
	this.startTimer()
	return this
} // SetTimeout()

// Close hides the OSD window.
func (this *OSD) Close() error {
	var err error

	if this.isOpen {
		err = wm.PostMessage(this.hwnd, wm.WM_CLOSE, 0, 0)
	}

	return err
} // Close()

// Destroy stops the OSD and optionally quits the app message loop.
func (this *OSD) Destroy(withQuit bool) {
	if this.hwnd != 0 {
		this.enforceQuit = withQuit
		_ = wm.PostMessage(this.hwnd, wm.WM_DESTROY, 0, 0)
	}
} // Destroy()

// windowProc handles Windows messages for the OSD display window.
func (this *OSD) windowProc(hwnd foundation.HWND, message uint32, wParam foundation.WPARAM, lParam foundation.LPARAM) uintptr {
	switch message {
	case wm.WM_PAINT:
		var paint gdi.PAINTSTRUCT
		dc := gdi.BeginPaint(hwnd, &paint)
		var clientRect foundation.RECT

		wm.GetClientRect(hwnd, &clientRect)
		backgroundBrush := gdi.GetSysColorBrush(gdi.COLOR_WINDOW)
		gdi.FillRect(dc, &clientRect, backgroundBrush)
		gdi.SetBkMode(dc, int32(gdi.TRANSPARENT))
		gdi.SetTextColor(dc, this.color)

		gdi.SelectObject(dc, gdi.HGDIOBJ(this.headingFont))
		gdi.DrawTextEx(dc,
			win32.UTF16Ptr(this.headingText), int32(len(this.headingText)),
			&this.uiTitleRect,
			gdi.DT_TOP+gdi.DT_LEFT, nil)

		if this.progress >= 0 {
			trackBrush := gdi.CreateSolidBrush(foundation.COLORREF(0x00D0eeD0))
			gdi.FillRect(dc, &this.uiMessageRect, trackBrush)
			gdi.DeleteObject(gdi.HGDIOBJ(trackBrush))

			progressRect := this.uiMessageRect
			progressRect.Right = progressRect.Left + ((progressRect.Right-progressRect.Left)*this.progress)/100
			completedBrush := gdi.CreateSolidBrush(foundation.COLORREF(0x00D0d0D0))
			gdi.FillRect(dc, &progressRect, completedBrush)
			gdi.DeleteObject(gdi.HGDIOBJ(completedBrush))
		}

		gdi.SelectObject(dc, gdi.HGDIOBJ(this.textFont))
		gdi.DrawTextEx(dc,
			win32.UTF16Ptr(this.text), int32(len(this.text)),
			&this.uiMessageRect,
			gdi.DT_TOP+gdi.DT_LEFT, nil)

		gdi.EndPaint(hwnd, &paint)
		return 0

	case wm.WM_TIMER:
		if uintptr(wParam) == this.activeTimerID {
			this.stopTimer()
			wm.ShowWindow(hwnd, wm.SW_HIDE)
		}
		return 0

	case wm.WM_DESTROY:
		wm.CloseWindow(this.hwnd)
		this.isOpen = false
		this.hwnd = 0
		if this.enforceQuit {
			wm.PostQuitMessage(0)
		}
		return 0

	default:
		return uintptr(wm.DefWindowProc(hwnd, message, wParam, lParam))
	}
} // windowProc()

// End.
