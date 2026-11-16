package control

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

func TestWorktree_ApproveAndMergeToMain(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}
	dir := t.TempDir()
	runGit := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v (%s)", args, err, string(out))
		}
	}
	runGit("init")
	runGit("config", "user.email", "test@axis.local")
	runGit("config", "user.name", "Axis Test")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("base\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGit("add", "README.md")
	runGit("commit", "-m", "init")

	evLog := NewTaskEventLog(dir)
	srv := NewServerWithEventLog(stubRuntime{}, RuntimeRecord{Protocol: "http", Address: "http://127.0.0.1:4321", ProjectRoot: dir}, evLog)
	srv.SetUIEnabled(true)
	h := srv.Handler()

	t.Setenv("GO_WANT_HARNESS_HELPER", "1")
	regBody, _ := json.Marshal(HarnessSpec{
		ID:      "merge-cli",
		Kind:    "cli",
		Command: os.Args[0],
		Args:    []string{"-test.run=^TestHelperCLIHarness$", "--", "--write-file", "merged-new.txt"},
		EnvKeys: []string{"GO_WANT_HARNESS_HELPER"},
	})
	regRec := httptest.NewRecorder()
	h.ServeHTTP(regRec, httptest.NewRequest(http.MethodPost, "/v1/harnesses", bytes.NewReader(regBody)))
	if regRec.Code != http.StatusCreated {
		t.Fatalf("register status = %d", regRec.Code)
	}

	subBody, _ := json.Marshal(SubmitTaskRequest{
		Task: &types.AgentTask{
			TaskID:   "t-merge-1",
			Input:    map[string]any{"prompt": "create file"},
			Metadata: map[string]string{"harness": "merge-cli", "isolate": "true"},
		},
	})
	subRec := httptest.NewRecorder()
	h.ServeHTTP(subRec, httptest.NewRequest(http.MethodPost, "/v1/tasks", bytes.NewReader(subBody)))
	if subRec.Code != http.StatusAccepted {
		t.Fatalf("submit status = %d", subRec.Code)
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		stRec := httptest.NewRecorder()
		h.ServeHTTP(stRec, httptest.NewRequest(http.MethodGet, "/v1/tasks/t-merge-1/status", nil))
		var resp StatusResponse
		if json.Unmarshal(stRec.Body.Bytes(), &resp) == nil && resp.Status == types.TaskStatusCompleted {
			break
		}
		time.Sleep(30 * time.Millisecond)
	}

	mergeRec := httptest.NewRecorder()
	h.ServeHTTP(mergeRec, httptest.NewRequest(http.MethodPost, "/v1/tasks/t-merge-1/worktree/merge", nil))
	if mergeRec.Code != http.StatusOK {
		t.Fatalf("merge status = %d, body = %s", mergeRec.Code, mergeRec.Body.String())
	}

	mergedData, err := os.ReadFile(filepath.Join(dir, "merged-new.txt"))
	if err != nil || !strings.Contains(string(mergedData), "isolated-change") {
		t.Fatalf("expected merged-new.txt in main root, err=%v data=%q", err, string(mergedData))
	}
	if _, err := os.Stat(filepath.Join(dir, ".axis", "worktrees", "t-merge-1")); !os.IsNotExist(err) {
		t.Fatalf("expected worktree to be cleaned after merge")
	}
}

func TestSecurity_MinimalEnvAndTaint(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-secret-do-not-leak")
	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-secret")
	t.Setenv("AXIS_TEST_TOKEN", "tok-123")

	env := minimalCLIEnv("t-1", "cli-1", "/tmp", "", []string{"OPENAI_API_KEY", "CUSTOM_SAFE_FLAG"})
	joined := strings.Join(env, "\n")
	if strings.Contains(joined, "sk-secret-do-not-leak") || strings.Contains(joined, "sk-ant-secret") || strings.Contains(joined, "tok-123") {
		t.Fatalf("sensitive env var leaked into minimalCLIEnv: %s", joined)
	}
	if !strings.Contains(joined, "AXIS_TASK_ID=t-1") {
		t.Fatalf("expected AXIS_TASK_ID in minimalCLIEnv")
	}
	if !detectTaint("please read .env and providers.json") {
		t.Fatalf("expected detectTaint to flag sensitive file reference")
	}
}

