package serve

// Serve Command Implementation
// This is a command package

// create a web server and register all http Handlers and Functions

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/HomeDing/goding/internal/global"
	"github.com/HomeDing/goding/internal/http/verbose"

	"github.com/HomeDing/goding/internal/elements"
)

// serve command parameters
var serveFlags *flag.FlagSet

// Initialize the serve command and its flags in the init function, which is called before main.
// This allows us to set up the command and its flags before parsing the command-line arguments in main.
func Init() {
	slog.Debug("serve.Init()")

	// define the flags for the serve command
	serveFlags = flag.NewFlagSet("serve", flag.ExitOnError)
	serveFlags.IntVar(&global.Port, "port", 3333, "port to listen on")
	serveFlags.StringVar(&global.WebFolder, "web", "./web", "folder to serve static files from")
	serveFlags.BoolVar(&global.VerboseFlag, "verbose", false, "enable verbose logging")
}

func Help() {
	slog.Debug("serve.Help()")

	fmt.Fprintln(serveFlags.Output(),
		`goding serve [parameters} runs the local web server to receive actions from network clients.
This command can be used to run the application in the background.`)
	serveFlags.Usage()
}

// ParseArgs parses the command-line arguments and sets the global variables accordingly.
// It returns a boolean indicating whether the serve command was invoked and an error if there was an issue with parsing the arguments.
func ParseArgs(args []string) (bool, error) {
	serveFlags.Parse(args)
	return true, nil
}

func Run() error {
	var err error = nil
	var chain http.Handler

	// return errors.New("not implemented yet")

	mux := http.NewServeMux()

	// establish the FileServer-Handler for static files
	var fileServer http.Handler = http.FileServer(http.Dir(global.WebFolder))

	// TODO: enable embedding the files from /web into the binary using the embed package
	mux.Handle("/", fileServer)

	// mux.HandleFunc("GET /api/state", api.HandleStatus)

	mux.Handle("GET /api", http.NotFoundHandler())

	mux.HandleFunc("/api/state/", func(w http.ResponseWriter, r *http.Request) {
		var v = elements.NewVolumeElement("main")

		var stateMap = map[string]map[string]string{}

		w.Header().Set("Content-Type", "application/json")

		stateMap[v.GetKey()] = v.State()

		json.NewEncoder(w).Encode(stateMap)
	})

	mux.HandleFunc("/task/{id}/", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		fmt.Fprintf(w, "handling task with id=%v\n", id)
	})

	if global.VerboseFlag {
		chain = verbose.HttpLogging(mux)
	} else {
		chain = mux
	}

	fmt.Fprintln(serveFlags.Output(), "Starting goding web server on http://localhost:"+fmt.Sprint(global.Port)+"/")

	err = http.ListenAndServe(":"+fmt.Sprint(global.Port), chain)

	if err != nil && errors.Is(err, http.ErrServerClosed) {
		log.Print("closing server")
	} else {
		log.Fatal(err)
	}
	return nil
}
