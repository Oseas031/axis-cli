package vigil

import (
	"os"
	"testing"
)

func TestLocker_LockUnlock(t *testing.T) {
	dir := t.TempDir()
	l := NewLocker(dir)

	// Lock succeeds on first call
	if err := l.Lock("item-1", "test"); err != nil {
		t.Fatalf("first lock failed: %v", err)
	}
	if !l.IsLocked("item-1") {
		t.Fatal("expected item-1 to be locked")
	}

	// Lock again from same process should fail (process is alive)
	if err := l.Lock("item-1", "other"); err == nil {
		t.Fatal("expected error on double lock")
	}

	// Unlock
	if err := l.Unlock("item-1"); err != nil {
		t.Fatalf("unlock failed: %v", err)
	}
	if l.IsLocked("item-1") {
		t.Fatal("expected item-1 to be unlocked")
	}

	// Unlock non-existent is fine
	if err := l.Unlock("item-999"); err != nil {
		t.Fatalf("unlock non-existent failed: %v", err)
	}
}

func TestLocker_StaleLockReclaimed(t *testing.T) {
	dir := t.TempDir()
	l := NewLocker(dir)

	// Write a lock with a dead PID
	if err := os.MkdirAll(l.dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// PID 2147483647 is almost certainly not running
	if err := os.WriteFile(l.lockPath("item-2"), []byte(`{"holder":"dead","pid":2147483647,"started_at":"2026-01-01T00:00:00Z"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Should not be considered locked (process dead)
	if l.IsLocked("item-2") {
		t.Fatal("stale lock should not be considered locked")
	}

	// Should be able to reclaim
	if err := l.Lock("item-2", "new-holder"); err != nil {
		t.Fatalf("reclaim stale lock failed: %v", err)
	}
	if !l.IsLocked("item-2") {
		t.Fatal("expected item-2 to be locked after reclaim")
	}
}

func TestLocker_EmptyID(t *testing.T) {
	l := NewLocker(t.TempDir())
	if err := l.Lock("", "test"); err == nil {
		t.Fatal("expected error for empty id")
	}
}
