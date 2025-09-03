package domain

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ExpressionContext provides variables and functions available during expression evaluation
type ExpressionContext struct {
	Self      map[string]interface{} `json:"self"`       // Current entity (player, NPC, etc.)
	Target    map[string]interface{} `json:"target"`     // Target entity (for combat, etc.)
	Party     map[string]interface{} `json:"party"`      // Party/group context
	Terrain   map[string]interface{} `json:"terrain"`    // Environmental context
	Game      map[string]interface{} `json:"game"`       // Global game state
	EventData map[string]interface{} `json:"event_data"` // Top-level event data
}

// ExpressionResult represents the result of evaluating an expression
type ExpressionResult struct {
	Value    interface{}            `json:"value"`
	Type     string                 `json:"type"` // "number", "string", "boolean", "dice_roll"
	Success  bool                   `json:"success"`
	Error    string                 `json:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ExpressionParser handles parsing and evaluation of game expressions
type ExpressionParser struct {
	functions map[string]ExpressionFunction
	operators map[string]OperatorPrecedence
}

// ExpressionFunction defines a callable function in expressions
type ExpressionFunction struct {
	Name        string
	Description string
	Parameters  []ParameterDefinition
	Handler     func(args []interface{}, ctx *ExpressionContext) (*ExpressionResult, error)
}

// ParameterDefinition defines a function parameter
type ParameterDefinition struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"` // "number", "string", "boolean", "any"
	Required bool        `json:"required"`
	Default  interface{} `json:"default,omitempty"`
}

// OperatorPrecedence defines operator precedence for expression parsing
type OperatorPrecedence struct {
	Precedence    int
	Associativity string // "left" or "right"
}

// NewExpressionParser creates a new expression parser with built-in functions
func NewExpressionParser() *ExpressionParser {
	parser := &ExpressionParser{
		functions: make(map[string]ExpressionFunction),
		operators: make(map[string]OperatorPrecedence),
	}

	parser.initializeBuiltinFunctions()
	parser.initializeOperators()

	return parser
}

// initializeBuiltinFunctions sets up the core game functions
func (ep *ExpressionParser) initializeBuiltinFunctions() {
	// Dice rolling functions
	ep.functions["roll"] = ExpressionFunction{
		Name:        "roll",
		Description: "Roll dice with specified notation (e.g., '2d6', '1d20+3')",
		Parameters: []ParameterDefinition{
			{Name: "notation", Type: "string", Required: true},
		},
		Handler: ep.rollDice,
	}

	ep.functions["d"] = ExpressionFunction{
		Name:        "d",
		Description: "Roll a single die (e.g., d(6) for 1d6)",
		Parameters: []ParameterDefinition{
			{Name: "sides", Type: "number", Required: true},
		},
		Handler: ep.rollSingleDie,
	}

	// Combat functions
	ep.functions["deal"] = ExpressionFunction{
		Name:        "deal",
		Description: "Deal damage to a target",
		Parameters: []ParameterDefinition{
			{Name: "target", Type: "any", Required: true},
			{Name: "amount", Type: "number", Required: true},
			{Name: "type", Type: "string", Required: false, Default: "physical"},
		},
		Handler: ep.dealDamage,
	}

	ep.functions["heal"] = ExpressionFunction{
		Name:        "heal",
		Description: "Heal a target",
		Parameters: []ParameterDefinition{
			{Name: "target", Type: "any", Required: true},
			{Name: "amount", Type: "number", Required: true},
		},
		Handler: ep.healTarget,
	}

	// Condition functions
	ep.functions["has_tag"] = ExpressionFunction{
		Name:        "has_tag",
		Description: "Check if target has a specific tag",
		Parameters: []ParameterDefinition{
			{Name: "target", Type: "any", Required: true},
			{Name: "tag", Type: "string", Required: true},
		},
		Handler: ep.hasTag,
	}

	ep.functions["has_ability"] = ExpressionFunction{
		Name:        "has_ability",
		Description: "Check if target has a specific ability",
		Parameters: []ParameterDefinition{
			{Name: "target", Type: "any", Required: true},
			{Name: "ability", Type: "string", Required: true},
		},
		Handler: ep.hasAbility,
	}

	// Math functions
	ep.functions["min"] = ExpressionFunction{
		Name:        "min",
		Description: "Return the minimum of two values",
		Parameters: []ParameterDefinition{
			{Name: "a", Type: "number", Required: true},
			{Name: "b", Type: "number", Required: true},
		},
		Handler: ep.minFunction,
	}

	ep.functions["max"] = ExpressionFunction{
		Name:        "max",
		Description: "Return the maximum of two values",
		Parameters: []ParameterDefinition{
			{Name: "a", Type: "number", Required: true},
			{Name: "b", Type: "number", Required: true},
		},
		Handler: ep.maxFunction,
	}

	ep.functions["clamp"] = ExpressionFunction{
		Name:        "clamp",
		Description: "Clamp a value between min and max",
		Parameters: []ParameterDefinition{
			{Name: "value", Type: "number", Required: true},
			{Name: "min", Type: "number", Required: true},
			{Name: "max", Type: "number", Required: true},
		},
		Handler: ep.clampFunction,
	}
}

