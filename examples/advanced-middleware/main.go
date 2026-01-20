package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/llmx-ai/llmx"
	"github.com/llmx-ai/llmx/middleware"

	// Import OpenAI provider
	_ "github.com/llmx-ai/llmx/provider/openai"
)

func main() {
	fmt.Println("=== llmx Advanced Middleware Example ===")
	fmt.Println()

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	ctx := context.Background()

	// Example 1: Rate Limiting
	fmt.Println("🚦 Example 1: Rate Limiting")
	fmt.Println("---")

	client1, err := llmx.NewClient(
		llmx.WithOpenAI(apiKey),
		llmx.WithDefaultModel("gpt-3.5-turbo"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create rate limiter: 2 requests per second, burst of 5
	rateLimiter := middleware.NewTokenBucketLimiter(2, 5)
	client1.Use(middleware.RateLimit(rateLimiter, false))

	// Try multiple rapid requests
	for i := 0; i < 3; i++ {
		start := time.Now()
		_, err := client1.Chat(ctx, &llmx.ChatRequest{
			Model: "gpt-3.5-turbo",
			Messages: []llmx.Message{
				{
					Role: llmx.RoleUser,
					Content: []llmx.ContentPart{
						llmx.TextPart{Text: fmt.Sprintf("Say hello #%d", i+1)},
					},
				},
			},
		})

		if err != nil {
			fmt.Printf("Request %d: Rate limited (%v)\n", i+1, time.Since(start))
		} else {
			fmt.Printf("Request %d: Success (%v)\n", i+1, time.Since(start))
		}
	}
	fmt.Println()

	// Example 2: Circuit Breaker
	fmt.Println("⚡ Example 2: Circuit Breaker")
	fmt.Println("---")

	client2, err := llmx.NewClient(
		llmx.WithOpenAI(apiKey),
		llmx.WithDefaultModel("gpt-3.5-turbo"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create circuit breaker: open after 3 failures, timeout 10 seconds
	circuitBreaker := middleware.NewCircuitBreaker(3, 10*time.Second)
	client2.Use(middleware.CircuitBreakerMiddleware(circuitBreaker))

	fmt.Printf("Circuit Breaker State: %s\n", circuitBreaker.State().String())
	fmt.Println()

	// Example 3: Timeout
	fmt.Println("⏱️  Example 3: Timeout")
	fmt.Println("---")

	client3, err := llmx.NewClient(
		llmx.WithOpenAI(apiKey),
		llmx.WithDefaultModel("gpt-3.5-turbo"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Add timeout middleware: 30 seconds
	client3.Use(middleware.Timeout(30 * time.Second))

	resp, err := client3.Chat(ctx, &llmx.ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []llmx.Message{
			{
				Role: llmx.RoleUser,
				Content: []llmx.ContentPart{
					llmx.TextPart{Text: "Say hello!"},
				},
			},
		},
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Response: %s\n", resp.Content)
	}
	fmt.Println()

	// Example 4: Adaptive Timeout
	fmt.Println("🎯 Example 4: Adaptive Timeout")
	fmt.Println("---")

	client4, err := llmx.NewClient(
		llmx.WithOpenAI(apiKey),
		llmx.WithDefaultModel("gpt-3.5-turbo"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create adaptive timeout
	adaptiveTimeout := middleware.NewAdaptiveTimeout().
		WithBaseTimeout(10 * time.Second).
		WithPerMessage(2 * time.Second).
		WithPerTool(5 * time.Second).
		WithMaxTimeout(60 * time.Second)

	client4.Use(adaptiveTimeout.Middleware())

	// Request with multiple messages
	messages := []llmx.Message{
		{
			Role: llmx.RoleSystem,
			Content: []llmx.ContentPart{
				llmx.TextPart{Text: "You are a helpful assistant."},
			},
		},
		{
			Role: llmx.RoleUser,
			Content: []llmx.ContentPart{
				llmx.TextPart{Text: "Tell me about Go."},
			},
		},
	}

	req := &llmx.ChatRequest{
		Model:    "gpt-3.5-turbo",
		Messages: messages,
	}

	timeout := adaptiveTimeout.CalculateTimeout(req)
	fmt.Printf("Calculated timeout: %v (base: 10s + 2 messages * 2s)\n", timeout)

	resp, err = client4.Chat(ctx, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Response: %s\n", resp.Content[:50])
	}
	fmt.Println()

	// Example 5: Combined Middleware Stack
	fmt.Println("🔗 Example 5: Combined Middleware Stack")
	fmt.Println("---")

	client5, err := llmx.NewClient(
		llmx.WithOpenAI(apiKey),
		llmx.WithDefaultModel("gpt-3.5-turbo"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Combine multiple middleware
	client5.Use(
		middleware.Logging(nil),                                      // Logging
		middleware.Timeout(30*time.Second),                           // Timeout
		middleware.RateLimit(middleware.NewTokenBucketLimiter(2, 5), true), // Rate limiting with wait
		middleware.Retry(3, middleware.NewExponentialBackoff()),      // Retry
		middleware.CacheMiddleware(nil, 5*time.Minute),               // Cache
	)

	fmt.Println("Middleware stack: Logging -> Timeout -> RateLimit -> Retry -> Cache")

	resp, err = client5.Chat(ctx, &llmx.ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []llmx.Message{
			{
				Role: llmx.RoleUser,
				Content: []llmx.ContentPart{
					llmx.TextPart{Text: "What is 2+2?"},
				},
			},
		},
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Response: %s\n", resp.Content)
	}
	fmt.Println()

	fmt.Println("=== Advanced Middleware Demo Complete ===")
	fmt.Println()
	fmt.Println("Features demonstrated:")
	fmt.Println("✅ Rate Limiting - Control request rates")
	fmt.Println("✅ Circuit Breaker - Fail fast and recover")
	fmt.Println("✅ Timeout - Prevent hanging requests")
	fmt.Println("✅ Adaptive Timeout - Dynamic timeout based on complexity")
	fmt.Println("✅ Combined Stack - Multiple middleware working together")
}
