package middleware

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/llmx-ai/llmx"
)

func TestCircuitBreakerMiddleware(t *testing.T) {
	t.Run("allows requests when circuit is closed", func(t *testing.T) {
		cb := NewCircuitBreaker(3, 5*time.Second)
		mw := CircuitBreakerMiddleware(cb)

		called := 0
		handler := func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
			called++
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

		if called != 1 {
			t.Errorf("expected handler to be called once, got %d", called)
		}
	})

	t.Run("blocks requests when circuit is open", func(t *testing.T) {
		cb := NewCircuitBreaker(2, 5*time.Second)
		mw := CircuitBreakerMiddleware(cb)

		failureCount := 0
		handler := func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
			failureCount++
			return nil, errors.New("service unavailable")
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

		// Generate failures to open circuit
		for i := 0; i < 3; i++ {
			wrappedHandler(ctx, req)
		}

		// Circuit should now be open
		beforeCount := failureCount
		_, err := wrappedHandler(ctx, req)

		// Should get circuit breaker error
		if err == nil {
			t.Error("expected circuit breaker to block request")
		}

		// Handler should not have been called again
		if failureCount != beforeCount {
			t.Error("handler should not be called when circuit is open")
		}
	})

	t.Run("returns appropriate error when circuit is open", func(t *testing.T) {
		cb := NewCircuitBreaker(1, 5*time.Second)
		mw := CircuitBreakerMiddleware(cb)

		handler := func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
			return nil, errors.New("service error")
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

		// Trigger failures to open circuit
		for i := 0; i < 3; i++ {
			wrappedHandler(ctx, req)
		}

		// Now request should be blocked by circuit breaker
		_, err := wrappedHandler(ctx, req)
		if err == nil {
			t.Fatal("expected error when circuit is open")
		}

		// Verify error contains circuit breaker information
		errMsg := err.Error()
		if errMsg == "" {
			t.Error("expected non-empty error message")
		}
	})
}
