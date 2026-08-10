// element.go - Shared contract for runtime elements
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

package common

// Element defines the common methods that runtime element instances must implement.
type Element interface {
	GetKey() string
	IsActive() bool
	Get(key string) string
	Set(key string, value string) bool
	Start()
	State() map[string]string
}

// End.
