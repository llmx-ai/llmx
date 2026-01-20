package llmx

import (
	"net/http"
	"time"
)

// Option is a function that modifies the config
type Option func(*Config)

// WithConfig uses an existing Config
func WithConfig(config *Config) Option {
	return func(c *Config) {
		*c = *config
	}
}

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

// WithGroq configures the client for Groq (ultra-fast inference)
func WithGroq(apiKey string) Option {
	return func(c *Config) {
		c.Provider = "groq"
		c.ProviderOptions = map[string]interface{}{
			"api_key": apiKey,
		}
	}
}

// WithDeepSeek configures the client for DeepSeek (cost-effective)
func WithDeepSeek(apiKey string) Option {
	return func(c *Config) {
		c.Provider = "deepseek"
		c.ProviderOptions = map[string]interface{}{
			"api_key": apiKey,
		}
	}
}

// WithOllama configures the client for Ollama (local inference)
func WithOllama(baseURL string) Option {
	return func(c *Config) {
		c.Provider = "ollama"
		c.ProviderOptions = map[string]interface{}{
			"base_url": baseURL,
		}
	}
}

// WithMistral configures the client for Mistral AI
func WithMistral(apiKey string) Option {
	return func(c *Config) {
		c.Provider = "mistral"
		c.ProviderOptions = map[string]interface{}{
			"api_key": apiKey,
		}
	}
}

// WithCohere configures the client for Cohere
func WithCohere(apiKey string) Option {
	return func(c *Config) {
		c.Provider = "cohere"
		c.ProviderOptions = map[string]interface{}{
			"api_key": apiKey,
		}
	}
}

// WithBedrock configures the client for Amazon Bedrock
func WithBedrock(region, accessKeyID, secretAccessKey string) Option {
	return func(c *Config) {
		c.Provider = "bedrock"
		c.ProviderOptions = map[string]interface{}{
			"region":            region,
			"access_key_id":     accessKeyID,
			"secret_access_key": secretAccessKey,
		}
	}
}

// WithZhipu configures the client for Zhipu AI (智谱 GLM)
func WithZhipu(apiKey string) Option {
	return func(c *Config) {
		c.Provider = "zhipu"
		c.ProviderOptions = map[string]interface{}{
			"api_key": apiKey,
		}
	}
}

// WithTongyi configures the client for Alibaba Tongyi (阿里通义千问)
func WithTongyi(apiKey string) Option {
	return func(c *Config) {
		c.Provider = "tongyi"
		c.ProviderOptions = map[string]interface{}{
			"api_key": apiKey,
		}
	}
}

// WithWenxin configures the client for Baidu Wenxin (百度文心一言)
func WithWenxin(apiKey, secretKey string) Option {
	return func(c *Config) {
		c.Provider = "wenxin"
		c.ProviderOptions = map[string]interface{}{
			"api_key":    apiKey,
			"secret_key": secretKey,
		}
	}
}

// WithDoubao configures the client for ByteDance Doubao (字节豆包)
func WithDoubao(apiKey, secretKey string) Option {
	return func(c *Config) {
		c.Provider = "doubao"
		c.ProviderOptions = map[string]interface{}{
			"api_key":    apiKey,
			"secret_key": secretKey,
		}
	}
}

// WithHuggingFace configures the client for Hugging Face
func WithHuggingFace(token string) Option {
	return func(c *Config) {
		c.Provider = "huggingface"
		c.ProviderOptions = map[string]interface{}{
			"api_key": token,
		}
	}
}

// WithLocalAI configures the client for LocalAI
func WithLocalAI(baseURL string) Option {
	return func(c *Config) {
		c.Provider = "localai"
		c.ProviderOptions = map[string]interface{}{
			"base_url": baseURL,
		}
	}
}

// WithLMStudio configures the client for LM Studio
func WithLMStudio(baseURL string) Option {
	return func(c *Config) {
		c.Provider = "lmstudio"
		c.ProviderOptions = map[string]interface{}{
			"base_url": baseURL,
		}
	}
}

// WithVLLM configures the client for vLLM
func WithVLLM(baseURL string) Option {
	return func(c *Config) {
		c.Provider = "vllm"
		c.ProviderOptions = map[string]interface{}{
			"base_url": baseURL,
		}
	}
}
