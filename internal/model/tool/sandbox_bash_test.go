package tool

import (
	"context"
	"os/exec"
	"strings"
	"testing"
)

func TestDockerAvailable(t *testing.T) {
	if !DockerAvailable() {
		t.Skip("docker not available")
	}
	t.Log("docker is available")
}

func TestSandboxedBashTool_Execute(t *testing.T) {
	if !DockerAvailable() {
		t.Skip("docker not available")
	}
	sb := NewSandboxedBashTool(SandboxConfig{WorkDir: t.TempDir()})
	defer sb.Cleanup()

	result, err := sb.Execute(context.Background(), map[string]any{"command": "echo hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := result["output"].(string)
	if strings.TrimSpace(output) != "hello" {
		t.Errorf("expected 'hello', got %q", output)
	}
	if result["exit_code"].(int) != 0 {
		t.Errorf("expected exit_code 0, got %v", result["exit_code"])
	}
}

func TestSandboxedBashTool_NetworkIsolation(t *testing.T) {
	if !DockerAvailable() {
		t.Skip("docker not available")
	}
	sb := NewSandboxedBashTool(SandboxConfig{WorkDir: t.TempDir(), Network: "none"})
	defer sb.Cleanup()

	result, err := sb.Execute(context.Background(), map[string]any{"command": "curl -s --max-time 3 http://google.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["exit_code"].(int) == 0 {
		t.Error("expected non-zero exit code with network=none")
	}
}

func TestSandboxedBashTool_FileSystemIsolation(t *testing.T) {
	if !DockerAvailable() {
		t.Skip("docker not available")
	}
	sb := NewSandboxedBashTool(SandboxConfig{WorkDir: t.TempDir()})
	defer sb.Cleanup()

	result, err := sb.Execute(context.Background(), map[string]any{"command": "touch /workspace/testfile"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["exit_code"].(int) == 0 {
		t.Error("expected non-zero exit code writing to read-only mount")
	}
}

func TestSandboxedBashTool_Cleanup(t *testing.T) {
	if !DockerAvailable() {
		t.Skip("docker not available")
	}
	sb := NewSandboxedBashTool(SandboxConfig{WorkDir: t.TempDir()})

	// Force container creation
	_, err := sb.Execute(context.Background(), map[string]any{"command": "true"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cid := sb.containerID
	if cid == "" {
		t.Fatal("container was not created")
	}

	if err := sb.Cleanup(); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	// Verify container no longer exists
	out, _ := exec.Command("docker", "inspect", cid).CombinedOutput()
	if !strings.Contains(string(out), "No such object") && !strings.Contains(string(out), "Error") {
		t.Errorf("container %s still exists after cleanup", cid)
	}
}
