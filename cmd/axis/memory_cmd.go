package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"text/tabwriter"

	"github.com/axis-cli/axis/internal/memory/working"
	"github.com/axis-cli/axis/internal/project"
	"github.com/spf13/cobra"
)

// memoryRootDir returns the absolute path to .axis/memory/working for the
// current project.
func memoryWorkingDir() string {
	return filepath.Join(project.MemoryDir(project.MustResolveRoot()), "working")
}

func newMemoryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Manage Axis layered memory (Working Memory operations)",
	}
	cmd.AddCommand(
		newMemoryRetainCommand(),
		newMemoryReleaseCommand(),
		newMemoryListCommand(),
		newMemoryInspectCommand(),
		newMemoryCompactCommand(),
		newMemoryRecallCommand(),
		newMemoryImmunityCommand(),
		newMemoryDreamCommand(),
		newMemoryForgetCommand(),
	)
	return cmd
}

func openWorkingMemory() (*working.Engine, error) {
	return working.Open(memoryWorkingDir())
}

func newMemoryRetainCommand() *cobra.Command {
	var reason string
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "retain [bundle-id]",
		Short: "Retain a context bundle in the working set",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bundleID := args[0]
			eng, err := openWorkingMemory()
			if err != nil {
				return err
			}
			defer eng.Close()
			if err := eng.Retain(context.Background(), bundleID, reason); err != nil {
				return fmt.Errorf("retain %s: %w", bundleID, err)
			}
			items, _ := eng.List(context.Background())
			if jsonOut {
				return writeJSON(cmd.OutOrStdout(), map[string]any{
					"action":    "retain",
					"bundle_id": bundleID,
					"reason":    reason,
					"set_size":  len(items),
				})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Retained wm:bundle:%s. Working set: %d bundles.\n", bundleID, len(items))
			fmt.Fprintln(cmd.OutOrStdout(), "Next: axis memory list")
			return nil
		},
	}
	cmd.Flags().StringVar(&reason, "reason", "", "Reason for retaining (required)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit machine-readable JSON")
	_ = cmd.MarkFlagRequired("reason")
	return cmd
}

func newMemoryReleaseCommand() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "release [bundle-id]",
		Short: "Release a context bundle from the working set",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bundleID := args[0]
			eng, err := openWorkingMemory()
			if err != nil {
				return err
			}
			defer eng.Close()
			if err := eng.Release(context.Background(), bundleID); err != nil {
				return fmt.Errorf("release %s: %w", bundleID, err)
			}
			items, _ := eng.List(context.Background())
			if jsonOut {
				return writeJSON(cmd.OutOrStdout(), map[string]any{
					"action":    "release",
					"bundle_id": bundleID,
					"set_size":  len(items),
				})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Released wm:bundle:%s. Working set: %d bundles.\n", bundleID, len(items))
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit machine-readable JSON")
	return cmd
}

func newMemoryListCommand() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all retained bundles in the working set",
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := openWorkingMemory()
			if err != nil {
				return err
			}
			defer eng.Close()
			items, err := eng.List(context.Background())
			if err != nil {
				return err
			}
			if jsonOut {
				return writeJSON(cmd.OutOrStdout(), map[string]any{
					"count": len(items),
					"items": items,
				})
			}
			if len(items) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "Working set is empty.")
				fmt.Fprintln(cmd.OutOrStdout(), "Next: axis memory retain <bundle-id> --reason \"...\"")
				return nil
			}
			tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "BUNDLE ID\tRETAINED AT\tACCESS\tREASON")
			for _, it := range items {
				fmt.Fprintf(tw, "%s\t%s\t%d\t%s\n",
					it.BundleID,
					it.RetainedAt.Format("2006-01-02 15:04:05"),
					it.AccessCount,
					it.Reason,
				)
			}
			return tw.Flush()
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit machine-readable JSON")
	return cmd
}

func newMemoryInspectCommand() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "inspect [bundle-id]",
		Short: "Inspect full contents of a retained bundle",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := openWorkingMemory()
			if err != nil {
				return err
			}
			defer eng.Close()
			bundle, err := eng.GetBundle(context.Background(), args[0])
			if err != nil {
				return fmt.Errorf("inspect %s: %w", args[0], err)
			}
			if jsonOut {
				return writeJSON(cmd.OutOrStdout(), bundle)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Bundle:       %s\n", bundle.BundleID)
			fmt.Fprintf(cmd.OutOrStdout(), "Task:         %s\n", bundle.TaskID)
			fmt.Fprintf(cmd.OutOrStdout(), "Contract:     %s\n", bundle.ContractID)
			fmt.Fprintf(cmd.OutOrStdout(), "Goal:         %s\n", bundle.Goal)
			fmt.Fprintf(cmd.OutOrStdout(), "Reason:       %s\n", bundle.Reason)
			fmt.Fprintf(cmd.OutOrStdout(), "Access count: %d\n", bundle.AccessCount)
			fmt.Fprintf(cmd.OutOrStdout(), "Retained at:  %s\n", bundle.RetainedAt.Format("2006-01-02 15:04:05"))
			fmt.Fprintf(cmd.OutOrStdout(), "Packets:      %d\n", len(bundle.Packets))
			for _, p := range bundle.Packets {
				fmt.Fprintf(cmd.OutOrStdout(), "  - [%s] %s (relevance=%.2f) %s\n", p.Type, p.Source, p.Relevance, p.Reason)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit machine-readable JSON")
	return cmd
}

func newMemoryCompactCommand() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "compact",
		Short: "Rebuild snapshot and index from authoritative in-memory state",
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := openWorkingMemory()
			if err != nil {
				return err
			}
			defer eng.Close()
			items, _ := eng.List(context.Background())
			if err := eng.Compact(); err != nil {
				return fmt.Errorf("compact: %w", err)
			}
			if jsonOut {
				return writeJSON(cmd.OutOrStdout(), map[string]any{
					"action":       "compact",
					"record_count": len(items),
				})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Compacted %d records. history.jsonl preserved.\n", len(items))
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit machine-readable JSON")
	return cmd
}

func newMemoryRecallCommand() *cobra.Command {
	var limit int
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "recall [query]",
		Short: "Recall relevant packets from retained bundles by keyword",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := openWorkingMemory()
			if err != nil {
				return err
			}
			defer eng.Close()
			hits, err := eng.Recall(context.Background(), args[0], limit)
			if err != nil {
				return err
			}
			if jsonOut {
				return writeJSON(cmd.OutOrStdout(), map[string]any{
					"query": args[0],
					"hits":  hits,
				})
			}
			if len(hits) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No matches for %q.\n", args[0])
				return nil
			}
			tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "BUNDLE\tTYPE\tSOURCE\tRELEVANCE")
			for _, h := range hits {
				fmt.Fprintf(tw, "%s\t%s\t%s\t%.2f\n", h.BundleID, h.Type, h.Source, h.Relevance)
			}
			return tw.Flush()
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of hits")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit machine-readable JSON")
	return cmd
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
