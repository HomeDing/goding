// sessions.go - Audio control wrapper audio devices and manipulation functions.
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// This package provides a wrapper around OS audio APIs without exposing COM interfaces.
package audio

import (
	"errors"
	"log/slog"
	"strings"

	win32 "github.com/deploymenttheory/go-bindings-win32/bindings/runtime/win32"
	w32audio "github.com/deploymenttheory/go-bindings-win32/bindings/win32/media/audio"
	win32com "github.com/deploymenttheory/go-bindings-win32/bindings/win32/system/com"
)

// SessionInfo describes an audio session and its owning process.
type SessionInfo struct {
	EndpointInfo
	PID         uint32
	SystemSound bool
	Active      bool
}

func sessionEnumerator() (*w32audio.IAudioSessionEnumerator, func(), error) {
	device, deviceCleanup, err := defaultDevice(Output)
	if err != nil {
		return nil, nil, err
	}

	var unknown *win32.IUnknown
	if err := device.Activate(&w32audio.IID_IAudioSessionManager2, win32com.CLSCTX_ALL, nil, &unknown); err != nil {
		deviceCleanup()
		return nil, nil, err
	}
	if unknown == nil {
		deviceCleanup()
		return nil, nil, errors.New("activate audio session manager returned nil")
	}
	manager := win32.Cast[w32audio.IAudioSessionManager2](unknown)

	sessions, err := manager.GetSessionEnumerator()
	if err != nil {
		manager.Release()
		deviceCleanup()
		return nil, nil, err
	}
	if sessions == nil {
		manager.Release()
		deviceCleanup()
		return nil, nil, errors.New("get audio session enumerator returned nil")
	}

	cleanup := func() {
		sessions.Release()
		manager.Release()
		deviceCleanup()
	}
	return sessions, cleanup, nil
}

func sessionInfoAt(sessions *w32audio.IAudioSessionEnumerator, index int32) (SessionInfo, error) {
	var control *w32audio.IAudioSessionControl
	if err := sessions.GetSession(index, &control); err != nil {
		return SessionInfo{}, err
	}
	if control == nil {
		return SessionInfo{}, errors.New("get audio session returned nil")
	}
	defer control.Release()

	control2, err := win32.QueryInterface[w32audio.IAudioSessionControl2](control, &w32audio.IID_IAudioSessionControl2)
	if err != nil {
		return SessionInfo{}, err
	}
	defer control2.Release()

	info := SessionInfo{EndpointInfo: EndpointInfo{Flow: "app"}}
	info.SystemSound = control2.IsSystemSoundsSession() == nil
	if err := control2.GetProcessId(&info.PID); err != nil {
		return SessionInfo{}, err
	}
	var state w32audio.AudioSessionState
	if err := control2.GetState(&state); err != nil {
		return SessionInfo{}, err
	}
	info.Active = state == w32audio.AudioSessionStateActive

	if info.SystemSound {
		info.Name = "System sounds"
	} else {
		info.Name, _ = FriendlyProcessName(info.PID)
	}
	return info, nil
}

func simpleVolumeForPID(sessions *w32audio.IAudioSessionEnumerator, pid uint32) (*w32audio.ISimpleAudioVolume, error) {
	var count int32
	if err := sessions.GetCount(&count); err != nil {
		return nil, err
	}

	for index := int32(0); index < count; index++ {
		var control *w32audio.IAudioSessionControl
		if err := sessions.GetSession(index, &control); err != nil {
			return nil, err
		}
		if control == nil {
			return nil, errors.New("get audio session returned nil")
		}
		control2, err := win32.QueryInterface[w32audio.IAudioSessionControl2](control, &w32audio.IID_IAudioSessionControl2)
		if err != nil {
			control.Release()
			return nil, err
		}

		var sessionPID uint32
		err = control2.GetProcessId(&sessionPID)
		if err != nil {
			control2.Release()
			control.Release()
			return nil, err
		}
		if sessionPID == pid {
			volume, queryErr := win32.QueryInterface[w32audio.ISimpleAudioVolume](control2, &w32audio.IID_ISimpleAudioVolume)
			control2.Release()
			control.Release()
			return volume, queryErr
		}

		control2.Release()
		control.Release()
	}
	return nil, errors.New("PID has no audio session")
}

func matchSessions(match string) ([]SessionInfo, error) {
	match = strings.ToLower(match)
	sessions, cleanup, err := sessionEnumerator()
	if err != nil {
		return nil, err
	}
	defer cleanup()

	var count int32
	if err := sessions.GetCount(&count); err != nil {
		return nil, err
	}

	result := make([]SessionInfo, 0, int(count))
	for index := int32(0); index < count; index++ {
		info, err := sessionInfoAt(sessions, index)
		if err != nil {
			slog.Info("GetSession Bad Index", "error", err, "sessionIdx", index)
			return nil, err
		}
		if len(info.Name) > 0 && (match == "" || strings.Contains(strings.ToLower(info.Name), match)) {
			result = append(result, info)
		}
	}
	return result, nil
}

// ListSessions returns audio sessions on the default output device.
func ListSessions() ([]SessionInfo, error) {
	return matchSessions("")
}

func FindSession(match string) *SessionInfo {
	sessions, _ := matchSessions(match)
	for _, session := range sessions {
		return &session
	}
	return nil
}

func (s *SessionInfo) GetVolume() (int, error) {
	sessions, cleanup, err := sessionEnumerator()
	if err != nil {
		return 0, err
	}
	defer cleanup()

	volume, err := simpleVolumeForPID(sessions, s.PID)
	if err != nil {
		return 0, err
	}
	defer volume.Release()

	var level float32
	if err := volume.GetMasterVolume(&level); err != nil {
		return 0, err
	}
	return int(level * 100), nil
}

func (s *SessionInfo) SetVolume(newVol int) error {
	sessions, cleanup, err := sessionEnumerator()
	if err != nil {
		return err
	}
	defer cleanup()

	volume, err := simpleVolumeForPID(sessions, s.PID)
	if err != nil {
		return err
	}
	defer volume.Release()

	scalarVolume := float32(newVol) / 100
	if scalarVolume < 0 {
		scalarVolume = 0
	} else if scalarVolume > 1 {
		scalarVolume = 1
	}
	return volume.SetMasterVolume(scalarVolume, nil)
}

// End.
