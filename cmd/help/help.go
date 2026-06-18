package help

// Help Command Implementation
// This is a command package

// provide high level help information
// and use Help() function in the command packages to send command specific help information

import (
	"flag"
	"fmt"
	"log"
	"log/slog"

	"github.com/HomeDing/goding/cmd/serve"
	"github.com/HomeDing/goding/internal/global"
)

var helpCommand string = ""

// help command parameters
var helpFlags *flag.FlagSet

// Initialize the serve command and its flags in the init function, which is called before main.
// This allows us to set up the command and its flags before parsing the command-line arguments in main.
func Init() {
	// define the flags for the serve command
	helpFlags = flag.NewFlagSet("help", flag.ExitOnError)
	helpFlags.BoolVar(&global.VerboseFlag, "verbose", false, "enable verbose logging")
}

func Help() {

	helpFlags.Usage()
}

// ParseArgs parses the command-line arguments and sets the global variables accordingly.
// It returns a boolean indicating whether the serve command was invoked and an error if there was an issue with parsing the arguments.
func ParseArgs(args []string) (bool, error) {
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

	// 			if len(os.Args) == 2 {
	// 			// global help info
	// 			os.Exit(0)

	// 		}

	// 		if helpcommand == "serve" {
	// 			serve.Help()
	// 		} else if helpcommand == "client" {
	// 			log.Print("client command help info goes here")
	// 			// client.Help()
	// 		}

}

func Run() error {
	slog.Debug("help.Run()", slog.String("command", helpCommand))

	fmt.Fprintln(helpFlags.Output(),
		`goding is tool for enabling remote control of windows devices through a web interface.`)

	switch helpCommand {
	case "serve":
		serve.Help()
		return nil

	case "client":
		log.Print("client command help info goes here")
		return nil

	default:
		if len(helpCommand) > 0 {

		}

		fmt.Fprintln(helpFlags.Output(),
			`  Usage:

    goding <command> [parameters]

  the commands are:
    serve          start a web server with the web interface and API for controlling devices.
    help           print this help information.
    help <command> print help for the specified command.
	`)
	}

	return nil
}
