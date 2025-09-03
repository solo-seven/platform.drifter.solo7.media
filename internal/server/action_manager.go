package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/solo7.media/platform.drifter.solo7.media/internal/domain"
)

// ActionManagerImpl implements the ActionDefinitionManager interface
type ActionManagerImpl struct {
	actions map[domain.ActionTypeId]*domain.ActionDefinition
	mu      sync.RWMutex
	logger  domain.Logger
}

// NewActionManager creates a new action manager
func NewActionManager(logger domain.Logger) *ActionManagerImpl {
	return &ActionManagerImpl{
		actions: make(map[domain.ActionTypeId]*domain.ActionDefinition),
		logger:  logger,
	}
}

// RegisterAction registers a new action definition
func (am *ActionManagerImpl) RegisterAction(ctx context.Context, action *domain.ActionDefinition) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if action.ID == "" {
		return fmt.Errorf("action ID cannot be empty")
	}

	if action.Name == "" {
		return fmt.Errorf("action name cannot be empty")
	}

	// Validate the action
	if err := am.validateAction(action); err != nil {
		return fmt.Errorf("action validation failed: %w", err)
	}

	// Check if action already exists
	if _, exists := am.actions[action.ID]; exists {
		return fmt.Errorf("action with ID %s already exists", action.ID)
	}

	am.actions[action.ID] = action

	am.logger.Info("Action registered", map[string]interface{}{
		"action_id":   action.ID,
		"action_name": action.Name,
		"action_type": action.Type,
	})

	return nil
}

// UnregisterAction removes an action definition
func (am *ActionManagerImpl) UnregisterAction(ctx context.Context, actionID domain.ActionTypeId) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	action, exists := am.actions[actionID]
	if !exists {
		return fmt.Errorf("action with ID %s not found", actionID)
	}

	delete(am.actions, actionID)

	am.logger.Info("Action unregistered", map[string]interface{}{
		"action_id":   actionID,
		"action_name": action.Name,
	})

	return nil
}

// GetAction retrieves an action definition by ID
func (am *ActionManagerImpl) GetAction(ctx context.Context, actionID domain.ActionTypeId) (*domain.ActionDefinition, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	action, exists := am.actions[actionID]
	if !exists {
		return nil, fmt.Errorf("action with ID %s not found", actionID)
	}

	// Return a copy to prevent external modification
	actionCopy := *action
	return &actionCopy, nil
}

// GetActionsByType retrieves all actions of a specific type
func (am *ActionManagerImpl) GetActionsByType(ctx context.Context, actionType domain.ActionType) ([]*domain.ActionDefinition, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var actions []*domain.ActionDefinition
	for _, action := range am.actions {
		if action.Type == actionType {
			// Return a copy to prevent external modification
			actionCopy := *action
			actions = append(actions, &actionCopy)
		}
	}

	return actions, nil
}

// ValidateAction validates an action definition
func (am *ActionManagerImpl) ValidateAction(ctx context.Context, action *domain.ActionDefinition) error {
	return am.validateAction(action)
}

// ListActions returns all registered actions
func (am *ActionManagerImpl) ListActions(ctx context.Context) ([]*domain.ActionDefinition, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var actions []*domain.ActionDefinition
	for _, action := range am.actions {
		// Return a copy to prevent external modification
		actionCopy := *action
		actions = append(actions, &actionCopy)
	}

	return actions, nil
}

// validateAction performs internal validation of an action
func (am *ActionManagerImpl) validateAction(action *domain.ActionDefinition) error {
	if action == nil {
		return fmt.Errorf("action cannot be nil")
	}

	if action.ID == "" {
		return fmt.Errorf("action ID is required")
	}

	if action.Name == "" {
		return fmt.Errorf("action name is required")
	}

	if action.Type == "" {
		return fmt.Errorf("action type is required")
	}

	// Validate action type
	validTypes := []domain.ActionType{
		domain.ActionTypeAttack,
		domain.ActionTypeMove,
		domain.ActionTypeCast,
		domain.ActionTypeInteract,
		domain.ActionTypeSkill,
		domain.ActionTypeSocial,
		domain.ActionTypeDefend,
		domain.ActionTypeUse,
	}

	validType := false
	for _, validTypeValue := range validTypes {
		if action.Type == validTypeValue {
			validType = true
			break
		}
	}

	if !validType {
		return fmt.Errorf("invalid action type: %s", action.Type)
	}

	// Validate cost
	if action.Cost.ActionPoints < 0 {
		return fmt.Errorf("action points cost cannot be negative")
	}

	if action.Cost.MovementPoints < 0 {
		return fmt.Errorf("movement points cost cannot be negative")
	}

	// Validate resolution method
	if action.Resolution.ID == "" {
		return fmt.Errorf("resolution method ID is required")
	}

	// Validate prerequisites
	for i, prereq := range action.Prerequisites {
		if prereq.Type == "" {
			return fmt.Errorf("prerequisite %d type cannot be empty", i)
		}
		if prereq.Condition == "" {
			return fmt.Errorf("prerequisite %d condition cannot be empty", i)
		}
	}

	// Validate context checks
	for i, check := range action.Validation.ContextChecks {
		if check.Type == "" {
			return fmt.Errorf("context check %d type cannot be empty", i)
		}
		if check.Expression == "" {
			return fmt.Errorf("context check %d expression cannot be empty", i)
		}
	}

	return nil
}
