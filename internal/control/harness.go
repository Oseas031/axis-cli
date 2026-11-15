package control

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

const defaultHarnessID = "axis-native"

// HarnessSpec describes an execution harness registered in the control plane.
// Builtin "axis-native" routes to the embedded Axis orchestrator; "cli" harnesses
// spawn external coding/agent CLIs as managed subprocesses.
type HarnessSpec struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Kind        string   `json:"kind"` // "native" | "cli"
	Command     string   `json:"command,omitempty"`
	Args        []string `json:"args,omitempty"`
	Description string   `json:"description,omitempty"`
	Builtin     bool     `json:"builtin,omitempty"`
}

type managedTask struct {
	task   *types.AgentTask
	result *types.TaskResult
	cancel context.CancelFunc
}

// HarnessManager coordinates multiple heterogeneous agent harnesses under a single
// control plane and bridges external CLI executions into TaskEvents & StatusResponse.
type HarnessManager struct {
	root      string
	events    *TaskEventLog
	mu        sync.RWMutex
	harnesses map[string]HarnessSpec
	tasks     map[string]*managedTask
}

func NewHarnessManager(root string, events *TaskEventLog) *HarnessManager {
	if root == "" {
		root = "."
	}
	m := &HarnessManager{
		root:      root,
		events:    events,
		harnesses: make(map[string]HarnessSpec),
		tasks:     make(map[string]*managedTask),
	}
	m.harnesses[defaultHarnessID] = HarnessSpec{
		ID:          defaultHarnessID,
		Name:        "Axis Native",
		Kind:        "native",
		Description: "Embedded Axis orchestrator and tool runtime",
		Builtin:     true,
	}
	_ = m.loadFromDisk()
	return m
}

func (m *HarnessManager) configPath() string {
	return filepath.Join(m.root, ".axis", "harnesses.json")
}

