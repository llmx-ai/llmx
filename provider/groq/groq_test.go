package groq

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/llmx-ai/llmx"
	"github.com/llmx-ai/llmx/core"
)

func TestNewGroqProvider(t *testing.T) {
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
				"base_url": "https://custom.groq.com/v1",
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
			provider, err := NewGroqProvider(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGroqProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && provider == nil {
				t.Error("NewGroqProvider() returned nil provider")
			}
			if !tt.wantErr && provider.Name() != "groq" {
				t.Errorf("Provider name = %v, want groq", provider.Name())
			}
		})
	}
}

func TestGroqProvider_Name(t *testing.T) {
	provider, err := NewGroqProvider(map[string]interface{}{
		"api_key": "test-key",
	})
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	if got := provider.Name(); got != "groq" {
		t.Errorf("Name() = %v, want groq", got)
	}
}

func TestGroqProvider_SupportedFeatures(t *testing.T) {
	provider, err := NewGroqProvider(map[string]interface{}{
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

func TestGroqProvider_SupportedModels(t *testing.T) {
	provider, err := NewGroqProvider(map[string]interface{}{
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
		"llama-3.3-70b-versatile",
		"llama-3.1-70b-versatile",
		"llama-3.1-8b-instant",
		"mixtral-8x7b-32768",
	}

	for _, expected := range expectedModels {
		if !modelIDs[expected] {
			t.Errorf("Expected model %s not found", expected)
		}
	}
}

func TestGroqProvider_Chat(t *testing.T) {
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
			"model":   "llama-3.3-70b-versatile",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "Hello from Groq!",
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
	provider, err := NewGroqProvider(map[string]interface{}{
		"api_key":  "test-key",
		"base_url": server.URL,
	})
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Create chat request
	req := &llmx.ChatRequest{
		Model: "llama-3.3-70b-versatile",
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

	if textPart.Text != "Hello from Groq!" {
		t.Errorf("Expected content 'Hello from Groq!', got '%s'", textPart.Text)
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
