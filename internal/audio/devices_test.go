// devices_test.go - Tests for Windows audio device handling
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

package audio

import (
	"testing"

	"github.com/MixyLabs/go-wca/pkg/wca"
)

func TestPolicyConfigGUIDs(t *testing.T) {
	if got, want := wca.GUID_CPolicyConfigVistaClient.Data1, uint32(0x294935ce); got != want {
		t.Fatalf("unexpected PolicyConfig CLSID Data1: got 0x%x, want 0x%x", got, want)
	}

	if got, want := wca.GUID_IPolicyConfigVista.Data1, uint32(0x568b9108); got != want {
		t.Fatalf("unexpected PolicyConfig IID Data1: got 0x%x, want 0x%x", got, want)
	}
}

func TestListDevices(t *testing.T) {
	var devices []DeviceInfo
	var err error

	// ===== output devices =====
	devices, err = ListDevices(wca.ERender)
	if err != nil {
		t.Skipf("Windows audio device enumeration unavailable: %v", err)
	}

	for index, device := range devices {
		if device.ID == "" {
			t.Errorf("device %d has an empty ID", index)
		}
		if device.Name == "" {
			t.Errorf("device %d has an empty friendly name", index)
		}
		if device.Flow != "out" {
			t.Errorf("device %d has wrong flow direction", index)
		}
	}

	// ===== input devices =====
	devices, err = ListDevices(wca.ECapture)
	if err != nil {
		t.Skipf("Windows audio device enumeration unavailable: %v", err)
	}

	for index, device := range devices {
		if device.ID == "" {
			t.Errorf("device %d has an empty ID", index)
		}
		if device.Name == "" {
			t.Errorf("device %d has an empty friendly name", index)
		}
		if device.Flow != "in" {
			t.Errorf("device %d has wrong flow direction", index)
		}
	}
}

func TestSetDefaultDevice(t *testing.T) {
	var defaultDevice *DeviceInfo
	var err error

	defaultDevice, err = GetDefaultDevice(Output)
	if err != nil {
		t.Skipf("GetDefaultDevice failed: %v", err)
	}

	devices, err := ListDevices(wca.ERender)
	if err != nil {
		t.Skipf("Windows audio device enumeration unavailable: %v", err)
	}

	for _, device := range devices {
		device.SetDefault()
	}

	if d, err := FindDevice(Output, "Dell"); err == nil {
		d.SetDefault()
	}

	defaultDevice.SetDefault()
}

func TestFindDevice(t *testing.T) {
	// var device *DeviceInfo
	var err error

	if _, err = FindDevice(Output, "Dell"); err != nil {
		t.Log("'Dell' audio device could not be found.", err)
	}

	if _, err = FindDevice(Output, "IBM"); err != nil {
		t.Log("'IBM' audio device could not be found.", err)
	}

	if _, err = FindDevice(Output, "RealTek"); err != nil {
		t.Log("'RealTek' audio device could not be found.", err)
	}

}

func TestVolume(t *testing.T) {
	var device *DeviceInfo
	var err error

	device, err = GetDefaultDevice(Output)
	if err != nil {
		t.Skipf("GetDefaultDevice failed: %v", err)
	}

	vol, _ := device.GetVolume()

	device.SetVolume(vol + 2)

	device.SetVolume(vol)

	device, err = FindDevice(Output, "AudioBox")
	if err != nil {
		t.Skipf("FindDevice failed: %v", err)
	}

	vol, _ = device.GetVolume()

	device.SetVolume(vol + 2)

	device.SetVolume(vol)
}

func TestMute(t *testing.T) {
	var defaultDevice *DeviceInfo
	var err error

	defaultDevice, err = GetDefaultDevice(Output)
	if err != nil {
		t.Skipf("GetDefaultDevice failed: %v", err)
	}

	muted, _ := defaultDevice.GetMute()

	defaultDevice.SetMute(!muted)
	defaultDevice.SetMute(muted)
}

// End.
