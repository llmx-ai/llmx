// Package vllm provides a vLLM provider implementation for llmx.
// vLLM is a high-performance inference engine with OpenAI-compatible API.
package vllm

import (
	"context"
	"fmt"

	"github.com/llmx-ai/llmx/provider"
	"github.com/llmx-ai/llmx/provider/openai"
)

const (
	// DefaultBaseURL is the default vLLM API endpoint
	DefaultBaseURL = "http://localhost:8000/v1"
)

// VLLMProvider implements the Provider interface for vLLM
type VLLMProvider struct {
	*openai.OpenAIProvider // Embed OpenAI provider for API compatibility
}

func init() {
	provider.Register("vllm", NewVLLMProvider)
}

// NewVLLMProvider creates a new vLLM provider
func NewVLLMProvider(opts map[string]interface{}) (provider.Provider, error) {
	baseURL, ok := opts["base_url"].(string)
	if !ok || baseURL == "" {
		baseURL = DefaultBaseURL
	}

	// vLLM doesn't require API key by default
	apiKey := "EMPTY"
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
		return nil, fmt.Errorf("vllm: failed to create provider: %w", err)
	}

	return &VLLMProvider{
		OpenAIProvider: openaiProvider.(*openai.OpenAIProvider),
	}, nil
}

// Name returns the provider name
func (p *VLLMProvider) Name() string {
	return "vllm"
}

// SupportedModels returns common models - actual models depend on vLLM deployment
func (p *VLLMProvider) SupportedModels() []provider.Model {
	return []provider.Model{
		{
			ID:              "deployed-model",
			Name:            "Currently Deployed Model",
			ContextWindow:   4096,
			MaxOutputTokens: 4096,
			InputCost:       0, // Pricing depends on deployment
			OutputCost:      0,
			// SupportToolCalling: true,
		},
	}
}

// SupportedFeatures returns the features supported by vLLM
func (p *VLLMProvider) SupportedFeatures() provider.Features {
	return provider.Features{
		Streaming:   true,
		ToolCalling: true,  // vLLM supports function calling
		Vision:      false, // Depends on deployed model
		JSONMode:    true,
		// SystemPrompt not needed
	}
}

// Chat sends a chat request to vLLM
func (p *VLLMProvider) Chat(ctx context.Context, req interface{}) (interface{}, error) {
	return p.OpenAIProvider.Chat(ctx, req)
}

// StreamChat sends a streaming chat request to vLLM
func (p *VLLMProvider) StreamChat(ctx context.Context, req interface{}) (interface{}, error) {
	return p.OpenAIProvider.StreamChat(ctx, req)
}
