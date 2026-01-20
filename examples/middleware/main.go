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
	fmt.Println("=== llmx Middleware Example ===")
	fmt.Println()

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Example 1: Logging Middleware
	fmt.Println("📝 Example 1: Logging Middleware")
	fmt.Println("---")

	client1, err := llmx.NewClient(
		llmx.WithOpenAI(apiKey),
		llmx.WithDefaultModel("gpt-3.5-turbo"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Use logging middleware
	client1.Use(middleware.Logging(nil))

	ctx := context.Background()
	req := &llmx.ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []llmx.Message{
			{
				Role: llmx.RoleUser,
				Content: []llmx.ContentPart{
					llmx.TextPart{Text: "Say hello!"},
				},
			},
		},
	}

	resp1, err := client1.Chat(ctx, req)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Response: %s\n\n", resp1.Content)
	}

	// Example 2: Retry Middleware
	fmt.Println("🔄 Example 2: Retry Middleware")
	fmt.Println("---")

	client2, err := llmx.NewClient(
		llmx.WithOpenAI(apiKey),
		llmx.WithDefaultModel("gpt-3.5-turbo"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Use retry middleware with 3 retries
	client2.Use(middleware.Retry(3, middleware.NewExponentialBackoff()))

	resp2, err := client2.Chat(ctx, req)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Response: %s\n\n", resp2.Content)
	}

	// Example 3: Cache Middleware
	fmt.Println("💾 Example 3: Cache Middleware")
	fmt.Println("---")

	client3, err := llmx.NewClient(
		llmx.WithOpenAI(apiKey),
		llmx.WithDefaultModel("gpt-3.5-turbo"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Use cache middleware with 5 minute TTL
	client3.Use(middleware.CacheMiddleware(nil, 5*time.Minute))

	// First call (miss)
	start := time.Now()
	resp3a, err := client3.Chat(ctx, &llmx.ChatRequest{
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
	duration1 := time.Since(start)

	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("First call (cache miss): %s (took %v)\n", resp3a.Content, duration1)
	}

	// Second call (hit)
	start = time.Now()
	resp3b, err := client3.Chat(ctx, &llmx.ChatRequest{
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
	duration2 := time.Since(start)

	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Second call (cache hit): %s (took %v)\n\n", resp3b.Content, duration2)
		fmt.Printf("Speedup: %.2fx faster\n\n", float64(duration1)/float64(duration2))
	}

	// Example 4: Middleware Chain
	fmt.Println("🔗 Example 4: Middleware Chain")
	fmt.Println("---")

	client4, err := llmx.NewClient(
		llmx.WithOpenAI(apiKey),
		llmx.WithDefaultModel("gpt-3.5-turbo"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Chain multiple middlewares: logging -> retry -> cache
	client4.Use(
		middleware.Logging(nil),
		middleware.Retry(3, middleware.NewExponentialBackoff()),
		middleware.CacheMiddleware(nil, 5*time.Minute),
	)

	resp4, err := client4.Chat(ctx, &llmx.ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []llmx.Message{
			{
				Role: llmx.RoleUser,
				Content: []llmx.ContentPart{
					llmx.TextPart{Text: "Tell me a fun fact!"},
				},
			},
		},
	})

	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Response: %s\n\n", resp4.Content)
	}

	fmt.Println("=== Middleware Demo Complete ===")
	fmt.Println("\nMiddlewares provide powerful cross-cutting concerns:")
	fmt.Println("- Logging: Track all requests and responses")
	fmt.Println("- Retry: Automatically retry failed requests")
	fmt.Println("- Cache: Speed up repeated requests")
	fmt.Println("- Custom: Build your own middleware!")
}
