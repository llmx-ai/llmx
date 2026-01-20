// Package deepseek provides a DeepSeek provider implementation for llmx.
// DeepSeek offers highly cost-effective AI models with competitive performance.
package deepseek

import (
	"context"
	"fmt"

	"github.com/llmx-ai/llmx/provider"
	"github.com/llmx-ai/llmx/provider/openai"
	openaisdk "github.com/sashabaranov/go-openai"
)

const (
	// DefaultBaseURL is the default DeepSeek API endpoint
	DefaultBaseURL = "https://api.deepseek.com/v1"
)

// DeepSeekProvider implements the Provider interface for DeepSeek
type DeepSeekProvider struct {
	*openai.OpenAIProvider // Embed OpenAI provider for API compatibility
}

func init() {
	provider.Register("deepseek", NewDeepSeekProvider)
}

// NewDeepSeekProvider creates a new DeepSeek provider
func NewDeepSeekProvider(opts map[string]interface{}) (provider.Provider, error) {
	apiKey, ok := opts["api_key"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("deepseek: api_key is required")
	}

	// Use custom base URL if provided, otherwise use default DeepSeek endpoint
	baseURL := DefaultBaseURL
	if customURL, ok := opts["base_url"].(string); ok && customURL != "" {
		baseURL = customURL
	}

	// Create OpenAI-compatible client with DeepSeek endpoint
	config := openaisdk.DefaultConfig(apiKey)
	config.BaseURL = baseURL

	// Create OpenAI provider with DeepSeek configuration
	openaiOpts := map[string]interface{}{
		"api_key":  apiKey,
		"base_url": baseURL,
	}

	openaiProvider, err := openai.NewOpenAIProvider(openaiOpts)
	if err != nil {
		return nil, fmt.Errorf("deepseek: failed to create provider: %w", err)
	}

	return &DeepSeekProvider{
		OpenAIProvider: openaiProvider.(*openai.OpenAIProvider),
	}, nil
}

// Name returns the provider name
func (p *DeepSeekProvider) Name() string {
	return "deepseek"
}

// SupportedModels returns the list of models supported by DeepSeek
func (p *DeepSeekProvider) SupportedModels() []provider.Model {
	return []provider.Model{
		{
			ID:              "deepseek-chat",
			Name:            "DeepSeek Chat",
			ContextWindow:   32768,
			MaxOutputTokens: 4096,
			InputCost:       0.14, // per 1M tokens (CNY 1.0, ~$0.14)
			OutputCost:      0.28, // per 1M tokens (CNY 2.0, ~$0.28)
			// SupportToolCalling: true,
		},
		{
			ID:              "deepseek-coder",
			Name:            "DeepSeek Coder",
			ContextWindow:   16384,
			MaxOutputTokens: 4096,
			InputCost:       0.14,
			OutputCost:      0.28,
			// SupportToolCalling: true,
		},
	}
}

// SupportedFeatures returns the features supported by DeepSeek
func (p *DeepSeekProvider) SupportedFeatures() provider.Features {
	return provider.Features{
		Streaming:   true,
		ToolCalling: true,
		Vision:      false, // DeepSeek doesn't support vision currently
		JSONMode:    true,
		// SystemPrompt not needed
	}
}

// Chat sends a chat request to DeepSeek
func (p *DeepSeekProvider) Chat(ctx context.Context, req interface{}) (interface{}, error) {
	// Delegate to OpenAI provider (API compatible)
	return p.OpenAIProvider.Chat(ctx, req)
}

// StreamChat sends a streaming chat request to DeepSeek
func (p *DeepSeekProvider) StreamChat(ctx context.Context, req interface{}) (interface{}, error) {
	// Delegate to OpenAI provider (API compatible)
	return p.OpenAIProvider.StreamChat(ctx, req)
}
