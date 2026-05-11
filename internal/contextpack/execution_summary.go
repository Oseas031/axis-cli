package contextpack

type ConsumptionMode string

const (
	ConsumptionModeNone            ConsumptionMode = "none"
	ConsumptionModeObserved        ConsumptionMode = "observed"
	ConsumptionModeSummary         ConsumptionMode = "summary"
	ConsumptionModePromptAugmented ConsumptionMode = "prompt_augmented"
)

type ExecutionContextSummary struct {
	BundleID         string          `json:"bundle_id,omitempty"`
	Status           PreflightStatus `json:"status"`
	Reason           string          `json:"reason,omitempty"`
	ConsumptionMode  ConsumptionMode `json:"consumption_mode"`
	PacketCount      int             `json:"packet_count,omitempty"`
	Truncated        bool            `json:"truncated"`
	SourceDigest     string          `json:"source_digest,omitempty"`
	Sources          []string        `json:"sources,omitempty"`
	RequestedSources []string        `json:"requested_sources,omitempty"`
	SatisfiedSources []string        `json:"satisfied_sources,omitempty"`
	MissingSources   []string        `json:"missing_sources,omitempty"`
}

func NewExecutionContextSummary(preflight PreflightResult, mode ConsumptionMode, sources []string) ExecutionContextSummary {
	return ExecutionContextSummary{
		BundleID:        preflight.BundleID,
		Status:          preflight.Status,
		Reason:          preflight.Reason,
		ConsumptionMode: mode,
		PacketCount:     preflight.PacketCount,
		Truncated:       preflight.Truncated,
		SourceDigest:    preflight.SourceDigest,
		Sources:         append([]string(nil), sources...),
	}
}
