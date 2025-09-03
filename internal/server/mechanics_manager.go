package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/solo7.media/platform.drifter.solo7.media/internal/domain"
)

// MechanicsManagerImpl implements the GameMechanicManager interface
type MechanicsManagerImpl struct {
	mechanics map[domain.MechanicId]*domain.GameMechanic
	mu        sync.RWMutex
	logger    domain.Logger
}

// NewMechanicsManager creates a new mechanics manager
func NewMechanicsManager(logger domain.Logger) *MechanicsManagerImpl {
	return &MechanicsManagerImpl{
		mechanics: make(map[domain.MechanicId]*domain.GameMechanic),
		logger:    logger,
	}
}

// RegisterMechanic registers a new game mechanic
func (mm *MechanicsManagerImpl) RegisterMechanic(ctx context.Context, mechanic *domain.GameMechanic) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if mechanic.ID == "" {
		return fmt.Errorf("mechanic ID cannot be empty")
	}

	if mechanic.Name == "" {
		return fmt.Errorf("mechanic name cannot be empty")
	}

	// Validate the mechanic
	if err := mm.validateMechanic(mechanic); err != nil {
		return fmt.Errorf("mechanic validation failed: %w", err)
	}

	// Check if mechanic already exists
	if _, exists := mm.mechanics[mechanic.ID]; exists {
		return fmt.Errorf("mechanic with ID %s already exists", mechanic.ID)
	}

	// Set version if not set
	if mechanic.Version == 0 {
		mechanic.Version = 1
	}

	mm.mechanics[mechanic.ID] = mechanic

	mm.logger.Info("Mechanic registered", map[string]interface{}{
		"mechanic_id":   mechanic.ID,
		"mechanic_name": mechanic.Name,
		"mechanic_type": mechanic.Type,
		"version":       mechanic.Version,
	})

	return nil
}

// UnregisterMechanic removes a game mechanic
func (mm *MechanicsManagerImpl) UnregisterMechanic(ctx context.Context, mechanicID domain.MechanicId) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	mechanic, exists := mm.mechanics[mechanicID]
	if !exists {
		return fmt.Errorf("mechanic with ID %s not found", mechanicID)
	}

	delete(mm.mechanics, mechanicID)

	mm.logger.Info("Mechanic unregistered", map[string]interface{}{
		"mechanic_id":   mechanicID,
		"mechanic_name": mechanic.Name,
	})

	return nil
}

// GetMechanic retrieves a game mechanic by ID
func (mm *MechanicsManagerImpl) GetMechanic(ctx context.Context, mechanicID domain.MechanicId) (*domain.GameMechanic, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	mechanic, exists := mm.mechanics[mechanicID]
	if !exists {
		return nil, fmt.Errorf("mechanic with ID %s not found", mechanicID)
	}

	// Return a copy to prevent external modification
	mechanicCopy := *mechanic
	return &mechanicCopy, nil
}

// GetMechanicsByType retrieves all mechanics of a specific type
func (mm *MechanicsManagerImpl) GetMechanicsByType(ctx context.Context, mechanicType domain.MechanicType) ([]*domain.GameMechanic, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	var mechanics []*domain.GameMechanic
	for _, mechanic := range mm.mechanics {
		if mechanic.Type == mechanicType {
			// Return a copy to prevent external modification
			mechanicCopy := *mechanic
			mechanics = append(mechanics, &mechanicCopy)
		}
	}

	return mechanics, nil
}

// ValidateMechanic validates a game mechanic
func (mm *MechanicsManagerImpl) ValidateMechanic(ctx context.Context, mechanic *domain.GameMechanic) error {
	return mm.validateMechanic(mechanic)
}

// ListMechanics returns all registered mechanics
func (mm *MechanicsManagerImpl) ListMechanics(ctx context.Context) ([]*domain.GameMechanic, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	var mechanics []*domain.GameMechanic
	for _, mechanic := range mm.mechanics {
		// Return a copy to prevent external modification
		mechanicCopy := *mechanic
		mechanics = append(mechanics, &mechanicCopy)
	}

	return mechanics, nil
}

// validateMechanic performs internal validation of a mechanic
func (mm *MechanicsManagerImpl) validateMechanic(mechanic *domain.GameMechanic) error {
	if mechanic == nil {
		return fmt.Errorf("mechanic cannot be nil")
	}

	if mechanic.ID == "" {
		return fmt.Errorf("mechanic ID is required")
	}

	if mechanic.Name == "" {
		return fmt.Errorf("mechanic name is required")
	}

	if mechanic.Type == "" {
		return fmt.Errorf("mechanic type is required")
	}

	// Validate mechanic type
	validTypes := []domain.MechanicType{
		domain.MechanicTypeCombat,
		domain.MechanicTypeMovement,
		domain.MechanicTypeInteraction,
		domain.MechanicTypeSkill,
		domain.MechanicTypeMagic,
		domain.MechanicTypeSocial,
	}

	validType := false
	for _, validTypeValue := range validTypes {
		if mechanic.Type == validTypeValue {
			validType = true
			break
		}
	}

	if !validType {
		return fmt.Errorf("invalid mechanic type: %s", mechanic.Type)
	}

	// Validate required inputs
	for _, input := range mechanic.Validation.RequiredInputs {
		if input.Name == "" {
			return fmt.Errorf("input requirement name cannot be empty")
		}
		if input.Type == "" {
			return fmt.Errorf("input requirement type cannot be empty for input: %s", input.Name)
		}
	}

	// Validate output validation
	if len(mechanic.Validation.OutputValidation.ExpectedTypes) == 0 {
		mm.logger.Warn("No expected output types defined for mechanic", map[string]interface{}{
			"mechanic_id": mechanic.ID,
		})
	}

	return nil
}
