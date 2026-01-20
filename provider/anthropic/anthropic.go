package anthropic

import (
	"context"
	"fmt"

	"github.com/llmx-ai/llmx"
	"github.com/llmx-ai/llmx/provider"
)

// AnthropicProvider implements the Provider interface for Anthropic Claude
type AnthropicProvider struct {
	apiKey string
}

func init() {
	provider.Register("anthropic", NewAnthropicProvider)
	provider.Register("claude", NewAnthropicProvider) // Alias
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(opts map[string]interface{}) (provider.Provider, error) {
	apiKey, ok := opts["api_key"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("api_key is required for Anthropic provider")
	}

	return &AnthropicProvider{
		apiKey: apiKey,
	}, nil
}

// Name returns the provider name
func (p *AnthropicProvider) Name() string {
	return "anthropic"
}

// Chat sends a chat request
func (p *AnthropicProvider) Chat(ctx context.Context, reqInterface interface{}) (interface{}, error) {
	_, ok := reqInterface.(*llmx.ChatRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type")
	}

	// TODO: Implement actual Anthropic API call
	// For now, return placeholder
	return nil, fmt.Errorf("Anthropic provider: not yet fully implemented")
}

// StreamChat sends a streaming chat request
func (p *AnthropicProvider) StreamChat(ctx context.Context, reqInterface interface{}) (interface{}, error) {
	_, ok := reqInterface.(*llmx.ChatRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type")
	}

	// TODO: Implement streaming
	return nil, fmt.Errorf("Anthropic streaming: not yet fully implemented")
}

// SupportedFeatures returns supported features
func (p *AnthropicProvider) SupportedFeatures() provider.Features {
	return provider.Features{
		Streaming:     true,
		ToolCalling:   true,
		Vision:        true,
		JSONMode:      false,
		ReasoningMode: true,  // Claude thinking
		CacheControl:  true,  // Prompt caching
		MultiModal:    true,
		Embedding:     false,
	}
}

// SupportedModels returns supported models
func (p *AnthropicProvider) SupportedModels() []provider.Model {
	return []provider.Model{
		{
			ID:              "claude-3-5-sonnet-20241022",
			Name:            "Claude 3.5 Sonnet",
			ContextWindow:   200000,
			MaxOutputTokens: 8192,
			InputCost:       3.0,
			OutputCost:      15.0,
			Capabilities:    []string{"chat", "vision", "tools", "thinking"},
		},
		{
			ID:              "claude-3-5-haiku-20241022",
			Name:            "Claude 3.5 Haiku",
			ContextWindow:   200000,
			MaxOutputTokens: 8192,
			InputCost:       0.8,
			OutputCost:      4.0,
			Capabilities:    []string{"chat", "vision", "tools"},
		},
		{
			ID:              "claude-3-opus-20240229",
			Name:            "Claude 3 Opus",
			ContextWindow:   200000,
			MaxOutputTokens: 4096,
			InputCost:       15.0,
			OutputCost:      75.0,
			Capabilities:    []string{"chat", "vision", "tools"},
		},
	}
}
