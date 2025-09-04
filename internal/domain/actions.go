package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Action implementations for MUD

// MoveAction represents a character movement action
type MoveAction struct {
	CharacterID CharacterId `json:"character_id"`
	Direction   string      `json:"direction"`
}

func (a *MoveAction) GetType() string             { return "move" }
func (a *MoveAction) GetCharacterID() CharacterId { return a.CharacterID }

func (a *MoveAction) Execute(ctx context.Context, world *World) (*MUDActionResult, error) {
	character, exists := world.Characters[a.CharacterID]
	if !exists {
		return &MUDActionResult{
			Success: false,
			Message: "Character not found",
		}, nil
	}

	location, exists := world.Locations[character.Position.LocationID]
	if !exists {
		return &MUDActionResult{
			Success: false,
			Message: "Current location not found",
		}, nil
	}

	// Find the exit in the specified direction
	exits := location.GetExits()
	var targetLocationID LocationId
	var exitDescription string
	var found bool

	for _, exit := range exits {
		if exit.Direction == a.Direction {
			targetLocationID = exit.TargetLocation
			exitDescription = exit.Description
			found = true
			break
		}
	}

	if !found {
		return &MUDActionResult{
			Success: false,
			Message: fmt.Sprintf("You cannot go %s from here", a.Direction),
		}, nil
	}

	// Check if target location exists
	targetLocation, exists := world.Locations[targetLocationID]
	if !exists {
		return &MUDActionResult{
			Success: false,
			Message: "Target location not found",
		}, nil
	}

	// Update character position
	oldLocationID := character.Position.LocationID
	character.Position.LocationID = targetLocationID

	// Update location player lists
	if area, ok := world.Locations[oldLocationID].(*Area); ok {
		area.Players = removeCharacterFromSlice(area.Players, a.CharacterID)
	}
	if structure, ok := world.Locations[oldLocationID].(*Structure); ok {
		structure.Players = removeCharacterFromSlice(structure.Players, a.CharacterID)
	}

	if area, ok := targetLocation.(*Area); ok {
		area.Players = append(area.Players, a.CharacterID)
	}
	if structure, ok := targetLocation.(*Structure); ok {
		structure.Players = append(structure.Players, a.CharacterID)
	}

	return &MUDActionResult{
		Success: true,
		Message: fmt.Sprintf("You %s %s", exitDescription, a.Direction),
		StateChanges: []MUDStateChange{
			{
				EntityID:  a.CharacterID.String(),
				Component: "position",
				Changes: map[string]interface{}{
					"location_id": targetLocationID.String(),
				},
				Timestamp: time.Now(),
			},
		},
		Events: []MUDGameEvent{
			{
				ID:   uuid.New().String(),
				Type: "player_moved",
				Data: map[string]interface{}{
					"character_id":  a.CharacterID.String(),
					"from_location": oldLocationID.String(),
					"to_location":   targetLocationID.String(),
					"direction":     a.Direction,
				},
				Timestamp: time.Now(),
			},
		},
	}, nil
}

// LookAction represents a look action
type LookAction struct {
	CharacterID CharacterId `json:"character_id"`
	Target      string      `json:"target"`
}

func (a *LookAction) GetType() string             { return "look" }
func (a *LookAction) GetCharacterID() CharacterId { return a.CharacterID }

func (a *LookAction) Execute(ctx context.Context, world *World) (*MUDActionResult, error) {
	character, exists := world.Characters[a.CharacterID]
	if !exists {
		return &MUDActionResult{
			Success: false,
			Message: "Character not found",
		}, nil
	}

	location, exists := world.Locations[character.Position.LocationID]
	if !exists {
		return &MUDActionResult{
			Success: false,
			Message: "Current location not found",
		}, nil
	}

	if a.Target == "" {
		// Look at current room
		description := location.GetDescription()
		exits := location.GetExits()
		items := location.GetItems()
		npcs := location.GetNPCs()
		players := location.GetPlayers()

		var exitNames []string
		for _, exit := range exits {
			exitNames = append(exitNames, exit.Direction)
		}

		var itemNames []string
		for _, item := range items {
			itemNames = append(itemNames, item.Name)
		}

		var npcNames []string
		for _, npc := range npcs {
			npcNames = append(npcNames, npc.Name)
		}

		var playerNames []string
		for _, playerID := range players {
			if playerID != a.CharacterID {
				if player, exists := world.Characters[playerID]; exists {
					playerNames = append(playerNames, player.Name)
				}
			}
		}

		message := fmt.Sprintf("%s\n\nExits: %v\nItems: %v\nNPCs: %v\nPlayers: %v",
			description, exitNames, itemNames, npcNames, playerNames)

		return &MUDActionResult{
			Success: true,
			Message: message,
		}, nil
	}

	// Look at specific target
	// TODO: Implement specific target looking
	return &MUDActionResult{
		Success: true,
		Message: fmt.Sprintf("You look at %s", a.Target),
	}, nil
}

