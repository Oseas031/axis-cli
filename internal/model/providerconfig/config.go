package providerconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/axis-cli/axis/internal/model/provider"
)

const (
	DefaultConfigDir  = ".axis"
	DefaultConfigFile = "providers.json"
)

type Config struct {
	ActiveProfile string             `json:"active_profile"`
	Profiles      map[string]Profile `json:"profiles"`
	Routes        map[string]Route   `json:"routes"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

type Profile struct {
	Name        string    `json:"name"`
	Provider    string    `json:"provider"`
	APIKey      string    `json:"api_key,omitempty"`
	BaseURL     string    `json:"base_url,omitempty"`
	Model       string    `json:"model"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxContext  int       `json:"max_context,omitempty"`
	Archived    bool      `json:"archived,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Route struct {
	Profile string `json:"profile"`
	Model   string `json:"model,omitempty"`
}

type Store struct {
	Root string
}

func NewStore(root string) *Store {
	if root == "" {
		root = "."
	}
	return &Store{Root: root}
}

func (s *Store) ConfigPath() string {
	return filepath.Join(s.Root, DefaultConfigDir, DefaultConfigFile)
}

func (s *Store) BackupDir() string {
	return filepath.Join(s.Root, DefaultConfigDir, "backups")
}

func NewConfig() *Config {
	return &Config{
		Profiles:  map[string]Profile{},
		Routes:    defaultRoutes(""),
		UpdatedAt: time.Now(),
	}
}

func defaultRoutes(profile string) map[string]Route {
	return map[string]Route{
		"reasoning": {Profile: profile},
		"code":      {Profile: profile},
		"writing":   {Profile: profile},
		"tool":      {Profile: profile},
	}
}

func (s *Store) Load() (*Config, error) {
	path := s.ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewConfig(), nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]Profile{}
	}
	if cfg.Routes == nil {
		cfg.Routes = defaultRoutes(cfg.ActiveProfile)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (s *Store) Save(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("provider config is nil")
	}
	cfg.UpdatedAt = time.Now()
	if err := cfg.Validate(); err != nil {
		return err
	}
	// Auto-backup existing config before overwriting.
	if _, err := os.Stat(s.ConfigPath()); err == nil {
		if _, err := s.Backup(); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Dir(s.ConfigPath()), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.ConfigPath() + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.ConfigPath())
}

func (s *Store) Backup() (string, error) {
	path := s.ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	if err := os.MkdirAll(s.BackupDir(), 0o700); err != nil {
		return "", err
	}
	backup := filepath.Join(s.BackupDir(), "providers-"+time.Now().Format("20060102-150405.000000")+".json")
	if err := os.WriteFile(backup, data, 0o600); err != nil {
		return "", err
	}
	return backup, nil
}

func (s *Store) AddProfile(profile Profile) error {
	cfg, err := s.Load()
	if err != nil {
		return err
	}
	if profile.Name == "" {
		return fmt.Errorf("profile name is required")
	}
	profile.UpdatedAt = time.Now()
	cfg.Profiles[profile.Name] = profile
	if cfg.ActiveProfile == "" && !profile.Archived {
		cfg.ActiveProfile = profile.Name
		cfg.Routes = defaultRoutes(profile.Name)
	}
	return s.Save(cfg)
}

func (s *Store) Switch(name string) (string, error) {
	cfg, err := s.Load()
	if err != nil {
		return "", err
	}
	profile, ok := cfg.Profiles[name]
	if !ok {
		return "", fmt.Errorf("provider profile %q not found", name)
	}
	if profile.Archived {
		return "", fmt.Errorf("provider profile %q is archived", name)
	}
	backup, err := s.Backup()
	if err != nil {
		return "", err
	}
	cfg.ActiveProfile = name
	cfg.Routes = defaultRoutes(name)
	return backup, s.Save(cfg)
}

func (s *Store) Remove(name string) error {
	cfg, err := s.Load()
	if err != nil {
		return err
	}
	if _, ok := cfg.Profiles[name]; !ok {
		return fmt.Errorf("provider profile %q not found", name)
	}
	if cfg.ActiveProfile == name {
		return fmt.Errorf("cannot remove active provider profile %q", name)
	}
	if _, err := s.Backup(); err != nil {
		return err
	}
	delete(cfg.Profiles, name)
	return s.Save(cfg)
}

func (s *Store) Archive(name string) error {
	cfg, err := s.Load()
	if err != nil {
		return err
	}
	profile, ok := cfg.Profiles[name]
	if !ok {
		return fmt.Errorf("provider profile %q not found", name)
	}
	if cfg.ActiveProfile == name {
		return fmt.Errorf("cannot archive active provider profile %q", name)
	}
	if _, err := s.Backup(); err != nil {
		return err
	}
	profile.Archived = true
	profile.UpdatedAt = time.Now()
	cfg.Profiles[name] = profile
	return s.Save(cfg)
}

func (cfg *Config) Validate() error {
	if cfg.Profiles == nil {
		return fmt.Errorf("profiles is required")
	}
	for name, profile := range cfg.Profiles {
		if name == "" || profile.Name == "" {
			return fmt.Errorf("profile name is required")
		}
		if name != profile.Name {
			return fmt.Errorf("profile key %q does not match profile name %q", name, profile.Name)
		}
		if profile.Provider == "" {
			return fmt.Errorf("provider is required for profile %q", name)
		}
		if profile.Model == "" && profile.Provider != "mock" && profile.Provider != "echo" {
			return fmt.Errorf("model is required for profile %q", name)
		}
	}
	if cfg.ActiveProfile != "" {
		profile, ok := cfg.Profiles[cfg.ActiveProfile]
		if !ok {
			return fmt.Errorf("active profile %q not found", cfg.ActiveProfile)
		}
		if profile.Archived {
			return fmt.Errorf("active profile %q is archived", cfg.ActiveProfile)
		}
	}
	for routeName, route := range cfg.Routes {
		if route.Profile == "" {
			continue
		}
		profile, ok := cfg.Profiles[route.Profile]
		if !ok {
			return fmt.Errorf("route %q references unknown profile %q", routeName, route.Profile)
		}
		if profile.Archived {
			return fmt.Errorf("route %q references archived profile %q", routeName, route.Profile)
		}
	}
	return nil
}

func (cfg *Config) Active() (Profile, bool) {
	if cfg == nil || cfg.ActiveProfile == "" {
		return Profile{}, false
	}
	profile, ok := cfg.Profiles[cfg.ActiveProfile]
	return profile, ok
}

func (p Profile) ProviderOptions() []provider.ProviderOption {
	opts := []provider.ProviderOption{}
	if p.Model != "" {
		opts = append(opts, provider.WithModel(p.Model))
	}
	if p.APIKey != "" {
		opts = append(opts, provider.WithAPIKey(p.APIKey))
	}
	if p.BaseURL != "" {
		opts = append(opts, provider.WithBaseURL(p.BaseURL))
	}
	if p.Temperature != 0 {
		opts = append(opts, provider.WithTemperature(p.Temperature))
	}
	if p.MaxContext > 0 {
		opts = append(opts, provider.WithMaxContext(p.MaxContext))
	}
	return opts
}

func SortedProfiles(cfg *Config) []Profile {
	if cfg == nil {
		return nil
	}
	profiles := make([]Profile, 0, len(cfg.Profiles))
	for _, profile := range cfg.Profiles {
		profiles = append(profiles, profile)
	}
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].Name < profiles[j].Name })
	return profiles
}
