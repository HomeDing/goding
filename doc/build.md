# Build and Run

## Build release binary

use the go compiler directly:

```cmd
go build -v -o goDing.exe main.go
```

or use:

```cmd
make build
```



  ./bin/server --port 8080

Sample curl requests

* Check API state:

  curl -v <http://localhost:8080/api/state>

* Fetch static file:

  curl -v <http://localhost:8080/static/index.html>

Logging & middleware

* Uses slog for structured logs.
* internal/http/verbose.go provides an HTTP middleware that logs method, path, remote addr, duration, and response status. Enable with --verbose or wrap the mux with verbose.HttpLogging(mux).

Embedding static files

* To embed web assets, place them under a web/ or static/ directory and use the //go:embed directive in cmd/server or an internal package. The server will serve embedded files when configured; otherwise it falls back to the on-disk file server.

Extending

* Add routes to cmd/server using http.ServeMux and mux.Handle or mux.HandleFunc.
* Keep middleware simple: chain logging and any authentication middlewares around the mux.

Testing

* Use curl or a browser to verify static files and API routes.
* For embedded assets, build the binary and run to confirm files are included.

Notes

* The server intentionally uses only the standard library to keep dependencies minimal and the binary portable.
* Routes should use ServeMux style patterns (e.g., "/api/state", "/static/") rather than framework-specific path templates.

