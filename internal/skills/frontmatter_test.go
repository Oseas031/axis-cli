package skills

import "testing"

func TestParseFrontmatter_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		key     string
		wantVal string
	}{
		{
			name:    "value with multiple colons",
			input:   "---\ndescription: Use when: something happens\n---\nbody",
			key:     "description",
			wantVal: "Use when: something happens",
		},
		{
			name:    "double-quoted value",
			input:   "---\ndescription: \"A skill with: colons\"\n---\nbody",
			key:     "description",
			wantVal: "A skill with: colons",
		},
		{
			name:    "single-quoted value",
			input:   "---\nname: 'my-skill'\n---\nbody",
			key:     "name",
			wantVal: "my-skill",
		},
		{
			name:    "empty value",
			input:   "---\nversion:\n---\nbody",
			key:     "version",
			wantVal: "",
		},
		{
			name:    "normal case",
			input:   "---\nname: diagnose\n---\nbody",
			key:     "name",
			wantVal: "diagnose",
		},
		{
			name:    "leading and trailing spaces",
			input:   "---\nname:   diagnose  \n---\nbody",
			key:     "name",
			wantVal: "diagnose",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta, _, err := parseFrontmatter(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := meta[tt.key]; got != tt.wantVal {
				t.Errorf("meta[%q] = %q, want %q", tt.key, got, tt.wantVal)
			}
		})
	}
}

func TestParseFrontmatter_TagsBracketSyntax(t *testing.T) {
	input := "---\nname: test\ntags: [engineering, debug]\n---\nbody"
	meta, _, err := parseFrontmatter(input)
	if err != nil {
		t.Fatal(err)
	}
	tags := parseTags(meta["tags"])
	if len(tags) != 2 || tags[0] != "engineering" || tags[1] != "debug" {
		t.Errorf("tags = %v, want [engineering debug]", tags)
	}
}

func TestStripQuotes(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`"hello"`, "hello"},
		{`'hello'`, "hello"},
		{`hello`, "hello"},
		{`""`, ""},
		{`"mismatched'`, `"mismatched'`},
		{`"`, `"`},
	}
	for _, tt := range tests {
		if got := stripQuotes(tt.input); got != tt.want {
			t.Errorf("stripQuotes(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
