# Web Server Documentation

## Overview

This service is a GO HTTP server providing a REST API and static file hosting for
frontend assets. It uses only the Go standard library and structured logging via slog.

## Implemented features

* Static file serving at /static/ (serves embedded or on-disk files).
* SPA-friendly fallback: requests to unknown paths under /static/ return index.html when
  present.
* API endpoint: GET /api/state (provided by api.HandleStatus).
* Embedding support: frontend files can be embedded using Go's embed package for a
  single-binary distribution.
* Verbose request logging middleware using internal/http/verbose.go and slog.
* Implement a custom FileServer `GoDingFileServer` for securing against folder listings
  and delivery of restricted files.

## CLI options

Using the CLI options is available to start GoDing for development and testing. In
production the configuration will be provided by the configuration files and the web
sever will be started by default.

* `serve`        -- command to run the web server
* --port <port>  -- (default: 8080) — set listening port.
* --verbose      -- (default: false) — enable verbose request logging (slog level/info output).

### Examples

To run the web server only (no midi receiving) you can use the following commandline:

``` cmd
go run main.go serve --port 8088 --verbose
```

To shut down the web server remotely the URL `/api/shutdown/` can be fetched (only in
dev mode) or the `serve.Stop()` function is called internally.

### REST Endpoints

Devices in the HomeDing eco system expose the following HTTP endpoints (excerpt):

* `GET /api/state/<type>/<id>` : Retrieve the current state and current values of a
  specific element by type and ID.
* `GET /api/state` : Retrieve the current state and current values for all configured
  elements of the device.
* `GET /env.json` : Retrieve the current JSON configuration of the device itself.
* `GET /config.json` : Retrieve the current JSON configuration of the running elements.
* `GET /api/state/<type>/<id>?<prop>=<value>` : Change a element property or
  configuration value at runtime or trigger an action in the device.
* `GET /api/shutdown` : Shut down the web server (in dev mode only)

These REST services can be used in the GoDing implementation to exchange Windows specific values
and trigger actions on the windows machine for the defined elements.


## Implementation approach

The web server is implemented in the folder [/cmd/serve](/cmd/serve).


### Request Routing

Routing is the process of deciding which function (or handler) should run when a request
is sent to the http server with a specific URL. This is implemented by multiplexers
for http requests (short `mux`).

Since this application is implemented starting with go 1.26.4 the enhanced standard
routing module `http.NewServeMux` of go is used in this application.

With the help of this mux the function that implement the functionality of a specific request
like RESTful API calls are registered.

### Implement Serving Static Files

The web server is not returning file content in all cases for security reasons.

* The folder containing the static files is per default `/web`and can be changed by
  parameter in development mode.
* Folder listing and delivery of files with restricted characters is prohibited.
* Request with the following substrings are denied: `..`, `:`
* Request of files starting with `_` are denied.
* Requesting files outside the web folder are denied.


## Implement Restfull services

* registering of HandleFunc in serve.go
* all using /api as prefix
  * JSON encoding of maps and sets


## See Also

* [How To Make an HTTP Server in Go](https://www.digitalocean.com/community/tutorials/how-to-make-an-http-server-in-go)
* [Golang HTTP server: basics](https://medium.com/@bartosz.piekny/golang-http-server-basics-6936ddab7474)

* [net/http documentation](https://pkg.go.dev/net/http)
* [Better HTTP server routing in Go 1.22](https://eli.thegreenplace.net/2023/better-http-server-routing-in-go-122/)

* TODO: check [multiplexer](https://qbit-glitch.github.io/golang_notes/projects/RestAPI/multiplexer.html)

