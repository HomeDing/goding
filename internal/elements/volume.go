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
	"maps"
	"strconv"

	"github.com/HomeDing/goding/internal/audiocontrol"
	"github.com/HomeDing/goding/internal/elements/registry"
)

type Volume struct {
	Base
	// shadow int values to avoid too much string conversion and parsing
	minimum, maximum, value int
	processName             string
}

// creates a new VolumeElement instance with default configuration values and registers it in the registry.
func NewVolumeElement(elementId string) *Volume {
	v := &Volume{Base: NewBaseElement("volume", elementId)}

	// set initial configuration parameters
	v.Config["min"] = "0"
	v.Config["max"] = "100"

	// set initial runtime value
	v.Values["value"] = "50"

	v.minimum = 0
	v.maximum = 100
	v.value = 50
	v.processName = "main"

	registry.Register(v)
	return v
} // NewVolumeElement()

// Set overrides the base element setter to validate volume-specific values and apply them to the system.
// Do not fail when there is no real audio element found. The control nothing.
func (e *Volume) Set(key, value string) bool {
	slog.Debug("volume.set", "element", e.GetKey(), "key", key, "value", value)
	var dev *audiocontrol.Device
	var err error
	var newValue int

	if key == "value" || key == "min" || key == "max" {
		if newValue, err = strconv.Atoi(value); err != nil {
			return false
		}
	}

	// call the base Set method to handle known keys
	changed := e.Base.Set(key, value)
	if changed {
		switch key {
		case "value":

			// constrain the new value to the min/max range
			if newValue < e.minimum {
				newValue = e.minimum
			} else if newValue > e.maximum {
				newValue = e.maximum
			}

			dev, err = audiocontrol.GetDefaultDevice()

			if err == nil {
				defer dev.Release()
				if err = dev.SetMasterVolume(newValue, e.minimum, e.maximum); err != nil {
					slog.Error("volume.start", "err", err)
					return false
				}
			}
			e.value = newValue
			e.Values["value"] = strconv.Itoa(newValue)

		case "min":
			e.minimum = newValue
		case "max":
			e.maximum = newValue
		}
	}
	return changed
}

func (e *Volume) Loop() bool {
	return false
}

func (e *Volume) Start() {
	var dev *audiocontrol.Device
	var err error
	var vol int

	// check parameters to be useful
	if e.maximum < e.minimum {
		slog.Error("volume.start Bad range min...max", "min", e.minimum, "max", e.maximum)
		return
	}

	if dev, err = audiocontrol.GetDefaultDevice(); err != nil {
		return
	}
	defer dev.Release()

	if vol, err = dev.GetMasterVolume(); err != nil {
		slog.Error("volume.start", "err", err)
		return
	}

	slog.Debug("volume.start", "currentVolume", vol)
	e.value = vol
	e.Values["value"] = strconv.Itoa(vol)

	if err := audiocontrol.ListSessions(); err != nil {
		slog.Warn("volume.start.list", "error", err)
	}

}

func (e *Volume) XState() map[string]string {
	res := map[string]string{}
	maps.Copy(res, e.Config)
	res["name"] = e.processName
	return res
}

// End.
