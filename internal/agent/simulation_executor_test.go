package agent

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

func TestSimulationAgentExecutor_FullPipeline(t *testing.T) {
	root := t.TempDir()
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("resolve project root: %v", err)
	}
	binary := filepath.Join(root, "axis.exe")

	build := exec.Command("go", "build", "-o", binary, "./cmd/axis")
	build.Dir = projectRoot
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build axis: %v\n%s", err, out)
	}

	taskID := fmt.Sprintf("sim-%d", time.Now().Unix())
	executor := NewSimulationAgentExecutor()
	req := &AgentExecutionRequest{
		Task: &types.AgentTask{
			TaskID:     taskID,
			ContractID: "default",
			Input: map[string]any{
				"axis_binary":  binary,
				"project_root": root,
				"task_prompt":  "run smoke test",
				"task_id":      taskID,
			},
			Metadata: map[string]string{},
		},
		Autonomy: AutonomyLevelExecute,
	}

	result, err := executor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	logs, _ := result.Output["logs"].([]string)
	for _, line := range logs {
		t.Log(line)
	}

	status, _ := result.Output["status"].(string)
	ok, _ := result.Output["ok"].(bool)
	if !ok {
		t.Fatalf("simulation failed: status=%s output=%v", status, result.Output)
	}
	t.Logf("simulation ok, status=%s", status)
}

func TestExtractFieldReadsRuntimeRecordJSON(t *testing.T) {
	runtimeJSON := `{"pid":12345,"protocol":"http","address":"http://127.0.0.1:1234"}`
	if got := extractField(runtimeJSON, "address"); got != "http://127.0.0.1:1234" {
		t.Fatalf("expected address field, got %q", got)
	}
	if got := extractField(runtimeJSON, "pid"); got != "12345" {
		t.Fatalf("expected pid field, got %q", got)
	}
}

func TestToBashPathUsesDetectedDrivePrefix(t *testing.T) {
	if got := toBashPath("C:/Users/ASUS/project", "/mnt"); got != "/mnt/c/Users/ASUS/project" {
		t.Fatalf("expected /mnt/c path, got %q", got)
	}
	if got := toBashPath("C:/Users/ASUS/project", "/"); got != "/c/Users/ASUS/project" {
		t.Fatalf("expected /c path, got %q", got)
	}
}
