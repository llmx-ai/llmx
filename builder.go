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

// Region sets the region (for Azure, Bedrock)
func (b *ClientBuilder) Region(region string) *ClientBuilder {
	if b.err != nil {
		return b
	}
	b.config.ProviderOptions["region"] = region
	return b
}

// ====================
// Provider-specific convenience methods
// ====================

// OpenAI configures the builder for OpenAI
func (b *ClientBuilder) OpenAI(apiKey string) *ClientBuilder {
	if b.err != nil {
		return b
	}
	b.config.Provider = "openai"
	b.config.ProviderOptions["api_key"] = apiKey
	return b
}

// Anthropic configures the builder for Anthropic
func (b *ClientBuilder) Anthropic(apiKey string) *ClientBuilder {
	if b.err != nil {
		return b
	}
	b.config.Provider = "anthropic"
	b.config.ProviderOptions["api_key"] = apiKey
	return b
}

// Google configures the builder for Google
func (b *ClientBuilder) Google(apiKey string) *ClientBuilder {
	if b.err != nil {
		return b
	}
	b.config.Provider = "google"
	b.config.ProviderOptions["api_key"] = apiKey
	return b
}

// Groq configures the builder for Groq (ultra-fast inference)
func (b *ClientBuilder) Groq(apiKey string) *ClientBuilder {
	if b.err != nil {
		return b
	}
	b.config.Provider = "groq"
	b.config.ProviderOptions["api_key"] = apiKey
	return b
}

// DeepSeek configures the builder for DeepSeek (cost-effective)
func (b *ClientBuilder) DeepSeek(apiKey string) *ClientBuilder {
	if b.err != nil {
		return b
	}
	b.config.Provider = "deepseek"
	b.config.ProviderOptions["api_key"] = apiKey
	return b
}

// Ollama configures the builder for Ollama (local inference)
func (b *ClientBuilder) Ollama(baseURL string) *ClientBuilder {
	if b.err != nil {
		return b
	}
	b.config.Provider = "ollama"
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	b.config.ProviderOptions["base_url"] = baseURL
	return b
}

// Mistral configures the builder for Mistral AI
func (b *ClientBuilder) Mistral(apiKey string) *ClientBuilder {
	if b.err != nil {
		return b
	}
	b.config.Provider = "mistral"
	b.config.ProviderOptions["api_key"] = apiKey
	return b
}

// Cohere configures the builder for Cohere
func (b *ClientBuilder) Cohere(apiKey string) *ClientBuilder {
	if b.err != nil {
		return b
	}
	b.config.Provider = "cohere"
	b.config.ProviderOptions["api_key"] = apiKey
	return b
}

// Bedrock configures the builder for Amazon Bedrock
func (b *ClientBuilder) Bedrock(region, accessKeyID, secretAccessKey string) *ClientBuilder {
	if b.err != nil {
		return b
	}
	b.config.Provider = "bedrock"
	b.config.ProviderOptions["region"] = region
	b.config.ProviderOptions["access_key_id"] = accessKeyID
	b.config.ProviderOptions["secret_access_key"] = secretAccessKey
	return b
}

// Zhipu configures the builder for Zhipu AI (智谱 GLM)
func (b *ClientBuilder) Zhipu(apiKey string) *ClientBuilder {
	if b.err != nil {
		return b
	}
	b.config.Provider = "zhipu"
	b.config.ProviderOptions["api_key"] = apiKey
	return b
}

// Tongyi configures the builder for Alibaba Tongyi (阿里通义千问)
func (b *ClientBuilder) Tongyi(apiKey string) *ClientBuilder {
	if b.err != nil {
		return b
	}
	b.config.Provider = "tongyi"
	b.config.ProviderOptions["api_key"] = apiKey
	return b
}

// Wenxin configures the builder for Baidu Wenxin (百度文心一言)
func (b *ClientBuilder) Wenxin(apiKey, secretKey string) *ClientBuilder {
	if b.err != nil {
		return b
	}
	b.config.Provider = "wenxin"
	b.config.ProviderOptions["api_key"] = apiKey
	b.config.ProviderOptions["secret_key"] = secretKey
	return b
}

// Doubao configures the builder for ByteDance Doubao (字节豆包)
func (b *ClientBuilder) Doubao(apiKey, secretKey string) *ClientBuilder {
	if b.err != nil {
		return b
	}
	b.config.Provider = "doubao"
	b.config.ProviderOptions["api_key"] = apiKey
	b.config.ProviderOptions["secret_key"] = secretKey
	return b
}

// HuggingFace configures the builder for Hugging Face
func (b *ClientBuilder) HuggingFace(token string) *ClientBuilder {
	if b.err != nil {
		return b
	}
	b.config.Provider = "huggingface"
	b.config.ProviderOptions["api_key"] = token
	return b
}

// LocalAI configures the builder for LocalAI
func (b *ClientBuilder) LocalAI(baseURL string) *ClientBuilder {
	if b.err != nil {
		return b
	}
	b.config.Provider = "localai"
	if baseURL == "" {
		baseURL = "http://localhost:8080/v1"
	}
	b.config.ProviderOptions["base_url"] = baseURL
	return b
}

// LMStudio configures the builder for LM Studio
func (b *ClientBuilder) LMStudio(baseURL string) *ClientBuilder {
	if b.err != nil {
		return b
	}
	b.config.Provider = "lmstudio"
	if baseURL == "" {
		baseURL = "http://localhost:1234/v1"
	}
	b.config.ProviderOptions["base_url"] = baseURL
	return b
}

// VLLM configures the builder for vLLM
func (b *ClientBuilder) VLLM(baseURL string) *ClientBuilder {
	if b.err != nil {
		return b
	}
	b.config.Provider = "vllm"
	if baseURL == "" {
		baseURL = "http://localhost:8000/v1"
	}
	b.config.ProviderOptions["base_url"] = baseURL
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

	// API key is optional for local providers (Ollama, LocalAI, LM Studio, vLLM)
	localProviders := map[string]bool{
		"ollama":   true,
		"localai":  true,
		"lmstudio": true,
		"vllm":     true,
	}

	apiKey, _ := b.config.ProviderOptions["api_key"].(string)
	baseURL, _ := b.config.ProviderOptions["base_url"].(string)

	if !localProviders[b.config.Provider] {
		// For non-local providers, API key is required
		if apiKey == "" {
			return nil, fmt.Errorf("api_key is required for provider '%s'", b.config.Provider)
		}
	} else {
		// For local providers, base_url is required
		if baseURL == "" {
			return nil, fmt.Errorf("base_url is required for provider '%s'", b.config.Provider)
		}
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
