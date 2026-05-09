package provider

import (
	"fmt"
	"net/http"
	"time"
)

// providerConfig holds configuration for a model provider.
type providerConfig struct {
	model      string
	apiKey     string
	baseURL    string
	timeout    time.Duration
	maxRetries int
	httpClient *http.Client
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

// NewProvider creates a ModelProvider by name. Supported: "mock", "echo", "anthropic", "openai".
func NewProvider(name string, opts ...ProviderOption) (ModelProvider, error) {
	cfg := &providerConfig{
		model:      "default",
		apiKey:     "",
		baseURL:    "",
		timeout:    30 * time.Second,
		maxRetries: 3,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	// Apply options
	for _, opt := range opts {
		opt(cfg)
	}

	switch name {
	case "mock":
		return NewMockModelProvider(), nil
	case "echo":
		return NewEchoModelProvider(), nil
	case "anthropic":
		return newAnthropicProvider(cfg), nil
	case "openai":
		return newOpenAIProvider(cfg), nil
	default:
		return nil, fmt.Errorf("unknown provider %q: supported values are \"mock\", \"echo\", \"anthropic\", and \"openai\"", name)
	}
}
