// devices.go - Audio control wrapper audio devices and manipulation functions.
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// This package provides a wrapper around OS audio APIs so the
// rest of the project does not have to deal with `ole` or `wca` specific directly.
// Every call will initialize OLE (CoInitializeEx) and uninitialize OLE (CoUninitialize)
// as goroutine are not bound to threads.
package audio

import (
	"strings"
)

type EndpointType int

const (
	Unknown     EndpointType = -1
	Output      EndpointType = 0
	Input       EndpointType = 1
	Application EndpointType = 2
	// Window      EndpointType = 3
)

// The EndpointInfo contains the common attributes and functions of devices and sessions.
type EndpointInfo struct {
	Name string // friendly name of device or application, used for matching
	Flow string // Endpoint Type "out", "in", "app"
}

func (ept EndpointType) String() string {
	switch ept {
	case Output:
		return "out"
	case Input:
		return "in"
	case Application:
		return "app"
		// case Window:
		// 	return "win"
	}
	return "no"
} // EndpointType.String()

// ParseEndpoint returns an endpoint identifier as EndpointType and name in lowercase
//
// Get endpoint type and clean identifier from a textual input
// Reformat the input for type and name combined string by removing extra characters and normalize
// Valid endpoint names:
//   - out:console, output:communication, output:multimedia
//   - out:con, out:com, out:mul
//   - out.con, out-com
//
// "out[put]:con[sole]" -- the current output device for Games, system notification sounds and voice commands
// "out[put]:com[munication]" -- the current output device for voice communications.
// "out[put]:mul[timedia]" -- the current output device for Music, movies, narration, and live music recording.
// "in[put]:..."
func ParseEndpoint(name string) (EndpointType, string) {
	var epType EndpointType = Unknown

	name = strings.ToLower(name)
	name = strings.Replace(name, ".", ":", 1)
	name = strings.Replace(name, "-", ":", 1)

	if !strings.Contains(name, (":")) {
		// use output as default
		name = "out:" + name
	}

	sType, sID, _ := strings.Cut(name, ":")

	// verify sType
	l := len(sType)
	switch {
	case l >= 2 && sType[0:2] == "in":
		epType = Input
	case l >= 3 && sType[0:3] == "out":
		epType = Output
	case l >= 3 && sType[0:3] == "app":
		epType = Application
		// case sType[0:2] == "win":
		// 	epType = Window
	}

	// verify sType
	if epType == Input || epType == Output {
		switch sID {
		case "main", "console":
			sID = "con"
		}
	}
	return epType, sID
} // ParseEndpoint()

// ===== default implementation for endpoints generating no errors =====

func (ep *EndpointInfo) GetVolume() (int, error) {
	return 0, nil
}

// default implementation for endpoint generating no errors
func (ep *EndpointInfo) SetVolume(vol int) error {
	return nil
}

func (d *EndpointInfo) GetMute() (bool, error) {
	return false, nil
} // GetMute

func (d *EndpointInfo) SetMute(muted bool) error {
	return nil
}

// End.
