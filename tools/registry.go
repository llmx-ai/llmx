package tools

import (
	"fmt"
	"sync"

	"github.com/llmx-ai/llmx"
)

// Registry manages tool registration and retrieval
type Registry struct {
	mu    sync.RWMutex
	tools map[string]llmx.Tool
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]llmx.Tool),
	}
}

// Register adds a tool to the registry
func (r *Registry) Register(tool llmx.Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if tool.Name == "" {
		return fmt.Errorf("tool name is required")
	}

	if tool.Execute == nil {
		return fmt.Errorf("tool execute function is required")
	}

	if _, exists := r.tools[tool.Name]; exists {
		return fmt.Errorf("tool %s already registered", tool.Name)
	}

	r.tools[tool.Name] = tool
	return nil
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (llmx.Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, ok := r.tools[name]
	return tool, ok
}

// List returns all registered tools
func (r *Registry) List() []llmx.Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]llmx.Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// Remove removes a tool from the registry
func (r *Registry) Remove(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[name]; !exists {
		return false
	}

	delete(r.tools, name)
	return true
}

// Has checks if a tool exists
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.tools[name]
	return ok
}

// Clear removes all tools
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tools = make(map[string]llmx.Tool)
}

// Count returns the number of registered tools
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.tools)
}

// Global registry instance
var globalRegistry = NewRegistry()

// GlobalRegistry returns the global tool registry
func GlobalRegistry() *Registry {
	return globalRegistry
}

// Register adds a tool to the global registry
func Register(tool llmx.Tool) error {
	return globalRegistry.Register(tool)
}

// Get retrieves a tool from the global registry
func Get(name string) (llmx.Tool, bool) {
	return globalRegistry.Get(name)
}

// List returns all tools from the global registry
func List() []llmx.Tool {
	return globalRegistry.List()
}
