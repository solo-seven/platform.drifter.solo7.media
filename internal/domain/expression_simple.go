package domain

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// SimpleExpressionParser provides a simpler approach to expression parsing
type SimpleExpressionParser struct {
	functions map[string]ExpressionFunction
}

// NewSimpleExpressionParser creates a new simple expression parser
func NewSimpleExpressionParser() *SimpleExpressionParser {
	parser := &SimpleExpressionParser{
		functions: make(map[string]ExpressionFunction),
	}

	parser.initializeBuiltinFunctions()
	return parser
}

// initializeBuiltinFunctions sets up the core game functions
func (sep *SimpleExpressionParser) initializeBuiltinFunctions() {
	// Dice rolling functions
	sep.functions["roll"] = ExpressionFunction{
		Name:        "roll",
		Description: "Roll dice with specified notation (e.g., '2d6', '1d20+3')",
		Parameters: []ParameterDefinition{
			{Name: "notation", Type: "string", Required: true},
		},
		Handler: sep.rollDice,
	}

	sep.functions["d"] = ExpressionFunction{
		Name:        "d",
		Description: "Roll a single die (e.g., d(6) for 1d6)",
		Parameters: []ParameterDefinition{
			{Name: "sides", Type: "number", Required: true},
		},
		Handler: sep.rollSingleDie,
	}

	// Combat functions
	sep.functions["deal"] = ExpressionFunction{
		Name:        "deal",
		Description: "Deal damage to a target",
		Parameters: []ParameterDefinition{
			{Name: "target", Type: "any", Required: true},
			{Name: "amount", Type: "number", Required: true},
			{Name: "type", Type: "string", Required: false, Default: "physical"},
		},
		Handler: sep.dealDamage,
	}

	sep.functions["heal"] = ExpressionFunction{
		Name:        "heal",
		Description: "Heal a target",
		Parameters: []ParameterDefinition{
			{Name: "target", Type: "any", Required: true},
			{Name: "amount", Type: "number", Required: true},
		},
		Handler: sep.healTarget,
	}

	// Condition functions
	sep.functions["has_tag"] = ExpressionFunction{
		Name:        "has_tag",
		Description: "Check if target has a specific tag",
		Parameters: []ParameterDefinition{
			{Name: "target", Type: "any", Required: true},
			{Name: "tag", Type: "string", Required: true},
		},
		Handler: sep.hasTag,
	}

	sep.functions["has_ability"] = ExpressionFunction{
		Name:        "has_ability",
		Description: "Check if target has a specific ability",
		Parameters: []ParameterDefinition{
			{Name: "target", Type: "any", Required: true},
			{Name: "ability", Type: "string", Required: true},
		},
		Handler: sep.hasAbility,
	}

	// Math functions
	sep.functions["min"] = ExpressionFunction{
		Name:        "min",
		Description: "Return the minimum of two values",
		Parameters: []ParameterDefinition{
			{Name: "a", Type: "number", Required: true},
			{Name: "b", Type: "number", Required: true},
		},
		Handler: sep.minFunction,
	}

	sep.functions["max"] = ExpressionFunction{
		Name:        "max",
		Description: "Return the maximum of two values",
		Parameters: []ParameterDefinition{
			{Name: "a", Type: "number", Required: true},
			{Name: "b", Type: "number", Required: true},
		},
		Handler: sep.maxFunction,
	}

	sep.functions["clamp"] = ExpressionFunction{
		Name:        "clamp",
		Description: "Clamp a value between min and max",
		Parameters: []ParameterDefinition{
			{Name: "value", Type: "number", Required: true},
			{Name: "min", Type: "number", Required: true},
			{Name: "max", Type: "number", Required: true},
		},
		Handler: sep.clampFunction,
	}
}

// Evaluate parses and evaluates an expression string using a simple approach
func (sep *SimpleExpressionParser) Evaluate(expression string, ctx *ExpressionContext) (*ExpressionResult, error) {
	if strings.TrimSpace(expression) == "" {
		return &ExpressionResult{
			Value:   nil,
			Type:    "null",
			Success: true,
		}, nil
	}

	// Check if this is a single function call that should preserve its type
	if sep.isSingleFunctionCall(expression) {
		return sep.evaluateSingleFunctionCall(expression, ctx)
	}

	// First, handle function calls
	expression = sep.expandFunctionCalls(expression, ctx)

	// Replace variables before making routing decision
	expression = sep.ReplaceVariables(expression, ctx)

	// Check if this is a string operation (contains quotes or comparison operators)
	if strings.Contains(expression, "\"") || strings.Contains(expression, "==") || strings.Contains(expression, "!=") {
		return sep.evaluateStringExpression(expression, ctx)
	}

	// Then evaluate the resulting expression
	return sep.evaluateSimpleExpression(expression, ctx)
}

