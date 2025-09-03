package domain

import (
	"testing"
	"time"

	"github.com/solo7.media/platform.drifter.solo7.media/internal/domain"
)

func TestExpressionParser_BasicArithmetic(t *testing.T) {
	parser := domain.NewSimpleExpressionParser()
	ctx := &domain.ExpressionContext{}

	tests := []struct {
		expression string
		expected   float64
	}{
		{"2 + 3", 5},
		{"10 - 4", 6},
		{"3 * 4", 12},
		{"15 / 3", 5},
		{"2 ^ 3", 8},
		{"10 % 3", 1},
		{"(2 + 3) * 4", 20},
		{"2 + 3 * 4", 14}, // Operator precedence
		{"(2 + 3) * (4 - 1)", 15},
	}

	for _, test := range tests {
		t.Run(test.expression, func(t *testing.T) {
			result, err := parser.Evaluate(test.expression, ctx)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !result.Success {
				t.Fatalf("Evaluation failed: %s", result.Error)
			}

			if result.Type != "number" {
				t.Fatalf("Expected number type, got %s", result.Type)
			}

			actual, ok := result.Value.(float64)
			if !ok {
				t.Fatalf("Expected float64 value, got %T", result.Value)
			}

			if actual != test.expected {
				t.Fatalf("Expected %f, got %f", test.expected, actual)
			}
		})
	}
}

func TestExpressionParser_StringOperations(t *testing.T) {
	parser := domain.NewSimpleExpressionParser()
	ctx := &domain.ExpressionContext{}

	tests := []struct {
		expression string
		expected   string
	}{
		{`"hello" + " " + "world"`, "hello world"},
		{`"test" == "test"`, "true"},
		{`"test" != "hello"`, "true"},
	}

	for _, test := range tests {
		t.Run(test.expression, func(t *testing.T) {
			result, err := parser.Evaluate(test.expression, ctx)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !result.Success {
				t.Fatalf("Evaluation failed: %s", result.Error)
			}

			if test.expected == "true" || test.expected == "false" {
				// Boolean result
				expectedBool := test.expected == "true"
				actual, ok := result.Value.(bool)
				if !ok {
					t.Fatalf("Expected bool value, got %T", result.Value)
				}
				if actual != expectedBool {
					t.Fatalf("Expected %v, got %v", expectedBool, actual)
				}
			} else {
				// String result
				actual, ok := result.Value.(string)
				if !ok {
					t.Fatalf("Expected string value, got %T", result.Value)
				}
				if actual != test.expected {
					t.Fatalf("Expected %s, got %s", test.expected, actual)
				}
			}
		})
	}
}

func TestExpressionParser_BooleanOperations(t *testing.T) {
	parser := domain.NewSimpleExpressionParser()
	ctx := &domain.ExpressionContext{}

	tests := []struct {
		expression string
		expected   bool
	}{
		{"true && true", true},
		{"true && false", false},
		{"false && true", false},
		{"false && false", false},
		{"true || true", true},
		{"true || false", true},
		{"false || true", true},
		{"false || false", false},
		{"5 > 3", true},
		{"3 > 5", false},
		{"5 >= 5", true},
		{"5 < 3", false},
		{"3 <= 5", true},
		{"5 == 5", true},
		{"5 != 3", true},
	}

	for _, test := range tests {
		t.Run(test.expression, func(t *testing.T) {
			result, err := parser.Evaluate(test.expression, ctx)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !result.Success {
				t.Fatalf("Evaluation failed: %s", result.Error)
			}

			if result.Type != "boolean" {
				t.Fatalf("Expected boolean type, got %s", result.Type)
			}

			actual, ok := result.Value.(bool)
			if !ok {
				t.Fatalf("Expected bool value, got %T", result.Value)
			}

			if actual != test.expected {
				t.Fatalf("Expected %v, got %v", test.expected, actual)
			}
		})
	}
}

