package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/axis-cli/axis/internal/comm"
	"github.com/axis-cli/axis/internal/types"
)

type Runtime interface {
	SubmitTask(task *types.AgentTask) error
	GetTaskStatus(taskID string) (types.TaskStatus, error)
	GetTaskResult(taskID string) (*types.TaskResult, error)
}

// InboxAPI abstracts the human inbox (design.md D7). *human.gate.Gate
// satisfies it; the control plane stays decoupled from the gate package.
type InboxAPI interface {
	List() ([]comm.InboxRecord, int, error)
	RespondCall(callID, choice, text string) (*comm.InboxRecord, error)
}

type Server struct {
	runtime   Runtime
	record    RuntimeRecord
	events    *TaskEventLog
	inbox     InboxAPI
	uiEnabled bool
	uiInfo    UIInfo
}

func NewServer(runtime Runtime, record RuntimeRecord) *Server {
	return &Server{runtime: runtime, record: record}
}

func NewServerWithEventLog(runtime Runtime, record RuntimeRecord, events *TaskEventLog) *Server {
	return &Server{runtime: runtime, record: record, events: events}
}

// SetInboxAPI wires the human inbox endpoints (optional capability).
func (s *Server) SetInboxAPI(inbox InboxAPI) { s.inbox = inbox }

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/health", s.handleHealth)
	mux.HandleFunc("/v1/tasks", s.handleTasks)
	mux.HandleFunc("/v1/tasks/", s.handleTaskStatus)
	mux.HandleFunc("/v1/inbox", s.handleInboxList)
	mux.HandleFunc("/v1/inbox/", s.handleInboxRespond)
	if s.uiEnabled {
		mux.HandleFunc("/ui", s.handleUI)
		mux.HandleFunc("/ui/", s.handleUI)
		mux.HandleFunc("/v1/events", s.handleEventsList)
		mux.HandleFunc("/v1/events/stream", s.handleEventsStream)
		mux.HandleFunc("/v1/ui/info", s.handleUIInfo)
	}
	return mux
}

func (s *Server) Listen(ctx context.Context, address string) (net.Listener, error) {
	if address == "" {
		address = "127.0.0.1:0"
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}
	if !tcpAddr.IP.IsLoopback() {
		return nil, fmt.Errorf("address must be loopback: %s", address)
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()
	return listener, nil
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", "")
		return
	}
	writeJSON(w, http.StatusOK, HealthResponse{Status: "ok", Protocol: s.record.Protocol, Address: s.record.Address, ProjectRoot: s.record.ProjectRoot})
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && s.uiEnabled {
		s.handleTasksList(w, r)
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", "")
		return
	}
	var req SubmitTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid submit request", "")
		return
	}
	if req.Task == nil || req.Task.TaskID == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "task is required", "")
		return
	}
	if err := s.runtime.SubmitTask(req.Task); err != nil {
		writeError(w, http.StatusInternalServerError, "submit_failed", err.Error(), "")
		return
	}
	s.appendEvent(TaskEvent{TaskID: req.Task.TaskID, EventType: "submitted", Actor: "local-control", Status: string(types.TaskStatusPending), Message: "task submitted"})
	writeJSON(w, http.StatusAccepted, SubmitTaskResponse{TaskID: req.Task.TaskID, Status: types.TaskStatusPending})
}

func (s *Server) handleTaskStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", "")
		return
	}
	prefix := "/v1/tasks/"
	if !strings.HasPrefix(r.URL.Path, prefix) || !strings.HasSuffix(r.URL.Path, "/status") {
		writeError(w, http.StatusNotFound, "not_found", "endpoint not found", "")
		return
	}
	taskID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, prefix), "/status")
	if taskID == "" || strings.Contains(taskID, "/") {
		writeError(w, http.StatusNotFound, "not_found", "endpoint not found", "")
		return
	}
	status, err := s.runtime.GetTaskStatus(taskID)
	if err != nil {
		writeError(w, http.StatusNotFound, "task_not_found", err.Error(), "")
		return
	}
	s.appendEvent(TaskEvent{TaskID: taskID, EventType: "status_requested", Actor: "local-control", Status: string(status), Message: "task status requested"})
	resp := StatusResponse{TaskID: taskID, Status: status}
	if status == types.TaskStatusCompleted || status == types.TaskStatusFailed {
		if result, err := s.runtime.GetTaskResult(taskID); err == nil && result != nil {
			resp.Output = result.Output
			resp.Error = result.Error
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) appendEvent(event TaskEvent) {
	if s.events == nil {
		return
	}
	if err := s.events.Append(event); err != nil {
		log.Printf("failed to append task event %s for task %s: %v", event.EventType, event.TaskID, err)
	}
}

// handleInboxList serves GET /v1/inbox — pending and resolved human requests.
func (s *Server) handleInboxList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", "")
		return
	}
	if s.inbox == nil {
		writeError(w, http.StatusNotImplemented, "inbox_unavailable", "human inbox not configured on this runtime", "start a runtime with inbox support")
		return
	}
	records, corrupted, err := s.inbox.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "inbox_read_failed", err.Error(), "retry or inspect .axis/comm/")
		return
	}
	writeJSON(w, http.StatusOK, InboxListResponse{Records: records, CorruptedLines: corrupted})
}

// handleInboxRespond serves POST /v1/inbox/{call-id}/respond.
func (s *Server) handleInboxRespond(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", "")
		return
	}
	if s.inbox == nil {
		writeError(w, http.StatusNotImplemented, "inbox_unavailable", "human inbox not configured on this runtime", "start a runtime with inbox support")
		return
	}
	prefix := "/v1/inbox/"
	suffix := "/respond"
	if !strings.HasPrefix(r.URL.Path, prefix) || !strings.HasSuffix(r.URL.Path, suffix) {
		writeError(w, http.StatusNotFound, "not_found", "endpoint not found", "")
		return
	}
	callID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, prefix), suffix)
	if callID == "" || strings.Contains(callID, "/") {
		writeError(w, http.StatusNotFound, "not_found", "endpoint not found", "")
		return
	}
	var req InboxRespondRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid respond request", `body: {"choice":"B"} or {"text":"..."}`)
		return
	}
	rec, err := s.inbox.RespondCall(callID, req.Choice, req.Text)
	if err != nil {
		switch {
		case errors.Is(err, comm.ErrCallNotFound):
			writeError(w, http.StatusNotFound, "call_not_found", err.Error(), "axis inbox list")
		case errors.Is(err, comm.ErrNotPending):
			writeError(w, http.StatusConflict, "call_not_pending", err.Error(), "axis inbox list --all")
		default:
			writeError(w, http.StatusInternalServerError, "respond_failed", err.Error(), "")
		}
		return
	}
	taskID := ""
	if raw, ok := rec.Message.Payload["request"]; ok {
		if data, merr := json.Marshal(raw); merr == nil {
			var hreq types.HumanCallRequest
			if json.Unmarshal(data, &hreq) == nil {
				taskID = hreq.TaskID
			}
		}
	}
	s.appendEvent(TaskEvent{TaskID: taskID, EventType: "human.request.resolved", Actor: "local-control", Status: rec.Status, Message: "responded via control plane (" + callID + ")"})
	writeJSON(w, http.StatusOK, InboxRespondResponse{CallID: callID, Status: rec.Status})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code string, message string, hint string) {
	writeJSON(w, status, ErrorResponse{Code: code, Message: message, Hint: hint})
}
