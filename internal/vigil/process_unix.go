//go:build !windows

package vigil

import "syscall"

// processAlive checks if a process with the given PID is still running.
func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	// Signal 0 checks existence without actually sending a signal.
	// ESRCH means no such process; EPERM means it exists but we lack permission.
	err := syscall.Kill(pid, 0)
	return err == nil || err == syscall.EPERM
}
