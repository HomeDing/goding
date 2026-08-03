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

type VolumeElement struct {
	BaseElement
	name           string
	endpointVolume *wca.IAudioEndpointVolume
}

// creates a new VolumeElement instance with default configuration values and registers it in the registry.
func NewVolumeElement(elementId string) *VolumeElement {
	e := NewElement("volume", elementId)
	v := &VolumeElement{Element: e, name: "Volume"}

	v.Config["min"] = "0"
	v.Config["max"] = "100"
	v.Values["value"] = "50"
	registry.Register(v.Element)
	return v
}

// Overwrite the New function to create a new VolumeElement instance with default configuration values.
func New(elementId string) *VolumeElement {
	return NewVolumeElement(elementId)
}

// overwrite the Element.Set method to handle volume-specific configuration changes and apply them to the system.
func (e *VolumeElement) Set(key, value string) bool {
	slog.Debug("volume.set", "element", e.GetKey(), "key", key, "value", value)
	slog.Debug("volume.set", slog.String("element", e.GetKey()), "key", key, "value", value)

	ret := e.Element.Set(key, value) // call the base Set method to handle known keys

	if !ret {
		return false // no change was made, so return false
	}

	if key == "value" {
		if volPercentage, err := strconv.ParseUint(value, 10, 16); err != nil {
			// constrain value to be within min and max
			min, _ := strconv.ParseUint(e.Config["min"], 10, 16)
			max, _ := strconv.ParseUint(e.Config["max"], 10, 16)
			if volPercentage < min {
				volPercentage = min
			}
			if volPercentage > max {
				volPercentage = max
			}
			slog.Debug("volume.change...")

			//todo: apply volume change to system
		}
	}

	return true
}

func (e *VolumeElement) Loop() bool {
	return false
}

func (e *VolumeElement) State() map[string]string {
	res := map[string]string{}
	maps.Copy(res, e.Config)
	res["name"] = e.name
	return res
}

func (e *VolumeElement) applyCurrentValue() error {
	value, err := strconv.ParseFloat(e.Config["value"], 64)
	if err != nil {
		return err
	}
	min, err := strconv.ParseFloat(e.Config["min"], 64)
	if err != nil {
		return err
	}
	max, err := strconv.ParseFloat(e.Config["max"], 64)
	if err != nil {
		return err
	}
	if max <= min {
		return fmt.Errorf("invalid volume range %v..%v", min, max)
	}

	scalar := (value - min) / (max - min)
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

	return aev.SetMasterVolumeLevelScalar(float32(scalar), nil)
}

// register volume element type in the registry
func init() {
	NewVolumeElement("default")
}
