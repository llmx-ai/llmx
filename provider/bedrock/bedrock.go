// Package bedrock provides an Amazon Bedrock provider implementation for llmx.
// Bedrock offers access to multiple foundation models (Claude, Llama, Titan, etc.) on AWS.
package bedrock

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/llmx-ai/llmx"
	"github.com/llmx-ai/llmx/core"
	"github.com/llmx-ai/llmx/provider"
)

// BedrockProvider implements the Provider interface for Amazon Bedrock
type BedrockProvider struct {
	client *bedrockruntime.Client
	region string
}

func init() {
	provider.Register("bedrock", NewBedrockProvider)
}

// NewBedrockProvider creates a new Bedrock provider
func NewBedrockProvider(opts map[string]interface{}) (provider.Provider, error) {
	region, ok := opts["region"].(string)
	if !ok || region == "" {
		region = "us-east-1" // Default region
	}

	var cfg aws.Config
	var err error

	// Check if credentials are provided
	accessKeyID, hasAccessKey := opts["access_key_id"].(string)
	secretAccessKey, hasSecretKey := opts["secret_access_key"].(string)

	if hasAccessKey && hasSecretKey && accessKeyID != "" && secretAccessKey != "" {
		// Use provided credentials
		cfg, err = config.LoadDefaultConfig(context.Background(),
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				accessKeyID,
				secretAccessKey,
				"",
			)),
		)
	} else {
		// Use default credential chain (环境变量, ~/.aws/credentials, IAM role等)
		cfg, err = config.LoadDefaultConfig(context.Background(),
			config.WithRegion(region),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("bedrock: failed to load AWS config: %w", err)
	}

	client := bedrockruntime.NewFromConfig(cfg)

	return &BedrockProvider{
		client: client,
		region: region,
	}, nil
}

// Name returns the provider name
func (p *BedrockProvider) Name() string {
	return "bedrock"
}

// SupportedModels returns the list of models supported by Bedrock
func (p *BedrockProvider) SupportedModels() []provider.Model {
	return []provider.Model{
		{
			ID:              "anthropic.claude-3-5-sonnet-20241022-v2:0",
			Name:            "Claude 3.5 Sonnet v2",
			ContextWindow:   200000,
			MaxOutputTokens: 8192,
			InputCost:       3.0,
			OutputCost:      15.0,
		},
		{
			ID:              "anthropic.claude-3-5-sonnet-20240620-v1:0",
			Name:            "Claude 3.5 Sonnet",
			ContextWindow:   200000,
			MaxOutputTokens: 8192,
			InputCost:       3.0,
			OutputCost:      15.0,
		},
		{
			ID:              "anthropic.claude-3-opus-20240229-v1:0",
			Name:            "Claude 3 Opus",
			ContextWindow:   200000,
			MaxOutputTokens: 4096,
			InputCost:       15.0,
			OutputCost:      75.0,
		},
		{
			ID:              "anthropic.claude-3-sonnet-20240229-v1:0",
			Name:            "Claude 3 Sonnet",
			ContextWindow:   200000,
			MaxOutputTokens: 4096,
			InputCost:       3.0,
			OutputCost:      15.0,
		},
		{
			ID:              "anthropic.claude-3-haiku-20240307-v1:0",
			Name:            "Claude 3 Haiku",
			ContextWindow:   200000,
			MaxOutputTokens: 4096,
			InputCost:       0.25,
			OutputCost:      1.25,
		},
		{
			ID:              "meta.llama3-1-70b-instruct-v1:0",
			Name:            "Llama 3.1 70B",
			ContextWindow:   131072,
			MaxOutputTokens: 4096,
			InputCost:       0.99,
			OutputCost:      0.99,
		},
		{
			ID:              "meta.llama3-1-8b-instruct-v1:0",
			Name:            "Llama 3.1 8B",
			ContextWindow:   131072,
			MaxOutputTokens: 4096,
			InputCost:       0.22,
			OutputCost:      0.22,
		},
	}
}

// SupportedFeatures returns the features supported by Bedrock
func (p *BedrockProvider) SupportedFeatures() provider.Features {
	return provider.Features{
		Streaming:   true,
		ToolCalling: true,  // Claude models support tools
		Vision:      true,  // Claude 3 models support vision
		JSONMode:    false, // Not standardized across models
	}
}

