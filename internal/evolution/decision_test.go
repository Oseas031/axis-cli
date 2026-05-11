package evolution

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDecisionGate_CanPromote_Success(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	gate := NewDecisionGate(store)

	runID := setupRunForPromotion(t, store)
	if err := gate.CanPromote(runID); err != nil {
		t.Fatalf("expected promotion allowed: %v", err)
	}
}

func TestDecisionGate_CanPromote_MissingRun(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	gate := NewDecisionGate(store)

	if err := gate.CanPromote("missing"); err == nil {
		t.Fatal("expected error for missing run")
	}
}

func TestDecisionGate_CanPromote_MissingVerification(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	gate := NewDecisionGate(store)

	runID := GenerateRunID()
	intent := EvolutionIntent{ID: "intent-" + runID, CreatedAt: time.Now()}
	run := EvolutionRun{RunID: runID, CreatedAt: time.Now()}
	if err := store.CreateRun(intent, run); err != nil {
		t.Fatalf("create run: %v", err)
	}
	// No workspace or verification
	if err := gate.CanPromote(runID); err == nil {
		t.Fatal("expected error for missing verification")
	}
}

func TestDecisionGate_CanPromote_FailedVerification(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	gate := NewDecisionGate(store)

	runID := GenerateRunID()
	intent := EvolutionIntent{ID: "intent-" + runID, CreatedAt: time.Now()}
	run := EvolutionRun{RunID: runID, CreatedAt: time.Now()}
	if err := store.CreateRun(intent, run); err != nil {
		t.Fatalf("create run: %v", err)
	}

	// Create workspace
	if _, err := NewWorkspace(store.RunDir(runID), runID); err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	// Write a failed verification record manually
	record := VerificationRecord{
		RunID:    runID,
		Command:  "test",
		Status:   VerificationFailed,
		ExitCode: 1,
	}
	verifier := NewVerifier(store)
	if err := verifier.writeVerificationRecord(store.RunDir(runID), &record); err != nil {
		t.Fatalf("write verification: %v", err)
	}

	if err := gate.CanPromote(runID); err == nil {
		t.Fatal("expected error for failed verification")
	}
}

func TestDecisionGate_Promote_Success(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	gate := NewDecisionGate(store)

	runID := setupRunForPromotion(t, store)

	// Create a file in workspace
	ws, err := NewWorkspace(store.RunDir(runID), runID)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if err := ws.WriteFile("promoted.txt", []byte("promoted content")); err != nil {
		t.Fatalf("write file: %v", err)
	}

	targetDir := t.TempDir()
	decision, err := gate.Promote(runID, targetDir, "test", "tests passed")
	if err != nil {
		t.Fatalf("promote failed: %v", err)
	}
	if decision.Decision != DecisionPromoted {
		t.Errorf("expected promoted, got %s", decision.Decision)
	}

	// Verify file was promoted
	content, err := os.ReadFile(filepath.Join(targetDir, "promoted.txt"))
	if err != nil {
		t.Fatalf("read promoted file: %v", err)
	}
	if string(content) != "promoted content" {
		t.Errorf("expected 'promoted content', got %s", string(content))
	}

	// Verify run status updated
	run, err := store.ReadRun(runID)
	if err != nil {
		t.Fatalf("read run: %v", err)
	}
	if run.Status != StatusPromoted {
		t.Errorf("expected run status promoted, got %s", run.Status)
	}
}

func TestDecisionGate_Promote_Blocked(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	gate := NewDecisionGate(store)

	runID := GenerateRunID()
	intent := EvolutionIntent{ID: "intent-" + runID, CreatedAt: time.Now()}
	run := EvolutionRun{RunID: runID, CreatedAt: time.Now()}
	if err := store.CreateRun(intent, run); err != nil {
		t.Fatalf("create run: %v", err)
	}

	_, err = gate.Promote(runID, t.TempDir(), "test", "")
	if err == nil {
		t.Fatal("expected promotion to be blocked")
	}
}

