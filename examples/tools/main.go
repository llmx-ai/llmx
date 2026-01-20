package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/llmx-ai/llmx"
	"github.com/llmx-ai/llmx/tools"
	"github.com/llmx-ai/llmx/tools/builtin"
	
	// Import OpenAI provider
	_ "github.com/llmx-ai/llmx/provider/openai"
)

func main() {
	fmt.Println("=== llmx Tool Calling Example ===")
	fmt.Println()

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

	// Create tool registry
	registry := tools.NewRegistry()

	// Register builtin tools
	if err := registry.Register(builtin.CalculatorTool()); err != nil {
		log.Fatal(err)
	}
	if err := registry.Register(builtin.DateTimeTool()); err != nil {
		log.Fatal(err)
	}

	// Create tool executor
	executor := tools.NewExecutor(registry).WithMaxDepth(5)

	ctx := context.Background()

	// Example 1: Calculator
	fmt.Println("📊 Example 1: Calculator")
	fmt.Println("Question: What is 25 * 4 + 10?")
	fmt.Println("---")

	resp1, err := executor.ExecuteLoop(ctx, client, &llmx.ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []llmx.Message{
			{
				Role: llmx.RoleUser,
				Content: []llmx.ContentPart{
					llmx.TextPart{Text: "What is 25 * 4 + 10? Use the calculator tool to compute this."},
				},
			},
		},
		Tools: registry.List(),
	})

	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("AI Response: %s\n", resp1.Content)
		fmt.Printf("Tokens: %d\n\n", resp1.Usage.TotalTokens)
	}

	// Example 2: DateTime
	fmt.Println("🕐 Example 2: DateTime")
	fmt.Println("Question: What time is it in Tokyo?")
	fmt.Println("---")

	resp2, err := executor.ExecuteLoop(ctx, client, &llmx.ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []llmx.Message{
			{
				Role: llmx.RoleUser,
				Content: []llmx.ContentPart{
					llmx.TextPart{Text: "What time is it in Tokyo right now? Use the datetime tool."},
				},
			},
		},
		Tools: registry.List(),
	})

	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("AI Response: %s\n", resp2.Content)
		fmt.Printf("Tokens: %d\n\n", resp2.Usage.TotalTokens)
	}

	// Example 3: Multiple tools
	fmt.Println("🔧 Example 3: Multiple Tools")
	fmt.Println("Question: Complex query requiring multiple tools")
	fmt.Println("---")

	resp3, err := executor.ExecuteLoop(ctx, client, &llmx.ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []llmx.Message{
			{
				Role: llmx.RoleUser,
				Content: []llmx.ContentPart{
					llmx.TextPart{Text: "Calculate 100 / 5, then tell me what time it is in New York."},
				},
			},
		},
		Tools: registry.List(),
	})

	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("AI Response: %s\n", resp3.Content)
		fmt.Printf("Tokens: %d\n\n", resp3.Usage.TotalTokens)
	}

	fmt.Println("=== Tool Calling Demo Complete ===")
	fmt.Println("\nNote: The AI automatically calls tools and processes results!")
	fmt.Println("You can add custom tools by implementing the Tool interface.")
}
