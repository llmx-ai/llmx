// Package wenxin provides a Baidu Wenxin (ERNIE) provider implementation for llmx.
// Wenxin is Baidu's large language model service.
package wenxin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/llmx-ai/llmx"
	"github.com/llmx-ai/llmx/core"
	"github.com/llmx-ai/llmx/provider"
)

const (
	// OAuth 2.0 token endpoint
	tokenURL = "https://aip.baidubce.com/oauth/2.0/token"
	// API base URL
	baseURL = "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat"
)

// WenxinProvider implements the Provider interface for Baidu Wenxin
type WenxinProvider struct {
	apiKey      string
	secretKey   string
	accessToken string
	tokenExpiry time.Time
	tokenMutex  sync.RWMutex
	httpClient  *http.Client
}

func init() {
	provider.Register("wenxin", NewWenxinProvider)
}

// NewWenxinProvider creates a new Wenxin provider
func NewWenxinProvider(opts map[string]interface{}) (provider.Provider, error) {
	apiKey, ok := opts["api_key"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("wenxin: api_key is required")
	}

	secretKey, ok := opts["secret_key"].(string)
	if !ok || secretKey == "" {
		return nil, fmt.Errorf("wenxin: secret_key is required")
	}

	return &WenxinProvider{
		apiKey:     apiKey,
		secretKey:  secretKey,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}, nil
}

// Name returns the provider name
func (p *WenxinProvider) Name() string {
	return "wenxin"
}

// SupportedModels returns the list of models supported by Wenxin
func (p *WenxinProvider) SupportedModels() []provider.Model {
	return []provider.Model{
		{
			ID:              "ernie-4.0-8k",
			Name:            "ERNIE 4.0",
			ContextWindow:   8192,
			MaxOutputTokens: 2048,
			InputCost:       16.8, // CNY 120/1M tokens
			OutputCost:      16.8,
		},
		{
			ID:              "ernie-3.5-8k",
			Name:            "ERNIE 3.5",
			ContextWindow:   8192,
			MaxOutputTokens: 2048,
			InputCost:       1.68, // CNY 12/1M tokens
			OutputCost:      1.68,
		},
		{
			ID:              "ernie-speed-8k",
			Name:            "ERNIE Speed",
			ContextWindow:   8192,
			MaxOutputTokens: 2048,
			InputCost:       0.56, // CNY 4/1M tokens
			OutputCost:      0.56,
		},
		{
			ID:              "ernie-lite-8k",
			Name:            "ERNIE Lite",
			ContextWindow:   8192,
			MaxOutputTokens: 2048,
			InputCost:       0.07, // CNY 0.5/1M tokens
			OutputCost:      0.07,
		},
	}
}

// SupportedFeatures returns the features supported by Wenxin
func (p *WenxinProvider) SupportedFeatures() provider.Features {
	return provider.Features{
		Streaming:   true,
		ToolCalling: true,
		Vision:      false,
		JSONMode:    true,
	}
}

// Chat sends a chat request to Wenxin
func (p *WenxinProvider) Chat(ctx context.Context, req interface{}) (interface{}, error) {
	chatReq, ok := req.(*llmx.ChatRequest)
	if !ok {
		return nil, fmt.Errorf("wenxin: invalid request type")
	}

	// Get access token
	token, err := p.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	// Build endpoint URL
	endpoint := p.getEndpoint(chatReq.Model)
	fullURL := fmt.Sprintf("%s?access_token=%s", endpoint, token)

	// Convert request
	wenxinReq := p.convertRequest(chatReq)
	wenxinReq["stream"] = false

	// Marshal request
	requestBody, err := json.Marshal(wenxinReq)
	if err != nil {
		return nil, fmt.Errorf("wenxin: failed to marshal request: %w", err)
	}

	// Send request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", fullURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("wenxin: failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("wenxin: request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("wenxin: failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return nil, p.convertError(body, resp.StatusCode)
	}

	// Parse response
	var wenxinResp map[string]interface{}
	if err := json.Unmarshal(body, &wenxinResp); err != nil {
		return nil, fmt.Errorf("wenxin: failed to unmarshal response: %w", err)
	}

	// Check for API errors
	if errCode, ok := wenxinResp["error_code"].(float64); ok && errCode != 0 {
		return nil, p.convertAPIError(wenxinResp)
	}

	// Convert to llmx response
	return p.convertResponse(wenxinResp, chatReq.Model), nil
}

