package skills

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
)

func testdataDir() string {
	_, f, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(f), "testdata", "skills")
}

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantKey string
		wantVal string
		wantErr bool
	}{
		{
			name:    "valid",
			input:   "---\nname: pdf\ndescription: Process PDFs\n---\n# Body",
			wantKey: "name",
			wantVal: "pdf",
		},
		{
			name:    "missing opening",
			input:   "name: pdf\n---\n# Body",
			wantErr: true,
		},
		{
			name:    "missing closing",
			input:   "---\nname: pdf\n# Body",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta, _, err := parseFrontmatter(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if meta[tt.wantKey] != tt.wantVal {
				t.Errorf("got %q, want %q", meta[tt.wantKey], tt.wantVal)
			}
		})
	}
}

func TestParseFrontmatterBody(t *testing.T) {
	input := "---\nname: pdf\n---\n# Title\n\nBody content."
	_, body, err := parseFrontmatter(input)
	if err != nil {
		t.Fatal(err)
	}
	if body != "# Title\n\nBody content." {
		t.Errorf("unexpected body: %q", body)
	}
}

func TestParseTags(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"tag1, tag2", 2},
		{"single", 1},
		{"", 0},
		{"  ", 0},
		{"a, b, c", 3},
	}
	for _, tt := range tests {
		got := parseTags(tt.input)
		if len(got) != tt.want {
			t.Errorf("parseTags(%q) = %d tags, want %d", tt.input, len(got), tt.want)
		}
	}
}

func TestDiscoverWithFixtures(t *testing.T) {
	loader := NewLoader(testdataDir())
	metas, err := loader.Discover(context.Background())
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if len(metas) != 2 {
		t.Fatalf("got %d skills, want 2", len(metas))
	}
	found := make(map[string]bool)
	for _, m := range metas {
		found[m.Name] = true
	}
	if !found["pdf"] || !found["code-review"] {
		t.Errorf("expected pdf and code-review, got %v", found)
	}
}

func TestDiscoverCachesResults(t *testing.T) {
	loader := NewLoader(testdataDir())
	m1, _ := loader.Discover(context.Background())
	m2, _ := loader.Discover(context.Background())
	if len(m1) != len(m2) {
		t.Error("cached results differ")
	}
}

func TestDiscoverEmptyDir(t *testing.T) {
	loader := NewLoader(t.TempDir())
	metas, err := loader.Discover(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(metas) != 0 {
		t.Errorf("expected 0 skills, got %d", len(metas))
	}
}

func TestDiscoverNonExistentDir(t *testing.T) {
	loader := NewLoader(filepath.Join(t.TempDir(), "nonexistent"))
	metas, err := loader.Discover(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if metas != nil {
		t.Errorf("expected nil, got %v", metas)
	}
}
