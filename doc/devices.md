# Windows device detection

## Detecting device changes

The application's Windows message loop dispatches messages to the OSD window. Its window
procedure handles `WM_DEVICECHANGE` and, for `DBT_DEVNODES_CHANGED`, calls
`devicechange.Notify()`. Windows broadcasts this message when a device node is added or
removed. The notification is system-wide; it does not identify a MIDI device, and the
application does not register a separate device-interface notification filter.

## Sharing notifications

The `internal/devicechange` package provides a shared subscription hub. A component
calls `devicechange.Subscribe()` to receive a buffered channel and an unsubscribe
function. `Notify()` sends a non-blocking signal to each current subscriber. The buffer
has capacity one, so multiple changes that arrive before a subscriber handles them are
coalesced into one rescan.

The hub only distributes change signals. It does not enumerate devices, filter for MIDI,
or open device handles. Keeping the notification handler lightweight also lets the
Windows window procedure return promptly.

## MIDI device discovery

The MIDI listener subscribes to device changes before its initial scan. It enumerates
input devices with WinMM's `midiInGetNumDevs` and reads each device's name with
`midiInGetDevCaps`. The configured `--device` value matches a device name
case-insensitively as a substring; `--device "*"` matches all input devices.

After the initial scan, the listener waits for either a stop signal or a device-change
notification. On notification, it enumerates devices again:

1. If an opened device is no longer present (or its name at that ID has changed), the
   listener logs an INFO message and stops, resets, and closes its WinMM handle.
2. Any newly available matching device is opened with `midiInOpen`, using the MIDI
   callback for that specific handle, then started with `midiInStart`.
3. If opening a matching device fails, the listener logs a warning. It will try again on
   the next device-change notification.

Devices already connected at startup are found by the initial scan. There is no periodic
polling; unrelated Windows device changes can also cause a MIDI rescan.

## MIDI message callbacks

Device-change notifications and MIDI-message callbacks are separate mechanisms. WinMM
does not provide one callback for every MIDI input device: the listener registers a
callback separately when it opens each matching device. That callback decodes supported
MIDI messages and passes them to the configured handler.

