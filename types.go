package llmx

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// MessageRole represents the role of a message sender
type MessageRole string

const (
	RoleSystem    MessageRole = "system"
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleTool      MessageRole = "tool"
)

// Message represents a chat message
type Message struct {
	Role      MessageRole   `json:"role"`
	Content   []ContentPart `json:"content"`
	ToolCalls []ToolCall    `json:"tool_calls,omitempty"`
}

// ContentPart is an interface for message content parts
type ContentPart interface {
	Type() ContentType
}

// ContentType represents the type of content
type ContentType string

const (
	ContentTypeText       ContentType = "text"
	ContentTypeImage      ContentType = "image"
	ContentTypeToolCall   ContentType = "tool_call"
	ContentTypeToolResult ContentType = "tool_result"
)

// TextPart represents text content
type TextPart struct {
	Text string `json:"text"`
}

func (t TextPart) Type() ContentType { return ContentTypeText }

// ImagePart represents image content
type ImagePart struct {
	URL    string `json:"url,omitempty"`
	Base64 string `json:"base64,omitempty"`
	Detail string `json:"detail,omitempty"` // "low", "high", "auto"
}

func (i ImagePart) Type() ContentType { return ContentTypeImage }

// ToolCall represents a tool call
type ToolCall struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

func (t ToolCall) Type() ContentType { return ContentTypeToolCall }

// ToolResultPart represents a tool execution result
type ToolResultPart struct {
	ToolCallID string `json:"tool_call_id"`
	Result     string `json:"result"`
	IsError    bool   `json:"is_error,omitempty"`
}

func (t ToolResultPart) Type() ContentType { return ContentTypeToolResult }

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`

	// Optional parameters
	Temperature *float64 `json:"temperature,omitempty"`
	MaxTokens   *int     `json:"max_tokens,omitempty"`
	TopP        *float64 `json:"top_p,omitempty"`
	TopK        *int     `json:"top_k,omitempty"`
	Stop        []string `json:"stop,omitempty"`

	// Tools
	Tools []Tool `json:"tools,omitempty"`

	// Provider-specific options
	ProviderOptions map[string]interface{} `json:"provider_options,omitempty"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	ID        string     `json:"id"`
	Model     string     `json:"model"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// Metadata
	Usage        Usage     `json:"usage"`
	FinishReason string    `json:"finish_reason"`
	CreatedAt    time.Time `json:"created_at"`

	// Raw response for debugging
	Raw interface{} `json:"raw,omitempty"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Tool represents a function that can be called by the AI
type Tool struct {
	Name        string
	Description string
	Parameters  *Schema
	Execute     ToolExecuteFunc
}

// ToolExecuteFunc is the function signature for tool execution
type ToolExecuteFunc func(ctx context.Context, args json.RawMessage) (*ToolResult, error)

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Output   string
	IsError  bool
	Metadata map[string]interface{}
}

// Schema represents a JSON Schema for tool parameters
type Schema struct {
	Type        string             `json:"type"`
	Description string             `json:"description,omitempty"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	Required    []string           `json:"required,omitempty"`
	Items       *Schema            `json:"items,omitempty"`
	Enum        []interface{}      `json:"enum,omitempty"`
}

// ValidateMessage validates a message
func ValidateMessage(message Message) error {
	if message.Role == "" {
		return NewInvalidRequestError("message role is required", nil)
	}

	if len(message.Content) == 0 && len(message.ToolCalls) == 0 {
		return NewInvalidRequestError("message must have content or tool calls", nil)
	}

	return nil
}

// ValidateMessages validates a slice of messages
func ValidateMessages(messages []Message) error {
	if len(messages) == 0 {
		return NewInvalidRequestError("at least one message is required", nil)
	}

	for i, msg := range messages {
		if err := ValidateMessage(msg); err != nil {
			return NewInvalidRequestError(fmt.Sprintf("invalid message at index %d: %v", i, err), nil)
		}
	}

	return nil
}

// ExtractText extracts text content from a message
func ExtractText(message Message) string {
	var text string
	for _, part := range message.Content {
		if textPart, ok := part.(TextPart); ok {
			text += textPart.Text
		}
	}
	return text
}

// Handler is a function that handles a chat request
type Handler func(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

// Middleware is a function that wraps a Handler
type Middleware func(next Handler) Handler
