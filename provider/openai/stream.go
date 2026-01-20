package openai

import (
	"context"
	"io"

	"github.com/llmx-ai/llmx"
	"github.com/llmx-ai/llmx/core"
	openai "github.com/sashabaranov/go-openai"
)

// handleStream processes the OpenAI stream and sends events to the chat stream
func (p *OpenAIProvider) handleStream(
	ctx context.Context,
	stream *openai.ChatCompletionStream,
	chatStream *llmx.ChatStream,
) {
	defer stream.Close()
	defer chatStream.Close()

	// Send start event
	chatStream.SendEvent(core.StreamEvent{
		Type: core.EventTypeStart,
	})

	// Use a channel for receiving stream responses to support cancellation
	type streamResult struct {
		response openai.ChatCompletionStreamResponse
		err      error
	}
	resultChan := make(chan streamResult, 1)

	for {
		// Start receiving in a goroutine to allow immediate cancellation
		go func() {
			response, err := stream.Recv()
			select {
			case resultChan <- streamResult{response: response, err: err}:
			case <-ctx.Done():
				// Context cancelled, don't send result
			}
		}()

		// Wait for either result or cancellation
		select {
		case <-ctx.Done():
			chatStream.SendError(ctx.Err())
			return

		case result := <-resultChan:
			if result.err == io.EOF {
				// Stream finished
				chatStream.SendEvent(core.StreamEvent{
					Type: core.EventTypeFinish,
				})
				return
			}
			if result.err != nil {
				chatStream.SendError(p.convertError(result.err))
				return
			}

			// Process response chunks
			for _, choice := range result.response.Choices {
				// Check context before processing each chunk
				select {
				case <-ctx.Done():
					chatStream.SendError(ctx.Err())
					return
				default:
				}

				// Text delta
				if choice.Delta.Content != "" {
					chatStream.SendEvent(core.StreamEvent{
						Type: core.EventTypeTextDelta,
						Data: choice.Delta.Content,
					})
				}

				// Tool calls
				if len(choice.Delta.ToolCalls) > 0 {
					for _, toolCall := range choice.Delta.ToolCalls {
						// For now, send full tool call (can be optimized later)
						if toolCall.Function.Name != "" {
							chatStream.SendEvent(core.StreamEvent{
								Type: core.EventTypeToolCall,
								Data: map[string]interface{}{
									"id":   toolCall.ID,
									"name": toolCall.Function.Name,
									"args": toolCall.Function.Arguments,
								},
							})
						}
					}
				}

				// Check finish reason
				if choice.FinishReason != "" {
					chatStream.SendEvent(core.StreamEvent{
						Type: core.EventTypeFinish,
						Data: string(choice.FinishReason),
					})
				}
			}
		}
	}
}
