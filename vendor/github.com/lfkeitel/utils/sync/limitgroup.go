package sync

import (
	"sync"
	"sync/atomic"
)

// A LimitGroup is a group used to limit the number of something. An example would be to
// limit the number of goroutines running at the same time
type LimitGroup struct {
	current int32
	done    chan bool
	max     int32
	wg      sync.WaitGroup
	sync.Mutex
}

// NewLimitGroup creates a limit group with l as the limit
func NewLimitGroup(l int32) *LimitGroup {
	lg := &LimitGroup{}
	lg.done = make(chan bool)
	lg.wg = sync.WaitGroup{}
	lg.Limit(l)
	return lg
}

// Limit sets the maximum limit, a limit of 0 (default) indicates no limit.
func (c *LimitGroup) Limit(l int32) {
	c.max = l
}

// Add adds delta number to the LimitGroup
func (c *LimitGroup) Add(delta int) {
	c.Lock()
	c.add(delta)
	c.Unlock()
}

func (c *LimitGroup) add(delta int) {
	c.current = atomic.AddInt32(&c.current, int32(delta))
	c.wg.Add(delta)
}

// Done removes one from the current LimitGroup
func (c *LimitGroup) Done() {
	c.Lock()
	c.add(-1)
	if c.max == 0 || c.current < c.max {
		// Don't wait if the channel is full
		select {
		case c.done <- true:
		default:
		}
	}
	c.Unlock()
}

// Wait will block until there are less than the max limit available. This would be placed
// at the end of a loop that's starting goroutines to wait for an available slot
// before starting a new one.
func (c *LimitGroup) Wait() {
	if c.max == 0 || atomic.LoadInt32(&c.current) < c.max {
		return
	}
	<-c.done
}

// WaitAll will block until the counter reaches 0.
func (c *LimitGroup) WaitAll() {
	c.wg.Wait()
}
