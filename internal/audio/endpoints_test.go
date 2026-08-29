// devices_test.go - Tests for Windows audio device handling
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

package audio

import (
	"log/slog"
	"testing"

	"github.com/HomeDing/goding/internal/common"

	"github.com/MixyLabs/go-wca/pkg/wca"
)

func TestListEndpoints(t *testing.T) {

	// endpoints can hold devices and sessions
	var endpoints []common.Endpoint

	// ===== output devices =====
	devices, _ := ListDevices(wca.ERender)
	for i := range devices {
		endpoints = append(endpoints, &devices[i])
	}

	sessions, _ := ListSessions()
	for i := range sessions {
		endpoints = append(endpoints, &sessions[i])
	}

	for _, ep := range endpoints {
		vol, _ := ep.GetVolume()
		slog.Info("TestListEndpoints", "ep", ep, "volume", vol)
	}
}

func TestEndpointVolume(t *testing.T) {

}

func TestEndpointMute(t *testing.T) {

}

// End.