// initializeOperators sets up operator precedence
func (ep *ExpressionParser) initializeOperators() {
	ep.operators["||"] = OperatorPrecedence{Precedence: 1, Associativity: "left"}
	ep.operators["&&"] = OperatorPrecedence{Precedence: 2, Associativity: "left"}
	ep.operators["=="] = OperatorPrecedence{Precedence: 3, Associativity: "left"}
	ep.operators["!="] = OperatorPrecedence{Precedence: 3, Associativity: "left"}
	ep.operators["<"] = OperatorPrecedence{Precedence: 4, Associativity: "left"}
	ep.operators[">"] = OperatorPrecedence{Precedence: 4, Associativity: "left"}
	ep.operators["<="] = OperatorPrecedence{Precedence: 4, Associativity: "left"}
	ep.operators[">="] = OperatorPrecedence{Precedence: 4, Associativity: "left"}
	ep.operators["+"] = OperatorPrecedence{Precedence: 5, Associativity: "left"}
	ep.operators["-"] = OperatorPrecedence{Precedence: 5, Associativity: "left"}
	ep.operators["*"] = OperatorPrecedence{Precedence: 6, Associativity: "left"}
	ep.operators["/"] = OperatorPrecedence{Precedence: 6, Associativity: "left"}
	ep.operators["%"] = OperatorPrecedence{Precedence: 6, Associativity: "left"}
	ep.operators["^"] = OperatorPrecedence{Precedence: 7, Associativity: "right"}
}

// Evaluate parses and evaluates an expression string
func (ep *ExpressionParser) Evaluate(expression string, ctx *ExpressionContext) (*ExpressionResult, error) {
	if strings.TrimSpace(expression) == "" {
		return &ExpressionResult{
			Value:   nil,
			Type:    "null",
			Success: true,
		}, nil
	}

	// Parse the expression into tokens
	tokens, err := ep.tokenize(expression)
	if err != nil {
		return &ExpressionResult{
			Success: false,
			Error:   fmt.Sprintf("Tokenization error: %v", err),
		}, err
	}

	// Parse tokens into an AST
	ast, err := ep.parseExpression(tokens)
	if err != nil {
		return &ExpressionResult{
			Success: false,
			Error:   fmt.Sprintf("Parse error: %v", err),
		}, err
	}

	// Evaluate the AST
	result, err := ep.evaluateNode(ast, ctx)
	if err != nil {
		return &ExpressionResult{
			Success: false,
			Error:   fmt.Sprintf("Evaluation error: %v", err),
		}, err
	}

	return result, nil
}

