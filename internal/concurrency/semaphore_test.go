package concurrency

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSemaphoreBasic(t *testing.T) {
	sem := NewSemaphore(2)

	var counter int64
	var wg sync.WaitGroup

	// Launch 5 goroutines, only 2 should run concurrently
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem.Acquire(1)
			defer sem.Release(1)
			atomic.AddInt64(&counter, 1)
			time.Sleep(50 * time.Millisecond)
		}()
	}

	wg.Wait()
	if counter != 5 {
		t.Errorf("counter = %d, want 5", counter)
	}
}

func TestSemaphoreMultiple(t *testing.T) {
	sem := NewSemaphore(2)

	var concurrent int64
	var maxConcurrent int64
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem.Acquire(1)
			defer sem.Release(1)

			cur := atomic.AddInt64(&concurrent, 1)
			for {
				max := atomic.LoadInt64(&maxConcurrent)
				if cur <= max {
					break
				}
				if atomic.CompareAndSwapInt64(&maxConcurrent, max, cur) {
					break
				}
			}

			time.Sleep(10 * time.Millisecond)
			atomic.AddInt64(&concurrent, -1)
		}()
	}

	wg.Wait()
	if maxConcurrent > 2 {
		t.Errorf("max concurrent = %d, expected <= 2", maxConcurrent)
	}
}

func BenchmarkSemaphore(b *testing.B) {
	sem := NewSemaphore(100)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sem.Acquire(1)
			sem.Release(1)
		}
	})
}
