package global

// Internal package for global variables and configuration.
// This is not intended to be used by external packages, but rather to hold global state and configuration for the application.
// It can be imported by other internal packages as needed.

// current executed command (e.g. "serve", "client", etc.)
var Command string = ""

// global flag for verbose logging, can be set by command-line arguments
var VerboseFlag bool = false

// global port variable for web server, can be set by command-line arguments
var Port int = 3333

var WebFolder string = "./web"
