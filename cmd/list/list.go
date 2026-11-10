// list.go - List available MIDI input devices
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

// Package list implements the command for listing MIDI input devices.
package list

import (
	"flag"
	"fmt"
	"log/slog"
	"sync"

	"github.com/HomeDing/goding/internal/global"
	midiDevices "github.com/HomeDing/goding/internal/midi"
)

var listFlags *flag.FlagSet

func init() {
	listFlags = flag.NewFlagSet("list", flag.ExitOnError)
	listFlags.BoolVar(&global.VerboseFlag, "verbose", global.VerboseFlag, "enable verbose logging")
}

func Help() {
	slog.Debug("list.Help()")

	fmt.Println()
	fmt.Println(`goDing list lists available MIDI input devices.

Usage:

  goding list [parameters]`)
	fmt.Println()
	listFlags.Usage()
}

// ParseArguments parses the list command's arguments.
func ParseArguments(args []string) (bool, error) {
	slog.Debug("list.ParseArguments()", slog.Any("args", args))
	listFlags.Parse(args)
	return true, nil
}

// Run lists the available MIDI input devices. It does not start a goroutine.
func Run(_ *sync.WaitGroup) error {
	slog.Debug("list.Run()")

	fmt.Println()
	fmt.Println("MIDI inbound ports:")

	devices := midiDevices.ListInputDevices()
	if len(devices) == 0 {
		fmt.Println("(none)")
		return nil
	}

	for _, device := range devices {
		fmt.Printf("%d: '%s'\n", device.ID, device.Name)
	}

	return nil
}

func Stop() {
	slog.Debug("list.Stop()")
}

// End.