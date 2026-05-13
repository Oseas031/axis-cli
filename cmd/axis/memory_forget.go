package main

import (
	"fmt"

	"github.com/axis-cli/axis/internal/memory/horizon"
	"github.com/axis-cli/axis/internal/project"
	"github.com/spf13/cobra"
)

func newMemoryForgetCommand() *cobra.Command {
	var dryRun bool
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "forget",
		Short: "Archive or delete old narrative memories (7d archive, 30d delete)",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := project.MustResolveRoot()
			memoryDir := project.MemoryDir(root)

			store := horizon.NewStore(memoryDir)
			if err := store.Init(); err != nil {
				return fmt.Errorf("forget: init memory: %w", err)
			}

			result, err := horizon.Forget(store, dryRun)
			if err != nil {
				return err
			}

			if jsonOut {
				return writeJSON(cmd.OutOrStdout(), result)
			}

			if dryRun {
				fmt.Fprintln(cmd.OutOrStdout(), "[dry-run] No files modified.")
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Forget complete.\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  Archived: %d\n", result.Archived)
			fmt.Fprintf(cmd.OutOrStdout(), "  Deleted:  %d\n", result.Deleted)
			fmt.Fprintf(cmd.OutOrStdout(), "  Skipped:  %d\n", result.Skipped)
			for _, d := range result.Details {
				fmt.Fprintf(cmd.OutOrStdout(), "    %s\n", d)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview what would happen without modifying files")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit machine-readable JSON")
	return cmd
}
