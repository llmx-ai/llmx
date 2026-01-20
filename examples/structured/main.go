package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/llmx-ai/llmx"
	"github.com/llmx-ai/llmx/structured"

	// Import OpenAI provider
	_ "github.com/llmx-ai/llmx/provider/openai"
)

// Person represents a person with structured fields
type Person struct {
	Name    string `json:"name" description:"The person's full name"`
	Age     int    `json:"age" description:"The person's age in years"`
	Email   string `json:"email" description:"The person's email address"`
	Country string `json:"country" description:"The person's country of residence"`
}

// Recipe represents a cooking recipe
type Recipe struct {
	Name        string   `json:"name" description:"Name of the recipe"`
	Ingredients []string `json:"ingredients" description:"List of ingredients"`
	Steps       []string `json:"steps" description:"Cooking steps"`
	PrepTime    int      `json:"prep_time" description:"Preparation time in minutes"`
	Difficulty  string   `json:"difficulty" description:"Difficulty level: easy, medium, hard"`
}

// Product represents a product review analysis
type Product struct {
	Name       string  `json:"name" description:"Product name"`
	Sentiment  string  `json:"sentiment" description:"Sentiment: positive, negative, neutral"`
	Rating     float64 `json:"rating" description:"Rating from 0 to 5"`
	KeyPoints  []string `json:"key_points" description:"Key points from the review"`
	Recommend  bool    `json:"recommend" description:"Would recommend this product"`
}

func main() {
	fmt.Println("=== llmx Structured Output Example ===")
	fmt.Println()

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Create client
	client, err := llmx.NewClient(
		llmx.WithOpenAI(apiKey),
		llmx.WithDefaultModel("gpt-4-turbo-preview"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create structured output helper
	structuredOutput := structured.New(client)
	ctx := context.Background()

	// Example 1: Extract person information
	fmt.Println("👤 Example 1: Extract Person Information")
	fmt.Println("---")

	var person Person
	err = structuredOutput.GenerateInto(ctx,
		"Extract information about: John Smith is a 35-year-old software engineer living in Canada. His email is john.smith@example.com",
		&person,
	)

	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Name: %s\n", person.Name)
		fmt.Printf("Age: %d\n", person.Age)
		fmt.Printf("Email: %s\n", person.Email)
		fmt.Printf("Country: %s\n\n", person.Country)
	}

	// Example 2: Generate recipe
	fmt.Println("🍳 Example 2: Generate Recipe")
	fmt.Println("---")

	var recipe Recipe
	err = structuredOutput.GenerateInto(ctx,
		"Create a simple recipe for chocolate chip cookies",
		&recipe,
	)

	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Recipe: %s\n", recipe.Name)
		fmt.Printf("Difficulty: %s\n", recipe.Difficulty)
		fmt.Printf("Prep Time: %d minutes\n", recipe.PrepTime)
		fmt.Printf("Ingredients: %d items\n", len(recipe.Ingredients))
		fmt.Printf("Steps: %d steps\n\n", len(recipe.Steps))
	}

	// Example 3: Analyze product review
	fmt.Println("📦 Example 3: Analyze Product Review")
	fmt.Println("---")

	var product Product
	err = structuredOutput.GenerateInto(ctx,
		`Analyze this product review: "This laptop is amazing! Great performance, 
		long battery life, and beautiful display. A bit pricey but worth every penny. 
		Highly recommend for developers."`,
		&product,
	)

	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Product: %s\n", product.Name)
		fmt.Printf("Sentiment: %s\n", product.Sentiment)
		fmt.Printf("Rating: %.1f/5\n", product.Rating)
		fmt.Printf("Recommend: %v\n", product.Recommend)
		fmt.Printf("Key Points:\n")
		for _, point := range product.KeyPoints {
			fmt.Printf("  - %s\n", point)
		}
		fmt.Println()
	}

	// Example 4: Custom schema
	fmt.Println("📋 Example 4: Custom Schema")
	fmt.Println("---")

	customSchema := &llmx.Schema{
		Type: "object",
		Properties: map[string]*llmx.Schema{
			"task": {
				Type:        "string",
				Description: "A task description",
			},
			"priority": {
				Type:        "string",
				Description: "Priority level",
				Enum:        []interface{}{"low", "medium", "high", "urgent"},
			},
			"estimated_hours": {
				Type:        "number",
				Description: "Estimated hours to complete",
			},
			"tags": {
				Type: "array",
				Items: &llmx.Schema{
					Type: "string",
				},
				Description: "Task tags",
			},
		},
		Required: []string{"task", "priority", "estimated_hours"},
	}

	result, err := structuredOutput.Generate(ctx,
		"Create a task: Fix the login bug in the authentication system",
		customSchema,
	)

	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Task: %v\n", result["task"])
		fmt.Printf("Priority: %v\n", result["priority"])
		fmt.Printf("Estimated Hours: %v\n", result["estimated_hours"])
		if tags, ok := result["tags"].([]interface{}); ok {
			fmt.Printf("Tags: %v\n\n", tags)
		}
	}

	fmt.Println("=== Structured Output Demo Complete ===")
	fmt.Println("\nStructured output ensures:")
	fmt.Println("- Type-safe JSON responses")
	fmt.Println("- Schema validation")
	fmt.Println("- Direct mapping to Go structs")
	fmt.Println("- Consistent and predictable outputs")
}
