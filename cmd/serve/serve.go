// serve.go - Serve command implementation for the GoDing web server
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// Command to start a web server and listen for Actions
package serve

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/HomeDing/goding/internal/elements/registry"
	"github.com/HomeDing/goding/internal/elements/volumeElement"
	"github.com/HomeDing/goding/internal/global"
	"github.com/HomeDing/goding/internal/http/verbose"
)

var isStarted = false
var quitChan = make(chan bool)
var srv *http.Server

// serve command parameters
var serveFlags *flag.FlagSet

// Initialize the serve command and its flags in the init function, which is called before main.
// This allows us to set up the command and its flags before parsing the command-line arguments in main.
func Init() {
	slog.Debug("serve.Init()")

	// define the flags for the serve command
	serveFlags = flag.NewFlagSet("serve", flag.ExitOnError)
	serveFlags.IntVar(&global.Port, "port", 3333, "port to listen on")
	serveFlags.StringVar(&global.WebFolder, "folder", "./web", "folder to serve static files from")
	serveFlags.BoolVar(&global.VerboseFlag, "verbose", false, "enable verbose logging")
}

func Help() {
	slog.Debug("serve.Help()")

	fmt.Fprintln(serveFlags.Output(),
		`
goDing serve starts the local web server receiving actions from network clients.
to execute create HomeDing actions as defined in the configuration.

Usage:

  goding serve [parameters]
	`)

	serveFlags.Usage()
}

// ParseArguments parses the command-line arguments and sets the global variables accordingly.
// It returns a boolean indicating whether the serve command was invoked and an error if there was an issue with parsing the arguments.
func ParseArguments(args []string) (bool, error) {
	slog.Debug("serve.ParseArguments()", slog.Any("args", args))
	serveFlags.Parse(args)
	return true, nil
}

func Run(wg *sync.WaitGroup) error {
	var chain http.Handler

	slog.Debug("serve.Run()")

	// return errors.New("not implemented yet")
	isStarted = true
	wg.Add(1)

	mux := http.NewServeMux()

	// TODO: enable embedding the files from /web into the bi^nary using the embed package

	// fs := GoDingFileSystem{fs: http.Dir(global.WebFolder)}

	// use extended FileServer-Handler for static files
	mux.Handle("/", GoDingFileServer(global.WebFolder))

	// mux.HandleFunc("GET /api/state", api.HandleStatus)

	mux.Handle("GET /api", http.NotFoundHandler())

	// mux.HandleFunc("GET /api/state/", func(w http.ResponseWriter, r *http.Request) {
	// 	slog.Debug("GET /api/state/")
	// 	var v = volumeElement.NewVolumeElement("main")
	// 	var stateMap = map[string]map[string]string{}
	// 	w.Header().Set("Content-Type", "application/json")
	// 	stateMap[v.GetKey()] = v.State()
	// 	json.NewEncoder(w).Encode(stateMap)
	// })

	mux.HandleFunc("GET /api/state/{t}/{id}", func(w http.ResponseWriter, r *http.Request) {
		t := r.PathValue("t")
		id := r.PathValue("id")

		slog.Debug("GET /api/state/{t}/{id}", slog.String("t", t), slog.String("id", id))

		e := registry.Find(t, id)
		if e == nil {
			slog.Warn("element not found", slog.String("t", t), slog.String("id", id))
			http.Error(w, "element not found", http.StatusNotFound)
			return
		}

		q := r.URL.Query()

		if len(q) == 0 {
			// no query parameters, return state as JSON
			v := e.State()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(v)
			return
		}

		// slog.Debug("Query", slog.Any("q", q), slog.Any("q.len", len(q)))

		for key, values := range q {
			if len(values) > 0 {
				value := values[0]
				slog.Debug("Query", slog.String("key", key), slog.String("value", value))
				e.Set(key, value)
			}
		}
	}) // handleFunc 

	if global.DevFlag {
		// enable shutdown endpoint for development and debugging purposes
		mux.HandleFunc("/api/shutdown/", func(w http.ResponseWriter, r *http.Request) {
			slog.Debug("/api/shutdown is called.")
			quitChan <- true
		})
	}

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

	// Create a server instance
	srv = &http.Server{
		Addr:    ":" + fmt.Sprint(global.Port),
		Handler: chain,
	}

	// Read https://www.codestudy.net/blog/how-to-stop-http-listenandserve/

	go func() {
		slog.Debug("serve web server starting ...")
		defer wg.Done() // let main know we server is done
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("serve", slog.Any("serve.error", err))
		}
		slog.Debug("serve web server stopped.")
	}()

	slog.Debug("serve Run() done.")

	return nil
}

func Stop() {
	slog.Debug("serve.Stop()")
	if isStarted && srv != nil {
		slog.Debug("serve Shutdown...")

		// Create a context with timeout to ensure shutdown completes
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// 7. Shutdown the server
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Server shutdown failed: %v", err)
		}
	}
}
