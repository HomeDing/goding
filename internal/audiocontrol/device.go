// audiocontrol.go - Audio control wrapper for system volume
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// Package audiocontrol provides a small wrapper around OS audio APIs so the
// rest of the project does not need to import `ole` or `wca` directly.
package audiocontrol

import (
	"errors"
	"log/slog"
	"runtime"
	"unsafe"

	"github.com/MixyLabs/go-wca/pkg/wca"
	"github.com/go-ole/go-ole"
)

// The Device implements a struct representing the current output device
type Device struct {
	id    string
	name   string
	active bool

	volume int

	// Sessions []*Session
	oleInit bool
	mmde    *wca.IMMDeviceEnumerator
	mmd     *wca.IMMDevice
	aev     *wca.IAudioEndpointVolume
	// ps      *wca.IPropertyStore
}

var commonDevice *Device = nil

// Release releases all underlying COM and Windows resources held by the
// Device. It is safe to call Release multiple times.
// Clean up all OLE and Windows resources when Device is no longer needed.
func (d *Device) Release() {
	slog.Info("Device.Release")
	if d.aev != nil {
		d.aev.Release()
		d.aev = nil
	}
	if d.mmd != nil {
		d.mmd.Release()
		d.mmd = nil
	}
	if d.mmde != nil {
		d.mmde.Release()
		d.mmde = nil
	}
	if d.oleInit == true {
		ole.CoUninitialize()
		d.oleInit = false
	}
}

// to create the default Device:
// var dev *audiocontrol.Device
// if dev, err = audiocontrol.GetDefaultDevice(); err != nil {
// 	return
// }
// defer dev.Release()

// the Device is effectively not created when returning a error
// make sure old stuff is released finally even after beeing released explicitely.
// See: https://victoriametrics.com/blog/go-runtime-finalizer-keepalive/

// Create a default Device
// As a fallback a finalizer is registered on the Device object to call Release().
func GetDefaultDevice() (*Device, error) {
	var d Device = Device{}

	// register a finalizer so resources are cleaned up if the caller forgets
	runtime.SetFinalizer(&d, (*Device).Release)

	// initialize ole for this Device object
	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		ole.CoUninitialize() // even when an error has occurred
		return nil, err
	}
	d.oleInit = true

	// IMMDeviceEnumerator in d.mmde
	if err := wca.CoCreateInstance(wca.CLSID_MMDeviceEnumerator, 0, wca.CLSCTX_ALL, wca.IID_IMMDeviceEnumerator, &d.mmde); err != nil {
		d.Release()
		return nil, err
	}

	// IMMDevice in d.mmd
	if err := d.mmde.GetDefaultAudioEndpoint(wca.ERender, wca.EConsole, &d.mmd); err != nil {
		d.Release()
		return nil, err
	}

	// IAudioEndpointVolume in d.aev
	if err := d.mmd.Activate(wca.IID_IAudioEndpointVolume, wca.CLSCTX_ALL, nil, &d.aev); err != nil {
		d.Release()
		return nil, err
	}

	// Process ID in d.did
	d.mmd.GetId(&d.id)
	slog.Info("Device.GetDefaultDevice", "deviceID", d.id)

	// State as bool in d.active, https://learn.microsoft.com/en-us/windows/win32/api/mmdeviceapi/nf-mmdeviceapi-immdevice-getstate
	var stat uint32
	d.mmd.GetState(&stat)
	slog.Info("Device.GetDefaultDevice", "stat", stat)
	d.active = (stat == 1)

	var ps *wca.IPropertyStore
	if err := d.mmd.OpenPropertyStore(wca.STGM_READ, &ps); err != nil {
		d.Release()
		return nil, err
	}
	defer ps.Release()

	var pv wca.PROPVARIANT
	if err := ps.GetValue(&wca.PKEY_DeviceInterface_FriendlyName, &pv); err != nil {
		// if err := ps.GetValue(&wca.PKEY_Device_FriendlyName, &pv); err != nil {
		// if err := ps.GetValue(&wca.PKEY_Device_DeviceDesc, &pv); err != nil {
		d.Release()
		return nil, err
	}
	d.name = pv.String()
	slog.Info("Device.GetDefaultDevice", "return", d)

	var vol float32
	if err := d.aev.GetMasterVolumeLevelScalar(&vol); err != nil {
		d.Release()
		return nil, err
	}
	d.volume = int(vol * 100)

	// finally return Device object ready to be used.
	return &d, nil
} // GetDefaultDevice()