// expandFunctionCalls expands function calls in the expression
func (sep *SimpleExpressionParser) expandFunctionCalls(expression string, ctx *ExpressionContext) string {
	// Simple approach: find function calls and replace them one by one
	// This avoids infinite loops by processing from left to right

	for {
		originalExpression := expression

		// Find the first function call
		for funcName := range sep.functions {
			pattern := funcName + "("
			if strings.Contains(expression, pattern) {
				start := strings.Index(expression, pattern)
				if start == -1 {
					continue
				}

				// Find the matching closing parenthesis
				parenCount := 0
				end := start + len(pattern) - 1 // Start after the opening parenthesis
				for i := end; i < len(expression); i++ {
					if expression[i] == '(' {
						parenCount++
					} else if expression[i] == ')' {
						parenCount--
						if parenCount == 0 {
							end = i
							break
						}
					}
				}

				if parenCount != 0 {
					continue // Mismatched parentheses, skip this function
				}

				// Extract the full function call
				fullMatch := expression[start : end+1]

				// Extract function name and arguments
				funcName := strings.TrimSpace(strings.Split(fullMatch, "(")[0])
				argsStr := strings.TrimSuffix(strings.Split(fullMatch, "(")[1], ")")

				// Parse arguments
				args := sep.parseArguments(argsStr, ctx)

				// Call the function
				function := sep.functions[funcName]
				result, err := function.Handler(args, ctx)
				if err != nil {
					// Replace with 0 on error
					expression = expression[:start] + "0" + expression[end+1:]
				} else {
					// For dice rolls, we need to preserve the type information
					// Store the result in a way that can be retrieved later
					if result.Type == "dice_roll" {
						// For dice rolls, we'll store the result and return the numeric value
						// The type information is preserved in the result
						expression = expression[:start] + fmt.Sprintf("%.0f", result.Value) + expression[end+1:]
					} else {
						// Convert result to string
						var resultStr string
						switch v := result.Value.(type) {
						case float64:
							resultStr = fmt.Sprintf("%.0f", v)
						case bool:
							if v {
								resultStr = "1"
							} else {
								resultStr = "0"
							}
						default:
							resultStr = fmt.Sprintf("%v", v)
						}
						expression = expression[:start] + resultStr + expression[end+1:]
					}
				}

				// Process this function call and break to start over
				break
			}
		}

		// If no changes were made, we're done
		if expression == originalExpression {
			break
		}
	}

	return expression
}

// isSingleFunctionCall checks if the expression is a single function call
func (sep *SimpleExpressionParser) isSingleFunctionCall(expression string) bool {
	expression = strings.TrimSpace(expression)

	// Check if it starts with a function name and ends with )
	for funcName := range sep.functions {
		if strings.HasPrefix(expression, funcName+"(") && strings.HasSuffix(expression, ")") {
			// Make sure there are no other operators outside the function call
			// by checking if the function call is the entire expression
			return true
		}
	}
	return false
}

// evaluateSingleFunctionCall evaluates a single function call and preserves its type
func (sep *SimpleExpressionParser) evaluateSingleFunctionCall(expression string, ctx *ExpressionContext) (*ExpressionResult, error) {
	expression = strings.TrimSpace(expression)

	// Find the function name
	for funcName := range sep.functions {
		if strings.HasPrefix(expression, funcName+"(") {
			// Extract arguments
			argsStr := strings.TrimSuffix(strings.TrimPrefix(expression, funcName+"("), ")")
			args := sep.parseArguments(argsStr, ctx)

			// Call the function
			function := sep.functions[funcName]
			return function.Handler(args, ctx)
		}
	}

	// Fall back to regular evaluation
	return sep.evaluateSimpleExpression(expression, ctx)
}

// evaluateStringExpression handles string operations
func (sep *SimpleExpressionParser) evaluateStringExpression(expression string, ctx *ExpressionContext) (*ExpressionResult, error) {
	// Handle string concatenation
	if strings.Contains(expression, "+") {
		// Simple string concatenation for expressions like "hello" + " " + "world"
		parts := strings.Split(expression, "+")
		var result strings.Builder

		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "\"") && strings.HasSuffix(part, "\"") {
				// Remove quotes and add to result
				content := part[1 : len(part)-1]
				result.WriteString(content)
			} else {
				// Try to evaluate as a variable or expression
				value := sep.GetContextValue(part, ctx)
				if value != nil {
					result.WriteString(fmt.Sprintf("%v", value))
				} else {
					result.WriteString(part)
				}
			}
		}

		return &ExpressionResult{
			Value:   result.String(),
			Type:    "string",
			Success: true,
		}, nil
	}

	// Handle string comparisons
	if strings.Contains(expression, "==") {
		parts := strings.Split(expression, "==")
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])

			// Remove quotes for comparison
			left = strings.Trim(left, "\"")
			right = strings.Trim(right, "\"")

			return &ExpressionResult{
				Value:   left == right,
				Type:    "boolean",
				Success: true,
			}, nil
		}
	}

	if strings.Contains(expression, "!=") {
		parts := strings.Split(expression, "!=")
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])

			// Remove quotes for comparison
			left = strings.Trim(left, "\"")
			right = strings.Trim(right, "\"")

			return &ExpressionResult{
				Value:   left != right,
				Type:    "boolean",
				Success: true,
			}, nil
		}
	}

	// Single string literal
	if strings.HasPrefix(expression, "\"") && strings.HasSuffix(expression, "\"") {
		content := expression[1 : len(expression)-1]
		return &ExpressionResult{
			Value:   content,
			Type:    "string",
			Success: true,
		}, nil
	}

	// Fall back to simple expression evaluation
	return sep.evaluateSimpleExpression(expression, ctx)
}