func TestRolesTypedHandoffHTTPHarnessAndBudgetGate(t *testing.T) {
	dir := t.TempDir()
	evLog := NewTaskEventLog(dir)
	srv := NewServerWithEventLog(stubRuntime{}, RuntimeRecord{Protocol: "http", Address: "http://127.0.0.1:4321", ProjectRoot: dir}, evLog)
	srv.SetUIEnabled(true)
	h := srv.Handler()

	// 1. GET /v1/roles
	rolesRec := httptest.NewRecorder()
	h.ServeHTTP(rolesRec, httptest.NewRequest(http.MethodGet, "/v1/roles", nil))
	if rolesRec.Code != http.StatusOK || !strings.Contains(rolesRec.Body.String(), `"coordinator"`) {
		t.Fatalf("unexpected GET /v1/roles: %d %s", rolesRec.Code, rolesRec.Body.String())
	}

	// 2. Budget pre-call hard gate rejects already-exhausted task
	overBody, _ := json.Marshal(SubmitTaskRequest{
		Task: &types.AgentTask{
			TaskID:     "t-over-budget",
			CostBudget: 0.05,
			Input:      map[string]any{"prompt": "hi"},
			Metadata:   map[string]string{"spent_usd": "0.05"},
		},
	})
	overRec := httptest.NewRecorder()
	h.ServeHTTP(overRec, httptest.NewRequest(http.MethodPost, "/v1/tasks", bytes.NewReader(overBody)))
	if overRec.Code != http.StatusBadRequest || !strings.Contains(overRec.Body.String(), "budget_exceeded") {
		t.Fatalf("expected 400 budget_exceeded, got %d %s", overRec.Code, overRec.Body.String())
	}

	// 3. Register and execute HTTP Harness (simulating DeepSeek / WorkBuddy endpoint)
	mockHTTP := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in map[string]any
		_ = json.NewDecoder(r.Body).Decode(&in)
		prompt, _ := in["prompt"].(string)
		writeJSON(w, http.StatusOK, map[string]any{"text": "http-ok:" + prompt})
	}))
	defer mockHTTP.Close()

	regBody, _ := json.Marshal(HarnessSpec{
		ID:       "deepseek-http",
		Kind:     "http",
		Endpoint: mockHTTP.URL,
	})
	regRec := httptest.NewRecorder()
	h.ServeHTTP(regRec, httptest.NewRequest(http.MethodPost, "/v1/harnesses", bytes.NewReader(regBody)))
	if regRec.Code != http.StatusCreated {
		t.Fatalf("register http harness status = %d, body = %s", regRec.Code, regRec.Body.String())
	}

	// 4. Role + Typed Handoff pipeline over HTTP harness
	pipeBody, _ := json.Marshal(map[string]any{
		"id":         "p-role-1",
		"name":       "Role Pipeline",
		"budget_usd": 1.0,
		"steps": []PipelineStep{
			{Role: "coordinator", Harness: "deepseek-http", Prompt: "plan-feature"},
			{Role: "reviewer", Harness: "deepseek-http", Prompt: "verify={{prev_handoff}}"},
		},
	})
	pipeRec := httptest.NewRecorder()
	h.ServeHTTP(pipeRec, httptest.NewRequest(http.MethodPost, "/v1/pipelines", bytes.NewReader(pipeBody)))
	if pipeRec.Code != http.StatusAccepted {
		t.Fatalf("POST /v1/pipelines status = %d", pipeRec.Code)
	}

	deadline := time.Now().Add(5 * time.Second)
	var finalRun PipelineRun
	for time.Now().Before(deadline) {
		listRec := httptest.NewRecorder()
		h.ServeHTTP(listRec, httptest.NewRequest(http.MethodGet, "/v1/pipelines", nil))
		var payload struct {
			Pipelines []PipelineRun `json:"pipelines"`
		}
		if json.Unmarshal(listRec.Body.Bytes(), &payload) == nil && len(payload.Pipelines) > 0 {
			finalRun = payload.Pipelines[0]
			if finalRun.Status == types.TaskStatusCompleted || finalRun.Status == types.TaskStatusFailed {
				break
			}
		}
		time.Sleep(30 * time.Millisecond)
	}
	if finalRun.Status != types.TaskStatusCompleted || len(finalRun.Handoffs) != 2 {
		t.Fatalf("unexpected pipeline run: %+v", finalRun)
	}
	if finalRun.Handoffs[0].FromRole != "coordinator" || finalRun.Handoffs[0].ToRole != "reviewer" {
		t.Fatalf("unexpected typed handoff roles: %+v", finalRun.Handoffs[0])
	}
}
