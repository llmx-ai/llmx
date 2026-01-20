package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/llmx-ai/llmx"
)

// DateTimeTool provides date and time information
func DateTimeTool() llmx.Tool {
	return llmx.Tool{
		Name:        "get_datetime",
		Description: "Gets the current date and time in a specified timezone",
		Parameters: &llmx.Schema{
			Type: "object",
			Properties: map[string]*llmx.Schema{
				"timezone": {
					Type:        "string",
					Description: "Timezone name (e.g., 'America/New_York', 'Asia/Shanghai', 'UTC'). Defaults to 'UTC' if not specified.",
				},
				"format": {
					Type:        "string",
					Description: "Date format ('short', 'long', 'timestamp'). Defaults to 'long'.",
				},
			},
		},
		Execute: executeDateTime,
	}
}

func executeDateTime(ctx context.Context, args json.RawMessage) (*llmx.ToolResult, error) {
	var params struct {
		Timezone string `json:"timezone"`
		Format   string `json:"format"`
	}

	if err := json.Unmarshal(args, &params); err != nil {
		return &llmx.ToolResult{
			Output:  fmt.Sprintf("Invalid arguments: %v", err),
			IsError: true,
		}, nil
	}

	// Default values
	if params.Timezone == "" {
		params.Timezone = "UTC"
	}
	if params.Format == "" {
		params.Format = "long"
	}

	// Load timezone
	loc, err := time.LoadLocation(params.Timezone)
	if err != nil {
		return &llmx.ToolResult{
			Output:  fmt.Sprintf("Invalid timezone: %v", err),
			IsError: true,
		}, nil
	}

	now := time.Now().In(loc)

	// Format output
	var output string
	switch params.Format {
	case "short":
		output = now.Format("2006-01-02 15:04:05")
	case "long":
		output = now.Format("Monday, January 2, 2006 at 3:04:05 PM MST")
	case "timestamp":
		output = fmt.Sprintf("%d", now.Unix())
	default:
		output = now.Format(time.RFC3339)
	}

	return &llmx.ToolResult{
		Output: fmt.Sprintf("Current time in %s: %s", params.Timezone, output),
		Metadata: map[string]interface{}{
			"timezone":  params.Timezone,
			"timestamp": now.Unix(),
			"iso8601":   now.Format(time.RFC3339),
		},
	}, nil
}
