package llmx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"

	"github.com/llmx-ai/llmx/provider"
)

// Client is the main llmx client
type Client struct {
	config      *Config
	provider    provider.Provider
	middlewares []Middleware // Store middlewares with correct type
	handler     Handler      // Cached middleware chain handler
	tools       []Tool       // Store tools

	// Resource management
	closeOnce sync.Once
	closed    bool
	mu        sync.RWMutex
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

	client := &Client{
		config:   config,
		provider: prov,
		tools:    []Tool{},
	}

	// Set finalizer to ensure resources are cleaned up
	runtime.SetFinalizer(client, func(c *Client) {
		c.Close()
	})

	return client, nil
}

// Chat sends a chat request and returns the response
func (c *Client) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// Check if client is closed
	if err := c.checkClosed(); err != nil {
		return nil, err
	}

	// Validate request
	if err := c.validateRequest(req); err != nil {
		return nil, err
	}

	// Apply defaults
	c.applyDefaults(req)

	// Use middleware chain if available
	if c.handler != nil {
		return c.handler(ctx, req)
	}

	// Fallback to direct provider call
	respInterface, err := c.provider.Chat(ctx, req)
	if err != nil {
		return nil, err
	}

	// Type assertion with detailed error handling
	resp, ok := respInterface.(*ChatResponse)
	if !ok {
		return nil, fmt.Errorf("llmx: invalid response type %T from provider %s", respInterface, c.provider.Name())
	}

	return resp, nil
}

// StreamChat sends a streaming chat request
func (c *Client) StreamChat(ctx context.Context, req *ChatRequest) (*ChatStream, error) {
	// Check if client is closed
	if err := c.checkClosed(); err != nil {
		return nil, err
	}

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

	// Type assertion with detailed error handling
	stream, ok := streamInterface.(*ChatStream)
	if !ok {
		return nil, fmt.Errorf("llmx: invalid stream type %T from provider %s", streamInterface, c.provider.Name())
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
func (c *Client) Use(mws ...Middleware) *Client {
	c.middlewares = append(c.middlewares, mws...)
	c.rebuildHandler()
	return c
}

// rebuildHandler rebuilds the middleware chain handler
func (c *Client) rebuildHandler() {
	// Base handler that calls the provider
	baseHandler := func(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
		respInterface, err := c.provider.Chat(ctx, req)
		if err != nil {
			return nil, err
		}

		// Type assertion with detailed error handling
		resp, ok := respInterface.(*ChatResponse)
		if !ok {
			return nil, fmt.Errorf("llmx: invalid response type %T from provider %s", respInterface, c.provider.Name())
		}

		return resp, nil
	}

	// Apply all middlewares in reverse order (last added = outermost)
	handler := baseHandler
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		handler = c.middlewares[i](handler)
	}

	c.handler = handler
}

// GenerateObject generates a structured object from a prompt
// This is a convenience method for structured output as specified in design docs
func (c *Client) GenerateObject(ctx context.Context, prompt string, output interface{}) error {
	// Check if client is closed
	if err := c.checkClosed(); err != nil {
		return err
	}

	// Create request with JSON mode
	req := &ChatRequest{
		Messages: []Message{
			{
				Role: RoleSystem,
				Content: []ContentPart{
					TextPart{Text: "You must respond with valid JSON that matches the provided schema. Do not include any text outside the JSON object."},
				},
			},
			{
				Role: RoleUser,
				Content: []ContentPart{
					TextPart{Text: prompt},
				},
			},
		},
		ProviderOptions: map[string]interface{}{
			"response_format": map[string]string{
				"type": "json_object",
			},
		},
	}

	// Apply defaults
	c.applyDefaults(req)

	// Execute request
	resp, err := c.Chat(ctx, req)
	if err != nil {
		return err
	}

	// Parse JSON into output
	return json.Unmarshal([]byte(resp.Content), output)
}

// Close closes the client and releases resources
// It's safe to call Close multiple times
func (c *Client) Close() error {
	var err error
	c.closeOnce.Do(func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		c.closed = true

		// Close HTTP client connections if using default transport
		if c.config.HTTPClient != nil {
			if transport, ok := c.config.HTTPClient.Transport.(*http.Transport); ok {
				transport.CloseIdleConnections()
			}
		}

		// Cancel finalizer since we're explicitly closing
		runtime.SetFinalizer(c, nil)
	})
	return err
}

// checkClosed returns an error if the client is closed
func (c *Client) checkClosed() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return fmt.Errorf("client is closed")
	}
	return nil
}

// IsClosed returns whether the client is closed
func (c *Client) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}