// StreamChat sends a streaming chat request to Wenxin
func (p *WenxinProvider) StreamChat(ctx context.Context, req interface{}) (interface{}, error) {
	chatReq, ok := req.(*llmx.ChatRequest)
	if !ok {
		return nil, fmt.Errorf("wenxin: invalid request type")
	}

	// Get access token
	token, err := p.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	// Build endpoint URL
	endpoint := p.getEndpoint(chatReq.Model)
	fullURL := fmt.Sprintf("%s?access_token=%s", endpoint, token)

	// Convert request
	wenxinReq := p.convertRequest(chatReq)
	wenxinReq["stream"] = true

	// Marshal request
	requestBody, err := json.Marshal(wenxinReq)
	if err != nil {
		return nil, fmt.Errorf("wenxin: failed to marshal request: %w", err)
	}

	// Send request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", fullURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("wenxin: failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("wenxin: request failed: %w", err)
	}

	// Create llmx stream
	chatStream := llmx.NewChatStream(ctx)

	// Start goroutine to handle streaming
	go p.handleStream(ctx, resp, chatStream)

	return chatStream, nil
}

// getAccessToken retrieves or refreshes the access token
func (p *WenxinProvider) getAccessToken(ctx context.Context) (string, error) {
	// Check if we have a valid cached token
	p.tokenMutex.RLock()
	if p.accessToken != "" && time.Now().Before(p.tokenExpiry) {
		token := p.accessToken
		p.tokenMutex.RUnlock()
		return token, nil
	}
	p.tokenMutex.RUnlock()

	// Need to refresh token
	p.tokenMutex.Lock()
	defer p.tokenMutex.Unlock()

	// Double-check after acquiring write lock
	if p.accessToken != "" && time.Now().Before(p.tokenExpiry) {
		return p.accessToken, nil
	}

	// Request new token
	params := url.Values{}
	params.Set("grant_type", "client_credentials")
	params.Set("client_id", p.apiKey)
	params.Set("client_secret", p.secretKey)

	fullURL := fmt.Sprintf("%s?%s", tokenURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "POST", fullURL, nil)
	if err != nil {
		return "", fmt.Errorf("wenxin: failed to create token request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("wenxin: token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("wenxin: failed to read token response: %w", err)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("wenxin: failed to unmarshal token response: %w", err)
	}

	if tokenResp.Error != "" {
		return "", fmt.Errorf("wenxin: token error: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("wenxin: empty access token received")
	}

	// Cache token (expire 5 minutes early for safety)
	p.accessToken = tokenResp.AccessToken
	p.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second)

	return p.accessToken, nil
}

// getEndpoint returns the API endpoint for a given model
func (p *WenxinProvider) getEndpoint(model string) string {
	// Map model IDs to endpoints
	endpoints := map[string]string{
		"ernie-4.0-8k":    baseURL + "/completions_pro",
		"ernie-3.5-8k":    baseURL + "/completions",
		"ernie-speed-8k":  baseURL + "/ernie_speed",
		"ernie-lite-8k":   baseURL + "/eb-instant",
		"ernie-tiny-8k":   baseURL + "/ernie-tiny-8k",
	}

	if endpoint, ok := endpoints[model]; ok {
		return endpoint
	}

	// Default to ERNIE 3.5
	return baseURL + "/completions"
}

// convertRequest converts llmx.ChatRequest to Wenxin format
func (p *WenxinProvider) convertRequest(req *llmx.ChatRequest) map[string]interface{} {
	wenxinReq := make(map[string]interface{})

	// Convert messages
	var messages []map[string]interface{}
	for _, msg := range req.Messages {
		var content string
		if len(msg.Content) > 0 {
			if textPart, ok := msg.Content[0].(*llmx.TextPart); ok {
				content = textPart.Text
			}
		}
		messages = append(messages, map[string]interface{}{
			"role":    convertRole(msg.Role),
			"content": content,
		})
	}
	wenxinReq["messages"] = messages

	// Set optional parameters
	if req.Temperature != nil {
		wenxinReq["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		wenxinReq["top_p"] = *req.TopP
	}
	if req.MaxTokens != nil {
		wenxinReq["max_output_tokens"] = *req.MaxTokens
	}

	// Convert tools if present
	if len(req.Tools) > 0 {
		wenxinReq["functions"] = convertTools(req.Tools)
	}

	return wenxinReq
}

// convertResponse converts Wenxin response to llmx format
func (p *WenxinProvider) convertResponse(resp map[string]interface{}, model string) *llmx.ChatResponse {
	llmxResp := &llmx.ChatResponse{
		Model:        model,
		FinishReason: "stop",
	}

	// Extract result
	if result, ok := resp["result"].(string); ok {
		llmxResp.Content = result
	}

	// Set usage if available
	if usage, ok := resp["usage"].(map[string]interface{}); ok {
		if promptTokens, ok := usage["prompt_tokens"].(float64); ok {
			llmxResp.Usage.PromptTokens = int(promptTokens)
		}
		if completionTokens, ok := usage["completion_tokens"].(float64); ok {
			llmxResp.Usage.CompletionTokens = int(completionTokens)
		}
		if totalTokens, ok := usage["total_tokens"].(float64); ok {
			llmxResp.Usage.TotalTokens = int(totalTokens)
		}
	}

	return llmxResp
}

// handleStream processes Wenxin streaming responses
func (p *WenxinProvider) handleStream(ctx context.Context, resp *http.Response, chatStream *llmx.ChatStream) {
	defer resp.Body.Close()
	defer chatStream.Close()

	chatStream.SendEvent(core.StreamEvent{
		Type: core.EventTypeStart,
	})

	decoder := json.NewDecoder(resp.Body)
	for {
		select {
		case <-ctx.Done():
			chatStream.SendError(ctx.Err())
			return
		default:
			var chunk map[string]interface{}
			if err := decoder.Decode(&chunk); err != nil {
				if err != io.EOF {
					chatStream.SendError(fmt.Errorf("wenxin: stream decode error: %w", err))
				}
				chatStream.SendEvent(core.StreamEvent{
					Type: core.EventTypeFinish,
				})
				return
			}

			// Check for errors
			if errCode, ok := chunk["error_code"].(float64); ok && errCode != 0 {
				chatStream.SendError(p.convertAPIError(chunk))
				return
			}

			// Extract result
			if result, ok := chunk["result"].(string); ok && result != "" {
				chatStream.SendEvent(core.StreamEvent{
					Type: core.EventTypeTextDelta,
					Data: result,
				})
			}

			// Check if finished
			if isEnd, ok := chunk["is_end"].(bool); ok && isEnd {
				chatStream.SendEvent(core.StreamEvent{
					Type: core.EventTypeFinish,
					Data: "stop",
				})
				return
			}
		}
	}
}

// Helper functions

func extractTextContent(content []llmx.ContentPart) string {
	// For simplified response structure, we don't use ContentPart
	// This function is kept for compatibility but not used
	return ""
}

func convertRole(role llmx.MessageRole) string {
	switch role {
	case llmx.RoleUser:
		return "user"
	case llmx.RoleAssistant:
		return "assistant"
	case llmx.RoleSystem:
		return "user" // Wenxin doesn't have dedicated system role
	default:
		return "user"
	}
}

func convertTools(tools []llmx.Tool) []map[string]interface{} {
	var wenxinTools []map[string]interface{}
	for _, tool := range tools {
		wenxinTools = append(wenxinTools, map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"parameters":  tool.Parameters,
		})
	}
	return wenxinTools
}

func (p *WenxinProvider) convertError(body []byte, statusCode int) error {
	var errResp map[string]interface{}
	json.Unmarshal(body, &errResp)

	errMsg := fmt.Sprintf("wenxin: HTTP %d", statusCode)
	if msg, ok := errResp["error_msg"].(string); ok {
		errMsg = msg
	}

	if statusCode == 429 {
		return llmx.NewRateLimitError(errMsg, 0)
	} else if statusCode == 401 || statusCode == 403 {
		return llmx.NewAuthenticationError(errMsg)
	} else if statusCode == 404 {
		return llmx.NewNotFoundError(errMsg, "model")
	} else if statusCode >= 400 && statusCode < 500 {
		return llmx.NewInvalidRequestError(errMsg, nil)
	}

	return llmx.NewInternalError(errMsg, nil)
}

func (p *WenxinProvider) convertAPIError(resp map[string]interface{}) error {
	errCode := int(resp["error_code"].(float64))
	errMsg := ""
	if msg, ok := resp["error_msg"].(string); ok {
		errMsg = msg
	}

	fullMsg := fmt.Sprintf("wenxin API error %d: %s", errCode, errMsg)

	// Map error codes to llmx error types
	switch errCode {
	case 18, 19: // Rate limit errors
		return llmx.NewRateLimitError(fullMsg, 0)
	case 110, 111: // Auth errors
		return llmx.NewAuthenticationError(fullMsg)
	case 1, 2, 3, 4, 6: // Invalid request
		return llmx.NewInvalidRequestError(fullMsg, nil)
	default:
		return llmx.NewInternalError(fullMsg, nil)
	}
}
