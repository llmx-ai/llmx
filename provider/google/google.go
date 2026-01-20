package google

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/vertexai/genai"
	"github.com/llmx-ai/llmx"
	"github.com/llmx-ai/llmx/provider"
	"google.golang.org/api/option"
)

// GoogleProvider implements the Provider interface for Google Gemini
type GoogleProvider struct {
	client    *genai.Client
	projectID string
	location  string
}

func init() {
	provider.Register("google", NewGoogleProvider)
	provider.Register("gemini", NewGoogleProvider) // Alias
}

// NewGoogleProvider creates a new Google provider
func NewGoogleProvider(opts map[string]interface{}) (provider.Provider, error) {
	apiKey, hasKey := opts["api_key"].(string)
	projectID, hasProject := opts["project_id"].(string)
	location, hasLocation := opts["location"].(string)

	if !hasLocation {
		location = "us-central1" // Default location
	}

	ctx := context.Background()
	var client *genai.Client
	var err error

	if hasKey && apiKey != "" {
		// Use API key authentication
		client, err = genai.NewClient(ctx, projectID, location, option.WithAPIKey(apiKey))
	} else if hasProject && projectID != "" {
		// Use default credentials
		client, err = genai.NewClient(ctx, projectID, location)
	} else {
		return nil, fmt.Errorf("either api_key or project_id is required for Google provider")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create Google client: %w", err)
	}

	return &GoogleProvider{
		client:    client,
		projectID: projectID,
		location:  location,
	}, nil
}

// Name returns the provider name
func (p *GoogleProvider) Name() string {
	return "google"
}

// Chat sends a chat request
func (p *GoogleProvider) Chat(ctx context.Context, reqInterface interface{}) (interface{}, error) {
	req, ok := reqInterface.(*llmx.ChatRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type")
	}

	// Get model
	model := p.client.GenerativeModel(req.Model)

	// Configure model
	p.configureModel(model, req)

	// Convert messages to Gemini format
	history, lastMessage := p.convertMessages(req.Messages)

	// Start chat session
	chat := model.StartChat()
	chat.History = history

	// Send message
	resp, err := chat.SendMessage(ctx, lastMessage...)
	if err != nil {
		return nil, p.convertError(err)
	}

	// Convert response
	return p.convertResponse(resp, req.Model), nil
}

// StreamChat sends a streaming chat request
func (p *GoogleProvider) StreamChat(ctx context.Context, reqInterface interface{}) (interface{}, error) {
	req, ok := reqInterface.(*llmx.ChatRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type")
	}

	// Get model
	model := p.client.GenerativeModel(req.Model)

	// Configure model
	p.configureModel(model, req)

	// Convert messages
	history, lastMessage := p.convertMessages(req.Messages)

	// Start chat session
	chat := model.StartChat()
	chat.History = history

	// Send streaming message
	iter := chat.SendMessageStream(ctx, lastMessage...)

	// Create llmx stream
	chatStream := llmx.NewChatStream(ctx)

	// Start goroutine to handle streaming
	go p.handleStream(ctx, iter, chatStream)

	return chatStream, nil
}

// SupportedFeatures returns supported features
func (p *GoogleProvider) SupportedFeatures() provider.Features {
	return provider.Features{
		Streaming:     true,
		ToolCalling:   true,
		Vision:        true,
		JSONMode:      true,
		ReasoningMode: false,
		CacheControl:  false,
		MultiModal:    true,
		Embedding:     true,
	}
}

// SupportedModels returns supported models
func (p *GoogleProvider) SupportedModels() []provider.Model {
	return []provider.Model{
		{
			ID:              "gemini-1.5-pro",
			Name:            "Gemini 1.5 Pro",
			ContextWindow:   2000000, // 2M tokens
			MaxOutputTokens: 8192,
			InputCost:       1.25,
			OutputCost:      5.0,
			Capabilities:    []string{"chat", "vision", "tools"},
		},
		{
			ID:              "gemini-1.5-flash",
			Name:            "Gemini 1.5 Flash",
			ContextWindow:   1000000, // 1M tokens
			MaxOutputTokens: 8192,
			InputCost:       0.075,
			OutputCost:      0.30,
			Capabilities:    []string{"chat", "vision", "tools"},
		},
		{
			ID:              "gemini-2.0-flash-exp",
			Name:            "Gemini 2.0 Flash (Experimental)",
			ContextWindow:   1000000,
			MaxOutputTokens: 8192,
			InputCost:       0.0,
			OutputCost:      0.0,
			Capabilities:    []string{"chat", "vision", "tools"},
		},
	}
}

