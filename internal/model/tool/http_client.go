package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

// HTTPClientTool makes HTTP requests to allowed hosts.
type HTTPClientTool struct {
	allowedHosts []string
}

// NewHTTPClientTool creates a new HTTPClientTool with the specified allowed hosts.
func NewHTTPClientTool(allowedHosts []string) *HTTPClientTool {
	return &HTTPClientTool{allowedHosts: allowedHosts}
}

// Name returns the tool name.
func (t *HTTPClientTool) Name() string {
	return "http_request"
}

// Schema returns the tool definition for http_request.
func (t *HTTPClientTool) Schema() types.ToolDefinition {
	return types.ToolDefinition{
		Name:        "http_request",
		Description: "Make HTTP requests to allowed hosts",
		Parameters: []types.FieldDef{
			{Name: "method", Type: types.FieldTypeString, Required: true, Description: "HTTP method (GET, POST, PUT, DELETE)", Enum: []string{"GET", "POST", "PUT", "DELETE"}},
			{Name: "url", Type: types.FieldTypeString, Required: true, Description: "URL to request"},
			{Name: "headers", Type: types.FieldTypeObject, Required: false, Description: "HTTP headers as key-value pairs"},
			{Name: "body", Type: types.FieldTypeString, Required: false, Description: "Request body"},
		},
	}
}

// validateHost checks if the URL host is in the allowed hosts list.
func (t *HTTPClientTool) validateHost(requestedURL string) error {
	parsedURL, err := url.Parse(requestedURL)
	if err != nil {
		return &HostValidationError{URL: requestedURL, Reason: "invalid URL format"}
	}

	host := parsedURL.Host
	if host == "" {
		return &HostValidationError{URL: requestedURL, Reason: "missing host in URL"}
	}

	// Strip port if present for comparison
	hostname := host
	if colonIdx := strings.LastIndex(host, ":"); colonIdx != -1 {
		hostname = host[:colonIdx]
	}

	for _, allowed := range t.allowedHosts {
		if hostname == allowed {
			return nil
		}
	}
	return &HostValidationError{URL: requestedURL, Reason: "host is not in allowed hosts list"}
}

// Execute makes an HTTP request.
func (t *HTTPClientTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
	method, ok := input["method"].(string)
	if !ok || method == "" {
		return map[string]any{"error": "method is required and must be a string"}, nil
	}

	urlStr, ok := input["url"].(string)
	if !ok || urlStr == "" {
		return map[string]any{"error": "url is required and must be a string"}, nil
	}

	// Validate host
	if err := t.validateHost(urlStr); err != nil {
		return map[string]any{"error": err.Error()}, nil
	}

	// Build request
	var body io.Reader
	if bodyStr, ok := input["body"].(string); ok && bodyStr != "" {
		body = bytes.NewBufferString(bodyStr)
	}

	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(method), urlStr, body)
	if err != nil {
		return map[string]any{"error": "failed to create request: " + err.Error()}, nil
	}

	// Set headers
	if headers, ok := input["headers"].(map[string]any); ok {
		for k, v := range headers {
			if vStr, ok := v.(string); ok {
				req.Header.Set(k, vStr)
			}
		}
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return map[string]any{"error": "request failed: " + err.Error()}, nil
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return map[string]any{"error": "failed to read response: " + err.Error()}, nil
	}

	// Try to parse as JSON for structured response
	var respData any
	if json.Unmarshal(respBody, &respData) != nil {
		respData = string(respBody)
	}

	return map[string]any{
		"status":  resp.StatusCode,
		"headers": resp.Header,
		"body":    respData,
	}, nil
}

// HostValidationError represents a host validation failure.
type HostValidationError struct {
	URL    string
	Reason string
}

func (e *HostValidationError) Error() string {
	return fmt.Sprintf("host validation failed: %s: %s", e.Reason, e.URL)
}
