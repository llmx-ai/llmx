package llmx

import (
	"context"
	"fmt"

	"github.com/llmx-ai/llmx/core"
)

// SimpleChat is a convenience method for single-turn text-only conversations
// It automatically handles context, request creation, and response extraction
func (c *Client) SimpleChat(ctx context.Context, message string) (string, error) {
	req := &ChatRequest{
		Messages: []Message{
			{
				Role:    RoleUser,
				Content: []ContentPart{TextPart{Text: message}},
			},
		},
	}

	resp, err := c.Chat(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// SimpleChatSync is like SimpleChat but uses a background context
func (c *Client) SimpleChatSync(message string) (string, error) {
	return c.SimpleChat(context.Background(), message)
}

// SimpleStreamChat is a convenience method for single-turn streaming conversations
// The onChunk callback is called for each text delta received
func (c *Client) SimpleStreamChat(ctx context.Context, message string, onChunk func(string)) error {
	req := &ChatRequest{
		Messages: []Message{
			{
				Role:    RoleUser,
				Content: []ContentPart{TextPart{Text: message}},
			},
		},
	}

	stream, err := c.StreamChat(ctx, req)
	if err != nil {
		return err
	}
	defer stream.Close()

	for {
		select {
		case event, ok := <-stream.Events():
			if !ok {
				// Stream closed
				return nil
			}

			switch event.Type {
			case core.EventTypeTextDelta:
				if text, ok := event.Data.(string); ok && onChunk != nil {
					onChunk(text)
				}
			case core.EventTypeError:
				if err, ok := event.Data.(error); ok {
					return err
				}
			case core.EventTypeFinish:
				return nil
			}

		case err := <-stream.Errors():
			return err

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// SimpleStreamChatSync is like SimpleStreamChat but uses a background context
func (c *Client) SimpleStreamChatSync(message string, onChunk func(string)) error {
	return c.SimpleStreamChat(context.Background(), message, onChunk)
}

// TextOnly extracts just the text content from a ChatResponse
func TextOnly(resp *ChatResponse) string {
	if resp == nil {
		return ""
	}
	return resp.Content
}

// MessageBuilder provides a fluent API for building messages
type MessageBuilder struct {
	role    MessageRole
	parts   []ContentPart
	cache   bool
	cacheID string
}

// NewUserMessage creates a builder for a user message
func NewUserMessage() *MessageBuilder {
	return &MessageBuilder{role: RoleUser}
}

// NewAssistantMessage creates a builder for an assistant message
func NewAssistantMessage() *MessageBuilder {
	return &MessageBuilder{role: RoleAssistant}
}

// NewSystemMessage creates a builder for a system message
func NewSystemMessage() *MessageBuilder {
	return &MessageBuilder{role: RoleSystem}
}

// Text adds a text part to the message
func (mb *MessageBuilder) Text(text string) *MessageBuilder {
	mb.parts = append(mb.parts, TextPart{Text: text})
	return mb
}

// Image adds an image part to the message
func (mb *MessageBuilder) Image(url string, detail ...string) *MessageBuilder {
	part := ImagePart{URL: url}
	if len(detail) > 0 {
		part.Detail = detail[0]
	}
	mb.parts = append(mb.parts, part)
	return mb
}

// ToolCall adds a tool call part to the message
func (mb *MessageBuilder) ToolCall(id, name string, arguments []byte) *MessageBuilder {
	mb.parts = append(mb.parts, ToolCall{
		ID:        id,
		Name:      name,
		Arguments: arguments,
	})
	return mb
}

// ToolResult adds a tool result part to the message
func (mb *MessageBuilder) ToolResult(toolCallID, result string) *MessageBuilder {
	mb.parts = append(mb.parts, ToolResultPart{
		ToolCallID: toolCallID,
		Result:     result,
	})
	return mb
}

// WithCache enables prompt caching for this message (provider-specific)
func (mb *MessageBuilder) WithCache(cacheID string) *MessageBuilder {
	mb.cache = true
	mb.cacheID = cacheID
	return mb
}

// Build creates the final Message
func (mb *MessageBuilder) Build() (Message, error) {
	if len(mb.parts) == 0 {
		return Message{}, fmt.Errorf("message must have at least one content part")
	}

	return Message{
		Role:    mb.role,
		Content: mb.parts,
	}, nil
}

// MustBuild is like Build but panics on error
func (mb *MessageBuilder) MustBuild() Message {
	msg, err := mb.Build()
	if err != nil {
		panic(err)
	}
	return msg
}

// RequestBuilder provides a fluent API for building chat requests
type RequestBuilder struct {
	req *ChatRequest
	err error
}

// NewRequest creates a new RequestBuilder
func NewRequest(model string) *RequestBuilder {
	return &RequestBuilder{
		req: &ChatRequest{
			Model:    model,
			Messages: []Message{},
		},
	}
}

// Message adds a message to the request
func (rb *RequestBuilder) Message(msg Message) *RequestBuilder {
	if rb.err != nil {
		return rb
	}
	rb.req.Messages = append(rb.req.Messages, msg)
	return rb
}

// UserMessage adds a simple user text message
func (rb *RequestBuilder) UserMessage(text string) *RequestBuilder {
	if rb.err != nil {
		return rb
	}
	rb.req.Messages = append(rb.req.Messages, Message{
		Role:    RoleUser,
		Content: []ContentPart{TextPart{Text: text}},
	})
	return rb
}

// AssistantMessage adds a simple assistant text message
func (rb *RequestBuilder) AssistantMessage(text string) *RequestBuilder {
	if rb.err != nil {
		return rb
	}
	rb.req.Messages = append(rb.req.Messages, Message{
		Role:    RoleAssistant,
		Content: []ContentPart{TextPart{Text: text}},
	})
	return rb
}

// SystemMessage adds a simple system text message
func (rb *RequestBuilder) SystemMessage(text string) *RequestBuilder {
	if rb.err != nil {
		return rb
	}
	rb.req.Messages = append(rb.req.Messages, Message{
		Role:    RoleSystem,
		Content: []ContentPart{TextPart{Text: text}},
	})
	return rb
}

// Temperature sets the temperature
func (rb *RequestBuilder) Temperature(temp float64) *RequestBuilder {
	if rb.err != nil {
		return rb
	}
	rb.req.Temperature = &temp
	return rb
}

// MaxTokens sets the max tokens
func (rb *RequestBuilder) MaxTokens(tokens int) *RequestBuilder {
	if rb.err != nil {
		return rb
	}
	rb.req.MaxTokens = &tokens
	return rb
}

// TopP sets the top_p
func (rb *RequestBuilder) TopP(topP float64) *RequestBuilder {
	if rb.err != nil {
		return rb
	}
	rb.req.TopP = &topP
	return rb
}

// Tools adds tools to the request
func (rb *RequestBuilder) Tools(tools ...Tool) *RequestBuilder {
	if rb.err != nil {
		return rb
	}
	rb.req.Tools = append(rb.req.Tools, tools...)
	return rb
}

// Build creates the final ChatRequest
func (rb *RequestBuilder) Build() (*ChatRequest, error) {
	if rb.err != nil {
		return nil, rb.err
	}
	if len(rb.req.Messages) == 0 {
		return nil, fmt.Errorf("request must have at least one message")
	}
	return rb.req, nil
}

// MustBuild is like Build but panics on error
func (rb *RequestBuilder) MustBuild() *ChatRequest {
	req, err := rb.Build()
	if err != nil {
		panic(err)
	}
	return req
}
