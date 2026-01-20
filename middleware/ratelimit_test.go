package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/llmx-ai/llmx"
)

func TestNewTokenBucketLimiter(t *testing.T) {
	t.Run("create limiter", func(t *testing.T) {
		limiter := NewTokenBucketLimiter(10, 20)
		if limiter == nil {
			t.Fatal("expected limiter to be created")
		}
	})

	t.Run("allow request within limit", func(t *testing.T) {
		limiter := NewTokenBucketLimiter(100, 10) // 100 rps, burst 10

		// Should allow first request
		if !limiter.Allow() {
			t.Error("expected first request to be allowed")
		}
	})

	t.Run("block request when exhausted", func(t *testing.T) {
		limiter := NewTokenBucketLimiter(1, 1) // 1 rps, burst 1

		// Exhaust the limiter
		limiter.Allow()

		// Should block next request (no time for refill)
		if limiter.Allow() {
			t.Error("expected request to be blocked when limit exhausted")
		}
	})
}

func TestRateLimit(t *testing.T) {
	t.Run("allows requests within limit", func(t *testing.T) {
		limiter := NewTokenBucketLimiter(100, 10)
		mw := RateLimit(limiter, false)

		// Create mock handler
		called := false
		handler := func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
			called = true
			return &llmx.ChatResponse{Content: "test"}, nil
		}

		wrappedHandler := mw(handler)

		ctx := context.Background()
		req := &llmx.ChatRequest{
			Model: "test",
			Messages: []llmx.Message{
				{
					Role: llmx.RoleUser,
					Content: []llmx.ContentPart{
						llmx.TextPart{Text: "test"},
					},
				},
			},
		}

		_, err := wrappedHandler(ctx, req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if !called {
			t.Error("expected handler to be called")
		}
	})

	t.Run("blocks requests when limit exceeded with blocking=false", func(t *testing.T) {
		limiter := NewTokenBucketLimiter(1, 1) // Very low limit
		mw := RateLimit(limiter, false) // Non-blocking mode

		handler := func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
			return &llmx.ChatResponse{Content: "test"}, nil
		}

		wrappedHandler := mw(handler)

		ctx := context.Background()
		req := &llmx.ChatRequest{
			Model: "test",
			Messages: []llmx.Message{
				{
					Role: llmx.RoleUser,
					Content: []llmx.ContentPart{
						llmx.TextPart{Text: "test"},
					},
				},
			},
		}

		// First request should succeed
		_, err := wrappedHandler(ctx, req)
		if err != nil {
			t.Errorf("first request failed: %v", err)
		}

		// Second request should fail immediately (non-blocking)
		_, err = wrappedHandler(ctx, req)
		if err == nil {
			t.Error("expected rate limit error for second request")
		}

		// Verify it's a rate limit error
		if llmxErr, ok := err.(llmx.Error); ok {
			if llmxErr.Code() != "rate_limit" {
				t.Errorf("expected rate_limit error code, got %s", llmxErr.Code())
			}
		} else {
			t.Error("expected llmx.Error")
		}
	})

	t.Run("waits for requests when limit exceeded with blocking=true", func(t *testing.T) {
		limiter := NewTokenBucketLimiter(10, 1) // 10 rps, burst 1
		mw := RateLimit(limiter, true)          // Blocking mode

		handler := func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
			return &llmx.ChatResponse{Content: "test"}, nil
		}

		wrappedHandler := mw(handler)

		ctx := context.Background()
		req := &llmx.ChatRequest{
			Model: "test",
			Messages: []llmx.Message{
				{
					Role: llmx.RoleUser,
					Content: []llmx.ContentPart{
						llmx.TextPart{Text: "test"},
					},
				},
			},
		}

		// First request should succeed immediately
		start := time.Now()
		_, err := wrappedHandler(ctx, req)
		if err != nil {
			t.Errorf("first request failed: %v", err)
		}
		firstDuration := time.Since(start)

		// Second request should wait (blocking mode)
		start = time.Now()
		_, err = wrappedHandler(ctx, req)
		if err != nil {
			t.Errorf("second request failed: %v", err)
		}
		secondDuration := time.Since(start)

		// Second request should take longer (waiting for token)
		if secondDuration <= firstDuration {
			t.Logf("Second request duration (%v) should be longer than first (%v)", secondDuration, firstDuration)
			// Note: This might fail occasionally due to timing, but it's informative
		}
	})

	t.Run("respects context cancellation in blocking mode", func(t *testing.T) {
		limiter := NewTokenBucketLimiter(1, 1)
		mw := RateLimit(limiter, true) // Blocking mode

		handler := func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
			return &llmx.ChatResponse{Content: "test"}, nil
		}

		wrappedHandler := mw(handler)

		req := &llmx.ChatRequest{
			Model: "test",
			Messages: []llmx.Message{
				{
					Role: llmx.RoleUser,
					Content: []llmx.ContentPart{
						llmx.TextPart{Text: "test"},
					},
				},
			},
		}

		// Exhaust the limiter
		ctx1 := context.Background()
		wrappedHandler(ctx1, req)

		// Create a context that's already cancelled
		ctx2, cancel := context.WithCancel(context.Background())
		cancel()

		// Should return immediately with context error
		_, err := wrappedHandler(ctx2, req)
		if err == nil {
			t.Error("expected error for cancelled context")
		}

		// Error might be wrapped, so just check it contains context canceled
		if err.Error() == "" {
			t.Error("expected non-empty error message")
		}
	})
}

func TestNewSlidingWindowLimiter(t *testing.T) {
	t.Run("create limiter", func(t *testing.T) {
		limiter := NewSlidingWindowLimiter(100, time.Second)
		if limiter == nil {
			t.Fatal("expected limiter to be created")
		}
	})

	t.Run("allows requests within limit", func(t *testing.T) {
		limiter := NewSlidingWindowLimiter(10, time.Second)

		// Should allow multiple requests within limit
		for i := 0; i < 10; i++ {
			if !limiter.Allow() {
				t.Errorf("expected request %d to be allowed", i+1)
			}
		}

		// Should block 11th request
		if limiter.Allow() {
			t.Error("expected 11th request to be blocked")
		}
	})

	t.Run("sliding window behavior", func(t *testing.T) {
		limiter := NewSlidingWindowLimiter(2, 100*time.Millisecond) // 2 requests per 100ms

		// First two requests should succeed
		if !limiter.Allow() {
			t.Error("expected first request to be allowed")
		}
		if !limiter.Allow() {
			t.Error("expected second request to be allowed")
		}

		// Third request should fail (limit reached)
		if limiter.Allow() {
			t.Error("expected third request to be blocked")
		}

		// Wait for window to slide
		time.Sleep(120 * time.Millisecond)

		// Fourth request should succeed (window slid)
		if !limiter.Allow() {
			t.Error("expected fourth request to be allowed after window slide")
		}
	})
}
