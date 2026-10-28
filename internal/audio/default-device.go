// Unfortunately Microsoft never published a public API for changing the default audio
// endpoint. All tools like NirCmd, SoundSwitch, EarTrumpet, etc. use the undocumented
// PolicyConfig COM interface.

package audio

import (
	"fmt"
	"syscall"
	"unsafe"

	win32 "github.com/deploymenttheory/go-bindings-win32/bindings/runtime/win32"
	win32com "github.com/deploymenttheory/go-bindings-win32/bindings/win32/system/com"
)

var (
	clsidPolicyConfigClient = win32.GUID{
		Data1: 0x870af99c,
		Data2: 0x171d,
		Data3: 0x4f9e,
		Data4: [8]byte{0xaf, 0x0d, 0xe6, 0x3d, 0xf4, 0x0c, 0x2b, 0xc9},
	}

	iidPolicyConfig = win32.GUID{
		Data1: 0xf8679f50,
		Data2: 0x850a,
		Data3: 0x41cf,
		Data4: [8]byte{0x9c, 0x72, 0x43, 0x0f, 0x29, 0x02, 0x90, 0xc8},
	}
)

type policyConfigRole uint32

const (
	policyConfigRoleConsole policyConfigRole = iota
	policyConfigRoleMultimedia
	policyConfigRoleCommunications
)

type policyConfig struct {
	win32.IUnknown
}

func (pc *policyConfig) setDefaultEndpoint(
	deviceID *uint16,
	role policyConfigRole,
) error {
	ret, _, _ := syscall.SyscallN(
		pc.LpVtbl[12],
		uintptr(unsafe.Pointer(pc)),
		uintptr(unsafe.Pointer(deviceID)),
		uintptr(role),
	)
	return win32.ErrIfFailed(int32(ret))
}

func SetDefaultAudioDevice(deviceID string) error {
	comCleanup, err := comInit(uint32(win32com.COINIT_APARTMENTTHREADED))
	if err != nil {
		return err
	}
	defer comCleanup()

	var unknown *win32com.IUnknown
	if err := win32com.CoCreateInstance(
		&clsidPolicyConfigClient,
		nil,
		win32com.CLSCTX_ALL,
		&iidPolicyConfig,
		&unknown,
	); err != nil {
		return fmt.Errorf("CoCreateInstance failed: %w", err)
	}
	if unknown == nil {
		return fmt.Errorf("CoCreateInstance returned a nil IPolicyConfig")
	}
	policy := (*policyConfig)(unsafe.Pointer(unknown))
	defer policy.Release()

	id, err := syscall.UTF16PtrFromString(deviceID)
	if err != nil {
		return fmt.Errorf("convert device ID to UTF-16 failed: %w", err)
	}

	roles := []policyConfigRole{
		policyConfigRoleConsole,
		policyConfigRoleMultimedia,
		policyConfigRoleCommunications,
	}
	for _, role := range roles {
		if err := policy.setDefaultEndpoint(id, role); err != nil {
			return fmt.Errorf("SetDefaultEndpoint for role %d failed: %w", role, err)
		}
	}
	return nil
}
