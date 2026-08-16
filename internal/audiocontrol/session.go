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
	"log/slog"
	"runtime"

	"github.com/MixyLabs/go-wca/pkg/wca"
)

// The Device implements a struct representing the current output device
type Session struct {
	d *Device
	// Sessions []*Session
	// oleInit bool
	// mmde    *wca.IMMDeviceEnumerator
	// mmd     *wca.IMMDevice
	// aev     *wca.IAudioEndpointVolume
	ps   *wca.IPropertyStore
	sm2  *wca.IAudioSessionManager2
	ase  *wca.IAudioSessionEnumerator
	asc2 *wca.IAudioSessionControl2
}

// Release releases all underlying COM and Windows resources held by the
// Session. It is safe to call Release multiple times.
// Clean up all OLE and Windows resources when Session is no longer needed.
func (s *Session) Release() {
	slog.Info("Session.Release")

	if s.sm2 != nil {
		s.sm2.Release()
	}
	if s.ps != nil {
		s.ps.Release()
	}
	if s.d != nil {
		s.d.Release()
	}

	// if s.aev != nil {
	// 	s.aev.Release()
	// }
	// if s.mmd != nil {
	// 	s.mmd.Release()
	// }
	// if s.mmde != nil {
	// 	s.mmde.Release()
	// }
	// if s.oleInit == true {
	// 	ole.CoUninitialize()
	// }
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
func GetSession(d *Device, find string) (*Session, error) {
	var s Session = Session{}
	s.d = d

	// register a finalizer so resources are cleaned up if the caller forgets
	runtime.SetFinalizer(&s, (*Session).Release)

	// PropertyStore in s.ps
	if err := d.mmd.OpenPropertyStore(wca.STGM_READ, &s.ps); err != nil {
		return nil, err
	}

	// IAudioSessionManager2 in s.sm2
	if err := d.mmd.Activate(wca.IID_IAudioSessionManager2, wca.CLSCTX_ALL, nil, &s.sm2); err != nil {
		s.Release()
		return nil, err
	}

	// var sessionEnum *wca.IAudioSessionEnumerator
	// if err := sessionManager.GetSessionEnumerator(&sessionEnum); err != nil {
	// 	return err
	// }
	// defer sessionEnum.Release()

	// IAudioSessionEnumerator in ase
	if err := s.sm2.GetSessionEnumerator(&s.ase); err != nil {
		s.Release()
		return nil, err
	}

	var sCount int
	s.ase.GetCount(&sCount)
	slog.Info("GetSession", "count", sCount)

	// finally return Session object ready to be used.
	return &s, nil
} // GetSession()

func (s *Session) GetProcessId() (uint32, error) {
	var (
		v   uint32
		err = s.asc2.GetProcessId(&v)
	)

	return v, err
}
