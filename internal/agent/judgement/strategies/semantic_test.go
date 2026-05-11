package strategies

import (
	"context"
	"errors"
	"testing"

	"github.com/axis-cli/axis/internal/model/provider"
)

// mockModelProvider is a test double for provider.ModelProvider.
type mockModelProvider struct {
	response *provider.ModelResponse
	err      error
}

func (m *mockModelProvider) Execute(ctx context.Context, req *provider.ModelRequest) (*provider.ModelResponse, error) {
	return m.response, m.err
}

func TestSemanticValidationStrategy_CanHandle(t *testing.T) {
	s := NewSemanticValidationStrategy(&mockModelProvider{})
	if !s.CanHandle(JudgementCriteria{Type: JudgementTypeSemantic}) {
		t.Error("expected to handle semantic type")
	}
	if s.CanHandle(JudgementCriteria{Type: JudgementTypeSyntax}) {
		t.Error("expected not to handle syntax type")
	}
}

func TestSemanticValidationStrategy_Validate_EmptyCode(t *testing.T) {
	s := NewSemanticValidationStrategy(&mockModelProvider{})

	result, err := s.Validate(map[string]any{"code": ""}, JudgementCriteria{
		Name:   "semantic",
		Type:   JudgementTypeSemantic,
		Weight: 1.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Passed {
		t.Error("expected empty code to pass")
	}
	if result.Score != 1.0 {
		t.Errorf("expected score 1.0, got %f", result.Score)
	}
}

func TestSemanticValidationStrategy_Validate_StringInput(t *testing.T) {
	s := NewSemanticValidationStrategy(&mockModelProvider{})

	result, err := s.Validate("package main\nfunc main() {}", JudgementCriteria{
		Name:   "semantic",
		Type:   JudgementTypeSemantic,
		Weight: 1.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Passed {
		t.Error("expected to fail without mock provider response")
	}
}

func TestSemanticValidationStrategy_Validate_ProviderError(t *testing.T) {
	mock := &mockModelProvider{err: errors.New("model unavailable")}
	s := NewSemanticValidationStrategy(mock)

	result, err := s.Validate("code", JudgementCriteria{
		Name:   "semantic",
		Type:   JudgementTypeSemantic,
		Weight: 1.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Passed {
		t.Error("expected provider error to result in failure")
	}
	if result.Error == "" {
		t.Error("expected error message in result")
	}
}

func TestSemanticValidationStrategy_Validate_ValidJSONResponse(t *testing.T) {
	mock := &mockModelProvider{
		response: &provider.ModelResponse{
			Output: map[string]any{
				"output": `{"passed": true, "score": 0.95, "issues": [], "confidence": 0.9}`,
			},
		},
	}
	s := NewSemanticValidationStrategy(mock)

	result, err := s.Validate("code", JudgementCriteria{
		Name:   "semantic",
		Type:   JudgementTypeSemantic,
		Weight: 1.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Passed {
		t.Errorf("expected valid JSON response to pass, got: %s", result.Error)
	}
	if result.Score != 0.95 {
		t.Errorf("expected score 0.95, got %f", result.Score)
	}
}

func TestSemanticValidationStrategy_Validate_InvalidJSONResponse(t *testing.T) {
	mock := &mockModelProvider{
		response: &provider.ModelResponse{
			Output: map[string]any{
				"output": `not json at all`,
			},
		},
	}
	s := NewSemanticValidationStrategy(mock)

	result, err := s.Validate("code", JudgementCriteria{
		Name:   "semantic",
		Type:   JudgementTypeSemantic,
		Weight: 1.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Fallback: response without JSON braces -> no error keyword, so passed=true with score 0.5
	if !result.Passed {
		t.Errorf("expected fallback to pass, got: %s", result.Error)
	}
}

func TestSemanticValidationStrategy_extractCode(t *testing.T) {
	s := NewSemanticValidationStrategy(&mockModelProvider{})

	tests := []struct {
		name     string
		input    any
		expected string
		wantErr  bool
	}{
		{"string", "hello", "hello", false},
		{"bytes", []byte("world"), "world", false},
		{"map with code", map[string]any{"code": "code"}, "code", false},
		{"map with files", map[string]any{"files": []string{"a", "b"}}, "a\nb", false},
		{"unsupported", 123, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := s.extractCode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("extractCode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if code != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, code)
			}
		})
	}
}

func TestSemanticValidationStrategy_parseResponse(t *testing.T) {
	s := NewSemanticValidationStrategy(&mockModelProvider{})

	tests := []struct {
		name         string
		content      string
		wantPassed   bool
		wantScore    float64
		wantErrField bool
	}{
		{
			name:       "valid JSON",
			content:    `{"passed": true, "score": 0.9, "issues": ["none"]}`,
			wantPassed: true,
			wantScore:  0.9,
		},
		{
			name:       "no JSON braces",
			content:    "no issues found here",
			wantPassed: true,
			wantScore:  0.5,
		},
		{
			name:       "no JSON braces with error",
			content:    "there is an error in the code",
			wantPassed: false,
			wantScore:  0.5,
		},
		{
			name:         "invalid JSON",
			content:      `{"passed":}`,
			wantErrField: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := s.parseResponse(tt.content, JudgementCriteria{Name: "semantic"})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if item.Passed != tt.wantPassed {
				t.Errorf("passed = %v, want %v", item.Passed, tt.wantPassed)
			}
			if item.Score != tt.wantScore {
				t.Errorf("score = %f, want %f", item.Score, tt.wantScore)
			}
			if tt.wantErrField && item.Error == "" {
				t.Error("expected error field to be set")
			}
		})
	}
}
