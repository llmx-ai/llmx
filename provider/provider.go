package provider

import (
	"context"
)

// Provider is the interface that all AI providers must implement
type Provider interface {
	// Name returns the provider name
	Name() string

	// Chat sends a chat request and returns the response
	Chat(ctx context.Context, req interface{}) (interface{}, error)

	// StreamChat sends a streaming chat request
	StreamChat(ctx context.Context, req interface{}) (interface{}, error)

	// SupportedFeatures returns the features supported by this provider
	SupportedFeatures() Features

	// SupportedModels returns the models supported by this provider
	SupportedModels() []Model
}

// Features represents the features supported by a provider
type Features struct {
	Streaming     bool
	ToolCalling   bool
	Vision        bool
	JSONMode      bool
	ReasoningMode bool // Claude thinking
	CacheControl  bool // Prompt caching
	MultiModal    bool
	Embedding     bool
}

// Model represents a model supported by a provider
type Model struct {
	ID              string
	Name            string
	ContextWindow   int
	MaxOutputTokens int
	InputCost       float64 // per 1M tokens
	OutputCost      float64 // per 1M tokens
	Capabilities    []string
}

// ProviderFactory is a function that creates a provider
type ProviderFactory func(opts map[string]interface{}) (Provider, error)
