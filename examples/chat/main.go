package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/llmx-ai/llmx"
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

	// Simple chat
	fmt.Println("=== Simple Chat Example ===")
	fmt.Println()

	resp, err := client.Chat(context.Background(), &llmx.ChatRequest{
		Messages: []llmx.Message{
			{
				Role: llmx.RoleUser,
				Content: []llmx.ContentPart{
					llmx.TextPart{Text: "What is Go programming language? Answer in one sentence."},
				},
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("AI: %s\n", resp.Content)
	fmt.Printf("\nTokens used: %d\n", resp.Usage.TotalTokens)
	fmt.Printf("Model: %s\n", resp.Model)

	// Multi-turn conversation
	fmt.Println()
	fmt.Println("=== Multi-turn Conversation ===")
	fmt.Println()

	messages := []llmx.Message{
		{
			Role: llmx.RoleSystem,
			Content: []llmx.ContentPart{
				llmx.TextPart{Text: "You are a helpful assistant that explains programming concepts."},
			},
		},
		{
			Role: llmx.RoleUser,
			Content: []llmx.ContentPart{
				llmx.TextPart{Text: "What are goroutines?"},
			},
		},
	}

	resp, err = client.Chat(context.Background(), &llmx.ChatRequest{
		Messages: messages,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("User: What are goroutines?\n")
	fmt.Printf("AI: %s\n\n", resp.Content)

	// Continue conversation
	messages = append(messages, llmx.Message{
		Role: llmx.RoleAssistant,
		Content: []llmx.ContentPart{
			llmx.TextPart{Text: resp.Content},
		},
	})

	messages = append(messages, llmx.Message{
		Role: llmx.RoleUser,
		Content: []llmx.ContentPart{
			llmx.TextPart{Text: "Can you give me a simple example?"},
		},
	})

	resp, err = client.Chat(context.Background(), &llmx.ChatRequest{
		Messages: messages,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("User: Can you give me a simple example?\n")
	fmt.Printf("AI: %s\n", resp.Content)
}
