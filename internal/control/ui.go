package control

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

//go:embed ui/index.html
var observatoryFS embed.FS

// UIInfo holds redacted runtime metadata displayed in the Observatory UI header.
// It must never contain API keys or secret credentials.
type UIInfo struct {
	Provider        string `json:"provider,omitempty"`
	ProfileName     string `json:"profile_name,omitempty"`
	Model           string `json:"model,omitempty"`
	RuntimeProfile  string `json:"runtime_profile,omitempty"`
	FallbackProfile string `json:"fallback_profile,omitempty"`
	HasAPIKey       bool   `json:"has_api_key"`
	ProjectRoot     string `json:"project_root,omitempty"`
}

// TaskLister is an optional extension of Runtime for listing known tasks in the UI.
type TaskLister interface {
	GetAllTasks() []*types.AgentTask
}

// TaskSummary represents a compact task record returned by GET /v1/tasks when UI is enabled.
type TaskSummary struct {
	TaskID      string           `json:"task_id"`
	Status      types.TaskStatus `json:"status"`
	Harness     string           `json:"harness,omitempty"`
	AgentID     string           `json:"agent_id,omitempty"`
	Role        string           `json:"role,omitempty"`
	Tainted     bool             `json:"tainted,omitempty"`
	CostBudget  float64          `json:"cost_budget,omitempty"`
	SpentUSD    float64          `json:"spent_usd,omitempty"`
	Prompt      string           `json:"prompt,omitempty"`
	Input       map[string]any   `json:"input,omitempty"`
	CreatedAt   time.Time        `json:"created_at,omitempty"`
	CompletedAt *time.Time       `json:"completed_at,omitempty"`
}

// SetUIEnabled toggles the built-in Observatory UI routes (/ui/, /v1/events, /v1/events/stream, /v1/ui/info).
func (s *Server) SetUIEnabled(enabled bool) {
	s.uiEnabled = enabled
}

// SetUIInfo sets redacted runtime metadata exposed via GET /v1/ui/info.
func (s *Server) SetUIInfo(info UIInfo) {
	s.uiInfo = info
}

func (s *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", "")
		return
	}
	if r.URL.Path == "/ui" {
		http.Redirect(w, r, "/ui/", http.StatusFound)
		return
	}
	data, err := observatoryFS.ReadFile("ui/index.html")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ui_load_failed", err.Error(), "")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (s *Server) handleUIInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", "")
		return
	}
	info := s.uiInfo
	if info.ProjectRoot == "" {
		info.ProjectRoot = s.record.ProjectRoot
	}
	writeJSON(w, http.StatusOK, info)
}

func taskSummaryFromAgentTask(t *types.AgentTask, fallbackHarness string) TaskSummary {
	harness := fallbackHarness
	agentID := ""
	role := ""
	tainted := false
	if t.Metadata != nil {
		if h := t.Metadata["harness"]; h != "" {
			harness = h
		}
		agentID = t.Metadata["agent_id"]
		role = t.Metadata["role"]
		tainted = t.Metadata["tainted"] == "true"
	}
	if agentID == "" {
		agentID = harness
	}
	return TaskSummary{
		TaskID:      t.TaskID,
		Status:      t.Status,
		Harness:     harness,
		AgentID:     agentID,
		Role:        role,
		Tainted:     tainted,
		CostBudget:  parseTaskBudgetUSD(t),
		SpentUSD:    parseTaskSpentUSD(t),
		Prompt:      extractTaskPrompt(t.Input),
		Input:       t.Input,
		CreatedAt:   t.CreatedAt,
		CompletedAt: t.CompletedAt,
	}
}

func (s *Server) handleTasksList(w http.ResponseWriter, r *http.Request) {
	byID := make(map[string]TaskSummary)

	if lister, ok := s.runtime.(TaskLister); ok {
		for _, t := range lister.GetAllTasks() {
			if t == nil || t.TaskID == "" {
				continue
			}
			byID[t.TaskID] = taskSummaryFromAgentTask(t, defaultHarnessID)
		}
	}

	for _, t := range s.ensureHarnesses().GetAllTasks() {
		if t == nil || t.TaskID == "" {
			continue
		}
		byID[t.TaskID] = taskSummaryFromAgentTask(t, defaultHarnessID)
	}

	// Merge event log records so tasks from earlier turns/restarts remain visible if absent from memory.
	for _, ev := range s.readAllEvents() {
		if ev.TaskID == "" {
			continue
		}
		if existing, ok := byID[ev.TaskID]; ok {
			if existing.CreatedAt.IsZero() && !ev.CreatedAt.IsZero() {
				existing.CreatedAt = ev.CreatedAt
			}
			if existing.AgentID == "" && ev.AgentID != "" {
				existing.AgentID = ev.AgentID
			}
			if existing.Harness == "" && ev.AgentID != "" {
				existing.Harness = ev.AgentID
			}
			byID[ev.TaskID] = existing
			continue
		}
		harness := ev.AgentID
		if harness == "" {
			harness = defaultHarnessID
		}
		byID[ev.TaskID] = TaskSummary{
			TaskID:    ev.TaskID,
			Status:    types.TaskStatus(ev.Status),
			Harness:   harness,
			AgentID:   ev.AgentID,
			CreatedAt: ev.CreatedAt,
		}
	}

	tasks := make([]TaskSummary, 0, len(byID))
	for _, item := range byID {
		tasks = append(tasks, item)
	}
	sort.SliceStable(tasks, func(i, j int) bool {
		if !tasks[i].CreatedAt.Equal(tasks[j].CreatedAt) {
			return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
		}
		return tasks[i].TaskID < tasks[j].TaskID
	})
	if len(tasks) > 50 {
		tasks = tasks[:50]
	}
	writeJSON(w, http.StatusOK, map[string]any{"tasks": tasks})
}

func extractTaskPrompt(input map[string]any) string {
	if input == nil {
		return ""
	}
	for _, key := range []string{"prompt", "message", "task", "query"} {
		if v, ok := input[key]; ok {
			if str, ok := v.(string); ok && str != "" {
				return str
			}
		}
	}
	return ""
}

func (s *Server) readAllEvents() []TaskEvent {
	events := []TaskEvent{}
	if s.events == nil {
		return events
	}
	f, err := os.Open(s.events.Path())
	if err != nil {
		return events
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var ev TaskEvent
		if json.Unmarshal(line, &ev) == nil {
			events = append(events, ev)
		}
	}
	return events
}

func (s *Server) handleEventsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", "")
		return
	}
	events := s.readAllEvents()
	if len(events) > 50 {
		events = events[len(events)-50:]
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": events})
}

func (s *Server) handleEventsStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", "")
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "stream_unsupported", "streaming unsupported", "")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	all := s.readAllEvents()
	cursor := len(all)
	_, _ = fmt.Fprintf(w, "event: ready\ndata: {\"count\":%d}\n\n", cursor)
	flusher.Flush()

	pollTicker := time.NewTicker(700 * time.Millisecond)
	defer pollTicker.Stop()
	pingTicker := time.NewTicker(15 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-pollTicker.C:
			current := s.readAllEvents()
			if len(current) < cursor {
				cursor = 0
			}
			for ; cursor < len(current); cursor++ {
				payload, err := json.Marshal(current[cursor])
				if err != nil {
					continue
				}
				_, _ = fmt.Fprintf(w, "event: task_event\ndata: %s\n\n", payload)
			}
			flusher.Flush()
		case <-pingTicker.C:
			_, _ = fmt.Fprint(w, ": ping\n\n")
			flusher.Flush()
		}
	}
}
