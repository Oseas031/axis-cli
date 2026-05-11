package contextpack

import (
	"strings"

	"github.com/axis-cli/axis/internal/types"
)

type ExecutionContextConsumer struct {
	Registry *ReadinessRegistry
}

func NewExecutionContextConsumer(registry *ReadinessRegistry) *ExecutionContextConsumer {
	return &ExecutionContextConsumer{Registry: registry}
}

// Summarize converts an AgentTask into an ExecutionContextSummary.
//
// Design intent (Query Is Context):
//
//	The Agent declares what it needs via task.Metadata["context.requested_sources"].
//	The system resolves those declarations against the readiness registry and
//	reports which are satisfied and which are missing.
//
//	Even when readiness is not ready or is untraceable, the Agent's declared
//	needs are still preserved in RequestedSources and MissingSources so the
//	caller has a complete picture of the context gap.
func (c *ExecutionContextConsumer) Summarize(task *types.AgentTask) ExecutionContextSummary {
	registry := c.registry()
	preflight := Preflight(task, registry)
	requested := parseRequestedSources(task)

	if preflight.Status != PreflightStatusReady {
		summary := NewExecutionContextSummary(preflight, ConsumptionModeNone, nil)
		summary.RequestedSources = requested
		summary.MissingSources = append([]string(nil), requested...)
		return summary
	}
	record, err := registry.Inspect(preflight.BundleID)
	if err != nil {
		preflight.Status = PreflightStatusUntraceable
		preflight.Reason = err.Error()
		summary := NewExecutionContextSummary(preflight, ConsumptionModeObserved, nil)
		summary.RequestedSources = requested
		summary.MissingSources = append([]string(nil), requested...)
		return summary
	}
	summary := NewExecutionContextSummary(preflight, ConsumptionModeSummary, record.Artifact.Sources)
	summary.RequestedSources = requested
	summary.SatisfiedSources, summary.MissingSources = resolveSources(requested, record.Artifact.Sources)
	return summary
}

// parseRequestedSources reads Agent-declared context needs from task metadata.
//
// AgentTask.Metadata is map[string]string, so the value is stored as a
// comma-separated string rather than a typed []string. This keeps the
// metadata layer lightweight and avoids a breaking change to AgentTask.
func parseRequestedSources(task *types.AgentTask) []string {
	if task == nil || task.Metadata == nil {
		return nil
	}
	s := task.Metadata[MetadataRequestedSources]
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// resolveSources performs exact string matching between Agent-declared source
// identifiers and the sources available in the readiness registry.
//
// It is intentionally not fuzzy or similarity-based: the Agent declares what
// it wants by exact source path, and the system reports a binary yes/no.
// Future P1+ work may add prefix or glob matching behind an explicit flag.
func resolveSources(requested, available []string) (satisfied, missing []string) {
	availSet := make(map[string]bool)
	for _, a := range available {
		availSet[a] = true
	}
	for _, r := range requested {
		if availSet[r] {
			satisfied = append(satisfied, r)
		} else {
			missing = append(missing, r)
		}
	}
	return satisfied, missing
}

func (c *ExecutionContextConsumer) registry() *ReadinessRegistry {
	if c == nil || c.Registry == nil {
		return DefaultRegistry
	}
	return c.Registry
}
