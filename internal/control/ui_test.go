package control

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

	for _, path := range []string{"/ui/", "/v1/events"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("GET %s when UI disabled: got %d, want 404", path, rec.Code)
		}
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
