// Package cohere provides a Cohere provider implementation for llmx.
// Cohere is known for its powerful embedding and RAG capabilities.
package cohere

import (
	"context"
	"encoding/json"
	"fmt"

	cohere "github.com/cohere-ai/cohere-go/v2"
	cohereclient "github.com/cohere-ai/cohere-go/v2/client"
	"github.com/llmx-ai/llmx"
	"github.com/llmx-ai/llmx/core"
	"github.com/llmx-ai/llmx/provider"
)

// CohereProvider implements the Provider interface for Cohere
type CohereProvider struct {
	client *cohereclient.Client
	apiKey string
}

func init() {
	provider.Register("cohere", NewCohereProvider)
}

// NewCohereProvider creates a new Cohere provider
func NewCohereProvider(opts map[string]interface{}) (provider.Provider, error) {
	apiKey, ok := opts["api_key"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("cohere: api_key is required")
	}

	client := cohereclient.NewClient(cohereclient.WithToken(apiKey))

	return &CohereProvider{
		client: client,
		apiKey: apiKey,
	}, nil
}

// Name returns the provider name
func (p *CohereProvider) Name() string {
	return "cohere"
}

// SupportedModels returns the list of models supported by Cohere
func (p *CohereProvider) SupportedModels() []provider.Model {
	return []provider.Model{
		{
			ID:              "command-r-plus",
			Name:            "Command R+",
			ContextWindow:   131072,
			MaxOutputTokens: 4096,
			InputCost:       3.0,
			OutputCost:      15.0,
		},
		{
			ID:              "command-r",
			Name:            "Command R",
			ContextWindow:   131072,
			MaxOutputTokens: 4096,
			InputCost:       0.5,
			OutputCost:      1.5,
		},
		{
			ID:              "command",
			Name:            "Command",
			ContextWindow:   4096,
			MaxOutputTokens: 4096,
			InputCost:       1.0,
			OutputCost:      2.0,
		},
		{
			ID:              "command-light",
			Name:            "Command Light",
			ContextWindow:   4096,
			MaxOutputTokens: 4096,
			InputCost:       0.3,
			OutputCost:      0.6,
		},
	}
}

// SupportedFeatures returns the features supported by Cohere
func (p *CohereProvider) SupportedFeatures() provider.Features {
	return provider.Features{
		Streaming:   true,
		ToolCalling: true,
		Vision:      false,
		JSONMode:    true,
	}
}

// Chat sends a chat request to Cohere
func (p *CohereProvider) Chat(ctx context.Context, req interface{}) (interface{}, error) {
	chatReq, ok := req.(*llmx.ChatRequest)
	if !ok {
		return nil, fmt.Errorf("cohere: invalid request type")
	}

	// Convert llmx request to Cohere format
	cohereReq := p.convertRequest(chatReq)

	// Call Cohere API
	resp, err := p.client.Chat(ctx, cohereReq)
	if err != nil {
		return nil, p.convertError(err)
	}

	// Convert Cohere response to llmx format
	return p.convertResponse(resp), nil
}

// StreamChat sends a streaming chat request to Cohere
func (p *CohereProvider) StreamChat(ctx context.Context, req interface{}) (interface{}, error) {
	chatReq, ok := req.(*llmx.ChatRequest)
	if !ok {
		return nil, fmt.Errorf("cohere: invalid request type")
	}

	// For now, use non-streaming and wrap in stream
	// Full streaming implementation requires more complex SDK handling
	resp, err := p.Chat(ctx, chatReq)
	if err != nil {
		return nil, err
	}

	chatResp := resp.(*llmx.ChatResponse)
	chatStream := llmx.NewChatStream(ctx)

	go func() {
		defer chatStream.Close()
		chatStream.SendEvent(core.StreamEvent{
			Type: core.EventTypeStart,
		})
		if chatResp.Content != "" {
			chatStream.SendEvent(core.StreamEvent{
				Type: core.EventTypeTextDelta,
				Data: chatResp.Content,
			})
		}
		chatStream.SendEvent(core.StreamEvent{
			Type: core.EventTypeFinish,
		})
	}()

	return chatStream, nil
}

// convertRequest converts llmx.ChatRequest to Cohere format
func (p *CohereProvider) convertRequest(req *llmx.ChatRequest) *cohere.ChatRequest {
	cohereReq := &cohere.ChatRequest{
		Model: cohere.String(req.Model),
	}

	// Extract system prompt (preamble in Cohere)
	var preamble string
	var chatHistory []*cohere.Message
	var message string

	for i, msg := range req.Messages {
		content := extractTextContent(msg.Content)

		if msg.Role == llmx.RoleSystem {
			// System messages become preamble
			if preamble == "" {
				preamble = content
			} else {
				preamble += "\n" + content
			}
		} else if i == len(req.Messages)-1 {
			// Last message is the current message
			message = content
		} else {
			// Other messages go into chat history
			cohereMsg := &cohere.Message{
				Role: convertRole(msg.Role),
			}

			// Cohere uses different field names for user/chatbot messages
			if msg.Role == llmx.RoleUser {
				cohereMsg.User = &cohere.ChatMessage{
					Message: content,
				}
			} else if msg.Role == llmx.RoleAssistant {
				cohereMsg.Chatbot = &cohere.ChatMessage{
					Message: content,
				}
			}

			chatHistory = append(chatHistory, cohereMsg)
		}
	}

	cohereReq.Message = message
	if preamble != "" {
		cohereReq.Preamble = cohere.String(preamble)
	}
	if len(chatHistory) > 0 {
		cohereReq.ChatHistory = chatHistory
	}

	// Set temperature if provided
	if req.Temperature != nil {
		cohereReq.Temperature = cohere.Float64(*req.Temperature)
	}

	// Set max_tokens if provided
	if req.MaxTokens != nil {
		cohereReq.MaxTokens = cohere.Int(*req.MaxTokens)
	}

	// Convert tools if provided
	if len(req.Tools) > 0 {
		cohereReq.Tools = p.convertTools(req.Tools)
	}

	return cohereReq
}

