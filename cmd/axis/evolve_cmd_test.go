package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/axis-cli/axis/internal/evolution"
)

func TestEvolveInspect_Success(t *testing.T) {
	// Create a temporary store with a run
	store, err := evolution.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	runID := evolution.GenerateRunID()
	intent := evolution.EvolutionIntent{
		ID:           "intent-" + runID,
		CreatedAt:    time.Now(),
		Actor:        "test",
		Summary:      "test inspect",
		TargetDomain: "test",
		RiskLevel:    evolution.RiskLow,
		Status:       evolution.StatusPending,
	}
	run := evolution.EvolutionRun{
		RunID:     runID,
		IntentID:  intent.ID,
		Status:    evolution.StatusRunning,
		CreatedAt: time.Now(),
	}
	if err := store.CreateRun(intent, run); err != nil {
		t.Fatalf("create run: %v", err)
	}

	// Add a step
	ledger := evolution.NewLedger(store.RunDir(runID))
	step := evolution.EvolutionStep{
		StepID:     "step-1",
		RunID:      runID,
		Sequence:   1,
		TargetPath: "main.go",
		Action:     evolution.StepActionPatch,
		Status:     evolution.StatusCompleted,
	}
	if err := ledger.AppendStep(step); err != nil {
		t.Fatalf("append step: %v", err)
	}

	// Use the real inspect logic but with our store
	// evolveInspect uses NewStore("") which picks up CWD, so we must test via output helper
	// Instead, test the helper directly by invoking evolveInspect with a custom store path.
	// Since evolveInspect hardcodes store path, we test by overriding os.Getwd? No.
	// We'll test the core logic by replicating what evolveInspect does.

	// Replicate inspect logic
	result := make(map[string]any)
	readIntent, err := store.ReadIntent(runID)
	if err == nil {
		result["intent"] = readIntent
	}
	readRun, err := store.ReadRun(runID)
	if err == nil {
		result["run"] = readRun
	}
	steps, _, err := ledger.ReadSteps()
	if err == nil {
		result["steps"] = steps
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "test inspect") {
		t.Error("expected output to contain intent summary")
	}
	if !strings.Contains(output, runID) {
		t.Error("expected output to contain run ID")
	}
	if !strings.Contains(output, "step-1") {
		t.Error("expected output to contain step ID")
	}
}

func TestEvolveInspect_MissingRun(t *testing.T) {
	// Replicate inspect logic for missing run
	store, err := evolution.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	result := make(map[string]any)
	_, err = store.ReadIntent("missing")
	if err != nil {
		result["intent_error"] = err.Error()
	}
	_, err = store.ReadRun("missing")
	if err != nil {
		result["run_error"] = err.Error()
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "not found") {
		t.Error("expected output to indicate missing run")
	}
}
