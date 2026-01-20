// Package groq provides a Groq provider implementation for llmx.
// Groq offers extremely fast inference speeds (500+ tokens/s) and is fully compatible with OpenAI's API.
package groq

import (
	"context"
	"fmt"

	"github.com/llmx-ai/llmx/provider"
	"github.com/llmx-ai/llmx/provider/openai"
	openaisdk "github.com/sashabaranov/go-openai"
)

const (
	// DefaultBaseURL is the default Groq API endpoint
	DefaultBaseURL = "https://api.groq.com/openai/v1"
)

// GroqProvider implements the Provider interface for Groq
type GroqProvider struct {
	*openai.OpenAIProvider // Embed OpenAI provider for API compatibility
}

func init() {
	provider.Register("groq", NewGroqProvider)
}

// NewGroqProvider creates a new Groq provider
func NewGroqProvider(opts map[string]interface{}) (provider.Provider, error) {
	apiKey, ok := opts["api_key"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("groq: api_key is required")
	}

	// Use custom base URL if provided, otherwise use default Groq endpoint
	baseURL := DefaultBaseURL
	if customURL, ok := opts["base_url"].(string); ok && customURL != "" {
		baseURL = customURL
	}

	// Create OpenAI-compatible client with Groq endpoint
	config := openaisdk.DefaultConfig(apiKey)
	config.BaseURL = baseURL

	// Create OpenAI provider with Groq configuration
	openaiOpts := map[string]interface{}{
		"api_key":  apiKey,
		"base_url": baseURL,
	}

	openaiProvider, err := openai.NewOpenAIProvider(openaiOpts)
	if err != nil {
		return nil, fmt.Errorf("groq: failed to create provider: %w", err)
	}

	return &GroqProvider{
		OpenAIProvider: openaiProvider.(*openai.OpenAIProvider),
	}, nil
}

// Name returns the provider name
func (p *GroqProvider) Name() string {
	return "groq"
}

// SupportedModels returns the list of models supported by Groq
func (p *GroqProvider) SupportedModels() []provider.Model {
	return []provider.Model{
		{
			ID:              "llama-3.3-70b-versatile",
			Name:            "Llama 3.3 70B Versatile",
			ContextWindow:   131072,
			MaxOutputTokens: 32768,
			InputCost:       0.59, // per 1M tokens
			OutputCost:      0.79, // per 1M tokens
			// SupportToolCalling: true,
		},
		{
			ID:              "llama-3.1-70b-versatile",
			Name:            "Llama 3.1 70B Versatile",
			ContextWindow:   131072,
			MaxOutputTokens: 32768,
			InputCost:       0.59,
			OutputCost:      0.79,
			// SupportToolCalling: true,
		},
		{
			ID:              "llama-3.1-8b-instant",
			Name:            "Llama 3.1 8B Instant",
			ContextWindow:   131072,
			MaxOutputTokens: 8192,
			InputCost:       0.05,
			OutputCost:      0.08,
			// SupportToolCalling: true,
		},
		{
			ID:              "llama3-70b-8192",
			Name:            "Llama 3 70B",
			ContextWindow:   8192,
			MaxOutputTokens: 8192,
			InputCost:       0.59,
			OutputCost:      0.79,
			// SupportToolCalling: true,
		},
		{
			ID:              "llama3-8b-8192",
			Name:            "Llama 3 8B",
			ContextWindow:   8192,
			MaxOutputTokens: 8192,
			InputCost:       0.05,
			OutputCost:      0.08,
			// SupportToolCalling: true,
		},
		{
			ID:              "mixtral-8x7b-32768",
			Name:            "Mixtral 8x7B",
			ContextWindow:   32768,
			MaxOutputTokens: 32768,
			InputCost:       0.24,
			OutputCost:      0.24,
			// SupportToolCalling: true,
		},
		{
			ID:              "gemma2-9b-it",
			Name:            "Gemma 2 9B",
			ContextWindow:   8192,
			MaxOutputTokens: 8192,
			InputCost:       0.20,
			OutputCost:      0.20,
			// SupportToolCalling: false,
		},
		{
			ID:              "gemma-7b-it",
			Name:            "Gemma 7B",
			ContextWindow:   8192,
			MaxOutputTokens: 8192,
			InputCost:       0.07,
			OutputCost:      0.07,
			// SupportToolCalling: false,
		},
	}
}

// SupportedFeatures returns the features supported by Groq
func (p *GroqProvider) SupportedFeatures() provider.Features {
	return provider.Features{
		Streaming:   true,
		ToolCalling: true,
		Vision:      false, // Groq doesn't support vision yet
		JSONMode:    true,
		// SystemPrompt not needed
	}
}

// Chat sends a chat request to Groq
func (p *GroqProvider) Chat(ctx context.Context, req interface{}) (interface{}, error) {
	// Delegate to OpenAI provider (API compatible)
	return p.OpenAIProvider.Chat(ctx, req)
}

// StreamChat sends a streaming chat request to Groq
func (p *GroqProvider) StreamChat(ctx context.Context, req interface{}) (interface{}, error) {
	// Delegate to OpenAI provider (API compatible)
	return p.OpenAIProvider.StreamChat(ctx, req)
}
