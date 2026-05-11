package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/axis-cli/axis/internal/contextpack"
	"github.com/axis-cli/axis/internal/control"
	"github.com/axis-cli/axis/internal/kernel/orchestrator"
	"github.com/axis-cli/axis/internal/model/provider"
	"github.com/axis-cli/axis/internal/model/providerconfig"
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

	shellCmd := &cobra.Command{
		Use:   "shell",
		Short: "Start an interactive Axis shell",
		RunE:  runShell,
	}
	shellCmd.Flags().Bool("no-prompt", false, "Suppress interactive shell prompt for pipe/automation drivers")

	rootCmd.AddCommand(runCmd, statusCmd, startCmd, shellCmd, newProviderCommand(), newAskCommand(), newContextCommand(), newJudgeCommand(), newEvolveCommand(), newMemoryCommand())

	rootCmd.PersistentFlags().StringVar(&app.providerName, "provider", "mock", "Model provider to use: mock, echo, anthropic, openai")
	rootCmd.PersistentFlags().StringVar(&app.modelName, "model", "", "Model name for real providers")

	return rootCmd
}

func runTask(cmd *cobra.Command, args []string) error {
	initOrchestrator()

	if err := submitTask(args[0]); err != nil {
		return fmt.Errorf("failed to submit task: %w", err)
	}

	fmt.Printf("Task %s submitted successfully\n", args[0])
	return nil
}

func getTaskStatus(cmd *cobra.Command, args []string) error {
	taskID := args[0]
	client := control.NewClient(control.NewRuntimeLocator("."), http.DefaultClient)
	status, err := client.Status(context.Background(), taskID)
	if err != nil {
		return fmt.Errorf("failed to get task %s status: %w", taskID, err)
	}

	fmt.Printf("Task %s status: %s\n", taskID, status.Status)
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
	return runLocalRuntime(ctx, ".", cmd.OutOrStdout())
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
		app.orch = orchestrator.NewOrchestrator(orchestrator.WithModelProvider(p))
		if err := app.orch.RegisterContract(defaultContract()); err != nil {
			fmt.Fprintf(os.Stderr, "Error registering default contract: %v\n", err)
		}
	})
}

func (app *App) resolveProvider() (string, []provider.ProviderOption) {
	if app.providerName != "mock" || app.modelName != "" {
		return app.providerName, app.providerOptions()
	}
	cfg, err := providerconfig.NewStore(".").Load()
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

func submitTask(taskID string) error {
	task := &types.AgentTask{
		TaskID:     taskID,
		ContractID: "default",
		Input:      map[string]any{"message": "test"},
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
