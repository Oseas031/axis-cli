package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/axis-cli/axis/internal/contextpack"
	"github.com/axis-cli/axis/internal/intent"
	"github.com/axis-cli/axis/internal/types"
	"github.com/spf13/cobra"
)

func newContextCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "context", Short: "Preview adaptive context bundles"}
	cmd.AddCommand(newContextPreviewCommand(), newContextInspectCommand(), newContextPreflightCommand(), newContextIndexCommand())
	return cmd
}

func newContextPreflightCommand() *cobra.Command {
	var strict bool
	cmd := &cobra.Command{
		Use:   "preflight [task-id]",
		Short: "Check whether a task has traceable context readiness",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			initOrchestrator()
			task := findSubmittedTask(args[0])
			result := contextpack.Preflight(task, contextpack.DefaultRegistry)
			if err := renderPreflightResult(result, cmd.OutOrStdout()); err != nil {
				return err
			}
			if strict && result.Status != contextpack.PreflightStatusReady {
				return fmt.Errorf("context readiness preflight failed for task %s: %s", args[0], result.Status)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&strict, "strict", false, "Return an error unless context readiness is ready")
	return cmd
}

func newContextInspectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect [bundle-id]",
		Short: "Inspect a registered context readiness record",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			record, err := contextpack.DefaultRegistry.Inspect(args[0])
			if err != nil {
				return err
			}
			return renderReadinessRecord(record, cmd.OutOrStdout())
		},
	}
}

func newContextPreviewCommand() *cobra.Command {
	var taskID string
	var contractID string
	var readStdin bool
	var maxPackets int
	var maxBytes int
	cmd := &cobra.Command{
		Use:   "preview [prompt]",
		Short: "Preview a rule-based context bundle without executing a task",
		Args: func(cmd *cobra.Command, args []string) error {
			if readStdin {
				return nil
			}
			if len(args) == 0 {
				return fmt.Errorf("prompt is required unless --stdin is set")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			prompt := strings.Join(args, " ")
			if readStdin {
				data, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return fmt.Errorf("failed to read stdin: %w", err)
				}
				prompt = string(data)
			}
			result, err := intent.NewDeterministicParser().Parse(context.Background(), intent.Request{Prompt: prompt, ContractID: contractID, TaskID: taskID})
			if err != nil {
				return err
			}
			budget := contextpack.DefaultBudget()
			if maxPackets > 0 {
				budget.MaxPackets = maxPackets
			}
			if maxBytes > 0 {
				budget.MaxBytes = maxBytes
			}
			opts := []contextpack.Option{contextpack.WithBudget(budget)}
			mgr := contextpack.NewIndexManager()
			if err := mgr.Load("."); err == nil {
				if idx := mgr.Index(); idx != nil && len(idx.Chunks) > 0 {
					opts = append(opts, contextpack.WithIndex(idx))
				}
			}
			bundle, err := contextpack.NewAssembler(opts...).Assemble(result.Task)
			if err != nil {
				return err
			}
			return renderContextBundle(bundle, cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVar(&contractID, "contract", "default", "Contract ID for the generated task")
	cmd.Flags().StringVar(&taskID, "task-id", "", "Task ID for the generated task")
	cmd.Flags().BoolVar(&readStdin, "stdin", false, "Read prompt from stdin")
	cmd.Flags().IntVar(&maxPackets, "max-packets", 0, "Maximum selected context packets")
	cmd.Flags().IntVar(&maxBytes, "max-bytes", 0, "Maximum selected context bytes")
	return cmd
}

func newContextIndexCommand() *cobra.Command {
	var rebuild bool
	var update bool
	var status bool
	var root string
	cmd := &cobra.Command{
		Use:   "index",
		Short: "Manage local TF-IDF context index",
		RunE: func(cmd *cobra.Command, args []string) error {
			count := 0
			if rebuild {
				count++
			}
			if update {
				count++
			}
			if status {
				count++
			}
			if count != 1 {
				return fmt.Errorf("exactly one of --rebuild, --update, or --status is required")
			}

			mgr := contextpack.NewIndexManager()
			var result *contextpack.IndexStatus
			var err error
			switch {
			case rebuild:
				result, err = mgr.Rebuild(root)
			case update:
				result, err = mgr.Update(root)
			case status:
				result = mgr.Status(root)
			}
			if err != nil {
				return err
			}
			return renderIndexStatus(result, cmd.OutOrStdout())
		},
	}
	cmd.Flags().BoolVar(&rebuild, "rebuild", false, "Rebuild the index from scratch")
	cmd.Flags().BoolVar(&update, "update", false, "Incrementally update the index based on mtime changes")
	cmd.Flags().BoolVar(&status, "status", false, "Show index status")
	cmd.Flags().StringVar(&root, "root", ".", "Project root to index")
	return cmd
}

func renderIndexStatus(status *contextpack.IndexStatus, writer io.Writer) error {
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(writer, "Context index:\n%s\n", data)
	return err
}

func renderContextBundle(bundle *contextpack.ContextBundle, writer io.Writer) error {
	data, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(writer, "Context bundle preview:\n%s\n", data)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(writer, "Not executed. Context preview does not submit or run tasks.")
	return err
}

func renderReadinessRecord(record contextpack.ReadinessRecord, writer io.Writer) error {
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(writer, "Context readiness record:\n%s\n", data)
	return err
}

func renderPreflightResult(result contextpack.PreflightResult, writer io.Writer) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(writer, "Context readiness preflight:\n%s\n", data)
	return err
}

func assembleContextForTask(task *types.AgentTask) (*contextpack.ContextBundle, error) {
	opts := []contextpack.Option{}
	mgr := contextpack.NewIndexManager()
	if err := mgr.Load("."); err == nil {
		if idx := mgr.Index(); idx != nil && len(idx.Chunks) > 0 {
			opts = append(opts, contextpack.WithIndex(idx))
		}
	}
	return contextpack.NewAssembler(opts...).Assemble(task)
}

func findSubmittedTask(taskID string) *types.AgentTask {
	if orch == nil {
		return nil
	}
	for _, task := range orch.GetAllTasks() {
		if task.TaskID == taskID {
			return task
		}
	}
	return nil
}
