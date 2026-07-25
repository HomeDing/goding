package registry

import (
	"testing"

	"github.com/HomeDing/goding/internal/elements/types"
)

func TestAddAndFindUsePointerIdentity(t *testing.T) {
	elementRegistry = map[string]*types.Element{}

	el := &types.Element{Type: "lamp", Id: "lamp1", Config: map[string]string{}, Values: map[string]string{}}
	add(el)

	got := find(el.GetKey())
	if got == nil {
		t.Fatal("expected element to be found")
	}

	if got != el {
		t.Fatalf("expected registry to preserve pointer identity")
	}
}
