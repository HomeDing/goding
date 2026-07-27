// element_type.go - Core Element type definitions
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// Package types provides shared type definitions used across element implementations.
package types

type Element struct {
	Type   string
	Id     string
	Config map[string]string
	Values map[string]string
}

func (e *Element) Init(elType string, elId string) bool {
	e.Type = elType
	e.Id = elId
	return true
}

func (e *Element) GetKey() string {
	return e.Type + "/" + e.Id
}

func (e Element) Set(key, value string) bool {
	if e.Config == nil {
		return false
	}
	e.Config[key] = value
	return false
}

func (e Element) Loop() bool {
	return false
}

func (e Element) State() map[string]string {
	return e.Values
}
