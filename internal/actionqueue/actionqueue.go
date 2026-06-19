package actionQueue

import (
	"sync"
)

// Interner Zustand der Aktionswarteschlange
var (
	// mu schützt die Queue vor gleichzeitigem Zugriff durch mehrere Goroutines
	mu    sync.Mutex
	// queue ist der statische FIFO-Puffer für die Aktions-Strings
	queue []string
)

// Add fügt eine neue Aktion zum statischen FIFO-Puffer hinzu.
// Diese Funktion ist sicher für den parallelen Zugriff durch mehrere Goroutines.
func Add(action string) {
	mu.Lock()
	defer mu.Unlock()
	
	// Das Anhängen an den Slice bewahrt die FIFO-Reihenfolge
	queue = append(queue, action)
}

// GetNext entnimmt die älteste Aktion aus dem Puffer.
// Gibt den Aktions-String und true zurück, oder einen leeren String und false, falls der Puffer leer ist.
func GetNext() (string, bool) {
	mu.Lock()
	defer mu.Unlock()

	if len(queue) == 0 {
		return "", false
	}

	// Das erste Element abrufen
	action := queue
	// Das Element aus dem Slice entfernen
	queue = queue[1:]

	return action, true
}