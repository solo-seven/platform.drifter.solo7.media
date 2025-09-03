package validation_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/solo-seven/platform.drifter.solo7.media/internal/validation"
)

func TestValidateEntitySchema_Simple(t *testing.T) {
	// Get the project root directory (go up from tests/unit/validation)
	projectRoot := filepath.Join("..", "..", "..")
	schema := filepath.Join(projectRoot, "schemas", "entity.schema.json")

	entity := map[string]any{
		"id":       "550e8400-e29b-41d4-a716-446655440000",
		"metadata": map[string]any{"created_at": "2024-01-01T00:00:00Z", "updated_at": "2024-01-01T00:00:00Z", "version": 1},
		"components": map[string]any{
			"transform": map[string]any{
				"type":    "transform",
				"version": 1,
				"data": map[string]any{
					"position": map[string]any{"x": 0.0, "y": 1.0, "z": 2.0},
					"rotation": map[string]any{"x": 0.0, "y": 0.0, "z": 0.0},
					"scale":    map[string]any{"x": 1.0, "y": 1.0, "z": 1.0},
				},
			},
		},
	}
	b, _ := json.Marshal(entity)
	if err := validation.ValidateJSON(schema, b); err != nil {
		t.Fatalf("expected valid entity, got error: %v", err)
	}
}

func TestValidateRuleSchema_Simple(t *testing.T) {
	// Get the project root directory (go up from tests/unit/validation)
	projectRoot := filepath.Join("..", "..", "..")
	schema := filepath.Join(projectRoot, "schemas", "rule.schema.json")
	rule := map[string]any{
		"id":       "550e8400-e29b-41d4-a716-446655440000",
		"priority": 10,
		"triggers": []any{
			map[string]any{"type": "combat_damage"},
		},
		"conditions": []any{
			map[string]any{"type": "expression", "property": "damage > 10 && target.armor < 15"},
		},
		"actions": []any{
			map[string]any{"type": "notification", "target": "combat", "properties": map[string]any{"message": "Critical hit!"}},
		},
	}
	b, _ := json.Marshal(rule)
	if err := validation.ValidateJSON(schema, b); err != nil {
		t.Fatalf("expected valid rule, got error: %v", err)
	}
}

func TestValidateExpressionSchema_Simple(t *testing.T) {
	// Get the project root directory (go up from tests/unit/validation)
	projectRoot := filepath.Join("..", "..", "..")
	schema := filepath.Join(projectRoot, "schemas", "expression.schema.json")
	b := []byte(`"roll(\"1d6\") + self.level"`)
	if err := validation.ValidateJSON(schema, b); err != nil {
		t.Fatalf("expected valid expression, got error: %v", err)
	}
}

func TestValidateEntitySchema_InvalidMissingFields(t *testing.T) {
	// Get the project root directory (go up from tests/unit/validation)
	projectRoot := filepath.Join("..", "..", "..")
	schema := filepath.Join(projectRoot, "schemas", "entity.schema.json")
	b := []byte(`{"id":"550e8400-e29b-41d4-a716-446655440000"}`)
	if err := validation.ValidateJSON(schema, b); err == nil {
		t.Fatal("expected validation error for missing fields, got nil")
	} else {
		_ = os.Getenv // keep os imported
	}
}
