# Project structure

The repository is organized around command packages, reusable internal packages,
and the web frontend:

```text
goding/
├── .github/                      -- GitHub configuration and Copilot instructions
│   └── copilot-instructions.md   -- project-specific Copilot guidance
├── api/                       HTTP API handlers
├── assets/                    Embedded application icon assets
├── cmd/                       Command-line entry points
│   ├── help/                  Help command
│   ├── list/                  MIDI input-device listing command
│   ├── midi/                  MIDI listener command and action dispatch
│   └── serve/                 HTTP server and static-file serving
├── doc/                       Project, package, and feature documentation
├── internal/                  Reusable internal packages
│   ├── actionqueue/           Asynchronous action queue
│   ├── audio/                 Windows audio devices, endpoints, and sessions
│   ├── common/                Interfaces shared by application elements
│   ├── config/                Configuration loading and element definitions
│   ├── devicechange/          Shared device-change notification subscriptions
│   ├── elements/              Volume and MIDI element implementations
│   │   └── registry/          Runtime element registry
│   ├── global/                Shared application configuration and state
│   ├── http/verbose/          Structured HTTP request-logging middleware
│   ├── midi/                  WinMM MIDI discovery and input callbacks
│   ├── osd/                   Windows on-screen display and message handling
│   └── win/                   Windows-specific utilities
├── web/                       Static frontend pages, styles, and scripts
├── main.go                    Application entry point and lifecycle
├── Makefile                   Build, test, and run targets
├── go.mod, go.sum             Go module and dependency metadata
└── README.md, memo.md          Project overview and design notes
```

## Go package responsibilities

| Package                      | Responsibility                                                          |
| ---------------------------- | ----------------------------------------------------------------------- |
| `api`                        | Implements HTTP API handlers, including the application state endpoint. |
| `assets`                     | Embeds application icon data for use by the tray.                       |
| `cmd/help`                   | Prints command-line help.                                               |
| `cmd/list`                   | Lists available MIDI input devices.                                     |
| `cmd/midi`                   | Starts the MIDI listener and dispatches matching actions.               |
| `cmd/serve`                  | Runs the HTTP server and serves API and web content.                    |
| `internal/actionqueue`       | Queues and dispatches actions asynchronously.                           |
| `internal/audio`             | Wraps Windows audio-device, endpoint, and session operations.           |
| `internal/common`            | Defines shared element and endpoint interfaces.                         |
| `internal/config`            | Loads configuration and describes configured elements.                  |
| `internal/devicechange`      | Distributes Windows device-change signals to subscribers.               |
| `internal/elements`          | Implements application elements, including volume and MIDI controls.    |
| `internal/elements/registry` | Registers and retrieves runtime element instances.                      |
| `internal/global`            | Holds shared application flags, settings, and runtime state.            |
| `internal/http/verbose`      | Provides structured HTTP request-logging middleware.                    |
| `internal/midi`              | Enumerates WinMM MIDI inputs and manages per-device callbacks.          |
| `internal/osd`               | Implements the Windows OSD window and forwards device-change messages.  |
| `internal/win`               | Provides small Windows-specific helpers.                                |


## Implementation using Goroutines

In Go the term "goroutine" is used to implement co-routines using the Go’s lightweight concurrent execution unit.

In the goding project goroutines are used for several tasks that run independently
but connected. Goroutines are implemented as a function that is called using `go func()`
and are using the same address space.

* [/main.go](/main.go) —- The main entry point in the program creates a signal-watching
  goroutine that waits for Ctrl+C or termination signals and triggers shutdown of the
  help, web, and MIDI services.

* goding/cmd/serve/serve.go -— The HTTP web server runs in a goroutine started by
  ListenAndServe so the app can keep serving requests while the main process remains
  active.

* goding/cmd/midi/midi.go — The MIDI listener runs in its own goroutine, continuously
  receiving MIDI events and dispatching matching actions when messages arrive.

The preferred way for communication between goroutines is to use channels. Channels behave like a FIFO queue for values and can buffer multiple entries.

* `quitChannel` in [/main.go](/main.go) is used to capture signals (interrupts) and trigger
  quitting the application (shutdown and exit). Also the goding/main.go creates a `sync.WaitGroup` to synchronize the lifecycle of all current goroutines.

## ActionQueue

The [/internal/actionqueue](/internal/actionqueue/actionqueue.go) package provides a
mutex-protected FIFO for actions. A background worker starts when the first action is
added and dispatches queued actions asynchronously in order. `AddOnce` replaces pending
actions for the same target (the part before `?`), which is used e.g. for MIDI-triggered
actions when they appear frequently and only the last action has to be in effect.


### Additional coroutine examples in the workspace

* goding/notify.go — This sample starts a goroutine to monitor Windows audio-session events and process session notifications asynchronously.

* goding/notify2.go — This companion sample also uses a goroutine to watch audio session activity and handle callback-based updates in the background.

