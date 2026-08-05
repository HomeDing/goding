// volume.go - Volume element implementation
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// Package elements contains implementations of UI/control elements used by GoDing.
package elements

import (
	"fmt"
	"log/slog"
	"maps"
	"strconv"

	"github.com/HomeDing/goding/internal/elements/registry"
	"github.com/MixyLabs/go-wca/pkg/wca"
	"github.com/go-ole/go-ole"
)

type Volume struct {
	Base
	// shadow int values to avoid too much string conversion and parsing
	minimum, maximum, value int
	processName             string
}

// creates a new VolumeElement instance with default configuration values and registers it in the registry.
func NewVolumeElement(elementId string) *Volume {
	v := &Volume{
		Base: NewBaseElement("volume", elementId)}

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
func (e *Volume) Set(key, value string) bool {
	slog.Debug("volume.set", "element", e.GetKey(), "key", key, "value", value)
	var newValue int
	var err error

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

			e.value = newValue
			e.Values["value"] = strconv.Itoa(newValue)
			if err := e.applyCurrentValue(); err != nil {
				slog.Warn("could not apply volume change", "element", e.GetKey(), "error", err)
			}
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

func (e *Volume) XState() map[string]string {
	res := map[string]string{}
	maps.Copy(res, e.Config)
	res["name"] = e.processName
	return res
}

func (e *Volume) applyCurrentValue() error {
	valueText := e.Values["value"]
	if valueText == "" {
		return fmt.Errorf("volume value is not set")
	}

	if e.maximum < e.minimum {
		return fmt.Errorf("invalid volume range %v..%v", e.minimum, e.maximum)
	}

	scalar := float32(e.value-e.minimum) / float32(e.maximum-e.minimum)
	if scalar < 0 {
		scalar = 0
	} else if scalar > 1 {
		scalar = 1
	}

	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		return err
	}
	defer ole.CoUninitialize()

	var mmde *wca.IMMDeviceEnumerator
	if err := wca.CoCreateInstance(wca.CLSID_MMDeviceEnumerator, 0, wca.CLSCTX_ALL, wca.IID_IMMDeviceEnumerator, &mmde); err != nil {
		return err
	}
	defer mmde.Release()

	var mmd *wca.IMMDevice
	if err := mmde.GetDefaultAudioEndpoint(wca.ERender, wca.EConsole, &mmd); err != nil {
		return err
	}
	defer mmd.Release()

	var aev *wca.IAudioEndpointVolume
	if err := mmd.Activate(wca.IID_IAudioEndpointVolume, wca.CLSCTX_ALL, nil, &aev); err != nil {
		return err
	}
	defer aev.Release()

	return aev.SetMasterVolumeLevelScalar(scalar, nil)
} // applyCurrentValue
