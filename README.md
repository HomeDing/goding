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


Go's concurrency model aka `Goroutines` is one of the language's most compelling
features hardly to find in other programming languages and by far less complex than
using Threads from the Operation system. Therefore Go was chosen to implement the
HomeDing compatible server for Windows PCs.
