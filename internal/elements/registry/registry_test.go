// registry_test.go - Tests for the element registry
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

package registry

import (
	"testing"

	"github.com/HomeDing/goding/internal/elements/types"
)

func TestAddAndFindUsePointerIdentity(t *testing.T) {
	elementRegistry = map[string]*types.Element{}

	el := &types.Element{Type: "lamp", Id: "lamp1", Config: map[string]string{}, Values: map[string]string{}}
	add(el)

	got := FindByKey(el.GetKey())
	if got == nil {
		t.Fatal("expected element to be found")
	}

	if got != el {
		t.Fatalf("expected registry to preserve pointer identity")
	}
}