// Chat sends a chat request to Bedrock
func (p *BedrockProvider) Chat(ctx context.Context, req interface{}) (interface{}, error) {
	chatReq, ok := req.(*llmx.ChatRequest)
	if !ok {
		return nil, fmt.Errorf("bedrock: invalid request type")
	}

	// Determine model family and convert request accordingly
	if isClaudeModel(chatReq.Model) {
		return p.chatClaude(ctx, chatReq)
	} else if isLlamaModel(chatReq.Model) {
		return p.chatLlama(ctx, chatReq)
	}

	return nil, fmt.Errorf("bedrock: unsupported model %s", chatReq.Model)
}

// StreamChat sends a streaming chat request to Bedrock
func (p *BedrockProvider) StreamChat(ctx context.Context, req interface{}) (interface{}, error) {
	chatReq, ok := req.(*llmx.ChatRequest)
	if !ok {
		return nil, fmt.Errorf("bedrock: invalid request type")
	}

	// Determine model family
	if isClaudeModel(chatReq.Model) {
		return p.streamChatClaude(ctx, chatReq)
	} else if isLlamaModel(chatReq.Model) {
		return p.streamChatLlama(ctx, chatReq)
	}

	return nil, fmt.Errorf("bedrock: unsupported model %s", chatReq.Model)
}

// chatClaude handles Claude-specific chat requests
func (p *BedrockProvider) chatClaude(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
	// Build Claude request
	claudeReq := map[string]interface{}{
		"anthropic_version": "bedrock-2023-05-31",
		"max_tokens":        4096,
	}

	// Extract system prompt and messages
	var systemPrompt string
	var messages []map[string]interface{}

	for _, msg := range req.Messages {
		var content string
		if len(msg.Content) > 0 {
			if textPart, ok := msg.Content[0].(*llmx.TextPart); ok {
				content = textPart.Text
			}
		}
		
		if msg.Role == llmx.RoleSystem {
			systemPrompt = content
		} else {
			messages = append(messages, map[string]interface{}{
				"role":    convertRoleForClaude(msg.Role),
				"content": content,
			})
		}
	}

	claudeReq["messages"] = messages
	if systemPrompt != "" {
		claudeReq["system"] = systemPrompt
	}

	// Set optional parameters
	if req.Temperature != nil {
		claudeReq["temperature"] = *req.Temperature
	}
	if req.MaxTokens != nil {
		claudeReq["max_tokens"] = *req.MaxTokens
	}
	if req.TopP != nil {
		claudeReq["top_p"] = *req.TopP
	}

	// Convert tools if present
	if len(req.Tools) > 0 {
		claudeReq["tools"] = convertToolsForClaude(req.Tools)
	}

	// Marshal request
	requestBody, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, fmt.Errorf("bedrock: failed to marshal request: %w", err)
	}

	// Call Bedrock API
	output, err := p.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(req.Model),
		Body:        requestBody,
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return nil, p.convertError(err)
	}

	// Parse response
	var claudeResp map[string]interface{}
	if err := json.Unmarshal(output.Body, &claudeResp); err != nil {
		return nil, fmt.Errorf("bedrock: failed to unmarshal response: %w", err)
	}

	// Convert to llmx response
	return p.convertClaudeResponse(claudeResp, req.Model), nil
}

// streamChatClaude handles Claude-specific streaming requests
func (p *BedrockProvider) streamChatClaude(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatStream, error) {
	// Build Claude request (same as non-streaming)
	claudeReq := map[string]interface{}{
		"anthropic_version": "bedrock-2023-05-31",
		"max_tokens":        4096,
	}

	var systemPrompt string
	var messages []map[string]interface{}

	for _, msg := range req.Messages {
		var content string
		if len(msg.Content) > 0 {
			if textPart, ok := msg.Content[0].(*llmx.TextPart); ok {
				content = textPart.Text
			}
		}
		
		if msg.Role == llmx.RoleSystem {
			systemPrompt = content
		} else {
			messages = append(messages, map[string]interface{}{
				"role":    convertRoleForClaude(msg.Role),
				"content": content,
			})
		}
	}

	claudeReq["messages"] = messages
	if systemPrompt != "" {
		claudeReq["system"] = systemPrompt
	}

	if req.Temperature != nil {
		claudeReq["temperature"] = *req.Temperature
	}
	if req.MaxTokens != nil {
		claudeReq["max_tokens"] = *req.MaxTokens
	}

	requestBody, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, fmt.Errorf("bedrock: failed to marshal request: %w", err)
	}

	// Call streaming API
	output, err := p.client.InvokeModelWithResponseStream(ctx, &bedrockruntime.InvokeModelWithResponseStreamInput{
		ModelId:     aws.String(req.Model),
		Body:        requestBody,
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return nil, p.convertError(err)
	}

	// Create llmx stream
	chatStream := llmx.NewChatStream(ctx)

	// Start goroutine to handle streaming
	go p.handleClaudeStream(ctx, output, chatStream)

	return chatStream, nil
}

