package validation

import (
	"encoding/json"
	"fmt"
	"os"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
)

// ValidateJSON validates the given JSON bytes against the JSON Schema at schemaPath.
func ValidateJSON(schemaPath string, jsonBytes []byte) error {
	c := jsonschema.NewCompiler()
	// Allow loading from file path
	if err := c.AddResource(schemaPath, mustOpen(schemaPath)); err != nil {
		return fmt.Errorf("failed to load schema %s: %w", schemaPath, err)
	}
	schema, err := c.Compile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to compile schema %s: %w", schemaPath, err)
	}
	var v any
	if err := json.Unmarshal(jsonBytes, &v); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}
	if err := schema.Validate(v); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}
	return nil
}

func mustOpen(path string) *os.File {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	return f
}
