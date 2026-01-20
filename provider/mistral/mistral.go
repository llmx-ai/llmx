// Package mistral provides a Mistral AI provider implementation for llmx.
// Mistral AI is a European AI company offering powerful open-source and proprietary models.
//
// Note: This is a simplified implementation that uses the OpenAI-compatible API.
// For full Mistral SDK integration, see: https://github.com/gage-technologies/mistral-go
package mistral

import (
	"context"
	"fmt"

	"github.com/llmx-ai/llmx/provider"
	"github.com/llmx-ai/llmx/provider/openai"
)

const (
	// DefaultBaseURL is the default Mistral API endpoint
	DefaultBaseURL = "https://api.mistral.ai/v1"
)

// MistralProvider implements the Provider interface for Mistral AI
type MistralProvider struct {
	*openai.OpenAIProvider // Embed OpenAI provider for API compatibility
}

func init() {
	provider.Register("mistral", NewMistralProvider)
}

// NewMistralProvider creates a new Mistral provider
func NewMistralProvider(opts map[string]interface{}) (provider.Provider, error) {
	apiKey, ok := opts["api_key"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("mistral: api_key is required")
	}

	baseURL := DefaultBaseURL
	if customURL, ok := opts["base_url"].(string); ok && customURL != "" {
		baseURL = customURL
	}

	// Mistral API is OpenAI-compatible
	openaiOpts := map[string]interface{}{
		"api_key":  apiKey,
		"base_url": baseURL,
	}

	openaiProvider, err := openai.NewOpenAIProvider(openaiOpts)
	if err != nil {
		return nil, fmt.Errorf("mistral: failed to create provider: %w", err)
	}

	return &MistralProvider{
		OpenAIProvider: openaiProvider.(*openai.OpenAIProvider),
	}, nil
}

// Name returns the provider name
func (p *MistralProvider) Name() string {
	return "mistral"
}

// SupportedModels returns the list of models supported by Mistral
func (p *MistralProvider) SupportedModels() []provider.Model {
	return []provider.Model{
		{
			ID:              "mistral-large-latest",
			Name:            "Mistral Large",
			ContextWindow:   128000,
			MaxOutputTokens: 32000,
			InputCost:       3.0,
			OutputCost:      9.0,
			// SupportToolCalling: true,
		},
		{
			ID:              "mistral-medium-latest",
			Name:            "Mistral Medium",
			ContextWindow:   32000,
			MaxOutputTokens: 8000,
			InputCost:       2.7,
			OutputCost:      8.1,
			// SupportToolCalling: true,
		},
		{
			ID:              "mistral-small-latest",
			Name:            "Mistral Small",
			ContextWindow:   32000,
			MaxOutputTokens: 8000,
			InputCost:       1.0,
			OutputCost:      3.0,
			// SupportToolCalling: true,
		},
		{
			ID:              "open-mixtral-8x7b",
			Name:            "Mixtral 8x7B",
			ContextWindow:   32000,
			MaxOutputTokens: 8000,
			InputCost:       0.7,
			OutputCost:      0.7,
			// SupportToolCalling: true,
		},
		{
			ID:              "open-mixtral-8x22b",
			Name:            "Mixtral 8x22B",
			ContextWindow:   64000,
			MaxOutputTokens: 16000,
			InputCost:       2.0,
			OutputCost:      6.0,
			// SupportToolCalling: true,
		},
		{
			ID:              "open-mistral-7b",
			Name:            "Mistral 7B",
			ContextWindow:   32000,
			MaxOutputTokens: 8000,
			InputCost:       0.25,
			OutputCost:      0.25,
			// SupportToolCalling: false,
		},
	}
}

// SupportedFeatures returns the features supported by Mistral
func (p *MistralProvider) SupportedFeatures() provider.Features {
	return provider.Features{
		Streaming:   true,
		ToolCalling: true,
		Vision:      false,
		JSONMode:    true,
		// SystemPrompt not needed
	}
}

// Chat sends a chat request to Mistral
func (p *MistralProvider) Chat(ctx context.Context, req interface{}) (interface{}, error) {
	return p.OpenAIProvider.Chat(ctx, req)
}

// StreamChat sends a streaming chat request to Mistral
func (p *MistralProvider) StreamChat(ctx context.Context, req interface{}) (interface{}, error) {
	return p.OpenAIProvider.StreamChat(ctx, req)
}
