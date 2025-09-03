package server

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/solo7.media/platform.drifter.solo7.media/internal/domain"
)

// RandomizationManagerImpl implements the RandomizationManager interface
type RandomizationManagerImpl struct {
	rules  map[string]*domain.RandomizationRule
	mu     sync.RWMutex
	logger domain.Logger
}

// NewRandomizationManager creates a new randomization manager
func NewRandomizationManager(logger domain.Logger) *RandomizationManagerImpl {
	return &RandomizationManagerImpl{
		rules:  make(map[string]*domain.RandomizationRule),
		logger: logger,
	}
}

// RegisterRandomizationRule registers a new randomization rule
func (rm *RandomizationManagerImpl) RegisterRandomizationRule(ctx context.Context, rule *domain.RandomizationRule) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rule.ID == "" {
		return fmt.Errorf("randomization rule ID cannot be empty")
	}

	if rule.Name == "" {
		return fmt.Errorf("randomization rule name cannot be empty")
	}

	// Validate the randomization rule
	if err := rm.validateRandomizationRule(rule); err != nil {
		return fmt.Errorf("randomization rule validation failed: %w", err)
	}

	// Check if rule already exists
	if _, exists := rm.rules[rule.ID]; exists {
		return fmt.Errorf("randomization rule with ID %s already exists", rule.ID)
	}

	rm.rules[rule.ID] = rule

	rm.logger.Info("Randomization rule registered", map[string]interface{}{
		"rule_id":   rule.ID,
		"rule_name": rule.Name,
		"rule_type": rule.Type,
	})

	return nil
}

// UnregisterRandomizationRule removes a randomization rule
func (rm *RandomizationManagerImpl) UnregisterRandomizationRule(ctx context.Context, ruleID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rule, exists := rm.rules[ruleID]
	if !exists {
		return fmt.Errorf("randomization rule with ID %s not found", ruleID)
	}

	delete(rm.rules, ruleID)

	rm.logger.Info("Randomization rule unregistered", map[string]interface{}{
		"rule_id":   ruleID,
		"rule_name": rule.Name,
	})

	return nil
}

// GetRandomizationRule retrieves a randomization rule by ID
func (rm *RandomizationManagerImpl) GetRandomizationRule(ctx context.Context, ruleID string) (*domain.RandomizationRule, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rule, exists := rm.rules[ruleID]
	if !exists {
		return nil, fmt.Errorf("randomization rule with ID %s not found", ruleID)
	}

	// Return a copy to prevent external modification
	ruleCopy := *rule
	return &ruleCopy, nil
}

// GetRandomizationRulesByType retrieves all randomization rules of a specific type
func (rm *RandomizationManagerImpl) GetRandomizationRulesByType(ctx context.Context, ruleType domain.RandomizationType) ([]*domain.RandomizationRule, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var rules []*domain.RandomizationRule
	for _, rule := range rm.rules {
		if rule.Type == ruleType {
			// Return a copy to prevent external modification
			ruleCopy := *rule
			rules = append(rules, &ruleCopy)
		}
	}

	return rules, nil
}

// ExecuteRandomization executes a randomization rule
func (rm *RandomizationManagerImpl) ExecuteRandomization(ctx context.Context, ruleID string, context map[string]interface{}) (*domain.RandomizationResult, error) {
	rm.mu.RLock()
	rule, exists := rm.rules[ruleID]
	rm.mu.RUnlock()

	if !exists {
		return &domain.RandomizationResult{
			RuleID:  ruleID,
			Success: false,
			Error:   fmt.Sprintf("randomization rule with ID %s not found", ruleID),
		}, nil
	}

	// Execute the randomization based on type
	result := rm.executeRandomizationByType(rule, context)
	return result, nil
}

// ValidateRandomizationRule validates a randomization rule
func (rm *RandomizationManagerImpl) ValidateRandomizationRule(ctx context.Context, rule *domain.RandomizationRule) error {
	return rm.validateRandomizationRule(rule)
}

// ListRandomizationRules returns all registered randomization rules
func (rm *RandomizationManagerImpl) ListRandomizationRules(ctx context.Context) ([]*domain.RandomizationRule, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var rules []*domain.RandomizationRule
	for _, rule := range rm.rules {
		// Return a copy to prevent external modification
		ruleCopy := *rule
		rules = append(rules, &ruleCopy)
	}

	return rules, nil
}

