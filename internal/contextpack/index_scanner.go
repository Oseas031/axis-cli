package contextpack

import (
	"os"
	"path/filepath"
	"strings"
)

// DocumentChunk represents a scanned document fragment ready for indexing.
type DocumentChunk struct {
	Source  string `json:"source"`
	Content string `json:"content"`
	ModTime int64  `json:"mod_time"`
	DocType string `json:"doc_type"`
}

// DocumentScanner scans a project tree for indexable documents.
type DocumentScanner struct {
	Root string
}

const maxFileSize = 1 << 20 // 1MB

// Scan walks the project root and returns indexable chunks.
// It includes .md and .go files, excluding .git, vendor, node_modules, and .axis directories.
func (s *DocumentScanner) Scan() ([]DocumentChunk, error) {
	var chunks []DocumentChunk
	err := filepath.WalkDir(s.Root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil // skip unreadable file, continue scanning
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "vendor" || name == "node_modules" || name == ".axis" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".go" {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil // skip file with unreadable metadata
		}
		if info.Size() > maxFileSize {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil // skip unreadable file
		}
		docType := "code"
		if ext == ".md" {
			docType = "doc"
		}
		rel, err := filepath.Rel(s.Root, path)
		if err != nil {
			rel = path
		}
		chunks = append(chunks, DocumentChunk{
			Source:  rel,
			Content: string(content),
			ModTime: info.ModTime().Unix(),
			DocType: docType,
		})
		return nil
	})
	return chunks, err
}
