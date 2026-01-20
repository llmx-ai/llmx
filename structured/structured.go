package structured

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/llmx-ai/llmx"
)

// Output is a helper for structured output generation
type Output struct {
	client *llmx.Client
}

// New creates a new structured output helper
func New(client *llmx.Client) *Output {
	return &Output{client: client}
}

// Generate generates structured output based on a schema
func (o *Output) Generate(
	ctx context.Context,
	prompt string,
	schema *llmx.Schema,
) (map[string]interface{}, error) {
	// Build system message with JSON instruction
	systemMsg := "You must respond with valid JSON that matches the provided schema. Do not include any text outside the JSON object."

	// Create request with JSON mode
	req := &llmx.ChatRequest{
		Messages: []llmx.Message{
			{
				Role: llmx.RoleSystem,
				Content: []llmx.ContentPart{
					llmx.TextPart{Text: systemMsg},
				},
			},
			{
				Role: llmx.RoleUser,
				Content: []llmx.ContentPart{
					llmx.TextPart{Text: prompt + "\n\nSchema:\n" + schemaToString(schema)},
				},
			},
		},
		ProviderOptions: map[string]interface{}{
			"response_format": map[string]string{
				"type": "json_object",
			},
		},
	}

	// Execute request
	resp, err := o.client.Chat(ctx, req)
	if err != nil {
		return nil, err
	}

	// Parse JSON response
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Validate against schema
	if err := validateAgainstSchema(result, schema); err != nil {
		return nil, fmt.Errorf("response doesn't match schema: %w", err)
	}

	return result, nil
}

// GenerateInto generates structured output and unmarshals into a Go struct
func (o *Output) GenerateInto(
	ctx context.Context,
	prompt string,
	target interface{},
) error {
	// Get schema from target type
	schema := schemaFromType(reflect.TypeOf(target))

	// Generate JSON
	result, err := o.Generate(ctx, prompt, schema)
	if err != nil {
		return err
	}

	// Marshal and unmarshal to populate target
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, target)
}

// schemaToString converts a schema to a readable string
func schemaToString(schema *llmx.Schema) string {
	data, _ := json.MarshalIndent(schema, "", "  ")
	return string(data)
}

// validateAgainstSchema validates data against a schema
func validateAgainstSchema(data map[string]interface{}, schema *llmx.Schema) error {
	if schema == nil {
		return nil
	}

	// Check required fields
	for _, required := range schema.Required {
		if _, ok := data[required]; !ok {
			return fmt.Errorf("missing required field: %s", required)
		}
	}

	// Check properties
	if schema.Properties != nil {
		for key, propSchema := range schema.Properties {
			if value, ok := data[key]; ok {
				if err := validateValue(value, propSchema); err != nil {
					return fmt.Errorf("field %s: %w", key, err)
				}
			}
		}
	}

	return nil
}

// validateValue validates a value against a schema
func validateValue(value interface{}, schema *llmx.Schema) error {
	if schema == nil {
		return nil
	}

	// Type checking
	switch schema.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "number":
		switch value.(type) {
		case float64, int, int64:
			// OK
		default:
			return fmt.Errorf("expected number, got %T", value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("expected array, got %T", value)
		}
	case "object":
		if obj, ok := value.(map[string]interface{}); ok {
			return validateAgainstSchema(obj, schema)
		}
		return fmt.Errorf("expected object, got %T", value)
	}

	// Enum checking
	if len(schema.Enum) > 0 {
		found := false
		for _, enumVal := range schema.Enum {
			if value == enumVal {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("value not in enum: %v", value)
		}
	}

	return nil
}

// schemaFromType generates a schema from a Go type
func schemaFromType(t reflect.Type) *llmx.Schema {
	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	schema := &llmx.Schema{
		Type:       "object",
		Properties: make(map[string]*llmx.Schema),
		Required:   []string{},
	}

	// Only handle struct types
	if t.Kind() != reflect.Struct {
		return schema
	}

	// Iterate over struct fields
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Get JSON tag
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Parse tag (handle omitempty)
		fieldName := jsonTag
		for idx := 0; idx < len(jsonTag); idx++ {
			if jsonTag[idx] == ',' {
				fieldName = jsonTag[:idx]
				break
			}
		}

		// Generate field schema
		fieldSchema := schemaFromFieldType(field.Type)
		if desc := field.Tag.Get("description"); desc != "" {
			fieldSchema.Description = desc
		}

		schema.Properties[fieldName] = fieldSchema

		// Check if required (no omitempty tag)
		if jsonTag == fieldName {
			schema.Required = append(schema.Required, fieldName)
		}
	}

	return schema
}

// schemaFromFieldType generates a schema from a field type
func schemaFromFieldType(t reflect.Type) *llmx.Schema {
	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return &llmx.Schema{Type: "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return &llmx.Schema{Type: "number"}
	case reflect.Bool:
		return &llmx.Schema{Type: "boolean"}
	case reflect.Slice, reflect.Array:
		return &llmx.Schema{
			Type:  "array",
			Items: schemaFromFieldType(t.Elem()),
		}
	case reflect.Struct:
		return schemaFromType(t)
	case reflect.Map:
		return &llmx.Schema{Type: "object"}
	default:
		return &llmx.Schema{Type: "string"}
	}
}
