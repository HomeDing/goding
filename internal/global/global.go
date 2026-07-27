// global.go - Global variables and configuration for the GoDing application
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// Internal package for global variables and configuration.
// This is not intended to be used by external packages,
// but rather to hold global state and configuration for the application.
// It can be imported by other internal packages as needed.
package global

// current executed command (e.g. "serve", "client", etc.)
var Command string = ""

// global flag for verbose logging, can be set by command-line arguments
var VerboseFlag bool = false

// global flag for development mode. This is set when starting from the command line with any command argument.
var DevFlag bool = false

// global port variable for web server, can be set by command-line arguments
var Port int = 3333

// global web folder variable for serving static files, can be set by command-line arguments
var WebFolder string = "./web"
