package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/axis-cli/axis/internal/control"
	"github.com/axis-cli/axis/internal/model/providerconfig"
	"github.com/spf13/cobra"
)

type configDump struct {
	ProjectRoot     string `json:"project_root"`
	RuntimeProfile  string `json:"runtime_profile"`
	Provider        string `json:"provider"`
	ProfileName     string `json:"profile_name,omitempty"`
	Model           string `json:"model,omitempty"`
	BaseURL         string `json:"base_url,omitempty"`
	HasAPIKey       bool   `json:"has_api_key"`
	FallbackProfile string `json:"fallback_profile,omitempty"`
	RuntimeLive     bool   `json:"runtime_live"`
	RuntimeURL      string `json:"runtime_url,omitempty"`
}

func newConfigCommand(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Inspect resolved Axis configuration",
	}
	cmd.AddCommand(newConfigDumpCommand(app))
	return cmd
}

func newConfigDumpCommand(app *App) *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "dump",
		Short: "Print resolved effective runtime configuration (secrets redacted)",
		RunE: func(cmd *cobra.Command, args []string) error {
			dump, err := buildConfigDump(app)
			if err != nil {
				return err
			}
			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(dump)
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "project_root: %s\n", dump.ProjectRoot)
			fmt.Fprintf(out, "runtime_profile: %s\n", dump.RuntimeProfile)
			fmt.Fprintf(out, "provider: %s\n", dump.Provider)
			if dump.ProfileName != "" {
				fmt.Fprintf(out, "profile_name: %s\n", dump.ProfileName)
			}
			if dump.Model != "" {
				fmt.Fprintf(out, "model: %s\n", dump.Model)
			}
			if dump.BaseURL != "" {
				fmt.Fprintf(out, "base_url: %s\n", dump.BaseURL)
			}
			fmt.Fprintf(out, "has_api_key: %v\n", dump.HasAPIKey)
			if dump.FallbackProfile != "" {
				fmt.Fprintf(out, "fallback_profile: %s\n", dump.FallbackProfile)
			}
			fmt.Fprintf(out, "runtime_live: %v\n", dump.RuntimeLive)
			if dump.RuntimeURL != "" {
				fmt.Fprintf(out, "runtime_url: %s\n", dump.RuntimeURL)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Machine-readable JSON output")
	return cmd
}

func buildConfigDump(app *App) (*configDump, error) {
	if app == nil {
		return nil, fmt.Errorf("config dump: app is nil")
	}
	profileName, err := normalizeRuntimeProfile(app.runtimeProfile)
	if err != nil {
		return nil, err
	}
	root := app.resolvedRoot()
	dump := &configDump{
		ProjectRoot:    root,
		RuntimeProfile: profileName,
		Provider:       app.providerName,
		Model:          app.modelName,
	}

	if cfg, loadErr := providerconfig.NewStore(root).Load(); loadErr == nil {
		dump.FallbackProfile = cfg.FallbackProfile
		if app.providerName != "mock" || app.modelName != "" {
			if p, ok := cfg.Profiles[app.providerName]; ok && !p.Archived {
				dump.ProfileName = app.providerName
				dump.Provider = p.Provider
				dump.Model = p.Model
				dump.BaseURL = p.BaseURL
				dump.HasAPIKey = p.APIKey != ""
			}
		} else if p, ok := cfg.Active(); ok {
			dump.ProfileName = cfg.ActiveProfile
			dump.Provider = p.Provider
			dump.Model = p.Model
			dump.BaseURL = p.BaseURL
			dump.HasAPIKey = p.APIKey != ""
		}
	}
	if !dump.HasAPIKey {
		if envKey := envAPIKeyForProvider(dump.Provider); envKey != "" && os.Getenv(envKey) != "" {
			dump.HasAPIKey = true
		}
	}

	locator := control.NewRuntimeLocator(root)
	if rec, lerr := locator.Load(); lerr == nil {
		client := control.NewClient(locator, control.LocalHTTPClient())
		ctx, cancel := context.WithTimeout(context.Background(), runtimeProbeTimeout)
		if _, herr := client.Health(ctx); herr == nil {
			dump.RuntimeLive = true
			if rec.Protocol != "" && rec.Address != "" {
				dump.RuntimeURL = fmt.Sprintf("%s://%s", rec.Protocol, rec.Address)
			} else {
				dump.RuntimeURL = rec.Address
			}
		}
		cancel()
	}
	return dump, nil
}
