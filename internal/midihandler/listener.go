package midihandler

import (
	"log/slog"
	"sync"

	"github.com/go-midi/midi"
	"github.com/go-midi/midi/reader"
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

	if len(in) == 0 {
		slog.Warn("No MIDI input ports found")
		return
	}

	if inputPort >= len(in) {
		slog.Error("MIDI input port index out of range", slog.Int("requested", inputPort), slog.Int("available", len(in)))
		return
	}

	// Open the MIDI input port
	rd, err := midi.ListenOn(in[inputPort])
	if err != nil {
		slog.Error("Failed to open MIDI input port", slog.Any("error", err))
		return
	}
	defer rd.Close()

	slog.Info("MIDI input port opened", slog.String("port", in[inputPort]))

	// Create a MIDI reader
	r := reader.New(rd)

	// Listen for events
	for {
		select {
		case <-ml.stopChan:
			return
		default:
			// Read the next MIDI message with a short timeout
			msg, err := r.ReadMessage()
			if err != nil {
				slog.Error("Failed to read MIDI message", slog.Any("error", err))
				return
			}

			if msg == nil {
				continue
			}

			// Process the MIDI message
			ml.processMIDIMessage(msg)
		}
	}
}

// processMIDIMessage converts MIDI messages to events and triggers handlers
func (ml *MIDIListener) processMIDIMessage(msg midi.Message) {
	ml.mu.RLock()
	handlers := ml.handlers
	ml.mu.RUnlock()

	// Handle Control Change messages
	if cc, ok := msg.(midi.ControlChange); ok {
		event := MIDIEvent{
			Type:       "control_change",
			Channel:    cc.Channel,
			Controller: cc.Controller,
			Value:      cc.Value,
		}
		slog.Debug("MIDI control change",
			slog.Uint8("channel", event.Channel),
			slog.Uint8("controller", event.Controller),
			slog.Uint8("value", event.Value))

		// Trigger registered handlers in separate goroutines
		for _, handler := range handlers["control_change"] {
			go handler(event)
		}
		for _, handler := range handlers["*"] {
			go handler(event)
		}
	}

	// Handle Note On messages
	if no, ok := msg.(midi.NoteOn); ok {
		event := MIDIEvent{
			Type:     "note_on",
			Channel:  no.Channel,
			Note:     no.Note,
			Velocity: no.Velocity,
		}
		slog.Debug("MIDI note on",
			slog.Uint8("channel", event.Channel),
			slog.Uint8("note", event.Note),
			slog.Uint8("velocity", event.Velocity))

		for _, handler := range handlers["note_on"] {
			go handler(event)
		}
		for _, handler := range handlers["*"] {
			go handler(event)
		}
	}

	// Handle Note Off messages
	if nf, ok := msg.(midi.NoteOff); ok {
		event := MIDIEvent{
			Type:     "note_off",
			Channel:  nf.Channel,
			Note:     nf.Note,
			Velocity: nf.Velocity,
		}
		slog.Debug("MIDI note off",
			slog.Uint8("channel", event.Channel),
			slog.Uint8("note", event.Note),
			slog.Uint8("velocity", event.Velocity))

		for _, handler := range handlers["note_off"] {
			go handler(event)
		}
		for _, handler := range handlers["*"] {
			go handler(event)
		}
	}
}

// ListInputPorts returns a list of available MIDI input ports
func ListInputPorts() ([]string, error) {
	ports, err := midi.FindInPort("*")
	if err != nil {
		return nil, err
	}
	return ports, nil
}
