package provider

import (
	"context"
	"strings"
	"sync/atomic"
	"time"
)

// FallbackProvider tries providers in order; on 429/rate-limit errors,
// switches to the next provider. After a cooldown, retries the primary.
// Includes a per-request rate limiter to avoid triggering 429 in the first place.
type FallbackProvider struct {
	providers []ModelProvider
	cooldown  time.Duration
	active    atomic.Int32
	coolUntil atomic.Int64 // unix nano when primary becomes eligible again
	minDelay  time.Duration
	lastReq   atomic.Int64 // unix nano of last request
}

// NewFallbackProvider creates a provider that falls back on 429 errors.
// providers[0] is primary; others are fallbacks tried in order.
func NewFallbackProvider(cooldown time.Duration, providers ...ModelProvider) *FallbackProvider {
	if cooldown == 0 {
		cooldown = 60 * time.Second
	}
	return &FallbackProvider{
		providers: providers,
		cooldown:  cooldown,
		minDelay:  5 * time.Second, // max ~12 req/min, under MiniMax's 14/min limit
	}
}

func (f *FallbackProvider) Execute(ctx context.Context, req *ModelRequest) (*ModelResponse, error) {
	// Rate limiting: wait if last request was too recent
	if f.minDelay > 0 {
		last := f.lastReq.Load()
		if last > 0 {
			elapsed := time.Since(time.Unix(0, last))
			if elapsed < f.minDelay {
				wait := f.minDelay - elapsed
				select {
				case <-time.After(wait):
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
		}
		f.lastReq.Store(time.Now().UnixNano())
	}

	// If cooldown expired, try primary again
	if cu := f.coolUntil.Load(); cu > 0 && time.Now().UnixNano() >= cu {
		f.active.Store(0)
		f.coolUntil.Store(0)
	}

	idx := int(f.active.Load())
	resp, err := f.providers[idx].Execute(ctx, req)
	if err != nil && isRateLimitError(err) && len(f.providers) > 1 {
		// Switch to next provider
		next := (idx + 1) % len(f.providers)
		f.active.Store(int32(next))
		if idx == 0 {
			f.coolUntil.Store(time.Now().Add(f.cooldown).UnixNano())
		}
		return f.providers[next].Execute(ctx, req)
	}
	return resp, err
}

func isRateLimitError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "status 429") ||
		strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "Rate limit") ||
		strings.Contains(msg, "Too Many Requests")
}
