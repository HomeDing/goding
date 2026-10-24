// midi.go - MIDI event handling and listener management
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// Command to start the midi listener
// The midi package provides functionality for handling MIDI events, including listening for MIDI messages,
// registering handlers for specific events, and managing the lifecycle of a MIDI listener.
// It uses the gomidi library to interface with MIDI devices and supports concurrent access through mutexes.
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
package midi

import (
	"flag"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"

	actionQueue "github.com/HomeDing/goding/internal/actionqueue"
	"github.com/HomeDing/goding/internal/global"
	midiDevices "github.com/HomeDing/goding/internal/midi"
)

// Package-level variables for managing the MIDI listener state and registered actions.

// isStarted indicates internally whether the MIDI listener has been started.
var isStarted = false

// quitChan is used to signal the MIDI listener goroutine to stop listening for events.
var quitChan = make(chan bool)

// ----- Command implementation

// MIDIListener handles MIDI event listening and processing
type MIDIListener struct {
	isRunning bool
	stopChan  chan struct{}
	handlers  map[string][]MIDIEventHandler
	mu        sync.RWMutex
}

// MIDIEventHandler is a function that processes MIDI events
type MIDIEventHandler func(event MIDIEvent)

// MIDIEvent represents a MIDI event
type MIDIEvent struct {
	Type       string // "control_change", "note_on", "note_off", etc.
	Channel    uint8
	Controller uint8 // for control_change
	Value      uint8
	Note       uint8 // for note_on/off
	Velocity   uint8
}

var listener *MIDIListener

var once sync.Once

var actionsRegistry map[string]string = make(map[string]string) // map of midi message to action string

// ----- Command implementation

// midi command parameters
var midiFlags *flag.FlagSet

// normalize the midi message to be used as a key for the registered actions.
func normalizeMidiMessage(midiMsg string) string {
	return strings.ToLower(strings.ReplaceAll(midiMsg, " ", ""))
}

// Initialize the midi command and its flags in the init function, which is called before main.
// This allows us to set up the command and its flags before parsing the command-line arguments in main.
func init() {
	// define the flags for the midi command
	midiFlags = flag.NewFlagSet("midi", flag.ExitOnError)
	midiFlags.BoolVar(&global.VerboseFlag, "verbose", global.VerboseFlag, "enable verbose logging")
} // init()

func Help() {
	slog.Debug("Midi.Help()")

	fmt.Println()
	fmt.Println(
		`goDing midi starts the midi listener receiving midi messages
to execute create HomeDing actions as defined in the configuration.

Usage:

  goding midi [parameters]`)
	fmt.Println()
	midiFlags.Usage()
} // Help()

// ParseArguments parses the command-line arguments and sets the global variables accordingly.
// It returns a boolean indicating whether the serve command was invoked and an error if there was an issue with parsing the arguments.
func ParseArguments(args []string) (bool, error) {
	slog.Debug("midi.ParseArguments()", slog.Any("args", args))

	midiFlags.Parse(args)
	return true, nil
} // ParseArguments()

// Register a action to be created on a specific signal.
// The action is a string that represents the action to be taken when the signal is received. The signal is a string that represents the signal to listen for. The action can be a command to execute, a function to call, or any other action that can be represented as a string.
func Register(midiMsg string, action string) {
	// normalize the midiMsg to be used as a key for the registered actions.
	midiMsg = normalizeMidiMessage(midiMsg)

	slog.Debug("midi.register", slog.String("midiMsg", midiMsg), slog.String("action", action))
	actionsRegistry[midiMsg] = action
} // Register()

// Go Routine to listen for MIDI events and process them using registered handlers
func listen(quitChan chan bool, wg *sync.WaitGroup) error {

	slog.Debug("midi.listen()")
	defer wg.Done()
	defer func() { isStarted = false }()

	err := midiDevices.ListenInputDevice("LPD8 mk2", quitChan, func(message midiDevices.Message) {
		midiMsg := normalizeMidiMessage(message.String())
		action := actionsRegistry[midiMsg]
		if len(action) > 0 {
			slog.Info("midi.action", "message", midiMsg, "value", message.Value, "action", action)
			action = strings.ReplaceAll(action, "$v", strconv.Itoa(int(message.Value)))
			actionQueue.AddOnce(action)
			return
		}

		slog.Info("midi.action", "message", midiMsg, "value", message.Value)
	})
	if err != nil {
		slog.Error("midi.listen", "error", err)
		return err
	}

	slog.Debug("midi.stopped!")
	return nil
}

func Run(wg *sync.WaitGroup) error {
	slog.Debug("midi.Run()")

	// start a goroutine to listen to midi events
	isStarted = true
	slog.Info("Listening for midi messages...")
	wg.Add(1)
	go listen(quitChan, wg)

	return nil
}

func Stop() {
	slog.Debug("midi.Stop()")
	if isStarted {
		slog.Debug("midi Quit Signal")
		quitChan <- true
	}
}

func List() error {
	slog.Debug("midi.List()")

	fmt.Println("in ports:")
	devices := midiDevices.ListInputDevices()
	if len(devices) == 0 {
		fmt.Println("(none)")
		return nil
	}

	for _, device := range devices {
		fmt.Printf("%d: %s\n", device.ID, device.Name)
	}

	return nil
}

// GetInstance returns the singleton MIDIListener instance
func GetInstance() *MIDIListener {
	once.Do(func() {
		listener = &MIDIListener{
			isRunning: false,
			stopChan:  make(chan struct{}),
			handlers:  make(map[string][]MIDIEventHandler),
		}
	})
	return listener
}

// RegisterHandler registers a handler for a specific MIDI event type
func (ml *MIDIListener) RegisterHandler(eventType string, handler MIDIEventHandler) {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	ml.handlers[eventType] = append(ml.handlers[eventType], handler)
	slog.Debug("MIDI handler registered", slog.String("eventType", eventType))
}

// Stop stops the MIDI listener
func (ml *MIDIListener) Stop() {
	ml.mu.Lock()
	if !ml.isRunning {
		ml.mu.Unlock()
		return
	}
	ml.isRunning = false
	ml.mu.Unlock()

	slog.Info("Stopping MIDI listener")
	close(ml.stopChan)
}

// IsRunning returns whether the listener is currently running
func (ml *MIDIListener) IsRunning() bool {
	ml.mu.RLock()
	defer ml.mu.RUnlock()
	return ml.isRunning
}

// End.
