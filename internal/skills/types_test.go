package skills

import (
	"encoding/json"
	"testing"
)

func TestSkillMetaValidate(t *testing.T) {
	tests := []struct {
		name    string
		meta    SkillMeta
		wantErr error
	}{
		{"valid", SkillMeta{Name: "pdf", Description: "Process PDFs"}, nil},
		{"valid with tags", SkillMeta{Name: "code-review", Description: "Review code", Tags: []string{"dev"}}, nil},
		{"empty name", SkillMeta{Name: "", Description: "desc"}, ErrSkillNameRequired},
		{"whitespace name", SkillMeta{Name: "  ", Description: "desc"}, ErrSkillNameRequired},
		{"invalid name uppercase", SkillMeta{Name: "PDF", Description: "desc"}, ErrInvalidSkillName},
		{"invalid name underscore", SkillMeta{Name: "code_review", Description: "desc"}, ErrInvalidSkillName},
		{"invalid name starts with dash", SkillMeta{Name: "-pdf", Description: "desc"}, ErrInvalidSkillName},
		{"invalid name ends with dash", SkillMeta{Name: "pdf-", Description: "desc"}, ErrInvalidSkillName},
		{"single char name", SkillMeta{Name: "a", Description: "desc"}, ErrInvalidSkillName},
		{"two char name", SkillMeta{Name: "ab", Description: "desc"}, nil},
		{"empty description", SkillMeta{Name: "pdf", Description: ""}, ErrDescriptionRequired},
		{"whitespace description", SkillMeta{Name: "pdf", Description: "   "}, ErrDescriptionRequired},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.meta.Validate()
			if err != tt.wantErr {
				t.Errorf("got %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSkillName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"valid three char", "pdf", nil},
		{"valid two char", "ab", nil},
		{"valid kebab", "code-review", nil},
		{"valid with numbers", "tool2go", nil},
		{"empty", "", ErrSkillNameRequired},
		{"dots", "my.skill", ErrInvalidSkillName},
		{"path escape", "../etc", ErrInvalidSkillName},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSkillName(tt.input)
			if err != tt.wantErr {
				t.Errorf("got %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestSkillMetaJSONRoundTrip(t *testing.T) {
	original := SkillMeta{
		Name:        "code-review",
		Description: "Review code for quality",
		Tags:        []string{"dev", "quality"},
		Version:     "1.0.0",
		Author:      "axis",
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded SkillMeta
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Name != original.Name || decoded.Description != original.Description {
		t.Errorf("round-trip mismatch: got %+v", decoded)
	}
	if len(decoded.Tags) != len(original.Tags) {
		t.Errorf("tags mismatch: got %v, want %v", decoded.Tags, original.Tags)
	}
}

func TestLoadSkillOutputJSON(t *testing.T) {
	out := LoadSkillOutput{
		Name:        "pdf",
		Description: "Process PDFs",
		Content:     "# PDF\n\nInstructions...",
	}
	data, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded LoadSkillOutput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Content != out.Content {
		t.Errorf("content mismatch")
	}
}
