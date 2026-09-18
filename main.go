// Package main provides the command line parsing and command execution for the GoDing application.
package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"sync"

	"github.com/gogpu/systray"

	"github.com/HomeDing/goding/assets"
	"github.com/HomeDing/goding/cmd/help"
	"github.com/HomeDing/goding/cmd/midi"
	"github.com/HomeDing/goding/cmd/serve"
	"github.com/HomeDing/goding/internal/common"
	"github.com/HomeDing/goding/internal/elements"
	"github.com/HomeDing/goding/internal/elements/registry"
	"github.com/HomeDing/goding/internal/global"
	"github.com/HomeDing/goding/internal/osd"
)

// TODO: move global variables into a config struct and use a config file for configuration

var activeCommands sync.WaitGroup

var quitChannel = make(chan os.Signal, 1)

var helpMode bool = false

// check if a string is a command.
// Commands are the first non-flag token in the command line arguments.
func isCommand(a string) bool {
	return ((a == "help") ||
		(a == "serve") ||
		(a == "midi") ||
		(a == "list"))
} // isCommand()

// runCommand executes the command with the given arguments.
func runCommand(command string, args []string) {
	slog.Debug("main.doCommand", slog.String("command", command), slog.Any("args", args))

	switch command {
	case "help":
		help.ParseArguments(args)
		help.Run(&activeCommands)

	case "serve":
		// switch on dev mode to enable supporting functionality for development and debugging.
		global.DevFlag = true
		serve.ParseArguments(args)
		serve.Run(&activeCommands)

	case "list":
		midi.ParseArguments(args)
		midi.List()

	case "midi":
		// switch on dev mode to enable supporting functionality for development and debugging.
		global.DevFlag = true
		midi.ParseArguments(args)
		midi.Run(&activeCommands)

	default:
		slog.Error("unknown command: " + command + ". Use 'goding help' for usage information.")
		os.Exit(-1)
	} // switch

} // runCommand()

// Command-line parsing behavior:
// A Command argument can be followed by any other arguments that are handled
// as parameters for that command.
// The next command argument will be treated as a new command and the previous command will be executed with the parameters that were found before it.
// If no command is given, the default command is 'serve'.
func parseAndRun(args []string) {
	var command = ""
	var firstParamIdx = 1
	var lastParamIdx = 0

	slog.Debug("parse", slog.Any("args", args))

	// args[0] is the program started. ignore.
	for argIdx := range args {
		slog.Debug("parse.process", slog.Any("arg", args[argIdx]))

		if !(argIdx == len(args) || isCommand(args[argIdx])) {
			// This argument is a parameter (short case first)
			lastParamIdx = argIdx
			continue
		}

		// First/last: handle command
		if len(command) > 0 {
			runCommand(command, args[firstParamIdx:lastParamIdx+1])
		}

		if argIdx < len(args) {
			// next command
			command = args[argIdx]
			firstParamIdx = argIdx + 1
			lastParamIdx = argIdx
		}
	} // for

	if len(command) > 0 {
		runCommand(command, args[firstParamIdx:lastParamIdx+1])
	}
}

// loadConfig loads the configuration from a JSON file and creates and initializes the Elements.
func loadConfig() {

	var data string = `{
	"Element": { 
	  "e1" : {
			"key1": "value1",
			"key2": "value2"
	  }
	},
	"Volume": {
	  "main" : {	
	  	"min": "0",
			"max": "100",
			"endpoint": "out:con",
			"value": "50"
	  },
	  "traydio" : {	
	  	"min": "0",
			"max": "100",
			"endpoint": "app:trdo",
			"value": "20"
	  },
	  "bee" : {	
	  	"min": "0",
			"max": "100",
			"endpoint": "app:bee",
			"value": "20"
	  },
	  "syssound" : {	
	  	"min": "0",
			"max": "100",
			"endpoint": "app:system",
			"value": "80"
	  }

	},
	"midi": {
	  "K5" : {	
	  	"message": "[14] CC 74",
			"onMessage": "volume/syssound?value=$v"
	  },
	  "K7" : {	
	  	"message": "[14] CC 76",
			"onMessage": "volume/traydio?value=$v"
	  },
	  "K8" : {	
	  	"message": "[14] CC 77",
			"onMessage": "volume/main?value=$v"
	  }
	}
}`

	// os.ReadFile(filename)

	// root object
	var root map[string]any
	slog.Debug("LoadConfig", slog.String("data", data))

	err := json.Unmarshal([]byte(data), &root)
	if err == nil {

		// Iterate groups: "person", "boss", ...
		for elType, elTypeRaw := range root {
			// slog.Debug("LoadConfig.Type", slog.String("Type", elType))

			group, ok := elTypeRaw.(map[string]any)
			if ok {
				// Iterate elements in group
				for elId, elRaw := range group {
					slog.Debug("LoadConfig.create", "key", elements.MakeKey(elType, elId))
					var el common.Element

					// create element of type elType with id elId
					// TODO: implement a generic factory for element creation instead of hardcoding the types here.
					switch strings.ToLower(elType) {
					case "volume":
						el = elements.NewVolumeElement(elId)
					case "midi":
						el = elements.NewMidiElement(elId)
					}
					slog.Debug("LoadConfig.created:", slog.Any("element", el))

					if el != nil {
						// Iterate attributes of element
						single, ok := elRaw.(map[string]any)
						if ok {
							for attr, attrRaw := range single {
								slog.Debug("LoadConfig.set", slog.String("attr", attr), slog.Any("attrRaw", attrRaw))

								el.Set(strings.ToLower(attr), attrRaw.(string))

								// attributes, ok := elRaw.(map[string]any)

								// // Iterate attributes
								// 	slog.Debug("LoadConfig.attributes", slog.Any("elRaw", elRaw))
							} // for
						}
					} // if el != nil
					// elConfig, ok := elRaw.(map[string]any)
				}
			}
		}
	}
}