func TestExpressionParser_DiceRolling(t *testing.T) {
	parser := domain.NewSimpleExpressionParser()
	ctx := &domain.ExpressionContext{}

	tests := []struct {
		expression string
		expectType string
	}{
		{`roll("1d6")`, "dice_roll"},
		{`roll("2d6+3")`, "dice_roll"},
		{`roll("1d20-1")`, "dice_roll"},
		{`d(6)`, "dice_roll"},
		{`d(20)`, "dice_roll"},
	}

	for _, test := range tests {
		t.Run(test.expression, func(t *testing.T) {
			result, err := parser.Evaluate(test.expression, ctx)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !result.Success {
				t.Fatalf("Evaluation failed: %s", result.Error)
			}

			if result.Type != test.expectType {
				t.Fatalf("Expected %s type, got %s", test.expectType, result.Type)
			}

			// Check that we got a numeric value
			_, ok := result.Value.(float64)
			if !ok {
				t.Fatalf("Expected numeric value, got %T", result.Value)
			}

			// Check metadata for dice rolls
			if test.expectType == "dice_roll" {
				if result.Metadata == nil {
					t.Fatal("Expected metadata for dice roll")
				}
				if _, exists := result.Metadata["notation"]; !exists {
					t.Fatal("Expected notation in metadata")
				}
			}
		})
	}
}

func TestExpressionParser_ContextVariables(t *testing.T) {
	parser := domain.NewSimpleExpressionParser()
	ctx := &domain.ExpressionContext{
		Self: map[string]interface{}{
			"level":    5.0,
			"name":     "TestPlayer",
			"health":   100.0,
			"strength": 18.0,
		},
		Target: map[string]interface{}{
			"armor": 15.0,
			"name":  "Goblin",
		},
		Game: map[string]interface{}{
			"difficulty": "normal",
		},
	}

	tests := []struct {
		expression string
		expected   interface{}
		expectType string
	}{
		{"level", 5.0, "number"},
		{"name", "TestPlayer", "string"},
		{"health", 100.0, "number"},
		{"strength", 18.0, "number"},
		{"level + 1", 6.0, "number"},
		{"health > 50", true, "boolean"},
		{"strength * 2", 36.0, "number"},
	}

	for _, test := range tests {
		t.Run(test.expression, func(t *testing.T) {
			result, err := parser.Evaluate(test.expression, ctx)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !result.Success {
				t.Fatalf("Evaluation failed: %s", result.Error)
			}

			if result.Type != test.expectType {
				t.Fatalf("Expected %s type, got %s", test.expectType, result.Type)
			}

			if result.Value != test.expected {
				t.Fatalf("Expected %v, got %v", test.expected, result.Value)
			}
		})
	}
}

func TestExpressionParser_GameFunctions(t *testing.T) {
	parser := domain.NewSimpleExpressionParser()
	ctx := &domain.ExpressionContext{
		Self: map[string]interface{}{
			"level": 5.0,
		},
		Target: map[string]interface{}{
			"name": "Goblin",
		},
	}

	tests := []struct {
		expression string
		expectType string
	}{
		{`deal(target, 10)`, "number"},
		{`deal(target, 15, "fire")`, "number"},
		{`heal(self, 20)`, "number"},
		{`has_tag(target, "undead")`, "boolean"},
		{`has_ability(self, "second_wind")`, "boolean"},
		{`min(5, 10)`, "number"},
		{`max(5, 10)`, "number"},
		{`clamp(15, 0, 10)`, "number"},
	}

	for _, test := range tests {
		t.Run(test.expression, func(t *testing.T) {
			result, err := parser.Evaluate(test.expression, ctx)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !result.Success {
				t.Fatalf("Evaluation failed: %s", result.Error)
			}

			if result.Type != test.expectType {
				t.Fatalf("Expected %s type, got %s", test.expectType, result.Type)
			}
		})
	}
}

func TestExpressionParser_ComplexExpressions(t *testing.T) {
	parser := domain.NewSimpleExpressionParser()
	ctx := &domain.ExpressionContext{
		Self: map[string]interface{}{
			"level":    5.0,
			"strength": 18.0,
			"health":   100.0,
		},
		Target: map[string]interface{}{
			"armor": 15.0,
		},
	}

	tests := []struct {
		expression string
		expectType string
	}{
		{`level + roll("1d4")`, "number"},
		{`strength > 15 && health > 50`, "boolean"},
		{`deal(target, roll("1d8") + strength)`, "number"},
		{`heal(self, min(roll("2d6"), health))`, "number"},
		{`(level * 2) + (strength / 2)`, "number"},
	}

	for _, test := range tests {
		t.Run(test.expression, func(t *testing.T) {
			result, err := parser.Evaluate(test.expression, ctx)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !result.Success {
				t.Fatalf("Evaluation failed: %s", result.Error)
			}

			if result.Type != test.expectType {
				t.Fatalf("Expected %s type, got %s", test.expectType, result.Type)
			}
		})
	}
}

