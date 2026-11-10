// console_windows.go - Windows console window detection
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// Package win provides Windows-specific utilities.
package win

import "syscall"

var getConsoleWindow = syscall.NewLazyDLL("kernel32.dll").NewProc("GetConsoleWindow")

// HasConsoleWindow reports whether the process is associated with a console window.
func HasConsoleWindow() bool {
	window, _, _ := getConsoleWindow.Call()
	return window != 0
}

// End.
