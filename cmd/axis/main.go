package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/axis-cli/axis/internal/agent"
	"github.com/axis-cli/axis/internal/contextpack"
	"github.com/axis-cli/axis/internal/control"
	"github.com/axis-cli/axis/internal/kernel/orchestrator"
	"github.com/axis-cli/axis/internal/memory/horizon"
	"github.com/axis-cli/axis/internal/model/compactor"
	"github.com/axis-cli/axis/internal/model/provider"
	"github.com/axis-cli/axis/internal/model/providerconfig"
	"github.com/axis-cli/axis/internal/project"
	"github.com/axis-cli/axis/internal/types"
	"github.com/spf13/cobra"
)

// envAPIKeyForProvider returns the environment variable name for a provider's API key.
func envAPIKeyForProvider(providerName string) string {
	switch providerName {
	case "anthropic":
		return "ANTHROPIC_API_KEY"
	case "openai":
		return "OPENAI_API_KEY"
	case "deepseek":
		return "DEEPSEEK_API_KEY"
	case "minimax":
		return "MINIMAX_API_KEY"
	default:
		return ""
	}
}

var (
	orch       *orchestrator.Orchestrator
	defaultApp = &App{providerName: "mock"}
)

type App struct {
	orch         *orchestrator.Orchestrator
	orchOnce     sync.Once
	providerName string
	modelName    string
	root         string // project root for file-backed stores; empty means in-memory
}

// resolvedRoot returns the project root, resolving from cwd if not explicitly set.
func (app *App) resolvedRoot() string {
	if app.root != "" {
		return app.root
	}
	app.root = project.MustResolveRoot()
	return app.root
}

func main() {
	rootCmd := NewRootCommand(defaultApp)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func NewRootCommand(app *App) *cobra.Command {
	if app.root != "" {
		_ = contextpack.InitDefaultRegistry(app.root)
	}

	rootCmd := &cobra.Command{
		Use:   "axis",
		Short: "Agent-native scheduling system",
		Long:  "Axis provides unified task scheduling capabilities for AI Agents.",
	}

	runCmd := &cobra.Command{
		Use:   "run [task-id]",
		Short: "Submit and run a task",
		Args:  cobra.ExactArgs(1),
		RunE:  runTask,
	}
	runCmd.Flags().String("prompt", "", "Natural language task input")
	runCmd.Flags().String("input", "", "JSON task input (e.g. '{\"message\": \"hello\"}')")
	runCmd.Flags().Bool("background", false, "Submit task and return immediately without waiting")

	statusCmd := &cobra.Command{
		Use:   "status [task-id]",
		Short: "Get task status",
		Args:  cobra.ExactArgs(1),
		RunE:  getTaskStatus,
	}

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the orchestrator",
		RunE:  startOrchestrator,
	}
	startCmd.Flags().Int("port", 0, "Fixed port for the control server (0 = random)")

	shellCmd := &cobra.Command{
		Use:   "shell",
		Short: "Start an interactive Axis shell",
		RunE:  runShell,
	}
	shellCmd.Flags().Bool("no-prompt", false, "Suppress interactive shell prompt for pipe/automation drivers")

	rootCmd.AddCommand(runCmd, statusCmd, startCmd, shellCmd, newProviderCommand(), newAskCommand(), newContextCommand(), newJudgeCommand(), newEvolveCommand(), newMemoryCommand(), newSkillsCommand(), newGUICommand(), newVigilCommand(), newDocsCommand())

	rootCmd.PersistentFlags().StringVar(&app.providerName, "provider", "mock", "Model provider to use: mock, echo, anthropic, openai")
	rootCmd.PersistentFlags().StringVar(&app.modelName, "model", "", "Model name for real providers")

	return rootCmd
}