// convertTools converts llmx.Tool to Cohere format
func (p *CohereProvider) convertTools(tools []llmx.Tool) []*cohere.Tool {
	// Simplified tool conversion
	// Full implementation requires proper schema handling
	var cohereTools []*cohere.Tool

	for _, tool := range tools {
		cohereTool := &cohere.Tool{
			Name:        tool.Name,
			Description: tool.Description,
		}
		cohereTools = append(cohereTools, cohereTool)
	}

	return cohereTools
}

// convertResponse converts Cohere response to llmx format
func (p *CohereProvider) convertResponse(resp *cohere.NonStreamedChatResponse) *llmx.ChatResponse {
	llmxResp := &llmx.ChatResponse{
		ID:      safeString(resp.GenerationId),
		Model:   "", // Cohere doesn't return model in response
		Content: resp.Text,
		FinishReason: func() string {
			if resp.FinishReason != nil {
				return convertFinishReason(*resp.FinishReason)
			}
			return "stop"
		}(),
	}

	// Handle tool calls if present
	if len(resp.ToolCalls) > 0 {
		for _, toolCall := range resp.ToolCalls {
			argsJSON := "{}"
			if toolCall.Parameters != nil {
				if data, err := json.Marshal(toolCall.Parameters); err == nil {
					argsJSON = string(data)
				}
			}
			
			llmxResp.ToolCalls = append(llmxResp.ToolCalls, llmx.ToolCall{
				ID:        toolCall.Name, // Cohere doesn't have separate ID
				Name:      toolCall.Name,
				Arguments: json.RawMessage(argsJSON),
			})
		}
	}

	// Set usage information if available
	if resp.Meta != nil && resp.Meta.Tokens != nil {
		llmxResp.Usage = llmx.Usage{
			PromptTokens:     safeFloat64ToInt(resp.Meta.Tokens.InputTokens),
			CompletionTokens: safeFloat64ToInt(resp.Meta.Tokens.OutputTokens),
			TotalTokens:      safeFloat64ToInt(resp.Meta.Tokens.InputTokens) + safeFloat64ToInt(resp.Meta.Tokens.OutputTokens),
		}
	}

	return llmxResp
}

// Helper functions for pointer safety
func safeString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func safeInt(i *int) int {
	if i != nil {
		return *i
	}
	return 0
}

func safeFloat64ToInt(f *float64) int {
	if f != nil {
		return int(*f)
	}
	return 0
}

// handleStream is no longer used - simplified to non-streaming wrapper

// Helper functions

func extractTextContent(content []llmx.ContentPart) string {
	// Extract text from first content part
	if len(content) > 0 {
		if textPart, ok := content[0].(*llmx.TextPart); ok {
			return textPart.Text
		}
	}
	return ""
}

func convertRole(role llmx.MessageRole) string {
	switch role {
	case llmx.RoleUser:
		return "USER"
	case llmx.RoleAssistant:
		return "CHATBOT"
	case llmx.RoleSystem:
		return "SYSTEM"
	default:
		return "USER"
	}
}

func convertFinishReason(reason interface{}) string {
	// Handle both string and FinishReason type
	var reasonStr string
	switch v := reason.(type) {
	case string:
		reasonStr = v
	default:
		return "stop"
	}
	
	switch reasonStr {
	case "COMPLETE":
		return "stop"
	case "MAX_TOKENS":
		return "length"
	case "ERROR":
		return "error"
	default:
		return "stop"
	}
}

func (p *CohereProvider) convertError(err error) error {
	// Convert Cohere errors to llmx errors
	errMsg := err.Error()

	if contains(errMsg, "rate limit") {
		return llmx.NewRateLimitError(errMsg, 0)
	} else if contains(errMsg, "unauthorized") || contains(errMsg, "invalid api key") {
		return llmx.NewAuthenticationError(errMsg)
	} else if contains(errMsg, "not found") {
		return llmx.NewNotFoundError(errMsg, "model")
	} else if contains(errMsg, "invalid") || contains(errMsg, "bad request") {
		return llmx.NewInvalidRequestError(errMsg, nil)
	}

	return llmx.NewInternalError(errMsg, err)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		len(s) > len(substr)+1 && s[1:len(substr)+1] == substr))
}
