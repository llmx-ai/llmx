package llmx

import (
	"context"
	"testing"
)

// BenchmarkClient_Chat benchmarks basic chat requests
func BenchmarkClient_Chat(b *testing.B) {
	// Create mock client
	client := &Client{
		config: &Config{
			DefaultModel: "test-model",
		},
		provider: &mockProvider{},
	}

	req := &ChatRequest{
		Model: "test-model",
		Messages: []Message{
			{
				Role: RoleUser,
				Content: []ContentPart{
					TextPart{Text: "Hello"},
				},
			},
		},
	}

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = client.Chat(ctx, req)
	}
}

// BenchmarkMessage_Create benchmarks message creation
func BenchmarkMessage_Create(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = Message{
			Role: RoleUser,
			Content: []ContentPart{
				TextPart{Text: "Hello, world!"},
			},
		}
	}
}

// BenchmarkValidateMessage benchmarks message validation
func BenchmarkValidateMessage(b *testing.B) {
	msg := Message{
		Role: RoleUser,
		Content: []ContentPart{
			TextPart{Text: "Hello"},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = ValidateMessage(msg)
	}
}

// BenchmarkChatRequest_Creation benchmarks request creation
func BenchmarkChatRequest_Creation(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = &ChatRequest{
			Model: "gpt-4",
			Messages: []Message{
				{
					Role: RoleUser,
					Content: []ContentPart{
						TextPart{Text: "Hello"},
					},
				},
			},
		}
	}
}

// BenchmarkMultiMessage benchmarks multiple messages
func BenchmarkMultiMessage(b *testing.B) {
	client := &Client{
		config: &Config{
			DefaultModel: "test-model",
		},
		provider: &mockProvider{},
	}

	messages := make([]Message, 10)
	for i := range messages {
		messages[i] = Message{
			Role: RoleUser,
			Content: []ContentPart{
				TextPart{Text: "Message " + string(rune(i))},
			},
		}
	}

	req := &ChatRequest{
		Model:    "test-model",
		Messages: messages,
	}

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = client.Chat(ctx, req)
	}
}

// BenchmarkConcurrentRequests benchmarks concurrent requests
func BenchmarkConcurrentRequests(b *testing.B) {
	client := &Client{
		config: &Config{
			DefaultModel: "test-model",
		},
		provider: &mockProvider{},
	}

	req := &ChatRequest{
		Model: "test-model",
		Messages: []Message{
			{
				Role: RoleUser,
				Content: []ContentPart{
					TextPart{Text: "Hello"},
				},
			},
		},
	}

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = client.Chat(ctx, req)
		}
	})
}
