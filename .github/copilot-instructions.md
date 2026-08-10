# Copilot instructions for goding

This repository is a Go web application and REST service project.
The project is built with `go 1.26.4` and uses the standard library (`net/http`, `log/slog`, etc.) rather than external frameworks.

## Current structure

* `/main.go` - entrypoint to the application and command line parsing
* `/cmd/` - all command line startable commands are implemented in this folder
* `/cmd/serve/serve.go` - HTTP server entrypoint, runs a goroutine
* `/cmd/midi/midi.go` - MIDI entrypoint, runs a goroutine
* `/cmd/help/help.go` - Help entrypoint, runs once and terminates the app after run.
* `/web/` -  This folder is the root folder for all static web content.
* `/internal/api/status.go` - current API handler for `/api/state`
* `/internal/http/verbose.go` - request logging middleware
* `/memo.md` - design notes, routing strategy, logging/middleware approach, and project goals
* `/readme.md` - About this project and further info


## Source code and documentation rules

### File Header

Every file containing go sourcecode must have a common file header above the package
definition with a short one line file description and the copyright statement like:

``` txt
// midi.go - MIDI event handling and listener management
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

```

This must be placed before the package statement and separated by a blank line at the end.

A short package description will follow to document the package purpose as a comment to
the package statement.


### Comments at the end of long statements

All statements that are longer than 4 lines and wrappend by brackets should have a
closing comment after the closing bracket mentioning the statement like

- `} // for` or 
- `} // if`.

Optional the closing comment can include the condition or var used in the statement.

For if statements the closing bracket will be the closing bracket of the `else` part of the statement.


### Implementation Style

* With `if` statements this project prefers not to have long if statements and short
  else statements.  
  Instead the if clause should be reversed so the short statement comes first and the
  else statement can be read in sight if the if clause.

* All sourcecode files must have a "End" comment like as the last non-empty line like:
  ```
  // End.
  ```

### Logging and Output

Use structured logging via `slog` for all logging and HTTP requests.

Any other output must be written using the `fmt` package.


## Project Goals

* Keep the existing REST API behavior from HomeDing library
* Support a static file web server for a frontend or assets
* Use `http.NewServeMux` to create a standard `ServeMux` and use standard library
  middleware patterns where possible.
* Add or improve command-line options for port selection and verbose logging
* Avoid unnecessary external dependencies

## Command and Concurrency Patterns

* Init() -- internal initialization of the command package
* Help() -- print a brief help message to output
* ParseArguments(args []string) -- parse the given arguments for configuration
* Run() -- run the command
  - When using no background Goroutine, ignore the WaitGroup parameter andjust do what
    needs to be done.
  - When using a background Goroutine, set the local state to "started", call Add() on
    the Waitgroup and then start the go Goroutine, stop when signalled to stop by the local control channel
    and call Done() on the Waitgroup
* Stop() -- signal a stop command to the local control channel.


## WebServer

The application includes a HTTP server providing a REST API and static file hosting for
frontend assets. It only uses the Go standard library.

### Implemented features

* Static file serving at /web (serves embedded or on-disk files).
* SPA-friendly fallback: requests to unknown paths under /web return index.html when present.
* API endpoints: e.g. GET /api/state (provided by api.HandleStatus).
* Optional Verbose request logging middleware using internal/http/verbose.go and slog.

WebServer CommandLine options:

* "--port"    -- port to listen on
* "--web"     -- folder to serve static files from
* "--verbose" -- enable verbose logging

## MIDI

The application includes a receiver for MIDI.

The relevant events can be configured and actions will be created to trigger changes in the elements.

## Startup Flow and Commands

The startup flow is centralized in `main.go`, which acts as the top-level dispatcher for
the application. Here the commandline parsing is implemented and the required commands are started.

* Use the flag package for commandline parsing.

* The program first initializes all command packages by calling `<command>.Init()`.
  These setup functions register the available flags for each command.

* It then inspects the command line arguments in os.Args:
  * If no subcommand is provided, it defaults to serve.
  * Otherwise, it uses the first argument as the selected command, such as help or serve.
* Before dispatching, it parses the command-specific arguments:
  * help.ParseArguments(os.Args[2:]) handles help-related arguments.
  * serve.ParseArguments(os.Args[2:]) parses the flags for the server command.
* The application then applies logging configuration based on the global verbose flag,
  adjusting the log level before execution.
* Finally, it calls the appropriate command runner:
  * help.Run() for help output
  * serve.Run() for starting the web server

In short, main.go performs three main steps for command startup:

* Initialize command modules and flags
* Parse the selected command and its arguments
* Execute the matching command handler
* This structure keeps command handling simple and makes it easy to extend with
  additional commands later.


All command implementations in the `/cmd` folder are implementing the following entrypoints:

* Init() -- for initialization of anything the command needs before starting any activity. This may include verifications of the environment.
* ParseArguments(args) --  parse all given arguments.
* Help() -- send any helpful information about functionality and parameters to the log output.
* Run() -- run the command once or start a goroutine to run a service.
* Stop() -- run the command once or start a goroutine to run a service.

To handle running of multiple goroutines the Run function has a 
sync.WaitGroup parameter that must be used to signal the running background process. 

func Run(wg *sync.WaitGroup) error {


## Logging

The project uses Go's standard library logging via `slog` rather than third-party
logging packages. Logging is implemented in a lightweight, structured way across
startup, HTTP serving, and MIDI handling.

* Startup and command flow in `main.go` configure the logging level from the `--verbose`
  flag. In normal mode, the app logs at warning level and above; in verbose mode, it
  enables debug-level output.
* The serve command in `cmd/serve/serve.go` exposes the `--verbose` option and applies
  request logging to the HTTP server.
* HTTP request logging is handled by the middleware in
  `internal/http/verbose/verbose.go`, which records the request method, path, content
  type, duration, and emits an `info` log entry for each request.
* MIDI-related events in `internal/midihandler/listener.go` use structured logs for
  lifecycle events such as starting/stopping the listener, registering handlers, and
  reporting errors or warnings.
* The project’s logging practice is to keep logs simple, structured, and
  dependency-light: use `slog` with key/value fields, choose appropriate levels
  (`Debug`, `Info`, `Warn`, `Error`), and avoid ad hoc plain-text logging where
  structured context is useful.

### Further development needed

TODO: Embedding support: frontend files can be embedded using Go's embed package for a single-binary distribution.

## What to do

* enable embedding the files from /web into the binary using the embed package

## Developer notes

* The memo emphasizes learning HTTP routing and middleware in Go.
* The current `main.go` uses patterns like `mux.HandleFunc("GET /api/state", api.HandleStatus)` and `verbose.HttpLogging(mux)`.
* Update route registration to use valid `http.ServeMux` patterns and avoid placeholder pseudo-routes such as `/task/{id}/` unless a custom router is added.
* If adding static file support, prefer a clean root path such as `/static/` or a SPA-friendly fallback.

> Use these instructions as the primary guide for changes in this repository. Keep the changes simple, idiomatic, and aligned with the project's existing standard library approach.

