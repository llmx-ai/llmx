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

	for {
		select {
		case <-ctx.Done():
			chatStream.SendError(ctx.Err())
			return
		default:
			response, err := stream.Recv()
			if err == io.EOF {
				// Stream finished
				chatStream.SendEvent(core.StreamEvent{
					Type: core.EventTypeFinish,
				})
				return
			}
			if err != nil {
				chatStream.SendError(p.convertError(err))
				return
			}

			// Process response chunks
			for _, choice := range response.Choices {
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
