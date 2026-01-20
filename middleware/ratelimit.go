package middleware

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/llmx-ai/llmx"
	"golang.org/x/time/rate"
)

// RateLimiter interface for custom rate limiters
type RateLimiter interface {
	Allow() bool
	Wait(ctx context.Context) error
}

// TokenBucketLimiter implements token bucket algorithm
type TokenBucketLimiter struct {
	limiter *rate.Limiter
}

// NewTokenBucketLimiter creates a new token bucket rate limiter
func NewTokenBucketLimiter(rps float64, burst int) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		limiter: rate.NewLimiter(rate.Limit(rps), burst),
	}
}

// Allow checks if a request can proceed
func (t *TokenBucketLimiter) Allow() bool {
	return t.limiter.Allow()
}

// Wait waits until the request can proceed
func (t *TokenBucketLimiter) Wait(ctx context.Context) error {
	return t.limiter.Wait(ctx)
}

// SlidingWindowLimiter implements sliding window algorithm
type SlidingWindowLimiter struct {
	mu           sync.Mutex
	requests     []time.Time
	maxRequests  int
	windowPeriod time.Duration
}

// NewSlidingWindowLimiter creates a new sliding window rate limiter
func NewSlidingWindowLimiter(maxRequests int, window time.Duration) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		requests:     make([]time.Time, 0, maxRequests),
		maxRequests:  maxRequests,
		windowPeriod: window,
	}
}

// Allow checks if a request can proceed
func (s *SlidingWindowLimiter) Allow() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-s.windowPeriod)

	// Remove old requests
	validRequests := s.requests[:0]
	for _, t := range s.requests {
		if t.After(cutoff) {
			validRequests = append(validRequests, t)
		}
	}
	s.requests = validRequests

	// Check if we can add a new request
	if len(s.requests) >= s.maxRequests {
		return false
	}

	s.requests = append(s.requests, now)
	return true
}

// Wait waits until the request can proceed
func (s *SlidingWindowLimiter) Wait(ctx context.Context) error {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		if s.Allow() {
			return nil
		}

		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// RateLimit creates a rate limiting middleware
func RateLimit(limiter RateLimiter, wait bool) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
			if wait {
				// Wait for rate limit
				if err := limiter.Wait(ctx); err != nil {
					return nil, llmx.NewRateLimitError(
						fmt.Sprintf("rate limit wait failed: %v", err),
						0,
					)
				}
			} else {
				// Check rate limit without waiting
				if !limiter.Allow() {
					return nil, llmx.NewRateLimitError(
						"rate limit exceeded",
						1*time.Second, // Suggested retry after
					)
				}
			}

			return next(ctx, req)
		}
	}
}

// RateLimitByModel creates a per-model rate limiting middleware
type ModelRateLimiter struct {
	mu       sync.RWMutex
	limiters map[string]RateLimiter
	factory  func() RateLimiter
}

// NewModelRateLimiter creates a new per-model rate limiter
func NewModelRateLimiter(factory func() RateLimiter) *ModelRateLimiter {
	return &ModelRateLimiter{
		limiters: make(map[string]RateLimiter),
		factory:  factory,
	}
}

// GetLimiter returns the limiter for a model
func (m *ModelRateLimiter) GetLimiter(model string) RateLimiter {
	m.mu.RLock()
	limiter, ok := m.limiters[model]
	m.mu.RUnlock()

	if ok {
		return limiter
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, ok := m.limiters[model]; ok {
		return limiter
	}

	limiter = m.factory()
	m.limiters[model] = limiter
	return limiter
}

// RateLimitPerModel creates a per-model rate limiting middleware
func RateLimitPerModel(modelLimiter *ModelRateLimiter, wait bool) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
			limiter := modelLimiter.GetLimiter(req.Model)

			if wait {
				if err := limiter.Wait(ctx); err != nil {
					return nil, llmx.NewRateLimitError(
						fmt.Sprintf("rate limit wait failed: %v", err),
						0,
					)
				}
			} else {
				if !limiter.Allow() {
					return nil, llmx.NewRateLimitError(
						fmt.Sprintf("rate limit exceeded for model: %s", req.Model),
						1*time.Second,
					)
				}
			}

			return next(ctx, req)
		}
	}
}
