// Package evolution provides data models and storage for the Staged Evolution Protocol.
package evolution

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LedgerFileName is the append-only trace ledger file name.
const LedgerFileName = "steps.jsonl"

// Ledger manages append-only step records for an evolution run.
//
// DESIGN NOTE: Append operations assume a single writer per run directory.
// Concurrent writes from multiple goroutines/processes to the same ledger
// may interleave JSON lines; this is acceptable for P0 because each run is
// owned by exactly one orchestrator instance at a time.
type Ledger struct {
	path string
}

// NewLedger opens or creates a ledger for a run directory.
func NewLedger(runDir string) *Ledger {
	return &Ledger{path: filepath.Join(runDir, LedgerFileName)}
}

// AppendStep appends a step to the ledger.
func (l *Ledger) AppendStep(step EvolutionStep) error {
	data, err := json.Marshal(step)
	if err != nil {
		return fmt.Errorf("marshal step: %w", err)
	}
	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open ledger: %w", err)
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("write step: %w", err)
	}
	if _, err := f.WriteString("\n"); err != nil {
		return fmt.Errorf("write newline: %w", err)
	}
	return nil
}

// ReadSteps reads all steps from the ledger in order.
// Malformed entries are skipped and returned as errors via the error slice.
func (l *Ledger) ReadSteps() ([]EvolutionStep, []error, error) {
	f, err := os.Open(l.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("open ledger: %w", err)
	}
	defer f.Close()

	var steps []EvolutionStep
	var errs []error
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if line == "" {
			continue
		}
		var step EvolutionStep
		if err := json.Unmarshal([]byte(line), &step); err != nil {
			errs = append(errs, fmt.Errorf("line %d: %w", lineNum, err))
			continue
		}
		steps = append(steps, step)
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("scan ledger: %w", err)
	}
	return steps, errs, nil
}

// ReadStepsStrict reads all steps and returns an error if any entry is malformed.
func (l *Ledger) ReadStepsStrict() ([]EvolutionStep, error) {
	steps, errs, err := l.ReadSteps()
	if err != nil {
		return nil, err
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("ledger has %d malformed entries", len(errs))
	}
	return steps, nil
}

// Exists returns true if the ledger file exists.
func (l *Ledger) Exists() bool {
	_, err := os.Stat(l.path)
	return err == nil
}
