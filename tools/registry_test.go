package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/llmx-ai/llmx"
)

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	tool := llmx.Tool{
		Name:        "test_tool",
		Description: "A test tool",
		Execute: func(ctx context.Context, args json.RawMessage) (*llmx.ToolResult, error) {
			return &llmx.ToolResult{Output: "test"}, nil
		},
	}

	// Test successful registration
	err := registry.Register(tool)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test duplicate registration
	err = registry.Register(tool)
	if err == nil {
		t.Error("Expected error for duplicate registration")
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()

	tool := llmx.Tool{
		Name:        "test_tool",
		Description: "A test tool",
		Execute: func(ctx context.Context, args json.RawMessage) (*llmx.ToolResult, error) {
			return &llmx.ToolResult{Output: "test"}, nil
		},
	}

	registry.Register(tool)

	// Test getting existing tool
	retrieved, ok := registry.Get("test_tool")
	if !ok {
		t.Error("Expected to find tool")
	}
	if retrieved.Name != "test_tool" {
		t.Errorf("Expected tool name 'test_tool', got '%s'", retrieved.Name)
	}

	// Test getting non-existent tool
	_, ok = registry.Get("nonexistent")
	if ok {
		t.Error("Expected not to find tool")
	}
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	tool1 := llmx.Tool{
		Name:        "tool1",
		Description: "Tool 1",
		Execute: func(ctx context.Context, args json.RawMessage) (*llmx.ToolResult, error) {
			return &llmx.ToolResult{Output: "test"}, nil
		},
	}

	tool2 := llmx.Tool{
		Name:        "tool2",
		Description: "Tool 2",
		Execute: func(ctx context.Context, args json.RawMessage) (*llmx.ToolResult, error) {
			return &llmx.ToolResult{Output: "test"}, nil
		},
	}

	registry.Register(tool1)
	registry.Register(tool2)

	tools := registry.List()
	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}
}

func TestRegistry_Remove(t *testing.T) {
	registry := NewRegistry()

	tool := llmx.Tool{
		Name:        "test_tool",
		Description: "A test tool",
		Execute: func(ctx context.Context, args json.RawMessage) (*llmx.ToolResult, error) {
			return &llmx.ToolResult{Output: "test"}, nil
		},
	}

	registry.Register(tool)

	// Test successful removal
	removed := registry.Remove("test_tool")
	if !removed {
		t.Error("Expected tool to be removed")
	}

	// Test removing non-existent tool
	removed = registry.Remove("nonexistent")
	if removed {
		t.Error("Expected removal to fail")
	}
}

func TestRegistry_Has(t *testing.T) {
	registry := NewRegistry()

	tool := llmx.Tool{
		Name:        "test_tool",
		Description: "A test tool",
		Execute: func(ctx context.Context, args json.RawMessage) (*llmx.ToolResult, error) {
			return &llmx.ToolResult{Output: "test"}, nil
		},
	}

	registry.Register(tool)

	if !registry.Has("test_tool") {
		t.Error("Expected tool to exist")
	}

	if registry.Has("nonexistent") {
		t.Error("Expected tool not to exist")
	}
}

func TestRegistry_Clear(t *testing.T) {
	registry := NewRegistry()

	tool := llmx.Tool{
		Name:        "test_tool",
		Description: "A test tool",
		Execute: func(ctx context.Context, args json.RawMessage) (*llmx.ToolResult, error) {
			return &llmx.ToolResult{Output: "test"}, nil
		},
	}

	registry.Register(tool)
	registry.Clear()

	if registry.Count() != 0 {
		t.Errorf("Expected 0 tools after clear, got %d", registry.Count())
	}
}
