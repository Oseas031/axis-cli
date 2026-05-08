package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/axis-cli/axis/internal/kernel/orchestrator"
	"github.com/axis-cli/axis/internal/types"
	"github.com/spf13/cobra"
)

var (
	orch      *orchestrator.Orchestrator
	orchMutex sync.Once
)

func main() {
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

	rootCmd.AddCommand(runCmd, statusCmd, startCmd, shellCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
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
	initOrchestrator()

	taskID := args[0]
	status, err := orch.GetTaskStatus(taskID)
	if err != nil {
		return fmt.Errorf("task %s not found in this local Axis process: %w", taskID, err)
	}

	fmt.Printf("Task %s status: %s\n", taskID, status)
	return nil
}

func startOrchestrator(cmd *cobra.Command, args []string) error {
	initOrchestrator()

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

	if err := orch.Start(ctx); err != nil {
		return fmt.Errorf("failed to start orchestrator: %w", err)
	}

	fmt.Println("Orchestrator started. Press Ctrl+C to stop.")

	// Keep the main goroutine running
	<-ctx.Done()
	return nil
}

func runShell(cmd *cobra.Command, args []string) error {
	initOrchestrator()

	ctx, cancel := context.WithCancel(context.Background())
	shutdownOnce := sync.Once{}
	shutdown := func() {
		shutdownOnce.Do(func() {
			cancel()
			if err := orch.Shutdown(context.Background()); err != nil {
				fmt.Fprintf(os.Stderr, "Error shutting down: %v\n", err)
			}
		})
	}
	defer shutdown()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down Axis shell...")
		shutdown()
		os.Exit(0)
	}()

	if err := orch.Start(ctx); err != nil {
		return fmt.Errorf("failed to start orchestrator: %w", err)
	}

	fmt.Println("Axis shell started. Type 'help' for commands, 'exit' to quit.")
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("axis> ")
		if !scanner.Scan() {
			fmt.Println()
			return scanner.Err()
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		command := strings.ToLower(fields[0])
		commandArgs := fields[1:]

		switch command {
		case "help":
			printShellHelp()
		case "run":
			if len(commandArgs) != 1 {
				fmt.Println("Usage: run <task-id>")
				continue
			}
			if err := submitTask(commandArgs[0]); err != nil {
				fmt.Printf("Could not submit task %s: %v\n", commandArgs[0], err)
				continue
			}
			fmt.Printf("Task %s submitted. Try: status %s\n", commandArgs[0], commandArgs[0])
		case "status":
			if len(commandArgs) != 1 {
				fmt.Println("Usage: status <task-id>")
				continue
			}
			status, err := orch.GetTaskStatus(commandArgs[0])
			if err != nil {
				fmt.Printf("Could not get status for task %s: %v\n", commandArgs[0], err)
				fmt.Printf("If this is a new task, try: run %s\n", commandArgs[0])
				continue
			}
			fmt.Printf("Task %s status: %s\n", commandArgs[0], status)
		case "exit", "quit":
			fmt.Println("Exiting Axis shell.")
			return nil
		default:
			fmt.Printf("Unknown command: %s\n", command)
			fmt.Println("Type 'help' to see available commands.")
		}
	}
}

func initOrchestrator() {
	orchMutex.Do(func() {
		orch = orchestrator.NewOrchestrator()
		if err := orch.RegisterContract(defaultContract()); err != nil {
			fmt.Fprintf(os.Stderr, "Error registering default contract: %v\n", err)
		}
	})
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

func printShellHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  help              Show this help message")
	fmt.Println("  run <task-id>     Submit a task")
	fmt.Println("  status <task-id>  Show task status")
	fmt.Println("  exit, quit        Shut down the shell")
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
					Required:    true,
					Description: "Execution status",
				},
				{
					Name:        "message",
					Type:        types.FieldTypeString,
					Required:    true,
					Description: "Execution message",
				},
			},
		},
	}
}
