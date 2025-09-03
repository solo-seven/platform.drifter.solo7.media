package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/solo7.media/platform.drifter.solo7.media/internal/domain"
)

// WorldStateManagerImpl implements the WorldStateManager interface
type WorldStateManagerImpl struct {
	worldState *domain.WorldState
	mu         sync.RWMutex
	logger     domain.Logger
}

// NewWorldStateManager creates a new world state manager
func NewWorldStateManager(logger domain.Logger) *WorldStateManagerImpl {
	return &WorldStateManagerImpl{
		worldState: &domain.WorldState{
			Regions: make(map[domain.RegionId]domain.RegionState),
			GlobalState: domain.GlobalGameState{
				GameTime:   time.Now(),
				GamePhase:  "active",
				Properties: make(map[string]interface{}),
			},
			PlayerStates: make(map[domain.PlayerId]domain.PlayerState),
		},
		logger: logger,
	}
}

// GetWorldState returns the current world state
func (wsm *WorldStateManagerImpl) GetWorldState(ctx context.Context) (*domain.WorldState, error) {
	wsm.mu.RLock()
	defer wsm.mu.RUnlock()

	// Return a copy to avoid race conditions
	// TODO: Implement proper deep copying
	return wsm.worldState, nil
}

// GetRegionState returns the state of a specific region
func (wsm *WorldStateManagerImpl) GetRegionState(ctx context.Context, regionID domain.RegionId) (*domain.RegionState, error) {
	wsm.mu.RLock()
	defer wsm.mu.RUnlock()

	regionState, exists := wsm.worldState.Regions[regionID]
	if !exists {
		return nil, fmt.Errorf("region %s not found", regionID)
	}

	return &regionState, nil
}

// UpdateRegionState updates the state of a region
func (wsm *WorldStateManagerImpl) UpdateRegionState(ctx context.Context, regionID domain.RegionId, updates []domain.StateChange) error {
	wsm.mu.Lock()
	defer wsm.mu.Unlock()

	regionState, exists := wsm.worldState.Regions[regionID]
	if !exists {
		// Create new region state
		regionState = domain.RegionState{
			Entities: make(map[domain.EntityId]domain.Entity),
			EnvironmentData: domain.EnvironmentData{
				Weather:    "clear",
				TimeOfDay:  "day",
				Properties: make(map[string]interface{}),
			},
			ActiveRules:   []domain.RuleId{},
			InterestAreas: []domain.InterestArea{},
		}
		wsm.worldState.Regions[regionID] = regionState
	}

	// Apply updates
	for _, update := range updates {
		// Update entity in region
		if entity, exists := regionState.Entities[update.EntityID]; exists {
			// Update component
			if component, exists := entity.Components[update.Component]; exists {
				// Merge changes
				for key, value := range update.Changes {
					// TODO: Implement proper component data merging
					// This is a simplified implementation
					if component.Data == nil {
						component.Data = make(map[string]interface{})
					}
					if dataMap, ok := component.Data.(map[string]interface{}); ok {
						dataMap[key] = value
					}
				}
				component.Version++
				entity.Components[update.Component] = component
			}

			// Update entity metadata
			entity.Metadata.UpdatedAt = update.Timestamp
			entity.Metadata.Version++

			regionState.Entities[update.EntityID] = entity
		}
	}

	wsm.worldState.Regions[regionID] = regionState

	wsm.logger.Debug("Region state updated", map[string]interface{}{
		"region_id": regionID,
		"updates":   len(updates),
	})

	return nil
}

// GetPlayerState returns the state of a specific player
func (wsm *WorldStateManagerImpl) GetPlayerState(ctx context.Context, playerID domain.PlayerId) (*domain.PlayerState, error) {
	wsm.mu.RLock()
	defer wsm.mu.RUnlock()

	playerState, exists := wsm.worldState.PlayerStates[playerID]
	if !exists {
		return nil, fmt.Errorf("player %s not found", playerID)
	}

	return &playerState, nil
}

