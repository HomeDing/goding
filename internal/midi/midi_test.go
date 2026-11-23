// winmm_test.go - Tests for Windows MIDI input device enumeration
//
// This file is part of the OpenSource GoDing project <https://github.com/HomeDing/goding>.
// Copyright (c) 2026-2026 by Matthias Hertel, <http://www.mathertel.de>
// This work is licensed under a BSD style license.
// See <https://github.com/HomeDing/goding/blob/main/LICENSE> for details.

package midi

import "testing"

func TestMatchingInputDevices(t *testing.T) {
	devices := []InputDevice{
		{ID: 0, Name: "LPD8 mk2"},
		{ID: 1, Name: "Virtual MIDI"},
		{ID: 2, Name: "LPD8"},
	}

	tests := []struct {
		name       string
		deviceName string
		wantIDs    []uint32
	}{
		{name: "matches name case insensitively", deviceName: "LpD8", wantIDs: []uint32{0, 2}},
		{name: "wildcard matches every device", deviceName: "*", wantIDs: []uint32{0, 1, 2}},
		{name: "unavailable device has no matches", deviceName: "missing", wantIDs: []uint32{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := matchingInputDevices(devices, test.deviceName)
			if len(got) != len(test.wantIDs) {
				t.Fatalf("matchingInputDevices() returned %d devices, want %d", len(got), len(test.wantIDs))
			}
			for index, device := range got {
				if device.ID != test.wantIDs[index] {
					t.Errorf("matchingInputDevices()[%d].ID = %d, want %d", index, device.ID, test.wantIDs[index])
				}
			}
		})
	}
}

func TestDecodeDeviceName(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "ASCII name terminated with null",
			data: []byte{'L', 0, 'P', 0, 'D', 0, '8', 0, 0, 0, 'x', 0},
			want: "LPD8",
		},
		{
			name: "UTF-16 surrogate pair",
			data: []byte{0x3c, 0xd8, 0xb9, 0xdf, 0, 0},
			want: "🎹",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := decodeDeviceName(test.data); got != test.want {
				t.Errorf("decodeDeviceName() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestDecodeShortMessage(t *testing.T) {
	tests := []struct {
		name      string
		status    byte
		data1     byte
		data2     byte
		want      Message
		wantText  string
		wantValid bool
	}{
		{
			name:      "note on",
			status:    0x92,
			data1:     60,
			data2:     100,
			want:      Message{Type: MessageNoteOn, Channel: 2, Data1: 60, Value: 100},
			wantText:  "[2] ON C5",
			wantValid: true,
		},
		{
			name:      "note off",
			status:    0x82,
			data1:     60,
			data2:     64,
			want:      Message{Type: MessageNoteOff, Channel: 2, Data1: 60},
			wantText:  "[2] OFF C5",
			wantValid: true,
		},
		{
			name:      "zero velocity note on is note off",
			status:    0x90,
			data1:     61,
			want:      Message{Type: MessageNoteOff, Data1: 61},
			wantText:  "[0] OFF Db5",
			wantValid: true,
		},
		{
			name:      "control change",
			status:    0xb3,
			data1:     7,
			data2:     91,
			want:      Message{Type: MessageControlChange, Channel: 3, Data1: 7, Value: 91},
			wantText:  "[3] CC 7",
			wantValid: true,
		},
		{
			name:      "program change",
			status:    0xc4,
			data1:     12,
			want:      Message{Type: MessageProgramChange, Channel: 4, Data1: 12, Value: 12},
			wantText:  "[4] PC 12",
			wantValid: true,
		},
		{
			name:   "unsupported pitch bend",
			status: 0xe0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := uintptr(test.status) | uintptr(test.data1)<<8 | uintptr(test.data2)<<16
			got, valid := decodeShortMessage(data)
			if valid != test.wantValid {
				t.Fatalf("decodeShortMessage() valid = %v, want %v", valid, test.wantValid)
			}
			if got != test.want {
				t.Errorf("decodeShortMessage() = %+v, want %+v", got, test.want)
			}
			if valid && got.String() != test.wantText {
				t.Errorf("Message.String() = %q, want %q", got.String(), test.wantText)
			}
		})
	}
}
