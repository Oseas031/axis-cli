package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/axis-cli/axis/internal/kernel/orchestrator"
	"github.com/axis-cli/axis/internal/types"
	"github.com/spf13/cobra"
)

var (
	orch *orchestrator.Orchestrator
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

	rootCmd.AddCommand(runCmd, statusCmd, startCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runTask(cmd *cobra.Command, args []string) error {
	if orch == nil {
		return fmt.Errorf("orchestrator not initialized. Please run 'axis start' first")
	}

	taskID := args[0]

	task := &types.AgentTask{
		TaskID:     taskID,
		ContractID: "default",
		Input:      map[string]any{"message": "test"},
		Status:     types.TaskStatusPending,
	}

	if err := orch.SubmitTask(task); err != nil {
		return fmt.Errorf("failed to submit task: %w", err)
	}

	fmt.Printf("Task %s submitted successfully\n", taskID)
	return nil
}

func getTaskStatus(cmd *cobra.Command, args []string) error {
	if orch == nil {
		return fmt.Errorf("orchestrator not initialized. Please run 'axis start' first")
	}

	taskID := args[0]
	status := orch.GetTaskStatus(taskID)
	if status == "" {
		return fmt.Errorf("task %s not found", taskID)
	}

	fmt.Printf("Task %s status: %s\n", taskID, status)
	return nil
}

func startOrchestrator(cmd *cobra.Command, args []string) error {
	orch = orchestrator.NewOrchestrator()

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
