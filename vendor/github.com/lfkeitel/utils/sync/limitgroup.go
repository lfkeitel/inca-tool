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
	sync.Mutex
}

// NewLimitGroup creates a limit group with l as the limit
func NewLimitGroup(l int32) *LimitGroup {
	lg := &LimitGroup{}
	lg.Limit(l)
	return lg
}

// Limit sets the maximum limit, a limit of 0 (default) indicates no limit.
func (c *LimitGroup) Limit(l int32) {
	c.prepare()
	c.max = l
}

// Add adds delta number to the LimitGroup
func (c *LimitGroup) Add(delta int32) {
	c.prepare()
	c.Lock()
	c.current = atomic.AddInt32(&c.current, delta)
	c.Unlock()
	return
}

// Done removes one from the current LimitGroup
func (c *LimitGroup) Done() {
	c.Add(-1)
	if c.max == 0 || c.current < c.max {
		select {
		case c.done <- true:
		default:
		}
	}
	return
}

// Wait will block until there are less than the max limit available. This would be placed
// at the end of a loop that's starting goroutines to wait for an available slot
// before starting a new one.
func (c *LimitGroup) Wait() {
	c.prepare()
	if c.max == 0 || c.current < c.max {
		return
	}
	<-c.done
	return
}

func (c *LimitGroup) prepare() {
	if c.done == nil {
		c.done = make(chan bool)
	}
}
