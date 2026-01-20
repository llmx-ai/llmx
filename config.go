package llmx

import (
	"net/http"
	"time"
)

// Config holds the client configuration
type Config struct {
	// Provider configuration
	Provider        string
	ProviderOptions map[string]interface{}

	// Default model settings
	DefaultModel string
	Temperature  *float64
	MaxTokens    *int
	TopP         *float64

	// HTTP client
	HTTPClient *http.Client

	// Timeout
	Timeout time.Duration

	// Debug mode
	Debug bool
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		Timeout:         60 * time.Second,
		Debug:           false,
		ProviderOptions: make(map[string]interface{}),
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate provider
	if c.Provider == "" {
		return NewInvalidRequestError("provider is required", nil)
	}

	// Validate provider options
	if c.ProviderOptions == nil {
		c.ProviderOptions = make(map[string]interface{})
	}

	// Check for API key (required for most providers)
	apiKey, _ := c.ProviderOptions["api_key"].(string)
	if apiKey == "" && c.Provider != "ollama" && c.Provider != "mock" { // ollama and mock don't need API key
		return NewInvalidRequestError("api_key is required", map[string]interface{}{
			"provider": c.Provider,
		})
	}

	// Validate temperature
	if c.Temperature != nil {
		if *c.Temperature < 0 || *c.Temperature > 2 {
			return NewInvalidRequestError("temperature must be between 0 and 2", map[string]interface{}{
				"temperature": *c.Temperature,
			})
		}
	}

	// Validate max_tokens
	if c.MaxTokens != nil {
		if *c.MaxTokens <= 0 {
			return NewInvalidRequestError("max_tokens must be positive", map[string]interface{}{
				"max_tokens": *c.MaxTokens,
			})
		}
		// Warn if max_tokens is too high (provider-specific limits vary)
		if *c.MaxTokens > 100000 {
			// This is just a warning, not an error
			// Different providers have different limits
		}
	}

	// Validate top_p
	if c.TopP != nil {
		if *c.TopP < 0 || *c.TopP > 1 {
			return NewInvalidRequestError("top_p must be between 0 and 1", map[string]interface{}{
				"top_p": *c.TopP,
			})
		}
	}

	// Validate HTTP client
	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{
			Timeout: 60 * time.Second,
		}
	}

	// Validate timeout
	if c.Timeout > 0 && c.HTTPClient.Timeout != c.Timeout {
		c.HTTPClient.Timeout = c.Timeout
	}
	if c.HTTPClient.Timeout <= 0 {
		c.HTTPClient.Timeout = 60 * time.Second
	}

	// Provider-specific validation
	switch c.Provider {
	case "azure":
		// Azure requires additional fields
		if _, ok := c.ProviderOptions["endpoint"].(string); !ok {
			return NewInvalidRequestError("azure provider requires 'endpoint' in provider options", nil)
		}
		if _, ok := c.ProviderOptions["deployment"].(string); !ok {
			return NewInvalidRequestError("azure provider requires 'deployment' in provider options", nil)
		}

	case "google":
		// Google might have specific requirements
		// Add validation as needed

	case "anthropic":
		// Anthropic might have specific requirements
		// Add validation as needed
	}

	return nil
}

// Clone creates a copy of the config
func (c *Config) Clone() *Config {
	clone := *c
	clone.ProviderOptions = make(map[string]interface{})
	for k, v := range c.ProviderOptions {
		clone.ProviderOptions[k] = v
	}
	return &clone
}
