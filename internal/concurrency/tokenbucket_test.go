package concurrency

import (
	"testing"
	"time"
)

func TestTokenBucketBurst(t *testing.T) {
	tb := NewTokenBucket(100, 100) // 100 RPS, burst 100

	start := time.Now()
	for i := 0; i < 100; i++ {
		tb.Wait()
	}
	elapsed := time.Since(start)

	// All 100 tokens should be consumed nearly instantly (full bucket)
	if elapsed > 200*time.Millisecond {
		t.Errorf("burst took %v, expected ~0 (100 tokens from full bucket)", elapsed)
	}
}

func TestTokenBucketRate(t *testing.T) {
	tb := NewTokenBucket(100, 1) // 100 RPS, burst 1 (no burst)

	start := time.Now()
	for i := 0; i < 10; i++ {
		tb.Wait()
	}
	elapsed := time.Since(start)

	// Should take approximately 10/100 = 0.1s = 100ms
	// Allow generous tolerance for test timing
	if elapsed < 30*time.Millisecond || elapsed > 500*time.Millisecond {
		t.Errorf("rate test took %v, expected ~100ms", elapsed)
	}
}

func TestTokenBucketAllow(t *testing.T) {
	tb := NewTokenBucket(100, 5)

	for i := 0; i < 5; i++ {
		if !tb.Allow() {
			t.Errorf("Allow() should return true, iteration %d", i)
		}
	}

	// Bucket should be empty now
	if tb.Allow() {
		t.Error("Allow() should return false when bucket is empty")
	}
}

func TestTokenBucketRefill(t *testing.T) {
	tb := NewTokenBucket(100, 1) // 100 RPS, burst 1

	tb.Wait() // consume the one token

	// Should be empty now
	if tb.Allow() {
		t.Error("bucket should be empty after Wait")
	}

	// Wait ~50ms, should have ~5 tokens
	time.Sleep(50 * time.Millisecond)

	if !tb.Allow() {
		t.Error("bucket should have tokens after refill")
	}
}

func TestTokenBucketZeroParams(t *testing.T) {
	// Zero rate/burst should default to 1
	tb := NewTokenBucket(0, 0)
	if tb == nil {
		t.Fatal("NewTokenBucket(0, 0) should not return nil")
	}

	start := time.Now()
	tb.Wait()
	elapsed := time.Since(start)
	// Should complete quickly (initial token available)
	if elapsed > 100*time.Millisecond {
		t.Errorf("Wait should complete quickly with initial token, got %v", elapsed)
	}
}

func BenchmarkTokenBucketWait(b *testing.B) {
	tb := NewTokenBucket(10000, 10000)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tb.Wait()
		}
	})
}
