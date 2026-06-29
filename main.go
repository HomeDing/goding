package main

import (
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"sync"

	"github.com/HomeDing/goding/cmd/help"
	"github.com/HomeDing/goding/cmd/midi"
	"github.com/HomeDing/goding/cmd/serve"
	"github.com/HomeDing/goding/internal/global"
)

// TODO: move global variables into a config struct and use a config file for configuration

var wg sync.WaitGroup

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
		help.ParseArgs(args)
		help.Run(&wg)

	case "serve":
		serve.ParseArgs(args)
		serve.Run(&wg)

	case "list":
		midi.ParseArgs(args)
		midi.List()

	case "midi":
		midi.ParseArgs(args)
		midi.Run(&wg)

	default:
		log.Fatal("unknown command: " + command + ". Use 'goding help' for usage information.")
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

	// os.Args[0] is the program started. ignore.
	for argIdx := 0; argIdx < len(args); argIdx++ {
		slog.Debug("parse.process", slog.Any("arg", args[argIdx]))

		if argIdx == len(args) || isCommand(args[argIdx]) {
			// First last handle command
			if len(command) > 0 {
				runCommand(command, args[firstParamIdx:lastParamIdx+1])
			}

			if argIdx < len(args) {
				// next command
				command = args[argIdx]
				firstParamIdx = argIdx + 1
				lastParamIdx = argIdx
			}

		} else {
			// This argument is a parameter
			lastParamIdx = argIdx
		}
	} // for

	if len(command) > 0 {
		runCommand(command, args[firstParamIdx:lastParamIdx+1])
	}
}

func main() {
	// Enable the following lines to get debug output from the start
	slog.SetLogLoggerLevel(slog.LevelDebug)
	global.VerboseFlag = true

	slog.Debug("main.main()")
	var quitChan = make(chan os.Signal, 1)

	help.Init()
	serve.Init()
	midi.Init()

	parseAndRun(os.Args[1:])

	signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM)

	// Run a goroutine to listen for signals
	go func() {
		slog.Debug("[Signal] waiting")
		sig := <-quitChan // Wait for a signal
		slog.Debug("[Signal] Caught signal", slog.Any("sig", sig))

		help.Stop()
		serve.Stop()
		midi.Stop()
		// os.Exit(0) // Exit after cleanup
	}()

	// if len(os.Args[argIdx]) > 0 && os.Args[argIdx][0] != '-' {
	//   switch command {
	//   case "help":
	//     help.ParseArgs(os.Args[firstParamIdx:argIdx])
	//   case "serve":
	//     serve.ParseArgs(os.Args[firstParamIdx:argIdx])
	//   case "midi", "list":
	//     midi.ParseArgs(os.Args[firstParamIdx:argIdx])
	//   default:
	//     log.Fatal("unknown command: " + command + ". Use 'goding help' for usage information.")
	//   }
	//   commandsToRun = append(commandsToRun, command)
	//   command = os.Args[argIdx]
	//   firstParamIdx = argIdx + 1
	// }

	// switch command {
	// case "help":
	// 	help.ParseArgs(os.Args[firstParamIdx:])
	// case "serve":
	// 	serve.ParseArgs(os.Args[firstParamIdx:])
	// case "midi", "list":
	// 	midi.ParseArgs(os.Args[firstParamIdx:])
	// default:
	// 	log.Fatal("unknown command: " + command + ". Use 'goding help' for usage information.")
	// }
	// commandsToRun = append(commandsToRun, command)
	// global.Command = commandsToRun[0]

	// if global.VerboseFlag {
	// 	slog.SetLogLoggerLevel(slog.LevelDebug)
	// 	slog.Debug("setting verbose mode.")
	// } else {
	// 	slog.SetLogLoggerLevel(slog.LevelWarn)
	// }

	// slog.Info("Startup", slog.String("command", global.Command), slog.Any("args", os.Args))

	// var wg sync.WaitGroup
	// for _, cmd := range commandsToRun {
	// 	switch cmd {
	// 	case "serve":
	// 		wg.Add(1)
	// 		go func() {
	// 			defer wg.Done()
	// 			serve.Run()
	// 		}()
	// 	case "midi":
	// 		wg.Add(1)
	// 		go func() {
	// 			defer wg.Done()
	// 			midi.Run()
	// 		}()
	// 	case "list":
	// 		wg.Add(1)
	// 		go func() {
	// 			defer wg.Done()
	// 			midi.List()
	// 		}()
	// 	}
	// }

	// wg.Wait()

	slog.Debug("main.wait...")
	wg.Wait() // Block until all workers finish
	slog.Debug("main.end")

	// time.Sleep(time.Second * 4)

}
