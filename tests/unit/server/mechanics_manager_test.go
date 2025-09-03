package server_test

import (
	"context"
	"testing"

	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
	"github.com/solo-seven/platform.drifter.solo7.media/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMechanicsManager_RegisterMechanic(t *testing.T) {
	mockLogger := &MockLogger{}
	manager := server.NewMechanicsManager(mockLogger)
	ctx := context.Background()

	t.Run("successful registration", func(t *testing.T) {
		mechanic := &domain.GameMechanic{
			ID:          "combat.basic_attack",
			Name:        "Basic Attack",
			Description: "A basic melee attack",
			Type:        domain.MechanicTypeCombat,
			Properties:  map[string]interface{}{"damage_type": "physical"},
			Validation: domain.MechanicValidation{
				RequiredInputs: []domain.InputRequirement{
					{
						Name:     "attacker",
						Type:     "entity",
						Required: true,
					},
					{
						Name:     "target",
						Type:     "entity",
						Required: true,
					},
				},
				OutputValidation: domain.OutputValidation{
					ExpectedTypes:  []string{"number"},
					RequiredFields: []string{"damage"},
				},
				ContextValidation: domain.ContextValidation{
					RequiredEntities: []string{"attacker", "target"},
					RequiredState:    []string{"combat_active"},
				},
			},
			Version: 1,
		}

		mockLogger.On("Info", "Mechanic registered", mock.AnythingOfType("map[string]interface {}")).Return()

		err := manager.RegisterMechanic(ctx, mechanic)
		assert.NoError(t, err)

		// Verify the mechanic was registered
		retrieved, err := manager.GetMechanic(ctx, "combat.basic_attack")
		assert.NoError(t, err)
		assert.Equal(t, mechanic.ID, retrieved.ID)
		assert.Equal(t, mechanic.Name, retrieved.Name)
		assert.Equal(t, mechanic.Type, retrieved.Type)

		mockLogger.AssertExpectations(t)
	})

	t.Run("duplicate registration", func(t *testing.T) {
		mechanic := &domain.GameMechanic{
			ID:          "combat.basic_attack",
			Name:        "Basic Attack",
			Description: "A basic melee attack",
			Type:        domain.MechanicTypeCombat,
			Validation:  domain.MechanicValidation{},
		}

		// Expect a warn call for the validation warning
		mockLogger.On("Warn", "No expected output types defined for mechanic", mock.AnythingOfType("map[string]interface {}")).Return()

		err := manager.RegisterMechanic(ctx, mechanic)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("invalid mechanic - empty ID", func(t *testing.T) {
		mechanic := &domain.GameMechanic{
			ID:   "",
			Name: "Test Mechanic",
			Type: domain.MechanicTypeCombat,
		}

		err := manager.RegisterMechanic(ctx, mechanic)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ID cannot be empty")
	})

	t.Run("invalid mechanic - empty name", func(t *testing.T) {
		mechanic := &domain.GameMechanic{
			ID:   "test.mechanic",
			Name: "",
			Type: domain.MechanicTypeCombat,
		}

		err := manager.RegisterMechanic(ctx, mechanic)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name cannot be empty")
	})

	t.Run("invalid mechanic - invalid type", func(t *testing.T) {
		mechanic := &domain.GameMechanic{
			ID:   "test.mechanic",
			Name: "Test Mechanic",
			Type: "invalid_type",
		}

		err := manager.RegisterMechanic(ctx, mechanic)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid mechanic type")
	})
}

func TestMechanicsManager_UnregisterMechanic(t *testing.T) {
	mockLogger := &MockLogger{}
	manager := server.NewMechanicsManager(mockLogger)
	ctx := context.Background()

	// Register a mechanic first
	mechanic := &domain.GameMechanic{
		ID:          "test.mechanic",
		Name:        "Test Mechanic",
		Description: "A test mechanic",
		Type:        domain.MechanicTypeCombat,
		Validation:  domain.MechanicValidation{},
	}

	mockLogger.On("Warn", "No expected output types defined for mechanic", mock.AnythingOfType("map[string]interface {}")).Return()
	mockLogger.On("Info", "Mechanic registered", mock.AnythingOfType("map[string]interface {}")).Return()
	err := manager.RegisterMechanic(ctx, mechanic)
	assert.NoError(t, err)

	t.Run("successful unregistration", func(t *testing.T) {
		mockLogger.On("Info", "Mechanic unregistered", mock.AnythingOfType("map[string]interface {}")).Return()

		err := manager.UnregisterMechanic(ctx, "test.mechanic")
		assert.NoError(t, err)

		// Verify the mechanic was removed
		_, err = manager.GetMechanic(ctx, "test.mechanic")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		mockLogger.AssertExpectations(t)
	})

	t.Run("unregister non-existent mechanic", func(t *testing.T) {
		err := manager.UnregisterMechanic(ctx, "non.existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestMechanicsManager_GetMechanicsByType(t *testing.T) {
	mockLogger := &MockLogger{}
	manager := server.NewMechanicsManager(mockLogger)
	ctx := context.Background()

	// Register multiple mechanics of different types
	combatMechanic := &domain.GameMechanic{
		ID:          "combat.attack",
		Name:        "Attack",
		Description: "Combat attack",
		Type:        domain.MechanicTypeCombat,
		Validation:  domain.MechanicValidation{},
	}

	movementMechanic := &domain.GameMechanic{
		ID:          "movement.walk",
		Name:        "Walk",
		Description: "Basic movement",
		Type:        domain.MechanicTypeMovement,
		Validation:  domain.MechanicValidation{},
	}

	anotherCombatMechanic := &domain.GameMechanic{
		ID:          "combat.defend",
		Name:        "Defend",
		Description: "Combat defense",
		Type:        domain.MechanicTypeCombat,
		Validation:  domain.MechanicValidation{},
	}

	// Expect warn calls for validation warnings and info calls for registration
	mockLogger.On("Warn", "No expected output types defined for mechanic", mock.AnythingOfType("map[string]interface {}")).Return().Times(3)
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}")).Return().Times(3)

	err := manager.RegisterMechanic(ctx, combatMechanic)
	assert.NoError(t, err)

	err = manager.RegisterMechanic(ctx, movementMechanic)
	assert.NoError(t, err)

	err = manager.RegisterMechanic(ctx, anotherCombatMechanic)
	assert.NoError(t, err)

	t.Run("get combat mechanics", func(t *testing.T) {
		mechanics, err := manager.GetMechanicsByType(ctx, domain.MechanicTypeCombat)
		assert.NoError(t, err)
		assert.Len(t, mechanics, 2)

		// Verify we got the right mechanics
		mechanicIDs := make([]string, len(mechanics))
		for i, m := range mechanics {
			mechanicIDs[i] = m.ID
		}
		assert.Contains(t, mechanicIDs, "combat.attack")
		assert.Contains(t, mechanicIDs, "combat.defend")
	})

	t.Run("get movement mechanics", func(t *testing.T) {
		mechanics, err := manager.GetMechanicsByType(ctx, domain.MechanicTypeMovement)
		assert.NoError(t, err)
		assert.Len(t, mechanics, 1)
		assert.Equal(t, "movement.walk", mechanics[0].ID)
	})

	t.Run("get non-existent type", func(t *testing.T) {
		mechanics, err := manager.GetMechanicsByType(ctx, domain.MechanicTypeMagic)
		assert.NoError(t, err)
		assert.Len(t, mechanics, 0)
	})
}

func TestMechanicsManager_ListMechanics(t *testing.T) {
	mockLogger := &MockLogger{}
	manager := server.NewMechanicsManager(mockLogger)
	ctx := context.Background()

	// Register multiple mechanics
	mechanic1 := &domain.GameMechanic{
		ID:         "test.mechanic1",
		Name:       "Test Mechanic 1",
		Type:       domain.MechanicTypeCombat,
		Validation: domain.MechanicValidation{},
	}

	mechanic2 := &domain.GameMechanic{
		ID:         "test.mechanic2",
		Name:       "Test Mechanic 2",
		Type:       domain.MechanicTypeMovement,
		Validation: domain.MechanicValidation{},
	}

	// Expect warn calls for validation warnings and info calls for registration
	mockLogger.On("Warn", "No expected output types defined for mechanic", mock.AnythingOfType("map[string]interface {}")).Return().Times(2)
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}")).Return().Times(2)

	err := manager.RegisterMechanic(ctx, mechanic1)
	assert.NoError(t, err)

	err = manager.RegisterMechanic(ctx, mechanic2)
	assert.NoError(t, err)

	mechanics, err := manager.ListMechanics(ctx)
	assert.NoError(t, err)
	assert.Len(t, mechanics, 2)

	// Verify we got both mechanics
	mechanicIDs := make([]string, len(mechanics))
	for i, m := range mechanics {
		mechanicIDs[i] = m.ID
	}
	assert.Contains(t, mechanicIDs, "test.mechanic1")
	assert.Contains(t, mechanicIDs, "test.mechanic2")
}

func TestMechanicsManager_ValidateMechanic(t *testing.T) {
	mockLogger := &MockLogger{}
	manager := server.NewMechanicsManager(mockLogger)
	ctx := context.Background()

	t.Run("valid mechanic", func(t *testing.T) {
		mechanic := &domain.GameMechanic{
			ID:          "test.mechanic",
			Name:        "Test Mechanic",
			Description: "A test mechanic",
			Type:        domain.MechanicTypeCombat,
			Validation: domain.MechanicValidation{
				RequiredInputs: []domain.InputRequirement{
					{
						Name:     "input1",
						Type:     "string",
						Required: true,
					},
				},
				OutputValidation: domain.OutputValidation{
					ExpectedTypes: []string{"number"},
				},
			},
		}

		err := manager.ValidateMechanic(ctx, mechanic)
		assert.NoError(t, err)
	})

	t.Run("invalid mechanic - nil", func(t *testing.T) {
		err := manager.ValidateMechanic(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("invalid mechanic - empty input name", func(t *testing.T) {
		mechanic := &domain.GameMechanic{
			ID:   "test.mechanic",
			Name: "Test Mechanic",
			Type: domain.MechanicTypeCombat,
			Validation: domain.MechanicValidation{
				RequiredInputs: []domain.InputRequirement{
					{
						Name: "", // Empty name
						Type: "string",
					},
				},
			},
		}

		err := manager.ValidateMechanic(ctx, mechanic)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "input requirement name cannot be empty")
	})
}
