// midicat.go - Helper API for midi used by web server
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// Package serve provides HTTP handlers and helpers for serving midi related APIs.
package serve

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/HomeDing/goding/internal/audio"
)

// HandleListDevices returns an HTTP GET handler that return all current midi devices as a JSON result.
func HandleListDevices() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("http.HandleListDevices")

		// Validate method, Expect GET and HEAD requests only for static files
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get Input Devices
		deviceList1, err := audio.ListDevices(audio.Input)

		// Get Output Devices
		deviceList2, err := audio.ListDevices(audio.Output)

		deviceList := append(deviceList1, deviceList2...)

		if err == nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(deviceList)
			return
		}

	}
} // HandleListDevices()

// HandleListSessions returns an HTTP GET handler that return all current audio sessions on the default device.
func HandleListSessions() http.HandlerFunc {

	// Get all sessions and return as JSON
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("http.HandleListSession")

		// Validate method, Expect GET and HEAD requests only for static files
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		sessionList, err := audio.ListSessions()

		if err == nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(sessionList)
			return
		}

	}
} // HandleListSessions()