// validateRandomizationRule performs internal validation of a randomization rule
func (rm *RandomizationManagerImpl) validateRandomizationRule(rule *domain.RandomizationRule) error {
	if rule == nil {
		return fmt.Errorf("randomization rule cannot be nil")
	}

	if rule.ID == "" {
		return fmt.Errorf("randomization rule ID is required")
	}

	if rule.Name == "" {
		return fmt.Errorf("randomization rule name is required")
	}

	if rule.Type == "" {
		return fmt.Errorf("randomization rule type is required")
	}

	// Validate rule type
	validTypes := []domain.RandomizationType{
		domain.RandomizationTypeDice,
		domain.RandomizationTypeTable,
		domain.RandomizationTypeWeighted,
		domain.RandomizationTypeCustom,
	}

	validType := false
	for _, validTypeValue := range validTypes {
		if rule.Type == validTypeValue {
			validType = true
			break
		}
	}

	if !validType {
		return fmt.Errorf("invalid randomization rule type: %s", rule.Type)
	}

	// Validate notation for dice rules
	if rule.Type == domain.RandomizationTypeDice {
		if rule.Notation == "" {
			return fmt.Errorf("dice notation is required for dice randomization rules")
		}
		// TODO: Add dice notation validation
	}

	return nil
}

// executeRandomizationByType executes randomization based on rule type
func (rm *RandomizationManagerImpl) executeRandomizationByType(rule *domain.RandomizationRule, context map[string]interface{}) *domain.RandomizationResult {
	switch rule.Type {
	case domain.RandomizationTypeDice:
		return rm.executeDiceRoll(rule, context)
	case domain.RandomizationTypeTable:
		return rm.executeTableRoll(rule, context)
	case domain.RandomizationTypeWeighted:
		return rm.executeWeightedRoll(rule, context)
	case domain.RandomizationTypeCustom:
		return rm.executeCustomRoll(rule, context)
	default:
		return &domain.RandomizationResult{
			RuleID:  rule.ID,
			Success: false,
			Error:   fmt.Sprintf("unsupported randomization type: %s", rule.Type),
		}
	}
}

// executeDiceRoll executes a dice roll randomization
func (rm *RandomizationManagerImpl) executeDiceRoll(rule *domain.RandomizationRule, context map[string]interface{}) *domain.RandomizationResult {
	// TODO: Implement proper dice notation parsing
	// For now, implement basic dice rolling

	// Simple implementation for common dice notation like "1d6", "2d10+3"
	// This is a placeholder - proper implementation would use the expression parser

	// Generate a random number for demonstration
	rand.Seed(time.Now().UnixNano())
	value := rand.Intn(20) + 1 // 1d20 for demonstration

	return &domain.RandomizationResult{
		RuleID:  rule.ID,
		Value:   value,
		Details: fmt.Sprintf("Rolled %s: %d", rule.Notation, value),
		Success: true,
		Metadata: map[string]interface{}{
			"notation":  rule.Notation,
			"timestamp": time.Now(),
		},
	}
}

// executeTableRoll executes a table-based randomization
func (rm *RandomizationManagerImpl) executeTableRoll(rule *domain.RandomizationRule, context map[string]interface{}) *domain.RandomizationResult {
	// TODO: Implement table-based randomization
	return &domain.RandomizationResult{
		RuleID:  rule.ID,
		Success: false,
		Error:   "table-based randomization not yet implemented",
	}
}

// executeWeightedRoll executes a weighted randomization
func (rm *RandomizationManagerImpl) executeWeightedRoll(rule *domain.RandomizationRule, context map[string]interface{}) *domain.RandomizationResult {
	// TODO: Implement weighted randomization
	return &domain.RandomizationResult{
		RuleID:  rule.ID,
		Success: false,
		Error:   "weighted randomization not yet implemented",
	}
}

// executeCustomRoll executes a custom randomization
func (rm *RandomizationManagerImpl) executeCustomRoll(rule *domain.RandomizationRule, context map[string]interface{}) *domain.RandomizationResult {
	// TODO: Implement custom randomization using expression parser
	return &domain.RandomizationResult{
		RuleID:  rule.ID,
		Success: false,
		Error:   "custom randomization not yet implemented",
	}
}
