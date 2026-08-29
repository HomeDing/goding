// sessions.go - Audio control wrapper audio devices and manipulation functions.
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
	// "errors"
	"errors"
	"log/slog"
	"unsafe"

	"github.com/MixyLabs/go-wca/pkg/wca"
	"github.com/go-ole/go-ole"

	"strings"
)

// The DeviceInfo struct saves public information about a corresponding device.
//   - PID -- the internal ID
//   - Name -- A friendly name, also used for specifying a device like "DELL U3223QE (Intel(R) Display-Audio)"
type SessionInfo struct {
	EndpointInfo
	PID         uint32
	SystemSound bool
	Active      bool
}

// ===== Public functions =====

// Create a List of all active audio sessions on the default device
func ListSessions() ([]SessionInfo, error) {
	var result []SessionInfo

	// Initialize COM
	if oleCleanup, err := oleInit(); err != nil {
		return nil, err
	} else {
		defer oleCleanup()
	}

	// Create enumerator
	var deviceEnum *wca.IMMDeviceEnumerator
	if err := wca.CoCreateInstance(wca.CLSID_MMDeviceEnumerator, 0, ole.CLSCTX_ALL, wca.IID_IMMDeviceEnumerator, &deviceEnum); err != nil {
		return nil, err
	}
	defer deviceEnum.Release()

	// Get default output device (Console role)
	var mmd *wca.IMMDevice
	if err := deviceEnum.GetDefaultAudioEndpoint(wca.ERender, wca.EConsole, &mmd); err != nil {
		return nil, err
	}
	defer mmd.Release()

	var sm2 *wca.IAudioSessionManager2
	if err := mmd.Activate(wca.IID_IAudioSessionManager2, wca.CLSCTX_ALL, nil, &sm2); err != nil {
		return nil, err
	}
	defer sm2.Release()

	// get its IAudioSessionEnumerator
	var sessionEnumerator *wca.IAudioSessionEnumerator
	if err := sm2.GetSessionEnumerator(&sessionEnumerator); err != nil {
		return nil, err
	}
	defer sessionEnumerator.Release()

	var sessionCount int
	if err := sessionEnumerator.GetCount(&sessionCount); err != nil {
		return nil, err
	}

	result = make([]SessionInfo, 0, sessionCount)

	// for each session:
	// https://github.com/omriharel/deej/blob/9c0307b96341538ec46f18d2e9b18aaa84441175/pkg/deej/session_finder_windows.go#L445
	for sessionIdx := 0; sessionIdx < sessionCount; sessionIdx++ {

		var audioSessionControl *wca.IAudioSessionControl
		if err := sessionEnumerator.GetSession(sessionIdx, &audioSessionControl); err != nil {
			slog.Info("GetSession Bad Index")
			return nil, err
		}

		// query its IAudioSessionControl2
		var audioSessionControl2 *wca.IAudioSessionControl2
		if dispatch, err := audioSessionControl.QueryInterface(wca.IID_IAudioSessionControl2); err != nil {
			slog.Error("Failed IAudioSessionControl2", "error", err, "sessionIdx", sessionIdx)
			audioSessionControl.Release()
			return nil, err

		} else {
			// receive a useful object instead of our dispatch
			audioSessionControl2 = (*wca.IAudioSessionControl2)(unsafe.Pointer(dispatch))
		}

		var si = SessionInfo{
			EndpointInfo: EndpointInfo{
				Flow: "app",
			},
		}

		si.SystemSound = audioSessionControl2.IsSystemSoundsSession() == nil
		audioSessionControl2.GetProcessId(&si.PID)
		audioSessionControl2.GetDisplayName(&si.Name)
		var sState uint32
		audioSessionControl2.GetState(&sState)
		si.Active = (sState == wca.AudioSessionStateActive)

		if si.SystemSound {
			si.Name = "System sounds"
		} else {
			si.Name, _ = FriendlyProcessName(si.PID)
		}

		result = append(result, si)

		// Release OLE stuff from inner loop
		// simpleAudioVolume.Release()
		audioSessionControl2.Release()
		audioSessionControl.Release()
	} // for

	return result, nil

} // ListSessions()

func FindSession(match string) (*SessionInfo, error) {
	match = strings.ToLower(match)

	sessions, _ := ListSessions()

	for _, session := range sessions {
		if strings.Contains(strings.ToLower(session.Name), match) {
			return &session, nil
		}
	} // for

	return nil, errors.New("Session not found")
} // FindSession()

