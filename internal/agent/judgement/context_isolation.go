package judgement

// IsolatedJudgeInput contains only the final-state data needed for judgement,
// stripped of intermediate tool calls and multi-turn transcript bloat.
// Retains a structured execution summary per Devil's Advocate constraint:
// blind artifact-only judging rewards brute-force solutions.
type IsolatedJudgeInput struct {
	TaskSpec       string           `json:"task_spec"`
	FinalArtifacts []Artifact       `json:"final_artifacts"`
	ContractSchema string           `json:"contract_schema"`
	ExecSummary    ExecutionSummary `json:"exec_summary"`
}

// ExecutionSummary retains minimal process signal without full transcript.
type ExecutionSummary struct {
	Attempts       int      `json:"attempts"`
	ErrorCategories []string `json:"error_categories"`
	TotalDuration  string   `json:"total_duration"`
	ToolsUsed      []string `json:"tools_used"`
}

// Artifact represents a single output artifact produced by execution.
type Artifact struct {
	Type    string `json:"type"`
	Path    string `json:"path"`
	Content string `json:"content"`
	Diff    string `json:"diff"`
}

// IsolateContext extracts only final-state artifacts from a full execution result,
// preventing Context Rot degradation in judgement strategies.
func IsolateContext(fullInput any) *IsolatedJudgeInput {
	if fullInput == nil {
		return &IsolatedJudgeInput{}
	}

	// If already isolated, return as-is
	if isolated, ok := fullInput.(*IsolatedJudgeInput); ok {
		return isolated
	}

	// Extract from map-based execution results (the common internal format)
	m, ok := fullInput.(map[string]any)
	if !ok {
		return &IsolatedJudgeInput{}
	}

	result := &IsolatedJudgeInput{}

	if spec, ok := m["task_spec"].(string); ok {
		result.TaskSpec = spec
	}
	if schema, ok := m["contract_schema"].(string); ok {
		result.ContractSchema = schema
	}

	// Extract only final artifacts, skip intermediate tool_calls/steps
	if artifacts, ok := m["final_artifacts"].([]any); ok {
		for _, a := range artifacts {
			if am, ok := a.(map[string]any); ok {
				result.FinalArtifacts = append(result.FinalArtifacts, extractArtifact(am))
			}
		}
	} else if artifacts, ok := m["artifacts"].([]any); ok {
		for _, a := range artifacts {
			if am, ok := a.(map[string]any); ok {
				result.FinalArtifacts = append(result.FinalArtifacts, extractArtifact(am))
			}
		}
	}

	// Extract structured execution summary (minimal process signal)
	if summary, ok := m["exec_summary"].(map[string]any); ok {
		if v, ok := summary["attempts"].(int); ok {
			result.ExecSummary.Attempts = v
		}
		if v, ok := summary["total_duration"].(string); ok {
			result.ExecSummary.TotalDuration = v
		}
		if cats, ok := summary["error_categories"].([]any); ok {
			for _, c := range cats {
				if s, ok := c.(string); ok {
					result.ExecSummary.ErrorCategories = append(result.ExecSummary.ErrorCategories, s)
				}
			}
		}
		if tools, ok := summary["tools_used"].([]any); ok {
			for _, t := range tools {
				if s, ok := t.(string); ok {
					result.ExecSummary.ToolsUsed = append(result.ExecSummary.ToolsUsed, s)
				}
			}
		}
	}

	return result
}

func extractArtifact(m map[string]any) Artifact {
	a := Artifact{}
	if v, ok := m["type"].(string); ok {
		a.Type = v
	}
	if v, ok := m["path"].(string); ok {
		a.Path = v
	}
	if v, ok := m["content"].(string); ok {
		a.Content = v
	}
	if v, ok := m["diff"].(string); ok {
		a.Diff = v
	}
	return a
}
