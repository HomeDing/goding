package elements

import (
	"maps"

	"github.com/HomeDing/goding/internal/elements/registry"
	"github.com/HomeDing/goding/internal/elements/types"
)

type VolumeElement struct {
	*types.Element
	name string
}

func NewVolumeElement(elementId string) *VolumeElement {
	e := NewElement("volume", elementId)
	v := &VolumeElement{Element: e, name: "Volume"}

	v.Config["min"] = "0"
	v.Config["max"] = "100"
	v.Config["value"] = "50"
	registry.Register(v.Element)
	return v
}

// Overwrite the New function to create a new VolumeElement instance with default configuration values.
func New(elementId string) *VolumeElement {
	return NewVolumeElement(elementId)
}

func (e *VolumeElement) Set(key, value string) bool {
	oldValue, ok := e.Config[key]
	if !ok {
		return false
	}
	if oldValue == value {
		return false
	}
	e.Config[key] = value
	return true
}

func (e *VolumeElement) Loop() bool {
	return false
}

func (e *VolumeElement) State() map[string]string {
	res := map[string]string{}
	maps.Copy(res, e.Config)
	res["name"] = e.name
	return res
}

// register volume element type in the registry
func init() {
	NewVolumeElement("default")
}
