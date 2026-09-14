# Project structure

Folder structure

``` txt
goding/
│
├── .github/                      -- GitHub configuration and Copilot instructions
│   └── copilot-instructions.md   -- project-specific Copilot guidance
├── cmd/                          -- command implementations
│   ├── help/                     -- help command and related output
│   ├── midi/                     -- MIDI listener and event handling
│   └── serve/                    -- HTTP/web server implementation
├── doc/                          -- project documentation and design notes
├── internal/                     -- internal packages and shared logic
│   └── actionqueue/              -- FIFO action queue for async work
├── public/                       -- static frontend assets (many files)
├── main.go                       -- application entry point and lifecycle setup
└── go.mod                        -- Go module configuration

└── notify.go                     -- sample Windows audio-session notification watcher
└── notify2.go                    -- companion async notification sample
```


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

For actions more functionality like look ahead is required so the build-in channel
mechanism cannot be used and the
[/internal/actionqueue](/internal/actionqueue/actionqueue.go) is implementing a similar
mechanism FIFO mechanism especially for actions.



### Additional coroutine examples in the workspace

* goding/notify.go — This sample starts a goroutine to monitor Windows audio-session events and process session notifications asynchronously.

* goding/notify2.go — This companion sample also uses a goroutine to watch audio session activity and handle callback-based updates in the background.

