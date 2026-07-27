// verbose.go - HTTP request logging middleware
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// Package verbose provides middleware for structured HTTP request logging.
package verbose

// verbose logging middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// HttpLogging prints method, path, and request duration
func HttpLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("http",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("content-type", w.Header().Get("Content-Type")),
			slog.Duration("duration", time.Since(start)))
	})
}
