package llmx

import (
	"fmt"
	"net/http"
	"time"
)

// ClientBuilder provides a fluent API for building Client instances
type ClientBuilder struct {
	config *Config
	err    error
}

// NewClientBuilder creates a new ClientBuilder with default configuration
func NewClientBuilder() *ClientBuilder {
	return &ClientBuilder{
		config: &Config{
			HTTPClient: &http.Client{
				Timeout: 60 * time.Second,
			},
			ProviderOptions: make(map[string]interface{}),
		},
	}
}

// Provider sets the provider configuration
func (b *ClientBuilder) Provider(provider, apiKey string) *ClientBuilder {
	if b.err != nil {
		return b
	}
	b.config.Provider = provider
	b.config.ProviderOptions["api_key"] = apiKey
	return b
}

// Model sets the default model
func (b *ClientBuilder) Model(model string) *ClientBuilder {
	if b.err != nil {
		return b
	}
	b.config.DefaultModel = model
	return b
}

// Temperature sets the default temperature
func (b *ClientBuilder) Temperature(temp float64) *ClientBuilder {
	if b.err != nil {
		return b
	}
	if temp < 0 || temp > 2 {
		b.err = fmt.Errorf("temperature must be between 0 and 2, got %.2f", temp)
		return b
	}
	b.config.Temperature = &temp
	return b
}

// MaxTokens sets the default max tokens
func (b *ClientBuilder) MaxTokens(tokens int) *ClientBuilder {
	if b.err != nil {
		return b
	}
	if tokens <= 0 {
		b.err = fmt.Errorf("max_tokens must be positive, got %d", tokens)
		return b
	}
	b.config.MaxTokens = &tokens
	return b
}

// TopP sets the default top_p value
func (b *ClientBuilder) TopP(topP float64) *ClientBuilder {
	if b.err != nil {
		return b
	}
	if topP < 0 || topP > 1 {
		b.err = fmt.Errorf("top_p must be between 0 and 1, got %.2f", topP)
		return b
	}
	b.config.TopP = &topP
	return b
}

// Timeout sets the HTTP client timeout
func (b *ClientBuilder) Timeout(timeout time.Duration) *ClientBuilder {
	if b.err != nil {
		return b
	}
	if timeout <= 0 {
		b.err = fmt.Errorf("timeout must be positive, got %v", timeout)
		return b
	}
	b.config.HTTPClient = &http.Client{
		Timeout: timeout,
	}
	return b
}

// HTTPClient sets a custom HTTP client
func (b *ClientBuilder) HTTPClient(client *http.Client) *ClientBuilder {
	if b.err != nil {
		return b
	}
	if client == nil {
		b.err = fmt.Errorf("http_client cannot be nil")
		return b
	}
	b.config.HTTPClient = client
	return b
}

// BaseURL sets the base URL for API requests (for OpenAI-compatible providers)
func (b *ClientBuilder) BaseURL(url string) *ClientBuilder {
	if b.err != nil {
		return b
	}
	if url == "" {
		b.err = fmt.Errorf("base_url cannot be empty")
		return b
	}
	b.config.ProviderOptions["base_url"] = url
	return b
}

// Organization sets the organization ID (for OpenAI)
func (b *ClientBuilder) Organization(org string) *ClientBuilder {
	if b.err != nil {
		return b
	}
	b.config.ProviderOptions["organization"] = org
	return b
}

// Region sets the region (for Azure)
func (b *ClientBuilder) Region(region string) *ClientBuilder {
	if b.err != nil {
		return b
	}
	b.config.ProviderOptions["region"] = region
	return b
}

// Build validates the configuration and creates the Client
func (b *ClientBuilder) Build() (*Client, error) {
	// Check for any accumulated errors
	if b.err != nil {
		return nil, fmt.Errorf("builder error: %w", b.err)
	}

	// Validate required fields
	if b.config.Provider == "" {
		return nil, fmt.Errorf("provider is required")
	}
	apiKey, _ := b.config.ProviderOptions["api_key"].(string)
	if apiKey == "" {
		return nil, fmt.Errorf("api_key is required")
	}

	// Create client using the existing NewClient function with WithConfig option
	return NewClient(WithConfig(b.config))
}

// MustBuild is like Build but panics on error
func (b *ClientBuilder) MustBuild() *Client {
	client, err := b.Build()
	if err != nil {
		panic(err)
	}
	return client
}
