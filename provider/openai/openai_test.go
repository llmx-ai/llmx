package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/llmx-ai/llmx"
)

func TestNewOpenAIProvider(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		opts := map[string]interface{}{
			"api_key": "test-key",
		}

		provider, err := NewOpenAIProvider(opts)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if provider == nil {
			t.Fatal("expected provider to be created")
		}

		if provider.Name() != "openai" {
			t.Errorf("expected provider name 'openai', got %s", provider.Name())
		}
	})

	t.Run("missing api key", func(t *testing.T) {
		opts := map[string]interface{}{}

		_, err := NewOpenAIProvider(opts)
		if err == nil {
			t.Error("expected error for missing api_key")
		}
	})

	t.Run("with custom base url", func(t *testing.T) {
		opts := map[string]interface{}{
			"api_key":  "test-key",
			"base_url": "https://custom.openai.com/v1",
		}

		provider, err := NewOpenAIProvider(opts)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if provider == nil {
			t.Fatal("expected provider to be created")
		}
	})
}

func TestOpenAIProvider_Chat(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		if r.URL.Path != "/chat/completions" {
			t.Errorf("expected path /chat/completions, got %s", r.URL.Path)
		}

		// Verify headers
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("missing or incorrect Authorization header")
		}

		// Send mock response
		response := map[string]interface{}{
			"id":      "chatcmpl-123",
			"object":  "chat.completion",
			"created": time.Now().Unix(),
			"model":   "gpt-4",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "Hello! How can I help you today?",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     10,
				"completion_tokens": 20,
				"total_tokens":      30,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create provider with mock server
	opts := map[string]interface{}{
		"api_key":  "test-key",
		"base_url": server.URL,
	}

	provider, err := NewOpenAIProvider(opts)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	// Create request
	req := &llmx.ChatRequest{
		Model: "gpt-4",
		Messages: []llmx.Message{
			{
				Role: llmx.RoleUser,
				Content: []llmx.ContentPart{
					llmx.TextPart{Text: "Hello"},
				},
			},
		},
	}

	// Call Chat
	ctx := context.Background()
	respInterface, err := provider.Chat(ctx, req)
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}

	// Verify response
	resp, ok := respInterface.(*llmx.ChatResponse)
	if !ok {
		t.Fatalf("expected *llmx.ChatResponse, got %T", respInterface)
	}

	if resp.Content != "Hello! How can I help you today?" {
		t.Errorf("expected content 'Hello! How can I help you today?', got %s", resp.Content)
	}

	if resp.Model != "gpt-4" {
		t.Errorf("expected model 'gpt-4', got %s", resp.Model)
	}

	if resp.Usage.PromptTokens != 10 {
		t.Errorf("expected 10 prompt tokens, got %d", resp.Usage.PromptTokens)
	}

	if resp.Usage.CompletionTokens != 20 {
		t.Errorf("expected 20 completion tokens, got %d", resp.Usage.CompletionTokens)
	}
}

func TestOpenAIProvider_Chat_InvalidRequest(t *testing.T) {
	opts := map[string]interface{}{
		"api_key": "test-key",
	}

	provider, err := NewOpenAIProvider(opts)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	ctx := context.Background()

	// Test with invalid type
	_, err = provider.Chat(ctx, "invalid")
	if err == nil {
		t.Error("expected error for invalid request type")
	}
}

func TestOpenAIProvider_SupportedFeatures(t *testing.T) {
	opts := map[string]interface{}{
		"api_key": "test-key",
	}

	provider, err := NewOpenAIProvider(opts)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	features := provider.SupportedFeatures()

	if !features.Streaming {
		t.Error("expected streaming to be supported")
	}

	if !features.ToolCalling {
		t.Error("expected tool calling to be supported")
	}

	if !features.Vision {
		t.Error("expected vision to be supported")
	}

	if !features.JSONMode {
		t.Error("expected JSON mode to be supported")
	}
}

func TestOpenAIProvider_SupportedModels(t *testing.T) {
	opts := map[string]interface{}{
		"api_key": "test-key",
	}

	provider, err := NewOpenAIProvider(opts)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	models := provider.SupportedModels()

	if len(models) == 0 {
		t.Error("expected some models to be supported")
	}

	// Check for GPT-4
	foundGPT4 := false
	for _, model := range models {
		if model.ID == "gpt-4" {
			foundGPT4 = true
			if model.ContextWindow != 8192 {
				t.Errorf("expected GPT-4 context window 8192, got %d", model.ContextWindow)
			}
			break
		}
	}

	if !foundGPT4 {
		t.Error("expected GPT-4 to be in supported models")
	}
}

func TestConvertRequest(t *testing.T) {
	opts := map[string]interface{}{
		"api_key": "test-key",
	}

	provider, err := NewOpenAIProvider(opts)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	p := provider.(*OpenAIProvider)

	t.Run("basic text message", func(t *testing.T) {
		req := &llmx.ChatRequest{
			Model: "gpt-4",
			Messages: []llmx.Message{
				{
					Role: llmx.RoleUser,
					Content: []llmx.ContentPart{
						llmx.TextPart{Text: "Hello"},
					},
				},
			},
		}

		openaiReq := p.convertRequest(req)

		if openaiReq.Model != "gpt-4" {
			t.Errorf("expected model 'gpt-4', got %s", openaiReq.Model)
		}

		if len(openaiReq.Messages) != 1 {
			t.Fatalf("expected 1 message, got %d", len(openaiReq.Messages))
		}

		if openaiReq.Messages[0].Role != "user" {
			t.Errorf("expected role 'user', got %s", openaiReq.Messages[0].Role)
		}

		if openaiReq.Messages[0].Content != "Hello" {
			t.Errorf("expected content 'Hello', got %s", openaiReq.Messages[0].Content)
		}
	})

	t.Run("with temperature", func(t *testing.T) {
		temp := 0.7
		req := &llmx.ChatRequest{
			Model:       "gpt-4",
			Temperature: &temp,
			Messages: []llmx.Message{
				{
					Role: llmx.RoleUser,
					Content: []llmx.ContentPart{
						llmx.TextPart{Text: "Hello"},
					},
				},
			},
		}

		openaiReq := p.convertRequest(req)

		if openaiReq.Temperature != 0.7 {
			t.Errorf("expected temperature 0.7, got %f", openaiReq.Temperature)
		}
	})

	t.Run("with max tokens", func(t *testing.T) {
		maxTokens := 1000
		req := &llmx.ChatRequest{
			Model:     "gpt-4",
			MaxTokens: &maxTokens,
			Messages: []llmx.Message{
				{
					Role: llmx.RoleUser,
					Content: []llmx.ContentPart{
						llmx.TextPart{Text: "Hello"},
					},
				},
			},
		}

		openaiReq := p.convertRequest(req)

		if openaiReq.MaxTokens != 1000 {
			t.Errorf("expected max_tokens 1000, got %d", openaiReq.MaxTokens)
		}
	})
}

func TestConvertError(t *testing.T) {
	opts := map[string]interface{}{
		"api_key": "test-key",
	}

	provider, err := NewOpenAIProvider(opts)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	p := provider.(*OpenAIProvider)

	t.Run("generic error", func(t *testing.T) {
		err := p.convertError(http.ErrServerClosed)

		llmxErr, ok := err.(llmx.Error)
		if !ok {
			t.Fatalf("expected llmx.Error, got %T", err)
		}

		if llmxErr.StatusCode() == 0 {
			t.Error("expected non-zero status code")
		}
	})
}
