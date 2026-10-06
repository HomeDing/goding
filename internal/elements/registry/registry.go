// registry.go - Element registry for runtime element instances
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// Package registry stores and retrieves element instances at runtime.
package registry

import (
	"regexp"

	"github.com/HomeDing/goding/internal/common"
)

// The element registry stores all configured elements at runtime.
var elementRegistry = map[string]common.Element{}

// Register stores a new element instance by its unique key.
func Register(e common.Element) {
	if e == nil {
		return
	}
	elementRegistry[e.GetKey()] = e
}

// Find retrieves an element from the registry by its unique key.
func Find(t string, id string) common.Element {
	k := t + "/" + id
	return FindByKey(k)
}

// Find retrieves an element from the registry by its unique key.
func FindByKey(key string) common.Element {
	if e, ok := elementRegistry[key]; ok {
		return e
	}
	return nil
}

// FindAll returns all registered elements whose keys match the regexp expression.
func FindAll(match string) []common.Element {
	pattern, err := regexp.Compile(match)
	if err != nil {
		return []common.Element{}
	}

	result := make([]common.Element, 0)
	for key := range elementRegistry {
		if pattern.MatchString(key) {
			result = append(result, elementRegistry[key])
		}
	}
	return result
}

// Start all registered elements by calling their Start() method.
func StartElements() {
	for _, e := range elementRegistry {
		e.Start()
	}
}

// End
