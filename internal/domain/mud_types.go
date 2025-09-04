package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// MUD-Specific Domain Types following DDD principles

// Core ID types for MUD
type CharacterId = uuid.UUID
type LocationId = uuid.UUID
type ItemId = uuid.UUID
type NPCId = uuid.UUID
type QuestId = uuid.UUID

// Location represents a place in the game world
type Location interface {
	GetID() LocationId
	GetName() string
	GetDescription() string
	GetType() LocationType
	GetExits() []Exit
	GetItems() []ItemInstance
	GetNPCs() []NPCInstance
	GetPlayers() []CharacterId
}

// LocationType distinguishes between Areas and Structures
type LocationType string

const (
	LocationTypeArea      LocationType = "area"
	LocationTypeStructure LocationType = "structure"
)

// Area represents an outdoor location
type Area struct {
	ID          LocationId     `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Boundaries  []Boundary     `json:"boundaries"`
	Entrances   []Entrance     `json:"entrances"`
	Items       []ItemInstance `json:"items"`
	NPCs        []NPCInstance  `json:"npcs"`
	Players     []CharacterId  `json:"players"`
}

func (a *Area) GetID() LocationId      { return a.ID }
func (a *Area) GetName() string        { return a.Name }
func (a *Area) GetDescription() string { return a.Description }
func (a *Area) GetType() LocationType  { return LocationTypeArea }
func (a *Area) GetExits() []Exit {
	var exits []Exit
	for _, boundary := range a.Boundaries {
		exits = append(exits, Exit{
			Direction:      boundary.Direction,
			TargetLocation: boundary.TargetAreaId,
			Description:    boundary.Description,
		})
	}
	for _, entrance := range a.Entrances {
		exits = append(exits, Exit{
			Direction:      entrance.Name,
			TargetLocation: entrance.TargetStructureId,
			Description:    entrance.Description,
		})
	}
	return exits
}
func (a *Area) GetItems() []ItemInstance  { return a.Items }
func (a *Area) GetNPCs() []NPCInstance    { return a.NPCs }
func (a *Area) GetPlayers() []CharacterId { return a.Players }

// Structure represents an enclosed location
type Structure struct {
	ID          LocationId     `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Exits       []Exit         `json:"exits"`
	Items       []ItemInstance `json:"items"`
	NPCs        []NPCInstance  `json:"npcs"`
	Players     []CharacterId  `json:"players"`
}

func (s *Structure) GetID() LocationId         { return s.ID }
func (s *Structure) GetName() string           { return s.Name }
func (s *Structure) GetDescription() string    { return s.Description }
func (s *Structure) GetType() LocationType     { return LocationTypeStructure }
func (s *Structure) GetExits() []Exit          { return s.Exits }
func (s *Structure) GetItems() []ItemInstance  { return s.Items }
func (s *Structure) GetNPCs() []NPCInstance    { return s.NPCs }
func (s *Structure) GetPlayers() []CharacterId { return s.Players }

// Boundary represents a path between Areas
type Boundary struct {
	Direction    string     `json:"direction"`
	TargetAreaId LocationId `json:"target_area_id"`
	Description  string     `json:"description"`
}

// Entrance represents a way into a Structure
type Entrance struct {
	Name              string     `json:"name"`
	TargetStructureId LocationId `json:"target_structure_id"`
	Description       string     `json:"description"`
}

// Exit represents a way out of a Location
type Exit struct {
	Direction      string     `json:"direction"`
	TargetLocation LocationId `json:"target_location"`
	Description    string     `json:"description"`
}

// Character represents a player's avatar
type Character struct {
	ID            CharacterId       `json:"id"`
	Name          string            `json:"name"`
	Position      Position          `json:"position"`
	Stats         Stats             `json:"stats"`
	Inventory     MUDInventory      `json:"inventory"`
	Abilities     []MUDAbility      `json:"abilities"`
	StatusEffects []MUDStatusEffect `json:"status_effects"`
	Metadata      CharacterMetadata `json:"metadata"`
}

// Position represents where a character is located
type Position struct {
	LocationID LocationId `json:"location_id"`
}

// Stats represents character attributes
type Stats struct {
	Health     int32            `json:"health"`
	MaxHealth  int32            `json:"max_health"`
	Mana       int32            `json:"mana"`
	MaxMana    int32            `json:"max_mana"`
	Attributes map[string]int32 `json:"attributes"` // strength, dexterity, etc.
}

