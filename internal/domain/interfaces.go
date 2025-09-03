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

// GameMechanicManager handles game mechanics registration and management
type GameMechanicManager interface {
	RegisterMechanic(ctx context.Context, mechanic *GameMechanic) error
	UnregisterMechanic(ctx context.Context, mechanicID MechanicId) error
	GetMechanic(ctx context.Context, mechanicID MechanicId) (*GameMechanic, error)
	GetMechanicsByType(ctx context.Context, mechanicType MechanicType) ([]*GameMechanic, error)
	ValidateMechanic(ctx context.Context, mechanic *GameMechanic) error
	ListMechanics(ctx context.Context) ([]*GameMechanic, error)
}

// ActionDefinitionManager handles action definitions
type ActionDefinitionManager interface {
	RegisterAction(ctx context.Context, action *ActionDefinition) error
	UnregisterAction(ctx context.Context, actionID ActionTypeId) error
	GetAction(ctx context.Context, actionID ActionTypeId) (*ActionDefinition, error)
	GetActionsByType(ctx context.Context, actionType ActionType) ([]*ActionDefinition, error)
	ValidateAction(ctx context.Context, action *ActionDefinition) error
	ListActions(ctx context.Context) ([]*ActionDefinition, error)
}

// ResolutionManager handles action resolution methods
type ResolutionManager interface {
	RegisterResolutionMethod(ctx context.Context, method *ResolutionMethod) error
	UnregisterResolutionMethod(ctx context.Context, resolutionID ResolutionId) error
	GetResolutionMethod(ctx context.Context, resolutionID ResolutionId) (*ResolutionMethod, error)
	GetResolutionMethodsForAction(ctx context.Context, actionID ActionTypeId) ([]*ResolutionMethod, error)
	ValidateResolutionMethod(ctx context.Context, method *ResolutionMethod) error
	ListResolutionMethods(ctx context.Context) ([]*ResolutionMethod, error)
}

// ModifierManager handles modifier systems
type ModifierManager interface {
	RegisterModifier(ctx context.Context, modifier *ModifierSystem) error
	UnregisterModifier(ctx context.Context, modifierID string) error
	GetModifier(ctx context.Context, modifierID string) (*ModifierSystem, error)
	GetModifiersByType(ctx context.Context, modifierType ModifierType) ([]*ModifierSystem, error)
	ApplyModifiers(ctx context.Context, entityID EntityId, target string, context map[string]interface{}) ([]ModifierResult, error)
	ValidateModifier(ctx context.Context, modifier *ModifierSystem) error
	ListModifiers(ctx context.Context) ([]*ModifierSystem, error)
}

// ModifierResult represents the result of applying a modifier
type ModifierResult struct {
	ModifierID string      `json:"modifier_id"`
	Value      interface{} `json:"value"`
	Operation  string      `json:"operation"`
	Success    bool        `json:"success"`
	Error      string      `json:"error,omitempty"`
}

// RandomizationManager handles randomization rules
type RandomizationManager interface {
	RegisterRandomizationRule(ctx context.Context, rule *RandomizationRule) error
	UnregisterRandomizationRule(ctx context.Context, ruleID string) error
	GetRandomizationRule(ctx context.Context, ruleID string) (*RandomizationRule, error)
	GetRandomizationRulesByType(ctx context.Context, ruleType RandomizationType) ([]*RandomizationRule, error)
	ExecuteRandomization(ctx context.Context, ruleID string, context map[string]interface{}) (*RandomizationResult, error)
	ValidateRandomizationRule(ctx context.Context, rule *RandomizationRule) error
	ListRandomizationRules(ctx context.Context) ([]*RandomizationRule, error)
}

// RandomizationResult represents the result of a randomization
type RandomizationResult struct {
	RuleID   string                 `json:"rule_id"`
	Value    interface{}            `json:"value"`
	Details  string                 `json:"details"`
	Success  bool                   `json:"success"`
	Error    string                 `json:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TOML Parser Interfaces for Content Authoring (Phase 5.1)

// TOMLParser defines the interface for parsing TOML content
type TOMLParser interface {
	ParseTOML(data []byte, options TOMLParseOptions) (*TOMLParseResult, error)
	ParseTOMLFile(filePath string, options TOMLParseOptions) (*TOMLParseResult, error)
	ValidateContent(content *TOMLContent, options TOMLParseOptions) (*ContentValidation, error)
	ExtractExpressions(content *TOMLContent) (map[string]string, error)
	ExtractReferences(content *TOMLContent) ([]string, error)
}

// ContentRepositoryManager defines the interface for managing content repositories
type ContentRepositoryManager interface {
	AddContent(content *TOMLContent) error
	GetContent(id string) (*TOMLContent, error)
	GetContentByType(contentType ContentType) ([]*TOMLContent, error)
	RemoveContent(id string) error
	UpdateContent(content *TOMLContent) error
	ListContent() ([]*TOMLContent, error)
	ValidateRepository() (*ContentValidation, error)
	ResolveReferences(content *TOMLContent) error
	GetRepository() *ContentRepository
}

// SchemaValidator defines the interface for validating content against schemas
type SchemaValidator interface {
	ValidateAgainstSchema(content *TOMLContent, schemaPath string) (*ContentValidation, error)
	LoadSchema(schemaPath string) error
	GetSchema(contentType ContentType) (interface{}, error)
	ValidateField(field string, value interface{}, schema interface{}) error
}

// ExpressionExtractor defines the interface for extracting expressions from content
type ExpressionExtractor interface {
	ExtractExpressions(data map[string]interface{}) (map[string]string, error)
	ValidateExpression(expression string) error
	ParseExpression(expression string) (interface{}, error)
}

// ReferenceResolver defines the interface for resolving cross-references
type ReferenceResolver interface {
	ResolveReferences(content *TOMLContent, repository *ContentRepository) error
	ValidateReferences(references []string, repository *ContentRepository) (*ReferenceValidation, error)
	FindMissingReferences(content *TOMLContent, repository *ContentRepository) ([]string, error)
}
