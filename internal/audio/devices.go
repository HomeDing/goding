// devices.go - Audio control wrapper audio devices and manipulation functions.
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// This package provides a wrapper around OS audio APIs so the
// rest of the project does not have to deal with Windows COM interfaces directly.
package audio

import (
	"errors"
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	win32 "github.com/deploymenttheory/go-bindings-win32/bindings/runtime/win32"
	"github.com/deploymenttheory/go-bindings-win32/bindings/win32/devices/functiondiscovery"
	"github.com/deploymenttheory/go-bindings-win32/bindings/win32/foundation"
	w32audio "github.com/deploymenttheory/go-bindings-win32/bindings/win32/media/audio"
	audioendpoints "github.com/deploymenttheory/go-bindings-win32/bindings/win32/media/audio/endpoints"
	win32com "github.com/deploymenttheory/go-bindings-win32/bindings/win32/system/com"
	"github.com/deploymenttheory/go-bindings-win32/bindings/win32/system/com/structuredstorage"
	"github.com/deploymenttheory/go-bindings-win32/bindings/win32/system/variant"
	"github.com/deploymenttheory/go-bindings-win32/bindings/win32/ui/shell/propertiessystem"
)

// The DeviceInfo struct saves public information about a corresponding device, extending the EndpointInfo.
type DeviceInfo struct {
	EndpointInfo
	ID string //  the internal device ID
}

func dataFlow(ept EndpointType) w32audio.EDataFlow {
	switch ept {
	case Output:
		return w32audio.ERender
	case Input:
		return w32audio.ECapture
	}
	return w32audio.EDataFlow(-1)
}

// validateEndpointType ensures the endpoint direction is a supported device type.
func validateEndpointType(ept EndpointType) error {
	if ept == Output || ept == Input {
		return nil
	}
	return errors.New("No Device Endpoint Type")
}

var mmDeviceEnumeratorCLSID = win32.GUID{Data1: 0xBCDE0395, Data2: 0xE52F, Data3: 0x467C, Data4: [8]byte{0x8E, 0x3D, 0xC4, 0x57, 0x92, 0x91, 0x69, 0x2E}}

func utf16PtrToString(ptr *uint16) string {
	if ptr == nil {
		return ""
	}

	length := 0
	for {
		ch := *(*uint16)(unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + uintptr(length*2)))
		if ch == 0 {
			break
		}
		length++
	}
	return syscall.UTF16ToString(unsafe.Slice(ptr, length))
}

func propVariantString(pv *structuredstorage.PROPVARIANT) (string, error) {
	if pv == nil {
		return "", errors.New("nil PROPVARIANT")
	}

	type rawPropVariant struct {
		vt  uint16
		_   [3]uint16
		ptr unsafe.Pointer
	}

	raw := (*rawPropVariant)(unsafe.Pointer(pv))
	vt := variant.VARENUM(raw.vt)

	switch vt {
	case variant.VT_LPWSTR:
		if raw.ptr == nil {
			return "", nil
		}
		return utf16PtrToString((*uint16)(raw.ptr)), nil
	case variant.VT_BSTR:
		if raw.ptr == nil {
			return "", nil
		}
		return utf16PtrToString((*uint16)(raw.ptr)), nil
	default:
		return "", fmt.Errorf("unsupported property variant type %v", vt)
	}
}

func deviceInfoFromIMMDevice(ept EndpointType, device *w32audio.IMMDevice, index int) (DeviceInfo, error) {
	var id foundation.PWSTR
	if err := device.GetId(&id); err != nil {
		return DeviceInfo{}, fmt.Errorf("get device %d ID failed: %w", index, err)
	}
	if id == nil {
		return DeviceInfo{}, fmt.Errorf("device %d has no ID", index)
	}
	deviceID := utf16PtrToString((*uint16)(unsafe.Pointer(id)))
	win32com.CoTaskMemFree(unsafe.Pointer(id))

	var store *propertiessystem.IPropertyStore
	if err := device.OpenPropertyStore(win32com.STGM_READ, &store); err != nil {
		return DeviceInfo{}, fmt.Errorf("open device %d property store failed: %w", index, err)
	}
	defer store.Release()

	var pv structuredstorage.PROPVARIANT
	if err := store.GetValue(&functiondiscovery.PKEY_Device_FriendlyName, &pv); err != nil {
		return DeviceInfo{}, fmt.Errorf("get device %d friendly name failed: %w", index, err)
	}
	defer structuredstorage.PropVariantClear(&pv)

	name, err := propVariantString(&pv)
	if err != nil {
		return DeviceInfo{}, fmt.Errorf("get device %d friendly name failed: %w", index, err)
	}

	return DeviceInfo{
		EndpointInfo: EndpointInfo{
			Flow: ept.String(),
			Name: name,
		},
		ID: deviceID,
	}, nil
}