func TestDecisionGate_Discard_Success(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	gate := NewDecisionGate(store)

	runID := GenerateRunID()
	intent := EvolutionIntent{ID: "intent-" + runID, CreatedAt: time.Now()}
	run := EvolutionRun{RunID: runID, CreatedAt: time.Now()}
	if err := store.CreateRun(intent, run); err != nil {
		t.Fatalf("create run: %v", err)
	}

	decision, err := gate.Discard(runID, "test", "abandoned")
	if err != nil {
		t.Fatalf("discard failed: %v", err)
	}
	if decision.Decision != DecisionDiscarded {
		t.Errorf("expected discarded, got %s", decision.Decision)
	}

	// Verify run status updated
	runPtr, err := store.ReadRun(runID)
	if err != nil {
		t.Fatalf("read run: %v", err)
	}
	if runPtr.Status != StatusDiscarded {
		t.Errorf("expected run status discarded, got %s", runPtr.Status)
	}

	// Verify trace files preserved
	if _, err := os.Stat(store.RunDir(runID)); os.IsNotExist(err) {
		t.Error("expected run directory to be preserved after discard")
	}
}

func TestDecisionGate_Discard_MissingRun(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	gate := NewDecisionGate(store)

	_, err = gate.Discard("missing", "test", "")
	if err == nil {
		t.Fatal("expected error for missing run")
	}
}

func TestDecisionGate_CanPromote_AlreadyDiscarded(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	gate := NewDecisionGate(store)

	runID := setupRunForPromotion(t, store)
	if _, err := gate.Discard(runID, "test", "abandoned"); err != nil {
		t.Fatalf("discard failed: %v", err)
	}

	if err := gate.CanPromote(runID); err == nil {
		t.Fatal("expected promotion blocked for already-discarded run")
	}
}

func TestDecisionGate_Promote_AlreadyPromoted(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	gate := NewDecisionGate(store)

	runID := setupRunForPromotion(t, store)
	ws, _ := NewWorkspace(store.RunDir(runID), runID)
	ws.WriteFile("x.txt", []byte("x"))
	if _, err := gate.Promote(runID, t.TempDir(), "test", "ok"); err != nil {
		t.Fatalf("first promote failed: %v", err)
	}

	_, err = gate.Promote(runID, t.TempDir(), "test", "again")
	if err == nil {
		t.Fatal("expected second promotion to be blocked")
	}
}

func TestDecisionGate_Discard_AlreadyDiscarded(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	gate := NewDecisionGate(store)

	runID := GenerateRunID()
	intent := EvolutionIntent{ID: "intent-" + runID, CreatedAt: time.Now()}
	run := EvolutionRun{RunID: runID, CreatedAt: time.Now()}
	if err := store.CreateRun(intent, run); err != nil {
		t.Fatalf("create run: %v", err)
	}

	if _, err := gate.Discard(runID, "test", "first"); err != nil {
		t.Fatalf("first discard failed: %v", err)
	}

	_, err = gate.Discard(runID, "test", "second")
	if err == nil {
		t.Fatal("expected second discard to be blocked")
	}
}

// setupRunForPromotion creates a run with workspace and passing verification.
func setupRunForPromotion(t *testing.T, store *Store) string {
	runID := GenerateRunID()
	intent := EvolutionIntent{ID: "intent-" + runID, CreatedAt: time.Now()}
	run := EvolutionRun{RunID: runID, CreatedAt: time.Now()}
	if err := store.CreateRun(intent, run); err != nil {
		t.Fatalf("create run: %v", err)
	}

	if _, err := NewWorkspace(store.RunDir(runID), runID); err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	record := VerificationRecord{
		RunID:     runID,
		Command:   "go test ./...",
		Status:    VerificationPassed,
		ExitCode:  0,
		StartedAt: time.Now(),
	}
	verifier := NewVerifier(store)
	if err := verifier.writeVerificationRecord(store.RunDir(runID), &record); err != nil {
		t.Fatalf("write verification: %v", err)
	}
	return runID
}
