package elements

import (
	"github.com/HomeDing/goding/internal/elements/registry"
	"github.com/HomeDing/goding/internal/elements/types"
)

// import (
// "maps"
// )

func NewElement(elType string, elId string) *types.Element {
	element := &types.Element{
		Id:     elId,
		Type:   elType,
		Config: map[string]string{},
		Values: map[string]string{},
	}
	registry.Register(element)
	return element
}