// tokenize breaks an expression into tokens
func (ep *ExpressionParser) tokenize(expression string) ([]Token, error) {
	var tokens []Token
	expression = strings.TrimSpace(expression)

	for len(expression) > 0 {
		expression = strings.TrimSpace(expression)
		if len(expression) == 0 {
			break
		}

		// Check for function calls first
		if isAlpha(expression[0]) || expression[0] == '_' {
			// Look ahead to see if this is a function call
			ident := ""
			i := 0
			for i < len(expression) && (isAlpha(expression[i]) || isDigit(expression[i]) || expression[i] == '_') {
				ident += string(expression[i])
				i++
			}

			// Check if next non-whitespace character is '('
			j := i
			for j < len(expression) && expression[j] == ' ' {
				j++
			}

			if j < len(expression) && expression[j] == '(' {
				// This is a function call
				tokens = append(tokens, Token{Type: "function", Value: ident})
				expression = expression[j:] // Keep the '(' for processing
				continue
			} else {
				// This is a regular identifier
				tokens = append(tokens, Token{Type: "identifier", Value: ident})
				expression = expression[i:]
				continue
			}
		}

		// Check for operators (longest first)
		matched := false
		for op := range ep.operators {
			if strings.HasPrefix(expression, op) {
				tokens = append(tokens, Token{Type: "operator", Value: op})
				expression = expression[len(op):]
				matched = true
				break
			}
		}
		if matched {
			continue
		}

		// Check for parentheses
		if expression[0] == '(' {
			tokens = append(tokens, Token{Type: "lparen", Value: "("})
			expression = expression[1:]
			continue
		}
		if expression[0] == ')' {
			tokens = append(tokens, Token{Type: "rparen", Value: ")"})
			expression = expression[1:]
			continue
		}

		// Check for comma
		if expression[0] == ',' {
			tokens = append(tokens, Token{Type: "comma", Value: ","})
			expression = expression[1:]
			continue
		}

		// Check for string literals
		if expression[0] == '"' {
			end := strings.Index(expression[1:], "\"")
			if end == -1 {
				return nil, fmt.Errorf("unterminated string literal")
			}
			value := expression[1 : end+1]
			tokens = append(tokens, Token{Type: "string", Value: value})
			expression = expression[end+2:]
			continue
		}

		// Check for numbers
		if isDigit(expression[0]) || expression[0] == '.' {
			numStr := ""
			for len(expression) > 0 && (isDigit(expression[0]) || expression[0] == '.') {
				numStr += string(expression[0])
				expression = expression[1:]
			}
			tokens = append(tokens, Token{Type: "number", Value: numStr})
			continue
		}

		// Check for boolean literals
		if strings.HasPrefix(expression, "true") {
			tokens = append(tokens, Token{Type: "boolean", Value: "true"})
			expression = expression[4:]
			continue
		}
		if strings.HasPrefix(expression, "false") {
			tokens = append(tokens, Token{Type: "boolean", Value: "false"})
			expression = expression[5:]
			continue
		}

		return nil, fmt.Errorf("unexpected character: %c", expression[0])
	}

	return tokens, nil
}

// Token represents a parsed token
type Token struct {
	Type  string
	Value string
}

// ASTNode represents a node in the abstract syntax tree
type ASTNode struct {
	Type     string
	Value    interface{}
	Children []*ASTNode
}

// parseExpression parses tokens into an AST using operator precedence
func (ep *ExpressionParser) parseExpression(tokens []Token) (*ASTNode, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty expression")
	}

	// Convert to postfix notation using shunting yard algorithm
	postfix, err := ep.shuntingYard(tokens)
	if err != nil {
		return nil, err
	}

	// Build AST from postfix notation
	return ep.buildAST(postfix)
}

