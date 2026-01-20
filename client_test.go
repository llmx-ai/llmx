package llmx

import (
	"context"
	"testing"

	"github.com/llmx-ai/llmx/provider"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr bool
	}{
		{
			name: "valid openai config",
			opts: []Option{
				WithOpenAI("test-key"),
			},
			wantErr: false,
		},
		{
			name:    "no provider",
			opts:    []Option{},
			wantErr: true,
		},
		{
			name: "with default model",
			opts: []Option{
				WithOpenAI("test-key"),
				WithDefaultModel("gpt-4"),
			},
			wantErr: false,
		},
		{
			name: "with temperature",
			opts: []Option{
				WithOpenAI("test-key"),
				WithTemperature(0.7),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client")
			}
		})
	}
}

func TestClient_validateRequest(t *testing.T) {
	client, _ := NewClient(WithOpenAI("test-key"))

	tests := []struct {
		name    string
		req     *ChatRequest
		wantErr bool
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name: "empty messages",
			req: &ChatRequest{
				Model:    "gpt-4",
				Messages: []Message{},
			},
			wantErr: true,
		},
		{
			name: "valid request",
			req: &ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{
						Role: RoleUser,
						Content: []ContentPart{
							TextPart{Text: "Hello"},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.validateRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_applyDefaults(t *testing.T) {
	defaultTemp := 0.7
	defaultTokens := 1000

	client, _ := NewClient(
		WithOpenAI("test-key"),
		WithDefaultModel("gpt-4"),
		WithTemperature(defaultTemp),
		WithMaxTokens(defaultTokens),
	)

	req := &ChatRequest{
		Messages: []Message{
			{
				Role: RoleUser,
				Content: []ContentPart{
					TextPart{Text: "Hello"},
				},
			},
		},
	}

	client.applyDefaults(req)

	if req.Model != "gpt-4" {
		t.Errorf("Expected model gpt-4, got %s", req.Model)
	}

	if req.Temperature == nil || *req.Temperature != defaultTemp {
		t.Errorf("Expected temperature %.1f, got %v", defaultTemp, req.Temperature)
	}

	if req.MaxTokens == nil || *req.MaxTokens != defaultTokens {
		t.Errorf("Expected max tokens %d, got %v", defaultTokens, req.MaxTokens)
	}
}

func TestClient_Config(t *testing.T) {
	client, _ := NewClient(
		WithOpenAI("test-key"),
		WithDefaultModel("gpt-4"),
	)

	config := client.Config()
	if config.DefaultModel != "gpt-4" {
		t.Errorf("Expected default model gpt-4, got %s", config.DefaultModel)
	}

	// Ensure it's a copy
	config.DefaultModel = "gpt-3.5-turbo"
	if client.config.DefaultModel != "gpt-4" {
		t.Error("Config() should return a copy, not the original")
	}
}

// Mock provider for testing
type mockProvider struct{}

func (m *mockProvider) Name() string {
	return "mock"
}

func (m *mockProvider) Chat(ctx context.Context, req interface{}) (interface{}, error) {
	return &ChatResponse{
		Content: "mock response",
	}, nil
}

func (m *mockProvider) StreamChat(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

func (m *mockProvider) SupportedFeatures() provider.Features {
	return provider.Features{}
}

func (m *mockProvider) SupportedModels() []provider.Model {
	return nil
}
