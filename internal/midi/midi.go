// midi.go - Windows MIDI input device enumeration
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// Package midi provides access to Windows MIDI input devices.
package midi

import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"unicode/utf16"

	"github.com/HomeDing/goding/internal/devicechange"
	winma "github.com/deploymenttheory/go-bindings-win32/bindings/win32/media/audio"
)

const mmMIMData uint32 = 0x03C3

// InputDevice describes a MIDI input device.
type InputDevice struct {
	ID   uint32
	Name string
}

// MessageType identifies a supported short MIDI message.
type MessageType uint8

const (
	MessageNoteOn MessageType = iota + 1
	MessageNoteOff
	MessageControlChange
	MessageProgramChange
)

// Message is a supported short MIDI input message. Channel numbers are zero-based.
type Message struct {
	Type    MessageType
	Channel uint8
	Data1   uint8
	Value   uint8
}

// String formats messages as "[channel] ON note", "[channel] OFF note",
// "[channel] CC controller", or "[channel] PC program". Note names use the
// conventional flat spellings (Db, Eb, Gb, Ab, Bb) and MIDI octave numbering.
func (message Message) String() string {
	switch message.Type {
	case MessageNoteOn:
		return fmt.Sprintf("[%d] ON %s", message.Channel, noteName(message.Data1))
	case MessageNoteOff:
		return fmt.Sprintf("[%d] OFF %s", message.Channel, noteName(message.Data1))
	case MessageControlChange:
		return fmt.Sprintf("[%d] CC %d", message.Channel, message.Data1)
	case MessageProgramChange:
		return fmt.Sprintf("[%d] PC %d", message.Channel, message.Value)
	default:
		return ""
	}
}

type inputListener struct {
	handler func(Message)
}

var inputListeners sync.Map
var nextInputListenerID atomic.Uintptr

type openedInputDevice struct {
	device     InputDevice
	handle     winma.HMIDIIN
	listenerID uintptr
}

// midiInCallback is the callback function for MIDI input events.
// It is called by the Windows API when a MIDI message is received.
var midiInCallback = syscall.NewCallback(func(_ uintptr, message uint32, instance uintptr, data uintptr, _ uintptr) uintptr {
	if message != mmMIMData || instance == 0 {
		return 0
	}

	value, ok := inputListeners.Load(instance)
	if !ok {
		return 0
	}
	listener := value.(*inputListener)
	if midiMessage, ok := decodeShortMessage(data); ok {
		listener.handler(midiMessage)
	}
	return 0
})

// ListInputDevices returns the available MIDI input devices.
func ListInputDevices() []InputDevice {
	deviceCount := winma.MidiInGetNumDevs()
	devices := make([]InputDevice, 0, deviceCount)

	for deviceID := uint32(0); deviceID < deviceCount; deviceID++ {
		var caps winma.MIDIINCAPSW
		result := winma.MidiInGetDevCaps(uintptr(deviceID), &caps, uint32(len(caps.Data)))
		if result != 0 {
			slog.Warn("midi device capabilities unavailable", "device", deviceID, "result", result)
			continue
		}

		devices = append(devices, InputDevice{
			ID:   deviceID,
			Name: decodeDeviceName(caps.Data[8:]),
		})
	}

	return devices
}

// ListenInputDevice listens for supported messages from matching MIDI input
// devices until a value is received from stop. It blocks until stopped.
// A device name of "*" matches every input device. Devices are rescanned so
// devices connected after this function starts are opened automatically.
func ListenInputDevice(deviceName string, stop <-chan bool, handler func(Message)) error {
	if handler == nil {
		return fmt.Errorf("MIDI message handler is nil")
	}

	deviceChanges, unsubscribe := devicechange.Subscribe()
	defer unsubscribe()

	opened := make(map[uint32]*openedInputDevice)
	defer func() {
		for _, device := range opened {
			closeInputDevice(device)
		}
	}()

	waiting := false
	for {
		devices := matchingInputDevices(ListInputDevices(), deviceName)
		wanted := make(map[uint32]InputDevice, len(devices))
		for _, device := range devices {
			wanted[device.ID] = device
		}

		for id, device := range opened {
			if candidate, ok := wanted[id]; ok && candidate.Name == device.device.Name {
				continue
			}
			slog.Info("midi input device disconnected", "device", device.device.Name, "id", device.device.ID)
			closeInputDevice(device)
			delete(opened, id)
		}

		for _, device := range devices {
			if _, ok := opened[device.ID]; ok {
				continue
			}
			openedDevice, err := openInputDevice(device, handler)
			if err != nil {
				slog.Warn("midi input device open failed", "device", device.Name, "id", device.ID, "error", err)
				continue
			}
			opened[device.ID] = openedDevice
			slog.Info("midi input device opened", "device", device.Name, "id", device.ID)
		}

		if len(opened) == 0 && !waiting {
			slog.Warn("waiting for MIDI input device", "device", deviceName)
			waiting = true
		} else if len(opened) > 0 {
			waiting = false
		}

		select {
		case <-stop:
			return nil
		case <-deviceChanges:
		}
	}
}