// shuntingYard converts infix notation to postfix using the shunting yard algorithm
func (ep *ExpressionParser) shuntingYard(tokens []Token) ([]Token, error) {
	var output []Token
	var operators []Token

	for _, token := range tokens {
		switch token.Type {
		case "number", "string", "identifier":
			output = append(output, token)
		case "lparen":
			operators = append(operators, token)
		case "rparen":
			for len(operators) > 0 && operators[len(operators)-1].Type != "lparen" {
				output = append(output, operators[len(operators)-1])
				operators = operators[:len(operators)-1]
			}
			if len(operators) == 0 {
				return nil, fmt.Errorf("mismatched parentheses")
			}
			operators = operators[:len(operators)-1] // Remove the '('
		case "operator":
			op1 := token
			for len(operators) > 0 && operators[len(operators)-1].Type == "operator" {
				op2 := operators[len(operators)-1]
				prec1 := ep.operators[op1.Value].Precedence
				prec2 := ep.operators[op2.Value].Precedence
				assoc1 := ep.operators[op1.Value].Associativity

				if (assoc1 == "left" && prec1 <= prec2) || (assoc1 == "right" && prec1 < prec2) {
					output = append(output, op2)
					operators = operators[:len(operators)-1]
				} else {
					break
				}
			}
			operators = append(operators, op1)
		}
	}

	// Pop remaining operators
	for len(operators) > 0 {
		if operators[len(operators)-1].Type == "lparen" {
			return nil, fmt.Errorf("mismatched parentheses")
		}
		output = append(output, operators[len(operators)-1])
		operators = operators[:len(operators)-1]
	}

	return output, nil
}

// buildAST builds an AST from postfix notation
func (ep *ExpressionParser) buildAST(postfix []Token) (*ASTNode, error) {
	var stack []*ASTNode

	for _, token := range postfix {
		if token.Type == "operator" {
			if len(stack) < 2 {
				return nil, fmt.Errorf("insufficient operands for operator %s", token.Value)
			}

			right := stack[len(stack)-1]
			left := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			node := &ASTNode{
				Type:     "binary_op",
				Value:    token.Value,
				Children: []*ASTNode{left, right},
			}
			stack = append(stack, node)
		} else {
			node := &ASTNode{
				Type:  token.Type,
				Value: token.Value,
			}
			stack = append(stack, node)
		}
	}

	if len(stack) != 1 {
		return nil, fmt.Errorf("invalid expression structure")
	}

	return stack[0], nil
}

// evaluateNode evaluates an AST node
func (ep *ExpressionParser) evaluateNode(node *ASTNode, ctx *ExpressionContext) (*ExpressionResult, error) {
	switch node.Type {
	case "number":
		val, err := strconv.ParseFloat(node.Value.(string), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", node.Value)
		}
		return &ExpressionResult{Value: val, Type: "number", Success: true}, nil

	case "string":
		return &ExpressionResult{Value: node.Value, Type: "string", Success: true}, nil

	case "boolean":
		value := node.Value.(string) == "true"
		return &ExpressionResult{Value: value, Type: "boolean", Success: true}, nil

	case "function":
		// This is a function call - we need to parse the arguments
		return ep.evaluateFunctionCall(node, ctx)

	case "identifier":
		// This is a variable reference
		return ep.evaluateVariable(node.Value.(string), ctx)

	case "binary_op":
		return ep.evaluateBinaryOperation(node, ctx)

	default:
		return nil, fmt.Errorf("unknown node type: %s", node.Type)
	}
}

// evaluateFunctionCall evaluates a function call
func (ep *ExpressionParser) evaluateFunctionCall(node *ASTNode, ctx *ExpressionContext) (*ExpressionResult, error) {
	funcName := node.Value.(string)

	// Parse function arguments
	var args []interface{}
	for _, child := range node.Children {
		result, err := ep.evaluateNode(child, ctx)
		if err != nil {
			return nil, err
		}
		args = append(args, result.Value)
	}

	// Look up function
	function, exists := ep.functions[funcName]
	if !exists {
		return nil, fmt.Errorf("unknown function: %s", funcName)
	}

	// Call function handler
	return function.Handler(args, ctx)
}

