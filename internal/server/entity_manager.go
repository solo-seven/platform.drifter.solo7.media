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
	entities         map[domain.EntityId]*domain.Entity
	componentSystems map[domain.ComponentType]ComponentSystem
	mu               sync.RWMutex
	logger           domain.Logger
}

// ComponentSystem defines operations for a specific component type
type ComponentSystem interface {
	GetType() domain.ComponentType
	Create(data interface{}) (domain.Component, error)
	Validate(component domain.Component) error
	Update(existing domain.Component, updates map[string]interface{}) (domain.Component, error)
	Serialize(component domain.Component) (map[string]interface{}, error)
	Deserialize(data map[string]interface{}) (domain.Component, error)
}

// NewEntityManager creates a new entity manager
func NewEntityManager(logger domain.Logger) *EntityManagerImpl {
	em := &EntityManagerImpl{
		entities:         make(map[domain.EntityId]*domain.Entity),
		componentSystems: make(map[domain.ComponentType]ComponentSystem),
		logger:           logger,
	}

	// Register built-in component systems
	em.registerBuiltinComponentSystems()

	return em
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

// RegisterComponentSystem registers a new component system
func (em *EntityManagerImpl) RegisterComponentSystem(system ComponentSystem) {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.componentSystems[system.GetType()] = system
	em.logger.Debug("Component system registered", map[string]interface{}{
		"component_type": system.GetType(),
	})
}

// GetComponentSystem retrieves a component system by type
func (em *EntityManagerImpl) GetComponentSystem(componentType domain.ComponentType) (ComponentSystem, bool) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	system, exists := em.componentSystems[componentType]
	return system, exists
}

// CreateComponent creates a component using the appropriate component system
func (em *EntityManagerImpl) CreateComponent(componentType domain.ComponentType, data interface{}) (domain.Component, error) {
	system, exists := em.GetComponentSystem(componentType)
	if !exists {
		return domain.Component{}, fmt.Errorf("unknown component type: %s", componentType)
	}

	return system.Create(data)
}

// ValidateComponent validates a component using its component system
func (em *EntityManagerImpl) ValidateComponent(component domain.Component) error {
	system, exists := em.GetComponentSystem(component.Type)
	if !exists {
		return fmt.Errorf("unknown component type: %s", component.Type)
	}

	return system.Validate(component)
}

// UpdateComponent updates a component using its component system
func (em *EntityManagerImpl) UpdateComponent(existing domain.Component, updates map[string]interface{}) (domain.Component, error) {
	system, exists := em.GetComponentSystem(existing.Type)
	if !exists {
		return domain.Component{}, fmt.Errorf("unknown component type: %s", existing.Type)
	}

	return system.Update(existing, updates)
}

// GetEntitiesWithComponent retrieves all entities that have a specific component type
func (em *EntityManagerImpl) GetEntitiesWithComponent(ctx context.Context, componentType domain.ComponentType) ([]*domain.Entity, error) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	var entities []*domain.Entity
	for _, entity := range em.entities {
		if _, exists := entity.Components[componentType]; exists {
			entities = append(entities, entity)
		}
	}

	return entities, nil
}

// GetEntitiesWithComponents retrieves entities that have all specified component types
func (em *EntityManagerImpl) GetEntitiesWithComponents(ctx context.Context, componentTypes []domain.ComponentType) ([]*domain.Entity, error) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	var entities []*domain.Entity
	for _, entity := range em.entities {
		hasAll := true
		for _, componentType := range componentTypes {
			if _, exists := entity.Components[componentType]; !exists {
				hasAll = false
				break
			}
		}
		if hasAll {
			entities = append(entities, entity)
		}
	}

	return entities, nil
}

// registerBuiltinComponentSystems registers the built-in component systems
func (em *EntityManagerImpl) registerBuiltinComponentSystems() {
	em.RegisterComponentSystem(&TransformComponentSystem{})
	em.RegisterComponentSystem(&PhysicsComponentSystem{})
	em.RegisterComponentSystem(&GameplayComponentSystem{})
	em.RegisterComponentSystem(&RenderableComponentSystem{})
	em.RegisterComponentSystem(&InteractiveComponentSystem{})
	em.RegisterComponentSystem(&NetworkComponentSystem{})
}

