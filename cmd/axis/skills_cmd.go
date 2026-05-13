package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/axis-cli/axis/internal/project"
	"github.com/axis-cli/axis/internal/skills"
	"github.com/spf13/cobra"
)

func newSkillsCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "skills", Short: "Manage on-demand knowledge skills"}
	cmd.AddCommand(newSkillsListCommand(), newSkillsShowCommand(), newSkillsValidateCommand(), newSkillsCreateCommand())
	return cmd
}

func skillsLoader() *skills.Loader {
	return skills.NewLoader(project.SkillsDir(project.MustResolveRoot()))
}

func newSkillsListCommand() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			loader := skillsLoader()
			metas, err := loader.Discover(context.Background())
			if err != nil {
				return err
			}
			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(metas)
			}
			if len(metas) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No skills found in .axis/skills/")
				return nil
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tDESCRIPTION")
			for _, m := range metas {
				fmt.Fprintf(w, "%s\t%s\n", m.Name, m.Description)
			}
			return w.Flush()
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func newSkillsShowCommand() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "show <skill-name>",
		Short: "Show full skill content",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			loader := skillsLoader()
			skill, err := loader.Load(context.Background(), args[0])
			if err != nil {
				return err
			}
			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(skill)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Name: %s\nDescription: %s\n---\n%s", skill.Meta.Name, skill.Meta.Description, skill.Content)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func newSkillsValidateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "validate [skill-name]",
		Short: "Validate skill format",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			loader := skillsLoader()
			if len(args) > 0 {
				if err := loader.Validate(context.Background(), args[0]); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Skill %s is valid\n", args[0])
				return nil
			}
			metas, err := loader.Discover(context.Background())
			if err != nil {
				return err
			}
			for _, m := range metas {
				if err := loader.Validate(context.Background(), m.Name); err != nil {
					return fmt.Errorf("skill %s: %w", m.Name, err)
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "All %d skills valid\n", len(metas))
			return nil
		},
	}
}

func newSkillsCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create <skill-name>",
		Short: "Create a new skill directory with template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := skills.ValidateSkillName(name); err != nil {
				return err
			}
			dir := project.MustResolveRoot()
			skillDir := filepath.Join(project.SkillsDir(dir), name)
			if err := os.MkdirAll(skillDir, 0o755); err != nil {
				return err
			}
			template := fmt.Sprintf("---\nname: %s\ndescription: TODO - add description\n---\n\n# %s\n\nAdd your instructions here.\n", name, name)
			skillFile := filepath.Join(skillDir, "SKILL.md")
			if err := os.WriteFile(skillFile, []byte(template), 0o600); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created: %s\n", skillFile)
			return nil
		},
	}
}
