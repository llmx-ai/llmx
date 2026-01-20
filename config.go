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
	if c.Provider == "" {
		return NewInvalidRequestError("provider is required", nil)
	}

	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{
			Timeout: 60 * time.Second,
		}
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
