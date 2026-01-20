package middleware

import (
	"context"
	"log"
	"time"

	"github.com/llmx-ai/llmx"
)

// Logger interface for custom loggers
type Logger interface {
	Log(level, message string, fields map[string]interface{})
}

// DefaultLogger uses standard log package
type DefaultLogger struct{}

func (l *DefaultLogger) Log(level, message string, fields map[string]interface{}) {
	log.Printf("[%s] %s %v", level, message, fields)
}

// Logging creates a logging middleware
func Logging(logger Logger) Middleware {
	if logger == nil {
		logger = &DefaultLogger{}
	}

	return func(next Handler) Handler {
		return func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
			start := time.Now()

			// Log request
			logger.Log("INFO", "Request started", map[string]interface{}{
				"model":        req.Model,
				"messages":     len(req.Messages),
				"tools":        len(req.Tools),
				"has_tools":    len(req.Tools) > 0,
			})

			// Execute request
			resp, err := next(ctx, req)

			duration := time.Since(start)

			// Log response
			if err != nil {
				logger.Log("ERROR", "Request failed", map[string]interface{}{
					"model":    req.Model,
					"duration": duration.String(),
					"error":    err.Error(),
				})
			} else {
				logger.Log("INFO", "Request completed", map[string]interface{}{
					"model":         req.Model,
					"duration":      duration.String(),
					"tokens":        resp.Usage.TotalTokens,
					"finish_reason": resp.FinishReason,
				})
			}

			return resp, err
		}
	}
}
