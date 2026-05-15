package dispatcher

import "github.com/axis-cli/axis/internal/agent"

// Permission represents the tri-state permission decision.
type Permission int

const (
	// PermissionAllow — operation proceeds without confirmation.
	PermissionAllow Permission = iota
	// PermissionAsk — operation pauses, awaiting external confirmation.
	PermissionAsk
	// PermissionDeny — operation is rejected.
	PermissionDeny
)

func (p Permission) String() string {
	switch p {
	case PermissionAllow:
		return "allow"
	case PermissionAsk:
		return "ask"
	case PermissionDeny:
		return "deny"
	default:
		return "unknown"
	}
}

// PermissionResolver determines the permission level for a task based on
// autonomy level and risk indicators in metadata.
type PermissionResolver struct {
	// HighRiskKeys are metadata keys whose presence triggers ask/deny.
	HighRiskKeys []string
}

// NewPermissionResolver creates a resolver with default high-risk indicators.
func NewPermissionResolver() *PermissionResolver {
	return &PermissionResolver{
		HighRiskKeys: []string{
			"axis.evolution_required",
			"axis.destructive",
			"tool.unrestricted",
		},
	}
}

// Resolve determines the permission for a task given its autonomy level and metadata.
// Rules:
//   - AutonomyLevelFull + no high-risk indicators → Allow
//   - AutonomyLevelFull + high-risk indicators → Ask
//   - AutonomyLevelLow → Ask for any agent task
//   - AutonomyLevelNone → Deny
//
// Invariants:
//   - Ask never auto-upgrades to Allow
//   - Deny is sticky for the same tool_use_id
func (pr *PermissionResolver) Resolve(autonomy agent.AutonomyLevel, metadata map[string]string) Permission {
	if autonomy == agent.AutonomyLevelNone {
		return PermissionDeny
	}
	if autonomy == agent.AutonomyLevelLow {
		return PermissionAsk
	}
	// AutonomyLevelFull — check for high-risk indicators
	if metadata != nil {
		for _, key := range pr.HighRiskKeys {
			if v, ok := metadata[key]; ok && v == "true" {
				return PermissionAsk
			}
		}
	}
	return PermissionAllow
}
