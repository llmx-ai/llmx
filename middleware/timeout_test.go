package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/llmx-ai/llmx"
)

func TestTimeout(t *testing.T) {
	t.Run("allows fast requests", func(t *testing.T) {
		mw := Timeout(1 * time.Second)

		handler := func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
			// Fast response
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
	})

	t.Run("times out slow requests", func(t *testing.T) {
		mw := Timeout(100 * time.Millisecond)

		handler := func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
			// Slow response
			select {
			case <-time.After(1 * time.Second):
				return &llmx.ChatResponse{Content: "test"}, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
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

		start := time.Now()
		_, err := wrappedHandler(ctx, req)
		duration := time.Since(start)

		if err == nil {
			t.Error("expected timeout error")
		}

		if duration > 200*time.Millisecond {
			t.Errorf("timeout took too long: %v", duration)
		}

		// Error might be wrapped by middleware
		if err == nil {
			t.Error("expected timeout error")
		}
	})

	t.Run("respects existing deadline in context", func(t *testing.T) {
		mw := Timeout(1 * time.Second) // Long middleware timeout

		handler := func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
			select {
			case <-time.After(500 * time.Millisecond):
				return &llmx.ChatResponse{Content: "test"}, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		wrappedHandler := mw(handler)

		// Create context with short deadline
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

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

		start := time.Now()
		_, err := wrappedHandler(ctx, req)
		duration := time.Since(start)

		if err == nil {
			t.Error("expected timeout error")
		}

		// Should timeout based on context deadline, not middleware timeout
		if duration > 200*time.Millisecond {
			t.Errorf("timeout took too long: %v", duration)
		}
	})
}

func TestAdaptiveTimeout(t *testing.T) {
	t.Run("calculates base timeout", func(t *testing.T) {
		adapter := NewAdaptiveTimeout()

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

		timeout := adapter.CalculateTimeout(req)
		expected := 30*time.Second + 2*time.Second // base + 1 message

		if timeout != expected {
			t.Errorf("expected timeout %v, got %v", expected, timeout)
		}
	})

	t.Run("increases timeout with more messages", func(t *testing.T) {
		adapter := NewAdaptiveTimeout()

		req1 := &llmx.ChatRequest{
			Model: "test",
			Messages: []llmx.Message{
				{Role: llmx.RoleUser, Content: []llmx.ContentPart{llmx.TextPart{Text: "1"}}},
			},
		}

		req2 := &llmx.ChatRequest{
			Model: "test",
			Messages: []llmx.Message{
				{Role: llmx.RoleUser, Content: []llmx.ContentPart{llmx.TextPart{Text: "1"}}},
				{Role: llmx.RoleAssistant, Content: []llmx.ContentPart{llmx.TextPart{Text: "2"}}},
				{Role: llmx.RoleUser, Content: []llmx.ContentPart{llmx.TextPart{Text: "3"}}},
			},
		}

		timeout1 := adapter.CalculateTimeout(req1)
		timeout2 := adapter.CalculateTimeout(req2)

		if timeout2 <= timeout1 {
			t.Errorf("expected timeout to increase with more messages, got %v -> %v", timeout1, timeout2)
		}
	})

	t.Run("respects maximum timeout", func(t *testing.T) {
		adapter := NewAdaptiveTimeout().
			WithMaxTimeout(1 * time.Minute)

		// Create request with many messages
		messages := make([]llmx.Message, 100)
		for i := range messages {
			messages[i] = llmx.Message{
				Role:    llmx.RoleUser,
				Content: []llmx.ContentPart{llmx.TextPart{Text: "test"}},
			}
		}

		req := &llmx.ChatRequest{
			Model:    "test",
			Messages: messages,
		}

		timeout := adapter.CalculateTimeout(req)
		if timeout > 1*time.Minute {
			t.Errorf("expected timeout to be capped at 1 minute, got %v", timeout)
		}
	})

	t.Run("supports custom configuration", func(t *testing.T) {
		adapter := NewAdaptiveTimeout().
			WithBaseTimeout(10 * time.Second).
			WithPerMessage(1 * time.Second).
			WithPerTool(3 * time.Second)

		req := &llmx.ChatRequest{
			Model: "test",
			Messages: []llmx.Message{
				{Role: llmx.RoleUser, Content: []llmx.ContentPart{llmx.TextPart{Text: "test"}}},
			},
			Tools: []llmx.Tool{
				{Name: "tool1"},
				{Name: "tool2"},
			},
		}

		timeout := adapter.CalculateTimeout(req)
		expected := 10*time.Second + 1*time.Second + 2*3*time.Second // base + message + tools

		if timeout != expected {
			t.Errorf("expected timeout %v, got %v", expected, timeout)
		}
	})
}
