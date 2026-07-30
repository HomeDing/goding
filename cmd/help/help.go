// help.go - Help command implementation for the GoDing CLI
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// Command implementation to offer high-level help on the command line
package help

// this command package doesn't start a goroutine.
// The WaitGroup is not used therefore.

import (
	"flag"
	"fmt"
	"log/slog"
	"sync"

	"github.com/HomeDing/goding/cmd/midi"
	"github.com/HomeDing/goding/cmd/serve"
	"github.com/HomeDing/goding/internal/global"
)

// Package-level variables for managing the help command state and registered actions.

// helpCommand stores the specific command for which help is requested.
var helpCommand string = ""

// help command parameters
var helpFlags *flag.FlagSet

// Initialize the serve command and its flags in the init function, which is called before main.
// This allows us to set up the command and its flags before parsing the command-line arguments in main.
func Init() {
	slog.Debug("help.Init()")

	// define the flags for the serve command
	helpFlags = flag.NewFlagSet("help", flag.ExitOnError)
	helpFlags.BoolVar(&global.VerboseFlag, "verbose", false, "enable verbose logging")
} // Init()

func Help() {
	slog.Debug("Help.Help()")

	fmt.Fprintln(helpFlags.Output(), `
Usage:

  goding <command> [parameters]

commands are:
  serve          start a web server with the web interface and API for controlling devices.
  midi           start the midi listener to create actions.
  help           print this help information.
  help <command> print help for the specified command.
	`)

	helpFlags.Usage()
} // Help()

// ParseArguments parses the command-line arguments and sets the global variables accordingly.
// It returns a boolean indicating whether the serve command was invoked and an error if there was an issue with parsing the arguments.
func ParseArguments(args []string) (bool, error) {
	slog.Debug("help.ParseArguments()", slog.Any("args", args))

	if len(args) > 0 {
		var s string = args[0]
		if len(s) > 0 && s[0] == '-' {
			// not a command, only parameters
			helpFlags.Parse(args)
		} else {
			helpCommand = s
			helpFlags.Parse(args[1:])
		}
	}
	return true, nil
} // ParseArguments()

func Run(wg *sync.WaitGroup) error {
	slog.Debug("help.Run()", slog.String("command", helpCommand))

	fmt.Fprintln(helpFlags.Output(),
		`
goDing is a Windows service application to use a Windows machine
to execute HomeDing compatible actions triggered by call to a web server and by midi messages.
  `)

	switch helpCommand {
	case "serve":
		serve.Help()

	case "midi":
		midi.Help()

	default:
		Help()
	}
	return nil
} // Run()

func Stop() {
	slog.Debug("help.Stop()")
} // Stop()
