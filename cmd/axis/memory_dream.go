package main

import (
	"context"
	"fmt"
	"time"

	"github.com/axis-cli/axis/internal/memory/horizon"
	"github.com/axis-cli/axis/internal/memory/longterm"
	"github.com/axis-cli/axis/internal/project"
	"github.com/spf13/cobra"
)

func newMemoryDreamCommand() *cobra.Command {
	var since string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "dream",
		Short: "Replay recent failed events, cluster patterns, and distill into long-horizon memory",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := project.MustResolveRoot()
			eventsDir := project.EventsDir(root)
			memoryDir := project.MemoryDir(root)

			eventStore, err := longterm.Open(eventsDir)
			if err != nil {
				return fmt.Errorf("dream: open events: %w", err)
			}
			defer eventStore.Close()

			hStore := horizon.NewStore(memoryDir)
			if err := hStore.Init(); err != nil {
				return fmt.Errorf("dream: init memory: %w", err)
			}

			opts := horizon.DreamOptions{}
			if since != "" {
				dur, err := time.ParseDuration(since)
				if err != nil {
					return fmt.Errorf("dream: invalid --since %q: %w", since, err)
				}
				opts.Since = time.Now().Add(-dur)
			}

			result, err := horizon.Dream(context.Background(), eventStore, hStore, opts)
			if err != nil {
				return err
			}

			if jsonOut {
				return writeJSON(cmd.OutOrStdout(), result)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Dream complete.\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  Events read:    %d\n", result.EventsRead)
			fmt.Fprintf(cmd.OutOrStdout(), "  Clusters found: %d\n", result.Clusters)
			fmt.Fprintf(cmd.OutOrStdout(), "  Patterns new:   %d\n", result.PatternsNew)
			for _, id := range result.PatternIDs {
				fmt.Fprintf(cmd.OutOrStdout(), "    → %s\n", id)
			}
			if result.PatternsNew == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "  No recurring failure patterns detected.")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&since, "since", "", "Replay events from this duration ago (e.g. 1h, 4h, 1d)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit machine-readable JSON")
	return cmd
}
