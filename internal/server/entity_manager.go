package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/solo7.media/platform.drifter.solo7.media/internal/domain"
)

// EntityManagerImpl implements the EntityManager interface
type EntityManagerImpl struct {
	entities map[domain.EntityId]*domain.Entity
	mu       sync.RWMutex
	logger   domain.Logger
}

// NewEntityManager creates a new entity manager
func NewEntityManager(logger domain.Logger) *EntityManagerImpl {
	return &EntityManagerImpl{
		entities: make(map[domain.EntityId]*domain.Entity),
		logger:   logger,
	}
}

// CreateEntity creates a new entity with the given components
func (em *EntityManagerImpl) CreateEntity(ctx context.Context, components map[domain.ComponentType]domain.Component) (*domain.Entity, error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	entityID := uuid.New()

	entity := &domain.Entity{
		ID:         entityID,
		Components: components,
		Metadata: domain.EntityMetadata{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Version:   1,
		},
	}

	em.entities[entityID] = entity

	em.logger.Debug("Entity created", map[string]interface{}{
		"entity_id":       entityID,
		"component_count": len(components),
	})

	return entity, nil
}

// GetEntity retrieves an entity by ID
func (em *EntityManagerImpl) GetEntity(ctx context.Context, id domain.EntityId) (*domain.Entity, error) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	entity, exists := em.entities[id]
	if !exists {
		return nil, fmt.Errorf("entity %s not found", id)
	}

	return entity, nil
}

// UpdateEntity updates an entity with new components
func (em *EntityManagerImpl) UpdateEntity(ctx context.Context, id domain.EntityId, updates map[domain.ComponentType]domain.Component) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	entity, exists := em.entities[id]
	if !exists {
		return fmt.Errorf("entity %s not found", id)
	}

	// Update components
	for componentType, component := range updates {
		entity.Components[componentType] = component
	}

	// Update metadata
	entity.Metadata.UpdatedAt = time.Now()
	entity.Metadata.Version++

	em.logger.Debug("Entity updated", map[string]interface{}{
		"entity_id": id,
		"version":   entity.Metadata.Version,
	})

	return nil
}

// DeleteEntity removes an entity
func (em *EntityManagerImpl) DeleteEntity(ctx context.Context, id domain.EntityId) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	_, exists := em.entities[id]
	if !exists {
		return fmt.Errorf("entity %s not found", id)
	}

	delete(em.entities, id)

	em.logger.Debug("Entity deleted", map[string]interface{}{
		"entity_id": id,
	})

	return nil
}

// GetEntitiesInRegion retrieves all entities in a specific region
func (em *EntityManagerImpl) GetEntitiesInRegion(ctx context.Context, regionID domain.RegionId) ([]*domain.Entity, error) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	var entities []*domain.Entity
	for _, entity := range em.entities {
		// TODO: Implement region-based filtering
		// For now, return all entities
		entities = append(entities, entity)
	}

	return entities, nil
}

// GetEntitiesInArea retrieves entities within a specific area
func (em *EntityManagerImpl) GetEntitiesInArea(ctx context.Context, regionID domain.RegionId, center domain.Vector3, radius float64) ([]*domain.Entity, error) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	var entities []*domain.Entity
	for _, entity := range em.entities {
		// Check if entity has transform component
		transformComp, exists := entity.Components["transform"]
		if !exists {
			continue
		}

		transformData, ok := transformComp.Data.(domain.TransformComponent)
		if !ok {
			continue
		}

		// Calculate distance from center
		dx := transformData.Position.X - center.X
		dy := transformData.Position.Y - center.Y
		dz := transformData.Position.Z - center.Z
		distance := dx*dx + dy*dy + dz*dz

		if distance <= radius*radius {
			entities = append(entities, entity)
		}
	}

	return entities, nil
}
