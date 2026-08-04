// element.go - Element factory helpers
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// Package BaseElement contains the common implementations of elements used by GoDing.
package baseElement

import (
	"github.com/HomeDing/goding/internal/elements"
	"github.com/HomeDing/goding/internal/elements/registry"
)

// The BaseElement struct provides a foundational implementation of the Element interface,
// encapsulating common properties and methods shared by all element types.
type BaseElement struct {
	// Type of the element, e.g., "Volume", "Temp", etc. will never be changed after creation.
	elementType string

	// ID of the element, e.g., "main", "0", etc. will never be changed after creation.
	elementID string

	// Config holds the configuration values for the element, which can be set and retrieved.
	Config map[string]string

	// Values holds the current state values of the element, which can be set and retrieved.
	Values map[string]string
}

func NewBaseElement(elType string, elId string) BaseElement {
	this := BaseElement{
		elementType: elType,
		elementID:   elId,
		Config:      map[string]string{},
		Values:      map[string]string{},
	}
	registry.Register(this)
	return this
}

func (e BaseElement) GetKey() string {
	return elements.MakeKey(e.elementType, e.elementID)
}

func (e BaseElement) Get(key string) string {
	if value, ok := e.Config[key]; ok {
		return value
	}
	if value, ok := e.Values[key]; ok {
		return value
	}
	return ""
}

// Set a configuration or value for the element.
// Returns true if the value was changed, false otherwise.
// This function will only change known configuration or value keys. Unknown keys will be ignored and return false.
func (e BaseElement) Set(key, value string) bool {
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

func (e BaseElement) Loop() bool {
	return false
}

func (e BaseElement) State() map[string]string {
	return e.Values
}
