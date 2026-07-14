# goDing -  Using Windows as a HomeDing device

HomeDing is a lightweight network‑communication protocol definition for devices that expose a small, self‑contained web server.

Instead of relying on cloud platforms or heavy protocols, HomeDing uses REST+JSON APIs and a simple action‑based messaging model to let devices communicate with each other over your local WiFi network.

This enables building an eco system of multiple devices that can run independently without the need on any central server.

There are multiple implementations available supporting the communication:

* Based on the Arduino programming environment for devices using the ESP8266 and ESP32
  microprocessors. This can be found in the [Github HomeDing project](https://github.com/HomeDing/HomeDing)
  and is documented in detail on <https://homeding.github.io/>.

* A web server based on nodeJS to implement a portal / dashboard for all local available
  devices that can be found in the project [Github HomeDing/WebFiles project](https://github.com/HomeDing/WebFiles).

* A Go language based implementation to enable a Windows based PC or Laptop to act as a device
  that can be found in the project [Github HomeDing/GoDing project](https://github.com/HomeDing/GoDing) (This project).

## Status and Plan

GoDing is still an experimental HomeDing device implementation.

Current focus is on building the project structure by implementing

* a web server for static files and REST services
* a midi receiver (sensor)
* a volume controller (actor)
* the action queuing and dispatching implementation.


## One application for multiple types services

* A Web Server to have access to settings and current values and control the through a web interface
* Midi Receiver to get any midi messages like fader changes, button pushes or playing notes via a midi device.
* A Serial Receiver for Arduino based devices sending fader changes, button pushes via a serial interface
* maybe some other too ?

These servers can run in parallel and handle / trigger actions like :

* Keyboard events
* Other Midi Events
* Change and Control System parameters like the output volume


## Why a Go based application ?


## GoRoutines

Go's concurrency model aka Goroutines is one of the language's most compelling features
hardly to find in other programming languages and by far less complex than using Threads
from the Operation system. Therefore Go was chosen to implement the HomeDing compatible
server for Windows PCs.

## Command / Production Concurrency Patterns

Init()
Help()
ParseArgs(args []string)

* `Run(wg *sync.WaitGroup)` without background GoRoutine
- ignore the WaitGroup parameter
- just do what needs to be done.
- return 

* `Run(wg *sync.WaitGroup)` with background GoRoutine
- Set state to "started" and Add() to Waitgroup
- start go goroutine
- return 

* GoRoutine (controlChannel chan, wg *sync.WaitGroup)
  - Run the task
  - controlChannel signals a stop, then wg.Done() and exit the goRoutine


wg.Add(n) before launching goroutines, not inside them. If the goroutine is


 — and one of
the most misunderstood.  and channels give you powerful primitives, but using
them correctly in production requires understanding not just the mechanics but the
patterns that prevent bugs, leaks, and subtle race conditions.

This tutorial covers production-ready concurrency patterns in Go, grounded in real-world usage. By the end you will understand when to use each pattern, what can go wrong, and how to measure the performance difference.

All code examples target Go 1.22+ and have been validated to compile and run correctly.

## See also

* [Go Goroutines and Channels](https://tutorials.technology/tutorials/go-goroutines-channels-concurrency-2026.html)
