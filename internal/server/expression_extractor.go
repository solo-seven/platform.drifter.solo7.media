package server

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/solo7.media/platform.drifter.solo7.media/internal/domain"
)

// ExpressionExtractorImpl implements the ExpressionExtractor interface
type ExpressionExtractorImpl struct {
	logger domain.Logger
}

// NewExpressionExtractor creates a new expression extractor
func NewExpressionExtractor(logger domain.Logger) domain.ExpressionExtractor {
	return &ExpressionExtractorImpl{
		logger: logger,
	}
}

// ExtractExpressions extracts expression strings from content data
func (ee *ExpressionExtractorImpl) ExtractExpressions(data map[string]interface{}) (map[string]string, error) {
	expressions := make(map[string]string)

	if data == nil {
		return expressions, nil
	}

	// Recursively search for expression patterns
	ee.extractExpressionsRecursive(data, "", expressions)

	ee.logger.Debug("Expressions extracted", map[string]interface{}{
		"count": len(expressions),
	})

	return expressions, nil
}

// ValidateExpression validates an expression string
func (ee *ExpressionExtractorImpl) ValidateExpression(expression string) error {
	if expression == "" {
		return fmt.Errorf("expression cannot be empty")
	}

	// Basic syntax validation
	if err := ee.validateBasicSyntax(expression); err != nil {
		return err
	}

	// Validate function calls
	if err := ee.validateFunctionCalls(expression); err != nil {
		return err
	}

	// Validate variable references
	if err := ee.validateVariableReferences(expression); err != nil {
		return err
	}

	// Validate dice notation
	if err := ee.validateDiceNotation(expression); err != nil {
		return err
	}

	return nil
}

// ParseExpression parses an expression string into an AST
func (ee *ExpressionExtractorImpl) ParseExpression(expression string) (interface{}, error) {
	if expression == "" {
		return nil, fmt.Errorf("expression cannot be empty")
	}

	// This is a placeholder implementation
	// In a real implementation, this would parse the expression into an AST
	// using the existing expression parser from the domain layer

	ee.logger.Debug("Expression parsed", map[string]interface{}{
		"expression": expression,
	})

	// Return a simple representation for now
	return map[string]interface{}{
		"type":      "expression",
		"value":     expression,
		"parsed_at": "placeholder",
	}, nil
}

// Helper methods

func (ee *ExpressionExtractorImpl) extractExpressionsRecursive(data interface{}, path string, expressions map[string]string) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			currentPath := key
			if path != "" {
				currentPath = path + "." + key
			}
			ee.extractExpressionsRecursive(value, currentPath, expressions)
		}
	case []interface{}:
		for i, value := range v {
			currentPath := fmt.Sprintf("%s[%d]", path, i)
			ee.extractExpressionsRecursive(value, currentPath, expressions)
		}
	case string:
		// Check if this looks like an expression
		if ee.isExpressionString(v) {
			expressions[path] = v
		}
	}
}

func (ee *ExpressionExtractorImpl) isExpressionString(str string) bool {
	// Simple heuristic to identify expression strings
	// Look for common expression patterns

	// Check for function calls
	if matched, _ := regexp.MatchString(`\w+\(`, str); matched {
		return true
	}

	// Check for dice notation
	if matched, _ := regexp.MatchString(`\d+d\d+`, str); matched {
		return true
	}

	// Check for variable references
	if matched, _ := regexp.MatchString(`\$\w+`, str); matched {
		return true
	}

	// Check for self/target references
	if matched, _ := regexp.MatchString(`(self|target|party|terrain)\.`, str); matched {
		return true
	}

	// Check for mathematical operations
	if matched, _ := regexp.MatchString(`[+\-*/]`, str); matched {
		return true
	}

	return false
}

func (ee *ExpressionExtractorImpl) validateBasicSyntax(expression string) error {
	// Check for balanced parentheses
	openCount := strings.Count(expression, "(")
	closeCount := strings.Count(expression, ")")
	if openCount != closeCount {
		return fmt.Errorf("unbalanced parentheses")
	}

	// Check for balanced brackets
	openBracketCount := strings.Count(expression, "[")
	closeBracketCount := strings.Count(expression, "]")
	if openBracketCount != closeBracketCount {
		return fmt.Errorf("unbalanced brackets")
	}

	// Check for balanced braces
	openBraceCount := strings.Count(expression, "{")
	closeBraceCount := strings.Count(expression, "}")
	if openBraceCount != closeBraceCount {
		return fmt.Errorf("unbalanced braces")
	}

	return nil
}

func (ee *ExpressionExtractorImpl) validateFunctionCalls(expression string) error {
	// Find all function calls
	functionCallRegex := regexp.MustCompile(`\w+\(`)
	matches := functionCallRegex.FindAllString(expression, -1)

	for _, match := range matches {
		functionName := strings.TrimSuffix(match, "(")

		// Check if function name is valid
		if !ee.isValidFunctionName(functionName) {
			return fmt.Errorf("invalid function name: %s", functionName)
		}
	}

	return nil
}

func (ee *ExpressionExtractorImpl) validateVariableReferences(expression string) error {
	// Find all variable references
	variableRegex := regexp.MustCompile(`\$\w+`)
	matches := variableRegex.FindAllString(expression, -1)

	for _, match := range matches {
		variableName := strings.TrimPrefix(match, "$")

		// Check if variable name is valid
		if !ee.isValidVariableName(variableName) {
			return fmt.Errorf("invalid variable name: %s", variableName)
		}
	}

	return nil
}

func (ee *ExpressionExtractorImpl) validateDiceNotation(expression string) error {
	// Find all dice notation
	diceRegex := regexp.MustCompile(`\d+d\d+`)
	matches := diceRegex.FindAllString(expression, -1)

	for _, match := range matches {
		// Validate dice notation format
		if !ee.isValidDiceNotation(match) {
			return fmt.Errorf("invalid dice notation: %s", match)
		}
	}

	return nil
}

func (ee *ExpressionExtractorImpl) isValidFunctionName(name string) bool {
	// List of allowed function names
	allowedFunctions := []string{
		"roll", "heal", "deal", "has_tag", "get_attribute", "set_attribute",
		"add", "subtract", "multiply", "divide", "mod", "min", "max",
		"abs", "floor", "ceil", "round", "random", "choose",
	}

	for _, allowed := range allowedFunctions {
		if name == allowed {
			return true
		}
	}

	return false
}

func (ee *ExpressionExtractorImpl) isValidVariableName(name string) bool {
	// Variable names should be alphanumeric and start with a letter
	if matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_]*$`, name); !matched {
		return false
	}

	// Check for reserved words
	reservedWords := []string{"self", "target", "party", "terrain", "context"}
	for _, reserved := range reservedWords {
		if name == reserved {
			return false
		}
	}

	return true
}

func (ee *ExpressionExtractorImpl) isValidDiceNotation(notation string) bool {
	// Dice notation should be in format: NdM where N is number of dice and M is sides
	// Examples: 1d6, 2d10, 1d20, etc.

	if matched, _ := regexp.MatchString(`^\d+d\d+$`, notation); !matched {
		return false
	}

	// Parse the notation
	parts := strings.Split(notation, "d")
	if len(parts) != 2 {
		return false
	}

	// Check if both parts are valid numbers
	if parts[0] == "0" || parts[1] == "0" {
		return false
	}

	return true
}
