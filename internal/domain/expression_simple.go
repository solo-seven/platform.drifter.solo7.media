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

	// First, handle function calls
	expression = sep.expandFunctionCalls(expression, ctx)

	// Then evaluate the resulting expression
	return sep.evaluateSimpleExpression(expression, ctx)
}

// expandFunctionCalls expands function calls in the expression
func (sep *SimpleExpressionParser) expandFunctionCalls(expression string, ctx *ExpressionContext) string {
	// Find function calls and replace them with their results
	for funcName := range sep.functions {
		pattern := fmt.Sprintf(`%s\s*\([^)]*\)`, funcName)
		re := regexp.MustCompile(pattern)

		expression = re.ReplaceAllStringFunc(expression, func(match string) string {
			// Extract function name and arguments
			funcName := strings.Split(match, "(")[0]
			argsStr := strings.TrimSuffix(strings.Split(match, "(")[1], ")")

			// Parse arguments
			args := sep.parseArguments(argsStr)

			// Call the function
			function := sep.functions[funcName]
			result, err := function.Handler(args, ctx)
			if err != nil {
				return "0" // Default to 0 on error
			}

			// Convert result to string
			switch v := result.Value.(type) {
			case float64:
				return fmt.Sprintf("%.0f", v)
			case bool:
				if v {
					return "1"
				}
				return "0"
			default:
				return fmt.Sprintf("%v", v)
			}
		})
	}

	return expression
}

