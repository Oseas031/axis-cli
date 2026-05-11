package control

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/axis-cli/axis/internal/types"
)

type Runtime interface {
	SubmitTask(task *types.AgentTask) error
	GetTaskStatus(taskID string) (types.TaskStatus, error)
}

type Server struct {
	runtime Runtime
	record  RuntimeRecord
	events  *TaskEventLog
}

func NewServer(runtime Runtime, record RuntimeRecord) *Server {
	return &Server{runtime: runtime, record: record}
}

func NewServerWithEventLog(runtime Runtime, record RuntimeRecord, events *TaskEventLog) *Server {
	return &Server{runtime: runtime, record: record, events: events}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/health", s.handleHealth)
	mux.HandleFunc("/v1/tasks", s.handleTasks)
	mux.HandleFunc("/v1/tasks/", s.handleTaskStatus)
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
	writeJSON(w, http.StatusOK, StatusResponse{TaskID: taskID, Status: status})
}

func (s *Server) appendEvent(event TaskEvent) {
	if s.events == nil {
		return
	}
	if err := s.events.Append(event); err != nil {
		log.Printf("failed to append task event %s for task %s: %v", event.EventType, event.TaskID, err)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code string, message string, hint string) {
	writeJSON(w, status, ErrorResponse{Code: code, Message: message, Hint: hint})
}
