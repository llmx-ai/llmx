package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/llmx-ai/llmx"
	"github.com/llmx-ai/llmx/middleware"
	"github.com/llmx-ai/llmx/observability"

	// Import OpenAI provider
	_ "github.com/llmx-ai/llmx/provider/openai"
)

func main() {
	fmt.Println("=== llmx Telemetry and Observability Example ===")
	fmt.Println()

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	ctx := context.Background()

	// Example 1: Basic Telemetry
	fmt.Println("📊 Example 1: Basic Telemetry (Metrics & Tracing)")
	fmt.Println("---")

	// Create telemetry instance
	// Note: In production, you would configure TracerProvider and MeterProvider
	// For this example, we use nil (no-op providers)
	tel, err := observability.New(&observability.Config{
		ServiceName:    "llmx-demo",
		ServiceVersion: "1.0.0",
		// TracerProvider: // Configure with OTLP exporter
		// MeterProvider:  // Configure with OTLP exporter
	})
	if err != nil {
		log.Fatal(err)
	}

	client, err := llmx.NewClient(
		llmx.WithOpenAI(apiKey),
		llmx.WithDefaultModel("gpt-3.5-turbo"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Add telemetry middleware
	client.Use(middleware.Telemetry(tel))

	// Make a request
	resp, err := client.Chat(ctx, &llmx.ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []llmx.Message{
			{
				Role: llmx.RoleUser,
				Content: []llmx.ContentPart{
					llmx.TextPart{Text: "What is OpenTelemetry?"},
				},
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Response: %s\n", resp.Content[:100]+"...")
	fmt.Printf("Tokens used: %d\n", resp.Usage.TotalTokens)
	fmt.Println()

	// Example 2: Custom Tracing
	fmt.Println("🔍 Example 2: Custom Tracing")
	fmt.Println("---")

	// Create a custom span
	spanCtx, span := tel.StartSpan(ctx, "custom-operation")
	defer span.End()

	// Make multiple requests within the span
	for i := 0; i < 3; i++ {
		_, innerSpan := tel.StartSpan(spanCtx, fmt.Sprintf("request-%d", i+1))

		_, err := client.Chat(spanCtx, &llmx.ChatRequest{
			Model: "gpt-3.5-turbo",
			Messages: []llmx.Message{
				{
					Role: llmx.RoleUser,
					Content: []llmx.ContentPart{
						llmx.TextPart{Text: fmt.Sprintf("Count to %d", i+1)},
					},
				},
			},
		})

		if err != nil {
			fmt.Printf("Request %d: Error - %v\n", i+1, err)
		} else {
			fmt.Printf("Request %d: Success\n", i+1)
		}

		innerSpan.End()
	}

	fmt.Println()

	// Example 3: Metrics Recording
	fmt.Println("📈 Example 3: Manual Metrics Recording")
	fmt.Println("---")

	// Record custom metrics
	tel.RecordRequest(ctx, "openai", "gpt-3.5-turbo", true)
	tel.RecordDuration(ctx, "openai", "gpt-3.5-turbo", 250.5)
	tel.RecordTokens(ctx, "openai", "gpt-3.5-turbo", "total", 150)

	fmt.Println("Metrics recorded:")
	fmt.Println("- Request: openai/gpt-3.5-turbo (success)")
	fmt.Println("- Duration: 250.5ms")
	fmt.Println("- Tokens: 150")
	fmt.Println()

	// Example 4: Full Observability Stack
	fmt.Println("🎯 Example 4: Full Observability Stack")
	fmt.Println("---")

	client2, err := llmx.NewClient(
		llmx.WithOpenAI(apiKey),
		llmx.WithDefaultModel("gpt-3.5-turbo"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Add full observability stack
	client2.Use(
		middleware.Logging(nil),        // Structured logging
		middleware.Telemetry(tel),      // Metrics & tracing
		middleware.Retry(3, nil),       // Retry with tracking
	)

	fmt.Println("Stack: Logging -> Telemetry -> Retry")
	fmt.Println()

	resp, err = client2.Chat(ctx, &llmx.ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []llmx.Message{
			{
				Role: llmx.RoleUser,
				Content: []llmx.ContentPart{
					llmx.TextPart{Text: "Explain observability in one sentence."},
				},
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Response: %s\n", resp.Content)
	fmt.Println()

	fmt.Println("=== Telemetry Demo Complete ===")
	fmt.Println()
	fmt.Println("Observability features:")
	fmt.Println("✅ Distributed Tracing - Track requests across services")
	fmt.Println("✅ Metrics Collection - Monitor performance and usage")
	fmt.Println("✅ Structured Logging - Correlated logs with traces")
	fmt.Println()
	fmt.Println("Integration options:")
	fmt.Println("- Jaeger for distributed tracing")
	fmt.Println("- Prometheus for metrics")
	fmt.Println("- Grafana for visualization")
	fmt.Println("- OTLP for unified telemetry")
}
