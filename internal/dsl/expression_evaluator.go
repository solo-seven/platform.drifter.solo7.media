package dsl

import (
	"context"
	"fmt"
	"time"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
)

// ExpressionEvaluator handles expression evaluation using expr-lang
type ExpressionEvaluator struct {
	logger domain.Logger
}

// NewExpressionEvaluator creates a new expression evaluator
func NewExpressionEvaluator(logger domain.Logger) *ExpressionEvaluator {
	return &ExpressionEvaluator{
		logger: logger,
	}
}

// Evaluate evaluates an expression with the given context
func (e *ExpressionEvaluator) Evaluate(ctx context.Context, expression string, context map[string]interface{}) (interface{}, error) {
	// Add common functions to context
	context["roll"] = e.rollDice
	context["max"] = e.max
	context["min"] = e.min
	context["now"] = time.Now
	context["time"] = map[string]interface{}{
		"is_night": e.isNight,
		"is_day":   e.isDay,
	}

	program, err := expr.Compile(expression, expr.Env(context))
	if err != nil {
		return nil, fmt.Errorf("failed to compile expression: %w", err)
	}

	result, err := expr.Run(program, context)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate expression: %w", err)
	}

	return result, nil
}

// Compile compiles an expression for later evaluation
func (e *ExpressionEvaluator) Compile(ctx context.Context, expression string) (domain.CompiledExpression, error) {
	program, err := expr.Compile(expression)
	if err != nil {
		return nil, fmt.Errorf("failed to compile expression: %w", err)
	}

	return &CompiledExpressionImpl{
		program: program,
	}, nil
}

// CompiledExpressionImpl implements CompiledExpression
type CompiledExpressionImpl struct {
	program *vm.Program
}

// Evaluate evaluates the compiled expression
func (c *CompiledExpressionImpl) Evaluate(ctx context.Context, context map[string]interface{}) (interface{}, error) {
	// Add common functions to context
	context["roll"] = func(dice string) int { return rollDice(dice) }
	context["max"] = func(a, b int) int {
		if a > b {
			return a
		}
		return b
	}
	context["min"] = func(a, b int) int {
		if a < b {
			return a
		}
		return b
	}
	context["now"] = time.Now
	context["time"] = map[string]interface{}{
		"is_night": func() bool { return isNight() },
		"is_day":   func() bool { return isDay() },
	}

	result, err := expr.Run(c.program, context)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate compiled expression: %w", err)
	}

	return result, nil
}

// Helper functions for expression evaluation

// rollDice rolls dice in the format "XdY" or "XdY+Z"
func (e *ExpressionEvaluator) rollDice(dice string) int {
	return rollDice(dice)
}

// rollDice is a standalone function for rolling dice
func rollDice(dice string) int {
	// Simple dice rolling implementation
	// Format: "XdY" or "XdY+Z" or "XdY-Z"
	// Examples: "1d6", "2d10+3", "1d20-1"

	// For now, return a simple random number
	// TODO: Implement proper dice rolling
	return 10 // Placeholder
}

// max returns the maximum of two integers
func (e *ExpressionEvaluator) max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the minimum of two integers
func (e *ExpressionEvaluator) min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// isNight checks if it's currently night time
func (e *ExpressionEvaluator) isNight() bool {
	return isNight()
}

// isNight is a standalone function
func isNight() bool {
	hour := time.Now().Hour()
	return hour < 6 || hour > 18
}

// isDay checks if it's currently day time
func (e *ExpressionEvaluator) isDay() bool {
	return isDay()
}

// isDay is a standalone function
func isDay() bool {
	hour := time.Now().Hour()
	return hour >= 6 && hour <= 18
}

// DSLInterpreter handles the complete DSL loading process
type DSLInterpreter struct {
	tomlParser          *TOMLParser
	expressionEvaluator *ExpressionEvaluator
	markdownParser      *MarkdownParser
	logger              domain.Logger
}

// NewDSLInterpreter creates a new DSL interpreter
func NewDSLInterpreter(logger domain.Logger) *DSLInterpreter {
	return &DSLInterpreter{
		tomlParser:          NewTOMLParser(logger),
		expressionEvaluator: NewExpressionEvaluator(logger),
		markdownParser:      NewMarkdownParser(logger),
		logger:              logger,
	}
}

// LoadWorld loads a complete world from DSL files
func (d *DSLInterpreter) LoadWorld(ctx context.Context, contentPath string) (*domain.World, error) {
	d.logger.Info("Loading world from DSL", map[string]interface{}{
		"content_path": contentPath,
	})

	// Parse TOML files
	world, err := d.tomlParser.ParseWorld(ctx, contentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TOML files: %w", err)
	}

	// TODO: Parse Markdown files for descriptions
	// TODO: Compile expressions in the world data
	// TODO: Resolve cross-references

	return world, nil
}

// ValidateContent validates DSL content
func (d *DSLInterpreter) ValidateContent(ctx context.Context, contentPath string) error {
	d.logger.Info("Validating DSL content", map[string]interface{}{
		"content_path": contentPath,
	})

	// Load world to validate
	_, err := d.LoadWorld(ctx, contentPath)
	if err != nil {
		return fmt.Errorf("content validation failed: %w", err)
	}

	// TODO: Add more validation rules
	// - Check for orphaned references
	// - Validate expression syntax
	// - Check for circular dependencies

	return nil
}

// ReloadContent reloads DSL content
func (d *DSLInterpreter) ReloadContent(ctx context.Context, contentPath string) (*domain.World, error) {
	d.logger.Info("Reloading DSL content", map[string]interface{}{
		"content_path": contentPath,
	})

	return d.LoadWorld(ctx, contentPath)
}
