package midi

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/HomeDing/goding/internal/global"

	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

var isStarted = false
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

// Initialize the midi command and its flags in the init function, which is called before main.
// This allows us to set up the command and its flags before parsing the command-line arguments in main.
func Init() {
	slog.Debug("midi.Init()")

	// always good to close the driver at the end
	defer midi.CloseDriver()

	// define the flags for the midi command
	midiFlags = flag.NewFlagSet("midi", flag.ExitOnError)
	midiFlags.BoolVar(&global.VerboseFlag, "verbose", false, "enable verbose logging")
} // Init()

func Help() {
	slog.Debug("Midi.Help()")
}

// ParseArgs parses the command-line arguments and sets the global variables accordingly.
// It returns a boolean indicating whether the serve command was invoked and an error if there was an issue with parsing the arguments.
func ParseArgs(args []string) (bool, error) {
	slog.Debug("midi.ParseArgs()", slog.Any("args", args))

	midiFlags.Parse(args)
	return true, nil
}

// Register a action to be created on a specific signal.
// The action is a string that represents the action to be taken when the signal is received. The signal is a string that represents the signal to listen for. The action can be a command to execute, a function to call, or any other action that can be represented as a string.
func Register(midiMsg string, action string) {
	// normalize the midiMsg to be used as a key for the registered actions.
	midiMsg = strings.ToLower(strings.ReplaceAll(midiMsg, " ", ""))

	slog.Debug("midi.register", slog.String("midiMsg", midiMsg), slog.String("action", action))
	actionsRegistry[midiMsg] = action
}

// Go Routine to listen for MIDI events and process them using registered handlers
func listen(quitChan chan bool, wg *sync.WaitGroup) error {

	slog.Debug("midi.listen()")

	in, err := midi.FindInPort("LPD8 mk2")
	if err != nil {
		fmt.Println("can't find LPD8 mk2")
		return nil

	} else {

		// Listen to the MIDI messages of type NoteOn, NoteOff, ControlChange, and ProgramChange on the specified input port.
		// and create a midi message string to be used as a key for the registered actions.
		stop, err := midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
			var ch, key, value uint8
			var ctl uint8
			var midiMsg string

			// [Channel] <Message> <Note/Control> <Value>
			value = 0

			switch {
			case msg.GetNoteStart(&ch, &key, &value):
				midiMsg = fmt.Sprintf("[%v] ON %s", ch, midi.Note(key))

			case msg.GetNoteEnd(&ch, &key):
				midiMsg = fmt.Sprintf("[%v] OFF %s", ch, midi.Note(key))

			case msg.GetControlChange(&ch, &ctl, &value):
				midiMsg = fmt.Sprintf("[%v] CC %v", ch, ctl)

			case msg.GetProgramChange(&ch, &value):
				midiMsg = fmt.Sprintf("[%v] PC %v", ch, value)

			default:
				// ignore !
				// fmt.Printf("midi.DEFAULT channel\n")
				// slog.Debug("midi", slog.Any("msg", msg))

			}

			// if global.VerboseFlag && len(midiMsg) > 0 {
			// 	log.Println("MIDI", midiMsg, value)
			// }

			// normalize the midiMsg to be used as a key for the registered actions.
			midiMsg = strings.ToLower(strings.ReplaceAll(midiMsg, " ", ""))

			var a = actionsRegistry[midiMsg]
			if len(a) > 0 {
				log.Printf("midi.action %s,%v => %s", midiMsg, value, a)
			} else {
				log.Printf("midi.noaction %s,%v", midiMsg, value)
			}
		}) // , midi.UseSysEx()

		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			return (nil)
		}

		// wait for quit signal
		<-quitChan

		stop() // stop the midi listener

		slog.Debug("midi.stopped!")
		wg.Done()
		isStarted = false
	}

	return nil
}

func Run(wg *sync.WaitGroup) error {
	slog.Debug("midi.Run()")

	// start a goroutine to listen to midi events

	// test some Registers
	Register("[11] ON E3", "Pad 1")
	Register("[14] CC 77", "volume change")

	isStarted = true
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

	fmt.Println("in ports: \n" + midi.GetInPorts().String())
	// fmt.Println("out ports: \n" + midi.GetOutPorts().String())

	return nil
}

// func FindInPort(name string) (drivers.In, error)

//   var ml = GetInstance()

// 	ml.Start()

// 	return nil
// }

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