func (m *HarnessManager) loadFromDisk() error {
	data, err := os.ReadFile(m.configPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	var specs []HarnessSpec
	if err := json.Unmarshal(data, &specs); err != nil {
		return err
	}
	for _, spec := range specs {
		if spec.ID == "" || spec.ID == defaultHarnessID {
			continue
		}
		if spec.Kind == "" {
			spec.Kind = "cli"
		}
		if spec.Name == "" {
			spec.Name = spec.ID
		}
		spec.Builtin = false
		m.harnesses[spec.ID] = spec
	}
	return nil
}

func (m *HarnessManager) saveToDiskLocked() error {
	var custom []HarnessSpec
	for _, spec := range m.harnesses {
		if spec.Builtin || spec.ID == defaultHarnessID {
			continue
		}
		custom = append(custom, spec)
	}
	sort.Slice(custom, func(i, j int) bool {
		return custom[i].ID < custom[j].ID
	})
	path := m.configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(custom, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func (m *HarnessManager) List() []HarnessSpec {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]HarnessSpec, 0, len(m.harnesses))
	for _, h := range m.harnesses {
		out = append(out, h)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Builtin != out[j].Builtin {
			return out[i].Builtin
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func (m *HarnessManager) Get(id string) (HarnessSpec, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	spec, ok := m.harnesses[id]
	return spec, ok
}

func (m *HarnessManager) Register(spec HarnessSpec) (HarnessSpec, error) {
	spec.ID = strings.TrimSpace(spec.ID)
	spec.Command = strings.TrimSpace(spec.Command)
	if spec.ID == "" || spec.ID == defaultHarnessID {
		return HarnessSpec{}, fmt.Errorf("invalid harness id %q", spec.ID)
	}
	if strings.ContainsAny(spec.ID, " /\\") {
		return HarnessSpec{}, fmt.Errorf("harness id must not contain spaces or slashes")
	}
	if spec.Kind == "" {
		spec.Kind = "cli"
	}
	if spec.Kind != "cli" {
		return HarnessSpec{}, fmt.Errorf("unsupported harness kind %q (only \"cli\" can be registered)", spec.Kind)
	}
	if spec.Command == "" {
		return HarnessSpec{}, fmt.Errorf("command is required for cli harness")
	}
	if spec.Name == "" {
		spec.Name = spec.ID
	}
	spec.Builtin = false

	m.mu.Lock()
	defer m.mu.Unlock()
	m.harnesses[spec.ID] = spec
	if err := m.saveToDiskLocked(); err != nil {
		return HarnessSpec{}, err
	}
	return spec, nil
}

func (m *HarnessManager) HasTask(taskID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.tasks[taskID]
	return ok
}

func (m *HarnessManager) CancelTask(taskID string) error {
	m.mu.Lock()
	mt, ok := m.tasks[taskID]
	if !ok || mt.task == nil {
		m.mu.Unlock()
		return fmt.Errorf("task %s not found", taskID)
	}
	if mt.task.Status == types.TaskStatusCompleted || mt.task.Status == types.TaskStatusFailed {
		m.mu.Unlock()
		return fmt.Errorf("task %s is already %s", taskID, mt.task.Status)
	}
	cancel := mt.cancel
	m.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	return nil
}

func (m *HarnessManager) GetTaskStatus(taskID string) (types.TaskStatus, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	mt, ok := m.tasks[taskID]
	if !ok || mt.task == nil {
		return "", false
	}
	return mt.task.Status, true
}

func (m *HarnessManager) GetTaskResult(taskID string) (*types.TaskResult, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	mt, ok := m.tasks[taskID]
	if !ok || mt.result == nil {
		return nil, false
	}
	return mt.result, true
}

func (m *HarnessManager) GetAllTasks() []*types.AgentTask {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*types.AgentTask, 0, len(m.tasks))
	for _, mt := range m.tasks {
		if mt.task != nil {
			cp := *mt.task
			out = append(out, &cp)
		}
	}
	return out
}

func (m *HarnessManager) SubmitCLI(spec HarnessSpec, task *types.AgentTask) error {
	if task == nil || task.TaskID == "" {
		return fmt.Errorf("task is required")
	}
	now := time.Now().UTC()
	cp := *task
	if cp.CreatedAt.IsZero() {
		cp.CreatedAt = now
	}
	cp.Status = types.TaskStatusPending
	if cp.Metadata == nil {
		cp.Metadata = map[string]string{}
	}
	cp.Metadata["harness"] = spec.ID
	if cp.Metadata["agent_id"] == "" {
		cp.Metadata["agent_id"] = spec.ID
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

	m.mu.Lock()
	if _, exists := m.tasks[cp.TaskID]; exists {
		m.mu.Unlock()
		cancel()
		return fmt.Errorf("task %s already exists", cp.TaskID)
	}
	m.tasks[cp.TaskID] = &managedTask{task: &cp, cancel: cancel}
	m.mu.Unlock()

	go m.runCLITask(ctx, cancel, spec, cp.TaskID)
	return nil
}

func (m *HarnessManager) runCLITask(ctx context.Context, cancel context.CancelFunc, spec HarnessSpec, taskID string) {
	defer func() {
		cancel()
		m.mu.Lock()
		if mt2, ok := m.tasks[taskID]; ok {
			mt2.cancel = nil
		}
		m.mu.Unlock()
	}()

	m.mu.Lock()
	mt := m.tasks[taskID]
	if mt == nil || mt.task == nil {
		m.mu.Unlock()
		return
	}
	if ctx.Err() != nil {
		m.mu.Unlock()
		m.finishCLITask(taskID, mt.task.Metadata["agent_id"], spec, time.Now().UTC(), "", "", ctx.Err())
		return
	}
	started := time.Now().UTC()
	mt.task.Status = types.TaskStatusRunning
	mt.task.StartedAt = &started
	prompt := extractTaskPrompt(mt.task.Input)
	agentID := mt.task.Metadata["agent_id"]
	workdir := mt.task.Metadata["workdir"]
	m.mu.Unlock()

	m.appendEvent(TaskEvent{
		TaskID:    taskID,
		EventType: "harness_started",
		Actor:     "harness-manager",
		AgentID:   agentID,
		Status:    string(types.TaskStatusRunning),
		Message:   fmt.Sprintf("dispatched to cli harness %s (%s)", spec.ID, spec.Command),
	})

	args, usedPlaceholder := buildCLIArgs(spec.Args, prompt)

	cmd := exec.CommandContext(ctx, spec.Command, args...)
	if workdir != "" && filepath.IsAbs(workdir) {
		cmd.Dir = workdir
	} else {
		cmd.Dir = m.root
	}
	cmd.Env = append(os.Environ(),
		"AXIS_TASK_ID="+taskID,
		"AXIS_HARNESS_ID="+spec.ID,
		"AXIS_PROJECT_ROOT="+m.root,
	)

	if !usedPlaceholder && prompt != "" {
		cmd.Stdin = strings.NewReader(prompt)
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		m.finishCLITask(taskID, agentID, spec, started, "", "", err)
		return
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		m.finishCLITask(taskID, agentID, spec, started, "", "", err)
		return
	}

	if err := cmd.Start(); err != nil {
		m.finishCLITask(taskID, agentID, spec, started, "", "", err)
		return
	}

	var outBuf, errBuf strings.Builder
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		m.streamPipe(stdoutPipe, &outBuf, taskID, agentID, "harness_stdout")
	}()
	go func() {
		defer wg.Done()
		m.streamPipe(stderrPipe, &errBuf, taskID, agentID, "harness_stderr")
	}()
	wg.Wait()

	waitErr := cmd.Wait()
	m.finishCLITask(taskID, agentID, spec, started, outBuf.String(), errBuf.String(), waitErr)
}

func (m *HarnessManager) streamPipe(r io.Reader, buf *strings.Builder, taskID, agentID, evType string) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 16*1024), 512*1024)
	for scanner.Scan() {
		line := scanner.Text()
		buf.WriteString(line)
		buf.WriteByte('\n')
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if len(trimmed) > 300 {
			trimmed = trimmed[:300] + "..."
		}
		m.appendEvent(TaskEvent{
			TaskID:    taskID,
			EventType: evType,
			Actor:     "harness-cli",
			AgentID:   agentID,
			Status:    string(types.TaskStatusRunning),
			Message:   trimmed,
		})
	}
}

