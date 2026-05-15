package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
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
		newVigilInstallHookCommand(),
	)
	return cmd
}

func vigilStore() *vigil.Store {
	return vigil.NewStore(project.MustResolveRoot())
}

func vigilLocker() *vigil.Locker {
	root := project.MustResolveRoot()
	return vigil.NewLocker(root + "/.axis/vigil")
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

			// Auto-triage on resume (silent archive + stale marking)
			result, active, toArchive := vigil.Triage(items, now)
			if len(result.Archived) > 0 || len(result.Staled) > 0 || len(result.Upgraded) > 0 {
				_ = store.Save(active)
				if len(toArchive) > 0 {
					_ = store.Archive(toArchive)
				}
				items = active
			}

			cutoff := now.Add(-24 * time.Hour)

			var inProgress, recentDone []*vigil.Item
			var pending []*vigil.Item
			for _, it := range items {
				switch {
				case it.Status == vigil.StatusInProgress:
					inProgress = append(inProgress, it)
				case it.Status == vigil.StatusCompleted && it.CompletedAt != nil && it.CompletedAt.After(cutoff):
					recentDone = append(recentDone, it)
				case it.Status == vigil.StatusPending || it.Status == vigil.StatusStale:
					pending = append(pending, it)
				}
			}

			// Sort pending by priority (P0 < P1 < P2 lexicographically)
			for i := 0; i < len(pending); i++ {
				for j := i + 1; j < len(pending); j++ {
					if pending[j].Priority < pending[i].Priority {
						pending[i], pending[j] = pending[j], pending[i]
					}
				}
			}
			if len(pending) > 10 {
				pending = pending[:10]
			}

			if len(inProgress) == 0 && len(recentDone) == 0 && len(pending) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No active work. Use: axis vigil add \"title\"")
				return nil
			}

			out := cmd.OutOrStdout()
			if len(inProgress) > 0 {
				fmt.Fprintln(out, "## In Progress")
				locker := vigilLocker()
				for _, it := range inProgress {
					if locker.IsLocked(it.ID) {
						info, _ := locker.Read(it.ID)
						fmt.Fprintf(out, "  %s  %s  [%s] 🔒 PID %d\n", it.ID, it.Title, it.Priority, info.PID)
					} else {
						fmt.Fprintf(out, "  %s  %s  [%s]\n", it.ID, it.Title, it.Priority)
					}
				}
			}
			if len(recentDone) > 0 {
				fmt.Fprintln(out, "## Recently Completed (24h)")
				for _, it := range recentDone {
					fmt.Fprintf(out, "  %s  %s\n", it.ID, it.Title)
				}
			}
			if len(pending) > 0 {
				fmt.Fprintln(out, "## Top Pending")
				for _, it := range pending {
					fmt.Fprintf(out, "  %s  %s  [%s]\n", it.ID, it.Title, it.Priority)
				}
				if len(inProgress) == 0 {
					fmt.Fprintf(out, "\nHint: start an item before working on it: axis vigil start %s\n", pending[0].ID)
				}
			}
			return nil
		},
	}
}

