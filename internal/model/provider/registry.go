package provider

import (
	"fmt"
	"net/http"
	"time"
)

// providerConfig holds configuration for a model provider.
type providerConfig struct {
	model       string
	apiKey      string
	baseURL     string
	timeout     time.Duration
	maxRetries  int
	temperature float64
	maxContext  int
	httpClient  *http.Client
}

// ProviderOption is a functional option for configuring a provider.
type ProviderOption func(*providerConfig)

// WithModel sets the model name.
func WithModel(model string) ProviderOption {
	return func(c *providerConfig) {
		c.model = model
	}
}

// WithAPIKey sets the API key.
func WithAPIKey(apiKey string) ProviderOption {
	return func(c *providerConfig) {
		c.apiKey = apiKey
	}
}

// WithBaseURL sets the base URL for the API.
func WithBaseURL(baseURL string) ProviderOption {
	return func(c *providerConfig) {
		c.baseURL = baseURL
	}
}

// WithTimeout sets the request timeout.
func WithTimeout(timeout time.Duration) ProviderOption {
	return func(c *providerConfig) {
		c.timeout = timeout
	}
}

// WithMaxRetries sets the maximum number of retries.
func WithMaxRetries(maxRetries int) ProviderOption {
	return func(c *providerConfig) {
		c.maxRetries = maxRetries
	}
}

// WithTemperature sets the sampling temperature.
func WithTemperature(temperature float64) ProviderOption {
	return func(c *providerConfig) {
		c.temperature = temperature
	}
}

// WithMaxContext sets the maximum context tokens / max tokens.
func WithMaxContext(maxContext int) ProviderOption {
	return func(c *providerConfig) {
		c.maxContext = maxContext
	}
}

// NewProvider creates a ModelProvider by name. Supported: "mock", "echo", "anthropic", "openai", "deepseek", "minimax".
func NewProvider(name string, opts ...ProviderOption) (ModelProvider, error) {
	cfg := &providerConfig{
		model:      "default",
		apiKey:     "",
		baseURL:    "",
		timeout:    30 * time.Second,
		maxRetries: 5,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	// Apply options
	for _, opt := range opts {
		opt(cfg)
	}

	// Sync httpClient timeout with configured timeout
	if cfg.httpClient.Timeout != cfg.timeout {
		cfg.httpClient.Timeout = cfg.timeout
	}

	// Real providers require an API key.
	needsKey := false
	switch name {
	case "mock":
		return NewMockModelProvider(), nil
	case "echo":
		return NewEchoModelProvider(), nil
	case "anthropic":
		if cfg.apiKey == "" {
			return nil, fmt.Errorf("anthropic: API key missing. Set ANTHROPIC_API_KEY or run 'axis provider add --type anthropic --api-key <key>'")
		}
		return newAnthropicProvider(cfg), nil
	case "openai":
		needsKey = true
	case "deepseek":
		needsKey = true
		if cfg.baseURL == "" {
			cfg.baseURL = "https://api.deepseek.com"
		}
	case "minimax":
		needsKey = true
		if cfg.baseURL == "" {
			cfg.baseURL = "https://api.minimaxi.com/v1"
		}
	default:
		return nil, fmt.Errorf("unknown provider %q: supported values are \"mock\", \"echo\", \"anthropic\", \"openai\", \"deepseek\", and \"minimax\"", name)
	}
	if needsKey && cfg.apiKey == "" {
		return nil, fmt.Errorf("%s: API key missing. Set the corresponding environment variable or run 'axis provider add --type %s --api-key <key>'", name, name)
	}
	return newOpenAIProvider(cfg), nil
}
