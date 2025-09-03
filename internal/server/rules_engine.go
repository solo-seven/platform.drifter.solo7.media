package server

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
)

// RulesEngineImpl implements the RulesEngine interface
type RulesEngineImpl struct {
	rules            map[domain.RuleId]*domain.GameRule
	expressionParser *domain.SimpleExpressionParser
	mu               sync.RWMutex
	logger           domain.Logger
}

// NewRulesEngine creates a new rules engine
func NewRulesEngine(logger domain.Logger) *RulesEngineImpl {
	return &RulesEngineImpl{
		rules:            make(map[domain.RuleId]*domain.GameRule),
		expressionParser: domain.NewSimpleExpressionParser(),
		logger:           logger,
	}
}

// RegisterRule registers a new game rule
func (re *RulesEngineImpl) RegisterRule(ctx context.Context, rule *domain.GameRule) error {
	re.mu.Lock()
	defer re.mu.Unlock()

	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
	}

	re.rules[rule.ID] = rule

	re.logger.Debug("Rule registered", map[string]interface{}{
		"rule_id":  rule.ID,
		"priority": rule.Priority,
	})

	return nil
}

// UnregisterRule removes a game rule
func (re *RulesEngineImpl) UnregisterRule(ctx context.Context, ruleID domain.RuleId) error {
	re.mu.Lock()
	defer re.mu.Unlock()

	_, exists := re.rules[ruleID]
	if !exists {
		return fmt.Errorf("rule %s not found", ruleID)
	}

	delete(re.rules, ruleID)

	re.logger.Debug("Rule unregistered", map[string]interface{}{
		"rule_id": ruleID,
	})

	return nil
}

// ProcessEvent processes a game event and returns the result
func (re *RulesEngineImpl) ProcessEvent(ctx context.Context, event domain.EventType, data map[string]interface{}) (*domain.ActionResult, error) {
	re.mu.RLock()
	defer re.mu.RUnlock()

	result := &domain.ActionResult{
		WorldStateChanges:   []domain.StateChange{},
		ClientNotifications: []domain.ClientNotification{},
		AestheticEvents:     []domain.AestheticEventData{},
	}

	// Find applicable rules
	var applicableRules []*domain.GameRule
	for _, rule := range re.rules {
		if re.isRuleApplicable(rule, event, data) {
			applicableRules = append(applicableRules, rule)
		}
	}

	// Sort rules by priority (higher priority first)
	sort.Slice(applicableRules, func(i, j int) bool {
		return applicableRules[i].Priority > applicableRules[j].Priority
	})

	// Execute applicable rules
	for _, rule := range applicableRules {
		ruleResult, err := re.executeRule(ctx, rule, event, data)
		if err != nil {
			re.logger.Error("Failed to execute rule", map[string]interface{}{
				"rule_id": rule.ID,
				"error":   err,
			})
			continue
		}

		// Merge results
		result.WorldStateChanges = append(result.WorldStateChanges, ruleResult.WorldStateChanges...)
		result.ClientNotifications = append(result.ClientNotifications, ruleResult.ClientNotifications...)
		result.AestheticEvents = append(result.AestheticEvents, ruleResult.AestheticEvents...)
	}

	re.logger.Debug("Event processed", map[string]interface{}{
		"event_type":       event,
		"applicable_rules": len(applicableRules),
		"state_changes":    len(result.WorldStateChanges),
		"notifications":    len(result.ClientNotifications),
		"aesthetic_events": len(result.AestheticEvents),
	})

	return result, nil
}

// GetActiveRules returns all active rules for a region
func (re *RulesEngineImpl) GetActiveRules(ctx context.Context, regionID domain.RegionId) ([]*domain.GameRule, error) {
	re.mu.RLock()
	defer re.mu.RUnlock()

	var activeRules []*domain.GameRule
	for _, rule := range re.rules {
		// TODO: Implement region-based rule filtering
		// For now, return all rules
		activeRules = append(activeRules, rule)
	}

	return activeRules, nil
}

// isRuleApplicable checks if a rule is applicable to an event
func (re *RulesEngineImpl) isRuleApplicable(rule *domain.GameRule, event domain.EventType, data map[string]interface{}) bool {
	// Check if any trigger matches the event
	for _, trigger := range rule.Triggers {
		if trigger.Type == event {
			// Check additional conditions if any
			if re.evaluateConditions(rule.Conditions, data) {
				return true
			}
		}
	}

	return false
}

