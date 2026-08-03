// volume.go - Volume element implementation
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// Package elements contains implementations of UI/control elements used by GoDing.
package volumeElement

import (
	"fmt"
	"log/slog"
	"maps"
	"strconv"

  "github.com/HomeDing/goding/internal/elements/baseElement"
	"github.com/HomeDing/goding/internal/elements/registry"
	"github.com/MixyLabs/go-wca/pkg/wca"
	"github.com/go-ole/go-ole"
)

type volumeElement struct {
	baseElement.BaseElement
	// baseElement.BaseElement
	// shadow int values to avoid too much string conversion and parsing
	minimum, maximum, value int
	processName             string
}

// creates a new VolumeElement instance with default configuration values and registers it in the registry.
func NewVolumeElement(elementId string) *volumeElement {
	v := volumeElement{
		BaseElement: baseElement.NewBaseElement("volume", elementId),
		processName: "main"}

	// set initial configuration parameters
	v.Config["min"] = "0"
	v.Config["max"] = "100"

	// set initial runtime value
	v.Values["value"] = "50"

	// set shadow variables to avoid too much conversions
	v.minimum = 0
	v.maximum = 100
	v.value = 50

	registry.Register(v)
	return &v
}

// Overwrite the New function to create a new VolumeElement instance with default configuration values.
func New(elementId string) *volumeElement {
	return NewVolumeElement(elementId)
}

// Set overrides the base element setter to validate volume-specific values and apply them to the system.
func (e volumeElement) Set(key, value string) bool {
	slog.Debug("volume.set", "element", e.GetKey(), "key", key, "value", value)
	var newValue int
	var err error

	if key == "value" || key == "min" || key == "max" {
		// check for good formatted value
		if newValue, err = strconv.Atoi(value); err != nil {
			return false
		}
	}

	changed := e.BaseElement.Set(key, value) // call the base Set method to handle known keys
	if changed {
		switch key {
		case "value":

			// constrain value to be within min and max and correct the Values map accordingly
			if newValue < e.minimum {
				newValue = e.minimum
				e.BaseElement.Values["value"] = strconv.Itoa(newValue)
			} else if newValue > e.maximum {
				newValue = e.maximum
				e.BaseElement.Values["value"] = strconv.Itoa(newValue)
			}

			e.value = newValue
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

func (e *volumeElement) Loop() bool {
	return false
}

func (e *volumeElement) XState() map[string]string {
	res := map[string]string{}
	maps.Copy(res, e.Config)
	res["name"] = e.processName
	return res
}

func (e *volumeElement) applyCurrentValue() error {
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

// register volume element type in the registry
func init() {
	NewVolumeElement("default")
}