// evaluateVariable evaluates a variable reference
func (ep *ExpressionParser) evaluateVariable(varName string, ctx *ExpressionContext) (*ExpressionResult, error) {
	// Check context variables in order: self, target, party, terrain, game
	contexts := []map[string]interface{}{
		ctx.Self, ctx.Target, ctx.Party, ctx.Terrain, ctx.Game,
	}

	for _, context := range contexts {
		if value, exists := context[varName]; exists {
			return &ExpressionResult{
				Value:   value,
				Type:    ep.getType(value),
				Success: true,
			}, nil
		}
	}

	return nil, fmt.Errorf("undefined variable: %s", varName)
}

// evaluateBinaryOperation evaluates a binary operation
func (ep *ExpressionParser) evaluateBinaryOperation(node *ASTNode, ctx *ExpressionContext) (*ExpressionResult, error) {
	left, err := ep.evaluateNode(node.Children[0], ctx)
	if err != nil {
		return nil, err
	}

	right, err := ep.evaluateNode(node.Children[1], ctx)
	if err != nil {
		return nil, err
	}

	operator := node.Value.(string)

	// Convert to numbers if both are numeric
	leftNum, leftIsNum := left.Value.(float64)
	rightNum, rightIsNum := right.Value.(float64)

	if leftIsNum && rightIsNum {
		return ep.evaluateNumericOperation(leftNum, rightNum, operator)
	}

	// String operations
	if left.Type == "string" && right.Type == "string" {
		return ep.evaluateStringOperation(left.Value.(string), right.Value.(string), operator)
	}

	// Boolean operations
	return ep.evaluateBooleanOperation(left, right, operator)
}

// evaluateNumericOperation evaluates numeric operations
func (ep *ExpressionParser) evaluateNumericOperation(left, right float64, operator string) (*ExpressionResult, error) {
	var result float64

	switch operator {
	case "+":
		result = left + right
	case "-":
		result = left - right
	case "*":
		result = left * right
	case "/":
		if right == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result = left / right
	case "%":
		result = math.Mod(left, right)
	case "^":
		result = math.Pow(left, right)
	case "==":
		return &ExpressionResult{Value: left == right, Type: "boolean", Success: true}, nil
	case "!=":
		return &ExpressionResult{Value: left != right, Type: "boolean", Success: true}, nil
	case "<":
		return &ExpressionResult{Value: left < right, Type: "boolean", Success: true}, nil
	case ">":
		return &ExpressionResult{Value: left > right, Type: "boolean", Success: true}, nil
	case "<=":
		return &ExpressionResult{Value: left <= right, Type: "boolean", Success: true}, nil
	case ">=":
		return &ExpressionResult{Value: left >= right, Type: "boolean", Success: true}, nil
	default:
		return nil, fmt.Errorf("unsupported numeric operator: %s", operator)
	}

	return &ExpressionResult{Value: result, Type: "number", Success: true}, nil
}

// evaluateStringOperation evaluates string operations
func (ep *ExpressionParser) evaluateStringOperation(left, right string, operator string) (*ExpressionResult, error) {
	switch operator {
	case "+":
		return &ExpressionResult{Value: left + right, Type: "string", Success: true}, nil
	case "==":
		return &ExpressionResult{Value: left == right, Type: "boolean", Success: true}, nil
	case "!=":
		return &ExpressionResult{Value: left != right, Type: "boolean", Success: true}, nil
	default:
		return nil, fmt.Errorf("unsupported string operator: %s", operator)
	}
}

// evaluateBooleanOperation evaluates boolean operations
func (ep *ExpressionParser) evaluateBooleanOperation(left, right *ExpressionResult, operator string) (*ExpressionResult, error) {
	leftBool := ep.toBool(left.Value)
	rightBool := ep.toBool(right.Value)

	switch operator {
	case "&&":
		return &ExpressionResult{Value: leftBool && rightBool, Type: "boolean", Success: true}, nil
	case "||":
		return &ExpressionResult{Value: leftBool || rightBool, Type: "boolean", Success: true}, nil
	case "==":
		return &ExpressionResult{Value: left.Value == right.Value, Type: "boolean", Success: true}, nil
	case "!=":
		return &ExpressionResult{Value: left.Value != right.Value, Type: "boolean", Success: true}, nil
	default:
		return nil, fmt.Errorf("unsupported boolean operator: %s", operator)
	}
}