// parseArguments parses function arguments from a string
func (sep *SimpleExpressionParser) parseArguments(argsStr string, ctx *ExpressionContext) []interface{} {
	if strings.TrimSpace(argsStr) == "" {
		return []interface{}{}
	}

	var args []interface{}
	parts := sep.splitArguments(argsStr)

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Try to parse as string literal first (remove quotes)
		if strings.HasPrefix(part, "\"") && strings.HasSuffix(part, "\"") {
			args = append(args, part[1:len(part)-1])
			continue
		}

		// Try to parse as boolean
		if part == "true" {
			args = append(args, true)
			continue
		}
		if part == "false" {
			args = append(args, false)
			continue
		}

		// Try to parse as number
		if num, err := strconv.ParseFloat(part, 64); err == nil {
			args = append(args, num)
			continue
		}

		// Check if this part contains a function call or is a complex expression
		if sep.containsFunctionCall(part) || sep.isComplexExpression(part) {
			// Evaluate the expression with the provided context
			result, err := sep.Evaluate(part, ctx)
			if err == nil && result.Success {
				// Convert the result to the appropriate type for function arguments
				if result.Type == "dice_roll" {
					// For dice rolls, use the numeric value
					if val, ok := result.Value.(float64); ok {
						args = append(args, val)
					} else {
						args = append(args, result.Value)
					}
				} else {
					args = append(args, result.Value)
				}
			} else {
				// If evaluation fails, treat as string
				args = append(args, part)
			}
			continue
		}

		// Check if this is a variable reference
		value := sep.GetContextValue(part, ctx)
		if value != nil {
			args = append(args, value)
			continue
		}

		// Default to string
		args = append(args, part)
	}

	return args
}

