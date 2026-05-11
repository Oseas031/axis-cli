package provider

import (
	"encoding/json"
	"fmt"
	"os"
)

// providerLogEntry is a structured log record for provider HTTP calls.
type providerLogEntry struct {
	Provider     string  `json:"provider"`
	Method       string  `json:"method"`
	URL          string  `json:"url"`
	Status       int     `json:"status"`
	DurationMs   int64   `json:"duration_ms"`
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	CostUSD      float64 `json:"cost_usd,omitempty"`
}

// logProviderCall writes a structured JSON log line to stderr when AXIS_PROVIDER_LOG is set.
func logProviderCall(entry providerLogEntry) {
	if os.Getenv("AXIS_PROVIDER_LOG") != "1" {
		return
	}
	b, _ := json.Marshal(entry)
	fmt.Fprintln(os.Stderr, string(b))
}
