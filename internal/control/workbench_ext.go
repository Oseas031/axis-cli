package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

// --- N7: Skills (.axis/skills/<name>/SKILL.md) ---

type SkillSummary struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Path        string `json:"path,omitempty"`
}

type SaveSkillRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Prompt      string `json:"prompt"`
	Harness     string `json:"harness,omitempty"`
	Output      string `json:"output_sample,omitempty"`
}

func (s *Server) projectRoot() string {
	if s.record.ProjectRoot != "" {
		return s.record.ProjectRoot
	}
	return "."
}

func (s *Server) handleSkills(w http.ResponseWriter, r *http.Request) {
	root := s.projectRoot()
	skillsDir := filepath.Join(root, ".axis", "skills")

	switch r.Method {
	case http.MethodGet:
		entries, err := os.ReadDir(skillsDir)
		out := []SkillSummary{}
		if err == nil {
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				mdPath := filepath.Join(skillsDir, e.Name(), "SKILL.md")
				raw, rerr := os.ReadFile(mdPath)
				if rerr != nil {
					continue
				}
				out = append(out, SkillSummary{
					Name:        e.Name(),
					Description: extractFrontmatterField(string(raw), "description"),
					Path:        mdPath,
				})
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"skills": out})

	case http.MethodPost:
		var req SaveSkillRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid skill payload", "")
			return
		}
		name := strings.ToLower(strings.TrimSpace(req.Name))
		name = sanitizePathToken(name)
		if name == "" || name == "task" && strings.TrimSpace(req.Name) == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "skill name is required", "")
			return
		}
		desc := strings.TrimSpace(req.Description)
		if desc == "" {
			desc = "Captured from Axis workbench task"
		}
		harness := strings.TrimSpace(req.Harness)
		if harness == "" {
			harness = defaultHarnessID
		}
		body := fmt.Sprintf("---\nname: %s\ndescription: %s\n---\n\n# %s\n\n- Preferred Harness: `%s`\n\n## Prompt Template\n\n```text\n%s\n```\n",
			name, desc, name, harness, strings.TrimSpace(req.Prompt))
		if strings.TrimSpace(req.Output) != "" {
			sample := strings.TrimSpace(req.Output)
			if len(sample) > 1200 {
				sample = sample[:1200] + "\n...(truncated)"
			}
			body += fmt.Sprintf("\n## Reference Output\n\n```text\n%s\n```\n", sample)
		}

		dir := filepath.Join(skillsDir, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			writeError(w, http.StatusInternalServerError, "skill_save_failed", err.Error(), "")
			return
		}
		mdPath := filepath.Join(dir, "SKILL.md")
		if err := os.WriteFile(mdPath, []byte(body), 0o600); err != nil {
			writeError(w, http.StatusInternalServerError, "skill_save_failed", err.Error(), "")
			return
		}
		writeJSON(w, http.StatusCreated, SkillSummary{Name: name, Description: desc, Path: mdPath})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", "")
	}
}

func extractFrontmatterField(content, key string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return ""
	}
	prefix := key + ":"
	for _, line := range lines[1:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			break
		}
		if strings.HasPrefix(trimmed, prefix) {
			return strings.Trim(strings.TrimSpace(strings.TrimPrefix(trimmed, prefix)), `"'`)
		}
	}
	return ""
}

// --- N9: Memory Notes (.axis/memory/notes.json) ---