func runTask(cmd *cobra.Command, args []string) error {
	prompt, _ := cmd.Flags().GetString("prompt")
	inputJSON, _ := cmd.Flags().GetString("input")
	background, _ := cmd.Flags().GetBool("background")

	// Background mode: submit to running Local Control Plane (requires `axis start`)
	if background {
		var input map[string]any
		switch {
		case inputJSON != "":
			if err := json.Unmarshal([]byte(inputJSON), &input); err != nil {
				return fmt.Errorf("invalid --input JSON: %w", err)
			}
		case prompt != "":
			input = map[string]any{"message": prompt}
		default:
			input = map[string]any{"message": args[0]}
		}
		task := &types.AgentTask{TaskID: args[0], ContractID: "default", Input: input, Status: types.TaskStatusPending}
		client := control.NewClient(control.NewRuntimeLocator(defaultApp.resolvedRoot()), http.DefaultClient)
		if _, err := client.SubmitTask(context.Background(), task); err != nil {
			return fmt.Errorf("failed to submit to runtime (is 'axis start' running?): %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Task %s submitted. Use 'axis status %s' to check progress.\n", args[0], args[0])
		return nil
	}

	// Synchronous mode: in-process execution
	initOrchestrator()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := orch.Start(ctx); err != nil {
		return fmt.Errorf("failed to start orchestrator: %w", err)
	}
	defer func() { _ = orch.Shutdown(context.Background()) }()

	if err := submitTask(args[0], prompt, inputJSON); err != nil {
		return fmt.Errorf("failed to submit task: %w", err)
	}

	// v1: poll for completion. TODO: event-driven notification from orchestrator.
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			fmt.Fprintf(cmd.ErrOrStderr(), "Task %s timed out\n", args[0])
			return ctx.Err()
		case <-ticker.C:
			status, err := orch.GetTaskStatus(args[0])
			if err != nil {
				continue // task may not be picked up yet
			}
			switch status {
			case types.TaskStatusCompleted:
				fmt.Fprintf(cmd.OutOrStdout(), "Task %s completed\n", args[0])
				return nil
			case types.TaskStatusFailed:
				fmt.Fprintf(cmd.OutOrStdout(), "Task %s failed\n", args[0])
				return fmt.Errorf("task %s failed", args[0])
			}
		}
	}
}

func getTaskStatus(cmd *cobra.Command, args []string) error {
	taskID := args[0]
	client := control.NewClient(control.NewRuntimeLocator(defaultApp.resolvedRoot()), http.DefaultClient)
	status, err := client.Status(context.Background(), taskID)
	if err != nil {
		return fmt.Errorf("failed to get task %s status: %w", taskID, err)
	}

	fmt.Printf("Task %s status: %s\n", taskID, status.Status)
	if status.Error != "" {
		fmt.Printf("Error: %s\n", status.Error)
	}
	if len(status.Output) > 0 {
		out, _ := json.MarshalIndent(status.Output, "", "  ")
		fmt.Printf("Output:\n%s\n", out)
	}
	return nil
}

func startOrchestrator(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down gracefully...")
		if err := orch.Shutdown(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Error shutting down: %v\n", err)
		}
		cancel()
	}()

	fmt.Println("Orchestrator started. Press Ctrl+C to stop.")
	port, _ := cmd.Flags().GetInt("port")
	return runLocalRuntime(ctx, defaultApp.resolvedRoot(), cmd.OutOrStdout(), port)
}

func initOrchestrator() {
	defaultApp.initOrchestrator()
	orch = defaultApp.orch
}

func (app *App) initOrchestrator() {
	app.orchOnce.Do(func() {
		providerName, opts := app.resolveProvider()
		p, err := provider.NewProvider(providerName, opts...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to create provider %q: %v, using mock\n", providerName, err)
			p = provider.NewMockModelProvider()
		}

		if providerName == "mock" {
			fmt.Fprintf(os.Stderr, "Warning: using mock provider. Configure a real provider with: axis provider add\n")
		}

		// Wrap with FallbackProvider if fallback_profile is configured
		p = app.wrapWithFallback(p)

		// Create LLM Agent Executor with the resolved provider
		eventLog := control.NewTaskEventLog(app.resolvedRoot())
		emitter := &eventLogEmitter{log: eventLog}
		memStore := horizon.NewStore(project.MemoryDir(app.resolvedRoot()))
		_ = memStore.Init()

		// Create offload compactor (degrades to noop on failure)
		offloadDir := project.MemoryDir(app.resolvedRoot()) + string(os.PathSeparator) + "offload"
		var compactorOpt agent.LLMAgentOption
		if c, err := compactor.New(compactor.DefaultConfig(offloadDir)); err == nil {
			compactorOpt = agent.WithHistoryCompactor(c)
		} else {
			compactorOpt = func(e *agent.LLMAgentExecutor) {} // noop: keep default compactor
		}

		agentExec := agent.NewLLMAgentExecutor(p, nil,
			agent.WithAgentID("axis-coding-agent"),
			agent.WithSystemPrompt("You are Axis Coding Agent. Use available tools to complete tasks. Be concise and direct. Do not over-analyze simple questions — if the user asks what tools you have, just list them briefly. When done, respond with your final output without tool calls.\n\nEnvironment: Windows with WSL bash. Tools available in PATH: go, git, find, grep, wc, cat. For Windows-specific commands use cmd.exe /c \"...\". Do NOT retry the same command if it fails — try a different approach."),
			agent.WithMaxIterations(20),
			agent.WithMaxErrors(5),
			agent.WithEventEmitter(emitter),
			agent.WithPostJudge(&agent.ExecutionJudge{}),
			agent.WithMemory(agent.NewHorizonMemory(memStore)),
			compactorOpt,
		)

		app.orch = orchestrator.NewOrchestrator(
			orchestrator.WithModelProvider(p),
			orchestrator.WithAgentExecutor(agentExec),
		)
		if err := app.orch.RegisterContract(defaultContract()); err != nil {
			fmt.Fprintf(os.Stderr, "Error registering default contract: %v\n", err)
		}
	})
}

