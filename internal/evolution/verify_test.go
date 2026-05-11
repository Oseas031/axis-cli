package evolution

import (
	"runtime"
	"testing"
	"time"
)

func TestVerifier_Run_Success(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("new store failed: %v", err)
	}
	verifier := NewVerifier(store)

	runID := GenerateRunID()
	intent := EvolutionIntent{ID: "intent-" + runID, CreatedAt: time.Now()}
	run := EvolutionRun{RunID: runID, CreatedAt: time.Now()}
	if err := store.CreateRun(intent, run); err != nil {
		t.Fatalf("create run failed: %v", err)
	}

	var command []string
	if runtime.GOOS == "windows" {
		command = []string{"cmd", "/c", "echo hello"}
	} else {
		command = []string{"echo", "hello"}
	}

	record, err := verifier.Run(runID, command, "")
	if err != nil {
		t.Fatalf("verification run failed: %v", err)
	}
	if record.Status != VerificationPassed {
		t.Errorf("expected status passed, got %s", record.Status)
	}
	if record.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", record.ExitCode)
	}
	if record.StdoutRef == "" {
		t.Error("expected stdout ref to be set")
	}
}

func TestVerifier_Run_Failure(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("new store failed: %v", err)
	}
	verifier := NewVerifier(store)

	runID := GenerateRunID()
	intent := EvolutionIntent{ID: "intent-" + runID, CreatedAt: time.Now()}
	run := EvolutionRun{RunID: runID, CreatedAt: time.Now()}
	if err := store.CreateRun(intent, run); err != nil {
		t.Fatalf("create run failed: %v", err)
	}

	// Use a command that will fail
	var command []string
	if runtime.GOOS == "windows" {
		command = []string{"powershell", "-Command", "exit 1"}
	} else {
		command = []string{"false"}
	}

	record, err := verifier.Run(runID, command, "")
	if err == nil {
		t.Fatal("expected error for failed command")
	}
	if record.Status != VerificationFailed {
		t.Errorf("expected status failed, got %s", record.Status)
	}
	if record.ExitCode == 0 {
		t.Error("expected non-zero exit code for failed command")
	}
}

func TestVerifier_Run_EmptyCommand(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("new store failed: %v", err)
	}
	verifier := NewVerifier(store)

	runID := GenerateRunID()
	intent := EvolutionIntent{ID: "intent-" + runID, CreatedAt: time.Now()}
	run := EvolutionRun{RunID: runID, CreatedAt: time.Now()}
	if err := store.CreateRun(intent, run); err != nil {
		t.Fatalf("create run failed: %v", err)
	}

	record, err := verifier.Run(runID, []string{}, "")
	if err == nil {
		t.Fatal("expected error for empty command")
	}
	if record.Status != VerificationFailed {
		t.Errorf("expected status failed, got %s", record.Status)
	}
}

func TestVerifier_Run_RunNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("new store failed: %v", err)
	}
	verifier := NewVerifier(store)

	_, err = verifier.Run("nonexistent", []string{"echo", "hello"}, "")
	if err == nil {
		t.Fatal("expected error for missing run")
	}
}

func TestVerifier_ReadVerification(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("new store failed: %v", err)
	}
	verifier := NewVerifier(store)

	runID := GenerateRunID()
	intent := EvolutionIntent{ID: "intent-" + runID, CreatedAt: time.Now()}
	run := EvolutionRun{RunID: runID, CreatedAt: time.Now()}
	if err := store.CreateRun(intent, run); err != nil {
		t.Fatalf("create run failed: %v", err)
	}

	var command []string
	if runtime.GOOS == "windows" {
		command = []string{"cmd", "/c", "echo hello"}
	} else {
		command = []string{"echo", "hello"}
	}
	_, err = verifier.Run(runID, command, "")
	if err != nil {
		t.Fatalf("verification run failed: %v", err)
	}

	record, err := verifier.ReadVerification(runID)
	if err != nil {
		t.Fatalf("read verification failed: %v", err)
	}
	if record.RunID != runID {
		t.Errorf("expected run ID %s, got %s", runID, record.RunID)
	}
	if record.Status != VerificationPassed {
		t.Errorf("expected status passed, got %s", record.Status)
	}
}

func TestVerifier_ReadVerification_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("new store failed: %v", err)
	}
	verifier := NewVerifier(store)

	_, err = verifier.ReadVerification("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing verification")
	}
}

func TestVerifier_ReadVerificationOutput(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("new store failed: %v", err)
	}
	verifier := NewVerifier(store)

	runID := GenerateRunID()
	intent := EvolutionIntent{ID: "intent-" + runID, CreatedAt: time.Now()}
	run := EvolutionRun{RunID: runID, CreatedAt: time.Now()}
	if err := store.CreateRun(intent, run); err != nil {
		t.Fatalf("create run failed: %v", err)
	}

	var command []string
	if runtime.GOOS == "windows" {
		command = []string{"cmd", "/c", "echo stdout_content && echo stderr_content >&2"}
	} else {
		command = []string{"sh", "-c", "echo stdout_content; echo stderr_content >&2"}
	}
	_, err = verifier.Run(runID, command, "")
	if err != nil {
		t.Fatalf("verification run failed: %v", err)
	}

	stdout, stderr, err := verifier.ReadVerificationOutput(runID)
	if err != nil {
		t.Fatalf("read output failed: %v", err)
	}
	if len(stdout) == 0 {
		t.Error("expected non-empty stdout")
	}
	if len(stderr) == 0 {
		t.Error("expected non-empty stderr")
	}
}

func TestVerifier_ReadVerificationOutput_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("new store failed: %v", err)
	}
	verifier := NewVerifier(store)

	stdout, stderr, err := verifier.ReadVerificationOutput("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stdout) != 0 || len(stderr) != 0 {
		t.Error("expected empty output for missing run")
	}
}
