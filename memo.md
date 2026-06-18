# goding

This project will implement a web UI and RESTful services that correspond to the
interface definitions defined in the homeding library.

As a result a Linux or Windows Server can be brought to the ecosystem of the homeding
library.

You can find in this project:

* The project is written in the GO language
* a full stack Web server implemented in the GO language
* element implementations to make advantage of a PC
* embedded user Interface from the from the library to control these elements

the planned elements are:

* controlling the output volume of running programmes and system
* ...


## Implementation approach

* [How To Make an HTTP Server in Go](https://www.digitalocean.com/community/tutorials/how-to-make-an-http-server-in-go)
* [Golang HTTP server: basics](https://medium.com/@bartosz.piekny/golang-http-server-basics-6936ddab7474)

## Implement Request Routing

Routing is the process of deciding which function (or handler) should run when a request
is sent to the http server with a specific URL. This is implemented by multiplexers
for http requests (short `mux`).

Since this application is implemented starting with go 1.26.4 the enhanced standard
routing module `http.NewServeMux` of go is used in this application.

With the help of this mux the function that implement the functionality of a specific request
like RESTful API calls are registered.

See:

* [net/http documentation](https://pkg.go.dev/net/http)
* [Better HTTP server routing in Go 1.22](https://eli.thegreenplace.net/2023/better-http-server-routing-in-go-122/)
* TODO: check [multiplexer](https://qbit-glitch.github.io/golang_notes/projects/RestAPI/multiplexer.html)

## Logging

Locking is implemented in this project using the `slog` package for structured locking to avoid dependencies to other external libraries.

``` go
import ( "log/slog")

slog.Info("http",
  slog.String("method", r.Method),
  slog.String("path", r.URL.Path),
  slog.String("content-type", w.Header().Get("Content-Type")),
  slog.Duration("duration", time.Since(start)))
```


## Verbose Request logging

Instead of wrapping a `HandlerFunc` that is registered in the mux into a logging function
the mux handler can be chained after the VerboseHandler that is implemented in the package
`verbose` in "goding/internal/http"

TODO: This is enabled on the command line when starting the application using `--verbose`.


How to add middleware handlers for verbose logging and http level analysis.


 mux.HandleFunc("GET /api/state", logTimeElapsed(api.HandleStatus))

[text](https://homeding.github.io/concepts/paper03.htm)


Sending Actions by using URLs by using the notation and syntax of Actions you can use a
URL to pass an action into a device manually e.g. by using
`http://(devicename)/api/state/digitalOut/D5?value=1`

.

* [Middleware Patterns in Go](https://drstearns.github.io/tutorials/gomiddleware/)

## GO

* `go version` -- go version go1.26.4 windows/amd64
* `go build cmd\server\main.go`
* `go test ./...`

``` txt
Usage:
        go <command> [arguments]

The commands are:

        bug         start a bug report
        build       compile packages and dependencies
        clean       remove object files and cached files
        doc         show documentation for package or symbol
        env         print Go environment information
        fix         apply fixes suggested by static checkers
        fmt         gofmt (reformat) package sources
        generate    generate Go files by processing source
        get         add dependencies to current module and install them
        install     compile and install packages and dependencies
        list        list packages or modules
        mod         module maintenance
        work        workspace maintenance
        run         compile and run Go program
        telemetry   manage telemetry data and settings
        test        test packages
        tool        run specified go tool
        version     print Go version
        vet         report likely mistakes in packages

Use "go help <command>" for more information about a command.

Additional help topics:

        buildconstraint build constraints
        buildjson       build -json encoding
        buildmode       build modes
        c               calling between Go and C
        cache           build and test caching
        environment     environment variables
        filetype        file types
        goauth          GOAUTH environment variable
        go.mod          the go.mod file
        gopath          GOPATH environment variable
        goproxy         module proxy protocol
        importpath      import path syntax
        modules         modules, module versions, and more
        module-auth     module authentication using go.sum
        packages        package lists and patterns
        private         configuration for downloading non-public code
        testflag        testing flags
        testfunc        testing functions
        vcs             controlling version control with GOVCS

Use "go help <topic>" for more information about that topic.
```

## Make (not installed)

make build
make test
make run
make clean

