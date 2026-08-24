package compactor

// StoreInterface defines the interface for compactor store operations.
// This enables dependency injection and testing.
type StoreInterface interface {
	// Offload writes the full tool result to refs/ and appends an entry to offload.jsonl.
	Offload(toolCallID, toolName, content string, summary string, score int) (*OffloadEntry, error)

	// ReadRef reads the full content of an offloaded tool result.
	ReadRef(refPath string) (string, error)
}

// Ensure Store implements StoreInterface at compile time.
var _ StoreInterface = (*Store)(nil)
