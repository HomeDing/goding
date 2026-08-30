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
	"fmt"
	"strings"

	"github.com/MixyLabs/go-wca/pkg/wca"
	"github.com/go-ole/go-ole"
)

// The DeviceInfo struct saves public information about a corresponding device, extending the EndpointInfo.
type DeviceInfo struct {
	EndpointInfo
	ID string //  the internal device ID
}

// get wca flow direction for devices from EndpointType
func wcaFlow(ept EndpointType) uint32 {
	switch ept {
	case Output:
		return wca.ERender
	case Input:
		return wca.ECapture
	}
	return 99
} // wcaFlow()

// validateEndpointType ensures the endpoint direction is a supported device type.
func validateEndpointType(ept EndpointType) error {
	if ept == Output || ept == Input {
		return nil
	}
	return errors.New("No Device Endpoint Type")
}

// deviceEnumerator initializes COM and creates the Core Audio device enumerator.
func deviceEnumerator() (*wca.IMMDeviceEnumerator, func(), error) {
	if oleCleanup, err := oleInit(); err != nil {
		return nil, nil, err
	} else {
		deviceEnum := new(wca.IMMDeviceEnumerator)
		if err := wca.CoCreateInstance(
			wca.CLSID_MMDeviceEnumerator,
			0,
			wca.CLSCTX_ALL,
			wca.IID_IMMDeviceEnumerator,
			&deviceEnum,
		); err != nil {
			oleCleanup()
			return nil, nil, err
		}

		cleanup := func() {
			deviceEnum.Release()
			oleCleanup()
		}
		return deviceEnum, cleanup, nil
	}
}

// defaultDevice returns the currently selected default audio device for the given flow.
func defaultDevice(ept EndpointType) (*wca.IMMDevice, func(), error) {
	if err := validateEndpointType(ept); err != nil {
		return nil, nil, err
	}

	deviceEnum, cleanup, err := deviceEnumerator()
	if err != nil {
		return nil, nil, err
	}

	mmd := new(wca.IMMDevice)
	if err := deviceEnum.GetDefaultAudioEndpoint(wcaFlow(ept), wca.EConsole, &mmd); err != nil {
		cleanup()
		return nil, nil, err
	}

	deviceCleanup := func() {
		mmd.Release()
		cleanup()
	}
	return mmd, deviceCleanup, nil
}

// deviceInfoFromIMMDevice converts a Core Audio device into the package DeviceInfo type.
func deviceInfoFromIMMDevice(ept EndpointType, device *wca.IMMDevice, index int) (DeviceInfo, error) {
	var id string
	if err := device.GetId(&id); err != nil {
		return DeviceInfo{}, fmt.Errorf("get device %d ID failed: %w", index, err)
	}

	var store *wca.IPropertyStore
	if err := device.OpenPropertyStore(wca.STGM_READ, &store); err != nil {
		return DeviceInfo{}, fmt.Errorf("open device %d property store failed: %w", index, err)
	}
	defer store.Release()

	var pv wca.PROPVARIANT
	if err := store.GetValue(&wca.PKEY_Device_FriendlyName, &pv); err != nil {
		return DeviceInfo{}, fmt.Errorf("get device %d friendly name failed: %w", index, err)
	}

	return DeviceInfo{
		EndpointInfo: EndpointInfo{
			Flow: ept.String(),
			Name: pv.String(),
		},
		ID: id,
	}, nil
}

// endpointVolume resolves the volume control interface for the current device.
func (d *DeviceInfo) endpointVolume() (*wca.IMMDevice, *wca.IAudioEndpointVolume, func(), error) {
	deviceEnum, cleanup, err := deviceEnumerator()
	if err != nil {
		return nil, nil, nil, err
	}

	var mmd *wca.IMMDevice
	if err := deviceEnum.GetDevice(d.ID, &mmd); err != nil {
		cleanup()
		return nil, nil, nil, err
	}

	var aev *wca.IAudioEndpointVolume
	if err := mmd.Activate(wca.IID_IAudioEndpointVolume, wca.CLSCTX_ALL, nil, &aev); err != nil {
		mmd.Release()
		cleanup()
		return nil, nil, nil, err
	}

	deviceCleanup := func() {
		aev.Release()
		mmd.Release()
		cleanup()
	}
	return mmd, aev, deviceCleanup, nil
}

// GetDefaultDevice retrieves the active audio endpoint devices for the specified
// data flow direction (render or capture) using the Windows Core Audio MMDevice API.
//
// The DeviceInfo result corresponds to an IMMDevice instance representing a single audio endpoint
// device in the system.
//
// GetDefaultDevice retrieves:
//   - The unique endpoint ID string
//   - The flow direction of the device
//   - The friendly device name (PKEY_Device_FriendlyName)
//
// Parameters:
//
//	ept — The audio device type. Typical values are:
//	       Output  (output devices)
//	       Input (capture devices)
//
// Returns:
//
//	A DeviceInfo struct describing the active default device, or an error
//	if function fails.
//
// This function initializes COM internally and ensures proper cleanup.
func GetDefaultDevice(ept EndpointType) (*DeviceInfo, error) {
	if err := validateEndpointType(ept); err != nil {
		return nil, err
	}

	mmd, cleanup, err := defaultDevice(ept)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	deviceInfo, err := deviceInfoFromIMMDevice(ept, mmd, 0)
	if err != nil {
		return nil, err
	}
	return &deviceInfo, nil
} // GetDefaultDevice()

