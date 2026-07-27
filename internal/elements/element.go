// element.go - Element factory helpers
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// Package elements contains implementations of elements used by GoDing.
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
