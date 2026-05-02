// Package concurrency provides shared concurrency primitives for the stress test engine.
package concurrency

import (
	"sync"
	"time"
)

// TokenBucket implements the token bucket algorithm for RPS rate limiting.
//
// The bucket starts fully filled. Tokens are added at a fixed rate (refillRate
// tokens per second). Each call to Wait() consumes one token; if none are
// available, it blocks until a token is added.
//
// This provides smooth traffic shaping with burst allowance up to the
// configured capacity.
type TokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	capacity   float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

// NewTokenBucket creates a fully-filled bucket at the given rate and burst capacity.
//
// Parameters:
//   - rate: target requests per second (refill rate)
//   - burst: maximum concurrent tokens (allowing short bursts)
func NewTokenBucket(rate uint64, burst uint64) *TokenBucket {
	if rate == 0 {
		rate = 1
	}
	if burst == 0 {
		burst = 1
	}
	return &TokenBucket{
		tokens:     float64(burst),
		capacity:   float64(burst),
		refillRate: float64(rate),
		lastRefill: time.Now(),
	}
}

// Allow checks if a request can proceed immediately (non-blocking).
// Returns true if a token is available, false otherwise.
func (tb *TokenBucket) Allow() bool {
	tb.refill()
	tb.mu.Lock()
	defer tb.mu.Unlock()
	if tb.tokens >= 1.0 {
		tb.tokens--
		return true
	}
	return false
}

// Wait blocks until a token becomes available, then consumes it.
func (tb *TokenBucket) Wait() {
	tb.refill()
	tb.mu.Lock()
	defer tb.mu.Unlock()
	for tb.tokens < 1.0 {
		sleep := time.Duration((1.0-tb.tokens)/tb.refillRate*1000) * time.Millisecond
		if sleep < time.Millisecond {
			sleep = time.Millisecond
		}
		tb.mu.Unlock()
		time.Sleep(sleep)
		tb.refill()
		tb.mu.Lock()
	}
	tb.tokens--
}

// refill adds tokens based on elapsed time since last refill.
// Must be called with mu held.
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens += elapsed * tb.refillRate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}
	tb.lastRefill = now
}

// CurrentTokens returns the current token count (approximate, non-blocking).
func (tb *TokenBucket) CurrentTokens() float64 {
	tb.refill()
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return tb.tokens
}
