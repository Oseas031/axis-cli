package evolution

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewStore_DefaultRoot(t *testing.T) {
	s, err := NewStore("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.root == "" {
		t.Fatal("expected non-empty root")
	}
	if _, err := os.Stat(s.root); os.IsNotExist(err) {
		t.Fatal("expected default root directory to exist")
	}
}

func TestNewStore_CustomRoot(t *testing.T) {
	tmpDir := t.TempDir()
	s, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.root != tmpDir {
		t.Fatalf("expected root %s, got %s", tmpDir, s.root)
	}
}

func TestStore_CreateAndReadRun(t *testing.T) {
	tmpDir := t.TempDir()
	s, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	runID := GenerateRunID()
	intent := EvolutionIntent{
		ID:           "intent-" + runID,
		CreatedAt:    time.Now(),
		Actor:        "test",
		Summary:      "test run",
		TargetDomain: "test",
		RiskLevel:    RiskLow,
		Status:       StatusPending,
	}
	run := EvolutionRun{
		RunID:     runID,
		IntentID:  intent.ID,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}

	if err := s.CreateRun(intent, run); err != nil {
		t.Fatalf("create run failed: %v", err)
	}

	// Verify directory exists
	runDir := s.RunDir(runID)
	if _, err := os.Stat(runDir); os.IsNotExist(err) {
		t.Fatal("expected run directory to exist")
	}

	// Read back run
	readRun, err := s.ReadRun(runID)
	if err != nil {
		t.Fatalf("read run failed: %v", err)
	}
	if readRun.RunID != runID {
		t.Errorf("expected run ID %s, got %s", runID, readRun.RunID)
	}

	// Read back intent
	readIntent, err := s.ReadIntent(runID)
	if err != nil {
		t.Fatalf("read intent failed: %v", err)
	}
	if readIntent.ID != intent.ID {
		t.Errorf("expected intent ID %s, got %s", intent.ID, readIntent.ID)
	}
}

func TestStore_CreateRun_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	s, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	runID := GenerateRunID()
	intent := EvolutionIntent{ID: "intent-" + runID, CreatedAt: time.Now()}
	run := EvolutionRun{RunID: runID, CreatedAt: time.Now()}

	if err := s.CreateRun(intent, run); err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	// Second create should fail
	if err := s.CreateRun(intent, run); err == nil {
		t.Fatal("expected error for duplicate run creation")
	}
}

func TestStore_ReadRun_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	s, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = s.ReadRun("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing run")
	}
}

func TestStore_ListRuns(t *testing.T) {
	tmpDir := t.TempDir()
	s, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Create two runs
	for i := 0; i < 2; i++ {
		runID := GenerateRunID()
		intent := EvolutionIntent{ID: "intent-" + runID, CreatedAt: time.Now()}
		run := EvolutionRun{RunID: runID, CreatedAt: time.Now()}
		if err := s.CreateRun(intent, run); err != nil {
			t.Fatalf("create run failed: %v", err)
		}
	}

	runs, err := s.ListRuns()
	if err != nil {
		t.Fatalf("list runs failed: %v", err)
	}
	if len(runs) != 2 {
		t.Errorf("expected 2 runs, got %d", len(runs))
	}
}

func TestStore_AppendAndReadDecision(t *testing.T) {
	tmpDir := t.TempDir()
	s, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	runID := GenerateRunID()
	intent := EvolutionIntent{ID: "intent-" + runID, CreatedAt: time.Now()}
	run := EvolutionRun{RunID: runID, CreatedAt: time.Now()}
	if err := s.CreateRun(intent, run); err != nil {
		t.Fatalf("create run failed: %v", err)
	}

	decision := EvolutionDecision{
		RunID:     runID,
		Decision:  DecisionPromoted,
		Actor:     "test",
		Reason:    "all tests passed",
		CreatedAt: time.Now(),
	}
	if err := s.AppendDecision(runID, decision); err != nil {
		t.Fatalf("append decision failed: %v", err)
	}

	readDecision, err := s.ReadDecision(runID)
	if err != nil {
		t.Fatalf("read decision failed: %v", err)
	}
	if readDecision.Decision != DecisionPromoted {
		t.Errorf("expected decision promoted, got %s", readDecision.Decision)
	}
}

func TestStore_AppendDecision_ReadsLatest(t *testing.T) {
	tmpDir := t.TempDir()
	s, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	runID := GenerateRunID()
	intent := EvolutionIntent{ID: "intent-" + runID, CreatedAt: time.Now()}
	run := EvolutionRun{RunID: runID, CreatedAt: time.Now()}
	if err := s.CreateRun(intent, run); err != nil {
		t.Fatalf("create run failed: %v", err)
	}

	first := EvolutionDecision{RunID: runID, Decision: DecisionDiscarded, CreatedAt: time.Now()}
	second := EvolutionDecision{RunID: runID, Decision: DecisionPromoted, CreatedAt: time.Now()}
	if err := s.AppendDecision(runID, first); err != nil {
		t.Fatalf("first append failed: %v", err)
	}
	if err := s.AppendDecision(runID, second); err != nil {
		t.Fatalf("second append failed: %v", err)
	}

	latest, err := s.ReadDecision(runID)
	if err != nil {
		t.Fatalf("read decision failed: %v", err)
	}
	if latest.Decision != DecisionPromoted {
		t.Errorf("expected latest decision promoted, got %s", latest.Decision)
	}
}

func TestStore_AppendDecision_RunNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	s, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	decision := EvolutionDecision{
		RunID:    "missing",
		Decision: DecisionPromoted,
	}
	if err := s.AppendDecision("missing", decision); err == nil {
		t.Fatal("expected error for missing run")
	}
}

func TestStore_writeJSON_Atomic(t *testing.T) {
	tmpDir := t.TempDir()
	s := &Store{root: tmpDir}

	path := filepath.Join(tmpDir, "test.json")
	data := map[string]string{"key": "value"}
	if err := s.writeJSON(path, data); err != nil {
		t.Fatalf("writeJSON failed: %v", err)
	}

	// Verify file exists and temp file does not remain
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected file to exist after atomic write")
	}
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("read dir failed: %v", err)
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".tmp" {
			t.Errorf("unexpected temp file remaining: %s", entry.Name())
		}
	}
}
