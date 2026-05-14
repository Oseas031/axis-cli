package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/axis-cli/axis/internal/contextpack"
	"github.com/axis-cli/axis/internal/control"
	"github.com/axis-cli/axis/internal/intent"
	"github.com/axis-cli/axis/internal/types"
	"github.com/spf13/cobra"
)

func newAskCommand() *cobra.Command {
	var submit bool
	var dryRun bool
	var contractID string
	var taskID string
	var readStdin bool
	var withContext bool
	cmd := &cobra.Command{
		Use:   "ask [prompt]",
		Short: "Convert natural language into an Axis task proposal",
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
			if submit && dryRun {
				return fmt.Errorf("--submit and --dry-run cannot both be set")
			}
			if submit {
				// Natural language tasks default to LLM Agent execution
				if result.Task.Metadata == nil {
					result.Task.Metadata = make(map[string]string)
				}
				if result.Task.Metadata[types.TaskMetadataKeyExecutor] == "" {
					result.Task.Metadata[types.TaskMetadataKeyExecutor] = types.ExecutorTypeAgent
				}
				if withContext {
					bundle, err := assembleContextForTask(result.Task)
					if err != nil {
						return err
					}
					artifact, err := contextpack.DefaultRegistry.Register(bundle)
					if err != nil {
						return err
					}
					if err := contextpack.AttachReadinessMetadata(result.Task, artifact); err != nil {
						return err
					}
				}
				client := control.NewClient(control.NewRuntimeLocator(defaultApp.resolvedRoot()), http.DefaultClient)
				if _, err := client.SubmitTask(context.Background(), result.Task); err != nil {
					return fmt.Errorf("failed to submit task: %w", err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Task %s submitted. Try: status %s\n", result.Task.TaskID, result.Task.TaskID)
				return nil
			}
			if err := renderTaskProposal(result.Task, cmd.OutOrStdout()); err != nil {
				return err
			}
			if withContext {
				bundle, err := assembleContextForTask(result.Task)
				if err != nil {
					return err
				}
				if err := renderContextBundle(bundle, cmd.OutOrStdout()); err != nil {
					return err
				}
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Not submitted. Use --submit to schedule this task.")
			return nil
		},
	}
	cmd.Flags().BoolVar(&submit, "submit", false, "Submit the generated task")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview the generated task without submitting")
	cmd.Flags().StringVar(&contractID, "contract", "default", "Contract ID for the generated task")
	cmd.Flags().StringVar(&taskID, "task-id", "", "Task ID for the generated task")
	cmd.Flags().BoolVar(&readStdin, "stdin", false, "Read prompt from stdin")
	cmd.Flags().BoolVar(&withContext, "with-context", false, "Preview an adaptive context bundle with the generated task")
	return cmd
}

// renderTaskProposal renders an AgentTask as JSON for preview.
// Used by both CLI ask and shell ask to ensure consistent output.
func renderTaskProposal(task *types.AgentTask, writer io.Writer) error {
	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(writer, "Task proposal:\n%s\n", data)
	return err
}
