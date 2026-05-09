package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// anthropicMessage represents a message in the Anthropic API format.
type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// anthropicTool represents a tool in the Anthropic API format.
type anthropicTool struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	InputSchema struct {
		Type string `json:"type"`
	} `json:"input_schema"`
}

// anthropicResponse is the response format from the Anthropic Messages API.
type anthropicResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text,omitempty"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// Execute calls the Anthropic Messages API.
func (p *AnthropicProvider) Execute(ctx context.Context, req *ModelRequest) (*ModelResponse, error) {
	baseURL := p.config.baseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}

	ar := anthropicRequest{
		Model:     p.config.model,
		MaxTokens: 4096,
	}

	// Convert history to Anthropic message format
	for _, msg := range req.History {
		// Build content string
		content := msg.Content
		if msg.Role == "assistant" && len(msg.ToolCalls) > 0 {
			// Convert tool calls to text
			for _, tc := range msg.ToolCalls {
				content += fmt.Sprintf("\n[ToolCall: %s(%v)]", tc.Name, tc.Input)
			}
		}
		ar.Messages = append(ar.Messages, anthropicMessage{
			Role:    msg.Role,
			Content: content,
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

	// Add system prompt from contract if available (using first tool description)
	if len(req.Tools) > 0 {
		ar.System = "You have access to the following tools. Use them when needed."
		for _, t := range req.Tools {
			ar.Tools = append(ar.Tools, anthropicTool{
				Name:        t.Name,
				Description: t.Description,
			})
			ar.Tools[len(ar.Tools)-1].InputSchema.Type = "object"
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

	resp, err := p.config.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var arResp anthropicResponse
	if err := json.Unmarshal(respBody, &arResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Extract output text
	output := make(map[string]any)
	if len(arResp.Content) > 0 && arResp.Content[0].Type == "text" {
		output["text"] = arResp.Content[0].Text
	}

	return &ModelResponse{
		Output:       output,
		InputTokens:  arResp.Usage.InputTokens,
		OutputTokens: arResp.Usage.OutputTokens,
	}, nil
}
