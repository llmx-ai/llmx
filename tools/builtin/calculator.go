package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/llmx-ai/llmx"
)

// CalculatorTool provides basic calculator functionality
func CalculatorTool() llmx.Tool {
	return llmx.Tool{
		Name:        "calculator",
		Description: "Performs basic arithmetic calculations. Supports +, -, *, /, %, ^(power), sqrt operations.",
		Parameters: &llmx.Schema{
			Type: "object",
			Properties: map[string]*llmx.Schema{
				"expression": {
					Type:        "string",
					Description: "The mathematical expression to evaluate (e.g., '2 + 2', '10 * 5', 'sqrt(16)')",
				},
			},
			Required: []string{"expression"},
		},
		Execute: executeCalculator,
	}
}

func executeCalculator(ctx context.Context, args json.RawMessage) (*llmx.ToolResult, error) {
	var params struct {
		Expression string `json:"expression"`
	}

	if err := json.Unmarshal(args, &params); err != nil {
		return &llmx.ToolResult{
			Output:  fmt.Sprintf("Invalid arguments: %v", err),
			IsError: true,
		}, nil
	}

	result, err := evaluateExpression(params.Expression)
	if err != nil {
		return &llmx.ToolResult{
			Output:  fmt.Sprintf("Calculation error: %v", err),
			IsError: true,
		}, nil
	}

	return &llmx.ToolResult{
		Output: fmt.Sprintf("%s = %v", params.Expression, result),
		Metadata: map[string]interface{}{
			"expression": params.Expression,
			"result":     result,
		},
	}, nil
}

// Simple expression evaluator (supports basic operations)
func evaluateExpression(expr string) (float64, error) {
	expr = strings.TrimSpace(expr)

	// Handle sqrt
	if strings.HasPrefix(expr, "sqrt(") && strings.HasSuffix(expr, ")") {
		inner := expr[5 : len(expr)-1]
		val, err := evaluateExpression(inner)
		if err != nil {
			return 0, err
		}
		return math.Sqrt(val), nil
	}

	// Handle basic operations
	for _, op := range []string{"+", "-", "*", "/", "%", "^"} {
		parts := strings.Split(expr, op)
		if len(parts) == 2 {
			left, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			if err != nil {
				continue
			}
			right, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			if err != nil {
				continue
			}

			switch op {
			case "+":
				return left + right, nil
			case "-":
				return left - right, nil
			case "*":
				return left * right, nil
			case "/":
				if right == 0 {
					return 0, fmt.Errorf("division by zero")
				}
				return left / right, nil
			case "%":
				return math.Mod(left, right), nil
			case "^":
				return math.Pow(left, right), nil
			}
		}
	}

	// Try to parse as single number
	return strconv.ParseFloat(expr, 64)
}