// evaluateConditions evaluates rule conditions against event data
func (re *RulesEngineImpl) evaluateConditions(conditions []domain.Condition, data map[string]interface{}) bool {
	// If no conditions, rule is applicable
	if len(conditions) == 0 {
		return true
	}

	// Create expression context from event data
	ctx := &domain.ExpressionContext{
		Self:    extractEntityData(data, "self"),
		Target:  extractEntityData(data, "target"),
		Party:   extractEntityData(data, "party"),
		Terrain: extractEntityData(data, "terrain"),
		Game:    extractEntityData(data, "game"),
		// Add top-level event data to context
		EventData: data,
	}

	// All conditions must be true (AND logic)
	for _, condition := range conditions {
		if !re.evaluateCondition(condition, ctx) {
			return false
		}
	}

	return true
}

// evaluateCondition evaluates a single condition using the expression parser
func (re *RulesEngineImpl) evaluateCondition(condition domain.Condition, ctx *domain.ExpressionContext) bool {
	// Build expression from condition
	var expression string

	// If condition has an expression, use it directly
	if condition.Type == "expression" {
		expression = condition.Property
	} else {
		// Build simple comparison expression
		// Handle string values by adding quotes
		var valueStr string
		switch v := condition.Value.(type) {
		case string:
			valueStr = fmt.Sprintf("\"%s\"", v)
		default:
			valueStr = fmt.Sprintf("%v", v)
		}
		expression = fmt.Sprintf("%s %s %s", condition.Property, condition.Operator, valueStr)
	}

	// Evaluate the expression
	result, err := re.expressionParser.Evaluate(expression, ctx)
	if err != nil {
		re.logger.Warn("Failed to evaluate condition", map[string]interface{}{
			"condition":  condition,
			"expression": expression,
			"error":      err,
		})
		return false
	}

	if !result.Success {
		re.logger.Warn("Condition evaluation failed", map[string]interface{}{
			"condition":  condition,
			"expression": expression,
			"error":      result.Error,
		})
		return false
	}

	// Convert result to boolean
	return re.toBool(result.Value)
}

// executeRule executes a rule and returns the result
func (re *RulesEngineImpl) executeRule(ctx context.Context, rule *domain.GameRule, event domain.EventType, data map[string]interface{}) (*domain.ActionResult, error) {
	result := &domain.ActionResult{
		WorldStateChanges:   []domain.StateChange{},
		ClientNotifications: []domain.ClientNotification{},
		AestheticEvents:     []domain.AestheticEventData{},
	}

	// Execute each action in the rule
	for _, action := range rule.Actions {
		actionResult, err := re.executeAction(ctx, action, data)
		if err != nil {
			re.logger.Error("Failed to execute action", map[string]interface{}{
				"rule_id":     rule.ID,
				"action_type": action.Type,
				"error":       err,
			})
			continue
		}

		// Merge action result
		result.WorldStateChanges = append(result.WorldStateChanges, actionResult.WorldStateChanges...)
		result.ClientNotifications = append(result.ClientNotifications, actionResult.ClientNotifications...)
		result.AestheticEvents = append(result.AestheticEvents, actionResult.AestheticEvents...)
	}

	return result, nil
}

