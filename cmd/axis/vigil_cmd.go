package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/axis-cli/axis/internal/project"
	"github.com/axis-cli/axis/internal/vigil"
	"github.com/spf13/cobra"
)

func newVigilCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "vigil", Short: "Track work items and priorities"}
	cmd.AddCommand(
		newVigilResumeCommand(),
		newVigilListCommand(),
		newVigilAddCommand(),
		newVigilStartCommand(),
		newVigilDoneCommand(),
		newVigilShowCommand(),
		newVigilTriageCommand(),
	)
	return cmd
}

func vigilStore() *vigil.Store {
	return vigil.NewStore(project.MustResolveRoot())
}

func newVigilResumeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "resume",
		Short: "Show current work context",
		RunE: func(cmd *cobra.Command, args []string) error {
			store := vigilStore()
			items, err := store.Load()
			if err != nil {
				return err
			}
			now := time.Now()
			cutoff := now.Add(-24 * time.Hour)

			var inProgress, recentDone, topPending []*vigil.Item
			for _, it := range items {
				switch {
				case it.Status == vigil.StatusInProgress:
					inProgress = append(inProgress, it)
				case it.Status == vigil.StatusCompleted && it.CompletedAt != nil && it.CompletedAt.After(cutoff):
					recentDone = append(recentDone, it)
				case it.Status == vigil.StatusPending:
					if len(topPending) == 0 || it.Priority < topPending[0].Priority {
						topPending = []*vigil.Item{it}
					} else if it.Priority == topPending[0].Priority {
						topPending = append(topPending, it)
					}
				}
			}

			if len(inProgress) == 0 && len(recentDone) == 0 && len(topPending) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No active work. Use: axis vigil add \"title\"")
				return nil
			}

			out := cmd.OutOrStdout()
			if len(inProgress) > 0 {
				fmt.Fprintln(out, "## In Progress")
				for _, it := range inProgress {
					fmt.Fprintf(out, "  %s  %s  [%s]\n", it.ID, it.Title, it.Priority)
				}
			}
			if len(recentDone) > 0 {
				fmt.Fprintln(out, "## Recently Completed (24h)")
				for _, it := range recentDone {
					fmt.Fprintf(out, "  %s  %s\n", it.ID, it.Title)
				}
			}
			if len(topPending) > 0 {
				fmt.Fprintln(out, "## Top Pending")
				for _, it := range topPending {
					fmt.Fprintf(out, "  %s  %s  [%s]\n", it.ID, it.Title, it.Priority)
				}
			}
			return nil
		},
	}
}

