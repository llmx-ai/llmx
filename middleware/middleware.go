package middleware

import (
	"github.com/llmx-ai/llmx"
)

// Re-export types from llmx package for convenience
type (
	Handler    = llmx.Handler
	Middleware = llmx.Middleware
)

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
