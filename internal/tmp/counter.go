package tmp

import "sync/atomic"

// Counter is a simple counter. Must be safe for concurrent use.
type Counter struct {
	n int64
}

func (c *Counter) Inc() { atomic.AddInt64(&c.n, 1) }
func (c *Counter) Get() int { return int(atomic.LoadInt64(&c.n)) }