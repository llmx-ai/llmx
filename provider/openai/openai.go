package openai

import (
	"context"
	"fmt"
	"time"

	"github.com/llmx-ai/llmx"
	"github.com/llmx-ai/llmx/provider"
	openai "github.com/sashabaranov/go-openai"
)

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	client *openai.Client
}

func init() {
	provider.Register("openai", NewOpenAIProvider)
	provider.Register("compatible", NewOpenAIProvider) // For OpenAI-compatible APIs
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(opts map[string]interface{}) (provider.Provider, error) {
	apiKey, ok := opts["api_key"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("api_key is required for OpenAI provider")
	}

	config := openai.DefaultConfig(apiKey)

	// Check for custom base URL (for compatible APIs)
	if baseURL, ok := opts["base_url"].(string); ok && baseURL != "" {
		config.BaseURL = baseURL
	}

	return &OpenAIProvider{
		client: openai.NewClientWithConfig(config),
	}, nil
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// Chat sends a chat request
func (p *OpenAIProvider) Chat(ctx context.Context, reqInterface interface{}) (interface{}, error) {
	// Type assertion with detailed error
	req, ok := reqInterface.(*llmx.ChatRequest)
	if !ok {
		return nil, fmt.Errorf("openai: invalid request type %T, expected *llmx.ChatRequest", reqInterface)
	}

	// Convert request
	openaiReq := p.convertRequest(req)

	// Call OpenAI API
	resp, err := p.client.CreateChatCompletion(ctx, openaiReq)
	if err != nil {
		return nil, p.convertError(err)
	}

	// Convert response
	return p.convertResponse(&resp), nil
}

// StreamChat sends a streaming chat request
func (p *OpenAIProvider) StreamChat(ctx context.Context, reqInterface interface{}) (interface{}, error) {
	// Type assertion with detailed error
	req, ok := reqInterface.(*llmx.ChatRequest)
	if !ok {
		return nil, fmt.Errorf("openai: invalid request type %T, expected *llmx.ChatRequest", reqInterface)
	}

	// Convert request
	openaiReq := p.convertRequest(req)

	// Create stream
	stream, err := p.client.CreateChatCompletionStream(ctx, openaiReq)
	if err != nil {
		return nil, p.convertError(err)
	}

	// Create llmx stream
	chatStream := llmx.NewChatStream(ctx)

	// Start goroutine to handle streaming
	go p.handleStream(ctx, stream, chatStream)

	return chatStream, nil
}

// SupportedFeatures returns supported features
func (p *OpenAIProvider) SupportedFeatures() provider.Features {
	return provider.Features{
		Streaming:    true,
		ToolCalling:  true,
		Vision:       true,
		JSONMode:     true,
		MultiModal:   true,
		Embedding:    true,
		CacheControl: false,
	}
}

// SupportedModels returns supported models
func (p *OpenAIProvider) SupportedModels() []provider.Model {
	return []provider.Model{
		{
			ID:              "gpt-4-turbo",
			Name:            "GPT-4 Turbo",
			ContextWindow:   128000,
			MaxOutputTokens: 4096,
			InputCost:       10.0,
			OutputCost:      30.0,
			Capabilities:    []string{"chat", "vision", "tools"},
		},
		{
			ID:              "gpt-4",
			Name:            "GPT-4",
			ContextWindow:   8192,
			MaxOutputTokens: 4096,
			InputCost:       30.0,
			OutputCost:      60.0,
			Capabilities:    []string{"chat", "tools"},
		},
		{
			ID:              "gpt-3.5-turbo",
			Name:            "GPT-3.5 Turbo",
			ContextWindow:   16385,
			MaxOutputTokens: 4096,
			InputCost:       0.5,
			OutputCost:      1.5,
			Capabilities:    []string{"chat", "tools"},
		},
	}
}

// convertRequest converts llmx request to OpenAI request
func (p *OpenAIProvider) convertRequest(req *llmx.ChatRequest) openai.ChatCompletionRequest {
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = p.convertMessage(msg)
	}

	openaiReq := openai.ChatCompletionRequest{
		Model:    req.Model,
		Messages: messages,
	}

	if req.Temperature != nil {
		openaiReq.Temperature = float32(*req.Temperature)
	}

	if req.MaxTokens != nil {
		openaiReq.MaxTokens = *req.MaxTokens
	}

	if req.TopP != nil {
		openaiReq.TopP = float32(*req.TopP)
	}

	if len(req.Stop) > 0 {
		openaiReq.Stop = req.Stop
	}

	return openaiReq
}

// convertMessage converts a llmx message to OpenAI message
func (p *OpenAIProvider) convertMessage(msg llmx.Message) openai.ChatCompletionMessage {
	var content string
	var multiContent []openai.ChatMessagePart

	// Process content parts
	for _, part := range msg.Content {
		switch v := part.(type) {
		case llmx.TextPart:
			if len(multiContent) == 0 {
				content = v.Text
			} else {
				multiContent = append(multiContent, openai.ChatMessagePart{
					Type: openai.ChatMessagePartTypeText,
					Text: v.Text,
				})
			}
		case llmx.ImagePart:
			// If we have an image, we need to use multiContent
			if content != "" && len(multiContent) == 0 {
				multiContent = append(multiContent, openai.ChatMessagePart{
					Type: openai.ChatMessagePartTypeText,
					Text: content,
				})
				content = ""
			}
			imageURL := v.URL
			if v.Base64 != "" {
				imageURL = v.Base64
			}
			multiContent = append(multiContent, openai.ChatMessagePart{
				Type: openai.ChatMessagePartTypeImageURL,
				ImageURL: &openai.ChatMessageImageURL{
					URL:    imageURL,
					Detail: openai.ImageURLDetail(v.Detail),
				},
			})
		}
	}

	message := openai.ChatCompletionMessage{
		Role: string(msg.Role),
	}

	if len(multiContent) > 0 {
		message.MultiContent = multiContent
	} else {
		message.Content = content
	}

	return message
}

// convertResponse converts OpenAI response to llmx response
func (p *OpenAIProvider) convertResponse(resp *openai.ChatCompletionResponse) *llmx.ChatResponse {
	if len(resp.Choices) == 0 {
		return &llmx.ChatResponse{}
	}

	choice := resp.Choices[0]

	return &llmx.ChatResponse{
		ID:      resp.ID,
		Model:   resp.Model,
		Content: choice.Message.Content,
		Usage: llmx.Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
		FinishReason: string(choice.FinishReason),
		CreatedAt:    time.Unix(int64(resp.Created), 0),
		Raw:          resp,
	}
}

// convertError converts OpenAI errors to llmx errors
func (p *OpenAIProvider) convertError(err error) error {
	if apiErr, ok := err.(*openai.APIError); ok {
		switch apiErr.HTTPStatusCode {
		case 401:
			return llmx.NewAuthenticationError(apiErr.Message)
		case 429:
			return llmx.NewRateLimitError(apiErr.Message, 60*time.Second)
		case 400:
			return llmx.NewInvalidRequestError(apiErr.Message, nil)
		case 404:
			return llmx.NewNotFoundError(apiErr.Message, "model")
		default:
			return llmx.NewProviderError("openai", apiErr.Message, apiErr.HTTPStatusCode, err)
		}
	}
	return llmx.NewInternalError("OpenAI API error", err)
}
