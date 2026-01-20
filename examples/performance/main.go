package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/llmx-ai/llmx"
	"github.com/llmx-ai/llmx/middleware"

	// Import OpenAI provider
	_ "github.com/llmx-ai/llmx/provider/openai"
)

func main() {
	fmt.Println("=== llmx Performance Optimization Example ===")
	fmt.Println()

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	ctx := context.Background()

	// Example 1: Sequential vs Concurrent Requests
	fmt.Println("⚡ Example 1: Sequential vs Concurrent")
	fmt.Println("---")

	client, err := llmx.NewClient(
		llmx.WithOpenAI(apiKey),
		llmx.WithDefaultModel("gpt-3.5-turbo"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Sequential requests
	fmt.Println("Sequential requests:")
	start := time.Now()
	for i := 0; i < 3; i++ {
		_, err := client.Chat(ctx, &llmx.ChatRequest{
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
			log.Printf("Request %d failed: %v", i+1, err)
		}
	}
	seqDuration := time.Since(start)
	fmt.Printf("Time: %v\n", seqDuration)
	fmt.Println()

	// Concurrent requests
	fmt.Println("Concurrent requests:")
	start = time.Now()

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_, err := client.Chat(ctx, &llmx.ChatRequest{
				Model: "gpt-3.5-turbo",
				Messages: []llmx.Message{
					{
						Role: llmx.RoleUser,
						Content: []llmx.ContentPart{
							llmx.TextPart{Text: fmt.Sprintf("Say hello #%d", n)},
						},
					},
				},
			})
			if err != nil {
				log.Printf("Request %d failed: %v", n, err)
			}
		}(i + 1)
	}
	wg.Wait()

	concDuration := time.Since(start)
	fmt.Printf("Time: %v\n", concDuration)
	fmt.Printf("Speedup: %.2fx\n", float64(seqDuration)/float64(concDuration))
	fmt.Println()

	// Example 2: Caching Performance
	fmt.Println("💾 Example 2: Caching Performance")
	fmt.Println("---")

	client2, err := llmx.NewClient(
		llmx.WithOpenAI(apiKey),
		llmx.WithDefaultModel("gpt-3.5-turbo"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Add cache middleware
	client2.Use(middleware.CacheMiddleware(nil, 5*time.Minute))

	req := &llmx.ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []llmx.Message{
			{
				Role: llmx.RoleUser,
				Content: []llmx.ContentPart{
					llmx.TextPart{Text: "What is 2+2?"},
				},
			},
		},
	}

	// First request (cache miss)
	start = time.Now()
	_, err = client2.Chat(ctx, req)
	if err != nil {
		log.Fatal(err)
	}
	missDuration := time.Since(start)
	fmt.Printf("Cache miss: %v\n", missDuration)

	// Second request (cache hit)
	start = time.Now()
	_, err = client2.Chat(ctx, req)
	if err != nil {
		log.Fatal(err)
	}
	hitDuration := time.Since(start)
	fmt.Printf("Cache hit: %v\n", hitDuration)
	fmt.Printf("Speedup: %.2fx\n", float64(missDuration)/float64(hitDuration))
	fmt.Println()

	// Example 3: Memory Usage
	fmt.Println("🧠 Example 3: Memory Usage")
	fmt.Println("---")

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Printf("Allocated: %v MB\n", bToMb(m.Alloc))
	fmt.Printf("Total Allocated: %v MB\n", bToMb(m.TotalAlloc))
	fmt.Printf("System: %v MB\n", bToMb(m.Sys))
	fmt.Printf("Goroutines: %d\n", runtime.NumGoroutine())
	fmt.Println()

	// Example 4: Batching Performance
	fmt.Println("📦 Example 4: Batching Performance")
	fmt.Println("---")

	client3, err := llmx.NewClient(
		llmx.WithOpenAI(apiKey),
		llmx.WithDefaultModel("gpt-3.5-turbo"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Add rate limiting
	rateLimiter := middleware.NewTokenBucketLimiter(5, 10)
	client3.Use(middleware.RateLimit(rateLimiter, true))

	fmt.Println("Sending 10 requests with rate limiting (5 RPS)...")
	start = time.Now()

	var wg2 sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg2.Add(1)
		go func(n int) {
			defer wg2.Done()
			_, err := client3.Chat(ctx, &llmx.ChatRequest{
				Model: "gpt-3.5-turbo",
				Messages: []llmx.Message{
					{
						Role: llmx.RoleUser,
						Content: []llmx.ContentPart{
							llmx.TextPart{Text: fmt.Sprintf("Number %d", n)},
						},
					},
				},
			})
			if err != nil {
				log.Printf("Request %d failed: %v", n, err)
			}
		}(i + 1)
	}
	wg2.Wait()

	batchDuration := time.Since(start)
	fmt.Printf("Time: %v\n", batchDuration)
	fmt.Printf("Throughput: %.2f req/s\n", 10/batchDuration.Seconds())
	fmt.Println()

	// Example 5: Optimized Stack
	fmt.Println("🚀 Example 5: Optimized Middleware Stack")
	fmt.Println("---")

	client4, err := llmx.NewClient(
		llmx.WithOpenAI(apiKey),
		llmx.WithDefaultModel("gpt-3.5-turbo"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Optimized stack for production
	client4.Use(
		middleware.Timeout(30*time.Second),                           // Prevent hanging
		middleware.RateLimit(middleware.NewTokenBucketLimiter(5, 10), true), // Rate control
		middleware.CacheMiddleware(nil, 5*time.Minute),               // Fast responses
		middleware.Retry(3, middleware.NewExponentialBackoff()),      // Reliability
	)

	fmt.Println("Stack: Timeout -> RateLimit -> Cache -> Retry")

	start = time.Now()
	resp, err := client4.Chat(ctx, &llmx.ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []llmx.Message{
			{
				Role: llmx.RoleUser,
				Content: []llmx.ContentPart{
					llmx.TextPart{Text: "What is performance optimization?"},
				},
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	duration := time.Since(start)
	fmt.Printf("Response time: %v\n", duration)
	fmt.Printf("Response: %s\n", resp.Content[:80]+"...")
	fmt.Println()

	fmt.Println("=== Performance Demo Complete ===")
	fmt.Println()
	fmt.Println("Performance tips:")
	fmt.Println("✅ Use concurrent requests for parallel operations")
	fmt.Println("✅ Enable caching for repeated queries")
	fmt.Println("✅ Apply rate limiting to avoid API throttling")
	fmt.Println("✅ Set timeouts to prevent resource leaks")
	fmt.Println("✅ Monitor memory usage in production")
	fmt.Println("✅ Combine middleware for optimal performance")
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