// deviceEnumerator initializes COM and creates the Core Audio device enumerator.
func deviceEnumerator() (*w32audio.IMMDeviceEnumerator, func(), error) {
	comCleanup, err := comInit(uint32(win32com.COINIT_MULTITHREADED))
	if err != nil {
		return nil, nil, err
	}

	var unknown *win32.IUnknown
	if err := win32com.CoCreateInstance(
		&mmDeviceEnumeratorCLSID,
		nil,
		win32com.CLSCTX_ALL,
		&w32audio.IID_IMMDeviceEnumerator,
		&unknown,
	); err != nil {
		comCleanup()
		return nil, nil, fmt.Errorf("create MMDevice enumerator failed: %w", err)
	}
	if unknown == nil {
		comCleanup()
		return nil, nil, errors.New("create MMDevice enumerator returned nil")
	}

	enumerator := win32.Cast[w32audio.IMMDeviceEnumerator](unknown)
	cleanup := func() {
		enumerator.Release()
		comCleanup()
	}
	return enumerator, cleanup, nil
}

// defaultDevice returns the currently selected default audio device for the given flow.
func defaultDevice(ept EndpointType) (*w32audio.IMMDevice, func(), error) {
	if err := validateEndpointType(ept); err != nil {
		return nil, nil, err
	}

	deviceEnum, cleanup, err := deviceEnumerator()
	if err != nil {
		return nil, nil, err
	}

	var device *w32audio.IMMDevice
	if err := deviceEnum.GetDefaultAudioEndpoint(dataFlow(ept), w32audio.EConsole, &device); err != nil {
		cleanup()
		return nil, nil, err
	}

	deviceCleanup := func() {
		device.Release()
		cleanup()
	}
	return device, deviceCleanup, nil
}

// endpointVolume resolves the volume control interface for the current device.
func (d *DeviceInfo) endpointVolume() (*w32audio.IMMDevice, *audioendpoints.IAudioEndpointVolume, func(), error) {
	deviceEnum, cleanup, err := deviceEnumerator()
	if err != nil {
		return nil, nil, nil, err
	}

	var device *w32audio.IMMDevice
	if err := deviceEnum.GetDevice(d.ID, &device); err != nil {
		cleanup()
		return nil, nil, nil, err
	}

	var unknown *win32.IUnknown
	if err := device.Activate(&audioendpoints.IID_IAudioEndpointVolume, win32com.CLSCTX_ALL, nil, &unknown); err != nil {
		device.Release()
		cleanup()
		return nil, nil, nil, err
	}
	aev := win32.Cast[audioendpoints.IAudioEndpointVolume](unknown)

	deviceCleanup := func() {
		aev.Release()
		device.Release()
		cleanup()
	}
	return device, aev, deviceCleanup, nil
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

	enumerator, cleanup, err := deviceEnumerator()
	if err != nil {
		return nil, err
	}
	defer cleanup()

	var collection *w32audio.IMMDeviceCollection
	if err := enumerator.EnumAudioEndpoints(dataFlow(ept), w32audio.DEVICE_STATE_ACTIVE, &collection); err != nil {
		return nil, fmt.Errorf("EnumAudioEndpoints failed: %w", err)
	}
	defer collection.Release()

	var count uint32
	if err := collection.GetCount(&count); err != nil {
		return nil, fmt.Errorf("GetCount failed: %w", err)
	}

	result := make([]DeviceInfo, 0, count)
	for i := uint32(0); i < count; i++ {
		var device *w32audio.IMMDevice
		if err := collection.Item(i, &device); err != nil {
			return nil, fmt.Errorf("get device %d failed: %w", i, err)
		}
		deviceInfo, err := deviceInfoFromIMMDevice(ept, device, int(i))
		if err != nil {
			device.Release()
			return nil, err
		}
		result = append(result, deviceInfo)
		device.Release()
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
	return SetDefaultAudioDevice(d.ID)
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

	return aev.SetMasterVolumeLevelScalar(scalarVolume, nil)
} // SetVolume()

// RegisterVolumeChangeNotify requests notifications for master volume changes.
// This generated binding set does not currently expose the callback implementation.
func (d *DeviceInfo) RegisterVolumeChangeNotify() error {
	return errors.New("volume change notifications are not implemented")
} // RegisterVolumeChangeNotify()

// UnregisterVolumeChangeNotify removes the notification subscription.
func (d *DeviceInfo) UnregisterVolumeChangeNotify() error {
	return errors.New("volume change notifications are not implemented")
} // UnregisterVolumeChangeNotify()

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

	var muted foundation.BOOL
	if err := aev.GetMute(&muted); err != nil {
		return false, err
	}

	return muted != 0, nil
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

	return aev.SetMute(muted, nil)
} // SetMute()

// End.
