// config.go - Configuration folder initialization for the GoDing application
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// Package config manages the location used for GoDing configuration files.

package config

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/HomeDing/goding/internal/elements"
	"github.com/HomeDing/goding/internal/elements/registry"
)

// Config stores all configured element types by type name.
type Config struct {
	ElementTypes map[string]ElementType `json:"-"`
}

// ElementType stores configured elements by element ID.
type ElementType struct {
	Elements map[string]Element `json:"-"`
}

// Element stores the string properties of one configured element.
type Element struct {
	Properties map[string]string `json:"-"`
}

// Storage contains the configuration loaded from config.json.
var Storage Config

var rawJSONMu sync.RWMutex
var rawJSON string

// Folder is the directory used for GoDing configuration files.
var Folder string

// Init determines the configuration folder and creates it when necessary.
func Init() {
	configRoot := os.Getenv("LOCALAPPDATA")
	if configRoot == "" {
		configRoot, _ = os.UserConfigDir()
	}

	Folder = filepath.Join(configRoot, "HomeDing")
	slog.Info("config", "folder", Folder)

	if err := os.MkdirAll(Folder, 0755); err != nil {
		slog.Error("failed to create config folder", "folder", Folder, "err", err)
		return
	}

	configFile := filepath.Join(Folder, "config.json")
	file, err := os.OpenFile(configFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if !os.IsExist(err) {
			slog.Error("failed to create config file", "file", configFile, "err", err)
		}
		return
	}

	if _, err := file.WriteString("{}\n"); err != nil {
		slog.Error("failed to initialize config file", "file", configFile, "err", err)
	}
	if err := file.Close(); err != nil {
		slog.Error("failed to close config file", "file", configFile, "err", err)
	}
} // Init()

// Load reads config.json and populates Storage without creating runtime elements.
func Load() (Config, error) {
	configFile := filepath.Join(Folder, "config.json")
	data, err := os.ReadFile(configFile)
	if err != nil {
		return Config{}, err
	}

	if err := unmarshalConfig(string(data)); err != nil {
		return Config{}, err
	}

	slog.Debug("config loaded", "file", configFile)
	return Storage, nil
} // Load()

// RawJSON returns the original JSON string loaded from config.json.
func RawJSON() string {
	rawJSONMu.RLock()
	defer rawJSONMu.RUnlock()
	return rawJSON
}

// unmarshalConfig populates Storage from a three-level JSON configuration string.
func unmarshalConfig(data string) error {
	var raw map[string]map[string]map[string]string
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		return err
	}

	loaded := Config{ElementTypes: make(map[string]ElementType, len(raw))}
	for typeName, elements := range raw {
		elementType := ElementType{Elements: make(map[string]Element, len(elements))}
		for elementID, properties := range elements {
			elementType.Elements[elementID] = Element{Properties: properties}
		}
		loaded.ElementTypes[typeName] = elementType
	}

	Storage = loaded
	rawJSONMu.Lock()
	rawJSON = data
	rawJSONMu.Unlock()
	return nil
} // unmarshalConfig()

// MarshalJSON serializes Config using the three-level configuration shape.
func (c Config) MarshalJSON() ([]byte, error) {
	raw := make(map[string]map[string]map[string]string, len(c.ElementTypes))
	for typeName, elementType := range c.ElementTypes {
		raw[typeName] = make(map[string]map[string]string, len(elementType.Elements))
		for elementID, element := range elementType.Elements {
			raw[typeName][elementID] = element.Properties
		}
	}
	return json.Marshal(raw)
} // MarshalJSON()

// CreateElements creates registered runtime elements from the stored configuration.
func (c Config) CreateElements() {
	for typeName, elementType := range c.ElementTypes {
		for elementID, elementConfig := range elementType.Elements {
			element, err := elements.NewElement(typeName, elementID)
			if err != nil {
				slog.Warn("unsupported config element type", "type", typeName, "id", elementID)
				continue
			}

			for property, value := range elementConfig.Properties {
				element.Set(strings.ToLower(property), value)
			}
			registry.Register(element)
		}
	}
} // CreateElements()

// End.