func TestExpressionParser_ErrorHandling(t *testing.T) {
	parser := domain.NewSimpleExpressionParser()
	ctx := &domain.ExpressionContext{}

	tests := []struct {
		expression  string
		expectError bool
	}{
		{"", false},                  // Empty expression should succeed
		{"1 / 0", true},              // Division by zero
		{"unknown_function()", true}, // Unknown function
		{"undefined_variable", true}, // Undefined variable
		{"(2 + 3", true},             // Mismatched parentheses
		{"2 + )", true},              // Invalid syntax
		{"roll()", true},             // Wrong number of arguments
		{"deal()", true},             // Wrong number of arguments
		{"min(1)", true},             // Wrong number of arguments
		{"clamp(1, 2, 3, 4)", true},  // Too many arguments
	}

	for _, test := range tests {
		t.Run(test.expression, func(t *testing.T) {
			result, err := parser.Evaluate(test.expression, ctx)

			if test.expectError {
				if err == nil && result.Success {
					t.Fatalf("Expected error for expression: %s", test.expression)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error for expression %s: %v", test.expression, err)
				}
				if !result.Success {
					t.Fatalf("Evaluation failed for expression %s: %s", test.expression, result.Error)
				}
			}
		})
	}
}

func TestExpressionParser_OperatorPrecedence(t *testing.T) {
	parser := domain.NewSimpleExpressionParser()
	ctx := &domain.ExpressionContext{}

	tests := []struct {
		expression string
		expected   float64
	}{
		{"2 + 3 * 4", 14},     // Multiplication before addition
		{"2 * 3 + 4", 10},     // Multiplication before addition
		{"2 + 3 * 4 + 5", 19}, // Multiple operations
		{"2 * 3 + 4 * 5", 26}, // Multiple multiplications
		{"2 ^ 3 * 4", 32},     // Exponentiation before multiplication
		{"2 * 3 ^ 2", 18},     // Exponentiation before multiplication
		{"(2 + 3) * 4", 20},   // Parentheses override precedence
		{"2 + (3 * 4)", 14},   // Parentheses override precedence
	}

	for _, test := range tests {
		t.Run(test.expression, func(t *testing.T) {
			result, err := parser.Evaluate(test.expression, ctx)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !result.Success {
				t.Fatalf("Evaluation failed: %s", result.Error)
			}

			actual, ok := result.Value.(float64)
			if !ok {
				t.Fatalf("Expected float64 value, got %T", result.Value)
			}

			if actual != test.expected {
				t.Fatalf("Expected %f, got %f", test.expected, actual)
			}
		})
	}
}

func TestExpressionParser_Associativity(t *testing.T) {
	parser := domain.NewSimpleExpressionParser()
	ctx := &domain.ExpressionContext{}

	tests := []struct {
		expression string
		expected   float64
	}{
		{"2 - 3 - 4", -5},                  // Left associative: (2-3)-4
		{"2 / 3 / 4", 0.16666666666666666}, // Left associative: (2/3)/4
		{"2 ^ 3 ^ 2", 512},                 // Right associative: 2^(3^2)
	}

	for _, test := range tests {
		t.Run(test.expression, func(t *testing.T) {
			result, err := parser.Evaluate(test.expression, ctx)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !result.Success {
				t.Fatalf("Evaluation failed: %s", result.Error)
			}

			actual, ok := result.Value.(float64)
			if !ok {
				t.Fatalf("Expected float64 value, got %T", result.Value)
			}

			// Use approximate equality for floating point
			if abs(actual-test.expected) > 0.0001 {
				t.Fatalf("Expected %f, got %f", test.expected, actual)
			}
		})
	}
}

func TestExpressionParser_Performance(t *testing.T) {
	parser := domain.NewSimpleExpressionParser()
	ctx := &domain.ExpressionContext{
		Self: map[string]interface{}{
			"level":    5.0,
			"strength": 18.0,
		},
	}

	expression := "level + strength * roll(\"1d6\") + min(health, 100)"

	// Warm up
	for i := 0; i < 10; i++ {
		_, _ = parser.Evaluate(expression, ctx)
	}

	// Measure performance
	start := time.Now()
	iterations := 1000

	for i := 0; i < iterations; i++ {
		result, err := parser.Evaluate(expression, ctx)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !result.Success {
			t.Fatalf("Evaluation failed: %s", result.Error)
		}
	}

	duration := time.Since(start)
	avgTime := duration / time.Duration(iterations)

	t.Logf("Average evaluation time: %v", avgTime)

	// Performance should be reasonable (less than 1ms per evaluation)
	if avgTime > time.Millisecond {
		t.Errorf("Performance too slow: %v per evaluation", avgTime)
	}
}

// Helper function for approximate equality
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
