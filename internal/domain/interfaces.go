package domain

import (
	"context"
	"time"
)

// Core Domain Interfaces

// EntityManager handles entity lifecycle and component management
type EntityManager interface {
	CreateEntity(ctx context.Context, components map[ComponentType]Component) (*Entity, error)
	GetEntity(ctx context.Context, id EntityId) (*Entity, error)
	UpdateEntity(ctx context.Context, id EntityId, updates map[ComponentType]Component) error
	DeleteEntity(ctx context.Context, id EntityId) error
	GetEntitiesInRegion(ctx context.Context, regionID RegionId) ([]*Entity, error)
	GetEntitiesInArea(ctx context.Context, regionID RegionId, center Vector3, radius float64) ([]*Entity, error)
}

// RulesEngine processes game rules and actions
type RulesEngine interface {
	RegisterRule(ctx context.Context, rule *GameRule) error
	UnregisterRule(ctx context.Context, ruleID RuleId) error
	ProcessEvent(ctx context.Context, event EventType, data map[string]interface{}) (*ActionResult, error)
	GetActiveRules(ctx context.Context, regionID RegionId) ([]*GameRule, error)
}

// WorldStateManager manages the authoritative game state
type WorldStateManager interface {
	GetWorldState(ctx context.Context) (*WorldState, error)
	GetRegionState(ctx context.Context, regionID RegionId) (*RegionState, error)
	UpdateRegionState(ctx context.Context, regionID RegionId, updates []StateChange) error
	GetPlayerState(ctx context.Context, playerID PlayerId) (*PlayerState, error)
	UpdatePlayerState(ctx context.Context, playerID PlayerId, updates map[string]interface{}) error
}

// NetworkProtocol defines the wire protocol for client-server communication
type NetworkProtocol interface {
	SendMessage(ctx context.Context, message *NetworkMessage) error
	ReceiveMessage(ctx context.Context) (*NetworkMessage, error)
	BroadcastToRegion(ctx context.Context, regionID RegionId, message *NetworkMessage) error
	BroadcastToPlayer(ctx context.Context, playerID PlayerId, message *NetworkMessage) error
	BroadcastToArea(ctx context.Context, regionID RegionId, center Vector3, radius float64, message *NetworkMessage) error
}

// ClientConnection represents a connected client
type ClientConnection interface {
	GetPlayerID() PlayerId
	GetConnectionID() string
	Send(ctx context.Context, message *NetworkMessage) error
	Receive(ctx context.Context) (*NetworkMessage, error)
	Close() error
	IsConnected() bool
	GetLastHeartbeat() time.Time
	UpdateHeartbeat()
}

// ConnectionManager handles client connections
type ConnectionManager interface {
	AddConnection(ctx context.Context, conn ClientConnection) error
	RemoveConnection(ctx context.Context, connectionID string) error
	GetConnection(ctx context.Context, connectionID string) (ClientConnection, error)
	GetPlayerConnection(ctx context.Context, playerID PlayerId) (ClientConnection, error)
	GetConnectionsInRegion(ctx context.Context, regionID RegionId) ([]ClientConnection, error)
	BroadcastToAll(ctx context.Context, message *NetworkMessage) error
	BroadcastToRegion(ctx context.Context, regionID RegionId, message *NetworkMessage) error
	BroadcastToPlayer(ctx context.Context, playerID PlayerId, message *NetworkMessage) error
	GetConnectionCount() int
	GetPlayerCount() int
	CleanupInactiveConnections(ctx context.Context, maxInactiveTime time.Duration) error
}

// GameServer represents the main game server
type GameServer interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool
	GetEntityManager() EntityManager
	GetRulesEngine() RulesEngine
	GetWorldStateManager() WorldStateManager
	GetConnectionManager() ConnectionManager
	GetNetworkProtocol() NetworkProtocol
}

// ContentValidator validates game content and rules
type ContentValidator interface {
	ValidateEntity(ctx context.Context, entity *Entity) error
	ValidateRule(ctx context.Context, rule *GameRule) error
	ValidateComponent(ctx context.Context, component Component) error
	ValidateWorldState(ctx context.Context, state *WorldState) error
}

// PersistenceLayer handles data persistence
type PersistenceLayer interface {
	SaveEntity(ctx context.Context, entity *Entity) error
	LoadEntity(ctx context.Context, id EntityId) (*Entity, error)
	SaveWorldState(ctx context.Context, state *WorldState) error
	LoadWorldState(ctx context.Context) (*WorldState, error)
	SavePlayerState(ctx context.Context, state *PlayerState) error
	LoadPlayerState(ctx context.Context, playerID PlayerId) (*PlayerState, error)
}

// EventBus handles event distribution
type EventBus interface {
	Publish(ctx context.Context, event EventType, data map[string]interface{}) error
	Subscribe(ctx context.Context, event EventType, handler EventHandler) error
	Unsubscribe(ctx context.Context, event EventType, handler EventHandler) error
}

// EventHandler processes events
type EventHandler interface {
	Handle(ctx context.Context, event EventType, data map[string]interface{}) error
}

// InterestManager handles interest management for network optimization
type InterestManager interface {
	AddInterestArea(ctx context.Context, playerID PlayerId, area InterestArea) error
	RemoveInterestArea(ctx context.Context, playerID PlayerId, areaID string) error
	GetEntitiesOfInterest(ctx context.Context, playerID PlayerId) ([]*Entity, error)
	UpdatePlayerPosition(ctx context.Context, playerID PlayerId, position Vector3) error
}

// GameSession represents a player's game session
type GameSession interface {
	GetPlayerID() PlayerId
	GetSessionID() string
	GetStartTime() time.Time
	GetLastActivity() time.Time
	UpdateActivity()
	IsActive() bool
	GetCurrentRegion() RegionId
	SetCurrentRegion(regionID RegionId)
}

// SessionManager handles player sessions
type SessionManager interface {
	CreateSession(ctx context.Context, playerID PlayerId) (GameSession, error)
	GetSession(ctx context.Context, playerID PlayerId) (GameSession, error)
	EndSession(ctx context.Context, playerID PlayerId) error
	GetActiveSessions(ctx context.Context) ([]GameSession, error)
	CleanupInactiveSessions(ctx context.Context, maxInactiveTime time.Duration) error
}

// MetricsCollector collects performance and game metrics
type MetricsCollector interface {
	RecordEntityCount(count int)
	RecordPlayerCount(count int)
	RecordMessageLatency(messageType MessageType, latency time.Duration)
	RecordRuleExecutionTime(ruleID RuleId, duration time.Duration)
	RecordRegionLoad(regionID RegionId, load float64)
	GetMetrics() map[string]interface{}
}

// Logger interface for structured logging
type Logger interface {
	Debug(msg string, fields map[string]interface{})
	Info(msg string, fields map[string]interface{})
	Warn(msg string, fields map[string]interface{})
	Error(msg string, fields map[string]interface{})
	Fatal(msg string, fields map[string]interface{})
}

// Configuration interface for server configuration
type Configuration interface {
	GetServerPort() int
	GetMaxConnections() int
	GetHeartbeatInterval() time.Duration
	GetRegionSize() float64
	GetMaxEntitiesPerRegion() int
	GetLogLevel() string
	GetDatabaseURL() string
	GetRedisURL() string
}
