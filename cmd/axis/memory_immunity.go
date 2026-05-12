package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/axis-cli/axis/internal/memory/immunity"
	"github.com/axis-cli/axis/internal/memory/longterm"
	"github.com/axis-cli/axis/internal/project"
	"github.com/spf13/cobra"
)

// memoryLongtermDir returns the directory hosting the long-term event log.
func memoryLongtermDir() string {
	return filepath.Join(project.MemoryDir(project.MustResolveRoot()), "longterm")
}

func openLongtermStore() (*longterm.FileStore, error) {
	return longterm.Open(memoryLongtermDir())
}

func newMemoryImmunityCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "immunity",
		Short: "Manage Immunity Memory (failure-as-asset layer)",
	}
	cmd.AddCommand(
		newImmunityPromoteCommand(),
		newImmunityListCommand(),
		newImmunityShowCommand(),
		newImmunityForgetCommand(),
	)
	return cmd
}

func newImmunityPromoteCommand() *cobra.Command {
	var cause, class, by string
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "promote [task-id]",
		Short: "Promote a failed task into an Immunity record",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			es, err := openLongtermStore()
			if err != nil {
				return err
			}
			defer es.Close()
			store := immunity.NewStore(es)
			rec, err := store.Promote(context.Background(), immunity.PromoteInput{
				SourceTaskID: args[0],
				Cause:        cause,
				FailureClass: immunity.FailureClass(class),
				PromotedBy:   by,
			})
			if err != nil {
				return fmt.Errorf("promote %s: %w", args[0], err)
			}
			if jsonOut {
				return writeJSON(cmd.OutOrStdout(), rec)
			}
			fmt.Fprintf(cmd.OutOrStdout(),
				"Promoted %s to immunity record %s.\n  cause:  %s\n  class:  %s\n  next:   axis memory immunity show %s\n",
				rec.SourceTaskID, rec.ImmunityID, rec.Cause, rec.FailureClass, rec.ImmunityID,
			)
			return nil
		},
	}
	cmd.Flags().StringVar(&cause, "cause", "", "One-line failure cause (required)")
	cmd.Flags().StringVar(&class, "class", "", "Failure class (e.g. failure.provider.timeout); auto-derived from event payload if empty")
	cmd.Flags().StringVar(&by, "by", "", "Promoter actor identifier (required)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit machine-readable JSON")
	_ = cmd.MarkFlagRequired("cause")
	_ = cmd.MarkFlagRequired("by")
	return cmd
}

func newImmunityListCommand() *cobra.Command {
	var class, since string
	var includeDeprecated bool
	var limit int
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Immunity records (newest first)",
		RunE: func(cmd *cobra.Command, args []string) error {
			es, err := openLongtermStore()
			if err != nil {
				return err
			}
			defer es.Close()
			store := immunity.NewStore(es)

			filter := immunity.ListFilter{
				Class:             immunity.FailureClass(class),
				IncludeDeprecated: includeDeprecated,
				Limit:             limit,
			}
			if since != "" {
				d, err := time.ParseDuration(since)
				if err != nil {
					return fmt.Errorf("--since: %w", err)
				}
				t := time.Now().Add(-d)
				filter.Since = &t
			}

			records, err := store.List(context.Background(), filter)
			if err != nil {
				return err
			}
			if jsonOut {
				return writeJSON(cmd.OutOrStdout(), map[string]any{
					"count":   len(records),
					"records": records,
				})
			}
			if len(records) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No immunity records.")
				fmt.Fprintln(cmd.OutOrStdout(), "Next: axis memory immunity promote <task-id> --cause \"...\" --by <actor>")
				return nil
			}
			tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "IMMUNITY ID\tCLASS\tPROMOTED AT\tDEPRECATED\tCAUSE")
			for _, r := range records {
				dep := "no"
				if r.Deprecated {
					dep = "yes"
				}
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
					r.ImmunityID, r.FailureClass,
					r.PromotedAt.Format("2006-01-02 15:04:05"),
					dep, truncate(r.Cause, 60),
				)
			}
			return tw.Flush()
		},
	}
	cmd.Flags().StringVar(&class, "class", "", "Filter by failure class")
	cmd.Flags().StringVar(&since, "since", "", "Only records newer than duration (e.g. 24h)")
	cmd.Flags().BoolVar(&includeDeprecated, "deprecated", false, "Include forgotten records")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum records to return")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit machine-readable JSON")
	return cmd
}

func newImmunityShowCommand() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "show [immunity-id]",
		Short: "Show a single Immunity record",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			es, err := openLongtermStore()
			if err != nil {
				return err
			}
			defer es.Close()
			store := immunity.NewStore(es)
			rec, err := store.Show(context.Background(), args[0])
			if err != nil {
				return fmt.Errorf("show %s: %w", args[0], err)
			}
			if jsonOut {
				return writeJSON(cmd.OutOrStdout(), rec)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Immunity ID:   %s\n", rec.ImmunityID)
			fmt.Fprintf(cmd.OutOrStdout(), "Source task:   %s\n", rec.SourceTaskID)
			fmt.Fprintf(cmd.OutOrStdout(), "Failure class: %s\n", rec.FailureClass)
			fmt.Fprintf(cmd.OutOrStdout(), "Cause:         %s\n", rec.Cause)
			fmt.Fprintf(cmd.OutOrStdout(), "Promoted by:   %s\n", rec.PromotedBy)
			fmt.Fprintf(cmd.OutOrStdout(), "Promoted at:   %s\n", rec.PromotedAt.Format(time.RFC3339))
			fmt.Fprintf(cmd.OutOrStdout(), "Sig hash:      %s\n", rec.SignatureHash)
			fmt.Fprintf(cmd.OutOrStdout(), "Intent:        %s\n", rec.Signature.IntentKind)
			if len(rec.Signature.ContractToolAllow) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Tools:         %s\n", strings.Join(rec.Signature.ContractToolAllow, ", "))
			}
			if rec.Deprecated {
				fmt.Fprintf(cmd.OutOrStdout(), "DEPRECATED:    %s\n", rec.DeprecateReason)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit machine-readable JSON")
	return cmd
}

func newImmunityForgetCommand() *cobra.Command {
	var reason, by string
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "forget [immunity-id]",
		Short: "Soft-mark an Immunity record as deprecated",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			es, err := openLongtermStore()
			if err != nil {
				return err
			}
			defer es.Close()
			store := immunity.NewStore(es)
			if err := store.Forget(context.Background(), args[0], reason, by); err != nil {
				return fmt.Errorf("forget %s: %w", args[0], err)
			}
			if jsonOut {
				return writeJSON(cmd.OutOrStdout(), map[string]any{
					"action":       "forget",
					"immunity_id":  args[0],
					"reason":       reason,
					"forgotten_by": by,
				})
			}
			fmt.Fprintf(cmd.OutOrStdout(),
				"Forgot immunity record %s.\n  reason: %s\n  next:   axis memory immunity show %s\n",
				args[0], reason, args[0],
			)
			return nil
		},
	}
	cmd.Flags().StringVar(&reason, "reason", "", "Reason for forgetting (required)")
	cmd.Flags().StringVar(&by, "by", "", "Actor identifier (required)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit machine-readable JSON")
	_ = cmd.MarkFlagRequired("reason")
	_ = cmd.MarkFlagRequired("by")
	return cmd
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}
