package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeDocFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

const testFrontmatter = "---\ntype: spec\nstatus: active\ncreated: 2026-08-26\nlast_verified: 2026-08-26\n---\n\n"

// Regression: directory-style links in index tables (e.g. "[actor-comm](actor-comm/)")
// must mark every markdown file under the target as referenced.
func TestOrphanDirectoryLinkRecognition(t *testing.T) {
	root := t.TempDir()
	writeDocFile(t, filepath.Join(root, "README.md"), "# Docs\n\n- [foo](specs/foo/)\n")
	writeDocFile(t, filepath.Join(root, "specs", "foo", "a.md"), testFrontmatter+"# A\n")
	writeDocFile(t, filepath.Join(root, "specs", "foo", "b.md"), testFrontmatter+"# B\n")
	writeDocFile(t, filepath.Join(root, "specs", "bar", "c.md"), testFrontmatter+"# C\n")

	var orphans []string
	for _, issue := range runDocsLint(root) {
		if strings.HasPrefix(issue, "[orphan]") {
			orphans = append(orphans, issue)
		}
	}
	want := "[orphan] docs/specs/bar/c.md"
	if len(orphans) != 1 || orphans[0] != want {
		t.Fatalf("expected exactly [%q], got %q", want, orphans)
	}
}

// Regression: --check restricts findings to the requested files.
func TestCheckFlagFiltersIssues(t *testing.T) {
	root := t.TempDir()
	fileA := filepath.Join(root, "a.md")
	fileB := filepath.Join(root, "b.md")
	writeDocFile(t, filepath.Join(root, "README.md"), "# Docs\n\n- [a](a.md)\n- [b](b.md)\n")
	writeDocFile(t, fileA, "# A\n\n[x](missing-a.md)\n")
	writeDocFile(t, fileB, "# B\n\n[y](missing-b.md)\n")

	all := runDocsLint(root)
	if len(all) != 2 {
		t.Fatalf("expected 2 dead-link issues, got %d: %q", len(all), all)
	}

	filtered, err := filterDocIssues(all, root, []string{fileA})
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 1 || !strings.Contains(filtered[0], "docs/a.md") {
		t.Fatalf("expected only a.md issue, got %q", filtered)
	}

	filtered, err = filterDocIssues(all, root, []string{filepath.Join(root, "..", filepath.Base(root), "b.md")})
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 1 || !strings.Contains(filtered[0], "docs/b.md") {
		t.Fatalf("expected only b.md issue after normalization, got %q", filtered)
	}
}

func TestNormalizeDocPathRejectsOutsidePaths(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(root, "..", "elsewhere.md")
	if _, err := normalizeDocPath(root, outside); err == nil {
		t.Fatal("expected error for path outside docs root")
	}
}

// Regression: placeholder template URLs like "[x]({path})" are not real links.
func TestPlaceholderLinksAreSkipped(t *testing.T) {
	root := t.TempDir()
	writeDocFile(t, filepath.Join(root, "a.md"), "# A\n\n> DEPRECATED: see [{replacement}]({path}).\n")

	for _, issue := range runDocsLint(root) {
		if strings.Contains(issue, "dead-link") {
			t.Fatalf("placeholder link must not be reported: %q", issue)
		}
	}
}

func TestDocIssueSubjectStripsTarget(t *testing.T) {
	got := docIssueSubject("[dead-link] docs/a.md -> missing.md")
	if got != "docs/a.md" {
		t.Fatalf("got %q", got)
	}
}
