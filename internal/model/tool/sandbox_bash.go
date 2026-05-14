package tool

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

// SandboxConfig holds configuration for a SandboxedBashTool.
type SandboxConfig struct {
	Image       string        // default "ubuntu:22.04"
	WorkDir     string        // required: host directory to mount as workspace
	Network     string        // default "none"
	MemoryLimit string        // default "512m"
	CPULimit    string        // default "1.0"
	Timeout     time.Duration // default 60s
}

// SandboxedBashTool executes commands inside a Docker container.
// Provides process, network, and filesystem isolation.
type SandboxedBashTool struct {
	image       string
	workDir     string
	networkMode string
	memoryLimit string
	cpuLimit    string
	timeout     time.Duration
	containerID string
	mu          sync.Mutex
}

// NewSandboxedBashTool creates a SandboxedBashTool with the given config.
func NewSandboxedBashTool(cfg SandboxConfig) *SandboxedBashTool {
	if cfg.Image == "" {
		cfg.Image = "axis-sandbox:latest"
	}
	if cfg.Network == "" {
		cfg.Network = "none"
	}
	if cfg.MemoryLimit == "" {
		cfg.MemoryLimit = "512m"
	}
	if cfg.CPULimit == "" {
		cfg.CPULimit = "1.0"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 60 * time.Second
	}
	return &SandboxedBashTool{
		image:       cfg.Image,
		workDir:     cfg.WorkDir,
		networkMode: cfg.Network,
		memoryLimit: cfg.MemoryLimit,
		cpuLimit:    cfg.CPULimit,
		timeout:     cfg.Timeout,
	}
}

// DockerAvailable checks if docker CLI is accessible.
func DockerAvailable() bool {
	return exec.Command("docker", "info").Run() == nil
}

// Name returns the tool name.
func (t *SandboxedBashTool) Name() string { return "bash" }

// Schema returns the tool definition.
func (t *SandboxedBashTool) Schema() types.ToolDefinition {
	return types.ToolDefinition{
		Name:        "bash",
		Description: "Execute a bash command inside an isolated Docker container",
		Parameters: []types.FieldDef{
			{Name: "command", Type: types.FieldTypeString, Required: true, Description: "The command to execute"},
		},
	}
}

// Execute runs a command inside the sandbox container.
func (t *SandboxedBashTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
	cmdStr, ok := input["command"].(string)
	if !ok || cmdStr == "" {
		return map[string]any{"output": "", "exit_code": 1}, fmt.Errorf("command is required and must be a string")
	}

	t.mu.Lock()
	if t.containerID == "" {
		if err := t.ensureContainer(); err != nil {
			t.mu.Unlock()
			return nil, fmt.Errorf("failed to create sandbox container: %w", err)
		}
	}
	cid := t.containerID
	t.mu.Unlock()

	timeoutCtx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, "docker", "exec", cid, "bash", "-c", cmdStr)
	output, execErr := cmd.CombinedOutput()

	exitCode := 0
	if execErr != nil {
		if exitErr, ok := execErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, execErr
		}
	}

	return map[string]any{
		"output":    string(output),
		"exit_code": exitCode,
	}, nil
}

// ensureContainer creates the sandbox container. Must be called with t.mu held.
func (t *SandboxedBashTool) ensureContainer() error {
	args := []string{
		"run", "-d",
		"--label", "axis-sandbox=true",
		"--network", t.networkMode,
		"--memory", t.memoryLimit,
		"--cpus", t.cpuLimit,
		"-v", t.workDir + ":/workspace:ro",
		"-w", "/workspace",
		t.image,
		"sleep", "infinity",
	}
	cmd := exec.Command("docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(out)), err)
	}
	t.containerID = strings.TrimSpace(string(out))
	return nil
}

// Cleanup removes the sandbox container.
func (t *SandboxedBashTool) Cleanup() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.containerID == "" {
		return nil
	}
	err := exec.Command("docker", "rm", "-f", t.containerID).Run()
	t.containerID = ""
	return err
}