// Built-in Component System Implementations

// TransformComponentSystem handles transform components
type TransformComponentSystem struct{}

func (tcs *TransformComponentSystem) GetType() domain.ComponentType {
	return "transform"
}

func (tcs *TransformComponentSystem) Create(data interface{}) (domain.Component, error) {
	transformData, ok := data.(domain.TransformComponent)
	if !ok {
		return domain.Component{}, fmt.Errorf("invalid transform component data")
	}

	return domain.Component{
		Type:    "transform",
		Data:    transformData,
		Version: 1,
	}, nil
}

func (tcs *TransformComponentSystem) Validate(component domain.Component) error {
	if component.Type != "transform" {
		return fmt.Errorf("invalid component type for transform system")
	}

	_, ok := component.Data.(domain.TransformComponent)
	if !ok {
		return fmt.Errorf("invalid transform component data")
	}

	return nil
}

func (tcs *TransformComponentSystem) Update(existing domain.Component, updates map[string]interface{}) (domain.Component, error) {
	transform, ok := existing.Data.(domain.TransformComponent)
	if !ok {
		return domain.Component{}, fmt.Errorf("invalid existing transform component")
	}

	// Update position
	if pos, exists := updates["position"]; exists {
		if posMap, ok := pos.(map[string]interface{}); ok {
			if x, ok := posMap["x"].(float64); ok {
				transform.Position.X = x
			}
			if y, ok := posMap["y"].(float64); ok {
				transform.Position.Y = y
			}
			if z, ok := posMap["z"].(float64); ok {
				transform.Position.Z = z
			}
		}
	}

	// Update rotation
	if rot, exists := updates["rotation"]; exists {
		if rotMap, ok := rot.(map[string]interface{}); ok {
			if x, ok := rotMap["x"].(float64); ok {
				transform.Rotation.X = x
			}
			if y, ok := rotMap["y"].(float64); ok {
				transform.Rotation.Y = y
			}
			if z, ok := rotMap["z"].(float64); ok {
				transform.Rotation.Z = z
			}
		}
	}

	// Update scale
	if scale, exists := updates["scale"]; exists {
		if scaleMap, ok := scale.(map[string]interface{}); ok {
			if x, ok := scaleMap["x"].(float64); ok {
				transform.Scale.X = x
			}
			if y, ok := scaleMap["y"].(float64); ok {
				transform.Scale.Y = y
			}
			if z, ok := scaleMap["z"].(float64); ok {
				transform.Scale.Z = z
			}
		}
	}

	return domain.Component{
		Type:    "transform",
		Data:    transform,
		Version: existing.Version + 1,
	}, nil
}

func (tcs *TransformComponentSystem) Serialize(component domain.Component) (map[string]interface{}, error) {
	transform, ok := component.Data.(domain.TransformComponent)
	if !ok {
		return nil, fmt.Errorf("invalid transform component")
	}

	return map[string]interface{}{
		"position": map[string]interface{}{
			"x": transform.Position.X,
			"y": transform.Position.Y,
			"z": transform.Position.Z,
		},
		"rotation": map[string]interface{}{
			"x": transform.Rotation.X,
			"y": transform.Rotation.Y,
			"z": transform.Rotation.Z,
		},
		"scale": map[string]interface{}{
			"x": transform.Scale.X,
			"y": transform.Scale.Y,
			"z": transform.Scale.Z,
		},
	}, nil
}

