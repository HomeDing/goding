package midihandler

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	// "github.com/go-delve/delve/pkg/dwarf/reader"
	"gitlab.com/gomidi/midi/v2"
)

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

// Start begins listening for MIDI events on the specified input port
func (ml *MIDIListener) Start(inputPort int) error {
	ml.mu.Lock()
	if ml.isRunning {
		ml.mu.Unlock()
		slog.Warn("MIDI listener already running")
		return nil
	}
	ml.isRunning = true
	ml.mu.Unlock()

	slog.Info("Starting MIDI listener", slog.Int("inputPort", inputPort))

	// Run the listen loop in this goroutine (caller handles the blocking)
	ml.listenLoop(inputPort)
	return nil
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

// listenLoop continuously reads MIDI events from the input port
func (ml *MIDIListener) listenLoop(inputPort int) {
	defer func() {
		ml.mu.Lock()
		ml.isRunning = false
		ml.mu.Unlock()
		slog.Info("MIDI listener stopped")
	}()

	// Get available MIDI input ports
	in, err := midi.FindInPort("*")
	if err != nil {
		slog.Error("Failed to find MIDI input ports", slog.Any("error", err))
		return
	}

	// Open the MIDI input port
	stop, err := midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		var bt []byte
		var ch, key, vel uint8
		switch {
		case msg.GetSysEx(&bt):
			fmt.Printf("got sysex: % X\n", bt)
		case msg.GetNoteStart(&ch, &key, &vel):
			fmt.Printf("starting note %s on channel %v with velocity %v\n", midi.Note(key), ch, vel)
		case msg.GetNoteEnd(&ch, &key):
			fmt.Printf("ending note %s on channel %v\n", midi.Note(key), ch)
		default:
			// ignore
		}
	}, midi.UseSysEx())

	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	time.Sleep(time.Second * 5)

	stop()
}
