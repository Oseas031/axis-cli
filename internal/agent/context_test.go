package agent

import (
	"testing"
)

func TestNewSelfContext(t *testing.T) {
	ctx := NewSelfContext("task-1")

	if ctx.TaskID != "task-1" {
		t.Errorf("expected TaskID 'task-1', got '%s'", ctx.TaskID)
	}
	if ctx.AutonomyLevel != AutonomyLevelExecute {
		t.Errorf("expected AutonomyLevelExecute, got %v", ctx.AutonomyLevel)
	}
	if ctx.CompetenceScore != 0.5 {
		t.Errorf("expected CompetenceScore 0.5, got %f", ctx.CompetenceScore)
	}
	if ctx.TaskLineage == nil || len(ctx.TaskLineage) != 0 {
		t.Errorf("expected empty TaskLineage, got %v", ctx.TaskLineage)
	}
	if ctx.CodeSnapshot == nil {
		t.Error("expected CodeSnapshot to be initialized")
	}
	if ctx.DocSnapshot == nil {
		t.Error("expected DocSnapshot to be initialized")
	}
	if ctx.StateSnapshot == nil {
		t.Error("expected StateSnapshot to be initialized")
	}
}

func TestNewCodeSnapshot(t *testing.T) {
	snapshot := NewCodeSnapshot()

	if snapshot.ModifiedFiles == nil || len(snapshot.ModifiedFiles) != 0 {
		t.Errorf("expected empty ModifiedFiles, got %v", snapshot.ModifiedFiles)
	}
	if snapshot.SpecVersion != "" {
		t.Errorf("expected empty SpecVersion, got '%s'", snapshot.SpecVersion)
	}
	if snapshot.TaskCount != 0 {
		t.Errorf("expected TaskCount 0, got %d", snapshot.TaskCount)
	}
	if snapshot.ToolCount != 0 {
		t.Errorf("expected ToolCount 0, got %d", snapshot.ToolCount)
	}
}

func TestNewDocSnapshot(t *testing.T) {
	snapshot := NewDocSnapshot()

	if snapshot.SpecFiles == nil || len(snapshot.SpecFiles) != 0 {
		t.Errorf("expected empty SpecFiles, got %v", snapshot.SpecFiles)
	}
	if snapshot.DocFiles == nil || len(snapshot.DocFiles) != 0 {
		t.Errorf("expected empty DocFiles, got %v", snapshot.DocFiles)
	}
	if snapshot.Reports == nil || len(snapshot.Reports) != 0 {
		t.Errorf("expected empty Reports, got %v", snapshot.Reports)
	}
}

func TestNewStateSnapshot(t *testing.T) {
	snapshot := NewStateSnapshot()

	if snapshot.RunningTasks != 0 {
		t.Errorf("expected RunningTasks 0, got %d", snapshot.RunningTasks)
	}
	if snapshot.PendingTasks != 0 {
		t.Errorf("expected PendingTasks 0, got %d", snapshot.PendingTasks)
	}
	if snapshot.CompletedTasks != 0 {
		t.Errorf("expected CompletedTasks 0, got %d", snapshot.CompletedTasks)
	}
	if snapshot.FailedTasks != 0 {
		t.Errorf("expected FailedTasks 0, got %d", snapshot.FailedTasks)
	}
}

func TestSelfContextClone(t *testing.T) {
	ctx := NewSelfContext("task-1")
	ctx.AddLineage("parent-1")
	ctx.AddLineage("parent-2")
	ctx.CodeSnapshot.SpecVersion = "M5"
	ctx.CodeSnapshot.TaskCount = 5
	ctx.DocSnapshot.SpecFiles = []string{"spec1.md", "spec2.md"}
	ctx.StateSnapshot.RunningTasks = 2

	clone := ctx.Clone()

	// Verify clone is equal but not same reference
	if clone == ctx {
		t.Error("clone should not be the same reference")
	}
	if clone.TaskID != ctx.TaskID {
		t.Errorf("clone TaskID mismatch: expected %s, got %s", ctx.TaskID, clone.TaskID)
	}
	if len(clone.TaskLineage) != len(ctx.TaskLineage) {
		t.Errorf("clone TaskLineage length mismatch")
	}
	if clone.CodeSnapshot == ctx.CodeSnapshot {
		t.Error("clone CodeSnapshot should not be same reference")
	}
	if clone.DocSnapshot == ctx.DocSnapshot {
		t.Error("clone DocSnapshot should not be same reference")
	}
	if clone.StateSnapshot == ctx.StateSnapshot {
		t.Error("clone StateSnapshot should not be same reference")
	}

	// Verify modifying clone doesn't affect original
	clone.TaskLineage = append(clone.TaskLineage, "new-parent")
	if len(ctx.TaskLineage) == len(clone.TaskLineage) {
		t.Error("modifying clone should not affect original")
	}
}

