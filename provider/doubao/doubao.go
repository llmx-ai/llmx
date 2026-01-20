// Package doubao provides a ByteDance Doubao provider implementation for llmx.
// Doubao is ByteDance's large language model service (powered by Volcano Engine).
package doubao

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/llmx-ai/llmx"
	"github.com/llmx-ai/llmx/core"
	"github.com/llmx-ai/llmx/provider"
)

const (
	// API base URL
	baseURL = "https://open.volcengineapi.com"
	// Service name
	serviceName = "ml_maas"
	// API version
	apiVersion = "2024-01-01"
)

// DoubaoProvider implements the Provider interface for ByteDance Doubao
type DoubaoProvider struct {
	apiKey     string
	secretKey  string
	httpClient *http.Client
}

func init() {
	provider.Register("doubao", NewDoubaoProvider)
}

// NewDoubaoProvider creates a new Doubao provider
func NewDoubaoProvider(opts map[string]interface{}) (provider.Provider, error) {
	apiKey, ok := opts["api_key"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("doubao: api_key is required")
	}

	secretKey, ok := opts["secret_key"].(string)
	if !ok || secretKey == "" {
		return nil, fmt.Errorf("doubao: secret_key is required")
	}

	return &DoubaoProvider{
		apiKey:     apiKey,
		secretKey:  secretKey,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}, nil
}

// Name returns the provider name
func (p *DoubaoProvider) Name() string {
	return "doubao"
}

// SupportedModels returns the list of models supported by Doubao
func (p *DoubaoProvider) SupportedModels() []provider.Model {
	return []provider.Model{
		{
			ID:              "doubao-pro-32k",
			Name:            "Doubao Pro 32K",
			ContextWindow:   32768,
			MaxOutputTokens: 4096,
			InputCost:       0.56, // CNY 4/1M tokens
			OutputCost:      2.1,  // CNY 15/1M tokens
		},
		{
			ID:              "doubao-lite-32k",
			Name:            "Doubao Lite 32K",
			ContextWindow:   32768,
			MaxOutputTokens: 4096,
			InputCost:       0.07, // CNY 0.5/1M tokens
			OutputCost:      0.21, // CNY 1.5/1M tokens
		},
		{
			ID:              "doubao-pro-4k",
			Name:            "Doubao Pro 4K",
			ContextWindow:   4096,
			MaxOutputTokens: 4096,
			InputCost:       0.14, // CNY 1/1M tokens
			OutputCost:      0.28, // CNY 2/1M tokens
		},
	}
}

// SupportedFeatures returns the features supported by Doubao
func (p *DoubaoProvider) SupportedFeatures() provider.Features {
	return provider.Features{
		Streaming:   true,
		ToolCalling: true,
		Vision:      false,
		JSONMode:    true,
	}
}