func (app *App) resolveProvider() (string, []provider.ProviderOption) {
	if app.providerName != "mock" || app.modelName != "" {
		return app.providerName, app.providerOptions()
	}
	cfg, err := providerconfig.NewStore(app.resolvedRoot()).Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load provider config: %v\n", err)
		return app.providerName, app.providerOptions()
	}
	profile, ok := cfg.Active()
	if !ok {
		return app.providerName, app.providerOptions()
	}
	return profile.Provider, profile.ProviderOptions()
}

func (app *App) wrapWithFallback(primary provider.ModelProvider) provider.ModelProvider {
	cfg, err := providerconfig.NewStore(app.resolvedRoot()).Load()
	if err != nil || cfg.FallbackProfile == "" {
		return primary
	}
	fbProfile, ok := cfg.Profiles[cfg.FallbackProfile]
	if !ok || fbProfile.Archived {
		return primary
	}
	fb, err := provider.NewProvider(fbProfile.Provider, fbProfile.ProviderOptions()...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create fallback provider %q: %v\n", cfg.FallbackProfile, err)
		return primary
	}
	fmt.Fprintf(os.Stderr, "Fallback provider: %s (model=%s)\n", cfg.FallbackProfile, fbProfile.Model)
	return provider.NewFallbackProvider(120*time.Second, primary, fb)
}

func (app *App) providerOptions() []provider.ProviderOption {
	opts := make([]provider.ProviderOption, 0, 3)
	modelName := app.modelName
	if modelName == "" {
		modelName = defaultModelForProvider(app.providerName)
	}
	if modelName != "" {
		opts = append(opts, provider.WithModel(modelName))
	}
	// Fallback to environment variable for API key when no project-local profile is active.
	if key := os.Getenv(envAPIKeyForProvider(app.providerName)); key != "" {
		opts = append(opts, provider.WithAPIKey(key))
	}
	return opts
}

func defaultModelForProvider(providerName string) string {
	switch providerName {
	case "anthropic":
		return "claude-3-5-sonnet-20241022"
	case "openai":
		return "gpt-4o-mini"
	case "deepseek":
		return "deepseek-v4-flash"
	case "minimax":
		return "MiniMax-M2.7"
	default:
		return ""
	}
}

func submitTask(taskID, prompt, inputJSON string) error {
	var input map[string]any
	switch {
	case inputJSON != "":
		if err := json.Unmarshal([]byte(inputJSON), &input); err != nil {
			return fmt.Errorf("invalid --input JSON: %w", err)
		}
	case prompt != "":
		input = map[string]any{"message": prompt}
	default:
		input = map[string]any{"message": taskID}
	}

	task := &types.AgentTask{
		TaskID:     taskID,
		ContractID: "default",
		Input:      input,
		Status:     types.TaskStatusPending,
	}

	return orch.SubmitTask(task)
}

func defaultContract() *types.AgentContract {
	return &types.AgentContract{
		ContractID: "default",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{
				{
					Name:        "message",
					Type:        types.FieldTypeString,
					Required:    true,
					Description: "Default shell task message",
				},
			},
		},
		OutputSchema: &types.OutputSchema{
			Fields: []types.FieldDef{
				{
					Name:        "status",
					Type:        types.FieldTypeString,
					Required:    false,
					Description: "Execution status",
				},
				{
					Name:        "message",
					Type:        types.FieldTypeString,
					Required:    false,
					Description: "Execution message",
				},
				{
					Name:        "text",
					Type:        types.FieldTypeString,
					Required:    false,
					Description: "Free-form text output",
				},
			},
		},
	}
}


// eventLogEmitter adapts control.TaskEventLog to agent.EventEmitter.
type eventLogEmitter struct {
	log *control.TaskEventLog
}

func (e *eventLogEmitter) Emit(taskID, eventType, message string) {
	_ = e.log.Append(control.TaskEvent{
		TaskID:    taskID,
		EventType: eventType,
		Actor:     "axis-coding-agent",
		Message:   message,
	})
}
