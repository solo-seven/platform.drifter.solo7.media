package domain

import (
	"time"

	"github.com/google/uuid"
)

// Core ID types
type EntityId = uuid.UUID
type PlayerId = uuid.UUID
type RegionId = uuid.UUID
type RuleId = uuid.UUID
type ComponentType = string
type ActionType = string
type EventType = string

// Base Entity structure
type Entity struct {
	ID         EntityId                    `json:"id"`
	Components map[ComponentType]Component `json:"components"`
	Metadata   EntityMetadata              `json:"metadata"`
}

type Component struct {
	Type    ComponentType `json:"type"`
	Data    interface{}   `json:"data"`
	Version int           `json:"version"`
}

type EntityMetadata struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int       `json:"version"`
}

// Core Component Types
type TransformComponent struct {
	Position Vector3 `json:"position"`
	Rotation Vector3 `json:"rotation"`
	Scale    Vector3 `json:"scale"`
}

type Vector3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type PhysicsComponent struct {
	CollisionBounds Bounds  `json:"collision_bounds"`
	Mass            float64 `json:"mass"`
	Velocity        Vector3 `json:"velocity"`
	IsStatic        bool    `json:"is_static"`
}

type Bounds struct {
	Min Vector3 `json:"min"`
	Max Vector3 `json:"max"`
}

type RenderableComponent struct {
	AssetID           string                 `json:"asset_id"`
	RenderingHints    RenderingHints         `json:"rendering_hints"`
	AnimationSetID    *string                `json:"animation_set_id,omitempty"`
	MaterialOverrides map[string]interface{} `json:"material_overrides,omitempty"`
}

type RenderingHints struct {
	LODLevels      []LODLevel `json:"lod_levels"`
	CullDistance   float64    `json:"cull_distance"`
	ShadowCasting  bool       `json:"shadow_casting"`
	StaticBatching bool       `json:"static_batching"`
}

type LODLevel struct {
	Level        int     `json:"level"`
	Distance     float64 `json:"distance"`
	AssetID      string  `json:"asset_id"`
	PolygonCount int     `json:"polygon_count"`
}

type InteractiveComponent struct {
	InputHandlers    []InputHandler    `json:"input_handlers"`
	InteractionZones []InteractionZone `json:"interaction_zones"`
	IsInteractable   bool              `json:"is_interactable"`
}

type InputHandler struct {
	Type       string                 `json:"type"`
	HandlerID  string                 `json:"handler_id"`
	Properties map[string]interface{} `json:"properties"`
}

type InteractionZone struct {
	Bounds      Bounds                 `json:"bounds"`
	Interaction string                 `json:"interaction"`
	Properties  map[string]interface{} `json:"properties"`
}

type GameplayComponent struct {
	Stats         map[string]float64 `json:"stats"`
	Abilities     []Ability          `json:"abilities"`
	Inventory     Inventory          `json:"inventory"`
	StatusEffects []StatusEffect     `json:"status_effects"`
}

type Ability struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Cooldown    time.Duration          `json:"cooldown"`
	Properties  map[string]interface{} `json:"properties"`
}

type Inventory struct {
	Items     []Item  `json:"items"`
	MaxSlots  int     `json:"max_slots"`
	MaxWeight float64 `json:"max_weight"`
}

type Item struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Weight      float64                `json:"weight"`
	Properties  map[string]interface{} `json:"properties"`
	Quantity    int                    `json:"quantity"`
}

type StatusEffect struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Duration   time.Duration          `json:"duration"`
	StartTime  time.Time              `json:"start_time"`
	Properties map[string]interface{} `json:"properties"`
}

type NetworkComponent struct {
	ReplicationRules ReplicationRules `json:"replication_rules"`
	Ownership        *PlayerId        `json:"ownership,omitempty"`
	InterestArea     InterestArea     `json:"interest_area"`
}