func (tcs *TransformComponentSystem) Deserialize(data map[string]interface{}) (domain.Component, error) {
	transform := domain.TransformComponent{}

	// Deserialize position
	if pos, exists := data["position"]; exists {
		if posMap, ok := pos.(map[string]interface{}); ok {
			if x, ok := posMap["x"].(float64); ok {
				transform.Position.X = x
			}
			if y, ok := posMap["y"].(float64); ok {
				transform.Position.Y = y
			}
			if z, ok := posMap["z"].(float64); ok {
				transform.Position.Z = z
			}
		}
	}

	// Deserialize rotation
	if rot, exists := data["rotation"]; exists {
		if rotMap, ok := rot.(map[string]interface{}); ok {
			if x, ok := rotMap["x"].(float64); ok {
				transform.Rotation.X = x
			}
			if y, ok := rotMap["y"].(float64); ok {
				transform.Rotation.Y = y
			}
			if z, ok := rotMap["z"].(float64); ok {
				transform.Rotation.Z = z
			}
		}
	}

	// Deserialize scale
	if scale, exists := data["scale"]; exists {
		if scaleMap, ok := scale.(map[string]interface{}); ok {
			if x, ok := scaleMap["x"].(float64); ok {
				transform.Scale.X = x
			}
			if y, ok := scaleMap["y"].(float64); ok {
				transform.Scale.Y = y
			}
			if z, ok := scaleMap["z"].(float64); ok {
				transform.Scale.Z = z
			}
		}
	}

	return domain.Component{
		Type:    "transform",
		Data:    transform,
		Version: 1,
	}, nil
}

// PhysicsComponentSystem handles physics components
type PhysicsComponentSystem struct{}

func (pcs *PhysicsComponentSystem) GetType() domain.ComponentType {
	return "physics"
}

func (pcs *PhysicsComponentSystem) Create(data interface{}) (domain.Component, error) {
	physicsData, ok := data.(domain.PhysicsComponent)
	if !ok {
		return domain.Component{}, fmt.Errorf("invalid physics component data")
	}

	return domain.Component{
		Type:    "physics",
		Data:    physicsData,
		Version: 1,
	}, nil
}

func (pcs *PhysicsComponentSystem) Validate(component domain.Component) error {
	if component.Type != "physics" {
		return fmt.Errorf("invalid component type for physics system")
	}

	_, ok := component.Data.(domain.PhysicsComponent)
	if !ok {
		return fmt.Errorf("invalid physics component data")
	}

	return nil
}

func (pcs *PhysicsComponentSystem) Update(existing domain.Component, updates map[string]interface{}) (domain.Component, error) {
	physics, ok := existing.Data.(domain.PhysicsComponent)
	if !ok {
		return domain.Component{}, fmt.Errorf("invalid existing physics component")
	}

	// Update mass
	if mass, exists := updates["mass"]; exists {
		if massFloat, ok := mass.(float64); ok {
			physics.Mass = massFloat
		}
	}

	// Update velocity
	if vel, exists := updates["velocity"]; exists {
		if velMap, ok := vel.(map[string]interface{}); ok {
			if x, ok := velMap["x"].(float64); ok {
				physics.Velocity.X = x
			}
			if y, ok := velMap["y"].(float64); ok {
				physics.Velocity.Y = y
			}
			if z, ok := velMap["z"].(float64); ok {
				physics.Velocity.Z = z
			}
		}
	}

	// Update static flag
	if isStatic, exists := updates["is_static"]; exists {
		if staticBool, ok := isStatic.(bool); ok {
			physics.IsStatic = staticBool
		}
	}

	return domain.Component{
		Type:    "physics",
		Data:    physics,
		Version: existing.Version + 1,
	}, nil
}

func (pcs *PhysicsComponentSystem) Serialize(component domain.Component) (map[string]interface{}, error) {
	physics, ok := component.Data.(domain.PhysicsComponent)
	if !ok {
		return nil, fmt.Errorf("invalid physics component")
	}

	return map[string]interface{}{
		"mass": physics.Mass,
		"velocity": map[string]interface{}{
			"x": physics.Velocity.X,
			"y": physics.Velocity.Y,
			"z": physics.Velocity.Z,
		},
		"is_static": physics.IsStatic,
		"collision_bounds": map[string]interface{}{
			"min": map[string]interface{}{
				"x": physics.CollisionBounds.Min.X,
				"y": physics.CollisionBounds.Min.Y,
				"z": physics.CollisionBounds.Min.Z,
			},
			"max": map[string]interface{}{
				"x": physics.CollisionBounds.Max.X,
				"y": physics.CollisionBounds.Max.Y,
				"z": physics.CollisionBounds.Max.Z,
			},
		},
	}, nil
}