// UpdatePlayerState updates the state of a player
func (wsm *WorldStateManagerImpl) UpdatePlayerState(ctx context.Context, playerID domain.PlayerId, updates map[string]interface{}) error {
	wsm.mu.Lock()
	defer wsm.mu.Unlock()

	playerState, exists := wsm.worldState.PlayerStates[playerID]
	if !exists {
		// Create new player state
		playerState = domain.PlayerState{
			PlayerID:      playerID,
			CurrentRegion: uuid.Nil, // Default region
			Position:      domain.Vector3{X: 0, Y: 0, Z: 0},
			Properties:    make(map[string]interface{}),
			LastSeen:      time.Now(),
		}
	}

	// Apply updates
	for key, value := range updates {
		switch key {
		case "current_region":
			if regionID, ok := value.(domain.RegionId); ok {
				playerState.CurrentRegion = regionID
			}
		case "position":
			if position, ok := value.(domain.Vector3); ok {
				playerState.Position = position
			}
		case "last_seen":
			if lastSeen, ok := value.(time.Time); ok {
				playerState.LastSeen = lastSeen
			}
		default:
			// Store in properties
			playerState.Properties[key] = value
		}
	}

	// Update last seen timestamp
	playerState.LastSeen = time.Now()

	wsm.worldState.PlayerStates[playerID] = playerState

	wsm.logger.Debug("Player state updated", map[string]interface{}{
		"player_id": playerID,
		"updates":   len(updates),
	})

	return nil
}

// CreateRegion creates a new region
func (wsm *WorldStateManagerImpl) CreateRegion(ctx context.Context, regionID domain.RegionId) error {
	wsm.mu.Lock()
	defer wsm.mu.Unlock()

	// Check if region already exists
	if _, exists := wsm.worldState.Regions[regionID]; exists {
		return fmt.Errorf("region %s already exists", regionID)
	}

	// Create new region
	regionState := domain.RegionState{
		Entities: make(map[domain.EntityId]domain.Entity),
		EnvironmentData: domain.EnvironmentData{
			Weather:    "clear",
			TimeOfDay:  "day",
			Properties: make(map[string]interface{}),
		},
		ActiveRules:   []domain.RuleId{},
		InterestAreas: []domain.InterestArea{},
	}

	wsm.worldState.Regions[regionID] = regionState

	wsm.logger.Info("Region created", map[string]interface{}{
		"region_id": regionID,
	})

	return nil
}

// DeleteRegion removes a region
func (wsm *WorldStateManagerImpl) DeleteRegion(ctx context.Context, regionID domain.RegionId) error {
	wsm.mu.Lock()
	defer wsm.mu.Unlock()

	// Check if region exists
	if _, exists := wsm.worldState.Regions[regionID]; !exists {
		return fmt.Errorf("region %s not found", regionID)
	}

	// Remove region
	delete(wsm.worldState.Regions, regionID)

	wsm.logger.Info("Region deleted", map[string]interface{}{
		"region_id": regionID,
	})

	return nil
}

// GetRegions returns all region IDs
func (wsm *WorldStateManagerImpl) GetRegions(ctx context.Context) ([]domain.RegionId, error) {
	wsm.mu.RLock()
	defer wsm.mu.RUnlock()

	var regionIDs []domain.RegionId
	for regionID := range wsm.worldState.Regions {
		regionIDs = append(regionIDs, regionID)
	}

	return regionIDs, nil
}

// GetPlayers returns all player IDs
func (wsm *WorldStateManagerImpl) GetPlayers(ctx context.Context) ([]domain.PlayerId, error) {
	wsm.mu.RLock()
	defer wsm.mu.RUnlock()

	var playerIDs []domain.PlayerId
	for playerID := range wsm.worldState.PlayerStates {
		playerIDs = append(playerIDs, playerID)
	}

	return playerIDs, nil
}

// UpdateGlobalState updates the global game state
func (wsm *WorldStateManagerImpl) UpdateGlobalState(ctx context.Context, updates map[string]interface{}) error {
	wsm.mu.Lock()
	defer wsm.mu.Unlock()

	// Apply updates to global state
	for key, value := range updates {
		switch key {
		case "game_time":
			if gameTime, ok := value.(time.Time); ok {
				wsm.worldState.GlobalState.GameTime = gameTime
			}
		case "game_phase":
			if gamePhase, ok := value.(string); ok {
				wsm.worldState.GlobalState.GamePhase = gamePhase
			}
		default:
			// Store in properties
			wsm.worldState.GlobalState.Properties[key] = value
		}
	}

	wsm.logger.Debug("Global state updated", map[string]interface{}{
		"updates": len(updates),
	})

	return nil
}
