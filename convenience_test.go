package llmx

import (
	"testing"
)

func TestMessageBuilder(t *testing.T) {
	t.Run("basic text message", func(t *testing.T) {
		msg := NewUserMessage().
			Text("Hello, world!").
			MustBuild()

		if msg.Role != RoleUser {
			t.Errorf("expected role %s, got %s", RoleUser, msg.Role)
		}
		if len(msg.Content) != 1 {
			t.Fatalf("expected 1 content part, got %d", len(msg.Content))
		}
		if textPart, ok := msg.Content[0].(TextPart); ok {
			if textPart.Text != "Hello, world!" {
				t.Errorf("expected text 'Hello, world!', got %s", textPart.Text)
			}
		} else {
			t.Error("expected TextPart")
		}
	})

	t.Run("multimodal message", func(t *testing.T) {
		msg := NewUserMessage().
			Text("What's in this image?").
			Image("https://example.com/image.jpg", "high").
			MustBuild()

		if len(msg.Content) != 2 {
			t.Fatalf("expected 2 content parts, got %d", len(msg.Content))
		}

		if textPart, ok := msg.Content[0].(TextPart); ok {
			if textPart.Text != "What's in this image?" {
				t.Errorf("expected text 'What's in this image?', got %s", textPart.Text)
			}
		} else {
			t.Error("expected TextPart as first part")
		}

		if imagePart, ok := msg.Content[1].(ImagePart); ok {
			if imagePart.URL != "https://example.com/image.jpg" {
				t.Errorf("expected URL 'https://example.com/image.jpg', got %s", imagePart.URL)
			}
			if imagePart.Detail != "high" {
				t.Errorf("expected detail 'high', got %s", imagePart.Detail)
			}
		} else {
			t.Error("expected ImagePart as second part")
		}
	})

	t.Run("assistant message with tool call", func(t *testing.T) {
		msg := NewAssistantMessage().
			ToolCall("call_123", "calculator", []byte(`{"a": 1, "b": 2}`)).
			MustBuild()

		if msg.Role != RoleAssistant {
			t.Errorf("expected role %s, got %s", RoleAssistant, msg.Role)
		}
		if len(msg.Content) != 1 {
			t.Fatalf("expected 1 content part, got %d", len(msg.Content))
		}
		if toolCall, ok := msg.Content[0].(ToolCall); ok {
			if toolCall.ID != "call_123" {
				t.Errorf("expected ID 'call_123', got %s", toolCall.ID)
			}
			if toolCall.Name != "calculator" {
				t.Errorf("expected name 'calculator', got %s", toolCall.Name)
			}
		} else {
			t.Error("expected ToolCall")
		}
	})

	t.Run("empty message error", func(t *testing.T) {
		_, err := NewUserMessage().Build()
		if err == nil {
			t.Error("expected error for empty message")
		}
	})
}

func TestRequestBuilder(t *testing.T) {
	t.Run("basic request", func(t *testing.T) {
		req := NewRequest("gpt-4").
			UserMessage("Hello").
			AssistantMessage("Hi there!").
			SystemMessage("You are helpful").
			MustBuild()

		if req.Model != "gpt-4" {
			t.Errorf("expected model 'gpt-4', got %s", req.Model)
		}
		if len(req.Messages) != 3 {
			t.Fatalf("expected 3 messages, got %d", len(req.Messages))
		}
		if req.Messages[0].Role != RoleUser {
			t.Errorf("expected first message role %s, got %s", RoleUser, req.Messages[0].Role)
		}
		if req.Messages[1].Role != RoleAssistant {
			t.Errorf("expected second message role %s, got %s", RoleAssistant, req.Messages[1].Role)
		}
		if req.Messages[2].Role != RoleSystem {
			t.Errorf("expected third message role %s, got %s", RoleSystem, req.Messages[2].Role)
		}
	})

	t.Run("with parameters", func(t *testing.T) {
		temp := 0.7
		maxTokens := 1000
		topP := 0.9

		req := NewRequest("gpt-4").
			UserMessage("Hello").
			Temperature(temp).
			MaxTokens(maxTokens).
			TopP(topP).
			MustBuild()

		if *req.Temperature != temp {
			t.Errorf("expected temperature %.2f, got %.2f", temp, *req.Temperature)
		}
		if *req.MaxTokens != maxTokens {
			t.Errorf("expected max_tokens %d, got %d", maxTokens, *req.MaxTokens)
		}
		if *req.TopP != topP {
			t.Errorf("expected top_p %.2f, got %.2f", topP, *req.TopP)
		}
	})

	t.Run("with custom message", func(t *testing.T) {
		customMsg := NewUserMessage().
			Text("Hello").
			Image("https://example.com/image.jpg").
			MustBuild()

		req := NewRequest("gpt-4").
			Message(customMsg).
			MustBuild()

		if len(req.Messages) != 1 {
			t.Fatalf("expected 1 message, got %d", len(req.Messages))
		}
		if len(req.Messages[0].Content) != 2 {
			t.Errorf("expected 2 content parts, got %d", len(req.Messages[0].Content))
		}
	})

	t.Run("empty request error", func(t *testing.T) {
		_, err := NewRequest("gpt-4").Build()
		if err == nil {
			t.Error("expected error for request with no messages")
		}
	})
}

func TestTextOnly(t *testing.T) {
	t.Run("extract text from response", func(t *testing.T) {
		resp := &ChatResponse{
			Content: "Hello, world!",
		}

		text := TextOnly(resp)
		if text != "Hello, world!" {
			t.Errorf("expected 'Hello, world!', got %s", text)
		}
	})

	t.Run("nil response", func(t *testing.T) {
		text := TextOnly(nil)
		if text != "" {
			t.Errorf("expected empty string, got %s", text)
		}
	})
}