type MemoryNote struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Source    string    `json:"source,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type memoryStore struct {
	mu   sync.Mutex
	path string
}

func newMemoryStore(root string) *memoryStore {
	return &memoryStore{path: filepath.Join(root, ".axis", "memory", "notes.json")}
}

func (ms *memoryStore) loadLocked() []MemoryNote {
	data, err := os.ReadFile(ms.path)
	if err != nil {
		return []MemoryNote{}
	}
	var notes []MemoryNote
	if json.Unmarshal(data, &notes) != nil {
		return []MemoryNote{}
	}
	return notes
}

func (ms *memoryStore) saveLocked(notes []MemoryNote) error {
	if err := os.MkdirAll(filepath.Dir(ms.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(notes, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ms.path, data, 0o600)
}

func (s *Server) handleMemory(w http.ResponseWriter, r *http.Request) {
	ms := newMemoryStore(s.projectRoot())
	ms.mu.Lock()
	defer ms.mu.Unlock()

	switch r.Method {
	case http.MethodGet:
		notes := ms.loadLocked()
		sort.Slice(notes, func(i, j int) bool { return notes[i].CreatedAt.After(notes[j].CreatedAt) })
		writeJSON(w, http.StatusOK, map[string]any{"notes": notes})

	case http.MethodPost:
		var req MemoryNote
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid memory note", "")
			return
		}
		req.Content = strings.TrimSpace(req.Content)
		if req.Content == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "content is required", "")
			return
		}
		if req.ID == "" {
			req.ID = fmt.Sprintf("mem-%d", time.Now().UTC().UnixNano())
		}
		if strings.TrimSpace(req.Title) == "" {
			req.Title = "Note"
		}
		if req.CreatedAt.IsZero() {
			req.CreatedAt = time.Now().UTC()
		}
		notes := ms.loadLocked()
		notes = append(notes, req)
		if err := ms.saveLocked(notes); err != nil {
			writeError(w, http.StatusInternalServerError, "memory_save_failed", err.Error(), "")
			return
		}
		writeJSON(w, http.StatusCreated, req)

	case http.MethodDelete:
		id := strings.TrimSpace(r.URL.Query().Get("id"))
		if id == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "id query param is required", "")
			return
		}
		notes := ms.loadLocked()
		kept := make([]MemoryNote, 0, len(notes))
		for _, n := range notes {
			if n.ID != id {
				kept = append(kept, n)
			}
		}
		if err := ms.saveLocked(kept); err != nil {
			writeError(w, http.StatusInternalServerError, "memory_delete_failed", err.Error(), "")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"deleted": id})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", "")
	}
}

// --- N9: Routines (.axis/routines.json) & N10: Multi-Harness Pipelines ---

type Routine struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Harness    string     `json:"harness"`
	Prompt     string     `json:"prompt"`
	Interval   string     `json:"interval"` // Go duration e.g. "30m", "1h"
	Isolate    bool       `json:"isolate,omitempty"`
	Enabled    bool       `json:"enabled"`
	LastRunAt  *time.Time `json:"last_run_at,omitempty"`
	LastTaskID string     `json:"last_task_id,omitempty"`
}

type PipelineStep struct {
	Name       string  `json:"name,omitempty"`
	Role       string  `json:"role,omitempty"`
	Harness    string  `json:"harness"`
	Prompt     string  `json:"prompt"`
	Isolate    bool    `json:"isolate,omitempty"`
	CostBudget float64 `json:"cost_budget,omitempty"`
}

type PipelineRun struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Status      types.TaskStatus `json:"status"`
	CurrentStep int              `json:"current_step"`
	TotalSteps  int              `json:"total_steps"`
	BudgetUSD   float64          `json:"budget_usd,omitempty"`
	SpentUSD    float64          `json:"spent_usd,omitempty"`
	StepTaskIDs []string         `json:"step_task_ids,omitempty"`
	Handoffs    []HandoffRecord  `json:"handoffs,omitempty"`
	Error       string           `json:"error,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	CompletedAt *time.Time       `json:"completed_at,omitempty"`
}

type WorkbenchExtensions struct {
	server    *Server
	mu        sync.RWMutex
	routines  map[string]Routine
	pipelines map[string]*PipelineRun
	stopOnce  sync.Once
}

func newWorkbenchExtensions(s *Server) *WorkbenchExtensions {
	ext := &WorkbenchExtensions{
		server:    s,
		routines:  make(map[string]Routine),
		pipelines: make(map[string]*PipelineRun),
	}
	ext.loadRoutines()
	return ext
}

func (e *WorkbenchExtensions) routinesPath() string {
	return filepath.Join(e.server.projectRoot(), ".axis", "routines.json")
}

