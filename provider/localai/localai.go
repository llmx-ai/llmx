// Package localai provides a LocalAI provider implementation for llmx.
// LocalAI is a drop-in replacement REST API that's OpenAI-compatible and runs locally.
package localai

import (
	"context"
	"fmt"

	"github.com/llmx-ai/llmx/provider"
	"github.com/llmx-ai/llmx/provider/openai"
)

const (
	// DefaultBaseURL is the default LocalAI API endpoint
	DefaultBaseURL = "http://localhost:8080/v1"
)

// LocalAIProvider implements the Provider interface for LocalAI
type LocalAIProvider struct {
	*openai.OpenAIProvider // Embed OpenAI provider for API compatibility
}

func init() {
	provider.Register("localai", NewLocalAIProvider)
}

// NewLocalAIProvider creates a new LocalAI provider
func NewLocalAIProvider(opts map[string]interface{}) (provider.Provider, error) {
	baseURL, ok := opts["base_url"].(string)
	if !ok || baseURL == "" {
		baseURL = DefaultBaseURL
	}

	// LocalAI doesn't require API key, but we set a dummy one for compatibility
	apiKey := "not-needed"
	if key, ok := opts["api_key"].(string); ok && key != "" {
		apiKey = key
	}

	// Create OpenAI-compatible provider with LocalAI configuration
	openaiOpts := map[string]interface{}{
		"api_key":  apiKey,
		"base_url": baseURL,
	}

	openaiProvider, err := openai.NewOpenAIProvider(openaiOpts)
	if err != nil {
		return nil, fmt.Errorf("localai: failed to create provider: %w", err)
	}

	return &LocalAIProvider{
		OpenAIProvider: openaiProvider.(*openai.OpenAIProvider),
	}, nil
}

// Name returns the provider name
func (p *LocalAIProvider) Name() string {
	return "localai"
}

// SupportedModels returns the list of common models for LocalAI
// Note: Actual available models depend on LocalAI configuration
func (p *LocalAIProvider) SupportedModels() []provider.Model {
	return []provider.Model{
		{
			ID:              "ggml-gpt4all-j",
			Name:            "GPT4All-J",
			ContextWindow:   2048,
			MaxOutputTokens: 2048,
			InputCost:       0, // Free (local)
			OutputCost:      0,
			// SupportToolCalling: false,
		},
		{
			ID:              "llama-2-7b-chat",
			Name:            "Llama 2 7B Chat",
			ContextWindow:   4096,
			MaxOutputTokens: 4096,
			InputCost:       0,
			OutputCost:      0,
			// SupportToolCalling: false,
		},
		{
			ID:              "mistral-7b-instruct",
			Name:            "Mistral 7B Instruct",
			ContextWindow:   8192,
			MaxOutputTokens: 8192,
			InputCost:       0,
			OutputCost:      0,
			// SupportToolCalling: false,
		},
	}
}

// SupportedFeatures returns the features supported by LocalAI
func (p *LocalAIProvider) SupportedFeatures() provider.Features {
	return provider.Features{
		Streaming:   true,
		ToolCalling: false, // LocalAI function calling support varies by model
		Vision:      false,
		JSONMode:    true,
		// SystemPrompt not needed
	}
}

// Chat sends a chat request to LocalAI
func (p *LocalAIProvider) Chat(ctx context.Context, req interface{}) (interface{}, error) {
	return p.OpenAIProvider.Chat(ctx, req)
}

// StreamChat sends a streaming chat request to LocalAI
func (p *LocalAIProvider) StreamChat(ctx context.Context, req interface{}) (interface{}, error) {
	return p.OpenAIProvider.StreamChat(ctx, req)
}
