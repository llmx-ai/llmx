package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/llmx-ai/llmx"
	"github.com/llmx-ai/llmx/core"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Create client
	client, err := llmx.NewClient(
		llmx.WithOpenAI(apiKey),
		llmx.WithDefaultModel("gpt-3.5-turbo"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Streaming example
	fmt.Println("=== Streaming Chat Example ===")
	fmt.Println()
	fmt.Println("AI: ")

	stream, err := client.StreamChat(context.Background(), &llmx.ChatRequest{
		Messages: []llmx.Message{
			{
				Role: llmx.RoleUser,
				Content: []llmx.ContentPart{
					llmx.TextPart{Text: "Write a short poem about Go programming language."},
				},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	// Process streaming events
	for event := range stream.Events() {
		switch event.Type {
		case core.EventTypeStart:
			// Stream started
		case core.EventTypeTextDelta:
			// Print text as it arrives
			if text, ok := event.Data.(string); ok {
				fmt.Print(text)
			}
		case core.EventTypeFinish:
			// Stream finished
			fmt.Println("\n\n[Stream completed]")
		case core.EventTypeError:
			// Error occurred
			if err, ok := event.Data.(error); ok {
				log.Printf("Error: %v\n", err)
			}
		}
	}

	// Check for errors
	select {
	case err := <-stream.Errors():
		log.Printf("Stream error: %v\n", err)
	default:
		// No error
	}

	// Example using Accumulate
	fmt.Println()
	fmt.Println("=== Using Accumulate ===")
	fmt.Println()

	stream2, err := client.StreamChat(context.Background(), &llmx.ChatRequest{
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
		log.Fatal(err)
	}

	// Accumulate waits for the full response
	resp, err := stream2.Accumulate()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Complete response: %s\n", resp.Content)
}
