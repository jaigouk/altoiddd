package mcp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(5, time.Second)

	for i := range 5 {
		assert.True(t, rl.Allow("tools/call"), "request %d should be allowed", i)
	}
}

func TestRateLimiter_RejectsOverLimit(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(3, time.Second)

	// Exhaust tokens.
	for range 3 {
		rl.Allow("tools/call")
	}

	assert.False(t, rl.Allow("tools/call"), "4th request should be rejected")
	assert.False(t, rl.Allow("tools/call"), "5th request should also be rejected")
}

func TestRateLimiter_ResetsAfterWindow(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(2, 10*time.Millisecond)

	// Exhaust tokens.
	rl.Allow("tools/call")
	rl.Allow("tools/call")
	assert.False(t, rl.Allow("tools/call"), "should be rate limited")

	// Wait for refill.
	time.Sleep(15 * time.Millisecond)
	assert.True(t, rl.Allow("tools/call"), "should be allowed after refill")
}

func TestRateLimiter_PerMethodBuckets(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(1, time.Second)

	assert.True(t, rl.Allow("tools/call"), "tools/call first request allowed")
	assert.False(t, rl.Allow("tools/call"), "tools/call second request rejected")

	// Different method gets its own bucket.
	assert.True(t, rl.Allow("resources/read"), "resources/read should have its own bucket")
}

func TestRateLimiter_ZeroTokens(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(0, time.Second)

	assert.False(t, rl.Allow("tools/call"), "zero-token limiter should reject all")
}

func TestNewRateLimiter_PanicsOnZeroDuration(t *testing.T) {
	t.Parallel()
	assert.Panics(t, func() {
		NewRateLimiter(5, 0)
	}, "zero refillRate should panic")
}

func TestNewRateLimiter_PanicsOnNegativeDuration(t *testing.T) {
	t.Parallel()
	assert.Panics(t, func() {
		NewRateLimiter(5, -time.Second)
	}, "negative refillRate should panic")
}
