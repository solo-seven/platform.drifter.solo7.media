package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/solo7.media/platform.drifter.solo7.media/internal/domain"
)

func TestEntity_Creation(t *testing.T) {
	t.Run("should create entity with valid components", func(t *testing.T) {
		// Given
		entityID := uuid.New()
		components := map[domain.ComponentType]domain.Component{
			"transform": {
				Type: "transform",
				Data: domain.TransformComponent{
					Position: domain.Vector3{X: 1.0, Y: 2.0, Z: 3.0},
					Rotation: domain.Vector3{X: 0.0, Y: 0.0, Z: 0.0},
					Scale:    domain.Vector3{X: 1.0, Y: 1.0, Z: 1.0},
				},
				Version: 1,
			},
		}

		// When
		entity := &domain.Entity{
			ID:         entityID,
			Components: components,
			Metadata: domain.EntityMetadata{
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Version:   1,
			},
		}

		// Then
		assert.Equal(t, entityID, entity.ID)
		assert.Len(t, entity.Components, 1)
		assert.Contains(t, entity.Components, "transform")
		
		transformComp := entity.Components["transform"]
		assert.Equal(t, "transform", transformComp.Type)
		assert.Equal(t, 1, transformComp.Version)
		
		transformData, ok := transformComp.Data.(domain.TransformComponent)
		require.True(t, ok)
		assert.Equal(t, 1.0, transformData.Position.X)
		assert.Equal(t, 2.0, transformData.Position.Y)
		assert.Equal(t, 3.0, transformData.Position.Z)
	})

	t.Run("should create entity with multiple components", func(t *testing.T) {
		// Given
		entityID := uuid.New()
		components := map[domain.ComponentType]domain.Component{
			"transform": {
				Type: "transform",
				Data: domain.TransformComponent{
					Position: domain.Vector3{X: 0.0, Y: 0.0, Z: 0.0},
					Rotation: domain.Vector3{X: 0.0, Y: 0.0, Z: 0.0},
					Scale:    domain.Vector3{X: 1.0, Y: 1.0, Z: 1.0},
				},
				Version: 1,
			},
			"physics": {
				Type: "physics",
				Data: domain.PhysicsComponent{
					Mass: 1.0,
					Velocity: domain.Vector3{X: 0.0, Y: 0.0, Z: 0.0},
					IsStatic: false,
				},
				Version: 1,
			},
			"gameplay": {
				Type: "gameplay",
				Data: domain.GameplayComponent{
					Stats: map[string]float64{
						"health": 100.0,
						"mana":   50.0,
					},
					Abilities: []domain.Ability{},
					Inventory: domain.Inventory{
						MaxSlots:  20,
						MaxWeight: 100.0,
					},
					StatusEffects: []domain.StatusEffect{},
				},
				Version: 1,
			},
		}

		// When
		entity := &domain.Entity{
			ID:         entityID,
			Components: components,
			Metadata: domain.EntityMetadata{
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Version:   1,
			},
		}

		// Then
		assert.Equal(t, entityID, entity.ID)
		assert.Len(t, entity.Components, 3)
		assert.Contains(t, entity.Components, "transform")
		assert.Contains(t, entity.Components, "physics")
		assert.Contains(t, entity.Components, "gameplay")
		
		// Verify gameplay component data
		gameplayComp := entity.Components["gameplay"]
		gameplayData, ok := gameplayComp.Data.(domain.GameplayComponent)
		require.True(t, ok)
		assert.Equal(t, 100.0, gameplayData.Stats["health"])
		assert.Equal(t, 50.0, gameplayData.Stats["mana"])
		assert.Equal(t, 20, gameplayData.Inventory.MaxSlots)
		assert.Equal(t, 100.0, gameplayData.Inventory.MaxWeight)
	})
}

func TestEntity_ComponentAccess(t *testing.T) {
	t.Run("should access transform component", func(t *testing.T) {
		// Given
		entity := createTestEntity(t)
		
		// When
		transformComp, exists := entity.Components["transform"]
		
		// Then
		require.True(t, exists)
		assert.Equal(t, "transform", transformComp.Type)
		
		transformData, ok := transformComp.Data.(domain.TransformComponent)
		require.True(t, ok)
		assert.Equal(t, 5.0, transformData.Position.X)
		assert.Equal(t, 10.0, transformData.Position.Y)
		assert.Equal(t, 15.0, transformData.Position.Z)
	})

	t.Run("should return false for non-existent component", func(t *testing.T) {
		// Given
		entity := createTestEntity(t)
		
		// When
		_, exists := entity.Components["non_existent"]
		
		// Then
		assert.False(t, exists)
	})
}

func TestEntity_ComponentUpdate(t *testing.T) {
	t.Run("should update component version", func(t *testing.T) {
		// Given
		entity := createTestEntity(t)
		originalVersion := entity.Components["transform"].Version
		
		// When
		transformComp := entity.Components["transform"]
		transformComp.Version++
		entity.Components["transform"] = transformComp
		
		// Then
		assert.Equal(t, originalVersion+1, entity.Components["transform"].Version)
	})

	t.Run("should update component data", func(t *testing.T) {
		// Given
		entity := createTestEntity(t)
		
		// When
		transformComp := entity.Components["transform"]
		transformData := transformComp.Data.(domain.TransformComponent)
		transformData.Position.X = 999.0
		transformComp.Data = transformData
		entity.Components["transform"] = transformComp
		
		// Then
		updatedTransformData := entity.Components["transform"].Data.(domain.TransformComponent)
		assert.Equal(t, 999.0, updatedTransformData.Position.X)
	})
}

func TestEntity_Metadata(t *testing.T) {
	t.Run("should have valid metadata", func(t *testing.T) {
		// Given
		now := time.Now()
		entity := &domain.Entity{
			ID:         uuid.New(),
			Components: make(map[domain.ComponentType]domain.Component),
			Metadata: domain.EntityMetadata{
				CreatedAt: now,
				UpdatedAt: now,
				Version:   1,
			},
		}
		
		// Then
		assert.Equal(t, now, entity.Metadata.CreatedAt)
		assert.Equal(t, now, entity.Metadata.UpdatedAt)
		assert.Equal(t, 1, entity.Metadata.Version)
	})

	t.Run("should update metadata timestamp", func(t *testing.T) {
		// Given
		entity := createTestEntity(t)
		originalUpdatedAt := entity.Metadata.UpdatedAt
		
		// When
		time.Sleep(1 * time.Millisecond) // Ensure time difference
		entity.Metadata.UpdatedAt = time.Now()
		
		// Then
		assert.True(t, entity.Metadata.UpdatedAt.After(originalUpdatedAt))
	})
}

// Helper function to create a test entity
func createTestEntity(t *testing.T) *domain.Entity {
	return &domain.Entity{
		ID: uuid.New(),
		Components: map[domain.ComponentType]domain.Component{
			"transform": {
				Type: "transform",
				Data: domain.TransformComponent{
					Position: domain.Vector3{X: 5.0, Y: 10.0, Z: 15.0},
					Rotation: domain.Vector3{X: 0.0, Y: 0.0, Z: 0.0},
					Scale:    domain.Vector3{X: 1.0, Y: 1.0, Z: 1.0},
				},
				Version: 1,
			},
		},
		Metadata: domain.EntityMetadata{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Version:   1,
		},
	}
}

