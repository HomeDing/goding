package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/HomeDing/goding/internal/global"
	"github.com/HomeDing/goding/cmd/help"
	"github.com/HomeDing/goding/cmd/serve"
)

// TODO: move global variables into a config struct and use a config file for configuration

func main() {

	// Enable the following line to get debug output from the start
	// slog.SetLogLoggerLevel(slog.LevelDebug)

	help.Init()
	serve.Init()

	// parse cmdline parameters using the flag package and sub-commands for different modes of operation (e.g. server, client, etc.)

	if len(os.Args) < 2 {
		global.Command = "serve"
		slog.Info("Startup", "message", "using serve with default parameters...")

	} else {
		global.Command = os.Args[1]
	} // if

	switch global.Command {
	case "help":
		help.ParseArgs(os.Args[2:])

	case "serve":
		if len(os.Args) > 2 {
			serve.ParseArgs(os.Args[2:])
		}

	default:
		log.Fatal("unknown command: " + global.Command + ". Use 'goding help' for usage information.")

	}

	if global.VerboseFlag {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("setting verbose mode.")

	} else {
		slog.SetLogLoggerLevel(slog.LevelWarn)
	}

	slog.Info("Startup", slog.String("command", global.Command), slog.Any("args", os.Args))

	switch global.Command {
	case "help":
		help.Run()

	case "serve":
		serve.Run()

	default:
		log.Fatal("unknown command: " + global.Command + ". Use 'goding help' for usage information.")

	}
}
