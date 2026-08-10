// goDingFileServer.go - File server helper for GoDing static files
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// Package serve provides HTTP handlers and helpers for serving web assets.
package serve

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/HomeDing/goding/internal/global"
)

// safeJoin ensures no directory traversal (../) can escape the base directory.
func safeJoin(baseDir, reqPath string) (string, error) {
	cleaned := filepath.Clean("/" + reqPath)
	fullPath := filepath.Join(baseDir, cleaned)

	// Ensure resulting path stays within baseDir
	if !strings.HasPrefix(fullPath+string(os.PathSeparator), filepath.Clean(baseDir)+string(os.PathSeparator)) {
		return "", os.ErrPermission
	}
	return fullPath, nil
}

// GoDingFileServer returns an HTTP handler that serves static files from the configured web folder.
// It validates the provided folder path, resolves the absolute base path, and prevents directory traversal.
// Only GET and HEAD requests are allowed. Requests for directories are mapped to index.htm if available.
func GoDingFileServer(webFolder string) http.HandlerFunc {
	var basePath string

	if strings.Contains(webFolder, "..") ||
		strings.Contains(webFolder, ":") {
		slog.Error("web folder parameter is invalid.")
		os.Exit(-1)
	}

	basePath, err := filepath.Abs(global.WebFolder)
	if err != nil {
		slog.Error("web folder parameter is not a folder.")
		os.Exit(-1)
	}

	stat, err := os.Stat(basePath)
	if err != nil {
		slog.Error("web folder doesn't exist.")
		os.Exit(-1)

	} else if !stat.IsDir() {
		slog.Error("web folder parameter is not a folder.")
		os.Exit(-1)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var path string // path in the request and full path in the fs to the requested file
		var info os.FileInfo
		var err error

		// Validate method, Expect GET and HEAD requests only for static files
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		path = r.URL.Path
		if strings.Contains(path, "..") ||
			strings.Contains(path, ":") {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Determine safe filesystem path
		path, err = safeJoin(basePath, r.URL.Path)
		if err != nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Now path is the full path.

		if strings.Contains(path, string(os.PathSeparator)+"_") {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// check file or folder exists
		info, err = os.Stat(path)

		if err != nil {
			http.NotFound(w, r)
			return
		}

		// If a folder is requested try to serve index.htm
		if info.IsDir() {
			indexPath := filepath.Join(path, "index.htm")
			info, err = os.Stat(indexPath)

			if err != nil || info.IsDir() {
				http.NotFound(w, r)
				return
			}
			path = indexPath
		}

		slog.Info("http.serving", "file", path)

		// Serve the file found
		http.ServeFile(w, r, path)
	}
}

// End.
