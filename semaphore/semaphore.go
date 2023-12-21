package semaphore

import (
	"sync"
)

type Semaphore struct {
	cond     *sync.Cond
	capacity int
}

func NewSemaphore(capacity int) *Semaphore {
	s := &Semaphore{capacity: capacity}
	s.cond = sync.NewCond(&sync.Mutex{})
	return s
}

func (s *Semaphore) Down() {
	s.cond.L.Lock()

	for s.capacity == 0 {
		s.cond.Wait()
	}
	s.capacity -= 1
	s.cond.L.Unlock()
}

func (s *Semaphore) Up() {
	s.cond.L.Lock()

	s.capacity += 1
	s.cond.Signal()

	s.cond.L.Unlock()
}
