# Web Server Documentation

## Overview

This service is a Go HTTP server providing a REST API and static file hosting for frontend assets. It uses only the Go standard library and structured logging via slog.

https://notebooklm.google.com/

Implemented features

- Static file serving at /static/ (serves embedded or on-disk files).
- SPA-friendly fallback: requests to unknown paths under /static/ return index.html when present.
- API endpoint: GET /api/state (provided by api.HandleStatus).
- Embedding support: frontend files can be embedded using Go's embed package for a single-binary distribution.
- Verbose request logging middleware using internal/http/verbose.go and slog.

CLI options

- --port <port>  (default: 8080) — set listening port.
- --verbose      (default: false) — enable verbose request logging (slog level/info output).

Examples

Build and run (development):

  go run ./cmd/server --port 8080 --verbose

Build release binary:

  go build -o bin/server ./cmd/server
  ./bin/server --port 8080

Sample curl requests

- Check API state:

  curl -v http://localhost:8080/api/state

- Fetch static file:

  curl -v http://localhost:8080/static/index.html

Logging & middleware

- Uses slog for structured logs.
- internal/http/verbose.go provides an HTTP middleware that logs method, path, remote addr, duration, and response status. Enable with --verbose or wrap the mux with verbose.HttpLogging(mux).

Embedding static files

- To embed web assets, place them under a web/ or static/ directory and use the //go:embed directive in cmd/server or an internal package. The server will serve embedded files when configured; otherwise it falls back to the on-disk file server.

Extending

- Add routes to cmd/server using http.ServeMux and mux.Handle or mux.HandleFunc.
- Keep middleware simple: chain logging and any authentication middlewares around the mux.

Testing

- Use curl or a browser to verify static files and API routes.
- For embedded assets, build the binary and run to confirm files are included.

Notes

- The server intentionally uses only the standard library to keep dependencies minimal and the binary portable.
- Routes should use ServeMux style patterns (e.g., "/api/state", "/static/") rather than framework-specific path templates.

