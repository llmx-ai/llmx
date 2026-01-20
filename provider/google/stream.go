package google

import (
	"context"

	"cloud.google.com/go/vertexai/genai"
	"github.com/llmx-ai/llmx"
	"github.com/llmx-ai/llmx/core"
)

// handleStream processes the Google stream and sends events to the chat stream
func (p *GoogleProvider) handleStream(
	ctx context.Context,
	iter *genai.GenerateContentResponseIterator,
	chatStream *llmx.ChatStream,
) {
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
			resp, err := iter.Next()
			if err != nil {
				// Check if it's end of stream
				if err.Error() == "iterator done" {
					chatStream.SendEvent(core.StreamEvent{
						Type: core.EventTypeFinish,
					})
					return
				}
				chatStream.SendError(p.convertError(err))
				return
			}

			// Process response chunks
			if len(resp.Candidates) > 0 {
				candidate := resp.Candidates[0]
				if candidate.Content != nil {
					for _, part := range candidate.Content.Parts {
						if text, ok := part.(genai.Text); ok {
							chatStream.SendEvent(core.StreamEvent{
								Type: core.EventTypeTextDelta,
								Data: string(text),
							})
						}
					}
				}
			}
		}
	}
}
