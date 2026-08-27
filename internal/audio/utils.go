package audio

import (
	"fmt"
	"strings"

	"github.com/go-ole/go-ole"

	"errors"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/StackExchange/wmi"
	"golang.org/x/sys/windows"
)

const S_FALSE = 1

// initialize ole but also accept that there was a initialization up from (e.g. from midi)
// Code=S_FALSE requires CoUninitialize to be called.
func oleInit() (func(), error) {
	err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)

	if err != nil {
		if hr, ok := err.(*ole.OleError); !ok || hr.Code() != S_FALSE {
			return nil, fmt.Errorf("CoInitializeEx failed: 0x%x", hr)
		}
	}
	return func() { ole.CoUninitialize() }, nil
}

func CleanEndpointName(name string) string {
	name = strings.Replace(name, ".", ":", 1)
	name = strings.Replace(name, "-", ":", 1)
	name = strings.ToLower(name)

	return name
} // CleanEndpointName()

// Step 1: Get full process path from PID
func processPathFromPID(pid uint32) (string, error) {
	const PROCESS_QUERY_LIMITED_INFORMATION = 0x1000

	h, err := windows.OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return "", err
	}
	defer windows.CloseHandle(h)

	var buf [windows.MAX_PATH]uint16
	size := uint32(len(buf))

	err = windows.QueryFullProcessImageName(h, 0, &buf[0], &size)
	if err != nil {
		return "", err
	}

	return syscall.UTF16ToString(buf[:size]), nil
}

// Step 2: Extract FileDescription from version resource
func fileDescriptionFromPath(path string) (string, error) {
	var handle windows.Handle
	size, err := windows.GetFileVersionInfoSize(path, &handle)
	if err != nil {
		return "", err
	}

	buf := make([]byte, size)
	err = windows.GetFileVersionInfo(path, 0, size, unsafe.Pointer(&buf[0]))
	if err != nil {
		return "", err
	}

	var block unsafe.Pointer
	var blockLen uint32

	// 040904b0 = US English + Unicode
	err = windows.VerQueryValue(
		unsafe.Pointer(&buf[0]),
		`\StringFileInfo\040904b0\FileDescription`,
		unsafe.Pointer(&block),
		&blockLen,
	)
	if err != nil {
		return "", err
	}

	s := syscall.UTF16ToString((*[1 << 20]uint16)(block)[:blockLen])
	if s == "" {
		return "", errors.New("empty description")
	}
	return s, nil
}

// Step 3: WMI fallback (Win32_Process.Description)
type Win32_Process struct {
	ProcessId   uint32
	Name        string
	Description string
}

func wmiDescriptionFromPID(pid uint32) (string, error) {
	var procs []Win32_Process
	q := fmt.Sprintf("WHERE ProcessId = %d", pid)

	err := wmi.Query("SELECT ProcessId, Name, Description FROM Win32_Process "+q, &procs)
	if err != nil {
		return "", err
	}
	if len(procs) == 0 {
		return "", errors.New("not found")
	}

	if procs[0].Description != "" {
		return procs[0].Description, nil
	}
	return "", errors.New("empty")
}

// Step 4: Combined helper
func FriendlyProcessName(pid uint32) (string, error) {
	// 1) Get full path
	path, err := processPathFromPID(pid)
	if err != nil {
		return "", err
	}

	// 2) Try FileDescription
	if desc, err := fileDescriptionFromPath(path); err == nil {
		return desc, nil
	}

	// 3) Try WMI Description
	if desc, err := wmiDescriptionFromPID(pid); err == nil {
		return desc, nil
	}

	// 4) Fallback: exe name
	base := filepath.Base(path)
	return strings.TrimSuffix(base, ".exe"), nil
}

// End.
