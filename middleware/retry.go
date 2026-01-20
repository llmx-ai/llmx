package middleware

import (
	"context"
	"time"

	"github.com/llmx-ai/llmx"
)

// BackoffStrategy defines how to calculate backoff delays
type BackoffStrategy interface {
	Next(attempt int) time.Duration
}

// ExponentialBackoff implements exponential backoff strategy
type ExponentialBackoff struct {
	Base time.Duration
	Max  time.Duration
}

// Next calculates the next backoff duration
func (e *ExponentialBackoff) Next(attempt int) time.Duration {
	duration := e.Base * time.Duration(1<<uint(attempt))
	if duration > e.Max {
		return e.Max
	}
	return duration
}

// NewExponentialBackoff creates a new exponential backoff strategy
func NewExponentialBackoff() BackoffStrategy {
	return &ExponentialBackoff{
		Base: 1 * time.Second,
		Max:  30 * time.Second,
	}
}

// Retry creates a retry middleware
func Retry(maxRetries int, backoff BackoffStrategy) Middleware {
	if backoff == nil {
		backoff = NewExponentialBackoff()
	}

	return func(next Handler) Handler {
		return func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
			var resp *llmx.ChatResponse
			var err error

			for attempt := 0; attempt <= maxRetries; attempt++ {
				resp, err = next(ctx, req)

				// If successful, return
				if err == nil {
					return resp, nil
				}

				// Check if error is retryable
				if !isRetryable(err) {
					return nil, err
				}

				// Don't sleep after last attempt
				if attempt < maxRetries {
					delay := backoff.Next(attempt)
					select {
					case <-time.After(delay):
						// Continue to next attempt
					case <-ctx.Done():
						return nil, ctx.Err()
					}
				}
			}

			return nil, err
		}
	}
}

// isRetryable checks if an error should be retried
func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check for llmx error types
	if llmxErr, ok := err.(llmx.Error); ok {
		return llmxErr.Retryable()
	}

	// Default: don't retry unknown errors
	return false
}
