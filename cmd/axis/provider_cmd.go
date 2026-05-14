package main

import (
	"context"
	"fmt"

	"github.com/axis-cli/axis/internal/model/provider"
	"github.com/axis-cli/axis/internal/model/providerconfig"
	"github.com/spf13/cobra"
)

func newProviderCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "provider", Short: "Manage project-local model provider profiles"}
	cmd.AddCommand(newProviderAddCommand(), newProviderListCommand(), newProviderUseCommand(), newProviderStatusCommand(), newProviderRemoveCommand(), newProviderArchiveCommand(), newProviderTestCommand())
	return cmd
}

func newProviderAddCommand() *cobra.Command {
	var profile providerconfig.Profile
	cmd := &cobra.Command{
		Use:   "add [name]",
		Short: "Add or update a project-local provider profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile.Name = args[0]
			if err := providerconfig.NewStore(defaultApp.resolvedRoot()).AddProfile(profile); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Provider profile %s saved\n", profile.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&profile.Provider, "type", "", "Provider type: mock, echo, anthropic, openai")
	cmd.Flags().StringVar(&profile.APIKey, "api-key", "", "Provider API key stored in .axis/providers.json")
	cmd.Flags().StringVar(&profile.BaseURL, "base-url", "", "Provider API base URL")
	cmd.Flags().StringVar(&profile.Model, "model", "", "Default model")
	cmd.Flags().Float64Var(&profile.Temperature, "temperature", 0, "Default temperature")
	cmd.Flags().IntVar(&profile.MaxContext, "max-context", 0, "Maximum context window")
	_ = cmd.MarkFlagRequired("type")
	return cmd
}

func newProviderListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List project-local provider profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := providerconfig.NewStore(defaultApp.resolvedRoot()).Load()
			if err != nil {
				return err
			}
			for _, profile := range providerconfig.SortedProfiles(cfg) {
				active := " "
				if cfg.ActiveProfile == profile.Name {
					active = "*"
				}
				archived := ""
				if profile.Archived {
					archived = " archived"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s %s provider=%s model=%s base_url=%s%s\n", active, profile.Name, profile.Provider, profile.Model, profile.BaseURL, archived)
			}
			return nil
		},
	}
}

func newProviderUseCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "use [name]",
		Short: "Switch active provider profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			backup, err := providerconfig.NewStore(defaultApp.resolvedRoot()).Switch(args[0])
			if err != nil {
				return err
			}
			if backup != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Backup written: %s\n", backup)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Active provider profile: %s\n", args[0])
			return nil
		},
	}
}

func newProviderStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show active project-local provider state",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := providerconfig.NewStore(defaultApp.resolvedRoot()).Load()
			if err != nil {
				return err
			}
			profile, ok := cfg.Active()
			if !ok {
				fmt.Fprintln(cmd.OutOrStdout(), "No active provider profile")
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "active_profile: %s\nprovider: %s\nmodel: %s\nbase_url: %s\nupdated_at: %s\n", profile.Name, profile.Provider, profile.Model, profile.BaseURL, cfg.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
			fmt.Fprintln(cmd.OutOrStdout(), "Run 'axis provider test' to verify connectivity")
			return nil
		},
	}
}

func newProviderTestCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Test provider connectivity with a lightweight ping (consumes ~1-10 tokens)",
		RunE: func(cmd *cobra.Command, args []string) error {
			store := providerconfig.NewStore(defaultApp.resolvedRoot())
			cfg, err := store.Load()
			if err != nil {
				return err
			}

			var providerName string
			var opts []provider.ProviderOption
			if p, ok := cfg.Active(); ok {
				providerName = p.Provider
				opts = p.ProviderOptions()
			} else {
				providerName = defaultApp.providerName
				opts = defaultApp.providerOptions()
			}

			p, err := provider.NewProvider(providerName, opts...)
			if err != nil {
				return fmt.Errorf("failed to create provider %q: %w", providerName, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Testing provider %q...\n", providerName)
			resp, err := p.Execute(context.Background(), &provider.ModelRequest{
				ContractID: "provider-test",
				Input:      map[string]any{"message": "ping"},
			})
			if err != nil {
				return fmt.Errorf("provider test failed: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Provider %q is reachable\n", providerName)
			fmt.Fprintf(cmd.OutOrStdout(), "input_tokens: %d\noutput_tokens: %d\n", resp.InputTokens, resp.OutputTokens)
			if resp.CostEstimateUSD > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "estimated_cost_usd: %.6f\n", resp.CostEstimateUSD)
			}
			return nil
		},
	}
}

func newProviderRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove [name]",
		Short: "Remove a non-active provider profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := providerconfig.NewStore(defaultApp.resolvedRoot()).Remove(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Provider profile %s removed\n", args[0])
			return nil
		},
	}
}

func newProviderArchiveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "archive [name]",
		Short: "Archive a non-active provider profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := providerconfig.NewStore(defaultApp.resolvedRoot()).Archive(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Provider profile %s archived\n", args[0])
			return nil
		},
	}
}
