package vigil

// StoreInterface defines the interface for vigil store operations.
// This enables dependency injection and testing.
type StoreInterface interface {
	// Load returns all vigil items.
	Load() ([]*Item, error)

	// Save persists all vigil items.
	Save(items []*Item) error

	// Add appends a new vigil item.
	Add(item *Item) error

	// Get retrieves a vigil item by ID.
	Get(id string) (*Item, error)

	// Update modifies an existing vigil item.
	Update(item *Item) error

	// Archive moves completed items to archive storage.
	Archive(items []*Item) error
}

// Ensure Store implements StoreInterface at compile time.
var _ StoreInterface = (*Store)(nil)