// GetAction represents picking up an item
type GetAction struct {
	CharacterID CharacterId `json:"character_id"`
	ItemName    string      `json:"item_name"`
}

func (a *GetAction) GetType() string             { return "get" }
func (a *GetAction) GetCharacterID() CharacterId { return a.CharacterID }

func (a *GetAction) Execute(ctx context.Context, world *World) (*MUDActionResult, error) {
	character, exists := world.Characters[a.CharacterID]
	if !exists {
		return &MUDActionResult{
			Success: false,
			Message: "Character not found",
		}, nil
	}

	location, exists := world.Locations[character.Position.LocationID]
	if !exists {
		return &MUDActionResult{
			Success: false,
			Message: "Current location not found",
		}, nil
	}

	// Find the item in the location
	items := location.GetItems()
	var targetItem *ItemInstance
	var targetIndex int = -1

	for i, item := range items {
		if item.Name == a.ItemName {
			targetItem = &item
			targetIndex = i
			break
		}
	}

	if targetItem == nil {
		return &MUDActionResult{
			Success: false,
			Message: fmt.Sprintf("There is no %s here", a.ItemName),
		}, nil
	}

	// Check if character can carry the item
	if int32(len(character.Inventory.Items)) >= character.Inventory.Capacity {
		return &MUDActionResult{
			Success: false,
			Message: "Your inventory is full",
		}, nil
	}

	// Add item to character inventory
	character.Inventory.Items = append(character.Inventory.Items, *targetItem)

	// Remove item from location
	if area, ok := location.(*Area); ok {
		area.Items = append(area.Items[:targetIndex], area.Items[targetIndex+1:]...)
	}
	if structure, ok := location.(*Structure); ok {
		structure.Items = append(structure.Items[:targetIndex], structure.Items[targetIndex+1:]...)
	}

	return &MUDActionResult{
		Success: true,
		Message: fmt.Sprintf("You pick up the %s", a.ItemName),
		StateChanges: []MUDStateChange{
			{
				EntityID:  a.CharacterID.String(),
				Component: "inventory",
				Changes: map[string]interface{}{
					"added_item": targetItem.ID.String(),
				},
				Timestamp: time.Now(),
			},
		},
	}, nil
}

// SayAction represents speaking
type SayAction struct {
	CharacterID CharacterId `json:"character_id"`
	Message     string      `json:"message"`
	Channel     string      `json:"channel"`
	Target      string      `json:"target,omitempty"`
}

func (a *SayAction) GetType() string             { return "say" }
func (a *SayAction) GetCharacterID() CharacterId { return a.CharacterID }

func (a *SayAction) Execute(ctx context.Context, world *World) (*MUDActionResult, error) {
	character, exists := world.Characters[a.CharacterID]
	if !exists {
		return &MUDActionResult{
			Success: false,
			Message: "Character not found",
		}, nil
	}

	// Create chat message event
	event := MUDGameEvent{
		ID:   uuid.New().String(),
		Type: "chat_message",
		Data: map[string]interface{}{
			"character_id":   a.CharacterID.String(),
			"character_name": character.Name,
			"message":        a.Message,
			"channel":        a.Channel,
			"target":         a.Target,
		},
		Timestamp: time.Now(),
	}

	return &MUDActionResult{
		Success: true,
		Message: "Message sent",
		Events:  []MUDGameEvent{event},
	}, nil
}

// AttackAction represents an attack
type AttackAction struct {
	CharacterID CharacterId `json:"character_id"`
	Target      string      `json:"target"`
	Weapon      string      `json:"weapon,omitempty"`
}

func (a *AttackAction) GetType() string             { return "attack" }
func (a *AttackAction) GetCharacterID() CharacterId { return a.CharacterID }

func (a *AttackAction) Execute(ctx context.Context, world *World) (*MUDActionResult, error) {
	_, exists := world.Characters[a.CharacterID]
	if !exists {
		return &MUDActionResult{
			Success: false,
			Message: "Character not found",
		}, nil
	}

	// TODO: Implement actual combat logic
	// For now, return a mock response
	return &MUDActionResult{
		Success: true,
		Message: fmt.Sprintf("You attack %s", a.Target),
		Events: []MUDGameEvent{
			{
				ID:   uuid.New().String(),
				Type: "combat_log",
				Data: map[string]interface{}{
					"attacker_id": a.CharacterID.String(),
					"target":      a.Target,
					"damage":      5,
					"result":      "hit",
				},
				Timestamp: time.Now(),
			},
		},
	}, nil
}

// Helper function to remove character from slice
func removeCharacterFromSlice(slice []CharacterId, id CharacterId) []CharacterId {
	for i, v := range slice {
		if v == id {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
