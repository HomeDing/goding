// base_test.go - Tests for the shared BaseElement implementation
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

package elements

import "testing"

func TestNewBaseElementInitializesDefaults(t *testing.T) {
	el := NewBaseElement("lamp", "kitchen")

	if got := el.GetKey(); got != "lamp/kitchen" {
		t.Fatalf("expected key %q, got %q", "lamp/kitchen", got)
	}

	if got := el.Get("missing"); got != "" {
		t.Fatalf("expected empty value for missing key, got %q", got)
	}

	if got := el.State(); len(got) != 0 {
		t.Fatalf("expected empty state map, got %v", got)
	}

	if got := el.Loop(); got {
		t.Fatalf("expected Loop to return false")
	}

}

func TestBaseElementSetTracksKnownKeys(t *testing.T) {
	el := NewBaseElement("lamp", "office")
	el.Config["name"] = "initial"

	if ok := el.Set("name", "desk"); !ok {
		t.Fatal("expected Set to return true for a known config key")
	}

	if got := el.Get("name"); got != "desk" {
		t.Fatalf("expected Get to return %q, got %q", "desk", got)
	}

	if ok := el.Set("name", "desk"); ok {
		t.Fatal("expected Set to return false when unchanged")
	}

	el.Values["level"] = "5"
	if ok := el.Set("level", "10"); !ok {
		t.Fatal("expected Set to return true for a known value key")
	}

	if got := el.Get("level"); got != "10" {
		t.Fatalf("expected Get to return %q, got %q", "10", got)
	}
}

func TestBaseElementRejectsUnknownKeys(t *testing.T) {
	el := NewBaseElement("lamp", "hall")

	if ok := el.Set("unknown", "value"); ok {
		t.Fatal("expected Set to return false for unknown keys")
	}

	if got := el.Get("unknown"); got != "" {
		t.Fatalf("expected Get to return empty string for unknown key, got %q", got)
	}
}
