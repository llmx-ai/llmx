package llmx

import (
	"net/http"
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := &Config{
			Provider: "openai",
			ProviderOptions: map[string]interface{}{
				"api_key": "test-key",
			},
		}

		err := config.Validate()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("missing provider", func(t *testing.T) {
		config := &Config{}

		err := config.Validate()
		if err == nil {
			t.Error("expected error for missing provider")
		}
	})

	t.Run("missing api key", func(t *testing.T) {
		config := &Config{
			Provider:        "openai",
			ProviderOptions: map[string]interface{}{},
		}

		err := config.Validate()
		if err == nil {
			t.Error("expected error for missing api_key")
		}
	})

	t.Run("invalid temperature too high", func(t *testing.T) {
		temp := 3.0
		config := &Config{
			Provider: "openai",
			ProviderOptions: map[string]interface{}{
				"api_key": "test-key",
			},
			Temperature: &temp,
		}

		err := config.Validate()
		if err == nil {
			t.Error("expected error for temperature > 2")
		}
	})

	t.Run("invalid temperature too low", func(t *testing.T) {
		temp := -0.5
		config := &Config{
			Provider: "openai",
			ProviderOptions: map[string]interface{}{
				"api_key": "test-key",
			},
			Temperature: &temp,
		}

		err := config.Validate()
		if err == nil {
			t.Error("expected error for temperature < 0")
		}
	})

	t.Run("invalid max_tokens", func(t *testing.T) {
		maxTokens := -100
		config := &Config{
			Provider: "openai",
			ProviderOptions: map[string]interface{}{
				"api_key": "test-key",
			},
			MaxTokens: &maxTokens,
		}

		err := config.Validate()
		if err == nil {
			t.Error("expected error for max_tokens <= 0")
		}
	})

	t.Run("invalid top_p too high", func(t *testing.T) {
		topP := 1.5
		config := &Config{
			Provider: "openai",
			ProviderOptions: map[string]interface{}{
				"api_key": "test-key",
			},
			TopP: &topP,
		}

		err := config.Validate()
		if err == nil {
			t.Error("expected error for top_p > 1")
		}
	})

	t.Run("invalid top_p too low", func(t *testing.T) {
		topP := -0.1
		config := &Config{
			Provider: "openai",
			ProviderOptions: map[string]interface{}{
				"api_key": "test-key",
			},
			TopP: &topP,
		}

		err := config.Validate()
		if err == nil {
			t.Error("expected error for top_p < 0")
		}
	})

	t.Run("auto-create http client", func(t *testing.T) {
		config := &Config{
			Provider: "openai",
			ProviderOptions: map[string]interface{}{
				"api_key": "test-key",
			},
		}

		err := config.Validate()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if config.HTTPClient == nil {
			t.Error("expected HTTPClient to be auto-created")
		}
		if config.HTTPClient.Timeout != 60*time.Second {
			t.Errorf("expected default timeout 60s, got %v", config.HTTPClient.Timeout)
		}
	})

	t.Run("sync timeout with http client", func(t *testing.T) {
		config := &Config{
			Provider: "openai",
			ProviderOptions: map[string]interface{}{
				"api_key": "test-key",
			},
			HTTPClient: &http.Client{
				Timeout: 30 * time.Second,
			},
			Timeout: 45 * time.Second,
		}

		err := config.Validate()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if config.HTTPClient.Timeout != 45*time.Second {
			t.Errorf("expected timeout to be synced to 45s, got %v", config.HTTPClient.Timeout)
		}
	})

	t.Run("azure requires endpoint and deployment", func(t *testing.T) {
		config := &Config{
			Provider: "azure",
			ProviderOptions: map[string]interface{}{
				"api_key": "test-key",
			},
		}

		err := config.Validate()
		if err == nil {
			t.Error("expected error for missing azure endpoint")
		}

		config.ProviderOptions["endpoint"] = "https://example.openai.azure.com"
		err = config.Validate()
		if err == nil {
			t.Error("expected error for missing azure deployment")
		}

		config.ProviderOptions["deployment"] = "gpt-4"
		err = config.Validate()
		if err != nil {
			t.Errorf("expected no error with complete azure config, got %v", err)
		}
	})

	t.Run("ollama doesn't require api key", func(t *testing.T) {
		config := &Config{
			Provider:        "ollama",
			ProviderOptions: map[string]interface{}{},
		}

		err := config.Validate()
		if err != nil {
			t.Errorf("expected no error for ollama without api_key, got %v", err)
		}
	})
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.HTTPClient == nil {
		t.Error("expected HTTPClient to be set")
	}
	if config.HTTPClient.Timeout != 60*time.Second {
		t.Errorf("expected timeout 60s, got %v", config.HTTPClient.Timeout)
	}
	if config.Timeout != 60*time.Second {
		t.Errorf("expected timeout 60s, got %v", config.Timeout)
	}
	if config.Debug {
		t.Error("expected debug to be false")
	}
	if config.ProviderOptions == nil {
		t.Error("expected ProviderOptions to be initialized")
	}
}

func TestConfig_Clone(t *testing.T) {
	temp := 0.7
	maxTokens := 1000
	original := &Config{
		Provider:     "openai",
		DefaultModel: "gpt-4",
		Temperature:  &temp,
		MaxTokens:    &maxTokens,
		ProviderOptions: map[string]interface{}{
			"api_key": "test-key",
			"org":     "test-org",
		},
	}

	clone := original.Clone()

	// Check values are the same
	if clone.Provider != original.Provider {
		t.Error("provider not cloned correctly")
	}
	if clone.DefaultModel != original.DefaultModel {
		t.Error("model not cloned correctly")
	}

	// Check that modifying clone doesn't affect original
	clone.Provider = "anthropic"
	if original.Provider == "anthropic" {
		t.Error("modifying clone affected original")
	}

	// Check that ProviderOptions is deep cloned
	clone.ProviderOptions["api_key"] = "new-key"
	if original.ProviderOptions["api_key"] == "new-key" {
		t.Error("modifying clone ProviderOptions affected original")
	}
}
