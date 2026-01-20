// Package tongyi provides an Alibaba Tongyi (Qwen) provider implementation for llmx.
// Tongyi is Alibaba Cloud's large language model service.
//
// Note: This is a simplified implementation compatible with OpenAI API format.
// For full SDK, see: https://help.aliyun.com/zh/dashscope/
package tongyi

import (
	"context"
	"fmt"

	"github.com/llmx-ai/llmx/provider"
	"github.com/llmx-ai/llmx/provider/openai"
)

const (
	// DefaultBaseURL is the default Tongyi API endpoint
	DefaultBaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
)

// TongyiProvider implements the Provider interface for Tongyi
type TongyiProvider struct {
	*openai.OpenAIProvider
}

func init() {
	provider.Register("tongyi", NewTongyiProvider)
}

// NewTongyiProvider creates a new Tongyi provider
func NewTongyiProvider(opts map[string]interface{}) (provider.Provider, error) {
	apiKey, ok := opts["api_key"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("tongyi: api_key is required")
	}

	baseURL := DefaultBaseURL
	if customURL, ok := opts["base_url"].(string); ok && customURL != "" {
		baseURL = customURL
	}

	// Tongyi compatible mode uses OpenAI API format
	openaiOpts := map[string]interface{}{
		"api_key":  apiKey,
		"base_url": baseURL,
	}

	openaiProvider, err := openai.NewOpenAIProvider(openaiOpts)
	if err != nil {
		return nil, fmt.Errorf("tongyi: failed to create provider: %w", err)
	}

	return &TongyiProvider{
		OpenAIProvider: openaiProvider.(*openai.OpenAIProvider),
	}, nil
}

// Name returns the provider name
func (p *TongyiProvider) Name() string {
	return "tongyi"
}

// SupportedModels returns the list of models supported by Tongyi
func (p *TongyiProvider) SupportedModels() []provider.Model {
	return []provider.Model{
		{
			ID:              "qwen-max",
			Name:            "Qwen Max",
			ContextWindow:   32768,
			MaxOutputTokens: 8192,
			InputCost:       2.8, // CNY 20/1M tokens ≈ $2.8
			OutputCost:      8.4, // CNY 60/1M tokens
			// SupportToolCalling: true,
		},
		{
			ID:              "qwen-plus",
			Name:            "Qwen Plus",
			ContextWindow:   131072,
			MaxOutputTokens: 8192,
			InputCost:       0.56, // CNY 4/1M tokens
			OutputCost:      1.68, // CNY 12/1M tokens
			// SupportToolCalling: true,
		},
		{
			ID:              "qwen-turbo",
			Name:            "Qwen Turbo",
			ContextWindow:   131072,
			MaxOutputTokens: 8192,
			InputCost:       0.28, // CNY 2/1M tokens
			OutputCost:      0.84, // CNY 6/1M tokens
			// SupportToolCalling: true,
		},
		{
			ID:              "qwen-vl-plus",
			Name:            "Qwen VL Plus (Vision)",
			ContextWindow:   8192,
			MaxOutputTokens: 2048,
			InputCost:       1.12, // CNY 8/1M tokens
			OutputCost:      1.12,
			// SupportToolCalling: false,
		},
	}
}

// SupportedFeatures returns the features supported by Tongyi
func (p *TongyiProvider) SupportedFeatures() provider.Features {
	return provider.Features{
		Streaming:   true,
		ToolCalling: true,
		Vision:      true, // qwen-vl models support vision
		JSONMode:    true,
		// SystemPrompt not needed
	}
}

// Chat sends a chat request to Tongyi
func (p *TongyiProvider) Chat(ctx context.Context, req interface{}) (interface{}, error) {
	return p.OpenAIProvider.Chat(ctx, req)
}

// StreamChat sends a streaming chat request to Tongyi
func (p *TongyiProvider) StreamChat(ctx context.Context, req interface{}) (interface{}, error) {
	return p.OpenAIProvider.StreamChat(ctx, req)
}