func (e *WorkbenchExtensions) loadRoutines() {
	data, err := os.ReadFile(e.routinesPath())
	if err != nil {
		return
	}
	var list []Routine
	if json.Unmarshal(data, &list) != nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, r := range list {
		if r.ID != "" {
			e.routines[r.ID] = r
		}
	}
}

func (e *WorkbenchExtensions) saveRoutinesLocked() error {
	list := make([]Routine, 0, len(e.routines))
	for _, r := range e.routines {
		list = append(list, r)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	p := e.routinesPath()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o600)
}

func (e *WorkbenchExtensions) StartRoutineLoop(ctx context.Context) {
	e.stopOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case now := <-ticker.C:
					e.tickRoutines(now.UTC())
				}
			}
		}()
	})
}

func (e *WorkbenchExtensions) tickRoutines(now time.Time) {
	e.mu.Lock()
	var due []Routine
	for _, r := range e.routines {
		if !r.Enabled {
			continue
		}
		dur, err := time.ParseDuration(r.Interval)
		if err != nil || dur <= 0 {
			continue
		}
		if r.LastRunAt == nil || now.Sub(*r.LastRunAt) >= dur {
			due = append(due, r)
		}
	}
	e.mu.Unlock()

	for _, r := range due {
		_, _ = e.triggerRoutine(r.ID)
	}
}

func (e *WorkbenchExtensions) triggerRoutine(id string) (string, error) {
	e.mu.Lock()
	r, ok := e.routines[id]
	if !ok {
		e.mu.Unlock()
		return "", fmt.Errorf("routine %q not found", id)
	}
	now := time.Now().UTC()
	taskID := fmt.Sprintf("rtn-%s-%d", sanitizePathToken(r.ID), now.UnixNano()%1000000)
	r.LastRunAt = &now
	r.LastTaskID = taskID
	e.routines[id] = r
	_ = e.saveRoutinesLocked()
	e.mu.Unlock()

	meta := map[string]string{
		"harness":  r.Harness,
		"agent_id": r.Harness,
		"routine":  r.ID,
	}
	if r.Isolate {
		meta["isolate"] = "true"
	}
	task := &types.AgentTask{
		TaskID:     taskID,
		ContractID: "default",
		Input:      map[string]any{"prompt": r.Prompt},
		Metadata:   meta,
		CreatedAt:  now,
	}
	if err := e.server.submitTaskInternal(task); err != nil {
		return "", err
	}
	return taskID, nil
}

func (s *Server) ensureExt() *WorkbenchExtensions {
	if s.ext == nil {
		s.ext = newWorkbenchExtensions(s)
	}
	return s.ext
}

func (s *Server) submitTaskInternal(task *types.AgentTask) error {
	if task == nil || task.TaskID == "" {
		return fmt.Errorf("task is required")
	}
	harnessID := defaultHarnessID
	if task.Metadata != nil && strings.TrimSpace(task.Metadata["harness"]) != "" {
		harnessID = strings.TrimSpace(task.Metadata["harness"])
	}
	agentID := harnessID
	if task.Metadata != nil && strings.TrimSpace(task.Metadata["agent_id"]) != "" {
		agentID = strings.TrimSpace(task.Metadata["agent_id"])
	}

	if harnessID != defaultHarnessID {
		mgr := s.ensureHarnesses()
		spec, ok := mgr.Get(harnessID)
		if !ok {
			return fmt.Errorf("harness %q is not registered", harnessID)
		}
		if err := mgr.SubmitCLI(spec, task); err != nil {
			return err
		}
		s.appendEvent(TaskEvent{TaskID: task.TaskID, EventType: "submitted", Actor: "local-control", AgentID: agentID, Status: string(types.TaskStatusPending), Message: "task submitted to harness " + harnessID})
		return nil
	}
	if err := s.runtime.SubmitTask(task); err != nil {
		return err
	}
	s.appendEvent(TaskEvent{TaskID: task.TaskID, EventType: "submitted", Actor: "local-control", AgentID: agentID, Status: string(types.TaskStatusPending), Message: "task submitted"})
	return nil
}