func (m *HarnessManager) finishCLITask(taskID, agentID string, spec HarnessSpec, started time.Time, stdout, stderr string, runErr error) {
	completed := time.Now().UTC()
	durationMs := completed.Sub(started).Milliseconds()
	status := types.TaskStatusCompleted
	errMsg := ""
	if runErr != nil {
		status = types.TaskStatusFailed
		errMsg = runErr.Error()
		if strings.TrimSpace(stderr) != "" {
			errMsg = fmt.Sprintf("%s: %s", errMsg, strings.TrimSpace(stderr))
		}
	}

	trace := map[string]any{
		"tool":        "cli_harness:" + spec.ID,
		"command":     spec.Command,
		"duration_ms": durationMs,
		"stdout":      strings.TrimSpace(stdout),
	}
	if strings.TrimSpace(stderr) != "" {
		trace["stderr"] = strings.TrimSpace(stderr)
	}
	if errMsg != "" {
		trace["error"] = errMsg
	}

	output := map[string]any{
		"text":         strings.TrimSpace(stdout),
		"harness":      spec.ID,
		"_tool_traces": []any{trace},
	}

	m.mu.Lock()
	if mt, ok := m.tasks[taskID]; ok && mt.task != nil {
		mt.task.Status = status
		mt.task.CompletedAt = &completed
		mt.result = &types.TaskResult{
			TaskID:    taskID,
			Status:    status,
			Output:    output,
			Error:     errMsg,
			Completed: completed,
		}
	}
	m.mu.Unlock()

	evType := "completed"
	msg := fmt.Sprintf("harness %s completed in %dms", spec.ID, durationMs)
	if status == types.TaskStatusFailed {
		evType = "failed"
		msg = fmt.Sprintf("harness %s failed: %s", spec.ID, errMsg)
	}
	m.appendEvent(TaskEvent{
		TaskID:    taskID,
		EventType: evType,
		Actor:     "harness-manager",
		AgentID:   agentID,
		Status:    string(status),
		Message:   msg,
	})
}

func (m *HarnessManager) appendEvent(ev TaskEvent) {
	if m.events == nil {
		return
	}
	_ = m.events.Append(ev)
}

func buildCLIArgs(template []string, prompt string) ([]string, bool) {
	used := false
	out := make([]string, len(template))
	for i, arg := range template {
		if strings.Contains(arg, "{prompt}") {
			out[i] = strings.ReplaceAll(arg, "{prompt}", prompt)
			used = true
		} else {
			out[i] = arg
		}
	}
	return out, used
}

func (s *Server) ensureHarnesses() *HarnessManager {
	if s.harnesses != nil {
		return s.harnesses
	}
	root := s.record.ProjectRoot
	if root == "" {
		root = "."
	}
	s.harnesses = NewHarnessManager(root, s.events)
	return s.harnesses
}

func (s *Server) handleHarnesses(w http.ResponseWriter, r *http.Request) {
	mgr := s.ensureHarnesses()
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{"harnesses": mgr.List()})
	case http.MethodPost:
		var spec HarnessSpec
		if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid harness spec", "")
			return
		}
		saved, err := mgr.Register(spec)
		if err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", err.Error(), "")
			return
		}
		writeJSON(w, http.StatusCreated, saved)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", "")
	}
}
