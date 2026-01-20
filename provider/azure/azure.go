package azure

import (
	"context"
	"fmt"
	"time"

	"github.com/llmx-ai/llmx"
	"github.com/llmx-ai/llmx/provider"
	openai "github.com/sashabaranov/go-openai"
)

// AzureProvider implements the Provider interface for Azure OpenAI
type AzureProvider struct {
	client *openai.Client
}

func init() {
	provider.Register("azure", NewAzureProvider)
	provider.Register("azure-openai", NewAzureProvider) // Alias
}

// NewAzureProvider creates a new Azure OpenAI provider
func NewAzureProvider(opts map[string]interface{}) (provider.Provider, error) {
	apiKey, hasKey := opts["api_key"].(string)
	endpoint, hasEndpoint := opts["endpoint"].(string)

	if !hasKey || apiKey == "" {
		return nil, fmt.Errorf("api_key is required for Azure provider")
	}

	if !hasEndpoint || endpoint == "" {
		return nil, fmt.Errorf("endpoint is required for Azure provider")
	}

	// Azure OpenAI configuration
	config := openai.DefaultAzureConfig(apiKey, endpoint)

	// Optional: API version
	if apiVersion, ok := opts["api_version"].(string); ok {
		config.APIVersion = apiVersion
	} else {
		config.APIVersion = "2024-02-15-preview" // Default version
	}

	return &AzureProvider{
		client: openai.NewClientWithConfig(config),
	}, nil
}

// Name returns the provider name
func (p *AzureProvider) Name() string {
	return "azure"
}

// Chat sends a chat request
func (p *AzureProvider) Chat(ctx context.Context, reqInterface interface{}) (interface{}, error) {
	req, ok := reqInterface.(*llmx.ChatRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type")
	}

	// Convert request (same as OpenAI)
	openaiReq := p.convertRequest(req)

	// Call Azure OpenAI API
	resp, err := p.client.CreateChatCompletion(ctx, openaiReq)
	if err != nil {
		return nil, p.convertError(err)
	}

	// Convert response (same as OpenAI)
	return p.convertResponse(&resp), nil
}

// StreamChat sends a streaming chat request
func (p *AzureProvider) StreamChat(ctx context.Context, reqInterface interface{}) (interface{}, error) {
	req, ok := reqInterface.(*llmx.ChatRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type")
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

	// Start goroutine to handle streaming (reuse OpenAI logic)
	go p.handleStream(ctx, stream, chatStream)

	return chatStream, nil
}

// SupportedFeatures returns supported features
func (p *AzureProvider) SupportedFeatures() provider.Features {
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
func (p *AzureProvider) SupportedModels() []provider.Model {
	return []provider.Model{
		{
			ID:              "gpt-4-turbo",
			Name:            "GPT-4 Turbo (Azure)",
			ContextWindow:   128000,
			MaxOutputTokens: 4096,
			InputCost:       10.0,
			OutputCost:      30.0,
			Capabilities:    []string{"chat", "vision", "tools"},
		},
		{
			ID:              "gpt-4",
			Name:            "GPT-4 (Azure)",
			ContextWindow:   8192,
			MaxOutputTokens: 4096,
			InputCost:       30.0,
			OutputCost:      60.0,
			Capabilities:    []string{"chat", "tools"},
		},
		{
			ID:              "gpt-35-turbo",
			Name:            "GPT-3.5 Turbo (Azure)",
			ContextWindow:   16385,
			MaxOutputTokens: 4096,
			InputCost:       0.5,
			OutputCost:      1.5,
			Capabilities:    []string{"chat", "tools"},
		},
	}
}

// convertRequest converts llmx request to OpenAI request
// (Identical to OpenAI provider)
func (p *AzureProvider) convertRequest(req *llmx.ChatRequest) openai.ChatCompletionRequest {
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
func (p *AzureProvider) convertMessage(msg llmx.Message) openai.ChatCompletionMessage {
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
func (p *AzureProvider) convertResponse(resp *openai.ChatCompletionResponse) *llmx.ChatResponse {
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
func (p *AzureProvider) convertError(err error) error {
	if apiErr, ok := err.(*openai.APIError); ok {
		switch apiErr.HTTPStatusCode {
		case 401:
			return llmx.NewAuthenticationError(apiErr.Message)
		case 429:
			return llmx.NewRateLimitError(apiErr.Message, 60*time.Second)
		case 400:
			return llmx.NewInvalidRequestError(apiErr.Message, nil)
		case 404:
			return llmx.NewNotFoundError(apiErr.Message, "deployment")
		default:
			return llmx.NewProviderError("azure", apiErr.Message, apiErr.HTTPStatusCode, err)
		}
	}
	return llmx.NewInternalError("Azure OpenAI API error", err)
}
