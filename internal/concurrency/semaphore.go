// Package concurrency provides shared concurrency primitives for the stress test engine.
package concurrency

// Semaphore provides goroutine concurrency limiting via a buffered channel.
type Semaphore struct {
	ch chan struct{}
}

// NewSemaphore creates a semaphore that allows max concurrent operations.
func NewSemaphore(max int) *Semaphore {
	return &Semaphore{ch: make(chan struct{}, max)}
}

// Acquire blocks until the semaphore count allows, then acquires n slots.
func (s *Semaphore) Acquire(n int64) {
	for range int(n) {
		s.ch <- struct{}{}
	}
}

// Release releases n slots back to the semaphore.
func (s *Semaphore) Release(n int64) {
	for range int(n) {
		<-s.ch
	}
}
