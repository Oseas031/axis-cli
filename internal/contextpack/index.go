package contextpack

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// IndexStatus reports the health and coverage of a context index.
type IndexStatus struct {
	Healthy      bool      `json:"healthy"`
	IndexedFiles int       `json:"indexed_files"`
	TotalChunks  int       `json:"total_chunks"`
	LastBuildAt  time.Time `json:"last_build_at"`
	Message      string    `json:"message,omitempty"`
}

// IndexPath returns the default index file path under the project root.
func IndexPath(root string) string {
	return filepath.Join(root, ".axis", "context", "index.json")
}

// IndexManager orchestrates index lifecycle: rebuild, update, status.
type IndexManager struct {
	index *TFIDFIndex
	store *IndexStore
}

// NewIndexManager creates a new index manager with default file store.
func NewIndexManager() *IndexManager {
	return &IndexManager{
		store: &IndexStore{},
	}
}

// Rebuild performs a full scan and rebuild of the index from scratch.
func (m *IndexManager) Rebuild(root string) (*IndexStatus, error) {
	scanner := &DocumentScanner{Root: root}
	chunks, err := scanner.Scan()
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	idx := &TFIDFIndex{}
	idx.Build(chunks)

	path := IndexPath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("mkdir failed: %w", err)
	}
	if err := m.store.Save(path, idx); err != nil {
		return nil, fmt.Errorf("save failed: %w", err)
	}

	m.index = idx
	return &IndexStatus{
		Healthy:      len(chunks) > 0,
		IndexedFiles: countFiles(chunks),
		TotalChunks:  len(chunks),
		LastBuildAt:  time.Now(),
		Message:      "rebuilt",
	}, nil
}

// Update incrementally refreshes the index based on file mtime changes.
// If no prior index exists, it falls back to Rebuild.
func (m *IndexManager) Update(root string) (*IndexStatus, error) {
	path := IndexPath(root)
	existing, err := m.store.Load(path)
	if err != nil {
		return m.Rebuild(root)
	}

	scanner := &DocumentScanner{Root: root}
	currentChunks, err := scanner.Scan()
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	existingMap := make(map[string]DocumentChunk)
	for _, c := range existing.Chunks {
		existingMap[c.Source] = c
	}

	var updatedChunks []DocumentChunk
	for _, c := range currentChunks {
		if ec, ok := existingMap[c.Source]; ok && ec.ModTime == c.ModTime {
			updatedChunks = append(updatedChunks, ec)
		} else {
			updatedChunks = append(updatedChunks, c)
		}
	}

	idx := &TFIDFIndex{}
	idx.Build(updatedChunks)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("mkdir failed: %w", err)
	}
	if err := m.store.Save(path, idx); err != nil {
		return nil, fmt.Errorf("save failed: %w", err)
	}

	m.index = idx
	return &IndexStatus{
		Healthy:      len(updatedChunks) > 0,
		IndexedFiles: countFiles(updatedChunks),
		TotalChunks:  len(updatedChunks),
		LastBuildAt:  time.Now(),
		Message:      "updated",
	}, nil
}

// Status loads and reports the current index state without modifying it.
func (m *IndexManager) Status(root string) *IndexStatus {
	path := IndexPath(root)
	info, err := os.Stat(path)
	if err != nil {
		return &IndexStatus{
			Healthy: false,
			Message: "index not found",
		}
	}

	idx, err := m.store.Load(path)
	if err != nil {
		return &IndexStatus{
			Healthy: false,
			Message: "index corrupt: " + err.Error(),
		}
	}

	return &IndexStatus{
		Healthy:      len(idx.Chunks) > 0,
		IndexedFiles: countFiles(idx.Chunks),
		TotalChunks:  len(idx.Chunks),
		LastBuildAt:  info.ModTime(),
		Message:      "index loaded",
	}
}

// Load reads the index from disk into memory without rebuilding.
func (m *IndexManager) Load(root string) error {
	path := IndexPath(root)
	idx, err := m.store.Load(path)
	if err != nil {
		return err
	}
	m.index = idx
	return nil
}

// Index returns the in-memory TF-IDF index, or nil if not yet loaded.
func (m *IndexManager) Index() *TFIDFIndex {
	return m.index
}

func countFiles(chunks []DocumentChunk) int {
	seen := make(map[string]bool)
	for _, c := range chunks {
		seen[c.Source] = true
	}
	return len(seen)
}
