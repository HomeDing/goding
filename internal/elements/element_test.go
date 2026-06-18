package elements

import "testing"

func TestNewElement(t *testing.T) {
	elements = map[string]Element{}

	el := NewElement("lamp", "lamp1")

	if el.Type != "lamp" {
		t.Fatalf("expected Type %q, got %q", "lamp", el.Type)
	}

	if el.Id != "lamp1" {
		t.Fatalf("expected Id %q, got %q", "lamp1", el.Id)
	}

	if got, want := el.key, "lamp/lamp1"; got != want {
		t.Fatalf("expected key %q, got %q", want, got)
	}

	if len(el.config) != 0 {
		t.Fatalf("expected empty config map, got %v", el.config)
	}

	if len(el.values) != 0 {
		t.Fatalf("expected empty values map, got %v", el.values)
	}

	if got, ok := elements["lamp1"]; !ok {
		t.Fatalf("expected elements map to contain key %q", "lamp1")
	} else if got.Id != el.Id || got.Type != el.Type {
		t.Fatalf("expected stored element to match returned element, got %v", got)
	}
}

func TestElementMethods(t *testing.T) {
	el := Element{Id: "test1", Type: "test"}

	if ok := el.Init("new-id"); !ok {
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
