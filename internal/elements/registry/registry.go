// registry.go - Element registry for runtime element instances
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// Package registry stores and retrieves element instances at runtime.
package registry

import "github.com/HomeDing/goding/internal/elements/element"

// The element registry stores all configured elements at runtime.
var elementRegistry = map[string]element.Element{}

// Register stores a new element instance by its unique key.
func Register(e element.Element) {
	if e == nil {
		return
	}
	elementRegistry[e.GetKey()] = e
}

// Find retrieves an element from the registry by its unique key.
func Find(t string, id string) element.Element {
	k := element.MakeKey(t, id)
	return FindByKey(k)
}

// Find retrieves an element from the registry by its unique key.
func FindByKey(key string) element.Element {
	if e, ok := elementRegistry[key]; ok {
		return e
	}
	return nil
}

// End