// chatLlama handles Llama-specific chat requests (simplified)
func (p *BedrockProvider) chatLlama(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
	// Build Llama request (Meta format)
	llamaReq := map[string]interface{}{
		"prompt":      buildLlamaPrompt(req.Messages),
		"max_gen_len": 2048,
	}

	if req.Temperature != nil {
		llamaReq["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		llamaReq["top_p"] = *req.TopP
	}

	requestBody, err := json.Marshal(llamaReq)
	if err != nil {
		return nil, fmt.Errorf("bedrock: failed to marshal request: %w", err)
	}

	output, err := p.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(req.Model),
		Body:        requestBody,
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return nil, p.convertError(err)
	}

	var llamaResp map[string]interface{}
	if err := json.Unmarshal(output.Body, &llamaResp); err != nil {
		return nil, fmt.Errorf("bedrock: failed to unmarshal response: %w", err)
	}

	return p.convertLlamaResponse(llamaResp, req.Model), nil
}

// streamChatLlama handles Llama streaming (simplified placeholder)
func (p *BedrockProvider) streamChatLlama(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatStream, error) {
	// Llama streaming is similar to Claude but with different event format
	// For now, return non-streaming response wrapped in stream
	resp, err := p.chatLlama(ctx, req)
	if err != nil {
		return nil, err
	}

	chatStream := llmx.NewChatStream(ctx)
	go func() {
		defer chatStream.Close()
		chatStream.SendEvent(core.StreamEvent{
			Type: core.EventTypeStart,
		})
		if resp.Content != "" {
			chatStream.SendEvent(core.StreamEvent{
				Type: core.EventTypeTextDelta,
				Data: resp.Content,
			})
		}
		chatStream.SendEvent(core.StreamEvent{
			Type: core.EventTypeFinish,
		})
	}()

	return chatStream, nil
}

// handleClaudeStream processes Claude streaming responses
func (p *BedrockProvider) handleClaudeStream(ctx context.Context, output *bedrockruntime.InvokeModelWithResponseStreamOutput, chatStream *llmx.ChatStream) {
	defer chatStream.Close()

	chatStream.SendEvent(core.StreamEvent{
		Type: core.EventTypeStart,
	})

	stream := output.GetStream()
	eventStream := stream.Events()
	
	for {
		select {
		case <-ctx.Done():
			chatStream.SendError(ctx.Err())
			return
		case event, ok := <-eventStream:
			if !ok {
				// Channel closed
				chatStream.SendEvent(core.StreamEvent{
					Type: core.EventTypeFinish,
				})
				return
			}
			
			// Handle payload chunk - simplified type handling
			chunkBytes, ok := extractEventBytes(event)
			if ok {
				// Parse chunk
				var chunkData map[string]interface{}
				if err := json.Unmarshal(chunkBytes, &chunkData); err != nil {
					continue
				}

				// Handle different chunk types
				if chunkType, ok := chunkData["type"].(string); ok {
					switch chunkType {
					case "content_block_delta":
						if delta, ok := chunkData["delta"].(map[string]interface{}); ok {
							if text, ok := delta["text"].(string); ok {
								chatStream.SendEvent(core.StreamEvent{
									Type: core.EventTypeTextDelta,
									Data: text,
								})
							}
						}
					case "message_stop":
						chatStream.SendEvent(core.StreamEvent{
							Type: core.EventTypeFinish,
							Data: "stop",
						})
						return
					}
				}
			}
		}
	}

	if err := stream.Err(); err != nil {
		chatStream.SendError(p.convertError(err))
		return
	}

	chatStream.SendEvent(core.StreamEvent{
		Type: core.EventTypeFinish,
	})
}

// Helper functions

func isClaudeModel(model string) bool {
	return len(model) >= 10 && model[:10] == "anthropic."
}

