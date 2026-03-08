package mcp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RateLimiter implements a per-key token bucket rate limiter.
// Keys are typically MCP method names (e.g. "tools/call").
type RateLimiter struct {
	maxTokens  int
	refillRate time.Duration
	mu         sync.Mutex
	buckets    map[string]*bucket
}

type bucket struct {
	tokens    int
	lastCheck time.Time
}

// NewRateLimiter creates a rate limiter that allows maxTokens requests per
// refillRate duration. Each key gets its own independent bucket.
// Panics if refillRate is zero or negative.
func NewRateLimiter(maxTokens int, refillRate time.Duration) *RateLimiter {
	if refillRate <= 0 {
		panic("rate_limiter: refillRate must be positive")
	}
	return &RateLimiter{
		maxTokens:  maxTokens,
		refillRate: refillRate,
		buckets:    make(map[string]*bucket),
	}
}

// Allow checks if a request with the given key is allowed.
// Returns true if allowed, false if rate limited.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, exists := rl.buckets[key]
	if !exists {
		b = &bucket{tokens: rl.maxTokens, lastCheck: now}
		rl.buckets[key] = b
	}

	// Refill tokens based on elapsed time.
	elapsed := now.Sub(b.lastCheck)
	refills := int(elapsed / rl.refillRate)
	if refills > 0 {
		b.tokens += refills
		if b.tokens > rl.maxTokens {
			b.tokens = rl.maxTokens
		}
		b.lastCheck = now
	}

	if b.tokens > 0 {
		b.tokens--
		return true
	}
	return false
}

// RateLimitMiddleware returns MCP middleware that rate-limits requests
// by method name. Non-tool methods are rate-limited by their method string.
func RateLimitMiddleware(limiter *RateLimiter) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			if !limiter.Allow(method) {
				return nil, fmt.Errorf("rate limit exceeded for %s", method)
			}
			return next(ctx, method, req)
		}
	}
}
