package main

import (
	"fmt"

	"github.com/axis-cli/axis/internal/guarantee"
	"github.com/spf13/cobra"
)

func newGuaranteeCommand(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "guarantee",
		Short: "Manage system guarantees",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all registered guarantees",
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, g := range app.guarantees.List() {
				level := "Hard"
				if g.Level == guarantee.LevelSoft {
					level = "Soft"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s — %s\n", level, g.ID, g.Description)
			}
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "verify",
		Short: "Verify all guarantees",
		RunE: func(cmd *cobra.Command, args []string) error {
			violations := app.guarantees.Verify()
			if len(violations) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "All guarantees satisfied.")
				return nil
			}
			for _, v := range violations {
				level := "Hard"
				if v.Level == guarantee.LevelSoft {
					level = "Soft"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "[%s VIOLATION] %s: %v\n", level, v.GuaranteeID, v.Error)
			}
			for _, v := range violations {
				if v.Level == guarantee.LevelHard {
					return fmt.Errorf("hard guarantee violation detected")
				}
			}
			return nil
		},
	})

	return cmd
}
