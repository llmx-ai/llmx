package llmx

import (
	"net/http"
	"time"
)

// Option is a function that modifies the config
type Option func(*Config)

// WithProvider sets the provider
func WithProvider(name string, opts map[string]interface{}) Option {
	return func(c *Config) {
		c.Provider = name
		c.ProviderOptions = opts
	}
}

// WithOpenAI configures the client for OpenAI
func WithOpenAI(apiKey string) Option {
	return func(c *Config) {
		c.Provider = "openai"
		c.ProviderOptions = map[string]interface{}{
			"api_key": apiKey,
		}
	}
}

// WithOpenAIBaseURL configures OpenAI with a custom base URL
func WithOpenAIBaseURL(apiKey string, baseURL string) Option {
	return func(c *Config) {
		c.Provider = "openai"
		c.ProviderOptions = map[string]interface{}{
			"api_key":  apiKey,
			"base_url": baseURL,
		}
	}
}

// WithAnthropic configures the client for Anthropic
func WithAnthropic(apiKey string) Option {
	return func(c *Config) {
		c.Provider = "anthropic"
		c.ProviderOptions = map[string]interface{}{
			"api_key": apiKey,
		}
	}
}

// WithGoogle configures the client for Google
func WithGoogle(apiKey string) Option {
	return func(c *Config) {
		c.Provider = "google"
		c.ProviderOptions = map[string]interface{}{
			"api_key": apiKey,
		}
	}
}

// WithOpenAICompatible configures for OpenAI-compatible APIs
func WithOpenAICompatible(baseURL, apiKey string) Option {
	return func(c *Config) {
		c.Provider = "compatible"
		c.ProviderOptions = map[string]interface{}{
			"base_url": baseURL,
			"api_key":  apiKey,
		}
	}
}

// WithDefaultModel sets the default model
func WithDefaultModel(model string) Option {
	return func(c *Config) {
		c.DefaultModel = model
	}
}

// WithTemperature sets the default temperature
func WithTemperature(temp float64) Option {
	return func(c *Config) {
		c.Temperature = &temp
	}
}

// WithMaxTokens sets the default max tokens
func WithMaxTokens(tokens int) Option {
	return func(c *Config) {
		c.MaxTokens = &tokens
	}
}

// WithTopP sets the default top_p
func WithTopP(p float64) Option {
	return func(c *Config) {
		c.TopP = &p
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) Option {
	return func(c *Config) {
		c.HTTPClient = client
	}
}

// WithTimeout sets the request timeout
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
		if c.HTTPClient != nil {
			c.HTTPClient.Timeout = timeout
		}
	}
}

// WithDebug enables debug mode
func WithDebug(debug bool) Option {
	return func(c *Config) {
		c.Debug = debug
	}
}
