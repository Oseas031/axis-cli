package evolution

// StoreInterface defines the interface for evolution store operations.
// This enables dependency injection and testing.
type StoreInterface interface {
	// CreateRun creates a new evolution run.
	CreateRun(intent EvolutionIntent, run EvolutionRun) error

	// ReadRun retrieves a run by ID.
	ReadRun(runID string) (*EvolutionRun, error)

	// ListRuns returns all run IDs.
	ListRuns() ([]string, error)

	// RunDir returns the workspace directory for a run.
	RunDir(runID string) string

	// ReadIntent retrieves an intent by ID.
	ReadIntent(intentID string) (*EvolutionIntent, error)

	// AppendDecision records an evolution decision for a specific run.
	AppendDecision(runID string, decision EvolutionDecision) error

	// ReadDecision retrieves the latest decision for a run.
	ReadDecision(runID string) (*EvolutionDecision, error)
}

// Ensure Store implements StoreInterface at compile time.
var _ StoreInterface = (*Store)(nil)