func newVigilListCommand() *cobra.Command {
	var priority, tag, status string
	var jsonOut, all bool
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
				if !all {
					if status == "" {
						if it.Status != vigil.StatusPending && it.Status != vigil.StatusInProgress && it.Status != vigil.StatusStale {
							continue
						}
					} else if string(it.Status) != status {
						continue
					}
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
			locker := vigilLocker()
			for _, it := range filtered {
				lock := ""
				if it.Status == vigil.StatusInProgress && locker.IsLocked(it.ID) {
					lock = " 🔒"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s  %s  [%s]  %s%s\n", it.ID, it.Title, it.Priority, it.Status, lock)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "Show all items regardless of status")
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
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "Tags (repeatable, comma-separated or multiple --tag)")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "Alias for --tag")
	cmd.Flags().MarkHidden("tags")
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
			locker := vigilLocker()
			item, err := store.Get(args[0])
			if err != nil {
				return fmt.Errorf("failed to start %s: %w", args[0], err)
			}
			if item.Status == vigil.StatusInProgress {
				if locker.IsLocked(args[0]) {
					info, _ := locker.Read(args[0])
					fmt.Fprintf(cmd.OutOrStdout(), "%s already in progress (locked by PID %d)\n", item.ID, info.PID)
					return nil
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s already in progress\n", item.ID)
				return nil
			}
			if err := locker.Lock(args[0], fmt.Sprintf("pid-%d", os.Getpid())); err != nil {
				return fmt.Errorf("cannot start %s: %w", args[0], err)
			}
			old := item.Status
			item.Status = vigil.StatusInProgress
			now := time.Now()
			item.StartedAt = &now
			item.History = append(item.History, vigil.StatusChange{From: old, To: vigil.StatusInProgress, At: now})
			if err := store.Update(item); err != nil {
				_ = locker.Unlock(args[0])
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
			locker := vigilLocker()
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
			_ = locker.Unlock(args[0])
			fmt.Fprintf(cmd.OutOrStdout(), "Completed %s — %s\nNext: axis vigil resume\n", item.ID, item.Title)
			printDocSyncReminders(cmd, commit)
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

// docSyncMap maps source directories to documentation that may need updating.
var docSyncMap = []struct {
	prefix string
	doc    string
}{
	{"internal/vigil/", ".axis/skills/vigil/SKILL.md"},
	{"internal/skills/", ".axis/skills/"},
	{"internal/memory/", "docs/architecture/semantic-boundaries.md (memory section)"},
	{"internal/kernel/", "docs/architecture/semantic-boundaries.md (kernel section)"},
	{"internal/contextpack/", "docs/architecture/semantic-boundaries.md (contextpack section)"},
	{"internal/agent/", "docs/architecture/semantic-boundaries.md (agent section)"},
	{"internal/model/", "docs/architecture/semantic-boundaries.md (provider/tool section)"},
	{"internal/control/", "docs/status/current-progress.md"},
	{"internal/evolution/", "docs/status/current-progress.md"},
	{"cmd/axis/", "README.md (CLI Commands table)"},
	{"tools/axis-gui/", "tools/axis-gui/README.md"},
	{"tools/axis-up/", "tools/axis-up/README.md"},
}

func printDocSyncReminders(cmd *cobra.Command, commit string) {
	var gitArgs []string
	if commit != "" {
		gitArgs = []string{"diff", "--name-only", commit + "^", commit}
	} else {
		gitArgs = []string{"diff", "--name-only", "HEAD"}
	}
	out, err := exec.Command("git", gitArgs...).Output()
	if err != nil || len(out) == 0 {
		return
	}
	files := strings.Split(strings.TrimSpace(string(out)), "\n")
	seen := map[string]bool{}
	for _, f := range files {
		f = strings.ReplaceAll(f, "\\", "/")
		for _, m := range docSyncMap {
			if strings.HasPrefix(f, m.prefix) && !seen[m.doc] {
				seen[m.doc] = true
			}
		}
	}
	if len(seen) > 0 {
		w := cmd.OutOrStdout()
		fmt.Fprintln(w, "⚠ Doc sync reminder:")
		for doc := range seen {
			fmt.Fprintf(w, "  → %s\n", doc)
		}
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


func newVigilInstallHookCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "install-hook",
		Short: "Install git post-commit hook for auto-completion",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := project.MustResolveRoot()
			hookDir := root + "/.git/hooks"
			hookPath := hookDir + "/post-commit"

			hookContent := `#!/bin/bash
# Axis Vigil: auto-complete work items referenced in commit messages.
msg=$(git log -1 --format=%B 2>/dev/null)
[ -z "$msg" ] && exit 0
ids=$(echo "$msg" | grep -oE 'vigil:[a-z0-9-]+' | sed 's/vigil://')
[ -z "$ids" ] && exit 0
commit=$(git rev-parse HEAD 2>/dev/null)
for id in $ids; do
    axis vigil done "$id" --commit "$commit" 2>/dev/null || true
done
`
			if err := os.MkdirAll(hookDir, 0o755); err != nil {
				return fmt.Errorf("failed to create hooks dir: %w", err)
			}
			if err := os.WriteFile(hookPath, []byte(hookContent), 0o755); err != nil {
				return fmt.Errorf("failed to write hook: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Installed post-commit hook at %s\n", hookPath)
			return nil
		},
	}
}
