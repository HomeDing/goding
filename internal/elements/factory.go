// factory.go - Element constructor registry
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

package elements

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/HomeDing/goding/internal/common"
)

// Factory creates an element instance for a configured ID.
type Factory func(id string) common.Element

var (
	factoryMu sync.RWMutex
	factories = make(map[string]Factory)
)

// RegisterFactory adds an element constructor to the factory registry.
func RegisterFactory(name string, factory Factory) {
	key := strings.ToLower(name)
	if key == "" || factory == nil {
		panic("element factory requires a name and constructor")
	}

	factoryMu.Lock()
	defer factoryMu.Unlock()

	if _, exists := factories[key]; exists {
		panic("duplicate element factory: " + key)
	}
	factories[key] = factory
}

// NewElement creates an element using the registered constructor for typeName.
func NewElement(typeName, id string) (common.Element, error) {
	factoryMu.RLock()
	factory, exists := factories[strings.ToLower(typeName)]
	factoryMu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("unsupported element type %q", typeName)
	}
	return factory(id), nil
}

// FactoryNames returns the registered element type names in sorted order.
func FactoryNames() []string {
	factoryMu.RLock()
	names := make([]string, 0, len(factories))
	for name := range factories {
		names = append(names, name)
	}
	factoryMu.RUnlock()

	sort.Strings(names)
	return names
}