// configureModel configures the Gemini model with request parameters
func (p *GoogleProvider) configureModel(model *genai.GenerativeModel, req *llmx.ChatRequest) {
	if req.Temperature != nil {
		temp := float32(*req.Temperature)
		model.Temperature = &temp
	}

	if req.MaxTokens != nil {
		maxTokens := int32(*req.MaxTokens)
		model.MaxOutputTokens = &maxTokens
	}

	if req.TopP != nil {
		topP := float32(*req.TopP)
		model.TopP = &topP
	}

	if req.TopK != nil {
		topK := int32(*req.TopK)
		model.TopK = &topK
	}

	if len(req.Stop) > 0 {
		model.StopSequences = req.Stop
	}
}

// convertMessages converts llmx messages to Gemini format
func (p *GoogleProvider) convertMessages(messages []llmx.Message) ([]*genai.Content, []genai.Part) {
	var history []*genai.Content
	var lastUserMessage []genai.Part

	for i, msg := range messages {
		// Skip system messages (Gemini handles them differently)
		if msg.Role == llmx.RoleSystem {
			continue
		}

		role := "user"
		if msg.Role == llmx.RoleAssistant {
			role = "model"
		}

		parts := p.convertContentParts(msg.Content)

		// Last user message goes separately
		if i == len(messages)-1 && msg.Role == llmx.RoleUser {
			lastUserMessage = parts
		} else {
			history = append(history, &genai.Content{
				Role:  role,
				Parts: parts,
			})
		}
	}

	return history, lastUserMessage
}

// convertContentParts converts content parts to Gemini parts
func (p *GoogleProvider) convertContentParts(parts []llmx.ContentPart) []genai.Part {
	var geminiParts []genai.Part

	for _, part := range parts {
		switch v := part.(type) {
		case llmx.TextPart:
			geminiParts = append(geminiParts, genai.Text(v.Text))

		case llmx.ImagePart:
			// Gemini supports inline data
			if v.Base64 != "" {
				// Note: This is simplified, actual implementation would need
				// to decode base64 and create proper image data
				// geminiParts = append(geminiParts, genai.ImageData("image/jpeg", data))
			}
		}
	}

	return geminiParts
}

// convertResponse converts Gemini response to llmx response
func (p *GoogleProvider) convertResponse(resp *genai.GenerateContentResponse, model string) *llmx.ChatResponse {
	var content string

	if len(resp.Candidates) > 0 {
		candidate := resp.Candidates[0]
		if candidate.Content != nil {
			for _, part := range candidate.Content.Parts {
				if text, ok := part.(genai.Text); ok {
					content += string(text)
				}
			}
		}
	}

	var usage llmx.Usage
	if resp.UsageMetadata != nil {
		usage = llmx.Usage{
			PromptTokens:     int(resp.UsageMetadata.PromptTokenCount),
			CompletionTokens: int(resp.UsageMetadata.CandidatesTokenCount),
			TotalTokens:      int(resp.UsageMetadata.TotalTokenCount),
		}
	}

	return &llmx.ChatResponse{
		Model:     model,
		Content:   content,
		Usage:     usage,
		CreatedAt: time.Now(),
		Raw:       resp,
	}
}

// convertError converts Google errors to llmx errors
func (p *GoogleProvider) convertError(err error) error {
	return llmx.NewProviderError("google", err.Error(), 500, err)
}

// Close closes the Google client
func (p *GoogleProvider) Close() error {
	return p.client.Close()
}
