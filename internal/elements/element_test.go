package elements

import (
	"testing"

	"github.com/HomeDing/goding/internal/elements/registry"
	"github.com/HomeDing/goding/internal/elements/types"
)

func TestNewElement(t *testing.T) {
	el := NewElement("lamp", "lamp1")
	if el == nil {
		t.Fatal("expected a new element instance")
	}

	elKey := el.GetKey()

	if el.Type != "lamp" {
		t.Fatalf("expected Type %q, got %q", "lamp", el.Type)
	}

	if el.Id != "lamp1" {
		t.Fatalf("expected Id %q, got %q", "lamp1", el.Id)
	}

	if got, want := el.GetKey(), "lamp/lamp1"; got != want {
		t.Fatalf("expected key %q, got %q", want, got)
	}

	if len(el.Config) != 0 {
		t.Fatalf("expected empty config map, got %v", el.Config)
	}

	if len(el.Values) != 0 {
		t.Fatalf("expected empty values map, got %v", el.Values)
	}

	got := registry.Find(elKey)
	if got == nil {
		t.Fatalf("expected registry to contain key %q", el.GetKey())
	} else if got != el {
		t.Fatalf("returned not the original object with %q", el.GetKey())
	} else if got.Id != el.Id || got.Type != el.Type {
		t.Fatalf("expected stored element to match returned element, got %v", got)
	}
}

func TestElementMethods(t *testing.T) {
	el := &types.Element{Id: "test1", Type: "test", Config: map[string]string{}, Values: map[string]string{}}

	if ok := el.Init("test", "new-id"); !ok {
		t.Fatal("expected Init to return true")
	}

	if el.Id != "new-id" {
		t.Fatalf("expected Id to update to %q, got %q", "new-id", el.Id)
	}

	if ok := el.Set("key", "value"); ok {
		t.Fatal("expected Set to return false for base Element")
	}

	if got := el.Loop(); got {
		t.Fatalf("expected Loop to return false, got %v", got)
	}

	state := el.State()
	if len(state) != 0 {
		t.Fatalf("expected State to return an empty map, got %v", state)
	}
}