// Chat sends a chat request to Doubao
func (p *DoubaoProvider) Chat(ctx context.Context, req interface{}) (interface{}, error) {
	chatReq, ok := req.(*llmx.ChatRequest)
	if !ok {
		return nil, fmt.Errorf("doubao: invalid request type")
	}

	// Convert request
	doubaoReq := p.convertRequest(chatReq)
	doubaoReq["stream"] = false

	// Marshal request
	requestBody, err := json.Marshal(doubaoReq)
	if err != nil {
		return nil, fmt.Errorf("doubao: failed to marshal request: %w", err)
	}

	// Build request
	endpoint := fmt.Sprintf("%s/%s/chat", baseURL, apiVersion)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("doubao: failed to create request: %w", err)
	}

	// Sign request
	if err := p.signRequest(httpReq, requestBody); err != nil {
		return nil, fmt.Errorf("doubao: failed to sign request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("doubao: request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("doubao: failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return nil, p.convertError(body, resp.StatusCode)
	}

	// Parse response
	var doubaoResp map[string]interface{}
	if err := json.Unmarshal(body, &doubaoResp); err != nil {
		return nil, fmt.Errorf("doubao: failed to unmarshal response: %w", err)
	}

	// Check for API errors
	if errCode, ok := doubaoResp["error_code"].(string); ok && errCode != "" {
		return nil, p.convertAPIError(doubaoResp)
	}

	// Convert to llmx response
	return p.convertResponse(doubaoResp, chatReq.Model), nil
}

// StreamChat sends a streaming chat request to Doubao
func (p *DoubaoProvider) StreamChat(ctx context.Context, req interface{}) (interface{}, error) {
	chatReq, ok := req.(*llmx.ChatRequest)
	if !ok {
		return nil, fmt.Errorf("doubao: invalid request type")
	}

	// Convert request
	doubaoReq := p.convertRequest(chatReq)
	doubaoReq["stream"] = true

	// Marshal request
	requestBody, err := json.Marshal(doubaoReq)
	if err != nil {
		return nil, fmt.Errorf("doubao: failed to marshal request: %w", err)
	}

	// Build request
	endpoint := fmt.Sprintf("%s/%s/chat", baseURL, apiVersion)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("doubao: failed to create request: %w", err)
	}

	// Sign request
	if err := p.signRequest(httpReq, requestBody); err != nil {
		return nil, fmt.Errorf("doubao: failed to sign request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("doubao: request failed: %w", err)
	}

	// Create llmx stream
	chatStream := llmx.NewChatStream(ctx)

	// Start goroutine to handle streaming
	go p.handleStream(ctx, resp, chatStream)

	return chatStream, nil
}

// signRequest signs the HTTP request using Volcano Engine signature algorithm
func (p *DoubaoProvider) signRequest(req *http.Request, body []byte) error {
	// This is a simplified version of Volcano Engine's signature
	// In production, you should use the official SDK or implement the full algorithm

	timestamp := time.Now().UTC().Format("20060102T150405Z")
	date := timestamp[:8]

	// Set required headers
	req.Header.Set("X-Date", timestamp)
	req.Header.Set("X-Content-Sha256", hashSHA256(body))

	// Build canonical request
	canonicalHeaders := p.buildCanonicalHeaders(req)
	signedHeaders := p.getSignedHeaders(req)

	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		req.Method,
		req.URL.Path,
		req.URL.RawQuery,
		canonicalHeaders,
		signedHeaders,
		hashSHA256(body),
	)

	// Build string to sign
	credentialScope := fmt.Sprintf("%s/%s/%s/request", date, "cn-north-1", serviceName)
	stringToSign := fmt.Sprintf("HMAC-SHA256\n%s\n%s\n%s",
		timestamp,
		credentialScope,
		hashSHA256([]byte(canonicalRequest)),
	)

	// Calculate signature
	signature := p.calculateSignature(stringToSign, date)

	// Build authorization header
	authorization := fmt.Sprintf("HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		p.apiKey,
		credentialScope,
		signedHeaders,
		signature,
	)

	req.Header.Set("Authorization", authorization)

	return nil
}

// Helper functions for signing

func (p *DoubaoProvider) buildCanonicalHeaders(req *http.Request) string {
	var headers []string
	for k, v := range req.Header {
		if shouldSignHeader(k) {
			headers = append(headers, fmt.Sprintf("%s:%s", strings.ToLower(k), strings.TrimSpace(v[0])))
		}
	}
	sort.Strings(headers)
	return strings.Join(headers, "\n") + "\n"
}

func (p *DoubaoProvider) getSignedHeaders(req *http.Request) string {
	var headers []string
	for k := range req.Header {
		if shouldSignHeader(k) {
			headers = append(headers, strings.ToLower(k))
		}
	}
	sort.Strings(headers)
	return strings.Join(headers, ";")
}

func (p *DoubaoProvider) calculateSignature(stringToSign, date string) string {
	kDate := hmacSHA256([]byte(p.secretKey), []byte(date))
	kRegion := hmacSHA256(kDate, []byte("cn-north-1"))
	kService := hmacSHA256(kRegion, []byte(serviceName))
	kSigning := hmacSHA256(kService, []byte("request"))
	signature := hmacSHA256(kSigning, []byte(stringToSign))
	return hex.EncodeToString(signature)
}

func shouldSignHeader(header string) bool {
	lower := strings.ToLower(header)
	return lower == "content-type" || lower == "host" || strings.HasPrefix(lower, "x-")
}

func hashSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// convertRequest converts llmx.ChatRequest to Doubao format
func (p *DoubaoProvider) convertRequest(req *llmx.ChatRequest) map[string]interface{} {
	doubaoReq := map[string]interface{}{
		"model": req.Model,
	}

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
	doubaoReq["messages"] = messages

	// Set optional parameters
	if req.Temperature != nil {
		doubaoReq["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		doubaoReq["top_p"] = *req.TopP
	}
	if req.MaxTokens != nil {
		doubaoReq["max_tokens"] = *req.MaxTokens
	}

	// Convert tools if present
	if len(req.Tools) > 0 {
		doubaoReq["tools"] = convertTools(req.Tools)
	}

	return doubaoReq
}

// convertResponse converts Doubao response to llmx format
func (p *DoubaoProvider) convertResponse(resp map[string]interface{}, model string) *llmx.ChatResponse {
	llmxResp := &llmx.ChatResponse{
		Model:        model,
		FinishReason: "stop",
	}

	// Extract choices
	if choices, ok := resp["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if text, ok := message["content"].(string); ok {
					llmxResp.Content = text
				}
			}
			llmxResp.FinishReason = getFinishReason(choice)
		}
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

// handleStream processes Doubao streaming responses
func (p *DoubaoProvider) handleStream(ctx context.Context, resp *http.Response, chatStream *llmx.ChatStream) {
	defer resp.Body.Close()
	defer chatStream.Close()

	chatStream.SendEvent(core.StreamEvent{
		Type: core.EventTypeStart,
	})

	reader := resp.Body
	decoder := json.NewDecoder(reader)

	for {
		select {
		case <-ctx.Done():
			chatStream.SendError(ctx.Err())
			return
		default:
			var chunk map[string]interface{}
			if err := decoder.Decode(&chunk); err != nil {
				if err != io.EOF {
					chatStream.SendError(fmt.Errorf("doubao: stream decode error: %w", err))
				}
				chatStream.SendEvent(core.StreamEvent{
					Type: core.EventTypeFinish,
				})
				return
			}

			// Check for errors
			if errCode, ok := chunk["error_code"].(string); ok && errCode != "" {
				chatStream.SendError(p.convertAPIError(chunk))
				return
			}

			// Extract delta
			if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						if content, ok := delta["content"].(string); ok && content != "" {
							chatStream.SendEvent(core.StreamEvent{
								Type: core.EventTypeTextDelta,
								Data: content,
							})
						}
					}

					// Check if finished
					if finishReason, ok := choice["finish_reason"].(string); ok && finishReason != "" {
						chatStream.SendEvent(core.StreamEvent{
							Type: core.EventTypeFinish,
							Data: finishReason,
						})
						return
					}
				}
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
		return "system"
	default:
		return "user"
	}
}

func convertTools(tools []llmx.Tool) []map[string]interface{} {
	var doubaoTools []map[string]interface{}
	for _, tool := range tools {
		doubaoTools = append(doubaoTools, map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  tool.Parameters,
			},
		})
	}
	return doubaoTools
}

func getFinishReason(choice map[string]interface{}) string {
	if reason, ok := choice["finish_reason"].(string); ok {
		return reason
	}
	return "stop"
}

func (p *DoubaoProvider) convertError(body []byte, statusCode int) error {
	var errResp map[string]interface{}
	json.Unmarshal(body, &errResp)

	errMsg := fmt.Sprintf("doubao: HTTP %d", statusCode)
	if msg, ok := errResp["error_msg"].(string); ok {
		errMsg = msg
	} else if msg, ok := errResp["message"].(string); ok {
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

func (p *DoubaoProvider) convertAPIError(resp map[string]interface{}) error {
	errCode := ""
	if code, ok := resp["error_code"].(string); ok {
		errCode = code
	}

	errMsg := ""
	if msg, ok := resp["error_msg"].(string); ok {
		errMsg = msg
	} else if msg, ok := resp["message"].(string); ok {
		errMsg = msg
	}

	fullMsg := fmt.Sprintf("doubao API error %s: %s", errCode, errMsg)

	// Map error codes to llmx error types
	switch errCode {
	case "RATE_LIMIT_EXCEEDED":
		return llmx.NewRateLimitError(fullMsg, 0)
	case "AUTHENTICATION_FAILED", "PERMISSION_DENIED":
		return llmx.NewAuthenticationError(fullMsg)
	case "INVALID_REQUEST", "INVALID_PARAMETER":
		return llmx.NewInvalidRequestError(fullMsg, nil)
	default:
		return llmx.NewInternalError(fullMsg, nil)
	}
}
