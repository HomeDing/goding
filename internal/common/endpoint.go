// devices.go - Audio control wrapper audio devices and manipulation functions.
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

package common

// Element defines the common methods that runtime element instances must implement.
type Endpoint interface {
	// GetName() ()
	GetVolume() (int, error)
	SetVolume(vol int) error
}

// End.