// ListDevices enumerates all active audio endpoint devices for the specified
// data flow direction (render or capture) using the Windows Core Audio MMDevice API.
//
// Each item in this collection
// corresponds to an IMMDevice instance representing a single audio endpoint
// device in the system.
//
// For each device, ListDevices retrieves:
//   - The unique endpoint ID string
//   - The flow direction of the device
//   - The friendly device name (PKEY_Device_FriendlyName)
//
// Parameters:
//
//	ept — The audio device type. Typical values are:
//	       Output  (output devices)
//	       Input (capture devices)
//
// Returns:
//
//	A slice of DeviceInfo structs describing each active device, or an error
//	if enumeration fails.
//
// This function initializes COM internally and ensures proper cleanup.
func ListDevices(ept EndpointType) ([]DeviceInfo, error) {
	if err := validateEndpointType(ept); err != nil {
		return nil, err
	}

	deviceEnum, cleanup, err := deviceEnumerator()
	if err != nil {
		return nil, err
	}
	defer cleanup()

	var coll *wca.IMMDeviceCollection
	if err := deviceEnum.EnumAudioEndpoints(wcaFlow(ept), wca.DEVICE_STATE_ACTIVE, &coll); err != nil {
		return nil, fmt.Errorf("EnumAudioEndpoints failed: %w", err)
	}
	defer coll.Release()

	var count uint32
	if err := coll.GetCount(&count); err != nil {
		return nil, fmt.Errorf("GetCount failed: %w", err)
	}

	result := make([]DeviceInfo, 0, count)
	for i := 0; i < int(count); i++ {
		var dev *wca.IMMDevice
		if err := coll.Item(uint32(i), &dev); err != nil {
			return nil, fmt.Errorf("get device %d failed: %w", i, err)
		}
		deviceInfo, err := deviceInfoFromIMMDevice(ept, dev, i)
		if err != nil {
			return nil, err
		}
		result = append(result, deviceInfo)
		dev.Release()
	}

	return result, nil
}

// find first matching device with the substring in the name
// compare without case sensitivity by converting to lowercase first.
func FindDevice(ept EndpointType, match string) (*DeviceInfo, error) {
	if err := validateEndpointType(ept); err != nil {
		return nil, err
	}

	match = strings.ToLower(match)
	devices, err := ListDevices(ept)
	if err != nil {
		return nil, err
	}

	for i := range devices {
		if devices[i].ID != "" && devices[i].Name != "" {
			if strings.Contains(strings.ToLower(devices[i].Name), match) {
				return &devices[i], nil
			}
		}
	} // for

	return nil, errors.New("Device not found")
} // FindDevice()

// Using IPolicyConfig to change device, partly undocumented !
func (d *DeviceInfo) SetDefault() error {
	if oleCleanup, err := oleInit(); err != nil {
		return err
	} else {
		defer oleCleanup()
	}

	var pc *wca.IPolicyConfigVista
	if err := wca.CoCreateInstance(
		wca.GUID_CPolicyConfigVistaClient,
		0,
		ole.CLSCTX_ALL,
		wca.GUID_IPolicyConfigVista,
		&pc,
	); err != nil {
		return errors.New("CoCreateInstance failed")
	}
	defer pc.Release()

	roles := []wca.ERole{wca.EConsole}
	for _, r := range roles {
		if err := pc.SetDefaultEndpoint(d.ID, r); err != nil {
			return fmt.Errorf("SetDefaultEndpoint failed: %w", err)
		}
	}

	return nil
}

// get the volume from a device or return 0 if not found
func (d *DeviceInfo) GetVolume() (int, error) {
	_, aev, cleanup, err := d.endpointVolume()
	if err != nil {
		return 0, err
	}
	defer cleanup()

	var vol float32
	if err := aev.GetMasterVolumeLevelScalar(&vol); err != nil {
		return 0, err
	}
	return int(vol * 100), nil
} // GetVolume

func (d *DeviceInfo) SetVolume(vol int) error {
	_, aev, cleanup, err := d.endpointVolume()
	if err != nil {
		return err
	}
	defer cleanup()

	scalarVolume := float32(vol) / 100
	if scalarVolume < 0 {
		scalarVolume = 0
	} else if scalarVolume > 1 {
		scalarVolume = 1
	}

	aev.SetMasterVolumeLevelScalar(scalarVolume, nil)
	return nil
} // SetVolume()

// GetMute queries the mute state of the audio endpoint device represented by
// this DeviceInfo instance.
//
// Behavior:
//   - Returns true if the device is currently muted.
//   - Returns false if the device is not muted.
//   - Returns an error if the device cannot be accessed or the mute state
//     cannot be retrieved.
//
// This method performs COM initialization internally and ensures proper
// cleanup. It requires that the DeviceInfo contains a valid endpoint ID
// corresponding to an active audio device.
func (d *DeviceInfo) GetMute() (bool, error) {
	_, aev, cleanup, err := d.endpointVolume()
	if err != nil {
		return false, err
	}
	defer cleanup()

	var muted bool
	if err := aev.GetMute(&muted); err != nil {
		return false, err
	}

	return muted, nil
} // GetMute

// SetMute changes the mute state of the audio endpoint device represented by
// this DeviceInfo instance.
//
// Behavior:
//   - Passing true mutes the device.
//   - Passing false un-mutes the device.
//   - Returns an error if the device cannot be accessed or the mute state
//     cannot be changed.
//
// This method performs COM initialization internally and ensures proper
// cleanup. It requires that the DeviceInfo contains a valid endpoint ID
// corresponding to an active audio device.
func (d *DeviceInfo) SetMute(muted bool) error {
	_, aev, cleanup, err := d.endpointVolume()
	if err != nil {
		return err
	}
	defer cleanup()

	aev.SetMute(muted, nil)
	return nil
} // SetMute()

// End.
