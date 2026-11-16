package control

import (
	"bufio"
	"bytes"
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
// spawn external coding/agent CLIs; "http" harnesses forward tasks to HTTP endpoints.
type HarnessSpec struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Kind        string   `json:"kind"` // "native" | "cli" | "http"
	Command     string   `json:"command,omitempty"`
	Args        []string `json:"args,omitempty"`
	Endpoint    string   `json:"endpoint,omitempty"`
	EnvKeys     []string `json:"env_keys,omitempty"`
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
	_ = m.loadSnapshotsFromDisk()
	return m
}

type persistedTaskSnapshot struct {
	Task   *types.AgentTask  `json:"task"`
	Result *types.TaskResult `json:"result,omitempty"`
}

func (m *HarnessManager) tasksDir() string {
	return filepath.Join(m.root, ".axis", "tasks")
}

func (m *HarnessManager) snapshotPath(taskID string) string {
	return filepath.Join(m.tasksDir(), sanitizePathToken(taskID)+".json")
}

func (m *HarnessManager) saveSnapshot(task *types.AgentTask, result *types.TaskResult) {
	if task == nil || task.TaskID == "" {
		return
	}
	dir := m.tasksDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	data, err := json.MarshalIndent(persistedTaskSnapshot{Task: task, Result: result}, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(m.snapshotPath(task.TaskID), data, 0o600)
}

func (m *HarnessManager) loadSnapshotsFromDisk() error {
	entries, err := os.ReadDir(m.tasksDir())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(m.tasksDir(), entry.Name()))
		if err != nil {
			continue
		}
		var snap persistedTaskSnapshot
		if json.Unmarshal(data, &snap) != nil || snap.Task == nil || snap.Task.TaskID == "" {
			continue
		}
		if snap.Task.Status == types.TaskStatusPending || snap.Task.Status == types.TaskStatusRunning {
			snap.Task.Status = types.TaskStatusFailed
			if snap.Result == nil {
				snap.Result = &types.TaskResult{
					TaskID: snap.Task.TaskID,
					Status: types.TaskStatusFailed,
					Error:  "interrupted by runtime restart",
				}
			}
		}
		m.tasks[snap.Task.TaskID] = &managedTask{task: snap.Task, result: snap.Result}
	}
	return nil
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
	spec.Endpoint = strings.TrimSpace(spec.Endpoint)
	if spec.ID == "" || spec.ID == defaultHarnessID {
		return HarnessSpec{}, fmt.Errorf("invalid harness id %q", spec.ID)
	}
	if strings.ContainsAny(spec.ID, " /\\") {
		return HarnessSpec{}, fmt.Errorf("harness id must not contain spaces or slashes")
	}
	if spec.Kind == "" {
		spec.Kind = "cli"
	}
	switch spec.Kind {
	case "cli":
		if spec.Command == "" {
			return HarnessSpec{}, fmt.Errorf("command is required for cli harness")
		}
	case "http":
		if !strings.HasPrefix(spec.Endpoint, "http://") && !strings.HasPrefix(spec.Endpoint, "https://") {
			return HarnessSpec{}, fmt.Errorf("endpoint must start with http:// or https:// for http harness")
		}
	default:
		return HarnessSpec{}, fmt.Errorf("unsupported harness kind %q (allowed: \"cli\", \"http\")", spec.Kind)
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
	if err := checkBudgetBeforeDispatch(task); err != nil {
		return err
	}
	now := time.Now().UTC()
	cp := *task
	if cp.CreatedAt.IsZero() {
		cp.CreatedAt = now
	}
	cp.Status = types.TaskStatusPending
	if cp.Metadata == nil {
		cp.Metadata = map[string]string{}
	} else {
		cloned := make(map[string]string, len(cp.Metadata))
		for k, v := range cp.Metadata {
			cloned[k] = v
		}
		cp.Metadata = cloned
	}
	cp.Metadata["harness"] = spec.ID
	if cp.Metadata["agent_id"] == "" {
		cp.Metadata["agent_id"] = spec.ID
	}
	if detectTaint(extractTaskPrompt(cp.Input)) {
		cp.Metadata["tainted"] = "true"
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
	m.saveSnapshot(&cp, nil)

	if cp.Metadata["tainted"] == "true" {
		m.appendEvent(TaskEvent{
			TaskID:    cp.TaskID,
			EventType: "taint_marked",
			Actor:     "security-gate",
			AgentID:   cp.Metadata["agent_id"],
			Status:    string(types.TaskStatusPending),
			Message:   "task marked as tainted; sensitive actions require approval",
		})
	}

	if spec.Kind == "http" {
		go m.runHTTPTask(ctx, cancel, spec, cp.TaskID)
	} else {
		go m.runCLITask(ctx, cancel, spec, cp.TaskID)
	}
	return nil
}

func (m *HarnessManager) runHTTPTask(ctx context.Context, cancel context.CancelFunc, spec HarnessSpec, taskID string) {
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
	started := time.Now().UTC()
	mt.task.Status = types.TaskStatusRunning
	mt.task.StartedAt = &started
	prompt := extractTaskPrompt(mt.task.Input)
	agentID := mt.task.Metadata["agent_id"]
	meta := mt.task.Metadata
	m.mu.Unlock()

	m.appendEvent(TaskEvent{
		TaskID:    taskID,
		EventType: "harness_started",
		Actor:     "harness-manager",
		AgentID:   agentID,
		Status:    string(types.TaskStatusRunning),
		Message:   fmt.Sprintf("dispatched to http harness %s (%s)", spec.ID, spec.Endpoint),
	})

	bodyBytes, _ := json.Marshal(map[string]any{
		"task_id":  taskID,
		"prompt":   prompt,
		"metadata": meta,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, spec.Endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		m.finishCLITask(taskID, agentID, spec, started, "", "", "", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		m.finishCLITask(taskID, agentID, spec, started, "", "", "", err)
		return
	}
	defer resp.Body.Close()
	respRaw, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if resp.StatusCode >= 400 {
		m.finishCLITask(taskID, agentID, spec, started, "", string(respRaw), "", fmt.Errorf("http harness returned %d", resp.StatusCode))
		return
	}
	var parsed map[string]any
	outText := strings.TrimSpace(string(respRaw))
	if json.Unmarshal(respRaw, &parsed) == nil {
		if txt, ok := parsed["text"].(string); ok && txt != "" {
			outText = txt
		} else if outMap, ok := parsed["output"].(map[string]any); ok {
			if txt, ok := outMap["text"].(string); ok && txt != "" {
				outText = txt
			}
		}
	}
	m.finishCLITask(taskID, agentID, spec, started, outText, "", "", nil)
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
		m.finishCLITask(taskID, mt.task.Metadata["agent_id"], spec, time.Now().UTC(), "", "", "", ctx.Err())
		return
	}
	started := time.Now().UTC()
	mt.task.Status = types.TaskStatusRunning
	mt.task.StartedAt = &started
	prompt := extractTaskPrompt(mt.task.Input)
	agentID := mt.task.Metadata["agent_id"]
	workdir := mt.task.Metadata["workdir"]
	isolate := mt.task.Metadata["isolate"] == "true"
	m.mu.Unlock()

	isolatedWT := ""
	if isolate && workdir == "" {
		if wtPath, wtErr := prepareTaskWorktree(m.root, taskID); wtErr == nil {
			isolatedWT = wtPath
			workdir = wtPath
			m.appendEvent(TaskEvent{
				TaskID:    taskID,
				EventType: "worktree_created",
				Actor:     "harness-manager",
				AgentID:   agentID,
				Status:    string(types.TaskStatusRunning),
				Message:   wtPath,
			})
		} else {
			m.appendEvent(TaskEvent{
				TaskID:    taskID,
				EventType: "worktree_skipped",
				Actor:     "harness-manager",
				AgentID:   agentID,
				Status:    string(types.TaskStatusRunning),
				Message:   wtErr.Error(),
			})
		}
	}

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
	env := minimalCLIEnv(taskID, spec.ID, m.root, isolatedWT, spec.EnvKeys)
	cmd.Env = env

	if !usedPlaceholder && prompt != "" {
		cmd.Stdin = strings.NewReader(prompt)
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		m.finishCLITask(taskID, agentID, spec, started, "", "", isolatedWT, err)
		return
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		m.finishCLITask(taskID, agentID, spec, started, "", "", isolatedWT, err)
		return
	}

	if err := cmd.Start(); err != nil {
		m.finishCLITask(taskID, agentID, spec, started, "", "", isolatedWT, err)
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
	m.finishCLITask(taskID, agentID, spec, started, outBuf.String(), errBuf.String(), isolatedWT, waitErr)
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

func (m *HarnessManager) finishCLITask(taskID, agentID string, spec HarnessSpec, started time.Time, stdout, stderr, isolatedWT string, runErr error) {
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
	if isolatedWT != "" {
		trace["worktree"] = isolatedWT
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
	if isolatedWT != "" {
		output["worktree"] = isolatedWT
		if diffText := captureWorktreeDiff(isolatedWT); diffText != "" {
			output["diff"] = diffText
		}
	}

	var snapTask *types.AgentTask
	var snapResult *types.TaskResult
	m.mu.Lock()
	if mt, ok := m.tasks[taskID]; ok && mt.task != nil {
		mt.task.Status = status
		mt.task.CompletedAt = &completed
		if mt.task.Metadata == nil {
			mt.task.Metadata = map[string]string{}
		}
		if detectTaint(stdout) {
			mt.task.Metadata["tainted"] = "true"
		}
		if mt.task.Metadata["tainted"] == "true" {
			output["tainted"] = true
		}
		spent := parseTaskSpentUSD(mt.task) + estimateCharCostUSD(extractTaskPrompt(mt.task.Input), stdout)
		mt.task.Metadata["spent_usd"] = fmt.Sprintf("%.5f", spent)
		output["spent_usd"] = spent
		if budget := parseTaskBudgetUSD(mt.task); budget > 0 {
			output["cost_budget"] = budget
		}
		mt.result = &types.TaskResult{
			TaskID:    taskID,
			Status:    status,
			Output:    output,
			Error:     errMsg,
			Completed: completed,
		}
		cp := *mt.task
		snapTask = &cp
		snapResult = mt.result
	}
	m.mu.Unlock()
	if snapTask != nil {
		m.saveSnapshot(snapTask, snapResult)
	}

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

// minimalCLIEnv builds a strictly allowlisted environment for external CLI subprocesses
// so host API keys and secrets are never leaked automatically.
func minimalCLIEnv(taskID, harnessID, root, isolatedWT string, extraKeys []string) []string {
	allow := []string{
		"PATH", "HOME", "USERPROFILE", "HOMEDRIVE", "HOMEPATH",
		"TEMP", "TMP", "TMPDIR", "SYSTEMROOT", "COMSPEC", "WINDIR",
		"LANG", "LC_ALL", "TERM", "GOPATH", "GOMODCACHE", "GOCACHE",
		"GIT_EXEC_PATH",
	}
	allowSet := make(map[string]struct{}, len(allow)+len(extraKeys))
	for _, k := range allow {
		allowSet[strings.ToUpper(k)] = struct{}{}
	}
	for _, k := range extraKeys {
		upper := strings.ToUpper(strings.TrimSpace(k))
		if upper == "" || isSensitiveEnvName(upper) {
			continue
		}
		allowSet[upper] = struct{}{}
	}

	env := make([]string, 0, len(allowSet)+4)
	for _, kv := range os.Environ() {
		idx := strings.IndexByte(kv, '=')
		if idx <= 0 {
			continue
		}
		key := strings.ToUpper(kv[:idx])
		if _, ok := allowSet[key]; ok {
			env = append(env, kv)
		}
	}
	env = append(env,
		"AXIS_TASK_ID="+taskID,
		"AXIS_HARNESS_ID="+harnessID,
		"AXIS_PROJECT_ROOT="+root,
	)
	if isolatedWT != "" {
		env = append(env, "AXIS_WORKTREE="+isolatedWT)
	}
	return env
}

func isSensitiveEnvName(upperKey string) bool {
	for _, kw := range []string{"KEY", "SECRET", "TOKEN", "PASSWORD", "PASSWD", "CREDENTIAL", "AUTH"} {
		if strings.Contains(upperKey, kw) {
			return true
		}
	}
	return false
}

func detectTaint(text string) bool {
	lower := strings.ToLower(text)
	for _, marker := range []string{".env", "providers.json", "id_rsa", "private_key", "api_key"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
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
