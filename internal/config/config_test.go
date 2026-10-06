// config_test.go - Tests for configuration JSON loading
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

package config

import (
	"reflect"
	"testing"

	"github.com/HomeDing/goding/internal/elements/registry"
)

func TestUnmarshalConfig(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		expected Config
	}{
		{
			name: "element properties",
			data: `{ "Element": { "e1": { "key1": "value1", "key2": "value2" } } }`,
			expected: Config{ElementTypes: map[string]ElementType{
				"Element": {Elements: map[string]Element{
					"e1": {Properties: map[string]string{"key1": "value1", "key2": "value2"}},
				}},
			}},
		},
		{
			name: "volume",
			data: `{ "Volume": { "main": { "min": "0", "max": "100", "value": "50" } } }`,
			expected: Config{ElementTypes: map[string]ElementType{
				"Volume": {Elements: map[string]Element{
					"main": {Properties: map[string]string{"min": "0", "max": "100", "value": "50"}},
				}},
			}},
		},
		{
			name: "midi",
			data: `{ "midi": { "K5": { "message": "[14] CC 74", "onMessage": "volume/main?value=$v" } } }`,
			expected: Config{ElementTypes: map[string]ElementType{
				"midi": {Elements: map[string]Element{
					"K5": {Properties: map[string]string{"message": "[14] CC 74", "onMessage": "volume/main?value=$v"}},
				}},
			}},
		},
		{
			name: "multiple types and elements",
			data: `{ "Volume": { "main": { "endpoint": "out:con" }, "bee": { "value": "20" } }, "midi": { "K7": { "message": "[14] CC 76" } } }`,
			expected: Config{ElementTypes: map[string]ElementType{
				"Volume": {Elements: map[string]Element{
					"main": {Properties: map[string]string{"endpoint": "out:con"}},
					"bee":  {Properties: map[string]string{"value": "20"}},
				}},
				"midi": {Elements: map[string]Element{
					"K7": {Properties: map[string]string{"message": "[14] CC 76"}},
				}},
			}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			Storage = Config{}
			if err := unmarshalConfig(test.data); err != nil {
				t.Fatalf("unmarshalConfig() returned error: %v", err)
			}
			if !reflect.DeepEqual(Storage, test.expected) {
				t.Fatalf("Storage mismatch:\n got: %#v\nwant: %#v", Storage, test.expected)
			}
			if got := RawJSON(); got != test.data {
				t.Fatalf("RawJSON() = %q, want %q", got, test.data)
			}
		})
	}
}

func TestCreateElementsUsesFactoriesAndRegistersConfiguredElement(t *testing.T) {
	const elementID = "config-factory-test"
	configuration := Config{ElementTypes: map[string]ElementType{
		"Volume": {Elements: map[string]Element{
			elementID: {Properties: map[string]string{"value": "73"}},
		}},
	}}

	configuration.CreateElements()

	element := registry.Find("volume", elementID)
	if element == nil {
		t.Fatal("expected configured element to be registered")
	}
	if got := element.Get("value"); got != "73" {
		t.Fatalf("registered element value = %q, want %q", got, "73")
	}
}

// End.
