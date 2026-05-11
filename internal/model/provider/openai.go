package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

// OpenAIProvider implements ModelProvider for OpenAI's API.
type OpenAIProvider struct {
	config *providerConfig
}

// newOpenAIProvider creates a new OpenAI provider.
func newOpenAIProvider(cfg *providerConfig) *OpenAIProvider {
	return &OpenAIProvider{config: cfg}
}

// openaiRequest is the request format for the OpenAI Chat Completions API.
type openaiRequest struct {
	Model       string          `json:"model"`
	Messages    []openaiMessage `json:"messages"`
	Tools       []openaiTool    `json:"tools,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
}

// openaiMessage represents a message in the OpenAI API format.
type openaiMessage struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	ToolCalls []struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		Function struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		} `json:"function"`
	} `json:"tool_calls,omitempty"`
	ToolCallID string `json:"tool_call_id,omitempty"`
}

// openaiTool represents a tool in the OpenAI API format.
type openaiTool struct {
	Type     string             `json:"type"`
	Function openaiToolFunction `json:"function"`
}

// openaiResponse is the response format from the OpenAI Chat Completions API.
type openaiResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role      string `json:"role"`
			Content   string `json:"content"`
			ToolCalls []struct {
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// openaiToolParamProperty represents a property in OpenAI tool parameters.
type openaiToolParamProperty struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// openaiToolParams represents tool parameters in OpenAI API format.
type openaiToolParams struct {
	Type       string                             `json:"type"`
	Properties map[string]openaiToolParamProperty `json:"properties"`
	Required   []string                           `json:"required"`
}

// openaiToolFunction represents a function in OpenAI tool format.
type openaiToolFunction struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Parameters  openaiToolParams `json:"parameters"`
}

// Execute calls the OpenAI Chat Completions API.
func (p *OpenAIProvider) Execute(ctx context.Context, req *ModelRequest) (*ModelResponse, error) {
	start := time.Now()
	baseURL := p.config.baseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}

	or := openaiRequest{
		Model:       p.config.model,
		Temperature: p.config.temperature,
	}

	// Convert history to OpenAI message format
	for _, msg := range req.History {
		om := openaiMessage{
			Role:       msg.Role,
			Content:    msg.Content,
			ToolCallID: msg.ToolCallID,
		}
		// Convert tool calls if present
		for _, tc := range msg.ToolCalls {
			om.ToolCalls = append(om.ToolCalls, struct {
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			}{
				ID:   tc.ID,
				Type: "function",
				Function: struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				}{
					Name:      tc.Name,
					Arguments: formatJSON(tc.Input),
				},
			})
		}
		or.Messages = append(or.Messages, om)
	}

	// Add current input as a user message
	inputContent := ""
	for k, v := range req.Input {
		inputContent += fmt.Sprintf("%s: %v\n", k, v)
	}
	or.Messages = append(or.Messages, openaiMessage{
		Role:    "user",
		Content: inputContent,
	})

	// Add tools if available
	for _, t := range req.Tools {
		params := openaiToolParams{
			Type:       "object",
			Properties: make(map[string]openaiToolParamProperty),
		}
		for _, p := range t.Parameters {
			params.Properties[p.Name] = openaiToolParamProperty{
				Type:        string(p.Type),
				Description: p.Description,
			}
			if p.Required {
				params.Required = append(params.Required, p.Name)
			}
		}
		or.Tools = append(or.Tools, openaiTool{
			Type: "function",
			Function: openaiToolFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  params,
			},
		})
	}

	body, err := json.Marshal(or)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, openAIChatCompletionsURL(baseURL), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.apiKey)

	// Retry loop with exponential backoff: retry on 5xx and network errors only.
	var lastErr error
	var respBody []byte
	for attempt := 0; attempt <= p.config.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt) * time.Second
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		resp, err := p.config.httpClient.Do(httpReq)
		if err != nil {
			lastErr = fmt.Errorf("request failed (attempt %d/%d): %w", attempt, p.config.maxRetries, err)
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

		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("API error (status %d, attempt %d/%d): %s", resp.StatusCode, attempt, p.config.maxRetries, string(respBody))
			continue
		}

		// 4xx errors are not retried.
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}
	if lastErr != nil && len(respBody) == 0 {
		return nil, fmt.Errorf("max retries (%d) exceeded: %w", p.config.maxRetries, lastErr)
	}

	var orResp openaiResponse
	if err := json.Unmarshal(respBody, &orResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(orResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := orResp.Choices[0]
	output := make(map[string]any)
	if choice.Message.Content != "" {
		output["text"] = choice.Message.Content
	}

	// Convert tool calls back to our format
	var toolCalls []types.ToolCall
	for i, tc := range choice.Message.ToolCalls {
		id := tc.ID
		if id == "" {
			id = fmt.Sprintf("call-%d", i+1)
		}
		var args map[string]any
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			args = map[string]any{"error": "failed to parse arguments"}
		}
		toolCalls = append(toolCalls, types.ToolCall{
			ID:    id,
			Name:  tc.Function.Name,
			Input: args,
		})
	}

	cost := estimateCost(p.config.model, orResp.Usage.PromptTokens, orResp.Usage.CompletionTokens)
	logProviderCall(providerLogEntry{
		Provider:     p.config.model,
		Method:       "POST",
		URL:          openAIChatCompletionsURL(baseURL),
		Status:       200,
		DurationMs:   time.Since(start).Milliseconds(),
		InputTokens:  orResp.Usage.PromptTokens,
		OutputTokens: orResp.Usage.CompletionTokens,
		CostUSD:      cost,
	})
	return &ModelResponse{
		Output:          output,
		ToolCalls:       toolCalls,
		InputTokens:     orResp.Usage.PromptTokens,
		OutputTokens:    orResp.Usage.CompletionTokens,
		CostEstimateUSD: cost,
	}, nil
}

func openAIChatCompletionsURL(baseURL string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	if strings.HasSuffix(baseURL, "/v1") {
		return baseURL + "/chat/completions"
	}
	return baseURL + "/v1/chat/completions"
}

// formatJSON converts a map to a JSON string.
func formatJSON(m map[string]any) string {
	b, _ := json.Marshal(m)
	return string(b)
}
