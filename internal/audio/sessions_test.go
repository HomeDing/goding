// devices_test.go - Tests for Windows audio device handling
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

package audio

import (
	"testing"
)

func TestListSessions(t *testing.T) {
	var sessions []SessionInfo
	var err error

	sessions, err = ListSessions()
	if err != nil {
		t.Skipf("Windows audio device enumeration unavailable: %v", err)
	}

	for index, session := range sessions {
		if session.Name == "" {
			t.Errorf("device %d has an empty friendly name", index)
		}
	}

}

func TestFindSession(t *testing.T) {
	var err error

	if _, err = FindSession("System sounds"); err != nil {
		t.Log("'System sounds' audio session could not be found.", err)
	}

	if _, err = FindSession("Edge"); err != nil {
		t.Log("'Edge' audio session could not be found.", err)
	}

	if _, err = FindSession("TrDo"); err != nil {
		t.Log("'TrDo' audio device could not be found.", err)
	}

}

func TestSessionVolume(t *testing.T) {

	if s, err := FindSession("Edge"); err != nil {
		t.Log("'Edge' audio session could not be found.", err)
	} else {
		vol, _ := s.GetVolume()
		t.Log("'Edge' volume", vol)

		s.SetVolume(vol + 10)
		s.SetVolume(vol)
	}

}

// End.