func (pcs *PhysicsComponentSystem) Deserialize(data map[string]interface{}) (domain.Component, error) {
	physics := domain.PhysicsComponent{}

	if mass, exists := data["mass"]; exists {
		if massFloat, ok := mass.(float64); ok {
			physics.Mass = massFloat
		}
	}

	if vel, exists := data["velocity"]; exists {
		if velMap, ok := vel.(map[string]interface{}); ok {
			if x, ok := velMap["x"].(float64); ok {
				physics.Velocity.X = x
			}
			if y, ok := velMap["y"].(float64); ok {
				physics.Velocity.Y = y
			}
			if z, ok := velMap["z"].(float64); ok {
				physics.Velocity.Z = z
			}
		}
	}

	if isStatic, exists := data["is_static"]; exists {
		if staticBool, ok := isStatic.(bool); ok {
			physics.IsStatic = staticBool
		}
	}

	return domain.Component{
		Type:    "physics",
		Data:    physics,
		Version: 1,
	}, nil
}

// GameplayComponentSystem handles gameplay components
type GameplayComponentSystem struct{}

func (gcs *GameplayComponentSystem) GetType() domain.ComponentType {
	return "gameplay"
}

func (gcs *GameplayComponentSystem) Create(data interface{}) (domain.Component, error) {
	gameplayData, ok := data.(domain.GameplayComponent)
	if !ok {
		return domain.Component{}, fmt.Errorf("invalid gameplay component data")
	}

	return domain.Component{
		Type:    "gameplay",
		Data:    gameplayData,
		Version: 1,
	}, nil
}

func (gcs *GameplayComponentSystem) Validate(component domain.Component) error {
	if component.Type != "gameplay" {
		return fmt.Errorf("invalid component type for gameplay system")
	}

	_, ok := component.Data.(domain.GameplayComponent)
	if !ok {
		return fmt.Errorf("invalid gameplay component data")
	}

	return nil
}

func (gcs *GameplayComponentSystem) Update(existing domain.Component, updates map[string]interface{}) (domain.Component, error) {
	gameplay, ok := existing.Data.(domain.GameplayComponent)
	if !ok {
		return domain.Component{}, fmt.Errorf("invalid existing gameplay component")
	}

	// Update stats
	if stats, exists := updates["stats"]; exists {
		if statsMap, ok := stats.(map[string]interface{}); ok {
			if gameplay.Stats == nil {
				gameplay.Stats = make(map[string]float64)
			}
			for key, value := range statsMap {
				if floatVal, ok := value.(float64); ok {
					gameplay.Stats[key] = floatVal
				}
			}
		}
	}

	return domain.Component{
		Type:    "gameplay",
		Data:    gameplay,
		Version: existing.Version + 1,
	}, nil
}

func (gcs *GameplayComponentSystem) Serialize(component domain.Component) (map[string]interface{}, error) {
	gameplay, ok := component.Data.(domain.GameplayComponent)
	if !ok {
		return nil, fmt.Errorf("invalid gameplay component")
	}

	return map[string]interface{}{
		"stats":          gameplay.Stats,
		"abilities":      gameplay.Abilities,
		"inventory":      gameplay.Inventory,
		"status_effects": gameplay.StatusEffects,
	}, nil
}

func (gcs *GameplayComponentSystem) Deserialize(data map[string]interface{}) (domain.Component, error) {
	gameplay := domain.GameplayComponent{
		Stats:         make(map[string]float64),
		Abilities:     []domain.Ability{},
		Inventory:     domain.Inventory{},
		StatusEffects: []domain.StatusEffect{},
	}

	if stats, exists := data["stats"]; exists {
		if statsMap, ok := stats.(map[string]interface{}); ok {
			for key, value := range statsMap {
				if floatVal, ok := value.(float64); ok {
					gameplay.Stats[key] = floatVal
				}
			}
		}
	}

	return domain.Component{
		Type:    "gameplay",
		Data:    gameplay,
		Version: 1,
	}, nil
}

