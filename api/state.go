// state.go - API handler for server status
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// Package api provides HTTP handlers for the application's API endpoints.
package api

import (
	"io"
	"log"
	"net/http"
	"time"
)

func HandleStatus(w http.ResponseWriter, r *http.Request) {
	log.Print("handling status request")
	io.WriteString(w, "Server is running at ")
	io.WriteString(w, time.Now().Format(time.RFC3339))
	io.WriteString(w, "\n")
}
