// Package lmstudio provides an LM Studio provider implementation for llmx.
// LM Studio is a desktop application for running LLMs locally with an OpenAI-compatible API.
package lmstudio

import (
	"context"
	"fmt"

	"github.com/llmx-ai/llmx/provider"
	"github.com/llmx-ai/llmx/provider/openai"
)

const (
	// DefaultBaseURL is the default LM Studio API endpoint
	DefaultBaseURL = "http://localhost:1234/v1"
)

// LMStudioProvider implements the Provider interface for LM Studio
type LMStudioProvider struct {
	*openai.OpenAIProvider // Embed OpenAI provider for API compatibility
}

func init() {
	provider.Register("lmstudio", NewLMStudioProvider)
}

// NewLMStudioProvider creates a new LM Studio provider
func NewLMStudioProvider(opts map[string]interface{}) (provider.Provider, error) {
	baseURL, ok := opts["base_url"].(string)
	if !ok || baseURL == "" {
		baseURL = DefaultBaseURL
	}

	// LM Studio doesn't require API key
	apiKey := "lm-studio"
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
		return nil, fmt.Errorf("lmstudio: failed to create provider: %w", err)
	}

	return &LMStudioProvider{
		OpenAIProvider: openaiProvider.(*openai.OpenAIProvider),
	}, nil
}

// Name returns the provider name
func (p *LMStudioProvider) Name() string {
	return "lmstudio"
}

// SupportedModels returns common models - actual models depend on what's loaded in LM Studio
func (p *LMStudioProvider) SupportedModels() []provider.Model {
	return []provider.Model{
		{
			ID:              "local-model",
			Name:            "Currently Loaded Model",
			ContextWindow:   4096,
			MaxOutputTokens: 4096,
			InputCost:       0, // Free (local)
			OutputCost:      0,
			// SupportToolCalling: false,
		},
	}
}

// SupportedFeatures returns the features supported by LM Studio
func (p *LMStudioProvider) SupportedFeatures() provider.Features {
	return provider.Features{
		Streaming:   true,
		ToolCalling: false,
		Vision:      false,
		JSONMode:    false,
		// SystemPrompt not needed
	}
}

// Chat sends a chat request to LM Studio
func (p *LMStudioProvider) Chat(ctx context.Context, req interface{}) (interface{}, error) {
	return p.OpenAIProvider.Chat(ctx, req)
}

// StreamChat sends a streaming chat request to LM Studio
func (p *LMStudioProvider) StreamChat(ctx context.Context, req interface{}) (interface{}, error) {
	return p.OpenAIProvider.StreamChat(ctx, req)
}
