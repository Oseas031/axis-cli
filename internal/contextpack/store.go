package contextpack

// ReadinessStore defines the persistence contract for readiness records.
// Implementations must be safe for use with ReadinessRegistry.
type ReadinessStore interface {
	LoadAll() (map[string]ReadinessRecord, error)
	SaveAll(records map[string]ReadinessRecord) error
	DeleteAll() error
}
