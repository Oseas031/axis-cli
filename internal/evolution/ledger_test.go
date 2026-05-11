package evolution

import (
	"fmt"
	"os"
	"testing"
)

func TestLedger_AppendAndReadSteps(t *testing.T) {
	runDir := t.TempDir()
	ledger := NewLedger(runDir)

	steps := []EvolutionStep{
		{StepID: "step-1", RunID: "run-1", Sequence: 1, TargetPath: "a.go", Action: StepActionPatch, Status: StatusCompleted},
		{StepID: "step-2", RunID: "run-1", Sequence: 2, TargetPath: "b.go", Action: StepActionCreate, Status: StatusCompleted},
		{StepID: "step-3", RunID: "run-1", Sequence: 3, TargetPath: "c.go", Action: StepActionDelete, Status: StatusFailed, Error: "not found"},
	}

	for _, step := range steps {
		if err := ledger.AppendStep(step); err != nil {
			t.Fatalf("append step failed: %v", err)
		}
	}

	readSteps, errs, err := ledger.ReadSteps()
	if err != nil {
		t.Fatalf("read steps failed: %v", err)
	}
	if len(errs) > 0 {
		t.Fatalf("unexpected ledger errors: %v", errs)
	}
	if len(readSteps) != len(steps) {
		t.Fatalf("expected %d steps, got %d", len(steps), len(readSteps))
	}
	for i, expected := range steps {
		if readSteps[i].StepID != expected.StepID {
			t.Errorf("step %d: expected StepID %s, got %s", i, expected.StepID, readSteps[i].StepID)
		}
		if readSteps[i].Sequence != expected.Sequence {
			t.Errorf("step %d: expected Sequence %d, got %d", i, expected.Sequence, readSteps[i].Sequence)
		}
		if readSteps[i].Action != expected.Action {
			t.Errorf("step %d: expected Action %s, got %s", i, expected.Action, readSteps[i].Action)
		}
		if readSteps[i].Status != expected.Status {
			t.Errorf("step %d: expected Status %s, got %s", i, expected.Status, readSteps[i].Status)
		}
	}
}

func TestLedger_ReadEmpty(t *testing.T) {
	runDir := t.TempDir()
	ledger := NewLedger(runDir)

	steps, errs, err := ledger.ReadSteps()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(steps) != 0 {
		t.Errorf("expected 0 steps, got %d", len(steps))
	}
	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errs))
	}
}

func TestLedger_ReadMalformedEntries(t *testing.T) {
	runDir := t.TempDir()
	ledger := NewLedger(runDir)

	// Append a valid step
	validStep := EvolutionStep{StepID: "step-1", RunID: "run-1", Sequence: 1, Action: StepActionPatch, Status: StatusCompleted}
	if err := ledger.AppendStep(validStep); err != nil {
		t.Fatalf("append valid step failed: %v", err)
	}

	// Append an invalid JSON line directly to the file
	f, err := openLedgerForAppend(ledger.path)
	if err != nil {
		t.Fatalf("open ledger failed: %v", err)
	}
	if _, err := f.WriteString("this is not json\n"); err != nil {
		t.Fatalf("write malformed line failed: %v", err)
	}
	f.Close()

	steps, errs, err := ledger.ReadSteps()
	if err != nil {
		t.Fatalf("read steps failed: %v", err)
	}
	if len(steps) != 1 {
		t.Errorf("expected 1 valid step, got %d", len(steps))
	}
	if len(errs) != 1 {
		t.Errorf("expected 1 malformed error, got %d", len(errs))
	}
}

func TestLedger_ReadStepsStrict(t *testing.T) {
	runDir := t.TempDir()
	ledger := NewLedger(runDir)

	step := EvolutionStep{StepID: "step-1", RunID: "run-1", Sequence: 1, Action: StepActionPatch, Status: StatusCompleted}
	if err := ledger.AppendStep(step); err != nil {
		t.Fatalf("append step failed: %v", err)
	}

	steps, err := ledger.ReadStepsStrict()
	if err != nil {
		t.Fatalf("read steps strict failed: %v", err)
	}
	if len(steps) != 1 {
		t.Errorf("expected 1 step, got %d", len(steps))
	}
}

func TestLedger_ReadStepsStrict_WithMalformed(t *testing.T) {
	runDir := t.TempDir()
	ledger := NewLedger(runDir)

	validStep := EvolutionStep{StepID: "step-1", RunID: "run-1", Sequence: 1, Action: StepActionPatch, Status: StatusCompleted}
	if err := ledger.AppendStep(validStep); err != nil {
		t.Fatalf("append valid step failed: %v", err)
	}

	f, err := openLedgerForAppend(ledger.path)
	if err != nil {
		t.Fatalf("open ledger failed: %v", err)
	}
	if _, err := f.WriteString("invalid json\n"); err != nil {
		t.Fatalf("write malformed line failed: %v", err)
	}
	f.Close()

	_, err = ledger.ReadStepsStrict()
	if err == nil {
		t.Fatal("expected error for malformed ledger in strict mode")
	}
}

func TestLedger_Exists(t *testing.T) {
	runDir := t.TempDir()
	ledger := NewLedger(runDir)
	if ledger.Exists() {
		t.Error("expected ledger not to exist before any append")
	}

	step := EvolutionStep{StepID: "step-1", RunID: "run-1", Sequence: 1, Action: StepActionPatch, Status: StatusCompleted}
	if err := ledger.AppendStep(step); err != nil {
		t.Fatalf("append step failed: %v", err)
	}
	if !ledger.Exists() {
		t.Error("expected ledger to exist after append")
	}
}

func TestLedger_AppendPreservesOrder(t *testing.T) {
	runDir := t.TempDir()
	ledger := NewLedger(runDir)

	for i := 1; i <= 5; i++ {
		step := EvolutionStep{
			StepID:   fmt.Sprintf("step-%d", i),
			RunID:    "run-1",
			Sequence: i,
			Action:   StepActionPatch,
			Status:   StatusCompleted,
		}
		if err := ledger.AppendStep(step); err != nil {
			t.Fatalf("append step %d failed: %v", i, err)
		}
	}

	steps, _, err := ledger.ReadSteps()
	if err != nil {
		t.Fatalf("read steps failed: %v", err)
	}
	if len(steps) != 5 {
		t.Fatalf("expected 5 steps, got %d", len(steps))
	}
	for i, step := range steps {
		expectedSeq := i + 1
		if step.Sequence != expectedSeq {
			t.Errorf("step %d: expected sequence %d, got %d", i, expectedSeq, step.Sequence)
		}
	}
}

// openLedgerForAppend opens the ledger file for appending raw malformed lines in tests.
func openLedgerForAppend(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
}
