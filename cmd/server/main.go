package main

import (
	"errors"
	"flag"
	"fmt"
	"goding/api"
	verbose "goding/internal/http" // Logging middleware is provided by the internal verbose package.
	"log"
	"log/slog"
	"os"

	"net/http"
)

// TODO: move global variables into a config struct and use a config file for configuration

var verboseMode bool = false
var helpMode = false
var command string = ""

var port int = 3333
var webFolder string = "./web"

func main() {
	var err error = nil

	// parse cmdline parameters using the flag package

	const usage = `Usage of using_flag:
  -v, --verbose  verbose output
  -h, --help     prints help information  
  --port int     port to listen on (default 3333)
  --web path     folder to serve static files from (default "./web")`

	// flag.BoolVar(&verboseMode, "v", false, "enable verbose logging (shorthand)")
	flag.StringVar(&command, "command", "", "command to execute")
	flag.BoolVar(&helpMode, "help", false, "prints help information")
	flag.BoolVar(&verboseMode, "verbose", false, "enable verbose logging")
	flag.IntVar(&port, "port", 3333, "port to listen on")
	flag.StringVar(&webFolder, "web", "./web", "folder to serve static files from")
	flag.Parse()

	if helpMode {
		// goding is tool for enabling remote control of various devices and services through a web interface. It provides a simple API for managing elements such as volumes, lights, and more. The server serves a web interface for controlling these elements and an API for programmatic access.

		// Usage:
		// goding <command> [arguments]

		flag.PrintDefaults()
		os.Exit(0)
	}

	slog.Info("Startup", slog.Any("args", os.Args[1:]))

	mux := http.NewServeMux()

	// establish the FileServer-Handler for static files
	var fileServer http.Handler = http.FileServer(http.Dir(webFolder))

	// TODO: enable embedding the files from /web into the binary using the embed package

	mux.Handle("/", fileServer)

	mux.HandleFunc("GET /api/state", api.HandleStatus)

	mux.Handle("GET /api", http.NotFoundHandler())

	mux.HandleFunc("GET /path/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "got path\n")
	})

	mux.HandleFunc("/state/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "got path\n")
	})

	mux.HandleFunc("/task/{id}/", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		fmt.Fprintf(w, "handling task with id=%v\n", id)
	})

	// err = http.ListenAndServe(":3333", mux)

	chain := verbose.HttpLogging(mux)

	slog.Info("Startup", slog.String("host", "http://localhost:"+fmt.Sprint(port)))

	err = http.ListenAndServe(":"+fmt.Sprint(port), chain)

	if err != nil && errors.Is(err, http.ErrServerClosed) {
		log.Print("closing server")
	} else {
		log.Fatal(err)
	}
}
