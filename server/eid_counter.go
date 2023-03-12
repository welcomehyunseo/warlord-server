package server

import "sync"

var eidCounter *EIDCounter

type EIDCounter struct {
	*sync.RWMutex

	last int32
}

func GetEIDCounter() *EIDCounter {
	if eidCounter == nil {
		eidCounter = &EIDCounter{
			new(sync.RWMutex),
			0,
		}
	}

	return eidCounter
}

func (c *EIDCounter) count() int32 {
	c.Lock()
	defer c.Unlock()

	v := c.last
	c.last++
	return v
}
