package contextpack

import (
	"fmt"
)

// InitDefaultRegistry replaces the global DefaultRegistry with a file-backed
// registry anchored at the given project root.
//
// It is safe to call multiple times; each call reloads from disk and replaces
// the in-memory instance. If the store file does not exist yet, the registry
// starts empty and will create the file on the first Register call.
func InitDefaultRegistry(root string) error {
	if root == "" {
		root = "."
	}
	store := NewFileStore(root)
	registry, err := NewReadinessRegistryWithStore(store)
	if err != nil {
		return fmt.Errorf("failed to init readiness registry: %w", err)
	}
	DefaultRegistry = registry
	return nil
}
