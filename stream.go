package llmx

import (
	"context"
	"sync"

	"github.com/llmx-ai/llmx/core"
)

// ChatStream represents a streaming chat response
type ChatStream struct {
	ctx    context.Context
	events chan core.StreamEvent
	errors chan error
	done   chan struct{}
	once   sync.Once
	mu     sync.Mutex

	// Accumulated response
	accumulated *ChatResponse
}

// NewChatStream creates a new chat stream
func NewChatStream(ctx context.Context) *ChatStream {
	return &ChatStream{
		ctx:    ctx,
		events: make(chan core.StreamEvent, 100),
		errors: make(chan error, 10),
		done:   make(chan struct{}),
		accumulated: &ChatResponse{
			Content: "",
		},
	}
}

// Events returns the events channel
func (s *ChatStream) Events() <-chan core.StreamEvent {
	return s.events
}

// Errors returns the errors channel
func (s *ChatStream) Errors() <-chan error {
	return s.errors
}

// SendEvent sends an event to the stream
func (s *ChatStream) SendEvent(event core.StreamEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.done:
		return
	case s.events <- event:
		// Accumulate text deltas
		if event.Type == core.EventTypeTextDelta {
			if text, ok := event.Data.(string); ok {
				s.accumulated.Content += text
			}
		}
	case <-s.ctx.Done():
		return
	}
}

// SendError sends an error to the stream
func (s *ChatStream) SendError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.done:
		return
	case s.errors <- err:
	case <-s.ctx.Done():
		return
	}
}

// Close closes the stream
func (s *ChatStream) Close() error {
	s.once.Do(func() {
		close(s.done)
		close(s.events)
		close(s.errors)
	})
	return nil
}

// Accumulate waits for the stream to complete and returns the full response
func (s *ChatStream) Accumulate() (*ChatResponse, error) {
	for {
		select {
		case event, ok := <-s.events:
			if !ok {
				return s.accumulated, nil
			}
			if event.Type == core.EventTypeError {
				if err, ok := event.Data.(error); ok {
					return nil, err
				}
			}
		case err := <-s.errors:
			return nil, err
		case <-s.ctx.Done():
			return nil, s.ctx.Err()
		}
	}
}

// GetAccumulated returns the accumulated response so far
func (s *ChatStream) GetAccumulated() *ChatResponse {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.accumulated
}