// MUDInventory represents a character's possessions
type MUDInventory struct {
	Items    []ItemInstance `json:"items"`
	Capacity int32          `json:"capacity"`
}

// ItemInstance represents a specific item in the world
type ItemInstance struct {
	ID          ItemId                 `json:"id"`
	TemplateID  string                 `json:"template_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Weight      int32                  `json:"weight"`
	Properties  map[string]interface{} `json:"properties"`
	Quantity    int32                  `json:"quantity"`
}

// NPCInstance represents a specific NPC in the world
type NPCInstance struct {
	ID          NPCId        `json:"id"`
	TemplateID  string       `json:"template_id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Stats       Stats        `json:"stats"`
	Inventory   MUDInventory `json:"inventory"`
	Abilities   []MUDAbility `json:"abilities"`
	AI          AIBehavior   `json:"ai"`
}

// MUDAbility represents a character or NPC ability
type MUDAbility struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Cooldown    time.Duration          `json:"cooldown"`
	Properties  map[string]interface{} `json:"properties"`
}

// MUDStatusEffect represents a temporary effect on a character
type MUDStatusEffect struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Duration   time.Duration          `json:"duration"`
	StartTime  time.Time              `json:"start_time"`
	Properties map[string]interface{} `json:"properties"`
}

// AIBehavior represents NPC AI logic
type AIBehavior struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
}

// CharacterMetadata contains character creation and update info
type CharacterMetadata struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int32     `json:"version"`
}

// World represents the entire game world
type World struct {
	ID         string                     `json:"id"`
	Name       string                     `json:"name"`
	Locations  map[LocationId]Location    `json:"locations"`
	Characters map[CharacterId]*Character `json:"characters"`
	NPCs       map[NPCId]*NPCInstance     `json:"npcs"`
	Items      map[ItemId]*ItemInstance   `json:"items"`
	Quests     map[QuestId]*Quest         `json:"quests"`
	Metadata   WorldMetadata              `json:"metadata"`
}

// Quest represents a quest or mission
type Quest struct {
	ID          QuestId                `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Status      QuestStatus            `json:"status"`
	Objectives  []QuestObjective       `json:"objectives"`
	Rewards     []QuestReward          `json:"rewards"`
	Properties  map[string]interface{} `json:"properties"`
}

// QuestStatus represents the state of a quest
type QuestStatus string

const (
	QuestStatusAvailable QuestStatus = "available"
	QuestStatusActive    QuestStatus = "active"
	QuestStatusCompleted QuestStatus = "completed"
	QuestStatusFailed    QuestStatus = "failed"
)

// QuestObjective represents a quest goal
type QuestObjective struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Target      string                 `json:"target"`
	Quantity    int32                  `json:"quantity"`
	Completed   bool                   `json:"completed"`
	Properties  map[string]interface{} `json:"properties"`
}

// QuestReward represents a quest reward
type QuestReward struct {
	Type       string                 `json:"type"`
	ItemID     string                 `json:"item_id,omitempty"`
	Experience int32                  `json:"experience,omitempty"`
	Gold       int32                  `json:"gold,omitempty"`
	Properties map[string]interface{} `json:"properties"`
}

// WorldMetadata contains world creation and update info
type WorldMetadata struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int32     `json:"version"`
}

// MUDAction represents a player action
type MUDAction interface {
	GetType() string
	GetCharacterID() CharacterId
	Execute(ctx context.Context, world *World) (*MUDActionResult, error)
}

// MUDActionResult represents the result of an action
type MUDActionResult struct {
	Success       bool                    `json:"success"`
	Message       string                  `json:"message"`
	StateChanges  []MUDStateChange        `json:"state_changes"`
	Events        []MUDGameEvent          `json:"events"`
	Notifications []MUDClientNotification `json:"notifications"`
}

// MUDStateChange represents a change to the game state
type MUDStateChange struct {
	EntityID  string                 `json:"entity_id"`
	Component string                 `json:"component"`
	Changes   map[string]interface{} `json:"changes"`
	Timestamp time.Time              `json:"timestamp"`
}

// MUDGameEvent represents an event in the game
type MUDGameEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// MUDClientNotification represents a notification to a client
type MUDClientNotification struct {
	PlayerID   CharacterId            `json:"player_id"`
	Type       string                 `json:"type"`
	Message    string                 `json:"message"`
	Properties map[string]interface{} `json:"properties"`
}