func (s *Server) handleRoutines(w http.ResponseWriter, r *http.Request) {
	ext := s.ensureExt()
	switch r.Method {
	case http.MethodGet:
		ext.mu.RLock()
		list := make([]Routine, 0, len(ext.routines))
		for _, item := range ext.routines {
			list = append(list, item)
		}
		ext.mu.RUnlock()
		sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
		writeJSON(w, http.StatusOK, map[string]any{"routines": list})

	case http.MethodPost:
		var req Routine
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid routine payload", "")
			return
		}
		req.ID = sanitizePathToken(strings.TrimSpace(req.ID))
		req.Prompt = strings.TrimSpace(req.Prompt)
		if req.ID == "" || req.Prompt == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "id and prompt are required", "")
			return
		}
		if strings.TrimSpace(req.Name) == "" {
			req.Name = req.ID
		}
		if strings.TrimSpace(req.Harness) == "" {
			req.Harness = defaultHarnessID
		}
		if strings.TrimSpace(req.Interval) == "" {
			req.Interval = "1h"
		}
		if d, err := time.ParseDuration(req.Interval); err != nil || d <= 0 {
			writeError(w, http.StatusBadRequest, "bad_request", "interval must be a valid positive duration (e.g. 30m, 1h)", "")
			return
		}

		ext.mu.Lock()
		ext.routines[req.ID] = req
		err := ext.saveRoutinesLocked()
		ext.mu.Unlock()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "routine_save_failed", err.Error(), "")
			return
		}
		writeJSON(w, http.StatusCreated, req)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", "")
	}
}

func (s *Server) handleRoutineAction(w http.ResponseWriter, r *http.Request) {
	ext := s.ensureExt()
	prefix := "/v1/routines/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		writeError(w, http.StatusNotFound, "not_found", "endpoint not found", "")
		return
	}
	rest := strings.TrimPrefix(r.URL.Path, prefix)
	if strings.HasSuffix(rest, "/run") && r.Method == http.MethodPost {
		id := strings.TrimSuffix(rest, "/run")
		taskID, err := ext.triggerRoutine(id)
		if err != nil {
			writeError(w, http.StatusNotFound, "routine_run_failed", err.Error(), "")
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"routine_id": id, "task_id": taskID})
		return
	}
	if strings.HasSuffix(rest, "/toggle") && r.Method == http.MethodPost {
		id := strings.TrimSuffix(rest, "/toggle")
		ext.mu.Lock()
		item, ok := ext.routines[id]
		if !ok {
			ext.mu.Unlock()
			writeError(w, http.StatusNotFound, "not_found", "routine not found", "")
			return
		}
		item.Enabled = !item.Enabled
		ext.routines[id] = item
		_ = ext.saveRoutinesLocked()
		ext.mu.Unlock()
		writeJSON(w, http.StatusOK, item)
		return
	}
	if !strings.Contains(rest, "/") && r.Method == http.MethodDelete {
		ext.mu.Lock()
		delete(ext.routines, rest)
		_ = ext.saveRoutinesLocked()
		ext.mu.Unlock()
		writeJSON(w, http.StatusOK, map[string]any{"deleted": rest})
		return
	}
	writeError(w, http.StatusNotFound, "not_found", "endpoint not found", "")
}