func isLlamaModel(model string) bool {
	return len(model) >= 5 && model[:5] == "meta."
}

func extractTextContent(content []llmx.ContentPart) string {
	// For simplified response structure, we don't use ContentPart
	// This function is kept for compatibility but not used
	return ""
}

func convertRoleForClaude(role llmx.MessageRole) string {
	if role == llmx.RoleAssistant {
		return "assistant"
	}
	return "user"
}

// convertContentForClaude is no longer used - simplified to string content

func convertToolsForClaude(tools []llmx.Tool) []map[string]interface{}  {
	var claudeTools []map[string]interface{}
	for _, tool := range tools {
		claudeTools = append(claudeTools, map[string]interface{}{
			"name":         tool.Name,
			"description":  tool.Description,
			"input_schema": tool.Parameters,
		})
	}
	return claudeTools
}

func (p *BedrockProvider) convertClaudeResponse(resp map[string]interface{}, model string) *llmx.ChatResponse {
	llmxResp := &llmx.ChatResponse{
		Model:        model,
		FinishReason: getFinishReason(resp),
	}

	// Extract text content
	if contentArray, ok := resp["content"].([]interface{}); ok {
		for _, item := range contentArray {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if text, ok := itemMap["text"].(string); ok {
					llmxResp.Content += text
				}
			}
		}
	}

	// Set usage if available
	if usage, ok := resp["usage"].(map[string]interface{}); ok {
		if inputTokens, ok := usage["input_tokens"].(float64); ok {
			llmxResp.Usage.PromptTokens = int(inputTokens)
		}
		if outputTokens, ok := usage["output_tokens"].(float64); ok {
			llmxResp.Usage.CompletionTokens = int(outputTokens)
			llmxResp.Usage.TotalTokens = llmxResp.Usage.PromptTokens + int(outputTokens)
		}
	}

	return llmxResp
}

func (p *BedrockProvider) convertLlamaResponse(resp map[string]interface{}, model string) *llmx.ChatResponse {
	llmxResp := &llmx.ChatResponse{
		Model:        model,
		FinishReason: "stop",
	}

	if generation, ok := resp["generation"].(string); ok {
		llmxResp.Content = generation
	}

	return llmxResp
}

func buildLlamaPrompt(messages []llmx.Message) string {
	// Simple prompt building for Llama
	var prompt string
	for _, msg := range messages {
		if len(msg.Content) == 0 {
			continue
		}
		// Extract text from first content part
		var text string
		if textPart, ok := msg.Content[0].(*llmx.TextPart); ok {
			text = textPart.Text
		}
		
		if msg.Role == llmx.RoleSystem {
			prompt += "[INST] <<SYS>>\n" + text + "\n<</SYS>>\n\n"
		} else if msg.Role == llmx.RoleUser {
			prompt += text + " [/INST] "
		} else {
			prompt += text + " "
		}
	}
	return prompt
}

func getFinishReason(resp map[string]interface{}) string {
	if stopReason, ok := resp["stop_reason"].(string); ok {
		switch stopReason {
		case "end_turn":
			return "stop"
		case "max_tokens":
			return "length"
		case "stop_sequence":
			return "stop"
		default:
			return stopReason
		}
	}
	return "stop"
}

func (p *BedrockProvider) convertError(err error) error {
	errMsg := err.Error()

	if contains(errMsg, "throttling") || contains(errMsg, "rate") {
		return llmx.NewRateLimitError(errMsg, 0)
	} else if contains(errMsg, "unauthorized") || contains(errMsg, "access denied") {
		return llmx.NewAuthenticationError(errMsg)
	} else if contains(errMsg, "not found") {
		return llmx.NewNotFoundError(errMsg, "model")
	} else if contains(errMsg, "invalid") || contains(errMsg, "validation") {
		return llmx.NewInvalidRequestError(errMsg, nil)
	}

	return llmx.NewInternalError(errMsg, err)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr)+1 && s[1:len(substr)+1] == substr))
}

// extractEventBytes extracts bytes from Bedrock stream event
func extractEventBytes(event interface{}) ([]byte, bool) {
	// Use reflection to extract bytes from event
	// This is a workaround for the complex event types in AWS SDK
	type hasBytes interface {
		GetBytes() []byte
	}
	
	if v, ok := event.(hasBytes); ok {
		return v.GetBytes(), true
	}
	
	return nil, false
}