// splitArguments splits function arguments, handling nested parentheses and quotes
func (sep *SimpleExpressionParser) splitArguments(argsStr string) []string {
	var parts []string
	var current strings.Builder
	parenDepth := 0
	inQuotes := false

	for _, char := range argsStr {
		switch char {
		case '"':
			inQuotes = !inQuotes
			current.WriteRune(char)
		case '(':
			if !inQuotes {
				parenDepth++
			}
			current.WriteRune(char)
		case ')':
			if !inQuotes {
				parenDepth--
			}
			current.WriteRune(char)
		case ',':
			if !inQuotes && parenDepth == 0 {
				parts = append(parts, current.String())
				current.Reset()
			} else {
				current.WriteRune(char)
			}
		default:
			current.WriteRune(char)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// containsFunctionCall checks if a string contains a function call
func (sep *SimpleExpressionParser) containsFunctionCall(expr string) bool {
	for funcName := range sep.functions {
		if strings.Contains(expr, funcName+"(") {
			return true
		}
	}
	return false
}

// isComplexExpression checks if a string is a complex expression
func (sep *SimpleExpressionParser) isComplexExpression(expr string) bool {
	// Check for operators that indicate a complex expression
	operators := []string{"+", "-", "*", "/", ">", "<", ">=", "<=", "==", "!=", "&&", "||"}
	for _, op := range operators {
		if strings.Contains(expr, op) {
			return true
		}
	}
	return false
}

// GetContextValue gets a value from the expression context
func (sep *SimpleExpressionParser) GetContextValue(key string, ctx *ExpressionContext) interface{} {
	// Handle dot notation (e.g., target.armor, self.level)
	if strings.Contains(key, ".") {
		parts := strings.Split(key, ".")
		if len(parts) != 2 {
			return nil // Only support one level of nesting for now
		}

		parentKey := parts[0]
		childKey := parts[1]

		// Map context section names to their actual context maps
		var contextMap map[string]interface{}
		switch parentKey {
		case "self":
			contextMap = ctx.Self
		case "target":
			contextMap = ctx.Target
		case "party":
			contextMap = ctx.Party
		case "terrain":
			contextMap = ctx.Terrain
		case "game":
			contextMap = ctx.Game
		default:
			// If it's not a known context section, check event data
			if ctx.EventData != nil {
				if parentValue, exists := ctx.EventData[parentKey]; exists {
					if parentMap, ok := parentValue.(map[string]interface{}); ok {
						if childValue, exists := parentMap[childKey]; exists {
							return childValue
						}
					}
				}
			}
			return nil
		}

		// Look for the child key in the appropriate context map
		if contextMap != nil {
			if childValue, exists := contextMap[childKey]; exists {
				return childValue
			}
		}

		return nil
	}

	// Handle simple variable names (no dot notation)
	// Check context variables in order: self, target, party, terrain, game, event data
	contexts := []map[string]interface{}{
		ctx.Self, ctx.Target, ctx.Party, ctx.Terrain, ctx.Game,
	}

	for _, context := range contexts {
		if value, exists := context[key]; exists {
			return value
		}
	}

	// Check event data if available
	if ctx.EventData != nil {
		if value, exists := ctx.EventData[key]; exists {
			return value
		}
	}

	return nil
}

// evaluateSimpleExpression evaluates a simple mathematical expression
func (sep *SimpleExpressionParser) evaluateSimpleExpression(expression string, ctx *ExpressionContext) (*ExpressionResult, error) {
	// Variables are already replaced in the main Evaluate function

	// Check if this is a string expression (contains quotes or is a single string)
	if strings.HasPrefix(expression, "\"") && strings.HasSuffix(expression, "\"") {
		content := expression[1 : len(expression)-1]
		return &ExpressionResult{
			Value:   content,
			Type:    "string",
			Success: true,
		}, nil
	}

	// Check if this is a boolean expression
	isBoolean := strings.Contains(expression, "&&") || strings.Contains(expression, "||") ||
		strings.Contains(expression, "==") || strings.Contains(expression, "!=") ||
		strings.Contains(expression, ">") || strings.Contains(expression, "<") ||
		strings.Contains(expression, ">=") || strings.Contains(expression, "<=")

	// Evaluate the expression
	result, err := sep.evaluateMathExpression(expression)
	if err != nil {
		return &ExpressionResult{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Determine result type and value
	if isBoolean {
		boolValue := result != 0
		return &ExpressionResult{
			Value:   boolValue,
			Type:    "boolean",
			Success: true,
		}, nil
	}

	return &ExpressionResult{
		Value:   result,
		Type:    "number",
		Success: true,
	}, nil
}

// ReplaceVariables replaces variable references with their values
func (sep *SimpleExpressionParser) ReplaceVariables(expression string, ctx *ExpressionContext) string {
	// Find variable references including dot notation (e.g., target.armor, self.level)
	re := regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*(?:\.[a-zA-Z_][a-zA-Z0-9_]*)*\b`)

	return re.ReplaceAllStringFunc(expression, func(match string) string {
		// Skip if it's a boolean literal
		if match == "true" || match == "false" {
			return match
		}

		// Get value from context (handles both simple and dot notation)
		value := sep.GetContextValue(match, ctx)
		if value != nil {
			switch v := value.(type) {
			case float64:
				return fmt.Sprintf("%.0f", v)
			case bool:
				if v {
					return "1"
				}
				return "0"
			case string:
				// For string values, we need to handle them differently
				// If this is a string comparison or concatenation, keep as string
				if strings.Contains(expression, "==") || strings.Contains(expression, "!=") ||
					strings.Contains(expression, "+") {
					return fmt.Sprintf("\"%s\"", v)
				}
				// Otherwise, try to convert to number if possible
				if num, err := strconv.ParseFloat(v, 64); err == nil {
					return fmt.Sprintf("%.0f", num)
				}
				return fmt.Sprintf("\"%s\"", v)
			default:
				return fmt.Sprintf("%v", v)
			}
		}

		// Return original if not found
		return match
	})
}

// evaluateMathExpression evaluates a mathematical expression
func (sep *SimpleExpressionParser) evaluateMathExpression(expression string) (float64, error) {
	// Simple evaluation using Go's expression evaluation
	// This is a simplified version - in production you'd want a proper math parser

	// Handle boolean literals
	expression = strings.ReplaceAll(expression, "true", "1")
	expression = strings.ReplaceAll(expression, "false", "0")

	// Handle boolean operations first (they have lower precedence than comparisons)
	if strings.Contains(expression, "&&") || strings.Contains(expression, "||") {
		return sep.evaluateBooleanOperation(expression)
	}

	// Handle comparison operators (they have higher precedence than boolean ops)
	if strings.Contains(expression, ">=") || strings.Contains(expression, "<=") ||
		strings.Contains(expression, ">") || strings.Contains(expression, "<") ||
		strings.Contains(expression, "==") || strings.Contains(expression, "!=") {
		return sep.evaluateComparison(expression)
	}

	// Handle basic arithmetic
	return sep.evaluateArithmetic(expression)
}

// evaluateComparison evaluates comparison operations
func (sep *SimpleExpressionParser) evaluateComparison(expression string) (float64, error) {
	// Handle >= first (before >)
	if strings.Contains(expression, ">=") {
		parts := strings.Split(expression, ">=")
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])

			leftNum, err1 := strconv.ParseFloat(left, 64)
			rightNum, err2 := strconv.ParseFloat(right, 64)

			if err1 == nil && err2 == nil {
				if leftNum >= rightNum {
					return 1, nil
				}
				return 0, nil
			}
		}
	}

	// Handle <= first (before <)
	if strings.Contains(expression, "<=") {
		parts := strings.Split(expression, "<=")
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])

			leftNum, err1 := strconv.ParseFloat(left, 64)
			rightNum, err2 := strconv.ParseFloat(right, 64)

			if err1 == nil && err2 == nil {
				if leftNum <= rightNum {
					return 1, nil
				}
				return 0, nil
			}
		}
	}

	// Handle >
	if strings.Contains(expression, ">") {
		parts := strings.Split(expression, ">")
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])

			leftNum, err1 := strconv.ParseFloat(left, 64)
			rightNum, err2 := strconv.ParseFloat(right, 64)

			if err1 == nil && err2 == nil {
				if leftNum > rightNum {
					return 1, nil
				}
				return 0, nil
			}
		}
	}

	// Handle <
	if strings.Contains(expression, "<") {
		parts := strings.Split(expression, "<")
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])

			leftNum, err1 := strconv.ParseFloat(left, 64)
			rightNum, err2 := strconv.ParseFloat(right, 64)

			if err1 == nil && err2 == nil {
				if leftNum < rightNum {
					return 1, nil
				}
				return 0, nil
			}
		}
	}

	// Handle ==
	if strings.Contains(expression, "==") {
		parts := strings.Split(expression, "==")
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])

			// Try numeric comparison first
			leftNum, err1 := strconv.ParseFloat(left, 64)
			rightNum, err2 := strconv.ParseFloat(right, 64)

			if err1 == nil && err2 == nil {
				if leftNum == rightNum {
					return 1, nil
				}
				return 0, nil
			}

			// Fall back to string comparison
			left = strings.Trim(left, "\"")
			right = strings.Trim(right, "\"")

			if left == right {
				return 1, nil
			}
			return 0, nil
		}
	}

	// Handle !=
	if strings.Contains(expression, "!=") {
		parts := strings.Split(expression, "!=")
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])

			// Try numeric comparison first
			leftNum, err1 := strconv.ParseFloat(left, 64)
			rightNum, err2 := strconv.ParseFloat(right, 64)

			if err1 == nil && err2 == nil {
				if leftNum != rightNum {
					return 1, nil
				}
				return 0, nil
			}

			// Fall back to string comparison
			left = strings.Trim(left, "\"")
			right = strings.Trim(right, "\"")

			if left != right {
				return 1, nil
			}
			return 0, nil
		}
	}

	return 0, fmt.Errorf("invalid comparison expression: %s", expression)
}

// evaluateBooleanOperation evaluates boolean operations
func (sep *SimpleExpressionParser) evaluateBooleanOperation(expression string) (float64, error) {
	// Handle && operations (higher precedence than ||)
	for strings.Contains(expression, "&&") {
		// Find the rightmost && to handle left associativity
		lastAndIndex := strings.LastIndex(expression, "&&")
		if lastAndIndex == -1 {
			break
		}

		left := strings.TrimSpace(expression[:lastAndIndex])
		right := strings.TrimSpace(expression[lastAndIndex+2:])

		// Evaluate left and right sides
		leftResult, err1 := sep.evaluateMathExpression(left)
		rightResult, err2 := sep.evaluateMathExpression(right)

		if err1 != nil || err2 != nil {
			return 0, fmt.Errorf("error evaluating boolean operation: %v, %v", err1, err2)
		}

		// Convert to boolean: 0 is false, anything else is true
		leftBool := leftResult != 0
		rightBool := rightResult != 0

		result := 0.0
		if leftBool && rightBool {
			result = 1.0
		}

		// Replace the entire expression with the result
		expression = fmt.Sprintf("%.0f", result)
	}

	// Handle || operations
	for strings.Contains(expression, "||") {
		// Find the rightmost || to handle left associativity
		lastOrIndex := strings.LastIndex(expression, "||")
		if lastOrIndex == -1 {
			break
		}

		left := strings.TrimSpace(expression[:lastOrIndex])
		right := strings.TrimSpace(expression[lastOrIndex+2:])

		// Evaluate left and right sides
		leftResult, err1 := sep.evaluateMathExpression(left)
		rightResult, err2 := sep.evaluateMathExpression(right)

		if err1 != nil || err2 != nil {
			return 0, fmt.Errorf("error evaluating boolean operation: %v, %v", err1, err2)
		}

		// Convert to boolean: 0 is false, anything else is true
		leftBool := leftResult != 0
		rightBool := rightResult != 0

		result := 0.0
		if leftBool || rightBool {
			result = 1.0
		}

		// Replace the entire expression with the result
		expression = fmt.Sprintf("%.0f", result)
	}

	// If we still have boolean operators, something went wrong
	if strings.Contains(expression, "&&") || strings.Contains(expression, "||") {
		return 0, fmt.Errorf("invalid boolean operation: %s", expression)
	}

	// Parse the final result
	result, err := strconv.ParseFloat(expression, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid boolean result: %s", expression)
	}

	return result, nil
}

// evaluateArithmetic evaluates arithmetic expressions
func (sep *SimpleExpressionParser) evaluateArithmetic(expression string) (float64, error) {
	// This is a very simplified arithmetic evaluator
	// In production, you'd want a proper expression parser

	// Basic validation - check for obviously invalid expressions
	if strings.Contains(expression, ")") && !strings.Contains(expression, "(") {
		return 0, fmt.Errorf("invalid expression: %s", expression)
	}
	if strings.Contains(expression, "(") && !strings.Contains(expression, ")") {
		return 0, fmt.Errorf("invalid expression: %s", expression)
	}

	// Handle parentheses first
	for strings.Contains(expression, "(") {
		start := strings.LastIndex(expression, "(")
		end := strings.Index(expression[start:], ")")
		if end == -1 {
			return 0, fmt.Errorf("mismatched parentheses")
		}
		end += start

		subExpr := expression[start+1 : end]
		result, err := sep.evaluateArithmetic(subExpr)
		if err != nil {
			return 0, err
		}

		expression = expression[:start] + fmt.Sprintf("%.0f", result) + expression[end+1:]
	}

	// Handle exponentiation (right associative) - process from right to left
	for strings.Contains(expression, "^") {
		// Find the rightmost ^ operator
		lastPowIndex := strings.LastIndex(expression, "^")
		if lastPowIndex == -1 {
			break
		}

		// Find the numbers around this operator
		re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*\^\s*(\d+(?:\.\d+)?)`)
		matches := re.FindAllStringIndex(expression, -1)

		// Find the match that contains our lastPowIndex
		var matchStart, matchEnd int
		for _, match := range matches {
			if match[0] <= lastPowIndex && lastPowIndex <= match[1] {
				matchStart = match[0]
				matchEnd = match[1]
				break
			}
		}

		if matchStart != 0 || matchEnd != 0 {
			// Extract and evaluate this specific match
			matchStr := expression[matchStart:matchEnd]
			parts := re.FindStringSubmatch(matchStr)
			if len(parts) == 3 {
				base, _ := strconv.ParseFloat(parts[1], 64)
				exp, _ := strconv.ParseFloat(parts[2], 64)
				result := math.Pow(base, exp)
				expression = expression[:matchStart] + fmt.Sprintf("%.0f", result) + expression[matchEnd:]
			}
		} else {
			break
		}
	}

	// Handle multiplication, division, and modulo (left associative)
	for strings.Contains(expression, "*") || strings.Contains(expression, "/") || strings.Contains(expression, "%") {
		// Multiplication
		re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*\*\s*(\d+(?:\.\d+)?)`)
		expression = re.ReplaceAllStringFunc(expression, func(match string) string {
			parts := re.FindStringSubmatch(match)
			if len(parts) == 3 {
				left, _ := strconv.ParseFloat(parts[1], 64)
				right, _ := strconv.ParseFloat(parts[2], 64)
				result := left * right
				return fmt.Sprintf("%.0f", result)
			}
			return match
		})

		// Division (left associative)
		re = regexp.MustCompile(`(\d+(?:\.\d+)?)\s*/\s*(\d+(?:\.\d+)?)`)
		expression = re.ReplaceAllStringFunc(expression, func(match string) string {
			parts := re.FindStringSubmatch(match)
			if len(parts) == 3 {
				left, _ := strconv.ParseFloat(parts[1], 64)
				right, _ := strconv.ParseFloat(parts[2], 64)
				if right == 0 {
					// Return a special marker for division by zero that will cause an error
					return "DIVISION_BY_ZERO"
				}
				result := left / right
				return fmt.Sprintf("%.6f", result) // Keep more precision for division
			}
			return match
		})

		// Modulo
		re = regexp.MustCompile(`(\d+(?:\.\d+)?)\s*%\s*(\d+(?:\.\d+)?)`)
		expression = re.ReplaceAllStringFunc(expression, func(match string) string {
			parts := re.FindStringSubmatch(match)
			if len(parts) == 3 {
				left, _ := strconv.ParseFloat(parts[1], 64)
				right, _ := strconv.ParseFloat(parts[2], 64)
				if right == 0 {
					return "0" // Modulo by zero
				}
				result := math.Mod(left, right)
				return fmt.Sprintf("%.0f", result)
			}
			return match
		})
	}

	// Handle addition and subtraction (left associative)
	for strings.Contains(expression, "+") || (strings.Contains(expression, "-") && !strings.HasPrefix(expression, "-")) {
		prev := expression

		// Addition
		re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*\+\s*(\d+(?:\.\d+)?)`)
		expression = re.ReplaceAllStringFunc(expression, func(match string) string {
			parts := re.FindStringSubmatch(match)
			if len(parts) == 3 {
				left, _ := strconv.ParseFloat(parts[1], 64)
				right, _ := strconv.ParseFloat(parts[2], 64)
				result := left + right
				return fmt.Sprintf("%.0f", result)
			}
			return match
		})

		// Subtraction (left associative) - handle negative numbers
		// Find the first subtraction operator that's not at the beginning
		firstMinusIndex := -1
		for i, char := range expression {
			if char == '-' && i > 0 {
				// Check if this is not part of a negative number
				if i > 0 && (expression[i-1] == ' ' || expression[i-1] == '+' || expression[i-1] == '-' || expression[i-1] == '*' || expression[i-1] == '/' || expression[i-1] == '%' || expression[i-1] == '^') {
					firstMinusIndex = i
					break
				}
			}
		}

		if firstMinusIndex != -1 {
			// Find the numbers around this operator
			re := regexp.MustCompile(`(-?\d+(?:\.\d+)?)\s*-\s*(\d+(?:\.\d+)?)`)
			matches := re.FindAllStringIndex(expression, -1)

			// Find the match that contains our firstMinusIndex
			var matchStart, matchEnd int
			for _, match := range matches {
				if match[0] <= firstMinusIndex && firstMinusIndex <= match[1] {
					matchStart = match[0]
					matchEnd = match[1]
					break
				}
			}

			if matchStart != 0 || matchEnd != 0 {
				// Extract and evaluate this specific match
				matchStr := expression[matchStart:matchEnd]
				parts := re.FindStringSubmatch(matchStr)
				if len(parts) == 3 {
					left, _ := strconv.ParseFloat(parts[1], 64)
					right, _ := strconv.ParseFloat(parts[2], 64)
					result := left - right
					expression = expression[:matchStart] + fmt.Sprintf("%.0f", result) + expression[matchEnd:]
				}
			}
		}

		// If no progress was made in this iteration, break to avoid infinite loop
		if expression == prev {
			break
		}
	}

	// If after processing we still have '+' or a '-' between numbers, expression is invalid
	if strings.Contains(expression, "+") || regexp.MustCompile(`\d\s*-\s*\d`).FindStringIndex(expression) != nil {
		// But allow negative numbers at the beginning
		if !strings.HasPrefix(expression, "-") {
			return 0, fmt.Errorf("invalid expression: %s", expression)
		}
	}

	// Check for division by zero marker
	if expression == "DIVISION_BY_ZERO" {
		return 0, fmt.Errorf("division by zero")
	}

	// Parse final result
	result, err := strconv.ParseFloat(expression, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid expression: %s", expression)
	}

	return result, nil
}

// Built-in function implementations (same as before)

func (sep *SimpleExpressionParser) rollDice(args []interface{}, ctx *ExpressionContext) (*ExpressionResult, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("roll() expects exactly 1 argument")
	}

	notation, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("roll() argument must be a string")
	}

	result, err := sep.parseDiceNotation(notation)
	if err != nil {
		return nil, err
	}

	return &ExpressionResult{
		Value:   float64(result.Total),
		Type:    "dice_roll",
		Success: true,
		Metadata: map[string]interface{}{
			"notation": notation,
			"rolls":    result.Rolls,
			"modifier": result.Modifier,
		},
	}, nil
}

