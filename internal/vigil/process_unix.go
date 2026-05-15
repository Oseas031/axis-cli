//go:build !windows

package vigil

import "os"

// processAlive checks if a process with the given PID is still running.
func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Unix: signal 0 checks existence without actually sending a signal.
	return p.Signal(nil) == nil
}