// GetMasterVolume returns the current master volume as an integer
// percentage in the range [0,100].
//
// It reads the scalar volume from the underlying IAudioEndpointVolume
// interface and converts it to a percentage. If the device's audio
// endpoint volume interface is not available an error is returned.
func (d *Device) GetMasterVolume() (int, error) {
	var vol float32

	if d.aev == nil {
		return 0, errors.New("Device.aev is missing")
	}

	if err := d.aev.GetMasterVolumeLevelScalar(&vol); err != nil {
		return 0, err
	}

	return int(vol * 100), nil
} // GetMasterVolume()

// SetMasterVolume sets the master volume using `vol` interpreted relative to
// the provided min/max range. `vol`, `min` and `max` are integer percentages.
//
// It uses vol as a scalar as percentage.
// If the device's audio endpoint volume interface is not available an error is returned.
func (d *Device) SetMasterVolume(vol, min, max int) error {
	if d.aev == nil {
		return errors.New("Device.aev is missing")
	}

	newScalarVolume := float32(vol-min) / float32(max-min)
	if newScalarVolume < 0 {
		newScalarVolume = 0
	} else if newScalarVolume > 1 {
		newScalarVolume = 1
	}

	return d.aev.SetMasterVolumeLevelScalar(newScalarVolume, nil)
} // SetMasterVolume()

// ListSessions enumerates audio sessions and emits informational logs similar
// to the previous `volume.list()` method. It returns an error on failure.
func ListSessions() error {
	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		return err
	}
	defer ole.CoUninitialize()

	var mmde *wca.IMMDeviceEnumerator
	if err := wca.CoCreateInstance(wca.CLSID_MMDeviceEnumerator, 0, wca.CLSCTX_ALL, wca.IID_IMMDeviceEnumerator, &mmde); err != nil {
		return err
	}
	defer mmde.Release()

	// mmde.EnumAudioEndpoints()

	var device2 *wca.IMMDevice
	if err := mmde.GetDefaultAudioEndpoint(wca.ERender, wca.EConsole, &device2); err != nil {
		return err
	}
	defer device2.Release()

	var deviceID string
	device2.GetId(&deviceID)
	slog.Info("audiocontrol.list", "deviceID", deviceID)

	var sessionManager *wca.IAudioSessionManager2
	if err := device2.Activate(wca.IID_IAudioSessionManager2, wca.CLSCTX_ALL, nil, &sessionManager); err != nil {
		return err
	}
	defer sessionManager.Release()

	var sessionEnum *wca.IAudioSessionEnumerator
	if err := sessionManager.GetSessionEnumerator(&sessionEnum); err != nil {
		return err
	}
	defer sessionEnum.Release()

	var sessionCount int
	if err := sessionEnum.GetCount(&sessionCount); err != nil {
		return err
	}

	slog.Info("audiocontrol.list", "sessions", sessionCount)

	for i := 0; i < sessionCount; i++ {
		var session *wca.IAudioSessionControl
		sessionEnum.GetSession(i, &session)

		var displayName string
		var procId uint32
		var sessionIdent string
		session.GetDisplayName(&displayName)

		dispatch, _ := session.QueryInterface(wca.IID_IAudioSessionControl2)
		session2 := (*wca.IAudioSessionControl2)(unsafe.Pointer(dispatch))

		session2.GetProcessId(&procId)
		session2.GetSessionIdentifier(&sessionIdent)
		var iconPath string
		session2.GetIconPath(&iconPath)

		slog.Info("audiocontrol.list",
			"displayName", displayName,
			"procId", procId,
			"sessionIdent", sessionIdent,
			"sysSound", session2.IsSystemSoundsSession(),
			"ico", iconPath)
	}

	return nil
}
