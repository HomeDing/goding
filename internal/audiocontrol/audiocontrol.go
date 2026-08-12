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
	"unsafe"

	"github.com/MixyLabs/go-wca/pkg/wca"
	"github.com/go-ole/go-ole"
)

// The Device implements a struct representing the current output device
type Device struct {
	// Sessions []*Session
	mmd *wca.IMMDevice
	aev *wca.IAudioEndpointVolume
	ps  *wca.IPropertyStore
}

func (d *Device) Release() {
	if d.ps != nil {
		d.ps.Release()
	}
	if d.aev != nil {
		d.aev.Release()
	}
	if d.mmd != nil {
		d.mmd.Release()
	}
	ole.CoUninitialize()
}

// create the default Device
// var dev *audiocontrol.Device
// if dev, err = audiocontrol.GetDefaultDevice(); err != nil {
// 	return
// }
// defer dev.Release()

// the Device is effectively not created when returning a error
func GetDefaultDevice() (*Device, error) {
	var ret *Device
	var mmde *wca.IMMDeviceEnumerator
	// var err error

	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		return nil, err
	}

	ret = &Device{}

	// IMMDeviceEnumerator only needed locally, not stored in Device
	if err := wca.CoCreateInstance(wca.CLSID_MMDeviceEnumerator, 0, wca.CLSCTX_ALL, wca.IID_IMMDeviceEnumerator, &mmde); err != nil {
		ret.Release()
		return nil, err
	}
	defer mmde.Release()

	if err := mmde.GetDefaultAudioEndpoint(wca.ERender, wca.EConsole, &ret.mmd); err != nil {
		ret.Release()
		return nil, err
	}

	if err := ret.mmd.Activate(wca.IID_IAudioEndpointVolume, wca.CLSCTX_ALL, nil, &ret.aev); err != nil {
		ret.Release()
		return nil, err
	}

	return ret, nil

	/*

		if err = device.mmd.OpenPropertyStore(wca.STGM_READ, &device.ps); err != nil {
			return nil, err
		}

		if err = device.mmd.Activate(wca.IID_IAudioEndpointVolume, wca.CLSCTX_ALL, nil, &device.aev); err != nil {
			return nil, err
		}

		err = device.mmd.Activate(wca.IID_IAudioSessionManager2, wca.CLSCTX_ALL, nil, &sm)
		if err != nil {
			return nil, err
		}

		defer sm.Release()

		device.Sessions, err = sessions(sm)
		if err != nil {
			return nil, err
		}

		return device, nil
	*/
} // GetDefaultDevice()

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
