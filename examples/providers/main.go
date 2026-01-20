package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/llmx-ai/llmx"
	
	// Import providers to trigger registration
	_ "github.com/llmx-ai/llmx/provider/anthropic"
	_ "github.com/llmx-ai/llmx/provider/azure"
	_ "github.com/llmx-ai/llmx/provider/google"
	_ "github.com/llmx-ai/llmx/provider/openai"
)

func main() {
	fmt.Println("=== llmx Multi-Provider Example ===")
	fmt.Println()

	// Common request
	request := &llmx.ChatRequest{
		Messages: []llmx.Message{
			{
				Role: llmx.RoleUser,
				Content: []llmx.ContentPart{
					llmx.TextPart{Text: "Explain what makes Go a great programming language in one sentence."},
				},
			},
		},
	}

	ctx := context.Background()

	// 1. OpenAI
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		fmt.Println("📘 OpenAI (GPT-3.5 Turbo)")
		fmt.Println("---")
		
		client, err := llmx.NewClient(
			llmx.WithOpenAI(apiKey),
			llmx.WithDefaultModel("gpt-3.5-turbo"),
		)
		if err != nil {
			log.Printf("OpenAI error: %v\n", err)
		} else {
			resp, err := client.Chat(ctx, request)
			if err != nil {
				log.Printf("OpenAI chat error: %v\n", err)
			} else {
				fmt.Printf("Response: %s\n", resp.Content)
				fmt.Printf("Tokens: %d\n\n", resp.Usage.TotalTokens)
			}
		}
	}

	// 2. Anthropic (Claude)
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		fmt.Println("🤖 Anthropic (Claude 3.5 Haiku)")
		fmt.Println("---")
		
		client, err := llmx.NewClient(
			llmx.WithAnthropic(apiKey),
			llmx.WithDefaultModel("claude-3-5-haiku-20241022"),
		)
		if err != nil {
			log.Printf("Anthropic error: %v\n", err)
		} else {
			resp, err := client.Chat(ctx, request)
			if err != nil {
				log.Printf("Anthropic chat error: %v\n", err)
			} else {
				fmt.Printf("Response: %s\n", resp.Content)
				fmt.Printf("Tokens: %d\n\n", resp.Usage.TotalTokens)
			}
		}
	}

	// 3. Google (Gemini)
	if projectID := os.Getenv("GOOGLE_PROJECT_ID"); projectID != "" {
		fmt.Println("🔷 Google (Gemini 1.5 Flash)")
		fmt.Println("---")
		
		client, err := llmx.NewClient(
			llmx.WithGoogle(projectID),
			llmx.WithDefaultModel("gemini-1.5-flash"),
		)
		if err != nil {
			log.Printf("Google error: %v\n", err)
		} else {
			resp, err := client.Chat(ctx, request)
			if err != nil {
				log.Printf("Google chat error: %v\n", err)
			} else {
				fmt.Printf("Response: %s\n", resp.Content)
				fmt.Printf("Tokens: %d\n\n", resp.Usage.TotalTokens)
			}
		}
	}

	// 4. Azure OpenAI
	if apiKey := os.Getenv("AZURE_OPENAI_KEY"); apiKey != "" {
		if endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT"); endpoint != "" {
			fmt.Println("☁️  Azure OpenAI")
			fmt.Println("---")
			
			client, err := llmx.NewClient(
				llmx.WithProvider("azure", map[string]interface{}{
					"api_key":  apiKey,
					"endpoint": endpoint,
				}),
				llmx.WithDefaultModel("gpt-35-turbo"),
			)
			if err != nil {
				log.Printf("Azure error: %v\n", err)
			} else {
				resp, err := client.Chat(ctx, request)
				if err != nil {
					log.Printf("Azure chat error: %v\n", err)
				} else {
					fmt.Printf("Response: %s\n", resp.Content)
					fmt.Printf("Tokens: %d\n\n", resp.Usage.TotalTokens)
				}
			}
		}
	}

	// 5. Local Model (Ollama)
	fmt.Println("🏠 Local Model (Ollama)")
	fmt.Println("---")
	
	localClient, err := llmx.NewClient(
		llmx.WithOpenAICompatible("http://localhost:11434/v1", ""),
		llmx.WithDefaultModel("llama3.2"),
	)
	if err != nil {
		log.Printf("Ollama error: %v\n", err)
	} else {
		resp, err := localClient.Chat(ctx, request)
		if err != nil {
			log.Printf("Ollama chat error: %v (Make sure Ollama is running)\n", err)
		} else {
			fmt.Printf("Response: %s\n", resp.Content)
			fmt.Printf("Tokens: %d\n\n", resp.Usage.TotalTokens)
		}
	}

	fmt.Println("=== Provider Switching Example ===")
	fmt.Println()
	fmt.Println("Notice how the SAME code works with ALL providers!")
	fmt.Println("Just change the WithXXX() option and the model name.")
}
