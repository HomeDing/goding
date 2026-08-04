// element.go - Core Element type definitions
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// The file provides the shared interface definitions used by element implementations.

package elements

// Element interface defines the common methods that all element types must implement.
type Element interface {
	GetKey() string
	Get(key string) string
	Set(key string, value string) bool
	State() map[string]string
}

type ElementCreator interface {
	New(elType string, elId string) Element
}

// MakeKey generates a unique key for an element based on its type and ID.
func MakeKey(t string, id string) string {
	return t + "/" + id
}