func matchingInputDevices(devices []InputDevice, deviceName string) []InputDevice {
	deviceName = strings.ToLower(deviceName)
	matches := make([]InputDevice, 0, len(devices))
	for _, device := range devices {
		if deviceName == "*" || strings.Contains(strings.ToLower(device.Name), deviceName) {
			matches = append(matches, device)
		}
	}
	return matches
}

func openInputDevice(device InputDevice, handler func(Message)) (*openedInputDevice, error) {
	listenerID := nextInputListenerID.Add(1)
	inputListeners.Store(listenerID, &inputListener{handler: handler})
	opened := &openedInputDevice{device: device, listenerID: listenerID}
	openResult := winma.MidiInOpen(&opened.handle, device.ID, midiInCallback, listenerID, winma.CALLBACK_FUNCTION)
	if openResult != 0 {
		inputListeners.Delete(listenerID)
		return nil, fmt.Errorf("midiInOpen failed with MMRESULT %d", openResult)
	}
	if opened.handle == 0 {
		inputListeners.Delete(listenerID)
		return nil, fmt.Errorf("midiInOpen succeeded without returning a handle")
	}
	if result := winma.MidiInStart(opened.handle); result != 0 {
		if closeResult := winma.MidiInClose(opened.handle); closeResult != 0 {
			slog.Error("midiInClose failed after start failure", "device", device.Name, "result", closeResult)
		}
		inputListeners.Delete(listenerID)
		return nil, fmt.Errorf("midiInStart failed with MMRESULT %d (handle %#x)", result, uintptr(opened.handle))
	}
	return opened, nil
}

func closeInputDevice(device *openedInputDevice) {
	if result := winma.MidiInStop(device.handle); result != 0 {
		slog.Warn("midiInStop failed", "device", device.device.Name, "result", result)
	}
	if result := winma.MidiInReset(device.handle); result != 0 {
		slog.Warn("midiInReset failed", "device", device.device.Name, "result", result)
	}
	if result := winma.MidiInClose(device.handle); result != 0 {
		slog.Error("midiInClose failed", "device", device.device.Name, "result", result)
	}
	inputListeners.Delete(device.listenerID)
}

// decodeShortMessage decodes a short MIDI message from the given data pointer.
// It returns the decoded message and a boolean indicating whether the message is supported.
func decodeShortMessage(data uintptr) (Message, bool) {
	status := byte(data)
	message := Message{
		Channel: status & 0x0f,
		Data1:   byte(data >> 8),
		Value:   byte(data >> 16),
	}

	switch status & 0xf0 {
	case 0x80:
		message.Type = MessageNoteOff
		message.Value = 0
	case 0x90:
		if message.Value == 0 {
			message.Type = MessageNoteOff
		} else {
			message.Type = MessageNoteOn
		}
	case 0xb0:
		message.Type = MessageControlChange
	case 0xc0:
		message.Type = MessageProgramChange
		message.Value = message.Data1
	case 0xe0, 0xa0, 0xd0:
		return Message{}, false
	default:
		return Message{}, false
	}

	return message, true
}

// noteName returns the conventional flat spelling of the given MIDI note number, using MIDI octave numbering.
// For example, note 60 is "C4", note 61 is "Db4", and note 72 is "C5".
func noteName(note uint8) string {
	names := [...]string{"C", "Db", "D", "Eb", "E", "F", "Gb", "G", "Ab", "A", "Bb", "B"}
	return fmt.Sprintf("%s%d", names[note%12], note/12)
}

// decodeDeviceName decodes a UTF-16LE encoded device name from the given byte slice.
// It returns the decoded string, stopping at the first null terminator.
func decodeDeviceName(data []byte) string {
	name := make([]uint16, 0, 32)
	for offset := 0; offset+1 < len(data); offset += 2 {
		char := binary.LittleEndian.Uint16(data[offset : offset+2])
		if char == 0 {
			break
		}
		name = append(name, char)
	}
	return string(utf16.Decode(name))
}

// End.
