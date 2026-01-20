package llmx

import (
	"context"
	"fmt"

	"github.com/llmx-ai/llmx/provider"
)

// Client is the main llmx client
type Client struct {
	config      *Config
	provider    provider.Provider
	middlewares []interface{} // Store middlewares
	tools       []Tool        // Store tools
}

// NewClient creates a new llmx client
func NewClient(opts ...Option) (*Client, error) {
	// Start with default config
	config := DefaultConfig()

	// Apply options
	for _, opt := range opts {
		opt(config)
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Create provider
	prov, err := provider.New(config.Provider, config.ProviderOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	return &Client{
		config:   config,
		provider: prov,
	}, nil
}

// Chat sends a chat request and returns the response
func (c *Client) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// Validate request
	if err := c.validateRequest(req); err != nil {
		return nil, err
	}

	// Apply defaults
	c.applyDefaults(req)

	// Call provider
	respInterface, err := c.provider.Chat(ctx, req)
	if err != nil {
		return nil, err
	}

	resp, ok := respInterface.(*ChatResponse)
	if !ok {
		return nil, fmt.Errorf("invalid response type from provider")
	}

	return resp, nil
}

// StreamChat sends a streaming chat request
func (c *Client) StreamChat(ctx context.Context, req *ChatRequest) (*ChatStream, error) {
	// Validate request
	if err := c.validateRequest(req); err != nil {
		return nil, err
	}

	// Apply defaults
	c.applyDefaults(req)

	// Call provider
	streamInterface, err := c.provider.StreamChat(ctx, req)
	if err != nil {
		return nil, err
	}

	stream, ok := streamInterface.(*ChatStream)
	if !ok {
		return nil, fmt.Errorf("invalid stream type from provider")
	}

	return stream, nil
}

// validateRequest validates a chat request
func (c *Client) validateRequest(req *ChatRequest) error {
	if req == nil {
		return NewInvalidRequestError("request is nil", nil)
	}

	if err := ValidateMessages(req.Messages); err != nil {
		return err
	}

	return nil
}

// applyDefaults applies default values from config to request
func (c *Client) applyDefaults(req *ChatRequest) {
	if req.Model == "" && c.config.DefaultModel != "" {
		req.Model = c.config.DefaultModel
	}

	if req.Temperature == nil && c.config.Temperature != nil {
		temp := *c.config.Temperature
		req.Temperature = &temp
	}

	if req.MaxTokens == nil && c.config.MaxTokens != nil {
		tokens := *c.config.MaxTokens
		req.MaxTokens = &tokens
	}

	if req.TopP == nil && c.config.TopP != nil {
		topP := *c.config.TopP
		req.TopP = &topP
	}
}

// Provider returns the underlying provider
func (c *Client) Provider() provider.Provider {
	return c.provider
}

// Config returns the client configuration
func (c *Client) Config() *Config {
	return c.config.Clone()
}

// WithTools adds tools to the client
func (c *Client) WithTools(tools ...Tool) *Client {
	c.tools = append(c.tools, tools...)
	return c
}

// Tools returns the client's tools
func (c *Client) Tools() []Tool {
	return c.tools
}

// Use adds middleware to the client
func (c *Client) Use(middlewares ...interface{}) *Client {
	c.middlewares = append(c.middlewares, middlewares...)
	return c
}
