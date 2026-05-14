package tmp

import (
	"sync"
	"testing"
)

func TestCounter_Concurrent(t *testing.T) {
	c := &Counter{}
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); c.Inc() }()
	}
	wg.Wait()
	if c.Get() != 1000 {
		t.Errorf("got %d, want 1000", c.Get())
	}
}
