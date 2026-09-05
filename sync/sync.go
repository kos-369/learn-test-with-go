package sync

import "sync/atomic"

type Counter struct {
	value atomic.Int64
}

func NewCounter() *Counter {
	return &Counter{}
}

func (c *Counter) Inc() {
	c.value.Add(1)
}

func (c *Counter) Value() int64 {
	return c.value.Load()
}
