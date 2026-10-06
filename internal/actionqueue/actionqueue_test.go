// actionqueue_test.go - Tests for the in-memory action queue
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

package actionQueue

import (
	"reflect"
	"testing"
)

func TestRemoveActionsForTarget(t *testing.T) {
	actions := []string{
		"volume/1?value=10",
		"switch/2?value=1",
		"volume/1?value=20",
	}

	got := removeActionsForTarget(actions, "volume/1")
	want := []string{"switch/2?value=1"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("removeActionsForTarget() = %v, want %v", got, want)
	}
}
