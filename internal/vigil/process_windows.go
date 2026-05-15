//go:build windows

package vigil

import (
	"syscall"
	"unsafe"
)

var (
	modkernel32              = syscall.NewLazyDLL("kernel32.dll")
	procOpenProcess          = modkernel32.NewProc("OpenProcess")
	procGetExitCodeProcess   = modkernel32.NewProc("GetExitCodeProcess")
)

const (
	processQueryLimitedInfo = 0x1000
	stillActive             = 259
)

// processAlive checks if a process with the given PID is still running.
func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	h, _, err := procOpenProcess.Call(processQueryLimitedInfo, 0, uintptr(pid))
	if h == 0 {
		_ = err
		return false
	}
	defer syscall.CloseHandle(syscall.Handle(h))
	var exitCode uint32
	ret, _, _ := procGetExitCodeProcess.Call(h, uintptr(unsafe.Pointer(&exitCode)))
	if ret == 0 {
		return false
	}
	return exitCode == stillActive
}