// RenderableComponentSystem handles renderable components
type RenderableComponentSystem struct{}

func (rcs *RenderableComponentSystem) GetType() domain.ComponentType {
	return "renderable"
}

func (rcs *RenderableComponentSystem) Create(data interface{}) (domain.Component, error) {
	renderableData, ok := data.(domain.RenderableComponent)
	if !ok {
		return domain.Component{}, fmt.Errorf("invalid renderable component data")
	}

	return domain.Component{
		Type:    "renderable",
		Data:    renderableData,
		Version: 1,
	}, nil
}

func (rcs *RenderableComponentSystem) Validate(component domain.Component) error {
	if component.Type != "renderable" {
		return fmt.Errorf("invalid component type for renderable system")
	}

	_, ok := component.Data.(domain.RenderableComponent)
	if !ok {
		return fmt.Errorf("invalid renderable component data")
	}

	return nil
}

func (rcs *RenderableComponentSystem) Update(existing domain.Component, updates map[string]interface{}) (domain.Component, error) {
	renderable, ok := existing.Data.(domain.RenderableComponent)
	if !ok {
		return domain.Component{}, fmt.Errorf("invalid existing renderable component")
	}

	// Update asset ID
	if assetID, exists := updates["asset_id"]; exists {
		if assetStr, ok := assetID.(string); ok {
			renderable.AssetID = assetStr
		}
	}

	return domain.Component{
		Type:    "renderable",
		Data:    renderable,
		Version: existing.Version + 1,
	}, nil
}

func (rcs *RenderableComponentSystem) Serialize(component domain.Component) (map[string]interface{}, error) {
	renderable, ok := component.Data.(domain.RenderableComponent)
	if !ok {
		return nil, fmt.Errorf("invalid renderable component")
	}

	return map[string]interface{}{
		"asset_id":           renderable.AssetID,
		"animation_set_id":   renderable.AnimationSetID,
		"material_overrides": renderable.MaterialOverrides,
		"rendering_hints":    renderable.RenderingHints,
	}, nil
}

func (rcs *RenderableComponentSystem) Deserialize(data map[string]interface{}) (domain.Component, error) {
	renderable := domain.RenderableComponent{}

	if assetID, exists := data["asset_id"]; exists {
		if assetStr, ok := assetID.(string); ok {
			renderable.AssetID = assetStr
		}
	}

	if animSetID, exists := data["animation_set_id"]; exists {
		if animStr, ok := animSetID.(string); ok {
			renderable.AnimationSetID = &animStr
		}
	}

	return domain.Component{
		Type:    "renderable",
		Data:    renderable,
		Version: 1,
	}, nil
}

// InteractiveComponentSystem handles interactive components
type InteractiveComponentSystem struct{}

func (ics *InteractiveComponentSystem) GetType() domain.ComponentType {
	return "interactive"
}

func (ics *InteractiveComponentSystem) Create(data interface{}) (domain.Component, error) {
	interactiveData, ok := data.(domain.InteractiveComponent)
	if !ok {
		return domain.Component{}, fmt.Errorf("invalid interactive component data")
	}

	return domain.Component{
		Type:    "interactive",
		Data:    interactiveData,
		Version: 1,
	}, nil
}

func (ics *InteractiveComponentSystem) Validate(component domain.Component) error {
	if component.Type != "interactive" {
		return fmt.Errorf("invalid component type for interactive system")
	}

	_, ok := component.Data.(domain.InteractiveComponent)
	if !ok {
		return fmt.Errorf("invalid interactive component data")
	}

	return nil
}

