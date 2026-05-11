// Package evolution provides data models and storage for the Sandboxed Evolution Protocol.
package evolution

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Verifier executes verification commands and captures evidence.
type Verifier struct {
	store *Store
}

// NewVerifier creates a new verifier bound to a store.
func NewVerifier(store *Store) *Verifier {
	return &Verifier{store: store}
}

// Run executes a verification command for a run and records the result.
// Output is stored as files in the run directory and referenced in the record.
func (v *Verifier) Run(runID string, command []string, workDir string) (*VerificationRecord, error) {
	runDir := v.store.RunDir(runID)
	if _, err := os.Stat(runDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("run %s not found", runID)
	}

	record := &VerificationRecord{
		RunID:     runID,
		Command:   strings.Join(command, " "),
		StartedAt: time.Now(),
		Status:    VerificationPending,
	}

	// Write initial pending record
	if err := v.writeVerificationRecord(runDir, record); err != nil {
		return nil, fmt.Errorf("write initial record: %w", err)
	}

	// Execute command
	var cmd *exec.Cmd
	if len(command) == 0 {
		record.Status = VerificationFailed
		record.ExitCode = -1
		now := time.Now()
		record.CompletedAt = &now
		_ = v.writeVerificationRecord(runDir, record)
		return record, fmt.Errorf("empty command")
	}

	// P0 intentionally omits command timeouts. Verification commands
	// are expected to be short-lived; hung commands are a known
	// limitation addressed by the caller or in future phases.
	cmd = exec.Command(command[0], command[1:]...)
	if workDir != "" {
		cmd.Dir = workDir
	}

	stdout, stderr, err := runCommand(cmd)
	now := time.Now()
	record.CompletedAt = &now

	// Write output files. Write errors are intentionally ignored:
	// the verification record (not the output files) is the source of truth.
	// Output file absence is detectable via the empty StdoutRef/StderrRef
	// or by inspecting the run directory directly.
	stdoutRef := filepath.Join(runDir, "verification.stdout.txt")
	stderrRef := filepath.Join(runDir, "verification.stderr.txt")
	_ = os.WriteFile(stdoutRef, stdout, 0644)
	_ = os.WriteFile(stderrRef, stderr, 0644)
	record.StdoutRef = stdoutRef
	record.StderrRef = stderrRef

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			record.ExitCode = exitErr.ExitCode()
		} else {
			record.ExitCode = -1
		}
		record.Status = VerificationFailed
	} else {
		record.ExitCode = 0
		record.Status = VerificationPassed
	}

	if writeErr := v.writeVerificationRecord(runDir, record); writeErr != nil {
		return record, fmt.Errorf("write final record: %w", writeErr)
	}

	if record.Status == VerificationFailed {
		return record, fmt.Errorf("verification failed: exit code %d", record.ExitCode)
	}
	return record, nil
}

// ReadVerification reads the verification record for a run.
func (v *Verifier) ReadVerification(runID string) (*VerificationRecord, error) {
	runDir := v.store.RunDir(runID)
	path := filepath.Join(runDir, "verification.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("verification for run %s not found", runID)
		}
		return nil, fmt.Errorf("read verification: %w", err)
	}
	var record VerificationRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("parse verification: %w", err)
	}
	return &record, nil
}

// ReadVerificationOutput reads the stdout and stderr files for a run.
func (v *Verifier) ReadVerificationOutput(runID string) (stdout, stderr []byte, err error) {
	runDir := v.store.RunDir(runID)
	stdoutPath := filepath.Join(runDir, "verification.stdout.txt")
	stderrPath := filepath.Join(runDir, "verification.stderr.txt")

	stdout, err = os.ReadFile(stdoutPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("read stdout: %w", err)
	}
	stderr, err = os.ReadFile(stderrPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("read stderr: %w", err)
	}
	return stdout, stderr, nil
}

func (v *Verifier) writeVerificationRecord(runDir string, record *VerificationRecord) error {
	path := filepath.Join(runDir, "verification.json")
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// runCommand executes a command and captures stdout/stderr separately.
// On Windows, some commands do not return an *exec.ExitError even for non-zero
// exit codes, so we explicitly check ProcessState.ExitCode().
func runCommand(cmd *exec.Cmd) (stdout, stderr []byte, err error) {
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	runErr := cmd.Run()
	stdout = outBuf.Bytes()
	stderr = errBuf.Bytes()
	if runErr != nil {
		return stdout, stderr, runErr
	}
	if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() != 0 {
		return stdout, stderr, fmt.Errorf("exit code %d", cmd.ProcessState.ExitCode())
	}
	return stdout, stderr, nil
}
