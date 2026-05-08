// Package lifecycle provides agent and task lifecycle management.
package lifecycle

import (
	"context"
	"sync"
)

// LifecycleManager manages agent and task lifecycle
type LifecycleManager interface {
	Shutdown(ctx context.Context) error
	IsRunning() bool
}

// LifecycleManagerImpl implements lifecycle management
type LifecycleManagerImpl struct {
	mu      sync.RWMutex
	running bool
	cancel  context.CancelFunc
	done    chan struct{}
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager() *LifecycleManagerImpl {
	ctx, cancel := context.WithCancel(context.Background())
	lm := &LifecycleManagerImpl{
		running: true,
		cancel:  cancel,
		done:    make(chan struct{}),
	}
	go lm.monitorContext(ctx)
	return lm
}

// monitorContext monitors the context for cancellation
func (lm *LifecycleManagerImpl) monitorContext(ctx context.Context) {
	<-ctx.Done()
	lm.mu.Lock()
	lm.running = false
	close(lm.done)
	lm.mu.Unlock()
}

// Shutdown gracefully shuts down the lifecycle manager
func (lm *LifecycleManagerImpl) Shutdown(ctx context.Context) error {
	lm.mu.Lock()
	if !lm.running {
		lm.mu.Unlock()
		return nil
	}
	lm.running = false
	lm.cancel()
	lm.mu.Unlock()

	select {
	case <-lm.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// IsRunning returns whether the lifecycle manager is running
func (lm *LifecycleManagerImpl) IsRunning() bool {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.running
}
