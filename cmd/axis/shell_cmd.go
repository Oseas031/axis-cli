package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/axis-cli/axis/internal/intent"
	"github.com/spf13/cobra"
)

func runShell(cmd *cobra.Command, args []string) error {
	initOrchestrator()
	noPrompt, err := cmd.Flags().GetBool("no-prompt")
	if err != nil {
		noPrompt = false
	}

	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()

	ctx, cancel := context.WithCancel(context.Background())
	shutdownOnce := sync.Once{}
	shutdown := func() {
		shutdownOnce.Do(func() {
			// Drain in-flight workers first, keeping ctx valid for dispatch.
			if err := orch.Shutdown(context.Background()); err != nil {
				fmt.Fprintf(errOut, "Error shutting down: %v\n", err)
			}
			// Cancel ctx only after workers have exited.
			cancel()
		})
	}
	defer shutdown()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	go func() {
		<-sigChan
		fmt.Fprintln(out, "\nShutting down Axis shell...")
		shutdown()
	}()

	if err := orch.Start(ctx); err != nil {
		return fmt.Errorf("failed to start orchestrator: %w", err)
	}

	fmt.Fprintln(out, "Axis shell started. Type 'help' for commands, 'exit' to quit.")
	scanner := bufio.NewScanner(os.Stdin)
	for {
		if !noPrompt {
			fmt.Fprint(out, "axis> ")
		}
		if !scanner.Scan() {
			fmt.Fprintln(out)
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
			printShellHelp(out)
		case "run":
			if len(commandArgs) != 1 {
				fmt.Fprintln(out, "Usage: run <task-id>")
				continue
			}
			if err := submitTask(commandArgs[0], "", ""); err != nil {
				fmt.Fprintf(out, "Could not submit task %s: %v\n", commandArgs[0], err)
				continue
			}
			fmt.Fprintf(out, "Task %s submitted. Try: status %s\n", commandArgs[0], commandArgs[0])
		case "status":
			if len(commandArgs) != 1 {
				fmt.Fprintln(out, "Usage: status <task-id>")
				continue
			}
			status, err := orch.GetTaskStatus(commandArgs[0])
			if err != nil {
				fmt.Fprintf(out, "Could not get status for task %s: %v\n", commandArgs[0], err)
				fmt.Fprintf(out, "If this is a new task, try: run %s\n", commandArgs[0])
				continue
			}
			fmt.Fprintf(out, "Task %s status: %s\n", commandArgs[0], status)
		case "dag":
			printDAG(out)
		case "resolve":
			if len(commandArgs) != 1 {
				fmt.Fprintln(out, "Usage: resolve <call-id>")
				continue
			}
			if err := orch.ResolveCall(commandArgs[0], map[string]any{"status": "resolved"}); err != nil {
				fmt.Fprintf(out, "Could not resolve call %s: %v\n", commandArgs[0], err)
				continue
			}
			fmt.Fprintf(out, "Call %s resolved.\n", commandArgs[0])
		case "tools":
			printTools(out)
		case "ask":
			handleShellAsk(commandArgs, out)
		case "judge":
			if err := runJudge(cmd, nil); err != nil {
				fmt.Fprintf(out, "Judgement error: %v\n", err)
			}
		case "exit", "quit":
			fmt.Fprintln(out, "Exiting Axis shell.")
			return nil
		default:
			fmt.Fprintf(out, "Unknown command: %s\n", command)
			fmt.Fprintln(out, "Type 'help' to see available commands.")
		}
	}
}

func printShellHelp(out io.Writer) {
	fmt.Fprintln(out, "Available commands:")
	fmt.Fprintln(out, "  help              Show this help message")
	fmt.Fprintln(out, "  run <task-id>     Submit a task")
	fmt.Fprintln(out, "  ask <prompt>      Preview a natural-language task")
	fmt.Fprintln(out, "  ask --submit <prompt> Submit a natural-language task")
	fmt.Fprintln(out, "  status <task-id>  Show task status")
	fmt.Fprintln(out, "  judge             Run self-judgement diagnostic")
	fmt.Fprintln(out, "  dag               Show dependency graph")
	fmt.Fprintln(out, "  resolve <call-id> Resolve a pending human call")
	fmt.Fprintln(out, "  tools              Show available tools")
	fmt.Fprintln(out, "  exit, quit        Shut down the shell")
}

func handleShellAsk(args []string, out io.Writer) {
	if len(args) == 0 {
		fmt.Fprintln(out, "Usage: ask [--submit] <prompt>")
		return
	}
	submit := false
	if args[0] == "--submit" {
		submit = true
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintln(out, "Usage: ask [--submit] <prompt>")
		return
	}
	prompt := strings.Join(args, " ")
	result, err := intent.NewDeterministicParser().Parse(context.Background(), intent.Request{Prompt: prompt, ContractID: "default"})
	if err != nil {
		fmt.Fprintf(out, "Could not parse prompt: %v\n", err)
		return
	}
	if submit {
		if err := orch.SubmitTask(result.Task); err != nil {
			fmt.Fprintf(out, "Could not submit task %s: %v\n", result.Task.TaskID, err)
			return
		}
		fmt.Fprintf(out, "Task %s submitted. Try: status %s\n", result.Task.TaskID, result.Task.TaskID)
		return
	}
	if err := renderTaskProposal(result.Task, out); err != nil {
		fmt.Fprintf(out, "Could not render task proposal: %v\n", err)
		return
	}
	fmt.Fprintln(out, "Not submitted. Use: ask --submit <prompt>")
}

func printTools(out io.Writer) {
	fmt.Fprintln(out, "Available tools:")
	fmt.Fprintln(out, "  bash           Execute shell commands")
	fmt.Fprintln(out, "  file_read      Read file contents")
	fmt.Fprintln(out, "  file_write     Write file contents")
	fmt.Fprintln(out, "  http_request   Make HTTP requests")
}

func printDAG(out io.Writer) {
	tasks := orch.GetAllTasks()
	deps := orch.GetDependencyGraph()
	if len(tasks) == 0 {
		fmt.Fprintln(out, "No tasks registered.")
		return
	}
	fmt.Fprintf(out, "%-20s %-12s %s\n", "TASK", "STATUS", "DEPENDS ON")
	fmt.Fprintf(out, "%-20s %-12s %s\n", "----", "------", "----------")
	for _, task := range tasks {
		depList := deps[task.TaskID]
		depStr := "(none)"
		if len(depList) > 0 {
			depStr = ""
			for i, d := range depList {
				if i > 0 {
					depStr += ", "
				}
				depStr += d
			}
		}
		fmt.Fprintf(out, "%-20s %-12s %s\n", task.TaskID, task.Status, depStr)
	}
}
