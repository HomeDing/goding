# Copilot instructions for goding

This repository is a Go web application and REST service project.
The project is built with `go 1.26.4` and uses the standard library (`net/http`, `log/slog`, etc.) rather than external frameworks.

## Current structure

* `cmd/server/main.go` - main HTTP server entrypoint
* `api/status.go` - current API handler for `/api/state`
* `internal/http/verbose.go` - request logging middleware
* `memo.md` - design notes, routing strategy, logging/middleware approach, and project goals

## Goals

* Keep the existing REST API behavior intact
* Support a static file web server for a frontend or assets
* Use `http.NewServeMux` to create a standard `ServeMux` and use standard library
  middleware patterns where possible.
* Add or improve command-line options for port selection and verbose logging
* Avoid unnecessary external dependencies
* use the flag package for commandline parsing.

## What to do

* Create a folder structure for static web content, such as `static/` or `web/`
* Serve those static assets from the HTTP server
* Make the file server compatible with the existing mux and route handling
* Keep the existing `api.HandleStatus` route and logging flow
* Implement or improve CLI options for `--port` and `--verbose`
* Use structured logging via `slog` for any logging and HTTP requests
* enable embedding the files from /web into the binary using the embed package

## Developer notes

* The memo emphasizes learning HTTP routing and middleware in Go.
* The current `main.go` uses patterns like `mux.HandleFunc("GET /api/state", api.HandleStatus)` and `verbose.HttpLogging(mux)`.
* Update route registration to use valid `http.ServeMux` patterns and avoid placeholder pseudo-routes such as `/task/{id}/` unless a custom router is added.
* If adding static file support, prefer a clean root path such as `/static/` or a SPA-friendly fallback.

> Use these instructions as the primary guide for changes in this repository. Keep the changes simple, idiomatic, and aligned with the project's existing standard library approach.
