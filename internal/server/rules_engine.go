package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/solo7.media/platform.drifter.solo7.media/internal/domain"
)

// RulesEngineImpl implements the RulesEngine interface
type RulesEngineImpl struct {
	rules  map[domain.RuleId]*domain.GameRule
	mu     sync.RWMutex
	logger domain.Logger
}

// NewRulesEngine creates a new rules engine
func NewRulesEngine(logger domain.Logger) *RulesEngineImpl {
	return &RulesEngineImpl{
		rules:  make(map[domain.RuleId]*domain.GameRule),
		logger: logger,
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
	// TODO: Implement proper sorting

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

	// All conditions must be true (AND logic)
	for _, condition := range conditions {
		if !re.evaluateCondition(condition, data) {
			return false
		}
	}

	return true
}

// evaluateCondition evaluates a single condition
func (re *RulesEngineImpl) evaluateCondition(condition domain.Condition, data map[string]interface{}) bool {
	// Get the value from data
	value, exists := data[condition.Property]
	if !exists {
		return false
	}

	// Simple string comparison for now
	// TODO: Implement more sophisticated condition evaluation
	switch condition.Operator {
	case "==":
		return fmt.Sprintf("%v", value) == condition.Value
	case "!=":
		return fmt.Sprintf("%v", value) != condition.Value
	case ">":
		// TODO: Implement numeric comparison
		return false
	case "<":
		// TODO: Implement numeric comparison
		return false
	default:
		return false
	}
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

	switch action.Type {
	case "state_change":
		// Create a state change
		change := domain.StateChange{
			EntityID:  uuid.MustParse(data["entity_id"].(string)),
			Component: action.Target,
			Changes:   action.Properties,
			Timestamp: time.Now(),
		}
		result.WorldStateChanges = append(result.WorldStateChanges, change)

	case "notification":
		// Create a client notification
		notification := domain.ClientNotification{
			PlayerID:   uuid.MustParse(data["player_id"].(string)),
			Type:       action.Target,
			Message:    action.Properties["message"].(string),
			Properties: action.Properties,
		}
		result.ClientNotifications = append(result.ClientNotifications, notification)

	case "aesthetic_event":
		// Create an aesthetic event
		aestheticEvent := domain.AestheticEventData{
			Type:       action.Target,
			EntityID:   uuid.MustParse(data["entity_id"].(string)),
			Properties: action.Properties,
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