// executeAction executes a single action
func (re *RulesEngineImpl) executeAction(ctx context.Context, action domain.Action, data map[string]interface{}) (*domain.ActionResult, error) {
	result := &domain.ActionResult{
		WorldStateChanges:   []domain.StateChange{},
		ClientNotifications: []domain.ClientNotification{},
		AestheticEvents:     []domain.AestheticEventData{},
	}

	// Create expression context for action execution
	exprCtx := &domain.ExpressionContext{
		Self:    extractEntityData(data, "self"),
		Target:  extractEntityData(data, "target"),
		Party:   extractEntityData(data, "party"),
		Terrain: extractEntityData(data, "terrain"),
		Game:    extractEntityData(data, "game"),
	}

	switch action.Type {
	case "expression":
		// Execute an expression directly
		expression, ok := action.Properties["expression"].(string)
		if !ok {
			return nil, fmt.Errorf("expression action requires 'expression' property")
		}

		exprResult, err := re.expressionParser.Evaluate(expression, exprCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate expression: %v", err)
		}

		if !exprResult.Success {
			return nil, fmt.Errorf("expression evaluation failed: %s", exprResult.Error)
		}

		// Store result in action metadata
		if action.Properties == nil {
			action.Properties = make(map[string]interface{})
		}
		action.Properties["result"] = exprResult.Value
		action.Properties["result_type"] = exprResult.Type

	case "state_change":
		// Create a state change with evaluated properties
		evaluatedProps := re.evaluateActionProperties(action.Properties, exprCtx)

		entityID, err := re.extractEntityID(data, "entity_id")
		if err != nil {
			return nil, fmt.Errorf("invalid entity_id: %v", err)
		}

		change := domain.StateChange{
			EntityID:  entityID,
			Component: action.Target,
			Changes:   evaluatedProps,
			Timestamp: time.Now(),
		}
		result.WorldStateChanges = append(result.WorldStateChanges, change)

	case "notification":
		// Create a client notification with evaluated properties
		evaluatedProps := re.evaluateActionProperties(action.Properties, exprCtx)

		playerID, err := re.extractEntityID(data, "player_id")
		if err != nil {
			return nil, fmt.Errorf("invalid player_id: %v", err)
		}

		message, ok := evaluatedProps["message"].(string)
		if !ok {
			message = "System notification"
		}

		notification := domain.ClientNotification{
			PlayerID:   playerID,
			Type:       action.Target,
			Message:    message,
			Properties: evaluatedProps,
		}
		result.ClientNotifications = append(result.ClientNotifications, notification)

	case "aesthetic_event":
		// Create an aesthetic event with evaluated properties
		evaluatedProps := re.evaluateActionProperties(action.Properties, exprCtx)

		entityID, err := re.extractEntityID(data, "entity_id")
		if err != nil {
			return nil, fmt.Errorf("invalid entity_id: %v", err)
		}

		aestheticEvent := domain.AestheticEventData{
			Type:       action.Target,
			EntityID:   entityID,
			Properties: evaluatedProps,
			Timestamp:  time.Now(),
		}
		result.AestheticEvents = append(result.AestheticEvents, aestheticEvent)

	default:
		re.logger.Warn("Unknown action type", map[string]interface{}{
			"action_type": action.Type,
		})
	}

	return result, nil
}

// Helper functions

// extractEntityData extracts entity data from event data
func extractEntityData(data map[string]interface{}, key string) map[string]interface{} {
	if entityData, exists := data[key]; exists {
		if entityMap, ok := entityData.(map[string]interface{}); ok {
			return entityMap
		}
	}
	return make(map[string]interface{})
}

// extractEntityID extracts and parses an entity ID from data
func (re *RulesEngineImpl) extractEntityID(data map[string]interface{}, key string) (domain.EntityId, error) {
	entityIDStr, exists := data[key]
	if !exists {
		return domain.EntityId{}, fmt.Errorf("missing %s in event data", key)
	}

	switch v := entityIDStr.(type) {
	case string:
		return uuid.Parse(v)
	case domain.EntityId:
		return v, nil
	default:
		return domain.EntityId{}, fmt.Errorf("invalid %s type: %T", key, entityIDStr)
	}
}

// evaluateActionProperties evaluates expressions in action properties
func (re *RulesEngineImpl) evaluateActionProperties(properties map[string]interface{}, ctx *domain.ExpressionContext) map[string]interface{} {
	if properties == nil {
		return make(map[string]interface{})
	}

	evaluated := make(map[string]interface{})

	for key, value := range properties {
		if strValue, ok := value.(string); ok && re.isExpression(strValue) {
			// Evaluate as expression
			result, err := re.expressionParser.Evaluate(strValue, ctx)
			if err != nil {
				re.logger.Warn("Failed to evaluate property expression", map[string]interface{}{
					"key":   key,
					"value": strValue,
					"error": err,
				})
				evaluated[key] = value // Use original value on error
			} else if result.Success {
				evaluated[key] = result.Value
			} else {
				re.logger.Warn("Property expression evaluation failed", map[string]interface{}{
					"key":   key,
					"value": strValue,
					"error": result.Error,
				})
				evaluated[key] = value // Use original value on error
			}
		} else {
			// Use value as-is
			evaluated[key] = value
		}
	}

	return evaluated
}

// isExpression checks if a string looks like an expression
func (re *RulesEngineImpl) isExpression(str string) bool {
	// Simple heuristic: if it contains operators, variables, or function calls, it's likely an expression
	operators := []string{"+", "-", "*", "/", "==", "!=", "<", ">", "&&", "||", "roll(", "deal(", "heal(", "has_tag(", "has_ability("}

	for _, op := range operators {
		if strings.Contains(str, op) {
			return true
		}
	}

	return false
}

// toBool converts a value to boolean
func (re *RulesEngineImpl) toBool(value interface{}) bool {
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
