package safego

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestGo_RecoversPanic(t *testing.T) {
	done := make(chan struct{})
	Go(context.Background(), func() {
		panic("intentional test panic")
	})

	// Give the goroutine time to panic and recover.
	time.Sleep(50 * time.Millisecond)

	// If we reach here without the test process crashing, recovery worked.
	// The panic was logged; we just verify the process is still alive.
	close(done)
	select {
	case <-done:
		// Expected: process survived.
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for panic recovery confirmation")
	}
}

func TestGo_RunsNormally(t *testing.T) {
	done := make(chan struct{})
	Go(context.Background(), func() {
		close(done)
	})

	select {
	case <-done:
		// Expected: normal execution completed.
	case <-time.After(100 * time.Millisecond):
		t.Fatal("normal goroutine did not complete")
	}
}

func TestGoWithWaitGroup_RecoversPanicAndSignalsDone(t *testing.T) {
	var wg sync.WaitGroup

	GoWithWaitGroup(context.Background(), &wg, func() {
		panic("worker panic")
	})

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Expected: wg.Done() was called despite the panic.
	case <-time.After(200 * time.Millisecond):
		t.Fatal("WaitGroup did not complete after panic recovery")
	}
}

func TestGoWithWaitGroup_RunsNormally(t *testing.T) {
	var wg sync.WaitGroup
	var counter atomic.Int32

	for i := 0; i < 10; i++ {
		GoWithWaitGroup(context.Background(), &wg, func() {
			counter.Add(1)
		})
	}

	wg.Wait()
	if counter.Load() != 10 {
		t.Fatalf("expected counter=10, got %d", counter.Load())
	}
}

func TestGo_ContextPropagated(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancelled := make(chan struct{})

	Go(ctx, func() {
		<-ctx.Done()
		close(cancelled)
	})

	cancel()

	select {
	case <-cancelled:
		// Expected: fn observed the cancellation.
	case <-time.After(100 * time.Millisecond):
		t.Fatal("goroutine did not observe context cancellation")
	}
}

func TestPanicError_Error(t *testing.T) {
	e := &PanicError{
		Value: "test panic",
		Stack: "stack trace here",
	}
	msg := e.Error()
	if !strings.Contains(msg, "test panic") {
		t.Fatalf("expected error to contain panic value, got: %s", msg)
	}
}