func (s *Server) handlePipelines(w http.ResponseWriter, r *http.Request) {
	ext := s.ensureExt()
	switch r.Method {
	case http.MethodGet:
		ext.mu.RLock()
		list := make([]PipelineRun, 0, len(ext.pipelines))
		for _, p := range ext.pipelines {
			cp := *p
			if len(p.StepTaskIDs) > 0 {
				cp.StepTaskIDs = append([]string(nil), p.StepTaskIDs...)
			}
			if len(p.Handoffs) > 0 {
				cp.Handoffs = append([]HandoffRecord(nil), p.Handoffs...)
			}
			list = append(list, cp)
		}
		ext.mu.RUnlock()
		sort.Slice(list, func(i, j int) bool { return list[i].CreatedAt.After(list[j].CreatedAt) })
		writeJSON(w, http.StatusOK, map[string]any{"pipelines": list})

	case http.MethodPost:
		var req struct {
			ID        string         `json:"id"`
			Name      string         `json:"name"`
			BudgetUSD float64        `json:"budget_usd,omitempty"`
			Steps     []PipelineStep `json:"steps"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid pipeline payload", "")
			return
		}
		if len(req.Steps) == 0 {
			writeError(w, http.StatusBadRequest, "bad_request", "at least one step is required", "")
			return
		}
		now := time.Now().UTC()
		id := strings.TrimSpace(req.ID)
		if id == "" {
			id = fmt.Sprintf("pipe-%d", now.UnixNano()%1000000)
		}
		name := strings.TrimSpace(req.Name)
		if name == "" {
			name = id
		}
		run := &PipelineRun{
			ID:         id,
			Name:       name,
			Status:     types.TaskStatusRunning,
			TotalSteps: len(req.Steps),
			BudgetUSD:  req.BudgetUSD,
			CreatedAt:  now,
		}
		ext.mu.Lock()
		ext.pipelines[id] = run
		resp := *run
		ext.mu.Unlock()

		s.appendEvent(TaskEvent{TaskID: id, EventType: "pipeline_started", Actor: "pipeline", Status: string(types.TaskStatusRunning), Message: fmt.Sprintf("pipeline %s started (%d steps)", name, len(req.Steps))})
		go ext.executePipeline(id, req.BudgetUSD, req.Steps)
		writeJSON(w, http.StatusAccepted, resp)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", "")
	}
}

func (e *WorkbenchExtensions) executePipeline(pipeID string, pipeBudget float64, steps []PipelineStep) {
	prevOut := ""
	prevDiff := ""
	prevHandoffJSON := ""
	spentTotal := 0.0

	for i, step := range steps {
		if pipeBudget > 0 && spentTotal >= pipeBudget {
			e.server.appendEvent(TaskEvent{TaskID: pipeID, EventType: "budget_exceeded", Actor: "pipeline", Status: string(types.TaskStatusFailed), Message: fmt.Sprintf("pipeline budget $%.4f exceeded (spent $%.4f)", pipeBudget, spentTotal)})
			e.finishPipeline(pipeID, types.TaskStatusFailed, fmt.Sprintf("budget_exceeded: spent $%.4f >= limit $%.4f", spentTotal, pipeBudget))
			return
		}

		roleSpec, hasRole := getRoleSpec(step.Role)
		harnessID := strings.TrimSpace(step.Harness)
		if harnessID == "" {
			if hasRole && roleSpec.DefaultHarness != "" {
				harnessID = roleSpec.DefaultHarness
			} else {
				harnessID = defaultHarnessID
			}
		}

		prompt := strings.ReplaceAll(step.Prompt, "{{prev_output}}", prevOut)
		prompt = strings.ReplaceAll(prompt, "{{prev_diff}}", prevDiff)
		prompt = strings.ReplaceAll(prompt, "{{prev_handoff}}", prevHandoffJSON)
		if hasRole && roleSpec.SystemPrompt != "" {
			prompt = roleSpec.SystemPrompt + "\n\n" + prompt
		}

		taskID := fmt.Sprintf("%s-s%d", pipeID, i+1)
		meta := map[string]string{
			"harness":  harnessID,
			"agent_id": harnessID,
			"pipeline": pipeID,
		}
		if hasRole {
			meta["role"] = roleSpec.ID
		}
		if step.Isolate || (hasRole && roleSpec.Isolate) {
			meta["isolate"] = "true"
		}
		task := &types.AgentTask{
			TaskID:     taskID,
			ContractID: "default",
			Input:      map[string]any{"prompt": prompt},
			Metadata:   meta,
			CostBudget: step.CostBudget,
			CreatedAt:  time.Now().UTC(),
		}

		e.mu.Lock()
		if p, ok := e.pipelines[pipeID]; ok {
			p.CurrentStep = i + 1
			p.StepTaskIDs = append(p.StepTaskIDs, taskID)
		}
		e.mu.Unlock()

		e.server.appendEvent(TaskEvent{
			TaskID:    taskID,
			EventType: "pipeline_step",
			Actor:     "pipeline",
			AgentID:   harnessID,
			Status:    string(types.TaskStatusRunning),
			Message:   fmt.Sprintf("pipeline %s step %d/%d -> %s", pipeID, i+1, len(steps), harnessID),
		})

		if err := e.server.submitTaskInternal(task); err != nil {
			e.finishPipeline(pipeID, types.TaskStatusFailed, err.Error())
			return
		}

		status, outText, diffText, errMsg := e.waitTaskCompletion(taskID, harnessID, 5*time.Minute)
		stepCost := estimateCharCostUSD(prompt, outText)
		spentTotal += stepCost

		e.mu.Lock()
		if p, ok := e.pipelines[pipeID]; ok {
			p.SpentUSD = spentTotal
		}
		e.mu.Unlock()

		if status != types.TaskStatusCompleted {
			if errMsg == "" {
				errMsg = fmt.Sprintf("step %d (%s) ended with %s", i+1, taskID, status)
			}
			e.finishPipeline(pipeID, types.TaskStatusFailed, errMsg)
			return
		}

		nextRole := ""
		if i+1 < len(steps) {
			nextRole = steps[i+1].Role
		}
		handoff, herr := buildTypedHandoff(step.Role, nextRole, taskID, outText, diffText)
		if herr != nil {
			e.finishPipeline(pipeID, types.TaskStatusFailed, herr.Error())
			return
		}
		if b, merr := json.MarshalIndent(handoff, "", "  "); merr == nil {
			prevHandoffJSON = string(b)
		}
		e.mu.Lock()
		if p, ok := e.pipelines[pipeID]; ok {
			p.Handoffs = append(p.Handoffs, handoff)
		}
		e.mu.Unlock()

		prevOut = outText
		prevDiff = diffText
	}

	e.finishPipeline(pipeID, types.TaskStatusCompleted, "")
}

func (e *WorkbenchExtensions) waitTaskCompletion(taskID, harnessID string, timeout time.Duration) (types.TaskStatus, string, string, string) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var status types.TaskStatus
		var res *types.TaskResult
		if harnessID != defaultHarnessID {
			mgr := e.server.ensureHarnesses()
			st, ok := mgr.GetTaskStatus(taskID)
			if ok {
				status = st
				if status == types.TaskStatusCompleted || status == types.TaskStatusFailed {
					res, _ = mgr.GetTaskResult(taskID)
				}
			}
		} else {
			st, err := e.server.runtime.GetTaskStatus(taskID)
			if err == nil {
				status = st
				if status == types.TaskStatusCompleted || status == types.TaskStatusFailed {
					res, _ = e.server.runtime.GetTaskResult(taskID)
				}
			} else if !errors.Is(err, os.ErrNotExist) {
				// keep polling briefly
			}
		}

		if status == types.TaskStatusCompleted || status == types.TaskStatusFailed {
			outText := ""
			diffText := ""
			errMsg := ""
			if res != nil {
				errMsg = res.Error
				if res.Output != nil {
					if v, ok := res.Output["text"].(string); ok {
						outText = v
					}
					if d, ok := res.Output["diff"].(string); ok {
						diffText = d
					}
				}
			}
			return status, outText, diffText, errMsg
		}
		time.Sleep(40 * time.Millisecond)
	}
	return types.TaskStatusFailed, "", "", "pipeline step timed out"
}

func (e *WorkbenchExtensions) finishPipeline(pipeID string, status types.TaskStatus, errMsg string) {
	now := time.Now().UTC()
	e.mu.Lock()
	if p, ok := e.pipelines[pipeID]; ok {
		p.Status = status
		p.Error = errMsg
		p.CompletedAt = &now
	}
	e.mu.Unlock()

	evType := "pipeline_completed"
	msg := fmt.Sprintf("pipeline %s completed", pipeID)
	if status == types.TaskStatusFailed {
		evType = "pipeline_failed"
		msg = fmt.Sprintf("pipeline %s failed: %s", pipeID, errMsg)
	}
	e.server.appendEvent(TaskEvent{TaskID: pipeID, EventType: evType, Actor: "pipeline", Status: string(status), Message: msg})
}
