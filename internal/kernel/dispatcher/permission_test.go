package dispatcher

import (
	"testing"

	"github.com/axis-cli/axis/internal/agent"
)

func TestPermission_String(t *testing.T) {
	if PermissionAllow.String() != "allow" {
		t.Fatal("bad string")
	}
	if PermissionAsk.String() != "ask" {
		t.Fatal("bad string")
	}
	if PermissionDeny.String() != "deny" {
		t.Fatal("bad string")
	}
}

func TestPermissionResolver_FullAutonomy_NoRisk(t *testing.T) {
	pr := NewPermissionResolver()
	got := pr.Resolve(agent.AutonomyLevelFull, map[string]string{"foo": "bar"})
	if got != PermissionAllow {
		t.Fatalf("expected Allow, got %s", got)
	}
}

func TestPermissionResolver_FullAutonomy_HighRisk(t *testing.T) {
	pr := NewPermissionResolver()
	got := pr.Resolve(agent.AutonomyLevelFull, map[string]string{"axis.destructive": "true"})
	if got != PermissionAsk {
		t.Fatalf("expected Ask, got %s", got)
	}
}

func TestPermissionResolver_Limited(t *testing.T) {
	pr := NewPermissionResolver()
	got := pr.Resolve(agent.AutonomyLevelLow, nil)
	if got != PermissionAsk {
		t.Fatalf("expected Ask, got %s", got)
	}
}

func TestPermissionResolver_None(t *testing.T) {
	pr := NewPermissionResolver()
	got := pr.Resolve(agent.AutonomyLevelNone, map[string]string{"foo": "bar"})
	if got != PermissionDeny {
		t.Fatalf("expected Deny, got %s", got)
	}
}

func TestPermissionResolver_NilMetadata(t *testing.T) {
	pr := NewPermissionResolver()
	got := pr.Resolve(agent.AutonomyLevelFull, nil)
	if got != PermissionAllow {
		t.Fatalf("expected Allow for nil metadata, got %s", got)
	}
}
