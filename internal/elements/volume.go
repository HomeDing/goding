// volume.go - Volume element implementation
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// Package elements contains implementations of UI/control elements used by GoDing.
package elements

import (
	"log/slog"
	"strconv"

	"github.com/HomeDing/goding/internal/audio"
	"github.com/HomeDing/goding/internal/common"
	"github.com/HomeDing/goding/internal/global"
)

// The Volume element exists for controlling the volume of audio endpoints
// like speakers or applications
type Volume struct {
	Base
	// shadow int values to avoid too much string conversion and parsing
	minimum, maximum, value int
	et                      audio.EndpointType
	epName                  string
	device                  *audio.DeviceInfo
	session                 *audio.SessionInfo
}

// init registers the Volume element type with the element factory.
// init() is called by the go runtime before main() and before any other package-level variables are initialized.
func init() {
	RegisterFactory("volume", NewVolumeElement)
} // init()

// NewVolumeElement creates a Volume element with its default configuration values.
//
// The endpoint to be controlled is defined by the "endpoint" parameter (default out:con)
// as a string with colon separated elements (no space) like:
// [out|in]... : [con|com|mul]... or
// [app]...: `application` name
// [win]... : [top]
// "out[put]:con[sole]" -- the current output device for Games, system notification sounds and voice commands
// "out[put]:com[munication]" -- the current output device for voice communications.
// "out[put]:mul[timedia]" -- the current output device for Music, movies, narration, and live music recording.
// "in[put]:..."
func NewVolumeElement(elementId string) common.Element {
	v := &Volume{Base: *newBaseElement("volume", elementId)}

	// set initial / default configuration parameters
	v.Config["min"] = "0"
	v.Config["max"] = "100"
	v.Config["endpoint"] = "out:con"

	// set initial runtime value
	v.Values["value"] = "50"

	v.minimum = 0
	v.maximum = 100
	v.value = 50

	return v
} // NewVolumeElement()

// ===== private functions =====

// Get the deviceInfo when controlling a input or output device
func (e *Volume) getDeviceInfo() *audio.DeviceInfo {
	var ret *audio.DeviceInfo
	var err error

	switch e.et {
	case audio.Input, audio.Output:
		if e.epName == "con" {
			ret, err = audio.GetDefaultDevice(e.et)
		} else {
			ret, err = audio.FindDevice(e.et, e.epName)
		}
		if err == nil {
			e.device = ret
			return ret
		}
	}
	return nil
} // getDeviceInfo()

// Get the current volume, fail without error returning 0
func (e *Volume) getVolume() int {
	slog.Debug("volume.getVolume")

	di := e.getDeviceInfo()
	if di != nil {
		vol, _ := di.GetVolume()
		return vol
	}
	return 0
} // getVolume()

func (e *Volume) display(name string, value int) {
	global.Display.SetHeading(name)
	global.Display.SetProgress(int32(value))
	global.Display.SetMessage(strconv.Itoa(value) + "%")
}

// Set the volume, fail without error
func (e *Volume) setVolume(value int) {
	slog.Debug("volume.setVolume", "value", value)

	switch e.et {
	case audio.Output, audio.Input:
		di := e.getDeviceInfo()
		if di != nil {
			di.SetVolume(value)
			e.display(di.Name, value)
		}
	case audio.Application:
		e.session = audio.FindSession(e.epName)
		if e.session != nil {
			e.session.SetVolume(value)
			e.display(e.session.Name, value)
		}
	}
} // setVolume()

// Set overrides the base element setter to validate volume-specific values and apply them to the system.
// Do not fail when there is no real audio element found. The control nothing.
func (e *Volume) Set(key, value string) bool {
	slog.Debug("volume.set", "element", e.GetKey(), "key", key, "value", value)
	var err error
	var newValue int

	if key == "value" || key == "min" || key == "max" {
		if newValue, err = strconv.Atoi(value); err != nil {
			return false
		}
	}

	// call the base Set method to handle only set requests for known keys.
	if changed := e.Base.Set(key, value); changed {

		switch key {
		case "value", "volume":
			// constrain the new value to the min/max range
			if newValue < e.minimum {
				newValue = e.minimum
			} else if newValue > e.maximum {
				newValue = e.maximum
			}

			if e.isActive {
				e.setVolume(newValue)
			}
			e.value = newValue
			e.Values["value"] = strconv.Itoa(newValue)

		case "mute":

			// ===== parameters

		case "endpoint":
			e.et, e.epName = audio.ParseEndpoint(value)
			e.Config["endpoint"] = e.et.String() + ":" + e.epName
			if e.et == audio.Unknown {
				return false
			}

		case "min":
			e.minimum = newValue
		case "max":
			e.maximum = newValue
		}
		return true
	}
	return false
}

func (e *Volume) Loop() bool {
	return false
}

func (e *Volume) Start() {
	// check parameters to be useful
	if e.maximum < e.minimum {
		slog.Error("volume.start Bad range min...max", "min", e.minimum, "max", e.maximum)
		return
	}

	e.Base.Start()

	e.value = e.getVolume()
	e.Values["value"] = strconv.Itoa(e.value)
	slog.Debug("volume.start", "currentVolume", e.value)
}

// End.