func newVigilListCommand() *cobra.Command {
	var priority, tag, status string
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List work items",
		RunE: func(cmd *cobra.Command, args []string) error {
			store := vigilStore()
			items, err := store.Load()
			if err != nil {
				return err
			}
			var filtered []*vigil.Item
			for _, it := range items {
				if status == "" {
					if it.Status != vigil.StatusPending && it.Status != vigil.StatusInProgress && it.Status != vigil.StatusStale {
						continue
					}
				} else if string(it.Status) != status {
					continue
				}
				if priority != "" && it.Priority != priority {
					continue
				}
				if tag != "" && !containsTag(it.Tags, tag) {
					continue
				}
				filtered = append(filtered, it)
			}
			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				if filtered == nil {
					filtered = []*vigil.Item{}
				}
				return enc.Encode(filtered)
			}
			if len(filtered) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No items found.")
				return nil
			}
			for _, it := range filtered {
				fmt.Fprintf(cmd.OutOrStdout(), "%s  %s  [%s]  %s\n", it.ID, it.Title, it.Priority, it.Status)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&priority, "priority", "", "Filter by priority")
	cmd.Flags().StringVar(&tag, "tag", "", "Filter by tag")
	cmd.Flags().StringVar(&status, "status", "", "Filter by status (default: active)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func newVigilAddCommand() *cobra.Command {
	var priority, origin, notes string
	var tags, dependsOn []string
	cmd := &cobra.Command{
		Use:   "add <title>",
		Short: "Add a new work item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := vigilStore()
			now := time.Now()
			item := &vigil.Item{
				ID:        vigil.GenerateID(args[0], now),
				Title:     args[0],
				Priority:  priority,
				Status:    vigil.StatusPending,
				Tags:      tags,
				Origin:    vigil.Origin{Type: origin},
				DependsOn: dependsOn,
				Notes:     notes,
				CreatedAt: now,
				History:   []vigil.StatusChange{},
			}
			if err := store.Add(item); err != nil {
				return fmt.Errorf("failed to add item: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Added %s — %s\nNext: axis vigil start %s\n", item.ID, item.Title, item.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&priority, "priority", "P1", "Priority level")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "Tags (repeatable)")
	cmd.Flags().StringVar(&origin, "origin", "", "Origin type")
	cmd.Flags().StringSliceVar(&dependsOn, "depends-on", nil, "Dependencies (repeatable)")
	cmd.Flags().StringVar(&notes, "notes", "", "Notes")
	return cmd
}

func newVigilStartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start <id>",
		Short: "Start working on an item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := vigilStore()
			item, err := store.Get(args[0])
			if err != nil {
				return fmt.Errorf("failed to start %s: %w", args[0], err)
			}
			if item.Status == vigil.StatusInProgress {
				fmt.Fprintf(cmd.OutOrStdout(), "%s already in progress\n", item.ID)
				return nil
			}
			old := item.Status
			item.Status = vigil.StatusInProgress
			now := time.Now()
			item.StartedAt = &now
			item.History = append(item.History, vigil.StatusChange{From: old, To: vigil.StatusInProgress, At: now})
			if err := store.Update(item); err != nil {
				return fmt.Errorf("failed to start %s: %w", args[0], err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Started %s — %s\nNext: axis vigil done %s\n", item.ID, item.Title, item.ID)
			return nil
		},
	}
}

func newVigilDoneCommand() *cobra.Command {
	var commit string
	cmd := &cobra.Command{
		Use:   "done <id>",
		Short: "Mark an item as completed",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := vigilStore()
			item, err := store.Get(args[0])
			if err != nil {
				return fmt.Errorf("failed to complete %s: %w", args[0], err)
			}
			old := item.Status
			item.Status = vigil.StatusCompleted
			now := time.Now()
			item.CompletedAt = &now
			if commit != "" {
				item.CommitHash = commit
			}
			item.History = append(item.History, vigil.StatusChange{From: old, To: vigil.StatusCompleted, At: now})
			if err := store.Update(item); err != nil {
				return fmt.Errorf("failed to complete %s: %w", args[0], err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Completed %s — %s\nNext: axis vigil resume\n", item.ID, item.Title)
			return nil
		},
	}
	cmd.Flags().StringVar(&commit, "commit", "", "Associated commit hash")
	return cmd
}

func newVigilShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show full item details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := vigilStore()
			item, err := store.Get(args[0])
			if err != nil {
				return fmt.Errorf("failed to show %s: %w", args[0], err)
			}
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(item)
		},
	}
}

func newVigilTriageCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "triage",
		Short: "Run triage rules on all items",
		RunE: func(cmd *cobra.Command, args []string) error {
			store := vigilStore()
			items, err := store.Load()
			if err != nil {
				return err
			}
			result, active, toArchive := vigil.Triage(items, time.Now())
			if err := store.Save(active); err != nil {
				return fmt.Errorf("failed to save after triage: %w", err)
			}
			if err := store.Archive(toArchive); err != nil {
				return fmt.Errorf("failed to archive: %w", err)
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Triage complete. Staled: %d, Upgraded: %d, Archived: %d\n",
				len(result.Staled), len(result.Upgraded), len(result.Archived))
			if len(result.Staled) > 0 {
				fmt.Fprintf(out, "  Staled: %v\n", result.Staled)
			}
			if len(result.Upgraded) > 0 {
				fmt.Fprintf(out, "  Upgraded to P0: %v\n", result.Upgraded)
			}
			if len(result.Archived) > 0 {
				fmt.Fprintf(out, "  Archived: %v\n", result.Archived)
			}
			return nil
		},
	}
}

func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}
