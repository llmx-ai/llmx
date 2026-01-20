package llmx

import (
	"context"
	"testing"
)

func TestClient_ResourceManagement(t *testing.T) {
	t.Run("close client", func(t *testing.T) {
		config := &Config{
			Provider:        "mock",
			ProviderOptions: map[string]interface{}{},
		}

		client, err := NewClient(WithConfig(config))
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}

		if client.IsClosed() {
			t.Error("client should not be closed initially")
		}

		err = client.Close()
		if err != nil {
			t.Errorf("expected no error closing client, got %v", err)
		}

		if !client.IsClosed() {
			t.Error("client should be closed after Close()")
		}
	})

	t.Run("close multiple times", func(t *testing.T) {
		config := &Config{
			Provider:        "mock",
			ProviderOptions: map[string]interface{}{},
		}

		client, err := NewClient(WithConfig(config))
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}

		// Close multiple times should not panic
		err1 := client.Close()
		err2 := client.Close()
		err3 := client.Close()

		if err1 != nil || err2 != nil || err3 != nil {
			t.Error("multiple Close() calls should not error")
		}

		if !client.IsClosed() {
			t.Error("client should be closed")
		}
	})

	t.Run("use closed client", func(t *testing.T) {
		config := &Config{
			Provider:        "mock",
			ProviderOptions: map[string]interface{}{},
		}

		client, err := NewClient(WithConfig(config))
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}

		err = client.Close()
		if err != nil {
			t.Errorf("expected no error closing client, got %v", err)
		}

		// Try to use closed client
		ctx := context.Background()
		req := &ChatRequest{
			Model: "gpt-4",
			Messages: []Message{
				{
					Role:    RoleUser,
					Content: []ContentPart{TextPart{Text: "Hello"}},
				},
			},
		}

		_, err = client.Chat(ctx, req)
		if err == nil {
			t.Error("expected error when using closed client for Chat")
		}

		_, err = client.StreamChat(ctx, req)
		if err == nil {
			t.Error("expected error when using closed client for StreamChat")
		}
	})
}
