package core

// EventType represents the type of streaming event
type EventType string

const (
	EventTypeStart          EventType = "start"
	EventTypeTextDelta      EventType = "text-delta"
	EventTypeToolCall       EventType = "tool-call"
	EventTypeToolCallDelta  EventType = "tool-call-delta"
	EventTypeReasoning      EventType = "reasoning"
	EventTypeReasoningDelta EventType = "reasoning-delta"
	EventTypeFinish         EventType = "finish"
	EventTypeError          EventType = "error"
)

// StreamEvent represents a streaming event
type StreamEvent struct {
	Type EventType
	Data interface{}
}
