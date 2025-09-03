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

// Game Mechanics Types (Phase 2.1: Core Mechanics Repository)

type MechanicId = string
type ActionTypeId = string
type OutcomeId = string
type ResolutionId = string

// GameMechanic represents a core game mechanic
type GameMechanic struct {
	ID          MechanicId             `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        MechanicType           `json:"type"`
	Properties  map[string]interface{} `json:"properties"`
	Validation  MechanicValidation     `json:"validation"`
	Version     int                    `json:"version"`
}

type MechanicType string

const (
	MechanicTypeCombat      MechanicType = "combat"
	MechanicTypeMovement    MechanicType = "movement"
	MechanicTypeInteraction MechanicType = "interaction"
	MechanicTypeSkill       MechanicType = "skill"
	MechanicTypeMagic       MechanicType = "magic"
	MechanicTypeSocial      MechanicType = "social"
)

type MechanicValidation struct {
	RequiredInputs    []InputRequirement `json:"required_inputs"`
	OutputValidation  OutputValidation   `json:"output_validation"`
	ContextValidation ContextValidation  `json:"context_validation"`
}

type InputRequirement struct {
	Name        string       `json:"name"`
	Type        string       `json:"type"`
	Required    bool         `json:"required"`
	Default     interface{}  `json:"default,omitempty"`
	Constraints []Constraint `json:"constraints,omitempty"`
}

type Constraint struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

type OutputValidation struct {
	ExpectedTypes  []string `json:"expected_types"`
	RequiredFields []string `json:"required_fields"`
}

type ContextValidation struct {
	RequiredEntities []string `json:"required_entities"`
	RequiredState    []string `json:"required_state"`
}

// ActionDefinition represents a game action
type ActionDefinition struct {
	ID            ActionTypeId           `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Type          ActionType             `json:"type"`
	Cost          ActionCost             `json:"cost"`
	Prerequisites []Prerequisite         `json:"prerequisites"`
	Resolution    ResolutionMethod       `json:"resolution"`
	Properties    map[string]interface{} `json:"properties"`
	Validation    ActionValidation       `json:"validation"`
}

// Action type constants
const (
	ActionTypeAttack   ActionType = "attack"
	ActionTypeMove     ActionType = "move"
	ActionTypeCast     ActionType = "cast"
	ActionTypeInteract ActionType = "interact"
	ActionTypeSkill    ActionType = "skill"
	ActionTypeSocial   ActionType = "social"
	ActionTypeDefend   ActionType = "defend"
	ActionTypeUse      ActionType = "use"
)

type ActionCost struct {
	ActionPoints   int                    `json:"action_points"`
	MovementPoints int                    `json:"movement_points"`
	Resources      map[string]interface{} `json:"resources"`
	Cooldown       time.Duration          `json:"cooldown"`
	Uses           *UsesLimit             `json:"uses,omitempty"`
}

type UsesLimit struct {
	Per     string `json:"per"` // "turn", "encounter", "short_rest", "long_rest", "day"
	Count   int    `json:"count"`
	Current int    `json:"current"`
}

type Prerequisite struct {
	Type       string                 `json:"type"`
	Condition  string                 `json:"condition"`
	Properties map[string]interface{} `json:"properties"`
}

type ActionValidation struct {
	RequiredComponents []ComponentType `json:"required_components"`
	RequiredStats      []string        `json:"required_stats"`
	ContextChecks      []ContextCheck  `json:"context_checks"`
}

type ContextCheck struct {
	Type       string                 `json:"type"`
	Expression string                 `json:"expression"`
	Properties map[string]interface{} `json:"properties"`
}

// ResolutionMethod defines how an action is resolved
type ResolutionMethod struct {
	ID                ResolutionId           `json:"id"`
	Name              string                 `json:"name"`
	ApplicableActions []ActionTypeId         `json:"applicable_actions"`
	RequiredInputs    []InputRequirement     `json:"required_inputs"`
	CalculationSteps  []CalculationStep      `json:"calculation_steps"`
	PossibleOutcomes  []OutcomeRange         `json:"possible_outcomes"`
	Properties        map[string]interface{} `json:"properties"`
}

type CalculationStep struct {
	Step        int                    `json:"step"`
	Type        string                 `json:"type"`
	Expression  string                 `json:"expression"`
	Description string                 `json:"description"`
	Properties  map[string]interface{} `json:"properties"`
}

type OutcomeRange struct {
	Type        string                 `json:"type"`
	MinValue    interface{}            `json:"min_value"`
	MaxValue    interface{}            `json:"max_value"`
	Description string                 `json:"description"`
	Properties  map[string]interface{} `json:"properties"`
}

// OutcomeFormula defines how outcomes are calculated
type OutcomeFormula struct {
	ID          OutcomeId              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Formula     string                 `json:"formula"`
	Variables   []FormulaVariable      `json:"variables"`
	Validation  FormulaValidation      `json:"validation"`
	Properties  map[string]interface{} `json:"properties"`
}

type FormulaVariable struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Source      string      `json:"source"`
	Default     interface{} `json:"default,omitempty"`
	Description string      `json:"description"`
}

