package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
)

// ModifierManagerImpl implements the ModifierManager interface
type ModifierManagerImpl struct {
	modifiers map[string]*domain.ModifierSystem
	mu        sync.RWMutex
	logger    domain.Logger
}

// NewModifierManager creates a new modifier manager
func NewModifierManager(logger domain.Logger) *ModifierManagerImpl {
	return &ModifierManagerImpl{
		modifiers: make(map[string]*domain.ModifierSystem),
		logger:    logger,
	}
}

// RegisterModifier registers a new modifier system
func (mm *ModifierManagerImpl) RegisterModifier(ctx context.Context, modifier *domain.ModifierSystem) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if modifier.ID == "" {
		return fmt.Errorf("modifier ID cannot be empty")
	}

	if modifier.Name == "" {
		return fmt.Errorf("modifier name cannot be empty")
	}

	// Validate the modifier
	if err := mm.validateModifier(modifier); err != nil {
		return fmt.Errorf("modifier validation failed: %w", err)
	}

	// Check if modifier already exists
	if _, exists := mm.modifiers[modifier.ID]; exists {
		return fmt.Errorf("modifier with ID %s already exists", modifier.ID)
	}

	mm.modifiers[modifier.ID] = modifier

	mm.logger.Info("Modifier registered", map[string]interface{}{
		"modifier_id":   modifier.ID,
		"modifier_name": modifier.Name,
		"modifier_type": modifier.Type,
	})

	return nil
}

// UnregisterModifier removes a modifier system
func (mm *ModifierManagerImpl) UnregisterModifier(ctx context.Context, modifierID string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	modifier, exists := mm.modifiers[modifierID]
	if !exists {
		return fmt.Errorf("modifier with ID %s not found", modifierID)
	}

	delete(mm.modifiers, modifierID)

	mm.logger.Info("Modifier unregistered", map[string]interface{}{
		"modifier_id":   modifierID,
		"modifier_name": modifier.Name,
	})

	return nil
}

// GetModifier retrieves a modifier system by ID
func (mm *ModifierManagerImpl) GetModifier(ctx context.Context, modifierID string) (*domain.ModifierSystem, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	modifier, exists := mm.modifiers[modifierID]
	if !exists {
		return nil, fmt.Errorf("modifier with ID %s not found", modifierID)
	}

	// Return a copy to prevent external modification
	modifierCopy := *modifier
	return &modifierCopy, nil
}

// GetModifiersByType retrieves all modifiers of a specific type
func (mm *ModifierManagerImpl) GetModifiersByType(ctx context.Context, modifierType domain.ModifierType) ([]*domain.ModifierSystem, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	var modifiers []*domain.ModifierSystem
	for _, modifier := range mm.modifiers {
		if modifier.Type == modifierType {
			// Return a copy to prevent external modification
			modifierCopy := *modifier
			modifiers = append(modifiers, &modifierCopy)
		}
	}

	return modifiers, nil
}

// ApplyModifiers applies all applicable modifiers to an entity
func (mm *ModifierManagerImpl) ApplyModifiers(ctx context.Context, entityID domain.EntityId, target string, context map[string]interface{}) ([]domain.ModifierResult, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	var results []domain.ModifierResult

	for _, modifier := range mm.modifiers {
		// Check if modifier applies to the target
		if modifier.Application.Target != target {
			continue
		}

		// Check if conditions are met
		applies, err := mm.checkModifierConditions(modifier, context)
		if err != nil {
			results = append(results, domain.ModifierResult{
				ModifierID: modifier.ID,
				Success:    false,
				Error:      fmt.Sprintf("condition check failed: %v", err),
			})
			continue
		}

		if !applies {
			continue
		}

		// Apply the modifier
		result := mm.applyModifier(modifier, context)
		results = append(results, result)
	}

	return results, nil
}

// ValidateModifier validates a modifier system
func (mm *ModifierManagerImpl) ValidateModifier(ctx context.Context, modifier *domain.ModifierSystem) error {
	return mm.validateModifier(modifier)
}

// ListModifiers returns all registered modifiers
func (mm *ModifierManagerImpl) ListModifiers(ctx context.Context) ([]*domain.ModifierSystem, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	var modifiers []*domain.ModifierSystem
	for _, modifier := range mm.modifiers {
		// Return a copy to prevent external modification
		modifierCopy := *modifier
		modifiers = append(modifiers, &modifierCopy)
	}

	return modifiers, nil
}

// validateModifier performs internal validation of a modifier
func (mm *ModifierManagerImpl) validateModifier(modifier *domain.ModifierSystem) error {
	if modifier == nil {
		return fmt.Errorf("modifier cannot be nil")
	}

	if modifier.ID == "" {
		return fmt.Errorf("modifier ID is required")
	}

	if modifier.Name == "" {
		return fmt.Errorf("modifier name is required")
	}

	if modifier.Type == "" {
		return fmt.Errorf("modifier type is required")
	}

	// Validate modifier type
	validTypes := []domain.ModifierType{
		domain.ModifierTypeAttribute,
		domain.ModifierTypeEquipment,
		domain.ModifierTypeEnvironmental,
		domain.ModifierTypeTemporary,
		domain.ModifierTypeSkill,
		domain.ModifierTypeMagic,
	}

	validType := false
	for _, validTypeValue := range validTypes {
		if modifier.Type == validTypeValue {
			validType = true
			break
		}
	}

	if !validType {
		return fmt.Errorf("invalid modifier type: %s", modifier.Type)
	}

	// Validate source
	if modifier.Source.Type == "" {
		return fmt.Errorf("modifier source type is required")
	}

	if modifier.Source.Property == "" {
		return fmt.Errorf("modifier source property is required")
	}

	// Validate application
	if modifier.Application.Target == "" {
		return fmt.Errorf("modifier application target is required")
	}

	if modifier.Application.Operation == "" {
		return fmt.Errorf("modifier application operation is required")
	}

	// Validate operation
	validOperations := []string{"add", "multiply", "set", "conditional"}
	validOperation := false
	for _, validOp := range validOperations {
		if modifier.Application.Operation == validOp {
			validOperation = true
			break
		}
	}

	if !validOperation {
		return fmt.Errorf("invalid modifier operation: %s", modifier.Application.Operation)
	}

	return nil
}

// checkModifierConditions checks if modifier conditions are met
func (mm *ModifierManagerImpl) checkModifierConditions(modifier *domain.ModifierSystem, context map[string]interface{}) (bool, error) {
	// If no conditions, modifier always applies
	if len(modifier.Application.Conditions) == 0 {
		return true, nil
	}

	// TODO: Implement condition evaluation using expression parser
	// For now, return true if no conditions are specified
	return true, nil
}

// applyModifier applies a modifier and returns the result
func (mm *ModifierManagerImpl) applyModifier(modifier *domain.ModifierSystem, context map[string]interface{}) domain.ModifierResult {
	// TODO: Implement actual modifier application using expression parser
	// For now, return a placeholder result
	return domain.ModifierResult{
		ModifierID: modifier.ID,
		Value:      0,
		Operation:  modifier.Application.Operation,
		Success:    true,
	}
}
