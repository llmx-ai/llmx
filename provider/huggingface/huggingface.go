// Package huggingface provides a Hugging Face provider implementation for llmx.
// Hugging Face offers Inference API for thousands of models.
package huggingface

import (
	"context"
	"fmt"

	"github.com/llmx-ai/llmx/provider"
	"github.com/llmx-ai/llmx/provider/openai"
)

const (
	// DefaultBaseURL for Hugging Face Inference API (OpenAI-compatible endpoint)
	DefaultBaseURL = "https://api-inference.huggingface.co/v1"
)

// HuggingFaceProvider implements the Provider interface for Hugging Face
type HuggingFaceProvider struct {
	*openai.OpenAIProvider
}

func init() {
	provider.Register("huggingface", NewHuggingFaceProvider)
}

// NewHuggingFaceProvider creates a new Hugging Face provider
func NewHuggingFaceProvider(opts map[string]interface{}) (provider.Provider, error) {
	apiKey, ok := opts["api_key"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("huggingface: api_key (HF token) is required")
	}

	baseURL := DefaultBaseURL
	if customURL, ok := opts["base_url"].(string); ok && customURL != "" {
		baseURL = customURL
	}

	// HF Inference API is OpenAI-compatible
	openaiOpts := map[string]interface{}{
		"api_key":  apiKey,
		"base_url": baseURL,
	}

	openaiProvider, err := openai.NewOpenAIProvider(openaiOpts)
	if err != nil {
		return nil, fmt.Errorf("huggingface: failed to create provider: %w", err)
	}

	return &HuggingFaceProvider{
		OpenAIProvider: openaiProvider.(*openai.OpenAIProvider),
	}, nil
}

// Name returns the provider name
func (p *HuggingFaceProvider) Name() string {
	return "huggingface"
}

// SupportedModels returns popular models on Hugging Face
func (p *HuggingFaceProvider) SupportedModels() []provider.Model {
	return []provider.Model{
		{
			ID:              "meta-llama/Meta-Llama-3.1-70B-Instruct",
			Name:            "Llama 3.1 70B Instruct",
			ContextWindow:   131072,
			MaxOutputTokens: 4096,
			InputCost:       0, // Depends on serverless/dedicated endpoint
			OutputCost:      0,
			// SupportToolCalling: false,
		},
		{
			ID:              "mistralai/Mistral-7B-Instruct-v0.3",
			Name:            "Mistral 7B Instruct",
			ContextWindow:   32768,
			MaxOutputTokens: 8192,
			InputCost:       0,
			OutputCost:      0,
			// SupportToolCalling: false,
		},
		{
			ID:              "Qwen/Qwen2.5-72B-Instruct",
			Name:            "Qwen 2.5 72B Instruct",
			ContextWindow:   131072,
			MaxOutputTokens: 32768,
			InputCost:       0,
			OutputCost:      0,
			// SupportToolCalling: false,
		},
	}
}

// SupportedFeatures returns the features supported by Hugging Face
func (p *HuggingFaceProvider) SupportedFeatures() provider.Features {
	return provider.Features{
		Streaming:   true,
		ToolCalling: false, // Depends on model
		Vision:      false, // Depends on model
		JSONMode:    false,
		// SystemPrompt not needed
	}
}

// Chat sends a chat request to Hugging Face
func (p *HuggingFaceProvider) Chat(ctx context.Context, req interface{}) (interface{}, error) {
	return p.OpenAIProvider.Chat(ctx, req)
}

// StreamChat sends a streaming chat request to Hugging Face
func (p *HuggingFaceProvider) StreamChat(ctx context.Context, req interface{}) (interface{}, error) {
	return p.OpenAIProvider.StreamChat(ctx, req)
}
