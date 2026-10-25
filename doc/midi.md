# MIDI Receiver

## Device discovery

`midi list` enumerates Windows MIDI input devices with `midiInGetNumDevs` and
`midiInGetDevCaps`. It prints each device's numeric ID and name. If reading a device's
capabilities fails, that device is skipped and a warning is logged.

The listener currently opens the input device named `LPD8 mk2`. The name match is
case-insensitive. If the device is not found or cannot be opened, the listener logs an
error and exits.

## Messages and actions

WinMM reports short input messages through its `MIM_DATA` callback. The listener decodes
each callback and accepts note on, note off, control change, and program change
messages. Note on with velocity zero is treated as note off. Other message types,
including pitch bend, channel pressure, and system exclusive messages, are ignored and
do not trigger actions.

Supported messages become action keys using these formats:

| MIDI message   | Key format                | `$v` value       |
| -------------- | ------------------------- | ---------------- |
| Note on        | `[channel] ON note`       | Velocity         |
| Note off       | `[channel] OFF note`      | `0`              |
| Control change | `[channel] CC controller` | Controller value |
| Program change | `[channel] PC program`    | Program number   |

Channels are zero-based. Note names use flat spellings and MIDI octave numbering; for
example, note 60 is `C5`. Before lookup, the key is lowercased and spaces are removed.
For example, `[2] ON C5` becomes `[2]onc5`.

MIDI elements register a key and action through `midi.Register`. When a message matches
a registered key, `$v` in the action is replaced with the value shown above, then the
action is added to the action queue. Messages without a registered action are logged but
do not enqueue anything.

## Implementation

`internal/midi/winmm.go` encapsulates the Windows-specific device discovery, WinMM
handle lifecycle, callback, and short-message decoding. It uses
`github.com/deploymenttheory/go-bindings-win32/bindings/win32/media/audio` for the
Windows MIDI interfaces. The `cmd/midi` package handles command lifecycle and action
dispatch, keeping direct Win32 calls out of the command code.

