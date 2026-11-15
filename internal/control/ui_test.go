package control

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

type stubRuntime struct{}

func (s stubRuntime) SubmitTask(_ *types.AgentTask) error { return nil }
func (s stubRuntime) GetTaskStatus(_ string) (types.TaskStatus, error) {
	return types.TaskStatusCompleted, nil
}
func (s stubRuntime) GetTaskResult(_ string) (*types.TaskResult, error) { return nil, nil }

func TestServerUI_DisabledByDefault(t *testing.T) {
	srv := NewServer(stubRuntime{}, RuntimeRecord{Protocol: "http", Address: "http://127.0.0.1:1234"})
	h := srv.Handler()

	for _, path := range []string{"/ui/", "/v1/events", "/v1/ui/info", "/v1/events/stream"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("GET %s when UI disabled: got %d, want 404", path, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/tasks", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /v1/tasks when UI disabled: got %d, want 405", rec.Code)
	}
}

func TestServerUI_EnabledServesObservatoryAndEvents(t *testing.T) {
	dir := t.TempDir()
	evLog := NewTaskEventLog(dir)
	_ = evLog.Append(TaskEvent{
		TaskID:    "t-ui-1",
		EventType: "completed",
		Status:    "completed",
		Message:   "ok",
	})

	srv := NewServerWithEventLog(stubRuntime{}, RuntimeRecord{
		Protocol:    "http",
		Address:     "http://127.0.0.1:4321",
		ProjectRoot: dir,
	}, evLog)
	srv.SetUIEnabled(true)
	h := srv.Handler()

	// Check /ui/
	req := httptest.NewRequest(http.MethodGet, "/ui/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /ui/ status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "AXIS // SINGLE-BINARY OBSERVATORY") {
		t.Fatalf("unexpected UI HTML body")
	}

	// Check /v1/events
	evReq := httptest.NewRequest(http.MethodGet, "/v1/events", nil)
	evRec := httptest.NewRecorder()
	h.ServeHTTP(evRec, evReq)
	if evRec.Code != http.StatusOK {
		t.Fatalf("GET /v1/events status = %d, want 200", evRec.Code)
	}
	var payload struct {
		Events []TaskEvent `json:"events"`
	}
	if err := json.Unmarshal(evRec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid JSON from /v1/events: %v", err)
	}
	if len(payload.Events) != 1 || payload.Events[0].TaskID != "t-ui-1" {
		t.Fatalf("unexpected events payload: %+v", payload.Events)
	}
}

type stubListerRuntime struct {
	stubRuntime
	tasks []*types.AgentTask
}

func (s stubListerRuntime) GetAllTasks() []*types.AgentTask {
	return s.tasks
}

func TestServerUI_InfoTasksAndStream(t *testing.T) {
	dir := t.TempDir()
	evLog := NewTaskEventLog(dir)
	_ = evLog.Append(TaskEvent{TaskID: "t-old", EventType: "completed", Status: "completed"})

	rt := stubListerRuntime{
		tasks: []*types.AgentTask{
			{
				TaskID:    "t-mem-1",
				Status:    types.TaskStatusRunning,
				Input:     map[string]any{"prompt": "hello axis"},
				CreatedAt: time.Now().UTC(),
			},
		},
	}
	srv := NewServerWithEventLog(rt, RuntimeRecord{
		Protocol:    "http",
		Address:     "http://127.0.0.1:4321",
		ProjectRoot: dir,
	}, evLog)
	srv.SetUIEnabled(true)
	srv.SetUIInfo(UIInfo{
		Provider:       "openai",
		ProfileName:    "ark-huoshan",
		Model:          "deepseek-v4-1-flash-260910",
		RuntimeProfile: "full",
		HasAPIKey:      true,
	})
	h := srv.Handler()

	// /v1/ui/info
	infoReq := httptest.NewRequest(http.MethodGet, "/v1/ui/info", nil)
	infoRec := httptest.NewRecorder()
	h.ServeHTTP(infoRec, infoReq)
	if infoRec.Code != http.StatusOK {
		t.Fatalf("GET /v1/ui/info status = %d, want 200", infoRec.Code)
	}
	if strings.Contains(strings.ToLower(infoRec.Body.String()), `"api_key":`) {
		t.Fatalf("unexpected api_key field in /v1/ui/info: %s", infoRec.Body.String())
	}
	var info UIInfo
	if err := json.Unmarshal(infoRec.Body.Bytes(), &info); err != nil {
		t.Fatalf("decode /v1/ui/info failed: %v", err)
	}
	if info.ProfileName != "ark-huoshan" || info.Model != "deepseek-v4-1-flash-260910" || !info.HasAPIKey {
		t.Fatalf("unexpected UIInfo: %+v", info)
	}

	// GET /v1/tasks
	tasksReq := httptest.NewRequest(http.MethodGet, "/v1/tasks", nil)
	tasksRec := httptest.NewRecorder()
	h.ServeHTTP(tasksRec, tasksReq)
	if tasksRec.Code != http.StatusOK {
		t.Fatalf("GET /v1/tasks status = %d, want 200", tasksRec.Code)
	}
	var tasksPayload struct {
		Tasks []TaskSummary `json:"tasks"`
	}
	if err := json.Unmarshal(tasksRec.Body.Bytes(), &tasksPayload); err != nil {
		t.Fatalf("decode /v1/tasks failed: %v", err)
	}
	if len(tasksPayload.Tasks) != 2 || tasksPayload.Tasks[0].TaskID != "t-mem-1" || tasksPayload.Tasks[0].Prompt != "hello axis" {
		t.Fatalf("unexpected tasks payload: %+v", tasksPayload.Tasks)
	}

	// GET /v1/events/stream (cancel context immediately after initial ready flush)
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	sseReq := httptest.NewRequest(http.MethodGet, "/v1/events/stream", nil).WithContext(ctx)
	sseRec := httptest.NewRecorder()
	h.ServeHTTP(sseRec, sseReq)
	if sseRec.Code != http.StatusOK {
		t.Fatalf("GET /v1/events/stream status = %d, want 200", sseRec.Code)
	}
	if ct := sseRec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("unexpected Content-Type: %s", ct)
	}
	if !strings.Contains(sseRec.Body.String(), "event: ready") {
		t.Fatalf("expected SSE ready event, got: %s", sseRec.Body.String())
	}
}
