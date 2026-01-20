package middleware

import (
	"context"
	"time"

	"github.com/llmx-ai/llmx"
	"github.com/llmx-ai/llmx/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Telemetry creates a telemetry middleware
func Telemetry(tel *observability.Telemetry) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
			// Start tracing span
			ctx, span := tel.StartSpan(ctx, "llmx.chat",
				trace.WithAttributes(
					attribute.String("model", req.Model),
					attribute.Int("messages", len(req.Messages)),
					attribute.Int("tools", len(req.Tools)),
				),
			)
			defer span.End()

			start := time.Now()

			// Execute request
			resp, err := next(ctx, req)

			duration := time.Since(start)
			durationMs := float64(duration.Milliseconds())

			// Determine provider from model (simplified)
			provider := getProviderFromModel(req.Model)

			if err != nil {
				// Record error
				tel.RecordRequest(ctx, provider, req.Model, false)
				tel.RecordError(ctx, provider, req.Model, getErrorType(err))
				tel.RecordDuration(ctx, provider, req.Model, durationMs)

				span.SetStatus(codes.Error, err.Error())
				span.RecordError(err)

				return nil, err
			}

			// Record success metrics
			tel.RecordRequest(ctx, provider, req.Model, true)
			tel.RecordDuration(ctx, provider, req.Model, durationMs)

			// Record token usage
			if resp.Usage.PromptTokens > 0 {
				tel.RecordTokens(ctx, provider, req.Model, "prompt", int64(resp.Usage.PromptTokens))
			}
			if resp.Usage.CompletionTokens > 0 {
				tel.RecordTokens(ctx, provider, req.Model, "completion", int64(resp.Usage.CompletionTokens))
			}
			if resp.Usage.TotalTokens > 0 {
				tel.RecordTokens(ctx, provider, req.Model, "total", int64(resp.Usage.TotalTokens))
			}

			// Set span attributes
			span.SetAttributes(
				attribute.String("response.id", resp.ID),
				attribute.String("response.model", resp.Model),
				attribute.String("finish_reason", resp.FinishReason),
				attribute.Int("tokens.prompt", resp.Usage.PromptTokens),
				attribute.Int("tokens.completion", resp.Usage.CompletionTokens),
				attribute.Int("tokens.total", resp.Usage.TotalTokens),
				attribute.Float64("duration_ms", durationMs),
			)

			span.SetStatus(codes.Ok, "Request completed successfully")

			return resp, nil
		}
	}
}

// getProviderFromModel extracts provider name from model string
func getProviderFromModel(model string) string {
	// Simple heuristic
	if len(model) >= 3 {
		switch {
		case model[:3] == "gpt":
			return "openai"
		case len(model) >= 6 && model[:6] == "claude":
			return "anthropic"
		case len(model) >= 6 && model[:6] == "gemini":
			return "google"
		}
	}
	return "unknown"
}

// getErrorType returns error type for metrics
func getErrorType(err error) string {
	if llmxErr, ok := err.(llmx.Error); ok {
		return llmxErr.Code()
	}
	return "unknown"
}
