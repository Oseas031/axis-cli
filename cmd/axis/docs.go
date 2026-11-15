package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

func newDocsCommand() *cobra.Command {
	docsCmd := &cobra.Command{
		Use:   "docs",
		Short: "Documentation knowledge base operations",
	}
	docsCmd.AddCommand(newDocsLintCommand())
	return docsCmd
}

func newDocsLintCommand() *cobra.Command {
	var checkPaths []string
	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Check documentation health (orphans, dead links, missing frontmatter)",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := findDocsRoot()
			if err != nil {
				return err
			}
			issues := runDocsLint(root)
			if len(checkPaths) > 0 {
				issues, err = filterDocIssues(issues, root, checkPaths)
				if err != nil {
					return err
				}
			}
			for _, issue := range issues {
				fmt.Println(issue)
			}
			if len(issues) > 0 {
				fmt.Fprintf(os.Stderr, "\ndocs lint: %d issues found\n", len(issues))
				os.Exit(1)
			}
			fmt.Println("docs lint: no issues found")
			return nil
		},
	}
	cmd.Flags().StringSliceVar(&checkPaths, "check", nil, "restrict findings to the given files (absolute, repo-relative, or docs-relative paths)")
	return cmd
}

func findDocsRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	docsDir := filepath.Join(wd, "docs")
	if info, err := os.Stat(docsDir); err == nil && info.IsDir() {
		return docsDir, nil
	}
	return "", fmt.Errorf("docs/ directory not found in %s", wd)
}

// filterDocIssues keeps only issues whose subject file appears in checkPaths.
func filterDocIssues(issues []string, docsRoot string, checkPaths []string) ([]string, error) {
	want := make(map[string]bool)
	for _, p := range checkPaths {
		relToDocs, err := normalizeDocPath(docsRoot, p)
		if err != nil {
			return nil, err
		}
		want["docs/"+relToDocs] = true
	}
	var kept []string
	for _, issue := range issues {
		if want[docIssueSubject(issue)] {
			kept = append(kept, issue)
		}
	}
	return kept, nil
}

// docIssueSubject extracts "docs/<rel>" from an issue line of form
// "[tag] docs/<rel>[ -> target]".
func docIssueSubject(issue string) string {
	idx := strings.Index(issue, "] ")
	if idx < 0 {
		return issue
	}
	rest := issue[idx+2:]
	if cut := strings.Index(rest, " ->"); cut >= 0 {
		rest = rest[:cut]
	}
	return rest
}

// normalizeDocPath maps a user-supplied path to its slash-separated path
// relative to docsRoot. Accepted inputs: absolute paths, repo-root-relative
// paths ("docs/specs/x.md"), and docs-root-relative paths ("specs/x.md").
func normalizeDocPath(docsRoot, p string) (string, error) {
	abs := p
	if !filepath.IsAbs(p) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		abs = filepath.Join(cwd, p)
	}
	abs = filepath.Clean(abs)

	if relToRepo, err := filepath.Rel(filepath.Dir(docsRoot), abs); err == nil {
		if slashed := filepath.ToSlash(relToRepo); strings.HasPrefix(slashed, "docs/") {
			return slashed[len("docs/"):], nil
		}
	}
	relToDocs, err := filepath.Rel(docsRoot, abs)
	if err != nil {
		return "", fmt.Errorf("path %s is outside docs root %s", p, docsRoot)
	}
	if strings.HasPrefix(relToDocs, "..") {
		return "", fmt.Errorf("path %s is outside docs root %s", p, docsRoot)
	}
	return filepath.ToSlash(relToDocs), nil
}

func runDocsLint(docsRoot string) []string {
	var issues []string

	allFiles := collectMDFiles(docsRoot)
	referenced := collectReferencedFiles(docsRoot, allFiles)

	// Check orphans: files not referenced in any README.md
	for _, f := range allFiles {
		rel, _ := filepath.Rel(docsRoot, f)
		rel = filepath.ToSlash(rel)
		base := filepath.Base(f)
		if base == "README.md" || base == "CHANGELOG.md" || base == "PURPOSE.md" {
			continue
		}
		if !referenced[rel] {
			issues = append(issues, fmt.Sprintf("[orphan] docs/%s", rel))
		}
	}

	// Check dead links
	linkRe := regexp.MustCompile(`\]\(([^)]+)\)`)
	for _, f := range allFiles {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		dir := filepath.Dir(f)
		matches := linkRe.FindAllStringSubmatch(string(data), -1)
		for _, m := range matches {
			link := m[1]
			if strings.HasPrefix(link, "http") || strings.HasPrefix(link, "#") || strings.HasPrefix(link, "mailto:") {
				continue
			}
			// Strip anchor
			if idx := strings.Index(link, "#"); idx >= 0 {
				link = link[:idx]
			}
			if link == "" {
				continue
			}
			if strings.ContainsAny(link, "{}") {
				// Placeholder URL in template text (e.g. "[x]({path})") is not
				// a real link; skip rather than report a dead link.
				continue
			}
			target := filepath.Join(dir, filepath.FromSlash(link))
			if _, err := os.Stat(target); os.IsNotExist(err) {
				rel, _ := filepath.Rel(docsRoot, f)
				issues = append(issues, fmt.Sprintf("[dead-link] docs/%s -> %s", filepath.ToSlash(rel), link))
			}
		}
	}

	// Check missing frontmatter in architecture/ and specs/
	for _, f := range allFiles {
		rel, _ := filepath.Rel(docsRoot, f)
		rel = filepath.ToSlash(rel)
		if !strings.HasPrefix(rel, "architecture/") && !strings.HasPrefix(rel, "specs/") {
			continue
		}
		if filepath.Base(f) == "README.md" {
			continue
		}
		if !hasFrontmatter(f) {
			issues = append(issues, fmt.Sprintf("[no-frontmatter] docs/%s", rel))
		}
	}

	return issues
}

func collectMDFiles(root string) []string {
	var files []string
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, ".md") {
			files = append(files, path)
		}
		return nil
	})
	return files
}

func collectReferencedFiles(docsRoot string, allFiles []string) map[string]bool {
	referenced := make(map[string]bool)
	linkRe := regexp.MustCompile(`\]\(([^)]+)\)`)
	for _, f := range allFiles {
		if filepath.Base(f) != "README.md" {
			continue
		}
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		dir := filepath.Dir(f)
		for _, m := range linkRe.FindAllStringSubmatch(string(data), -1) {
			link := m[1]
			if strings.HasPrefix(link, "http") || strings.HasPrefix(link, "#") {
				continue
			}
			if idx := strings.Index(link, "#"); idx >= 0 {
				link = link[:idx]
			}
			if link == "" {
				continue
			}
			target := filepath.Join(dir, filepath.FromSlash(link))
			if info, err := os.Stat(target); err == nil && info.IsDir() {
				// Directory-style link (e.g. "[actor-comm](actor-comm/)" in a
				// spec index) references every markdown file under the target.
				filepath.Walk(target, func(p string, fi os.FileInfo, err error) error {
					if err != nil || fi.IsDir() || !strings.HasSuffix(p, ".md") {
						return nil
					}
					if rel, err := filepath.Rel(docsRoot, p); err == nil {
						referenced[filepath.ToSlash(rel)] = true
					}
					return nil
				})
				continue
			}
			rel, err := filepath.Rel(docsRoot, target)
			if err == nil {
				referenced[filepath.ToSlash(rel)] = true
			}
		}
	}
	return referenced
}

func hasFrontmatter(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()) == "---"
	}
	return false
}
