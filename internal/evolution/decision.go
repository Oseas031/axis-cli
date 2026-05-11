// Package evolution provides data models and storage for the Sandboxed Evolution Protocol.
package evolution

import (
	"fmt"
	"os"
	"path/filepath"
)

// DecisionGate enforces explicit promotion/discard rules.
type DecisionGate struct {
	store *Store
}

// NewDecisionGate creates a gate bound to a store.
func NewDecisionGate(store *Store) *DecisionGate {
	return &DecisionGate{store: store}
}

// CanPromote checks whether a run is eligible for promotion.
// Promotion requires a successful verification record, an existing workspace,
// and no prior final decision (promoted or discarded).
func (g *DecisionGate) CanPromote(runID string) error {
	runDir := g.store.RunDir(runID)
	if _, err := os.Stat(runDir); os.IsNotExist(err) {
		return fmt.Errorf("run %s not found", runID)
	}

	existing, err := g.store.ReadDecision(runID)
	if err == nil && (existing.Decision == DecisionPromoted || existing.Decision == DecisionDiscarded) {
		return fmt.Errorf("run %s already has a final decision (%s); promotion blocked", runID, existing.Decision)
	}

	verifier := NewVerifier(g.store)
	record, err := verifier.ReadVerification(runID)
	if err != nil {
		return fmt.Errorf("missing verification: %w", err)
	}
	if record.Status != VerificationPassed {
		return fmt.Errorf("verification failed (status: %s); promotion blocked", record.Status)
	}

	wsPath := filepath.Join(runDir, "workspace")
	if info, err := os.Stat(wsPath); err != nil || !info.IsDir() {
		return fmt.Errorf("workspace missing; promotion blocked")
	}

	return nil
}

// Promote moves workspace files to the target root and records the decision.
// It fails if CanPromote returns an error.
func (g *DecisionGate) Promote(runID string, targetRoot string, actor string, reason string) (*EvolutionDecision, error) {
	if err := g.CanPromote(runID); err != nil {
		return nil, err
	}

	runDir := g.store.RunDir(runID)
	ws, err := NewWorkspace(runDir, runID)
	if err != nil {
		return nil, fmt.Errorf("load workspace: %w", err)
	}

	if err := ws.PromoteTo(targetRoot); err != nil {
		return nil, fmt.Errorf("promote workspace: %w", err)
	}

	decision := EvolutionDecision{
		RunID:    runID,
		Decision: DecisionPromoted,
		Actor:    actor,
		Reason:   reason,
	}
	if err := g.store.AppendDecision(runID, decision); err != nil {
		return nil, fmt.Errorf("write decision: %w", err)
	}

	// Update run status.
	// DESIGN NOTE: This is intentionally non-atomic. The filesystem-native
	// store accepts eventual consistency: decisions.jsonl and run.json are
	// separate append-only facts. A crash between writes leaves the run
	// with an outdated status but a consistent decision record.
	run, err := g.store.ReadRun(runID)
	if err != nil {
		return nil, fmt.Errorf("read run: %w", err)
	}
	run.Status = StatusPromoted
	if err := g.store.writeJSON(filepath.Join(runDir, "run.json"), run); err != nil {
		return nil, fmt.Errorf("update run status: %w", err)
	}

	return &decision, nil
}

// CanDiscard checks whether a run is eligible for discard.
// A run can be discarded only if it exists and has no prior final decision.
func (g *DecisionGate) CanDiscard(runID string) error {
	runDir := g.store.RunDir(runID)
	if _, err := os.Stat(runDir); os.IsNotExist(err) {
		return fmt.Errorf("run %s not found", runID)
	}
	existing, err := g.store.ReadDecision(runID)
	if err == nil && (existing.Decision == DecisionPromoted || existing.Decision == DecisionDiscarded) {
		return fmt.Errorf("run %s already has a final decision (%s); discard blocked", runID, existing.Decision)
	}
	return nil
}

// Discard records a discard decision without modifying the main project tree.
// It preserves all trace files.
func (g *DecisionGate) Discard(runID string, actor string, reason string) (*EvolutionDecision, error) {
	if err := g.CanDiscard(runID); err != nil {
		return nil, err
	}

	decision := EvolutionDecision{
		RunID:    runID,
		Decision: DecisionDiscarded,
		Actor:    actor,
		Reason:   reason,
	}
	if err := g.store.AppendDecision(runID, decision); err != nil {
		return nil, fmt.Errorf("write decision: %w", err)
	}

	// Update run status (see Promote for DESIGN NOTE on non-atomicity).
	runDir := g.store.RunDir(runID)
	run, err := g.store.ReadRun(runID)
	if err != nil {
		return nil, fmt.Errorf("read run: %w", err)
	}
	run.Status = StatusDiscarded
	if err := g.store.writeJSON(filepath.Join(runDir, "run.json"), run); err != nil {
		return nil, fmt.Errorf("update run status: %w", err)
	}

	return &decision, nil
}