func (sep *SimpleExpressionParser) rollSingleDie(args []interface{}, ctx *ExpressionContext) (*ExpressionResult, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("d() expects exactly 1 argument")
	}

	sides, ok := args[0].(float64)
	if !ok {
		return nil, fmt.Errorf("d() argument must be a number")
	}

	if sides < 1 {
		return nil, fmt.Errorf("die must have at least 1 side")
	}

	roll := sep.randomInt(1, int(sides))

	return &ExpressionResult{
		Value:   float64(roll),
		Type:    "dice_roll",
		Success: true,
		Metadata: map[string]interface{}{
			"notation": fmt.Sprintf("1d%.0f", sides),
			"rolls":    []int{roll},
			"modifier": 0,
		},
	}, nil
}

func (sep *SimpleExpressionParser) dealDamage(args []interface{}, ctx *ExpressionContext) (*ExpressionResult, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("deal() expects at least 2 arguments")
	}

	amount, ok := args[1].(float64)
	if !ok {
		return nil, fmt.Errorf("deal() amount must be a number")
	}

	damageType := "physical"
	if len(args) > 2 {
		if dt, ok := args[2].(string); ok {
			damageType = dt
		}
	}

	return &ExpressionResult{
		Value:   amount,
		Type:    "number",
		Success: true,
		Metadata: map[string]interface{}{
			"action":      "deal_damage",
			"target":      args[0],
			"amount":      amount,
			"damage_type": damageType,
		},
	}, nil
}

