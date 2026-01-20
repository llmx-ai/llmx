package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/llmx-ai/llmx"
)

// Executor executes tools and manages tool calling loops
type Executor struct {
	registry *Registry
	maxDepth int
}

// NewExecutor creates a new tool executor
func NewExecutor(registry *Registry) *Executor {
	return &Executor{
		registry: registry,
		maxDepth: 10, // Default max recursion depth
	}
}

// WithMaxDepth sets the maximum recursion depth
func (e *Executor) WithMaxDepth(depth int) *Executor {
	e.maxDepth = depth
	return e
}

// ExecuteLoop automatically executes tools until no more tool calls are needed
func (e *Executor) ExecuteLoop(
	ctx context.Context,
	client interface{},
	req *llmx.ChatRequest,
) (*llmx.ChatResponse, error) {
	// Type assert client
	llmxClient, ok := client.(*llmx.Client)
	if !ok {
		return nil, fmt.Errorf("invalid client type")
	}

	messages := append([]llmx.Message{}, req.Messages...)
	depth := 0

	for depth < e.maxDepth {
		// Create request with current messages
		currentReq := &llmx.ChatRequest{
			Model:       req.Model,
			Messages:    messages,
			Tools:       req.Tools,
			Temperature: req.Temperature,
			MaxTokens:   req.MaxTokens,
			TopP:        req.TopP,
			Stop:        req.Stop,
		}

		// Call AI
		resp, err := llmxClient.Chat(ctx, currentReq)
		if err != nil {
			return nil, err
		}

		// If no tool calls, return final response
		if len(resp.ToolCalls) == 0 {
			return resp, nil
		}

		// Add assistant message with tool calls
		messages = append(messages, llmx.Message{
			Role: llmx.RoleAssistant,
			Content: []llmx.ContentPart{
				llmx.TextPart{Text: resp.Content},
			},
			ToolCalls: resp.ToolCalls,
		})

		// Execute all tool calls
		for _, toolCall := range resp.ToolCalls {
			result, err := e.ExecuteSingle(ctx, toolCall)
			if err != nil {
				// Add error as tool result
				messages = append(messages, llmx.Message{
					Role: llmx.RoleTool,
					Content: []llmx.ContentPart{
						llmx.ToolResultPart{
							ToolCallID: toolCall.ID,
							Result:     fmt.Sprintf("Error: %v", err),
							IsError:    true,
						},
					},
				})
			} else {
				// Add successful result
				messages = append(messages, llmx.Message{
					Role: llmx.RoleTool,
					Content: []llmx.ContentPart{
						llmx.ToolResultPart{
							ToolCallID: toolCall.ID,
							Result:     result.Output,
							IsError:    result.IsError,
						},
					},
				})
			}
		}

		depth++
	}

	return nil, fmt.Errorf("max tool call depth reached: %d", e.maxDepth)
}

// ExecuteSingle executes a single tool call
func (e *Executor) ExecuteSingle(
	ctx context.Context,
	toolCall llmx.ToolCall,
) (*llmx.ToolResult, error) {
	// Get tool from registry
	tool, ok := e.registry.Get(toolCall.Name)
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", toolCall.Name)
	}

	// Validate arguments if schema is provided
	if tool.Parameters != nil {
		if err := ValidateToolArguments(tool.Parameters, toolCall.Arguments); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
	}

	// Execute tool
	return tool.Execute(ctx, toolCall.Arguments)
}

// ValidateToolArguments validates tool arguments against schema
func ValidateToolArguments(schema *llmx.Schema, args json.RawMessage) error {
	// Basic validation
	if schema == nil {
		return nil
	}

	// Parse arguments
	var argMap map[string]interface{}
	if err := json.Unmarshal(args, &argMap); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Check required fields
	for _, required := range schema.Required {
		if _, ok := argMap[required]; !ok {
			return fmt.Errorf("missing required field: %s", required)
		}
	}

	// TODO: Add more detailed validation (type checking, etc.)

	return nil
}
