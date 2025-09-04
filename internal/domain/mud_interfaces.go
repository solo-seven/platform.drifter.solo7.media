package domain

import (
	"context"
)

// MUD-Specific Domain Interfaces following DDD principles

// CharacterRepository handles character persistence
type CharacterRepository interface {
	Create(ctx context.Context, character *Character) error
	GetByID(ctx context.Context, id CharacterId) (*Character, error)
	GetByName(ctx context.Context, name string) (*Character, error)
	Update(ctx context.Context, character *Character) error
	Delete(ctx context.Context, id CharacterId) error
	ListByLocation(ctx context.Context, locationID LocationId) ([]*Character, error)
}

// WorldRepository handles world content loading
type WorldRepository interface {
	GetLocation(ctx context.Context, id LocationId) (Location, error)
	GetItemTemplate(ctx context.Context, templateID string) (*ItemTemplate, error)
	GetNPCTemplate(ctx context.Context, templateID string) (*NPCTemplate, error)
	GetQuest(ctx context.Context, id QuestId) (*Quest, error)
	LoadWorld(ctx context.Context) (*World, error)
}

// ItemTemplate represents a template for creating item instances
type ItemTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Weight      int32                  `json:"weight"`
	Properties  map[string]interface{} `json:"properties"`
}

// NPCTemplate represents a template for creating NPC instances
type NPCTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Stats       Stats                  `json:"stats"`
	Abilities   []Ability              `json:"abilities"`
	AI          AIBehavior             `json:"ai"`
	Properties  map[string]interface{} `json:"properties"`
}

// ActionService processes player actions
type ActionService interface {
	ProcessAction(ctx context.Context, action Action, world *World) (*ActionResult, error)
	ValidateAction(ctx context.Context, action Action, world *World) error
}

// GameLoopService manages the game loop
type GameLoopService interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool
	GetTickRate() int
	SetTickRate(rate int)
}

// LocationService manages location-related operations
type LocationService interface {
	MoveCharacter(ctx context.Context, characterID CharacterId, direction string, world *World) (*ActionResult, error)
	GetLocationDescription(ctx context.Context, locationID LocationId, world *World) (string, error)
	GetExits(ctx context.Context, locationID LocationId, world *World) ([]Exit, error)
	GetItemsInLocation(ctx context.Context, locationID LocationId, world *World) ([]ItemInstance, error)
	GetNPCsInLocation(ctx context.Context, locationID LocationId, world *World) ([]NPCInstance, error)
}

// InventoryService manages inventory operations
type InventoryService interface {
	AddItem(ctx context.Context, characterID CharacterId, item ItemInstance, world *World) (*ActionResult, error)
	RemoveItem(ctx context.Context, characterID CharacterId, itemID ItemId, world *World) (*ActionResult, error)
	GetInventory(ctx context.Context, characterID CharacterId, world *World) (Inventory, error)
	CanCarry(ctx context.Context, characterID CharacterId, item ItemInstance, world *World) bool
}

// CombatService manages combat operations
type CombatService interface {
	Attack(ctx context.Context, attackerID CharacterId, targetID string, world *World) (*ActionResult, error)
	CalculateDamage(ctx context.Context, attacker *Character, target *Character, weapon *ItemInstance) int32
	IsInCombat(ctx context.Context, characterID CharacterId, world *World) bool
	EndCombat(ctx context.Context, characterID CharacterId, world *World) error
}

// QuestService manages quest operations
type QuestService interface {
	StartQuest(ctx context.Context, characterID CharacterId, questID QuestId, world *World) (*ActionResult, error)
	CompleteQuest(ctx context.Context, characterID CharacterId, questID QuestId, world *World) (*ActionResult, error)
	UpdateQuestProgress(ctx context.Context, characterID CharacterId, questID QuestId, objectiveID string, world *World) (*ActionResult, error)
	GetAvailableQuests(ctx context.Context, characterID CharacterId, world *World) ([]*Quest, error)
	GetActiveQuests(ctx context.Context, characterID CharacterId, world *World) ([]*Quest, error)
}

// LLMService handles LLM interactions
type LLMService interface {
	GenerateText(ctx context.Context, prompt string) (string, error)
	ParseIntent(ctx context.Context, input string, context map[string]interface{}) (*ParsedIntent, error)
	GenerateDialogue(ctx context.Context, npc *NPCInstance, player *Character, topic string) (string, error)
	GenerateDescription(ctx context.Context, location Location, context map[string]interface{}) (string, error)
}

// ParsedIntent represents a parsed player intent
type ParsedIntent struct {
	Command    string                 `json:"command"`
	Target     string                 `json:"target,omitempty"`
	Topic      string                 `json:"topic,omitempty"`
	Message    string                 `json:"message,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// MUDEventBus handles event distribution
type MUDEventBus interface {
	Publish(ctx context.Context, event MUDGameEvent) error
	Subscribe(ctx context.Context, eventType string, handler MUDEventHandler) error
	Unsubscribe(ctx context.Context, eventType string, handler MUDEventHandler) error
}

// MUDEventHandler processes events
type MUDEventHandler interface {
	Handle(ctx context.Context, event MUDGameEvent) error
}

// DSLInterpreter handles DSL content loading
type DSLInterpreter interface {
	LoadWorld(ctx context.Context, contentPath string) (*World, error)
	ValidateContent(ctx context.Context, contentPath string) error
	ReloadContent(ctx context.Context, contentPath string) (*World, error)
}

// ExpressionEvaluator handles expression evaluation
type ExpressionEvaluator interface {
	Evaluate(ctx context.Context, expression string, context map[string]interface{}) (interface{}, error)
	Compile(ctx context.Context, expression string) (CompiledExpression, error)
}

// CompiledExpression represents a compiled expression
type CompiledExpression interface {
	Evaluate(ctx context.Context, context map[string]interface{}) (interface{}, error)
}

// MarkdownAST represents a parsed Markdown document
type MarkdownAST interface {
	RenderToText(ctx context.Context) (string, error)
}

// MUDContentValidator validates game content
type MUDContentValidator interface {
	ValidateLocation(ctx context.Context, location Location) error
	ValidateCharacter(ctx context.Context, character *Character) error
	ValidateItem(ctx context.Context, item *ItemInstance) error
	ValidateQuest(ctx context.Context, quest *Quest) error
	ValidateWorld(ctx context.Context, world *World) error
}

// MUDPersistenceLayer handles data persistence
type MUDPersistenceLayer interface {
	SaveWorld(ctx context.Context, world *World) error
	LoadWorld(ctx context.Context) (*World, error)
	SaveCharacter(ctx context.Context, character *Character) error
	LoadCharacter(ctx context.Context, id CharacterId) (*Character, error)
	SaveLocation(ctx context.Context, location Location) error
	LoadLocation(ctx context.Context, id LocationId) (Location, error)
}
