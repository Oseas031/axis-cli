package executor

import (
	"fmt"
	"testing"

	"github.com/axis-cli/axis/internal/types"
)

func TestHumanExecutor_ExecuteCall(t *testing.T) {
	exec := NewHumanExecutor()

	req := &types.HumanCallRequest{
		CallID: "call-1",
		Input:  map[string]any{"message": "test"},
	}

	result, err := exec.ExecuteCall(req)
	if err != nil {
		t.Fatalf("Failed to execute call: %v", err)
	}

	if result.CallID != req.CallID {
		t.Errorf("Expected call ID %s, got %s", req.CallID, result.CallID)
	}

	if result.Status != types.CallStatusPending {
		t.Errorf("Expected status %s, got %s", types.CallStatusPending, result.Status)
	}
}

func TestHumanExecutor_DuplicateCall(t *testing.T) {
	exec := NewHumanExecutor()

	req := &types.HumanCallRequest{
		CallID: "call-1",
		Input:  map[string]any{"message": "test"},
	}

	exec.ExecuteCall(req)

	// Duplicate call should fail
	_, err := exec.ExecuteCall(req)
	if err == nil {
		t.Error("Duplicate call should fail")
	}
}

func TestHumanExecutor_GetCallStatus(t *testing.T) {
	exec := NewHumanExecutor()

	req := &types.HumanCallRequest{
		CallID: "call-1",
		Input:  map[string]any{"message": "test"},
	}

	exec.ExecuteCall(req)

	status := exec.GetCallStatus("call-1")
	if status != types.CallStatusPending {
		t.Errorf("Expected status %s, got %s", types.CallStatusPending, status)
	}

	// Non-existent call should return empty string
	nonExistentStatus := exec.GetCallStatus("non-existent")
	if nonExistentStatus != "" {
		t.Error("Non-existent call should return empty string")
	}
}

func TestHumanExecutor_ResolveCall(t *testing.T) {
	exec := NewHumanExecutor()

	req := &types.HumanCallRequest{
		CallID: "call-1",
		Input:  map[string]any{"message": "test"},
	}

	exec.ExecuteCall(req)

	output := map[string]any{"result": "approved"}
	err := exec.ResolveCall("call-1", output, nil)
	if err != nil {
		t.Fatalf("Failed to resolve call: %v", err)
	}

	status := exec.GetCallStatus("call-1")
	if status != types.CallStatusCompleted {
		t.Errorf("Expected status %s, got %s", types.CallStatusCompleted, status)
	}

	// Resolve non-existent call should fail
	err = exec.ResolveCall("non-existent", output, nil)
	if err == nil {
		t.Error("Resolving non-existent call should fail")
	}
}

func TestHumanExecutor_ResolveCallWithError(t *testing.T) {
	exec := NewHumanExecutor()

	req := &types.HumanCallRequest{
		CallID: "call-1",
		Input:  map[string]any{"message": "test"},
	}

	exec.ExecuteCall(req)

	err := exec.ResolveCall("call-1", nil, fmt.Errorf("human rejected"))
	if err != nil {
		t.Fatalf("Failed to resolve call with error: %v", err)
	}

	status := exec.GetCallStatus("call-1")
	if status != types.CallStatusFailed {
		t.Errorf("Expected status %s, got %s", types.CallStatusFailed, status)
	}
}

func TestHumanExecutor_ResolveNonPendingCall(t *testing.T) {
	exec := NewHumanExecutor()

	req := &types.HumanCallRequest{
		CallID: "call-1",
		Input:  map[string]any{"message": "test"},
	}

	exec.ExecuteCall(req)

	// First resolve
	output := map[string]any{"result": "approved"}
	exec.ResolveCall("call-1", output, nil)

	// Try to resolve again - should fail
	err := exec.ResolveCall("call-1", output, nil)
	if err == nil {
		t.Error("Resolving already resolved call should fail")
	}
}
