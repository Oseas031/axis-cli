package control

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

func TestClientSubmitTask(t *testing.T) {
	var got SubmitTaskRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/tasks" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		writeJSON(w, http.StatusAccepted, SubmitTaskResponse{TaskID: got.Task.TaskID, Status: types.TaskStatusPending})
	}))
	defer server.Close()

	client := NewClient(locatorWithRecord(t, RuntimeRecord{Protocol: "http", Address: server.URL, ProjectRoot: t.TempDir(), StartedAt: time.Now().UTC()}), server.Client())
	resp, err := client.SubmitTask(context.Background(), &types.AgentTask{TaskID: "task-1", ContractID: "default", Input: map[string]any{"message": "hello"}})
	if err != nil {
		t.Fatalf("submit task: %v", err)
	}
	if resp.TaskID != "task-1" || resp.Status != types.TaskStatusPending {
		t.Fatalf("unexpected submit response: %#v", resp)
	}
	if got.Task == nil || got.Task.TaskID != "task-1" {
		t.Fatalf("expected task sent to server, got %#v", got)
	}
}

func TestClientStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/tasks/task-1/status" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		writeJSON(w, http.StatusOK, StatusResponse{TaskID: "task-1", Status: types.TaskStatusCompleted})
	}))
	defer server.Close()

	client := NewClient(locatorWithRecord(t, RuntimeRecord{Protocol: "http", Address: server.URL, ProjectRoot: t.TempDir(), StartedAt: time.Now().UTC()}), server.Client())
	resp, err := client.Status(context.Background(), "task-1")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if resp.TaskID != "task-1" || resp.Status != types.TaskStatusCompleted {
		t.Fatalf("unexpected status response: %#v", resp)
	}
}

func TestClientHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/health" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		writeJSON(w, http.StatusOK, HealthResponse{Status: "ok", Protocol: "http", Address: serverURL(r), ProjectRoot: "C:/project"})
	}))
	defer server.Close()

	client := NewClient(locatorWithRecord(t, RuntimeRecord{Protocol: "http", Address: server.URL, ProjectRoot: t.TempDir(), StartedAt: time.Now().UTC()}), server.Client())
	resp, err := client.Health(context.Background())
	if err != nil {
		t.Fatalf("health: %v", err)
	}
	if resp.Status != "ok" || resp.Protocol != "http" || resp.ProjectRoot != "C:/project" {
		t.Fatalf("unexpected health response: %#v", resp)
	}
}

func TestClientNoRuntime(t *testing.T) {
	client := NewClient(NewRuntimeLocator(t.TempDir()), http.DefaultClient)
	_, err := client.Status(context.Background(), "task-1")
	if !errors.Is(err, ErrRuntimeNotFound) {
		t.Fatalf("expected ErrRuntimeNotFound, got %v", err)
	}
	var clientErr *ClientError
	if !errors.As(err, &clientErr) {
		t.Fatalf("expected ClientError, got %T", err)
	}
	if clientErr.Hint == "" {
		t.Fatalf("expected actionable hint, got %#v", clientErr)
	}
}

func TestClientServerErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Code: "task_not_found", Message: "missing"})
	}))
	defer server.Close()

	client := NewClient(locatorWithRecord(t, RuntimeRecord{Protocol: "http", Address: server.URL, ProjectRoot: t.TempDir(), StartedAt: time.Now().UTC()}), server.Client())
	_, err := client.Status(context.Background(), "missing")
	var clientErr *ClientError
	if !errors.As(err, &clientErr) {
		t.Fatalf("expected ClientError, got %T: %v", err, err)
	}
	if clientErr.Code != "task_not_found" {
		t.Fatalf("expected task_not_found, got %#v", clientErr)
	}
}

func TestClientStatusURLEscapesTaskID(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		writeJSON(w, http.StatusOK, StatusResponse{TaskID: "task/1", Status: types.TaskStatusCompleted})
	}))
	defer server.Close()

	client := NewClient(locatorWithRecord(t, RuntimeRecord{Protocol: "http", Address: server.URL, ProjectRoot: t.TempDir(), StartedAt: time.Now().UTC()}), server.Client())
	_, err := client.Status(context.Background(), "task/1")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	want := "/v1/tasks/task%2F1/status"
	if gotPath != want {
		t.Fatalf("expected path %q, got %q", want, gotPath)
	}
}

func TestClientGETDoesNotSendBody(t *testing.T) {
	var bodyLen int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyLen = r.ContentLength
		writeJSON(w, http.StatusOK, StatusResponse{TaskID: "task-1", Status: types.TaskStatusCompleted})
	}))
	defer server.Close()

	client := NewClient(locatorWithRecord(t, RuntimeRecord{Protocol: "http", Address: server.URL, ProjectRoot: t.TempDir(), StartedAt: time.Now().UTC()}), server.Client())
	_, err := client.Status(context.Background(), "task-1")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if bodyLen != 0 {
		t.Fatalf("expected no body on GET, got ContentLength %d", bodyLen)
	}
}

func TestClientUsesTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		writeJSON(w, http.StatusOK, StatusResponse{TaskID: "task-1", Status: types.TaskStatusCompleted})
	}))
	defer server.Close()

	client := NewClient(locatorWithRecord(t, RuntimeRecord{Protocol: "http", Address: server.URL, ProjectRoot: t.TempDir(), StartedAt: time.Now().UTC()}), nil)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	_, err := client.Status(ctx, "task-1")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func locatorWithRecord(t *testing.T, record RuntimeRecord) *RuntimeLocator {
	t.Helper()
	root := t.TempDir()
	record.ProjectRoot = root
	locator := NewRuntimeLocator(root)
	if err := locator.Save(record); err != nil {
		t.Fatalf("save locator: %v", err)
	}
	return locator
}

func serverURL(r *http.Request) string {
	return "http://" + r.Host
}
