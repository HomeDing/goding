// midi.go - Midi Receiver element implementation
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// Package elements contains implementations of UI/control elements used by GoDing.
package elements

import (
	"log/slog"

	"github.com/HomeDing/goding/cmd/midi"
	"github.com/HomeDing/goding/internal/elements/registry"
)

type Midi struct {
	Base
	event  string
	action string
}

// creates a new MidiElement instance with default configuration values and registers it in the registry.
func NewMidiElement(elementId string) *Midi {
	v := &Midi{
		Base: NewBaseElement("midi", elementId)}

	// set initial configuration parameters
	v.Config["message"] = ""
	v.Config["onmessage"] = ""

	registry.Register(v)
	return v
} // NewMidiElement()

// Set overrides the base element setter to validate midi receiver specific values
// and apply them to the system.
func (e *Midi) Set(key, value string) bool {
	slog.Debug("midi.set", "element", e.GetKey(), "key", key, "value", value)

	// call the base Set method to handle known keys
	changed := e.Base.Set(key, value)
	if changed {
		switch key {
		case "message":
			slog.Warn("message specification")

		case "onmessage":
			slog.Warn("onmessage specification")
		}
	}
	return changed
} // Set()

func (e *Midi) Start() {
	slog.Debug("midi.start", "element", e.GetKey())

	// add midi message to the midi receiver
	midi.Register(e.Config["message"], e.Config["onmessage"])

	// Start the base element to ensure any base initialization is performed
	e.Base.Start()
}

// End.