type FormulaValidation struct {
	RequiredVariables []string     `json:"required_variables"`
	ExpectedOutput    string       `json:"expected_output"`
	RangeChecks       []RangeCheck `json:"range_checks"`
}

type RangeCheck struct {
	Variable string      `json:"variable"`
	Min      interface{} `json:"min,omitempty"`
	Max      interface{} `json:"max,omitempty"`
}

// ModifierSystem defines how modifiers are applied
type ModifierSystem struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        ModifierType           `json:"type"`
	Source      ModifierSource         `json:"source"`
	Application ModifierApplication    `json:"application"`
	Properties  map[string]interface{} `json:"properties"`
}

type ModifierType string

const (
	ModifierTypeAttribute     ModifierType = "attribute"
	ModifierTypeEquipment     ModifierType = "equipment"
	ModifierTypeEnvironmental ModifierType = "environmental"
	ModifierTypeTemporary     ModifierType = "temporary"
	ModifierTypeSkill         ModifierType = "skill"
	ModifierTypeMagic         ModifierType = "magic"
)

type ModifierSource struct {
	Type       string                 `json:"type"`
	EntityID   *EntityId              `json:"entity_id,omitempty"`
	Component  *ComponentType         `json:"component,omitempty"`
	Property   string                 `json:"property"`
	Properties map[string]interface{} `json:"properties"`
}

type ModifierApplication struct {
	Target     string                 `json:"target"`
	Operation  string                 `json:"operation"` // "add", "multiply", "set", "conditional"
	Expression string                 `json:"expression"`
	Conditions []string               `json:"conditions"`
	Properties map[string]interface{} `json:"properties"`
}

// RandomizationRule defines dice rolling and random mechanics
type RandomizationRule struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       RandomizationType      `json:"type"`
	Notation   string                 `json:"notation"`
	Expression string                 `json:"expression"`
	Properties map[string]interface{} `json:"properties"`
}

type RandomizationType string

const (
	RandomizationTypeDice     RandomizationType = "dice"
	RandomizationTypeTable    RandomizationType = "table"
	RandomizationTypeWeighted RandomizationType = "weighted"
	RandomizationTypeCustom   RandomizationType = "custom"
)

// DeterministicRule provides fallback for randomization
type DeterministicRule struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Condition  string                 `json:"condition"`
	Fallback   string                 `json:"fallback"`
	Properties map[string]interface{} `json:"properties"`
}

// TOML Parser Types for Content Authoring (Phase 5.1)

// ContentType represents the type of content being parsed
type ContentType string

const (
	ContentTypeClass         ContentType = "class"
	ContentTypeSpecies       ContentType = "species"
	ContentTypeItem          ContentType = "item"
	ContentTypeSpell         ContentType = "spell"
	ContentTypeAbility       ContentType = "ability"
	ContentTypeMonster       ContentType = "monster"
	ContentTypeEncounter     ContentType = "encounter"
	ContentTypeLocation      ContentType = "location"
	ContentTypeMechanic      ContentType = "mechanic"
	ContentTypeAction        ContentType = "action"
	ContentTypeResolution    ContentType = "resolution"
	ContentTypeModifier      ContentType = "modifier"
	ContentTypeRandomization ContentType = "randomization"
)

// TOMLContent represents parsed TOML content
type TOMLContent struct {
	Type        ContentType            `json:"type"`
	ID          string                 `json:"id"`
	Version     string                 `json:"version,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Data        map[string]interface{} `json:"data"`
	Expressions map[string]string      `json:"expressions,omitempty"`
	References  []string               `json:"references,omitempty"`
	Validation  ContentValidation      `json:"validation,omitempty"`
}

// ContentValidation represents validation results for parsed content
type ContentValidation struct {
	IsValid     bool                `json:"is_valid"`
	Errors      []ValidationError   `json:"errors,omitempty"`
	Warnings    []ValidationWarning `json:"warnings,omitempty"`
	SchemaValid bool                `json:"schema_valid"`
	References  ReferenceValidation `json:"references,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ReferenceValidation represents validation of cross-references
type ReferenceValidation struct {
	ValidReferences   []string `json:"valid_references"`
	InvalidReferences []string `json:"invalid_references"`
	MissingReferences []string `json:"missing_references"`
}

// TOMLParseResult represents the result of parsing TOML content
type TOMLParseResult struct {
	Content  *TOMLContent           `json:"content,omitempty"`
	Success  bool                   `json:"success"`
	Error    string                 `json:"error,omitempty"`
	Warnings []string               `json:"warnings,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TOMLParseOptions represents options for TOML parsing
type TOMLParseOptions struct {
	ValidateSchema    bool          `json:"validate_schema"`
	ResolveReferences bool          `json:"resolve_references"`
	AllowedTypes      []ContentType `json:"allowed_types,omitempty"`
	StrictMode        bool          `json:"strict_mode"`
	SchemaPath        string        `json:"schema_path,omitempty"`
}

// ContentRepository represents a repository of parsed content
type ContentRepository struct {
	Content     map[string]*TOMLContent  `json:"content"`
	ByType      map[ContentType][]string `json:"by_type"`
	References  map[string][]string      `json:"references"`
	LastUpdated time.Time                `json:"last_updated"`
	Version     string                   `json:"version"`
}
