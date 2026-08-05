// registry_test.go - Tests for the element registry
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

package registry

import (
	"testing"

	"github.com/HomeDing/goding/internal/common"
)

func TestAddAndFindUsePointerIdentity(t *testing.T) {
	elementRegistry = map[string]common.Element{}

	el := &testElement{key: "lamp/lamp1"}
	Register(el)

	got := FindByKey(el.GetKey())
	if got == nil {
		t.Fatal("expected element to be found")
	}

	if got != el {
		t.Fatalf("expected registry to preserve pointer identity")
	}
}

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
