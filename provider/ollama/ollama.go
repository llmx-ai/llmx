// Package ollama provides an Ollama provider implementation for llmx.
// Ollama is a tool for running LLMs locally with ease.
package ollama

import (
	"context"
	"fmt"

	"github.com/llmx-ai/llmx/provider"
	"github.com/llmx-ai/llmx/provider/openai"
)

const (
	// DefaultBaseURL is the default Ollama API endpoint
	DefaultBaseURL = "http://localhost:11434/v1"
)

// OllamaProvider implements the Provider interface for Ollama
type OllamaProvider struct {
	*openai.OpenAIProvider // Embed OpenAI provider for API compatibility
	baseURL                string
}

func init() {
	provider.Register("ollama", NewOllamaProvider)
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(opts map[string]interface{}) (provider.Provider, error) {
	baseURL, ok := opts["base_url"].(string)
	if !ok || baseURL == "" {
		baseURL = DefaultBaseURL
	}

	// Ollama doesn't require API key
	apiKey := "ollama"
	if key, ok := opts["api_key"].(string); ok && key != "" {
		apiKey = key
	}

	// Create OpenAI-compatible provider
	openaiOpts := map[string]interface{}{
		"api_key":  apiKey,
		"base_url": baseURL,
	}

	openaiProvider, err := openai.NewOpenAIProvider(openaiOpts)
	if err != nil {
		return nil, fmt.Errorf("ollama: failed to create provider: %w", err)
	}

	return &OllamaProvider{
		OpenAIProvider: openaiProvider.(*openai.OpenAIProvider),
		baseURL:        baseURL,
	}, nil
}

// Name returns the provider name
func (p *OllamaProvider) Name() string {
	return "ollama"
}

// SupportedModels returns popular Ollama models
// Note: Actual available models depend on what's pulled locally
func (p *OllamaProvider) SupportedModels() []provider.Model {
	return []provider.Model{
		{
			ID:              "llama3.3:70b",
			Name:            "Llama 3.3 70B",
			ContextWindow:   131072,
			MaxOutputTokens: 32768,
			InputCost:       0, // Free (local)
			OutputCost:      0,
			// SupportToolCalling: true,
		},
		{
			ID:              "llama3.1:8b",
			Name:            "Llama 3.1 8B",
			ContextWindow:   131072,
			MaxOutputTokens: 8192,
			InputCost:       0,
			OutputCost:      0,
			// SupportToolCalling: true,
		},
		{
			ID:              "qwen2.5:72b",
			Name:            "Qwen 2.5 72B",
			ContextWindow:   131072,
			MaxOutputTokens: 32768,
			InputCost:       0,
			OutputCost:      0,
			// SupportToolCalling: true,
		},
		{
			ID:              "deepseek-r1:70b",
			Name:            "DeepSeek R1 70B",
			ContextWindow:   65536,
			MaxOutputTokens: 16384,
			InputCost:       0,
			OutputCost:      0,
			// SupportToolCalling: true,
		},
		{
			ID:              "mistral:7b",
			Name:            "Mistral 7B",
			ContextWindow:   32768,
			MaxOutputTokens: 8192,
			InputCost:       0,
			OutputCost:      0,
			// SupportToolCalling: false,
		},
		{
			ID:              "gemma2:9b",
			Name:            "Gemma 2 9B",
			ContextWindow:   8192,
			MaxOutputTokens: 8192,
			InputCost:       0,
			OutputCost:      0,
			// SupportToolCalling: false,
		},
		{
			ID:              "phi3:mini",
			Name:            "Phi 3 Mini",
			ContextWindow:   4096,
			MaxOutputTokens: 4096,
			InputCost:       0,
			OutputCost:      0,
			// SupportToolCalling: false,
		},
		{
			ID:              "codellama:7b",
			Name:            "Code Llama 7B",
			ContextWindow:   16384,
			MaxOutputTokens: 4096,
			InputCost:       0,
			OutputCost:      0,
			// SupportToolCalling: false,
		},
	}
}

// SupportedFeatures returns the features supported by Ollama
func (p *OllamaProvider) SupportedFeatures() provider.Features {
	return provider.Features{
		Streaming:   true,
		ToolCalling: true,  // Newer Ollama models support function calling
		Vision:      false, // Some models support vision, but not standardized
		JSONMode:    true,
		// SystemPrompt not needed
	}
}

// Chat sends a chat request to Ollama
func (p *OllamaProvider) Chat(ctx context.Context, req interface{}) (interface{}, error) {
	return p.OpenAIProvider.Chat(ctx, req)
}

// StreamChat sends a streaming chat request to Ollama
func (p *OllamaProvider) StreamChat(ctx context.Context, req interface{}) (interface{}, error) {
	return p.OpenAIProvider.StreamChat(ctx, req)
}