func (sep *SimpleExpressionParser) healTarget(args []interface{}, ctx *ExpressionContext) (*ExpressionResult, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("heal() expects exactly 2 arguments")
	}

	amount, ok := args[1].(float64)
	if !ok {
		return nil, fmt.Errorf("heal() amount must be a number")
	}

	return &ExpressionResult{
		Value:   amount,
		Type:    "number",
		Success: true,
		Metadata: map[string]interface{}{
			"action": "heal",
			"target": args[0],
			"amount": amount,
		},
	}, nil
}

func (sep *SimpleExpressionParser) hasTag(args []interface{}, ctx *ExpressionContext) (*ExpressionResult, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("has_tag() expects exactly 2 arguments")
	}

	return &ExpressionResult{
		Value:   false,
		Type:    "boolean",
		Success: true,
		Metadata: map[string]interface{}{
			"action": "check_tag",
			"target": args[0],
			"tag":    args[1],
		},
	}, nil
}

func (sep *SimpleExpressionParser) hasAbility(args []interface{}, ctx *ExpressionContext) (*ExpressionResult, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("has_ability() expects exactly 2 arguments")
	}

	return &ExpressionResult{
		Value:   false,
		Type:    "boolean",
		Success: true,
		Metadata: map[string]interface{}{
			"action":  "check_ability",
			"target":  args[0],
			"ability": args[1],
		},
	}, nil
}

