package providerconfig

import (
	"os"
	"testing"
)

func TestStore_AddSwitchArchiveRemove(t *testing.T) {
	store := NewStore(t.TempDir())

	if err := store.AddProfile(Profile{Name: "local-openai", Provider: "openai", APIKey: "secret", BaseURL: "https://api.openai.com", Model: "gpt-4o-mini", Temperature: 0.2, MaxContext: 128000}); err != nil {
		t.Fatalf("AddProfile failed: %v", err)
	}
	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.ActiveProfile != "local-openai" {
		t.Fatalf("expected first profile to become active, got %s", cfg.ActiveProfile)
	}
	if cfg.Routes["code"].Profile != "local-openai" {
		t.Fatalf("expected route mapping to active profile, got %#v", cfg.Routes["code"])
	}

	if err := store.AddProfile(Profile{Name: "claude", Provider: "anthropic", APIKey: "secret", Model: "claude-3-5-sonnet-20241022"}); err != nil {
		t.Fatalf("AddProfile failed: %v", err)
	}
	backup, err := store.Switch("claude")
	if err != nil {
		t.Fatalf("Switch failed: %v", err)
	}
	if backup == "" {
		t.Fatal("expected switch to create a backup")
	}
	cfg, err = store.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.ActiveProfile != "claude" {
		t.Fatalf("expected claude active, got %s", cfg.ActiveProfile)
	}
	if cfg.Routes["reasoning"].Profile != "claude" {
		t.Fatalf("expected routes to switch globally, got %#v", cfg.Routes)
	}

	if err := store.Archive("local-openai"); err != nil {
		t.Fatalf("Archive failed: %v", err)
	}
	cfg, err = store.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if !cfg.Profiles["local-openai"].Archived {
		t.Fatal("expected local-openai archived")
	}
	if err := store.Remove("local-openai"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
}

func TestConfig_ValidateRejectsBrokenActiveProfile(t *testing.T) {
	cfg := NewConfig()
	cfg.ActiveProfile = "missing"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected missing active profile to fail validation")
	}
}

func TestProfile_ProviderOptions(t *testing.T) {
	profile := Profile{Name: "p", Provider: "openai", APIKey: "secret", BaseURL: "http://localhost", Model: "gpt-test", Temperature: 0.2, MaxContext: 128000}
	if got := len(profile.ProviderOptions()); got != 5 {
		t.Fatalf("expected 5 provider options, got %d", got)
	}
}

func TestStore_AddProfileCreatesBackup(t *testing.T) {
	store := NewStore(t.TempDir())
	// Seed an initial config so AddProfile overwrites an existing file.
	if err := store.Save(&Config{Profiles: map[string]Profile{"seed": {Name: "seed", Provider: "mock", Model: "m"}}}); err != nil {
		t.Fatalf("initial Save failed: %v", err)
	}
	if err := store.AddProfile(Profile{Name: "local-openai", Provider: "openai", APIKey: "secret", BaseURL: "https://api.openai.com", Model: "gpt-4o-mini"}); err != nil {
		t.Fatalf("AddProfile failed: %v", err)
	}
	entries, err := os.ReadDir(store.BackupDir())
	if err != nil {
		t.Fatalf("failed to read backup dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected AddProfile to auto-backup existing config")
	}
}

func TestStore_BackupTimestampUnique(t *testing.T) {
	store := NewStore(t.TempDir())
	if err := store.Save(&Config{Profiles: map[string]Profile{"a": {Name: "a", Provider: "mock", Model: "m"}}}); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	b1, err := store.Backup()
	if err != nil {
		t.Fatalf("Backup 1 failed: %v", err)
	}
	b2, err := store.Backup()
	if err != nil {
		t.Fatalf("Backup 2 failed: %v", err)
	}
	if b1 == b2 {
		t.Fatalf("expected unique backup names, got duplicate: %s", b1)
	}
}

func TestSortedProfiles_NilConfig(t *testing.T) {
	profiles := SortedProfiles(nil)
	if profiles != nil {
		t.Fatalf("expected nil for nil config, got %v", profiles)
	}
}

func TestConfig_ValidateRejectsDanglingRoute(t *testing.T) {
	cfg := NewConfig()
	cfg.ActiveProfile = "default"
	cfg.Profiles["default"] = Profile{Name: "default", Provider: "mock", Model: "m"}
	cfg.Routes["code"] = Route{Profile: "missing"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected dangling route to fail validation")
	}
}

func TestConfig_ValidateRejectsArchivedRouteProfile(t *testing.T) {
	cfg := NewConfig()
	cfg.ActiveProfile = "default"
	cfg.Profiles["default"] = Profile{Name: "default", Provider: "mock", Model: "m"}
	cfg.Profiles["old"] = Profile{Name: "old", Provider: "mock", Model: "m", Archived: true}
	cfg.Routes["code"] = Route{Profile: "old"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected archived route profile to fail validation")
	}
}

func TestConfig_ValidateAcceptsOpenAICompatibleProviders(t *testing.T) {
	for _, providerName := range []string{"deepseek", "minimax"} {
		cfg := NewConfig()
		cfg.ActiveProfile = providerName
		cfg.Profiles[providerName] = Profile{Name: providerName, Provider: providerName, APIKey: "secret", Model: "test-model"}
		if err := cfg.Validate(); err != nil {
			t.Fatalf("expected %s profile to validate: %v", providerName, err)
		}
	}
}
