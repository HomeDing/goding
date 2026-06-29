# goding

experimental HomeDing device implementation using the GO language


## One application for multiple services

* Web Server
* Midi Receiver
* Serial Receiver fpr Arduino

Trigger actions on the PC

* Keyboard events
* Midi Events
* Control System parameters


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