// parseArguments parses function arguments from a string
func (sep *SimpleExpressionParser) parseArguments(argsStr string) []interface{} {
	if strings.TrimSpace(argsStr) == "" {
		return []interface{}{}
	}

	var args []interface{}
	parts := sep.splitArguments(argsStr)

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Try to parse as number
		if num, err := strconv.ParseFloat(part, 64); err == nil {
			args = append(args, num)
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

		// Try to parse as string (remove quotes)
		if strings.HasPrefix(part, "\"") && strings.HasSuffix(part, "\"") {
			args = append(args, part[1:len(part)-1])
			continue
		}

		// Try to get from context (we'll need to pass ctx here, but for now skip)
		// if value := sep.getContextValue(part, ctx); value != nil {
		//	args = append(args, value)
		//	continue
		// }

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

// getContextValue gets a value from the expression context
func (sep *SimpleExpressionParser) getContextValue(key string, ctx *ExpressionContext) interface{} {
	// Check context variables in order: self, target, party, terrain, game
	contexts := []map[string]interface{}{
		ctx.Self, ctx.Target, ctx.Party, ctx.Terrain, ctx.Game,
	}

	for _, context := range contexts {
		if value, exists := context[key]; exists {
			return value
		}
	}

	return nil
}

// evaluateSimpleExpression evaluates a simple mathematical expression
func (sep *SimpleExpressionParser) evaluateSimpleExpression(expression string, ctx *ExpressionContext) (*ExpressionResult, error) {
	// Replace variables with their values
	expression = sep.replaceVariables(expression, ctx)

	// Evaluate the expression
	result, err := sep.evaluateMathExpression(expression)
	if err != nil {
		return &ExpressionResult{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	return &ExpressionResult{
		Value:   result,
		Type:    "number",
		Success: true,
	}, nil
}

// replaceVariables replaces variable references with their values
func (sep *SimpleExpressionParser) replaceVariables(expression string, ctx *ExpressionContext) string {
	// Find variable references (identifiers that aren't numbers or operators)
	re := regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`)

	return re.ReplaceAllStringFunc(expression, func(match string) string {
		// Skip if it's a boolean literal
		if match == "true" || match == "false" {
			return match
		}

		// Get value from context
		value := sep.getContextValue(match, ctx)
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

	// Handle string comparisons
	if strings.Contains(expression, "==") || strings.Contains(expression, "!=") {
		return sep.evaluateComparison(expression)
	}

	// Handle boolean operations
	if strings.Contains(expression, "&&") || strings.Contains(expression, "||") {
		return sep.evaluateBooleanOperation(expression)
	}

	// Handle basic arithmetic
	return sep.evaluateArithmetic(expression)
}

// evaluateComparison evaluates comparison operations
func (sep *SimpleExpressionParser) evaluateComparison(expression string) (float64, error) {
	if strings.Contains(expression, "==") {
		parts := strings.Split(expression, "==")
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])

			// Remove quotes for string comparison
			left = strings.Trim(left, "\"")
			right = strings.Trim(right, "\"")

			if left == right {
				return 1, nil
			}
			return 0, nil
		}
	}

	if strings.Contains(expression, "!=") {
		parts := strings.Split(expression, "!=")
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])

			// Remove quotes for string comparison
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
	if strings.Contains(expression, "&&") {
		parts := strings.Split(expression, "&&")
		if len(parts) == 2 {
			left, err1 := sep.evaluateMathExpression(strings.TrimSpace(parts[0]))
			right, err2 := sep.evaluateMathExpression(strings.TrimSpace(parts[1]))

			if err1 != nil || err2 != nil {
				return 0, fmt.Errorf("error evaluating boolean operation")
			}

			if left != 0 && right != 0 {
				return 1, nil
			}
			return 0, nil
		}
	}

	if strings.Contains(expression, "||") {
		parts := strings.Split(expression, "||")
		if len(parts) == 2 {
			left, err1 := sep.evaluateMathExpression(strings.TrimSpace(parts[0]))
			right, err2 := sep.evaluateMathExpression(strings.TrimSpace(parts[1]))

			if err1 != nil || err2 != nil {
				return 0, fmt.Errorf("error evaluating boolean operation")
			}

			if left != 0 || right != 0 {
				return 1, nil
			}
			return 0, nil
		}
	}

	return 0, fmt.Errorf("invalid boolean operation: %s", expression)
}

// evaluateArithmetic evaluates arithmetic expressions
func (sep *SimpleExpressionParser) evaluateArithmetic(expression string) (float64, error) {
	// This is a very simplified arithmetic evaluator
	// In production, you'd want a proper expression parser

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

	// Handle exponentiation
	for strings.Contains(expression, "^") {
		re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*\^\s*(\d+(?:\.\d+)?)`)
		expression = re.ReplaceAllStringFunc(expression, func(match string) string {
			parts := re.FindStringSubmatch(match)
			if len(parts) == 3 {
				base, _ := strconv.ParseFloat(parts[1], 64)
				exp, _ := strconv.ParseFloat(parts[2], 64)
				result := math.Pow(base, exp)
				return fmt.Sprintf("%.0f", result)
			}
			return match
		})
	}

	// Handle multiplication and division
	for strings.Contains(expression, "*") || strings.Contains(expression, "/") {
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

		// Division
		re = regexp.MustCompile(`(\d+(?:\.\d+)?)\s*/\s*(\d+(?:\.\d+)?)`)
		expression = re.ReplaceAllStringFunc(expression, func(match string) string {
			parts := re.FindStringSubmatch(match)
			if len(parts) == 3 {
				left, _ := strconv.ParseFloat(parts[1], 64)
				right, _ := strconv.ParseFloat(parts[2], 64)
				if right == 0 {
					return "0" // Division by zero
				}
				result := left / right
				return fmt.Sprintf("%.0f", result)
			}
			return match
		})
	}

	// Handle addition and subtraction
	for strings.Contains(expression, "+") || (strings.Contains(expression, "-") && !strings.HasPrefix(expression, "-")) {
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

		// Subtraction
		re = regexp.MustCompile(`(\d+(?:\.\d+)?)\s*-\s*(\d+(?:\.\d+)?)`)
		expression = re.ReplaceAllStringFunc(expression, func(match string) string {
			parts := re.FindStringSubmatch(match)
			if len(parts) == 3 {
				left, _ := strconv.ParseFloat(parts[1], 64)
				right, _ := strconv.ParseFloat(parts[2], 64)
				result := left - right
				return fmt.Sprintf("%.0f", result)
			}
			return match
		})
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
	if len(args) != 1 {
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
	if len(args) < 2 {
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
	if len(args) != 2 {
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
	if len(args) != 2 {
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