func (sep *SimpleExpressionParser) minFunction(args []interface{}, ctx *ExpressionContext) (*ExpressionResult, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("min() expects exactly 2 arguments")
	}

	a, ok1 := args[0].(float64)
	b, ok2 := args[1].(float64)

	if !ok1 || !ok2 {
		return nil, fmt.Errorf("min() arguments must be numbers")
	}

	result := a
	if b < a {
		result = b
	}

	return &ExpressionResult{Value: result, Type: "number", Success: true}, nil
}

func (sep *SimpleExpressionParser) maxFunction(args []interface{}, ctx *ExpressionContext) (*ExpressionResult, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("max() expects exactly 2 arguments")
	}

	a, ok1 := args[0].(float64)
	b, ok2 := args[1].(float64)

	if !ok1 || !ok2 {
		return nil, fmt.Errorf("max() arguments must be numbers")
	}

	result := a
	if b > a {
		result = b
	}

	return &ExpressionResult{Value: result, Type: "number", Success: true}, nil
}

func (sep *SimpleExpressionParser) clampFunction(args []interface{}, ctx *ExpressionContext) (*ExpressionResult, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("clamp() expects exactly 3 arguments")
	}

	value, ok1 := args[0].(float64)
	min, ok2 := args[1].(float64)
	max, ok3 := args[2].(float64)

	if !ok1 || !ok2 || !ok3 {
		return nil, fmt.Errorf("clamp() arguments must be numbers")
	}

	if min > max {
		return nil, fmt.Errorf("clamp() min cannot be greater than max")
	}

	result := value
	if result < min {
		result = min
	}
	if result > max {
		result = max
	}

	return &ExpressionResult{Value: result, Type: "number", Success: true}, nil
}

