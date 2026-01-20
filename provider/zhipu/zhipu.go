// Package zhipu provides a Zhipu AI (GLM) provider implementation for llmx.
// Zhipu AI is a Chinese AI company offering the GLM series of models.
//
// Note: This is a simplified implementation. For full SDK integration,
// see: https://github.com/zhipuai/zhipuai-sdk-go-v4
package zhipu

import (
	"context"
	"fmt"

	"github.com/llmx-ai/llmx/provider"
	"github.com/llmx-ai/llmx/provider/openai"
)

const (
	// DefaultBaseURL is the default Zhipu API endpoint
	DefaultBaseURL = "https://open.bigmodel.cn/api/paas/v4"
)

// ZhipuProvider implements the Provider interface for Zhipu AI
type ZhipuProvider struct {
	*openai.OpenAIProvider
}

func init() {
	provider.Register("zhipu", NewZhipuProvider)
}

// NewZhipuProvider creates a new Zhipu provider
func NewZhipuProvider(opts map[string]interface{}) (provider.Provider, error) {
	apiKey, ok := opts["api_key"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("zhipu: api_key is required")
	}

	baseURL := DefaultBaseURL
	if customURL, ok := opts["base_url"].(string); ok && customURL != "" {
		baseURL = customURL
	}

	// Zhipu API is similar to OpenAI
	openaiOpts := map[string]interface{}{
		"api_key":  apiKey,
		"base_url": baseURL,
	}

	openaiProvider, err := openai.NewOpenAIProvider(openaiOpts)
	if err != nil {
		return nil, fmt.Errorf("zhipu: failed to create provider: %w", err)
	}

	return &ZhipuProvider{
		OpenAIProvider: openaiProvider.(*openai.OpenAIProvider),
	}, nil
}

// Name returns the provider name
func (p *ZhipuProvider) Name() string {
	return "zhipu"
}

// SupportedModels returns the list of models supported by Zhipu
func (p *ZhipuProvider) SupportedModels() []provider.Model {
	return []provider.Model{
		{
			ID:              "glm-4-plus",
			Name:            "GLM-4 Plus",
			ContextWindow:   131072,
			MaxOutputTokens: 4096,
			InputCost:       0.7, // CNY 50/1M tokens ≈ $0.7
			OutputCost:      0.7,
			// SupportToolCalling: true,
		},
		{
			ID:              "glm-4-0520",
			Name:            "GLM-4-0520",
			ContextWindow:   131072,
			MaxOutputTokens: 4096,
			InputCost:       1.4, // CNY 100/1M tokens
			OutputCost:      1.4,
			// SupportToolCalling: true,
		},
		{
			ID:              "glm-4",
			Name:            "GLM-4",
			ContextWindow:   131072,
			MaxOutputTokens: 4096,
			InputCost:       1.4,
			OutputCost:      1.4,
			// SupportToolCalling: true,
		},
		{
			ID:              "glm-3-turbo",
			Name:            "GLM-3 Turbo",
			ContextWindow:   131072,
			MaxOutputTokens: 4096,
			InputCost:       0.07, // CNY 5/1M tokens
			OutputCost:      0.07,
			// SupportToolCalling: false,
		},
		{
			ID:              "glm-4v",
			Name:            "GLM-4V (Vision)",
			ContextWindow:   8192,
			MaxOutputTokens: 1024,
			InputCost:       0.7,
			OutputCost:      0.7,
			// SupportToolCalling: false,
		},
	}
}

// SupportedFeatures returns the features supported by Zhipu
func (p *ZhipuProvider) SupportedFeatures() provider.Features {
	return provider.Features{
		Streaming:   true,
		ToolCalling: true,
		Vision:      true, // glm-4v supports vision
		JSONMode:    true,
		// SystemPrompt not needed
	}
}

// Chat sends a chat request to Zhipu
func (p *ZhipuProvider) Chat(ctx context.Context, req interface{}) (interface{}, error) {
	return p.OpenAIProvider.Chat(ctx, req)
}

// StreamChat sends a streaming chat request to Zhipu
func (p *ZhipuProvider) StreamChat(ctx context.Context, req interface{}) (interface{}, error) {
	return p.OpenAIProvider.StreamChat(ctx, req)
}
