// devices.go - Audio control wrapper audio devices and manipulation functions.
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// This package provides a wrapper around OS audio APIs so the
// rest of the project does not have to deal with `ole` or `wca` specific directly.
// Every call will initialize OLE (CoInitializeEx) and uninitialize OLE (CoUninitialize)
// as goroutines are not bound to threads.
package audio

import (
	"errors"
	"strings"
)

type EndpointType int

const (
	Output      EndpointType = 0
	Input       EndpointType = 1
	Application EndpointType = 2
	// Window      EndpointType = 3
)

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
	return "unknown"
}

// find the endpoint type from a textual input
func ScanEndpointType(name string) (EndpointType, error) {
	var ret EndpointType = Output
	name = strings.ToLower(name)

	switch {
	case name[0:3] == "out":
		ret = Output
	case name[0:2] == "in":
		ret = Input
	case name[0:3] == "app":
		ret = Application
	// case name[0:2] == "win":
	// 	ret = Window
	default:
		return 0, errors.New("Unknown Endpoint type")
	}
	return ret, nil
}

// End.