// Helper functions

func (sep *SimpleExpressionParser) parseDiceNotation(notation string) (*DiceResult, error) {
	// Parse dice notation like "2d6+3", "1d20-1", "3d4"
	re := regexp.MustCompile(`^(\d+)d(\d+)([+-]\d+)?$`)
	matches := re.FindStringSubmatch(notation)

	if len(matches) == 0 {
		return nil, fmt.Errorf("invalid dice notation: %s", notation)
	}

	count, _ := strconv.Atoi(matches[1])
	sides, _ := strconv.Atoi(matches[2])
	modifier := 0

	if len(matches) > 3 && matches[3] != "" {
		modifier, _ = strconv.Atoi(matches[3])
	}

	if count < 1 || sides < 1 {
		return nil, fmt.Errorf("invalid dice count or sides")
	}

	// Roll the dice
	var rolls []int
	total := modifier

	for i := 0; i < count; i++ {
		roll := sep.randomInt(1, sides)
		rolls = append(rolls, roll)
		total += roll
	}

	return &DiceResult{
		Notation: notation,
		Count:    count,
		Sides:    sides,
		Modifier: modifier,
		Rolls:    rolls,
		Total:    total,
	}, nil
}

func (sep *SimpleExpressionParser) randomInt(min, max int) int {
	// Simple pseudo-random number generator
	// In a real implementation, you'd use a proper RNG with seeding
	return min + (int(time.Now().UnixNano()) % (max - min + 1))
}
