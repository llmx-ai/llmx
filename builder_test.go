package llmx

import (
	"testing"
	"time"
)

func TestClientBuilder_Validation(t *testing.T) {
	t.Run("missing provider", func(t *testing.T) {
		_, err := NewClientBuilder().
			Model("gpt-4").
			Build()

		if err == nil {
			t.Fatal("expected error for missing provider")
		}
	})

	t.Run("missing api key", func(t *testing.T) {
		_, err := NewClientBuilder().
			Provider("openai", "").
			Build()

		if err == nil {
			t.Fatal("expected error for missing api key")
		}
	})

	t.Run("invalid temperature", func(t *testing.T) {
		builder := NewClientBuilder().
			Provider("openai", "test-key").
			Temperature(3.0) // Invalid: > 2

		if builder.err == nil {
			t.Fatal("expected error for invalid temperature")
		}
	})

	t.Run("invalid max tokens", func(t *testing.T) {
		builder := NewClientBuilder().
			Provider("openai", "test-key").
			MaxTokens(-1) // Invalid: <= 0

		if builder.err == nil {
			t.Fatal("expected error for invalid max tokens")
		}
	})

	t.Run("invalid top_p", func(t *testing.T) {
		builder := NewClientBuilder().
			Provider("openai", "test-key").
			TopP(1.5) // Invalid: > 1

		if builder.err == nil {
			t.Fatal("expected error for invalid top_p")
		}
	})

	t.Run("invalid timeout", func(t *testing.T) {
		builder := NewClientBuilder().
			Provider("openai", "test-key").
			Timeout(-1 * time.Second) // Invalid: <= 0

		if builder.err == nil {
			t.Fatal("expected error for invalid timeout")
		}
	})

	t.Run("invalid base url", func(t *testing.T) {
		builder := NewClientBuilder().
			Provider("openai", "test-key").
			BaseURL("") // Invalid: empty

		if builder.err == nil {
			t.Fatal("expected error for empty base_url")
		}
	})

	t.Run("chaining with error", func(t *testing.T) {
		// Test that errors are accumulated through the chain
		builder := NewClientBuilder().
			Provider("openai", "test-key").
			Temperature(5.0). // Invalid
			MaxTokens(1000)   // This should still be called

		if builder.err == nil {
			t.Fatal("expected error to be propagated")
		}
	})
}

func TestClientBuilder_Configuration(t *testing.T) {
	t.Run("basic configuration", func(t *testing.T) {
		builder := NewClientBuilder().
			Provider("openai", "test-key").
			Model("gpt-4").
			Temperature(0.7).
			MaxTokens(1000)

		if builder.config.Provider != "openai" {
			t.Errorf("expected provider 'openai', got %s", builder.config.Provider)
		}
		if builder.config.DefaultModel != "gpt-4" {
			t.Errorf("expected model 'gpt-4', got %s", builder.config.DefaultModel)
		}
		if *builder.config.Temperature != 0.7 {
			t.Errorf("expected temperature 0.7, got %.2f", *builder.config.Temperature)
		}
		if *builder.config.MaxTokens != 1000 {
			t.Errorf("expected max_tokens 1000, got %d", *builder.config.MaxTokens)
		}

		apiKey, _ := builder.config.ProviderOptions["api_key"].(string)
		if apiKey != "test-key" {
			t.Errorf("expected api_key 'test-key', got %s", apiKey)
		}
	})

	t.Run("custom timeout", func(t *testing.T) {
		timeout := 30 * time.Second
		builder := NewClientBuilder().
			Provider("openai", "test-key").
			Timeout(timeout)

		if builder.config.HTTPClient.Timeout != timeout {
			t.Errorf("expected timeout %v, got %v", timeout, builder.config.HTTPClient.Timeout)
		}
	})

	t.Run("base url", func(t *testing.T) {
		builder := NewClientBuilder().
			Provider("compatible", "test-key").
			BaseURL("https://api.example.com/v1")

		baseURL, _ := builder.config.ProviderOptions["base_url"].(string)
		if baseURL != "https://api.example.com/v1" {
			t.Errorf("expected base_url 'https://api.example.com/v1', got %s", baseURL)
		}
	})

	t.Run("organization", func(t *testing.T) {
		builder := NewClientBuilder().
			Provider("openai", "test-key").
			Organization("org-123")

		org, _ := builder.config.ProviderOptions["organization"].(string)
		if org != "org-123" {
			t.Errorf("expected organization 'org-123', got %s", org)
		}
	})

	t.Run("region", func(t *testing.T) {
		builder := NewClientBuilder().
			Provider("azure", "test-key").
			Region("eastus")

		region, _ := builder.config.ProviderOptions["region"].(string)
		if region != "eastus" {
			t.Errorf("expected region 'eastus', got %s", region)
		}
	})
}