func (s *SessionInfo) GetVolume() (int, error) {

	// Initialize COM
	if oleCleanup, err := oleInit(); err != nil {
		return 0, err
	} else {
		defer oleCleanup()
	}

	// Create enumerator
	var deviceEnum *wca.IMMDeviceEnumerator
	if err := wca.CoCreateInstance(wca.CLSID_MMDeviceEnumerator, 0, ole.CLSCTX_ALL, wca.IID_IMMDeviceEnumerator, &deviceEnum); err != nil {
		return 0, err
	}
	defer deviceEnum.Release()

	// Get default output device (Console role)
	var mmd *wca.IMMDevice
	if err := deviceEnum.GetDefaultAudioEndpoint(wca.ERender, wca.EConsole, &mmd); err != nil {
		return 0, err
	}
	defer mmd.Release()

	var sm2 *wca.IAudioSessionManager2
	if err := mmd.Activate(wca.IID_IAudioSessionManager2, wca.CLSCTX_ALL, nil, &sm2); err != nil {
		return 0, err
	}
	defer sm2.Release()

	// get its IAudioSessionEnumerator
	var sessionEnumerator *wca.IAudioSessionEnumerator
	if err := sm2.GetSessionEnumerator(&sessionEnumerator); err != nil {
		return 0, err
	}
	defer sessionEnumerator.Release()

	var sessionCount int
	if err := sessionEnumerator.GetCount(&sessionCount); err != nil {
		return 0, err
	}
	// slog.Debug("GetSession", slog.Int("sessionCount", sessionCount))

	for sessionIdx := 0; sessionIdx < sessionCount; sessionIdx++ {

		var audioSessionControl *wca.IAudioSessionControl
		if err := sessionEnumerator.GetSession(sessionIdx, &audioSessionControl); err != nil {
			slog.Info("GetSession Bad Index")
			return 0, err
		}

		// query its IAudioSessionControl2
		dispatch, err := audioSessionControl.QueryInterface(wca.IID_IAudioSessionControl2)
		if err != nil {
			slog.Error("Failed IAudioSessionControl2", "error", err, "sessionIdx", sessionIdx)
			return 0, err
		}

		// receive a useful object instead of our dispatch
		audioSessionControl2 := (*wca.IAudioSessionControl2)(unsafe.Pointer(dispatch))
		defer audioSessionControl2.Release()

		var pid uint32
		audioSessionControl2.GetProcessId(&pid)

		if pid == s.PID {
			var sav *wca.ISimpleAudioVolume

			if dispatch2, err := audioSessionControl2.QueryInterface(wca.IID_ISimpleAudioVolume); err != nil {
				return 0, err
			} else {
				sav = (*wca.ISimpleAudioVolume)(unsafe.Pointer(dispatch2))
			}

			var vol float32
			if err := sav.GetMasterVolume(&vol); err != nil {
				return 0, err
			}
			return int(vol * 100), nil
		}
	}
	return 0, errors.New("PID has no audio session")

}

func (s *SessionInfo) SetVolume(newVol int) error {

	// Initialize COM
	if oleCleanup, err := oleInit(); err != nil {
		return err
	} else {
		defer oleCleanup()
	}

	// Create enumerator
	var deviceEnum *wca.IMMDeviceEnumerator
	if err := wca.CoCreateInstance(wca.CLSID_MMDeviceEnumerator, 0, ole.CLSCTX_ALL, wca.IID_IMMDeviceEnumerator, &deviceEnum); err != nil {
		return err
	}
	defer deviceEnum.Release()

	// Get default output device (Console role)
	var mmd *wca.IMMDevice
	if err := deviceEnum.GetDefaultAudioEndpoint(wca.ERender, wca.EConsole, &mmd); err != nil {
		return err
	}
	defer mmd.Release()

	var sm2 *wca.IAudioSessionManager2
	if err := mmd.Activate(wca.IID_IAudioSessionManager2, wca.CLSCTX_ALL, nil, &sm2); err != nil {
		return err
	}
	defer sm2.Release()

	// get its IAudioSessionEnumerator
	var sessionEnumerator *wca.IAudioSessionEnumerator
	if err := sm2.GetSessionEnumerator(&sessionEnumerator); err != nil {
		return err
	}
	defer sessionEnumerator.Release()

	var sessionCount int
	if err := sessionEnumerator.GetCount(&sessionCount); err != nil {
		return err
	}
	// slog.Debug("GetSession", slog.Int("sessionCount", sessionCount))

	for sessionIdx := 0; sessionIdx < sessionCount; sessionIdx++ {

		var audioSessionControl *wca.IAudioSessionControl
		if err := sessionEnumerator.GetSession(sessionIdx, &audioSessionControl); err != nil {
			slog.Info("GetSession Bad Index")
			return err
		}

		// query its IAudioSessionControl2
		dispatch, err := audioSessionControl.QueryInterface(wca.IID_IAudioSessionControl2)
		if err != nil {
			slog.Error("Failed IAudioSessionControl2", "error", err, "sessionIdx", sessionIdx)
			return err
		}

		// receive a useful object instead of our dispatch
		audioSessionControl2 := (*wca.IAudioSessionControl2)(unsafe.Pointer(dispatch))
		defer audioSessionControl2.Release()

		var pid uint32
		audioSessionControl2.GetProcessId(&pid)

		if pid == s.PID {
			var sav *wca.ISimpleAudioVolume

			if dispatch2, err := audioSessionControl2.QueryInterface(wca.IID_ISimpleAudioVolume); err != nil {
				return err
			} else {
				sav = (*wca.ISimpleAudioVolume)(unsafe.Pointer(dispatch2))
			}

			// Set volume by scalar value
			scalarVolume := float32(newVol) / 100
			if scalarVolume < 0 {
				scalarVolume = 0
			} else if scalarVolume > 1 {
				scalarVolume = 1
			}
			return sav.SetMasterVolume(scalarVolume, nil)
		}
	}
	return errors.New("PID has no audio session")

} // SetVolume

// End.
