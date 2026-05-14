package vigil

import (
	"testing"
	"time"
)

func TestTriageStale(t *testing.T) {
	now := time.Now()
	items := []*Item{{
		ID: "s1", Status: StatusPending, CreatedAt: now.Add(-8 * 24 * time.Hour),
	}}
	res, active, _ := Triage(items, now)
	if len(res.Staled) != 1 || res.Staled[0] != "s1" {
		t.Fatal("expected s1 to be staled")
	}
	if active[0].Status != StatusStale {
		t.Fatal("expected status stale")
	}
}

func TestTriageDependencyUpgrade(t *testing.T) {
	now := time.Now()
	target := &Item{ID: "target", Status: StatusPending, Priority: "P2", CreatedAt: now}
	items := []*Item{
		target,
		{ID: "d1", Status: StatusPending, DependsOn: []string{"target"}, CreatedAt: now},
		{ID: "d2", Status: StatusPending, DependsOn: []string{"target"}, CreatedAt: now},
		{ID: "d3", Status: StatusPending, DependsOn: []string{"target"}, CreatedAt: now},
	}
	res, _, _ := Triage(items, now)
	if len(res.Upgraded) != 1 || res.Upgraded[0] != "target" {
		t.Fatalf("expected target upgraded, got %v", res.Upgraded)
	}
	if target.Priority != "P0" {
		t.Fatal("expected P0")
	}
}

func TestTriageArchive(t *testing.T) {
	now := time.Now()
	completed := now.Add(-49 * time.Hour)
	items := []*Item{{
		ID: "c1", Status: StatusCompleted, CompletedAt: &completed, CreatedAt: now.Add(-72 * time.Hour),
	}}
	res, active, toArchive := Triage(items, now)
	if len(res.Archived) != 1 || res.Archived[0] != "c1" {
		t.Fatal("expected c1 archived")
	}
	if len(active) != 0 {
		t.Fatal("expected no active items")
	}
	if len(toArchive) != 1 {
		t.Fatal("expected 1 item to archive")
	}
}

func TestTriageNoChanges(t *testing.T) {
	now := time.Now()
	items := []*Item{{
		ID: "fresh", Status: StatusPending, CreatedAt: now, Priority: "P1",
	}}
	res, active, toArchive := Triage(items, now)
	if len(res.Staled) != 0 || len(res.Upgraded) != 0 || len(res.Archived) != 0 {
		t.Fatal("expected no changes")
	}
	if len(active) != 1 {
		t.Fatal("expected 1 active")
	}
	if len(toArchive) != 0 {
		t.Fatal("expected 0 to archive")
	}
}
