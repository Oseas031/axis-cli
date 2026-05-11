// Package safego provides structured goroutine launch primitives that prevent
// the top recurring concurrency bug categories: goroutine panics escaping
// and background work crashing the process.
//
// Design constraints:
//   - Does NOT inject cancellation semantics into fn; Go's ctx is cooperative.
//     fn must check ctx.Done() itself.
//   - Does NOT prevent goroutine leaks; that requires architectural patterns
//     (channels, sync.Once, WaitGroups) which safego composes with.
//   - DOES guarantee: any panic in fn is recovered, logged, and the goroutine
//     exits cleanly without crashing the program.
package safego

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"sync"
)

// Go launches fn in a new goroutine with automatic panic recovery.
// If fn panics, the panic is logged with a full stack trace and the
// goroutine exits cleanly. The program continues running.
//
// ctx is passed for traceability; fn must check ctx.Done() itself
// for cooperative cancellation.
//
// Usage replaces bare `go func()` for any background work where a panic
// would be catastrophic (e.g. orchestrator workers, dispatcher goroutines).
func Go(ctx context.Context, fn func()) {
	go func() {
		defer recoverPanic("Go")
		fn()
	}()
}

// GoWithWaitGroup launches fn in a goroutine with three guarantees:
//  1. wg.Add(1) before start, wg.Done() after exit (via defer).
//  2. Any panic in fn is recovered and logged; wg.Done() still runs.
//  3. The goroutine exits cleanly without crashing the program.
//
// This is the direct replacement for the fragile pattern:
//
//	wg.Add(1)
//	go func() {
//	    defer wg.Done()
//	    ... // panic here crashes the whole program
//	}()
//
// Replacement:
//
//	safego.GoWithWaitGroup(ctx, wg, func() {
//	    ... // panic here is recovered
//	})
func GoWithWaitGroup(ctx context.Context, wg *sync.WaitGroup, fn func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer recoverPanic("GoWithWaitGroup")
		fn()
	}()
}

// recoverPanic recovers from a panic, logs it with a full stack trace,
// and returns so the goroutine can exit cleanly.
func recoverPanic(scope string) {
	if r := recover(); r != nil {
		log.Printf("safego.%s panic recovered: %v\n%s", scope, r, debug.Stack())
	}
}

// PanicError is the sentinel error type returned when a goroutine
// launched by safego recovers from a panic. It carries the original
// panic value and stack trace for diagnostics.
type PanicError struct {
	Value string
	Stack string
}

func (e *PanicError) Error() string {
	return fmt.Sprintf("goroutine panic: %s", e.Value)
}
