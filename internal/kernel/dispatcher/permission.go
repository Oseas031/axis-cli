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

// v1 (human-agent-scheduling spec): Resolve is wired into executeAgentTask
// behind RequiresConfirmation so only tasks that explicitly declare autonomy
// or carry high-risk keys enter the ask path.

// RequiresConfirmation reports whether a task's metadata explicitly opts into
// permission gating: an explicitly declared autonomy level or any high-risk
// key. Tasks that never declared autonomy keep their legacy behavior
// (implicit Low runs ungated) so wiring the ask path does not retroactively
// gate every historical task; declarations are conscious acts.
func (pr *PermissionResolver) RequiresConfirmation(metadata map[string]string) bool {
	if metadata == nil {
		return false
	}
	if _, ok := metadata["agent.autonomy_level"]; ok {
		return true
	}
	if _, ok := metadata["autonomy_level"]; ok {
		return true
	}
	for _, key := range pr.HighRiskKeys {
		if _, ok := metadata[key]; ok {
			return true
		}
	}
	return false
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