type ReplicationRules struct {
	ReplicateToAll bool       `json:"replicate_to_all"`
	ReplicateTo    []PlayerId `json:"replicate_to"`
	ReplicateFrom  []PlayerId `json:"replicate_from"`
}

type InterestArea struct {
	Center   Vector3 `json:"center"`
	Radius   float64 `json:"radius"`
	Priority int     `json:"priority"`
}

// Game Rules and Actions
type GameRule struct {
	ID         RuleId         `json:"id"`
	Triggers   []EventTrigger `json:"triggers"`
	Conditions []Condition    `json:"conditions"`
	Actions    []Action       `json:"actions"`
	Priority   int            `json:"priority"`
}

type EventTrigger struct {
	Type      EventType              `json:"type"`
	Condition map[string]interface{} `json:"condition"`
}

type Condition struct {
	Type     string      `json:"type"`
	Property string      `json:"property"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

type Action struct {
	Type       ActionType             `json:"type"`
	Target     string                 `json:"target"`
	Properties map[string]interface{} `json:"properties"`
}

type ActionResult struct {
	WorldStateChanges   []StateChange        `json:"world_state_changes"`
	ClientNotifications []ClientNotification `json:"client_notifications"`
	AestheticEvents     []AestheticEventData `json:"aesthetic_events"`
}

type StateChange struct {
	EntityID  EntityId               `json:"entity_id"`
	Component ComponentType          `json:"component"`
	Changes   map[string]interface{} `json:"changes"`
	Timestamp time.Time              `json:"timestamp"`
}

type ClientNotification struct {
	PlayerID   PlayerId               `json:"player_id"`
	Type       string                 `json:"type"`
	Message    string                 `json:"message"`
	Properties map[string]interface{} `json:"properties"`
}

type AestheticEventData struct {
	Type       string                 `json:"type"`
	EntityID   EntityId               `json:"entity_id"`
	Properties map[string]interface{} `json:"properties"`
	Timestamp  time.Time              `json:"timestamp"`
}

// World State Management
type WorldState struct {
	Regions      map[RegionId]RegionState `json:"regions"`
	GlobalState  GlobalGameState          `json:"global_state"`
	PlayerStates map[PlayerId]PlayerState `json:"player_states"`
}

type RegionState struct {
	Entities        map[EntityId]Entity `json:"entities"`
	EnvironmentData EnvironmentData     `json:"environment_data"`
	ActiveRules     []RuleId            `json:"active_rules"`
	InterestAreas   []InterestArea      `json:"interest_areas"`
}

type EnvironmentData struct {
	Weather    string                 `json:"weather"`
	TimeOfDay  string                 `json:"time_of_day"`
	Properties map[string]interface{} `json:"properties"`
}

type GlobalGameState struct {
	GameTime   time.Time              `json:"game_time"`
	GamePhase  string                 `json:"game_phase"`
	Properties map[string]interface{} `json:"properties"`
}

type PlayerState struct {
	PlayerID      PlayerId               `json:"player_id"`
	CurrentRegion RegionId               `json:"current_region"`
	Position      Vector3                `json:"position"`
	Properties    map[string]interface{} `json:"properties"`
	LastSeen      time.Time              `json:"last_seen"`
}

// Network Protocol Types
type MessageType int

const (
	// Client → Server
	PlayerInput MessageType = iota
	ChatMessage
	AdminCommand

	// Server → Client
	StateUpdate
	AestheticEvent
	SystemNotification

	// Bidirectional
	Heartbeat
	ConnectionNegotiation
)

type NetworkMessage struct {
	Type      MessageType            `json:"type"`
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

type PlayerInputMessage struct {
	PlayerID  PlayerId               `json:"player_id"`
	InputType string                 `json:"input_type"`
	Data      map[string]interface{} `json:"data"`
}

type StateUpdateMessage struct {
	RegionID  RegionId      `json:"region_id"`
	Entities  []Entity      `json:"entities"`
	Changes   []StateChange `json:"changes"`
	Timestamp time.Time     `json:"timestamp"`
}
