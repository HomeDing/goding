package elements

import "testing"

func TestNewVolumeElement(t *testing.T) {
	volume := NewVolumeElement("volume", "speaker1")

	if volume.Type != "volume" {
		t.Fatalf("expected Type to be %q, got %q", "volume", volume.Type)
	}

	if volume.Id != "speaker1" {
		t.Fatalf("expected Id to be %q, got %q", "speaker1", volume.Id)
	}

	if got, want := volume.config["min"], "0"; got != want {
		t.Fatalf("expected min config %q, got %q", want, got)
	}

	if got, want := volume.config["max"], "100"; got != want {
		t.Fatalf("expected max config %q, got %q", want, got)
	}

	if got, want := volume.config["value"], "50"; got != want {
		t.Fatalf("expected value config %q, got %q", want, got)
	}
}

func TestVolumeElementSet(t *testing.T) {
	volume := NewVolumeElement("volume", "speaker1")

	if ok := volume.Set("value", "75"); !ok {
		t.Fatal("expected Set to return true when changing an existing value")
	}

	if got := volume.config["value"]; got != "75" {
		t.Fatalf("expected config value %q after Set, got %q", "75", got)
	}

	if ok := volume.Set("value", "75"); ok {
		t.Fatal("expected Set to return false when value is unchanged")
	}

	if ok := volume.Set("unknown", "10"); ok {
		t.Fatal("expected Set to return false for an unknown config key")
	}
}

func TestVolumeElementLoopAndState(t *testing.T) {
	volume := NewVolumeElement("volume", "speaker1")

	if got := volume.Loop(); got {
		t.Fatalf("expected Loop to return false, got %v", got)
	}

	state := volume.State()
	if len(state) != 0 {
		t.Fatalf("expected State to return an empty map, got %v", state)
	}
}
