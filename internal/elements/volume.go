package elements

import (
	"maps"
	// "github.com/HomeDing/goding/internal/elements"
)

type VolumeElement struct {
	Element
	name string
}

func NewVolumeElement(elementId string) VolumeElement {
	var e Element = NewElement("volume", elementId)
	var v VolumeElement = VolumeElement{Element: e, name: "Volume"}

	v.config["min"] = "0"
	v.config["max"] = "100"
	v.config["value"] = "50"
	return v
}

func (e VolumeElement) Set(key, value string) bool {

	var oldValue string
	var ok bool

	oldValue, ok = e.config[key]

	if (ok) && (oldValue != value) {
		e.config[key] = value
	} else {
		return false
	}
	return true
}

func (e VolumeElement) Loop() bool {
	return e.Element.Loop()
}

func (e VolumeElement) State() map[string]string {
	var res = map[string]string{}

	maps.Copy(res, e.Element.State())
	res["name"] = e.name

	return res
}
