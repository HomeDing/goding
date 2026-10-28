package audio

import (
	"testing"
	"unsafe"

	"github.com/deploymenttheory/go-bindings-win32/bindings/win32/system/com/structuredstorage"
	"github.com/deploymenttheory/go-bindings-win32/bindings/win32/system/variant"
	"golang.org/x/sys/windows"
)

type testPropVariantLayout struct {
	vt      uint16
	_       [3]uint16
	payload unsafe.Pointer
}

func TestPropVariantString(t *testing.T) {
	text := "Speakers"
	utf16, err := windows.UTF16FromString(text)
	if err != nil {
		t.Fatalf("UTF16FromString failed: %v", err)
	}
	utf16 = append(utf16, 0)
	pv := structuredstorage.PROPVARIANT{}
	raw := (*testPropVariantLayout)(unsafe.Pointer(&pv))
	raw.vt = uint16(variant.VT_LPWSTR)
	raw.payload = unsafe.Pointer(&utf16[0])

	got, err := propVariantString(&pv)
	if err != nil {
		t.Fatalf("propVariantString returned error: %v", err)
	}
	if got != text {
		t.Fatalf("propVariantString(%q) = %q, want %q", text, got, text)
	}
}