func (ics *InteractiveComponentSystem) Update(existing domain.Component, updates map[string]interface{}) (domain.Component, error) {
	interactive, ok := existing.Data.(domain.InteractiveComponent)
	if !ok {
		return domain.Component{}, fmt.Errorf("invalid existing interactive component")
	}

	// Update interactable flag
	if isInteractable, exists := updates["is_interactable"]; exists {
		if interactableBool, ok := isInteractable.(bool); ok {
			interactive.IsInteractable = interactableBool
		}
	}

	return domain.Component{
		Type:    "interactive",
		Data:    interactive,
		Version: existing.Version + 1,
	}, nil
}

func (ics *InteractiveComponentSystem) Serialize(component domain.Component) (map[string]interface{}, error) {
	interactive, ok := component.Data.(domain.InteractiveComponent)
	if !ok {
		return nil, fmt.Errorf("invalid interactive component")
	}

	return map[string]interface{}{
		"is_interactable":   interactive.IsInteractable,
		"input_handlers":    interactive.InputHandlers,
		"interaction_zones": interactive.InteractionZones,
	}, nil
}

func (ics *InteractiveComponentSystem) Deserialize(data map[string]interface{}) (domain.Component, error) {
	interactive := domain.InteractiveComponent{}

	if isInteractable, exists := data["is_interactable"]; exists {
		if interactableBool, ok := isInteractable.(bool); ok {
			interactive.IsInteractable = interactableBool
		}
	}

	return domain.Component{
		Type:    "interactive",
		Data:    interactive,
		Version: 1,
	}, nil
}

// NetworkComponentSystem handles network components
type NetworkComponentSystem struct{}

func (ncs *NetworkComponentSystem) GetType() domain.ComponentType {
	return "network"
}

func (ncs *NetworkComponentSystem) Create(data interface{}) (domain.Component, error) {
	networkData, ok := data.(domain.NetworkComponent)
	if !ok {
		return domain.Component{}, fmt.Errorf("invalid network component data")
	}

	return domain.Component{
		Type:    "network",
		Data:    networkData,
		Version: 1,
	}, nil
}

func (ncs *NetworkComponentSystem) Validate(component domain.Component) error {
	if component.Type != "network" {
		return fmt.Errorf("invalid component type for network system")
	}

	_, ok := component.Data.(domain.NetworkComponent)
	if !ok {
		return fmt.Errorf("invalid network component data")
	}

	return nil
}

func (ncs *NetworkComponentSystem) Update(existing domain.Component, updates map[string]interface{}) (domain.Component, error) {
	network, ok := existing.Data.(domain.NetworkComponent)
	if !ok {
		return domain.Component{}, fmt.Errorf("invalid existing network component")
	}

	// Update ownership
	if ownership, exists := updates["ownership"]; exists {
		if ownershipStr, ok := ownership.(string); ok {
			if playerID, err := uuid.Parse(ownershipStr); err == nil {
				network.Ownership = &playerID
			}
		}
	}

	return domain.Component{
		Type:    "network",
		Data:    network,
		Version: existing.Version + 1,
	}, nil
}

func (ncs *NetworkComponentSystem) Serialize(component domain.Component) (map[string]interface{}, error) {
	network, ok := component.Data.(domain.NetworkComponent)
	if !ok {
		return nil, fmt.Errorf("invalid network component")
	}

	ownership := ""
	if network.Ownership != nil {
		ownership = network.Ownership.String()
	}

	return map[string]interface{}{
		"ownership":         ownership,
		"replication_rules": network.ReplicationRules,
		"interest_area":     network.InterestArea,
	}, nil
}

func (ncs *NetworkComponentSystem) Deserialize(data map[string]interface{}) (domain.Component, error) {
	network := domain.NetworkComponent{}

	if ownership, exists := data["ownership"]; exists {
		if ownershipStr, ok := ownership.(string); ok {
			if playerID, err := uuid.Parse(ownershipStr); err == nil {
				network.Ownership = &playerID
			}
		}
	}

	return domain.Component{
		Type:    "network",
		Data:    network,
		Version: 1,
	}, nil
}
