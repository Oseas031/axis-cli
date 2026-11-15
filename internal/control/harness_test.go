package control

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func TestHelperCLIHarness(t *testing.T) {
	if os.Getenv("GO_WANT_HARNESS_HELPER") != "1" {
		return
	}
	fmt.Println("helper-line-1")
	for i, a := range os.Args {
		if a == "--write-file" && i+1 < len(os.Args) {
			_ = os.WriteFile(os.Args[i+1], []byte("isolated-change\n"), 0o600)
		}
		if a == "--sleep" {
			time.Sleep(10 * time.Second)
		}
		if a == "--prompt" && i+1 < len(os.Args) {
			fmt.Printf("echo:%s\n", os.Args[i+1])
		}
	}
	os.Exit(0)
}

func TestHarness_RegisterPersistAndExecuteCLI(t *testing.T) {
	dir := t.TempDir()
	evLog := NewTaskEventLog(dir)
	srv := NewServerWithEventLog(stubRuntime{}, RuntimeRecord{
		Protocol:    "http",
		Address:     "http://127.0.0.1:4321",
		ProjectRoot: dir,
	}, evLog)
	srv.SetUIEnabled(true)
	h := srv.Handler()

	// 1. Builtin axis-native exists by default
	listReq := httptest.NewRequest(http.MethodGet, "/v1/harnesses", nil)
	listRec := httptest.NewRecorder()
	h.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("GET /v1/harnesses status = %d, want 200", listRec.Code)
	}
	var listResp struct {
		Harnesses []HarnessSpec `json:"harnesses"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode harnesses: %v", err)
	}
	if len(listResp.Harnesses) != 1 || listResp.Harnesses[0].ID != defaultHarnessID {
		t.Fatalf("unexpected default harnesses: %+v", listResp.Harnesses)
	}

	// 2. Register a CLI harness backed by the test binary helper
	t.Setenv("GO_WANT_HARNESS_HELPER", "1")
	regBody, _ := json.Marshal(HarnessSpec{
		ID:          "helper-cli",
		Name:        "Helper CLI",
		Kind:        "cli",
		Command:     os.Args[0],
		Args:        []string{"-test.run=^TestHelperCLIHarness$", "--", "--prompt", "{prompt}"},
		Description: "test cli harness",
	})
	regReq := httptest.NewRequest(http.MethodPost, "/v1/harnesses", bytes.NewReader(regBody))
	regRec := httptest.NewRecorder()
	h.ServeHTTP(regRec, regReq)
	if regRec.Code != http.StatusCreated {
		t.Fatalf("POST /v1/harnesses status = %d, body = %s", regRec.Code, regRec.Body.String())
	}

	// Verify persistence on reload
	reloaded := NewHarnessManager(dir, evLog)
	if _, ok := reloaded.Get("helper-cli"); !ok {
		t.Fatalf("expected helper-cli to persist in .axis/harnesses.json")
	}

	// 3. Unknown harness returns 400
	badTask, _ := json.Marshal(SubmitTaskRequest{
		Task: &types.AgentTask{
			TaskID:   "t-bad-h",
			Input:    map[string]any{"prompt": "hi"},
			Metadata: map[string]string{"harness": "missing-h"},
		},
	})
	badReq := httptest.NewRequest(http.MethodPost, "/v1/tasks", bytes.NewReader(badTask))
	badRec := httptest.NewRecorder()
	h.ServeHTTP(badRec, badReq)
	if badRec.Code != http.StatusBadRequest {
		t.Fatalf("unknown harness status = %d, want 400", badRec.Code)
	}

	// 4. Submit task to helper-cli and poll completion
	subBody, _ := json.Marshal(SubmitTaskRequest{
		Task: &types.AgentTask{
			TaskID:   "t-cli-1",
			Input:    map[string]any{"prompt": "hello-multi-harness"},
			Metadata: map[string]string{"harness": "helper-cli", "agent_id": "helper-cli"},
		},
	})
	subReq := httptest.NewRequest(http.MethodPost, "/v1/tasks", bytes.NewReader(subBody))
	subRec := httptest.NewRecorder()
	h.ServeHTTP(subRec, subReq)
	if subRec.Code != http.StatusAccepted {
		t.Fatalf("POST /v1/tasks status = %d, body = %s", subRec.Code, subRec.Body.String())
	}

	deadline := time.Now().Add(5 * time.Second)
	var statusResp StatusResponse
	for time.Now().Before(deadline) {
		stReq := httptest.NewRequest(http.MethodGet, "/v1/tasks/t-cli-1/status", nil)
		stRec := httptest.NewRecorder()
		h.ServeHTTP(stRec, stReq)
		if stRec.Code != http.StatusOK {
			t.Fatalf("GET status code = %d", stRec.Code)
		}
		if err := json.Unmarshal(stRec.Body.Bytes(), &statusResp); err != nil {
			t.Fatalf("decode status: %v", err)
		}
		if statusResp.Status == types.TaskStatusCompleted || statusResp.Status == types.TaskStatusFailed {
			break
		}
		time.Sleep(30 * time.Millisecond)
	}
	if statusResp.Status != types.TaskStatusCompleted {
		t.Fatalf("cli task status = %s (err=%s), want completed", statusResp.Status, statusResp.Error)
	}
	text, _ := statusResp.Output["text"].(string)
	if !strings.Contains(text, "echo:hello-multi-harness") {
		t.Fatalf("unexpected cli output text: %q", text)
	}

	// 5. GET /v1/tasks includes the CLI task with harness metadata
	tasksReq := httptest.NewRequest(http.MethodGet, "/v1/tasks", nil)
	tasksRec := httptest.NewRecorder()
	h.ServeHTTP(tasksRec, tasksReq)
	if tasksRec.Code != http.StatusOK {
		t.Fatalf("GET /v1/tasks code = %d", tasksRec.Code)
	}
	var tasksPayload struct {
		Tasks []TaskSummary `json:"tasks"`
	}
	if err := json.Unmarshal(tasksRec.Body.Bytes(), &tasksPayload); err != nil {
		t.Fatalf("decode tasks: %v", err)
	}
	if len(tasksPayload.Tasks) == 0 || tasksPayload.Tasks[0].TaskID != "t-cli-1" || tasksPayload.Tasks[0].Harness != "helper-cli" {
		t.Fatalf("unexpected tasks list: %+v", tasksPayload.Tasks)
	}

	// 6. Simulate runtime restart and verify persisted task snapshot restores output
	srv2 := NewServerWithEventLog(stubRuntime{}, RuntimeRecord{
		Protocol:    "http",
		Address:     "http://127.0.0.1:4321",
		ProjectRoot: dir,
	}, evLog)
	srv2.SetUIEnabled(true)
	h2 := srv2.Handler()

	restoredReq := httptest.NewRequest(http.MethodGet, "/v1/tasks/t-cli-1/status", nil)
	restoredRec := httptest.NewRecorder()
	h2.ServeHTTP(restoredRec, restoredReq)
	if restoredRec.Code != http.StatusOK {
		t.Fatalf("restored status code = %d, want 200", restoredRec.Code)
	}
	var restoredResp StatusResponse
	if err := json.Unmarshal(restoredRec.Body.Bytes(), &restoredResp); err != nil {
		t.Fatalf("decode restored status: %v", err)
	}
	if restoredText, _ := restoredResp.Output["text"].(string); !strings.Contains(restoredText, "echo:hello-multi-harness") {
		t.Fatalf("expected persisted output after restart, got %q", restoredText)
	}
}

func TestHarness_CancelRunningCLITask(t *testing.T) {
	dir := t.TempDir()
	evLog := NewTaskEventLog(dir)
	srv := NewServerWithEventLog(stubRuntime{}, RuntimeRecord{
		Protocol:    "http",
		Address:     "http://127.0.0.1:4321",
		ProjectRoot: dir,
	}, evLog)
	srv.SetUIEnabled(true)
	h := srv.Handler()

	t.Setenv("GO_WANT_HARNESS_HELPER", "1")
	regBody, _ := json.Marshal(HarnessSpec{
		ID:      "slow-cli",
		Name:    "Slow CLI",
		Kind:    "cli",
		Command: os.Args[0],
		Args:    []string{"-test.run=^TestHelperCLIHarness$", "--", "--sleep"},
	})
	regReq := httptest.NewRequest(http.MethodPost, "/v1/harnesses", bytes.NewReader(regBody))
	regRec := httptest.NewRecorder()
	h.ServeHTTP(regRec, regReq)
	if regRec.Code != http.StatusCreated {
		t.Fatalf("register slow-cli status = %d", regRec.Code)
	}

	subBody, _ := json.Marshal(SubmitTaskRequest{
		Task: &types.AgentTask{
			TaskID:   "t-cancel-1",
			Input:    map[string]any{"prompt": "wait"},
			Metadata: map[string]string{"harness": "slow-cli"},
		},
	})
	subReq := httptest.NewRequest(http.MethodPost, "/v1/tasks", bytes.NewReader(subBody))
	subRec := httptest.NewRecorder()
	h.ServeHTTP(subRec, subReq)
	if subRec.Code != http.StatusAccepted {
		t.Fatalf("submit status = %d", subRec.Code)
	}

	// Wait briefly until task transitions to running
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		stReq := httptest.NewRequest(http.MethodGet, "/v1/tasks/t-cancel-1/status", nil)
		stRec := httptest.NewRecorder()
		h.ServeHTTP(stRec, stReq)
		var resp StatusResponse
		if json.Unmarshal(stRec.Body.Bytes(), &resp) == nil && resp.Status == types.TaskStatusRunning {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	cancelReq := httptest.NewRequest(http.MethodPost, "/v1/tasks/t-cancel-1/cancel", nil)
	cancelRec := httptest.NewRecorder()
	h.ServeHTTP(cancelRec, cancelReq)
	if cancelRec.Code != http.StatusOK {
		t.Fatalf("cancel status = %d, body = %s", cancelRec.Code, cancelRec.Body.String())
	}

	// Task should converge to failed after cancellation
	deadline = time.Now().Add(3 * time.Second)
	var finalStatus types.TaskStatus
	for time.Now().Before(deadline) {
		stReq := httptest.NewRequest(http.MethodGet, "/v1/tasks/t-cancel-1/status", nil)
		stRec := httptest.NewRecorder()
		h.ServeHTTP(stRec, stReq)
		var resp StatusResponse
		if json.Unmarshal(stRec.Body.Bytes(), &resp) == nil {
			finalStatus = resp.Status
			if finalStatus == types.TaskStatusFailed {
				break
			}
		}
		time.Sleep(25 * time.Millisecond)
	}
	if finalStatus != types.TaskStatusFailed {
		t.Fatalf("expected failed status after cancel, got %s", finalStatus)
	}
}

func TestHarness_WorktreeIsolationDiffAndCleanup(t *testing.T) {
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
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("init\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGit("add", "README.md")
	runGit("commit", "-m", "init")

	evLog := NewTaskEventLog(dir)
	srv := NewServerWithEventLog(stubRuntime{}, RuntimeRecord{
		Protocol:    "http",
		Address:     "http://127.0.0.1:4321",
		ProjectRoot: dir,
	}, evLog)
	srv.SetUIEnabled(true)
	h := srv.Handler()

	t.Setenv("GO_WANT_HARNESS_HELPER", "1")
	regBody, _ := json.Marshal(HarnessSpec{
		ID:      "wt-cli",
		Name:    "WT CLI",
		Kind:    "cli",
		Command: os.Args[0],
		Args:    []string{"-test.run=^TestHelperCLIHarness$", "--", "--write-file", "wt-note.txt"},
	})
	regReq := httptest.NewRequest(http.MethodPost, "/v1/harnesses", bytes.NewReader(regBody))
	regRec := httptest.NewRecorder()
	h.ServeHTTP(regRec, regReq)
	if regRec.Code != http.StatusCreated {
		t.Fatalf("register wt-cli status = %d", regRec.Code)
	}

	subBody, _ := json.Marshal(SubmitTaskRequest{
		Task: &types.AgentTask{
			TaskID:   "t-wt-1",
			Input:    map[string]any{"prompt": "write isolated file"},
			Metadata: map[string]string{"harness": "wt-cli", "isolate": "true"},
		},
	})
	subReq := httptest.NewRequest(http.MethodPost, "/v1/tasks", bytes.NewReader(subBody))
	subRec := httptest.NewRecorder()
	h.ServeHTTP(subRec, subReq)
	if subRec.Code != http.StatusAccepted {
		t.Fatalf("submit status = %d", subRec.Code)
	}

	deadline := time.Now().Add(5 * time.Second)
	var resp StatusResponse
	for time.Now().Before(deadline) {
		stReq := httptest.NewRequest(http.MethodGet, "/v1/tasks/t-wt-1/status", nil)
		stRec := httptest.NewRecorder()
		h.ServeHTTP(stRec, stReq)
		if json.Unmarshal(stRec.Body.Bytes(), &resp) == nil && resp.Status == types.TaskStatusCompleted {
			break
		}
		time.Sleep(30 * time.Millisecond)
	}
	if resp.Status != types.TaskStatusCompleted {
		t.Fatalf("expected completed, got %s (err=%s)", resp.Status, resp.Error)
	}
	if _, err := os.Stat(filepath.Join(dir, "wt-note.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected main workspace not to be modified when isolate=true")
	}
	diffStr, _ := resp.Output["diff"].(string)
	if !strings.Contains(diffStr, "wt-note.txt") {
		t.Fatalf("expected diff to capture wt-note.txt, got %q", diffStr)
	}

	cleanReq := httptest.NewRequest(http.MethodPost, "/v1/tasks/t-wt-1/worktree/cleanup", nil)
	cleanRec := httptest.NewRecorder()
	h.ServeHTTP(cleanRec, cleanReq)
	if cleanRec.Code != http.StatusOK {
		t.Fatalf("cleanup status = %d, body = %s", cleanRec.Code, cleanRec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(dir, ".axis", "worktrees", "t-wt-1")); !os.IsNotExist(err) {
		t.Fatalf("expected worktree directory to be removed after cleanup")
	}
}