func TestSelfContextCloneNil(t *testing.T) {
	var ctx *SelfContext
	clone := ctx.Clone()

	if clone != nil {
		t.Error("clone of nil should be nil")
	}
}

func TestAddLineage(t *testing.T) {
	ctx := NewSelfContext("task-1")

	ctx.AddLineage("parent-1")
	if len(ctx.TaskLineage) != 1 || ctx.TaskLineage[0] != "parent-1" {
		t.Errorf("unexpected lineage: %v", ctx.TaskLineage)
	}

	ctx.AddLineage("parent-2")
	if len(ctx.TaskLineage) != 2 || ctx.TaskLineage[1] != "parent-2" {
		t.Errorf("unexpected lineage: %v", ctx.TaskLineage)
	}
}

func TestAddLineageToNilSlice(t *testing.T) {
	ctx := &SelfContext{
		TaskID:      "task-1",
		TaskLineage: nil,
	}

	ctx.AddLineage("parent-1")

	if ctx.TaskLineage == nil {
		t.Error("lineage should not be nil after AddLineage")
	}
	if len(ctx.TaskLineage) != 1 {
		t.Errorf("expected 1 lineage, got %d", len(ctx.TaskLineage))
	}
}

func TestSetAutonomyLevel(t *testing.T) {
	ctx := NewSelfContext("task-1")

	ctx.SetAutonomyLevel(AutonomyLevelPlan)
	if ctx.AutonomyLevel != AutonomyLevelPlan {
		t.Errorf("expected AutonomyLevelPlan, got %v", ctx.AutonomyLevel)
	}
}

func TestUpdateCompetenceScore(t *testing.T) {
	tests := []struct {
		score         float64
		expectedLevel AutonomyLevel
		expectedScore float64
	}{
		{0.3, AutonomyLevelExecute, 0.3},
		{0.5, AutonomyLevelDecide, 0.5},
		{0.7, AutonomyLevelPlan, 0.7},
		{0.9, AutonomyLevelLearn, 0.9},
		{1.0, AutonomyLevelLearn, 1.0},
	}

	for _, tt := range tests {
		ctx := NewSelfContext("task-1")
		ctx.UpdateCompetenceScore(tt.score)

		if ctx.CompetenceScore != tt.expectedScore {
			t.Errorf("score %f: expected score %f, got %f", tt.score, tt.expectedScore, ctx.CompetenceScore)
		}
		if ctx.AutonomyLevel != tt.expectedLevel {
			t.Errorf("score %f: expected level %v, got %v", tt.score, tt.expectedLevel, ctx.AutonomyLevel)
		}
	}
}

func TestAutonomyLevelConstants(t *testing.T) {
	if AutonomyLevelExecute != 0 {
		t.Errorf("AutonomyLevelExecute should be 0, got %d", AutonomyLevelExecute)
	}
	if AutonomyLevelDecide != 1 {
		t.Errorf("AutonomyLevelDecide should be 1, got %d", AutonomyLevelDecide)
	}
	if AutonomyLevelPlan != 2 {
		t.Errorf("AutonomyLevelPlan should be 2, got %d", AutonomyLevelPlan)
	}
	if AutonomyLevelLearn != 3 {
		t.Errorf("AutonomyLevelLearn should be 3, got %d", AutonomyLevelLearn)
	}
}
