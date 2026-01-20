package deepseek

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/llmx-ai/llmx"
	"github.com/llmx-ai/llmx/core"
)

func TestNewDeepSeekProvider(t *testing.T) {
	tests := []struct {
		name    string
		opts    map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config",
			opts: map[string]interface{}{
				"api_key": "test-key",
			},
			wantErr: false,
		},
		{
			name: "valid config with custom base_url",
			opts: map[string]interface{}{
				"api_key":  "test-key",
				"base_url": "https://custom.deepseek.com/v1",
			},
			wantErr: false,
		},
		{
			name:    "missing api_key",
			opts:    map[string]interface{}{},
			wantErr: true,
		},
		{
			name: "empty api_key",
			opts: map[string]interface{}{
				"api_key": "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewDeepSeekProvider(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDeepSeekProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && provider == nil {
				t.Error("NewDeepSeekProvider() returned nil provider")
			}
			if !tt.wantErr && provider.Name() != "deepseek" {
				t.Errorf("Provider name = %v, want deepseek", provider.Name())
			}
		})
	}
}

func TestDeepSeekProvider_Name(t *testing.T) {
	provider, err := NewDeepSeekProvider(map[string]interface{}{
		"api_key": "test-key",
	})
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	if got := provider.Name(); got != "deepseek" {
		t.Errorf("Name() = %v, want deepseek", got)
	}
}

func TestDeepSeekProvider_SupportedFeatures(t *testing.T) {
	provider, err := NewDeepSeekProvider(map[string]interface{}{
		"api_key": "test-key",
	})
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	features := provider.SupportedFeatures()

	if !features.Chat {
		t.Error("Chat should be supported")
	}
	if !features.Streaming {
		t.Error("Streaming should be supported")
	}
	if !features.Tools {
		t.Error("Tools should be supported")
	}
	if features.Vision {
		t.Error("Vision should not be supported")
	}
	if !features.JSON {
		t.Error("JSON mode should be supported")
	}
	if !features.SystemPrompt {
		t.Error("System prompt should be supported")
	}
}

func TestDeepSeekProvider_SupportedModels(t *testing.T) {
	provider, err := NewDeepSeekProvider(map[string]interface{}{
		"api_key": "test-key",
	})
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	models := provider.SupportedModels()

	if len(models) == 0 {
		t.Error("Should return at least one model")
	}

	// Check for key models
	modelIDs := make(map[string]bool)
	for _, model := range models {
		modelIDs[model.ID] = true

		// Validate model has required fields
		if model.ID == "" {
			t.Error("Model ID should not be empty")
		}
		if model.Name == "" {
			t.Error("Model Name should not be empty")
		}
		if model.ContextSize == 0 {
			t.Errorf("Model %s should have context size", model.ID)
		}
	}

	// Check for specific models
	expectedModels := []string{
		"deepseek-chat",
		"deepseek-coder",
	}

	for _, expected := range expectedModels {
		if !modelIDs[expected] {
			t.Errorf("Expected model %s not found", expected)
		}
	}
}

func TestDeepSeekProvider_Chat(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-key" {
			t.Errorf("Expected Authorization header 'Bearer test-key', got '%s'", authHeader)
		}

		// Return mock response
		response := map[string]interface{}{
			"id":      "chatcmpl-test",
			"object":  "chat.completion",
			"created": 1234567890,
			"model":   "deepseek-chat",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "Hello from DeepSeek!",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     10,
				"completion_tokens": 5,
				"total_tokens":      15,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create provider with mock server URL
	provider, err := NewDeepSeekProvider(map[string]interface{}{
		"api_key":  "test-key",
		"base_url": server.URL,
	})
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Create chat request
	req := &llmx.ChatRequest{
		Model: "deepseek-chat",
		Messages: []core.Message{
			{
				Role: llmx.RoleUser,
				Content: []core.ContentPart{
					&core.TextPart{Text: "Hello"},
				},
			},
		},
	}

	// Send chat request
	ctx := context.Background()
	resp, err := provider.Chat(ctx, req)
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}

	// Check response
	if resp == nil {
		t.Fatal("Chat() returned nil response")
	}

	if len(resp.Choices) == 0 {
		t.Fatal("Chat() returned no choices")
	}

	choice := resp.Choices[0]
	if choice.Message.Role != llmx.RoleAssistant {
		t.Errorf("Expected role 'assistant', got '%s'", choice.Message.Role)
	}

	if len(choice.Message.Content) == 0 {
		t.Fatal("Message has no content")
	}

	textPart, ok := choice.Message.Content[0].(*core.TextPart)
	if !ok {
		t.Fatal("Expected TextPart content")
	}

	if textPart.Text != "Hello from DeepSeek!" {
		t.Errorf("Expected content 'Hello from DeepSeek!', got '%s'", textPart.Text)
	}

	// Check usage
	if resp.Usage.PromptTokens != 10 {
		t.Errorf("Expected PromptTokens 10, got %d", resp.Usage.PromptTokens)
	}
	if resp.Usage.CompletionTokens != 5 {
		t.Errorf("Expected CompletionTokens 5, got %d", resp.Usage.CompletionTokens)
	}
	if resp.Usage.TotalTokens != 15 {
		t.Errorf("Expected TotalTokens 15, got %d", resp.Usage.TotalTokens)
	}
}
