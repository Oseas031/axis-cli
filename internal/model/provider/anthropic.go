package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

// AnthropicProvider implements ModelProvider for Anthropic's API.
type AnthropicProvider struct {
	config *providerConfig
}

// newAnthropicProvider creates a new Anthropic provider.
func newAnthropicProvider(cfg *providerConfig) *AnthropicProvider {
	return &AnthropicProvider{config: cfg}
}

// anthropicRequest is the request format for the Anthropic Messages API.
type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
	Tools     []anthropicTool    `json:"tools,omitempty"`
	System    string             `json:"system,omitempty"`
}

// anthropicContentBlock represents a single block in Anthropic message content.
type anthropicContentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Content   string          `json:"content,omitempty"`
}

// anthropicMessage represents a message in the Anthropic API format.
// Content may be a string or []anthropicContentBlock.
type anthropicMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

// anthropicToolSchema matches the JSON Schema object Anthropic expects for tool input.
type anthropicToolSchema struct {
	Type       string         `json:"type"`
	Properties map[string]any `json:"properties,omitempty"`
	Required   []string       `json:"required,omitempty"`
}

// anthropicTool represents a tool in the Anthropic API format.
type anthropicTool struct {
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	InputSchema anthropicToolSchema `json:"input_schema"`
}

// anthropicUsage tracks token consumption.
type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// anthropicResponse is the response format from the Anthropic Messages API.
type anthropicResponse struct {
	ID           string                  `json:"id"`
	Type         string                  `json:"type"`
	Role         string                  `json:"role"`
	Content      []anthropicContentBlock `json:"content"`
	Model        string                  `json:"model"`
	StopReason   string                  `json:"stop_reason"`
	StopSequence string                  `json:"stop_sequence"`
	Usage        anthropicUsage          `json:"usage"`
}

// Execute calls the Anthropic Messages API.
func (p *AnthropicProvider) Execute(ctx context.Context, req *ModelRequest) (*ModelResponse, error) {
	start := time.Now()
	baseURL := p.config.baseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}

	ar := anthropicRequest{
		Model:     p.config.model,
		MaxTokens: p.config.maxContext,
	}
	if ar.MaxTokens <= 0 {
		ar.MaxTokens = 4096
	}

	// Convert history to Anthropic message format
	for _, msg := range req.History {
		if msg.Role == "assistant" && len(msg.ToolCalls) > 0 {
			blocks := []anthropicContentBlock{}
			if msg.Content != "" {
				blocks = append(blocks, anthropicContentBlock{Type: "text", Text: msg.Content})
			}
			for _, tc := range msg.ToolCalls {
				raw, _ := json.Marshal(tc.Input)
				blocks = append(blocks, anthropicContentBlock{Type: "tool_use", ID: tc.ID, Name: tc.Name, Input: raw})
			}
			ar.Messages = append(ar.Messages, anthropicMessage{Role: msg.Role, Content: blocks})
			continue
		}
		if msg.Role == "tool" {
			blocks := []anthropicContentBlock{
				{Type: "tool_result", ToolUseID: msg.ToolCallID, Content: msg.Content},
			}
			ar.Messages = append(ar.Messages, anthropicMessage{Role: "user", Content: blocks})
			continue
		}
		ar.Messages = append(ar.Messages, anthropicMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Add current input as a user message
	inputContent := ""
	for k, v := range req.Input {
		inputContent += fmt.Sprintf("%s: %v\n", k, v)
	}
	ar.Messages = append(ar.Messages, anthropicMessage{
		Role:    "user",
		Content: inputContent,
	})

	// Add tools with proper JSON Schema
	if len(req.Tools) > 0 {
		ar.System = "You have access to the following tools. Use them when needed."
		if req.SystemPrompt != "" {
			ar.System = req.SystemPrompt
		}
		for _, t := range req.Tools {
			schema := anthropicToolSchema{Type: "object"}
			schema.Properties = make(map[string]any)
			var required []string
			for _, field := range t.Parameters {
				prop := map[string]any{"type": string(field.Type)}
				if field.Description != "" {
					prop["description"] = field.Description
				}
				if len(field.Enum) > 0 {
					prop["enum"] = field.Enum
				}
				schema.Properties[field.Name] = prop
				if field.Required {
					required = append(required, field.Name)
				}
			}
			if len(required) > 0 {
				schema.Required = required
			}
			ar.Tools = append(ar.Tools, anthropicTool{
				Name:        t.Name,
				Description: t.Description,
				InputSchema: schema,
			})
		}
	}

	body, err := json.Marshal(ar)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.config.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	// Retry loop with exponential backoff: retry on 5xx, 429, and network errors.
	var lastErr error
	var respBody []byte
	for attempt := 0; attempt <= p.config.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second // exponential: 1s, 2s, 4s, 8s...
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			// Recreate request with fresh body for retry
			httpReq, err = http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/messages", bytes.NewReader(body))
			if err != nil {
				return nil, fmt.Errorf("failed to create request: %w", err)
			}
			httpReq.Header.Set("Content-Type", "application/json")
			httpReq.Header.Set("x-api-key", p.config.apiKey)
			httpReq.Header.Set("anthropic-version", "2023-06-01")
		}

		resp, err := p.config.httpClient.Do(httpReq)
		if err != nil {
			lastErr = fmt.Errorf("request failed (attempt %d/%d): %w", attempt+1, p.config.maxRetries+1, err)
			continue
		}

		respBody, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode == http.StatusOK {
			break
		}

		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			lastErr = fmt.Errorf("API error (status %d, attempt %d/%d): %s", resp.StatusCode, attempt+1, p.config.maxRetries+1, string(respBody))
			continue
		}

		// 4xx errors (except 429) are not retried.
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}
	if lastErr != nil && len(respBody) == 0 {
		return nil, fmt.Errorf("max retries (%d) exceeded: %w", p.config.maxRetries, lastErr)
	}

	var arResp anthropicResponse
	if err := json.Unmarshal(respBody, &arResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Extract output text and tool calls
	output := make(map[string]any)
	var toolCalls []types.ToolCall
	var textParts []string
	for _, block := range arResp.Content {
		switch block.Type {
		case "text":
			textParts = append(textParts, block.Text)
		case "tool_use":
			var input map[string]any
			if len(block.Input) > 0 {
				_ = json.Unmarshal(block.Input, &input)
			}
			toolCalls = append(toolCalls, types.ToolCall{
				ID:    block.ID,
				Name:  block.Name,
				Input: input,
			})
		}
	}
	if len(textParts) > 0 {
		output["text"] = textParts[0]
	}

	cost := estimateCost(p.config.model, arResp.Usage.InputTokens, arResp.Usage.OutputTokens)
	logProviderCall(providerLogEntry{
		Provider:     "anthropic",
		Method:       "POST",
		URL:          baseURL + "/v1/messages",
		Status:       200,
		DurationMs:   time.Since(start).Milliseconds(),
		InputTokens:  arResp.Usage.InputTokens,
		OutputTokens: arResp.Usage.OutputTokens,
		CostUSD:      cost,
	})
	return &ModelResponse{
		Output:          output,
		ToolCalls:       toolCalls,
		InputTokens:     arResp.Usage.InputTokens,
		OutputTokens:    arResp.Usage.OutputTokens,
		CostEstimateUSD: cost,
	}, nil
}