func initTray() *systray.SystemTray {
	tray := systray.New()

	menu := systray.NewMenu()

	menu.Add("Config...", func() {
		slog.Info("Config clicked!")
	})

	menu.Add("About GoDing...", func() {
		slog.Info("About clicked!")
		tray.ShowNotification("Update Available", "Version 2.0 is ready to install.")
	})

	menu.Add("Quit", func() {
		slog.Info("Quit clicked, removing tray...")
		quitChannel <- syscall.SIGTERM
	})

	// menu.AddSeparator()
	// menu.AddCheckbox("Check me", false, func() { slog.Info("Checkbox toggled") })
	// menu.AddSeparator()

	tray.SetIcon(assets.GodingIcon).
		SetDarkModeIcon(assets.GodingIcon).
		SetTooltip("GoDing!").
		SetMenu(menu)
	// tray.OnClick(func() { slog.Info("Left click!") })
	// tray.OnDoubleClick(func() { slog.Info("Double click!") })
	// tray.OnRightClick(func() { slog.Info("Right click!") })
	tray.Show()
	return tray
} // initTray()

// Main entry point for the GoDing application.
// It initializes the application, parses command-line arguments, and runs the specified commands.
func main() {
	// Set the default log level to warning to reduce verbosity in normal operation.
	slog.SetLogLoggerLevel(slog.LevelInfo)

	// Enable the following lines to get debug output from the start
	slog.SetLogLoggerLevel(slog.LevelDebug)
	global.VerboseFlag = true

	args := os.Args
	slog.Info("GoDing starting...")

	// Start the tray registration and the main window event loop.
	tray := initTray()

	// Start the OnScreenDisplay.
	global.Display, _ = osd.New()
	global.Display.SetTimeout(2)

	// init all commands , don't start them yet, just register them
	help.Init()
	serve.Init()
	midi.Init()

	if len(args) == 1 {
		slog.Info("main no parameters, defaulting to 'midi serve'")
		args = append(args, "midi", "serve")
	}

	if args[1] == "help" {
		helpMode = true
		help.ParseArguments(args[2:])
		help.Run(&activeCommands)

	} else {
		parseAndRun(args[1:])

		if global.VerboseFlag {
			slog.SetLogLoggerLevel(slog.LevelDebug)
		}

		loadConfig()
		registry.StartElements()

		signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)

		// Run a goroutine to listen for signals that will terminate the application.
		// When a signal is received, it will stop all running commands / goroutines and clean up resources before exiting.
		go func() {
			slog.Debug("[Signal] waiting")
			sig := <-quitChannel // Wait for a signal
			slog.Debug("[Signal] Caught signal", slog.Any("sig", sig))

			// Stop all co-routines and clean up resources before exiting.
			help.Stop()
			serve.Stop()
			midi.Stop()
			// And then stop the main window event loop.
			tray.Remove()
			global.Display.Destroy(false)
			// tray.Remove()
		}()

		if err := tray.Run(); err != nil {
			slog.Error("Tray initialization", "err", err)
		}

		activeCommands.Wait() // Block until all workers finish
	}

	if global.VerboseFlag {
		slog.Debug("main.end")
		time.Sleep(time.Second * 1) // Give some time for cleanup before exiting
	}
} // main ()

// End.
