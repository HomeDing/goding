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


## Planned elements

### Volume Element

* controlling the output volume of the system
* controlling the output volume of a specific application
* controlling the output volume of the program with the active window program
* smart volume change (read first and adjust delta only)

### CPU Element

* current usage

### MEM Element

* current size
* current usage

### DISK Element

* current size
* current usage


## Web Server

The web server is used for static files hosting and for
enabling RESTful services like other HomeDing devices do.

Details can be found in [Web Server](doc/webserver.md).


## MIDI receiver

TODO: Details can be found in [MIDI Receiver](doc/midi.md).

TODO: mute for session / applications
TODO: Volume element: Treat all devices and sessions as endpoints and equal
TODO: change output device element
TODO: less debugging info
TODO: configuration from file
TODO: Web UI from HomeDing/WebFiles


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

## Go tools

* go install github.com/go-delve/delve/cmd/dlv@latest

## WCA wrappers

* [roamer](https://github.com/Oppodelldog/roamer/blob/main/internal/audioctl/device.go)  
  replace github.com/moutend/go-wca v0.2.0 => github.com/Oppodelldog/go-wca v0.2.1-0.20211129184456-42e90fa2b475

## See also

* [Go goroutine and Channels](https://tutorials.technology/tutorials/go-goroutine-channels-concurrency-2026.html)

---

<https://github.com/NiklasRichter2222/deej>
