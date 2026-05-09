package lifecycle

import (
	"context"
	"testing"
	"time"
)

func TestLifecycleManager_Shutdown(t *testing.T) {
	mgr := NewLifecycleManager()

	if !mgr.IsRunning() {
		t.Error("Lifecycle manager should be running initially")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := mgr.Shutdown(ctx)
	if err != nil {
		t.Fatalf("Failed to shutdown: %v", err)
	}

	if mgr.IsRunning() {
		t.Error("Lifecycle manager should not be running after shutdown")
	}
}

func TestLifecycleManager_ShutdownWithExpiredContext(t *testing.T) {
	mgr := NewLifecycleManager()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := mgr.Shutdown(ctx)
	if err == nil {
		t.Error("Shutdown with expired context should return context error")
	}
}

func TestLifecycleManager_ShutdownTwice(t *testing.T) {
	mgr := NewLifecycleManager()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mgr.Shutdown(ctx)

	// Second shutdown should not fail
	err := mgr.Shutdown(ctx)
	if err != nil {
		t.Errorf("Second shutdown should not fail: %v", err)
	}
}