// Built-in function implementations

func (ep *ExpressionParser) rollDice(args []interface{}, ctx *ExpressionContext) (*ExpressionResult, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("roll() expects exactly 1 argument")
	}

	notation, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("roll() argument must be a string")
	}

	result, err := ep.parseDiceNotation(notation)
	if err != nil {
		return nil, err
	}

	return &ExpressionResult{
		Value:   result.Total,
		Type:    "dice_roll",
		Success: true,
		Metadata: map[string]interface{}{
			"notation": notation,
			"rolls":    result.Rolls,
			"modifier": result.Modifier,
		},
	}, nil
}

func (ep *ExpressionParser) rollSingleDie(args []interface{}, ctx *ExpressionContext) (*ExpressionResult, error) {
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

	roll := ep.randomInt(1, int(sides))

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

func (ep *ExpressionParser) dealDamage(args []interface{}, ctx *ExpressionContext) (*ExpressionResult, error) {
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

	// In a real implementation, this would actually deal damage to the target
	// For now, we'll just return the damage amount
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

func (ep *ExpressionParser) healTarget(args []interface{}, ctx *ExpressionContext) (*ExpressionResult, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("heal() expects exactly 2 arguments")
	}

	amount, ok := args[1].(float64)
	if !ok {
		return nil, fmt.Errorf("heal() amount must be a number")
	}

	// In a real implementation, this would actually heal the target
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

func (ep *ExpressionParser) hasTag(args []interface{}, ctx *ExpressionContext) (*ExpressionResult, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("has_tag() expects exactly 2 arguments")
	}

	tag, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("has_tag() tag must be a string")
	}

	// In a real implementation, this would check the target's tags
	// For now, we'll return false
	return &ExpressionResult{
		Value:   false,
		Type:    "boolean",
		Success: true,
		Metadata: map[string]interface{}{
			"action": "check_tag",
			"target": args[0],
			"tag":    tag,
		},
	}, nil
}

func (ep *ExpressionParser) hasAbility(args []interface{}, ctx *ExpressionContext) (*ExpressionResult, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("has_ability() expects exactly 2 arguments")
	}

	ability, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("has_ability() ability must be a string")
	}

	// In a real implementation, this would check the target's abilities
	// For now, we'll return false
	return &ExpressionResult{
		Value:   false,
		Type:    "boolean",
		Success: true,
		Metadata: map[string]interface{}{
			"action":  "check_ability",
			"target":  args[0],
			"ability": ability,
		},
	}, nil
}

func (ep *ExpressionParser) minFunction(args []interface{}, ctx *ExpressionContext) (*ExpressionResult, error) {
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

func (ep *ExpressionParser) maxFunction(args []interface{}, ctx *ExpressionContext) (*ExpressionResult, error) {
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

func (ep *ExpressionParser) clampFunction(args []interface{}, ctx *ExpressionContext) (*ExpressionResult, error) {
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

func (ep *ExpressionParser) parseDiceNotation(notation string) (*DiceResult, error) {
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
		roll := ep.randomInt(1, sides)
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

type DiceResult struct {
	Notation string `json:"notation"`
	Count    int    `json:"count"`
	Sides    int    `json:"sides"`
	Modifier int    `json:"modifier"`
	Rolls    []int  `json:"rolls"`
	Total    int    `json:"total"`
}

func (ep *ExpressionParser) randomInt(min, max int) int {
	// Simple pseudo-random number generator
	// In a real implementation, you'd use a proper RNG with seeding
	return min + (int(time.Now().UnixNano()) % (max - min + 1))
}

func (ep *ExpressionParser) getType(value interface{}) string {
	switch value.(type) {
	case float64:
		return "number"
	case string:
		return "string"
	case bool:
		return "boolean"
	default:
		return "unknown"
	}
}

func (ep *ExpressionParser) toBool(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case float64:
		return v != 0
	case string:
		return v != ""
	default:
		return value != nil
	}
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
