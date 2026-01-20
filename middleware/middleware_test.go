package middleware

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/llmx-ai/llmx"
)

// Mock handler for testing
func mockHandler(resp *llmx.ChatResponse, err error) Handler {
	return func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
		return resp, err
	}
}

func TestChain(t *testing.T) {
	called := []string{}

	m1 := func(next Handler) Handler {
		return func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
			called = append(called, "m1-before")
			resp, err := next(ctx, req)
			called = append(called, "m1-after")
			return resp, err
		}
	}

	m2 := func(next Handler) Handler {
		return func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
			called = append(called, "m2-before")
			resp, err := next(ctx, req)
			called = append(called, "m2-after")
			return resp, err
		}
	}

	handler := mockHandler(&llmx.ChatResponse{Content: "test"}, nil)
	chained := Chain(m1, m2)(handler)

	_, err := chained(context.Background(), &llmx.ChatRequest{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check execution order
	expected := []string{"m1-before", "m2-before", "m2-after", "m1-after"}
	if len(called) != len(expected) {
		t.Errorf("Expected %d calls, got %d", len(expected), len(called))
	}

	for i, exp := range expected {
		if i >= len(called) || called[i] != exp {
			t.Errorf("Expected call %d to be '%s', got '%s'", i, exp, called[i])
		}
	}
}

func TestLogging(t *testing.T) {
	// Create middleware
	middleware := Logging(nil)

	handler := mockHandler(&llmx.ChatResponse{Content: "test"}, nil)
	wrapped := middleware(handler)

	_, err := wrapped(context.Background(), &llmx.ChatRequest{Model: "test-model"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestRetry_Success(t *testing.T) {
	attempts := 0

	handler := func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
		attempts++
		if attempts < 2 {
			return nil, llmx.NewRateLimitError("rate limited", 0)
		}
		return &llmx.ChatResponse{Content: "success"}, nil
	}

	middleware := Retry(3, NewExponentialBackoff())
	wrapped := middleware(handler)

	resp, err := wrapped(context.Background(), &llmx.ChatRequest{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resp.Content != "success" {
		t.Errorf("Expected 'success', got '%s'", resp.Content)
	}
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestRetry_NonRetryableError(t *testing.T) {
	attempts := 0
	nonRetryableErr := errors.New("non-retryable error")

	handler := func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
		attempts++
		return nil, nonRetryableErr
	}

	middleware := Retry(3, NewExponentialBackoff())
	wrapped := middleware(handler)

	_, err := wrapped(context.Background(), &llmx.ChatRequest{})
	if err != nonRetryableErr {
		t.Errorf("Expected non-retryable error, got %v", err)
	}
	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", attempts)
	}
}

func TestCacheMiddleware(t *testing.T) {
	cache := NewMemoryCache()
	middleware := CacheMiddleware(cache, 10*time.Minute)

	callCount := 0
	handler := func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
		callCount++
		return &llmx.ChatResponse{Content: "test"}, nil
	}

	wrapped := middleware(handler)

	// First call (miss)
	resp1, err := wrapped(context.Background(), &llmx.ChatRequest{Model: "test"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resp1.Content != "test" {
		t.Errorf("Expected 'test', got '%s'", resp1.Content)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}

	// Second call (hit)
	resp2, err := wrapped(context.Background(), &llmx.ChatRequest{Model: "test"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resp2.Content != "test" {
		t.Errorf("Expected 'test', got '%s'", resp2.Content)
	}
	if callCount != 1 {
		t.Errorf("Expected still 1 call (cached), got %d", callCount)
	}
}
