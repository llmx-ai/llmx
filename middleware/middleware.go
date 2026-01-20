package middleware

import (
	"context"

	"github.com/llmx-ai/llmx"
)

// Middleware is a function that wraps a Handler
type Middleware func(next Handler) Handler

// Handler is a function that handles a chat request
type Handler func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error)

// Chain creates a middleware chain
func Chain(middlewares ...Middleware) Middleware {
	return func(next Handler) Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// Apply applies middleware to a handler
func Apply(handler Handler, middlewares ...Middleware) Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
