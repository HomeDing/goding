// element_test.go - Tests for the shared element contract
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

package common

import "testing"

type testElement struct {
	key string
}

func (e *testElement) GetKey() string {
	return e.key
}

func (e *testElement) Get(key string) string {
	return ""
}

func (e *testElement) Set(key, value string) bool {
	return false
}

func (e *testElement) State() map[string]string {
	return map[string]string{}
}

// TestElementInterfaceCanBeImplemented verifies that the Element interface
// can be implemented by a custom struct.
func TestElementInterfaceCanBeImplemented(t *testing.T) {
	var el Element = &testElement{key: "demo/item"}
	if got := el.GetKey(); got != "demo/item" {
		t.Fatalf("expected key %q, got %q", "demo/item", got)
	}
}
