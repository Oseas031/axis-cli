package compactor

import (
	"strings"
	"testing"
)

func TestRecoveryContext_Nil(t *testing.T) {
	var rc *RecoveryContext
	if got := rc.BuildRecoveryMessage(); got != "" {
		t.Fatal("nil should return empty")
	}
}

func TestRecoveryContext_Full(t *testing.T) {
	rc := &RecoveryContext{
		ActivePlan: "implement swarm topology",
		Feedback:   "previous attempt missed error handling",
		FileStates: []string{"internal/kernel/swarm/config.go", "internal/agent/executor.go"},
	}
	msg := rc.BuildRecoveryMessage()
	if !strings.Contains(msg, "[compact_boundary]") {
		t.Fatal("missing compact_boundary marker")
	}
	if !strings.Contains(msg, "implement swarm topology") {
		t.Fatal("missing active plan")
	}
	if !strings.Contains(msg, "previous attempt") {
		t.Fatal("missing feedback")
	}
	if !strings.Contains(msg, "config.go") {
		t.Fatal("missing file state")
	}
}

func TestRecoveryContext_Empty(t *testing.T) {
	rc := &RecoveryContext{}
	msg := rc.BuildRecoveryMessage()
	if !strings.Contains(msg, "[compact_boundary]") {
		t.Fatal("should always have boundary marker")
	}
	if strings.Contains(msg, "[active_plan]") {
		t.Fatal("should not have plan section when empty")
	}
}
