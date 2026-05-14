package control

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/axis-cli/axis/internal/types"
)

type fakeRuntime struct {
	submitted []*types.AgentTask
	statuses  map[string]types.TaskStatus
	results   map[string]*types.TaskResult
}

func (f *fakeRuntime) SubmitTask(task *types.AgentTask) error {
	f.submitted = append(f.submitted, task)
	return nil
}

func (f *fakeRuntime) GetTaskStatus(taskID string) (types.TaskStatus, error) {
	status, ok := f.statuses[taskID]
	if !ok {
		return "", errors.New("task not found")
	}
	return status, nil
}

func (f *fakeRuntime) GetTaskResult(taskID string) (*types.TaskResult, error) {
	if f.results == nil {
		return nil, nil
	}
	return f.results[taskID], nil
}

func TestControlServerSubmitTask(t *testing.T) {
	runtime := &fakeRuntime{statuses: map[string]types.TaskStatus{}}
	server := NewServer(runtime, RuntimeRecord{Protocol: "http", Address: "127.0.0.1:1234", ProjectRoot: "C:/project"})

	body, err := json.Marshal(SubmitTaskRequest{Task: &types.AgentTask{TaskID: "task-1", ContractID: "default", Input: map[string]any{"message": "hello"}}})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/tasks", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202 accepted, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(runtime.submitted) != 1 || runtime.submitted[0].TaskID != "task-1" {
		t.Fatalf("expected task submitted through runtime, got %#v", runtime.submitted)
	}
	var resp SubmitTaskResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.TaskID != "task-1" || resp.Status != types.TaskStatusPending {
		t.Fatalf("unexpected submit response: %#v", resp)
	}
}

func TestControlServerSubmitMalformedRequest(t *testing.T) {
	server := NewServer(&fakeRuntime{}, RuntimeRecord{})
	req := httptest.NewRequest(http.MethodPost, "/v1/tasks", strings.NewReader("{"))
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	var resp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error response: %v", err)
	}
	if resp.Code != "bad_request" || resp.Message == "" {
		t.Fatalf("expected stable bad_request response, got %#v", resp)
	}
}

func TestControlServerStatus(t *testing.T) {
	server := NewServer(&fakeRuntime{statuses: map[string]types.TaskStatus{"task-1": types.TaskStatusCompleted}}, RuntimeRecord{})
	req := httptest.NewRequest(http.MethodGet, "/v1/tasks/task-1/status", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp StatusResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal status response: %v", err)
	}
	if resp.TaskID != "task-1" || resp.Status != types.TaskStatusCompleted {
		t.Fatalf("unexpected status response: %#v", resp)
	}
}

func TestControlServerStatusWithResult(t *testing.T) {
	rt := &fakeRuntime{
		statuses: map[string]types.TaskStatus{"task-1": types.TaskStatusCompleted},
		results: map[string]*types.TaskResult{
			"task-1": {TaskID: "task-1", Output: map[string]any{"summary": "done"}, Status: types.TaskStatusCompleted},
		},
	}
	server := NewServer(rt, RuntimeRecord{})
	req := httptest.NewRequest(http.MethodGet, "/v1/tasks/task-1/status", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp StatusResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal status response: %v", err)
	}
	if resp.Output["summary"] != "done" {
		t.Fatalf("expected output with summary=done, got %#v", resp.Output)
	}
}

func TestControlServerStatusNotFound(t *testing.T) {
	server := NewServer(&fakeRuntime{statuses: map[string]types.TaskStatus{}}, RuntimeRecord{})
	req := httptest.NewRequest(http.MethodGet, "/v1/tasks/missing/status", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	var resp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error response: %v", err)
	}
	if resp.Code != "task_not_found" {
		t.Fatalf("expected task_not_found, got %#v", resp)
	}
}

func TestControlServerHealthDoesNotExposeSecrets(t *testing.T) {
	server := NewServer(&fakeRuntime{}, RuntimeRecord{Protocol: "http", Address: "127.0.0.1:1234", ProjectRoot: "C:/project"})
	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var raw map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal health response: %v", err)
	}
	for _, forbidden := range []string{"api_key", "token", "secret", "bearer"} {
		if _, ok := raw[forbidden]; ok {
			t.Fatalf("health response must not expose %q: %#v", forbidden, raw)
		}
	}
}

func TestControlServerListenLocalOnly(t *testing.T) {
	server := NewServer(&fakeRuntime{}, RuntimeRecord{})
	listener, err := server.Listen(context.Background(), "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()
	if !strings.HasPrefix(listener.Addr().String(), "127.0.0.1:") {
		t.Fatalf("expected loopback listener, got %s", listener.Addr().String())
	}
}

func TestControlServerListenRejectsNonLoopback(t *testing.T) {
	server := NewServer(&fakeRuntime{}, RuntimeRecord{})
	_, err := server.Listen(context.Background(), "0.0.0.0:8080")
	if err == nil {
		t.Fatal("expected non-loopback address to be rejected")
	}
}
