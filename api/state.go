package api

import (
	"io"
	"log"
	"net/http"
	"time"
)

func HandleStatus(w http.ResponseWriter, r *http.Request) {
	log.Print("handling status request")
	io.WriteString(w, "Server is running at ")
	io.WriteString(w, time.Now().Format(time.RFC3339))
	io.WriteString(w, "\n")
}
