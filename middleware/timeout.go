package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/llmx-ai/llmx"
)

// Timeout creates a timeout middleware
func Timeout(timeout time.Duration) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
			// Create context with timeout
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			// Channel to receive result
			type result struct {
				resp *llmx.ChatResponse
				err  error
			}
			resultChan := make(chan result, 1)

			// Execute request in goroutine
			go func() {
				resp, err := next(ctx, req)
				resultChan <- result{resp: resp, err: err}
			}()

			// Wait for result or timeout
			select {
			case res := <-resultChan:
				return res.resp, res.err
			case <-ctx.Done():
				return nil, llmx.NewInternalError(
					fmt.Sprintf("request timeout after %v", timeout),
					ctx.Err(),
				)
			}
		}
	}
}

// TimeoutPerModel creates a per-model timeout middleware
func TimeoutPerModel(timeouts map[string]time.Duration, defaultTimeout time.Duration) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
			// Get timeout for model
			timeout, ok := timeouts[req.Model]
			if !ok {
				timeout = defaultTimeout
			}

			// Create context with timeout
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			// Channel to receive result
			type result struct {
				resp *llmx.ChatResponse
				err  error
			}
			resultChan := make(chan result, 1)

			// Execute request in goroutine
			go func() {
				resp, err := next(ctx, req)
				resultChan <- result{resp: resp, err: err}
			}()

			// Wait for result or timeout
			select {
			case res := <-resultChan:
				return res.resp, res.err
			case <-ctx.Done():
				return nil, llmx.NewInternalError(
					fmt.Sprintf("request timeout for model %s after %v", req.Model, timeout),
					ctx.Err(),
				)
			}
		}
	}
}

// AdaptiveTimeout creates an adaptive timeout middleware
// It adjusts timeout based on request complexity
type AdaptiveTimeout struct {
	baseTimeout      time.Duration
	perMessage       time.Duration
	perTool          time.Duration
	maxTimeout       time.Duration
}

// NewAdaptiveTimeout creates a new adaptive timeout
func NewAdaptiveTimeout() *AdaptiveTimeout {
	return &AdaptiveTimeout{
		baseTimeout: 30 * time.Second,
		perMessage:  2 * time.Second,
		perTool:     5 * time.Second,
		maxTimeout:  5 * time.Minute,
	}
}

// WithBaseTimeout sets the base timeout
func (at *AdaptiveTimeout) WithBaseTimeout(d time.Duration) *AdaptiveTimeout {
	at.baseTimeout = d
	return at
}

// WithPerMessage sets the per-message timeout
func (at *AdaptiveTimeout) WithPerMessage(d time.Duration) *AdaptiveTimeout {
	at.perMessage = d
	return at
}

// WithPerTool sets the per-tool timeout
func (at *AdaptiveTimeout) WithPerTool(d time.Duration) *AdaptiveTimeout {
	at.perTool = d
	return at
}

// WithMaxTimeout sets the maximum timeout
func (at *AdaptiveTimeout) WithMaxTimeout(d time.Duration) *AdaptiveTimeout {
	at.maxTimeout = d
	return at
}

// CalculateTimeout calculates timeout based on request
func (at *AdaptiveTimeout) CalculateTimeout(req *llmx.ChatRequest) time.Duration {
	timeout := at.baseTimeout

	// Add time for messages
	timeout += time.Duration(len(req.Messages)) * at.perMessage

	// Add time for tools
	timeout += time.Duration(len(req.Tools)) * at.perTool

	// Cap at max timeout
	if timeout > at.maxTimeout {
		timeout = at.maxTimeout
	}

	return timeout
}

// Middleware creates an adaptive timeout middleware
func (at *AdaptiveTimeout) Middleware() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
			timeout := at.CalculateTimeout(req)

			// Create context with timeout
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			// Channel to receive result
			type result struct {
				resp *llmx.ChatResponse
				err  error
			}
			resultChan := make(chan result, 1)

			// Execute request in goroutine
			go func() {
				resp, err := next(ctx, req)
				resultChan <- result{resp: resp, err: err}
			}()

			// Wait for result or timeout
			select {
			case res := <-resultChan:
				return res.resp, res.err
			case <-ctx.Done():
				return nil, llmx.NewInternalError(
					fmt.Sprintf("adaptive timeout after %v (messages: %d, tools: %d)",
						timeout, len(req.Messages), len(req.Tools)),
					ctx.Err(),
				)
			}
		}
	}
}
