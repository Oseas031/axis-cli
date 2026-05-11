package control

import (
	"encoding/json"
	"testing"

	"github.com/axis-cli/axis/internal/types"
)

func TestSubmitTaskRequestJSONRoundTrip(t *testing.T) {
	req := SubmitTaskRequest{
		Task: &types.AgentTask{
			TaskID:     "task-1",
			ContractID: "default",
			Input:      map[string]any{"message": "hello"},
			Status:     types.TaskStatusPending,
			Metadata:   map[string]string{"intent.source": "natural_language"},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal submit request: %v", err)
	}

	var got SubmitTaskRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal submit request: %v", err)
	}
	if got.Task == nil {
		t.Fatal("expected task to round trip")
	}
	if got.Task.TaskID != "task-1" {
		t.Fatalf("expected task ID task-1, got %q", got.Task.TaskID)
	}
	if got.Task.Metadata["intent.source"] != "natural_language" {
		t.Fatalf("expected namespaced metadata to round trip, got %#v", got.Task.Metadata)
	}
}

func TestSubmitTaskResponseJSONRoundTrip(t *testing.T) {
	resp := SubmitTaskResponse{TaskID: "task-1", Status: types.TaskStatusPending}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal submit response: %v", err)
	}

	var got SubmitTaskResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal submit response: %v", err)
	}
	if got.TaskID != "task-1" || got.Status != types.TaskStatusPending {
		t.Fatalf("unexpected submit response: %#v", got)
	}
}

func TestStatusResponseJSONRoundTrip(t *testing.T) {
	resp := StatusResponse{TaskID: "task-1", Status: types.TaskStatusCompleted}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal status response: %v", err)
	}

	var got StatusResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal status response: %v", err)
	}
	if got.TaskID != "task-1" || got.Status != types.TaskStatusCompleted {
		t.Fatalf("unexpected status response: %#v", got)
	}
}

func TestErrorResponseJSONShape(t *testing.T) {
	resp := ErrorResponse{Code: "runtime_not_found", Message: "No local Axis runtime found", Hint: "Start one with: axis start"}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal error response: %v", err)
	}

	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal error response: %v", err)
	}
	if raw["code"] != "runtime_not_found" {
		t.Fatalf("expected stable code field, got %#v", raw)
	}
	if raw["message"] == "" || raw["hint"] == "" {
		t.Fatalf("expected message and hint, got %#v", raw)
	}
}

func TestHealthResponseJSONRoundTrip(t *testing.T) {
	resp := HealthResponse{Status: "ok", Protocol: "http", Address: "127.0.0.1:12345", ProjectRoot: "C:/project"}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal health response: %v", err)
	}

	var got HealthResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal health response: %v", err)
	}
	if got.Status != "ok" || got.Protocol != "http" || got.Address == "" || got.ProjectRoot == "" {
		t.Fatalf("unexpected health response: %#v", got)
	}
}
