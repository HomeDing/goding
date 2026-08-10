// base.go - Shared base implementation for element types
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

package elements

import (
	"log/slog"
)

// Base provides the common behaviors shared by element implementations.
type Base struct {
	// Type of the element, e.g., "Volume", "Temp", etc. will never be changed after creation.
	elementType string

	// ID of the element, e.g., "main", "0", etc. will never be changed after creation.
	elementID string

	// activated state
	isActive bool

	// Config holds the configuration values for the element, which can be set and retrieved.
	Config map[string]string

	// Values holds the current state values of the element, which can be set and retrieved.
	Values map[string]string
}

func NewBaseElement(elType string, elId string) Base {
	this := Base{
		elementType: elType,
		elementID:   elId,
		isActive:    false,
		Config:      map[string]string{},
		Values:      map[string]string{},
	}
	this.Config["active"] = "0"
	return this
}

func (e Base) GetKey() string {
	return MakeKey(e.elementType, e.elementID)
}

func (e Base) IsActive() bool {
	return e.isActive
}

func (e Base) Get(key string) string {
	if value, ok := e.Config[key]; ok {
		return value
	}
	if value, ok := e.Values[key]; ok {
		return value
	}
	return ""
}

// Set stores a configuration or value for the element.
// Returns true if the value was changed, false otherwise.
// This function will only change known configuration or value keys. Unknown keys will be ignored and return false.
// The keys are case-sensitive and must be passed as lowercase.
func (e Base) Set(key, value string) bool {
	if oldValue, ok := e.Config[key]; ok {
		if value != oldValue {
			e.Config[key] = value
			return true
		}
	}

	if oldValue, ok := e.Values[key]; ok {
		if value != oldValue {
			e.Values[key] = value
			return true
		}
	}
	return false
}

func (e Base) Start() {
	slog.Debug("base.start", "element", e.GetKey())
	// Base element does not have any specific start behavior.
	e.isActive = true
	e.Config["active"] = "1"
}

func (e Base) Loop() bool {
	return false
}

func (e Base) State() map[string]string {
	return e.Values
}

// End.
