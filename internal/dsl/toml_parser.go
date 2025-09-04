package dsl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pelletier/go-toml/v2"
	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
)

// TOMLParser handles parsing TOML files for game content
type TOMLParser struct {
	logger domain.Logger
}

// NewTOMLParser creates a new TOML parser
func NewTOMLParser(logger domain.Logger) *TOMLParser {
	return &TOMLParser{
		logger: logger,
	}
}

// ParseWorld loads a complete world from TOML files
func (p *TOMLParser) ParseWorld(ctx context.Context, contentPath string) (*domain.World, error) {
	world := &domain.World{
		ID:         "default_world",
		Name:       "Default World",
		Locations:  make(map[domain.LocationId]domain.Location),
		Characters: make(map[domain.CharacterId]*domain.Character),
		NPCs:       make(map[domain.NPCId]*domain.NPCInstance),
		Items:      make(map[domain.ItemId]*domain.ItemInstance),
		Quests:     make(map[domain.QuestId]*domain.Quest),
		Metadata: domain.WorldMetadata{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Version:   1,
		},
	}

	// Parse rooms/areas
	if err := p.parseRooms(ctx, contentPath, world); err != nil {
		return nil, fmt.Errorf("failed to parse rooms: %w", err)
	}

	// Parse items
	if err := p.parseItems(ctx, contentPath, world); err != nil {
		return nil, fmt.Errorf("failed to parse items: %w", err)
	}

	// Parse NPCs
	if err := p.parseNPCs(ctx, contentPath, world); err != nil {
		return nil, fmt.Errorf("failed to parse NPCs: %w", err)
	}

	// Parse quests
	if err := p.parseQuests(ctx, contentPath, world); err != nil {
		return nil, fmt.Errorf("failed to parse quests: %w", err)
	}

	return world, nil
}

// parseRooms parses room/area TOML files
func (p *TOMLParser) parseRooms(ctx context.Context, contentPath string, world *domain.World) error {
	roomsPath := filepath.Join(contentPath, "rooms")
	if _, err := os.Stat(roomsPath); os.IsNotExist(err) {
		p.logger.Info("No rooms directory found, skipping room parsing", map[string]interface{}{})
		return nil
	}

	return filepath.Walk(roomsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".toml") {
			return nil
		}

		room, err := p.parseRoomFile(ctx, path)
		if err != nil {
			return fmt.Errorf("failed to parse room file %s: %w", path, err)
		}

		world.Locations[room.GetID()] = room
		return nil
	})
}

// parseRoomFile parses a single room TOML file
func (p *TOMLParser) parseRoomFile(ctx context.Context, filePath string) (domain.Location, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var roomData map[string]interface{}
	if err := toml.Unmarshal(data, &roomData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal TOML: %w", err)
	}

	// Determine if this is an area or structure
	roomType, ok := roomData["type"].(string)
	if !ok {
		roomType = "structure" // default to structure
	}

	switch roomType {
	case "area":
		return p.parseArea(roomData)
	case "structure":
		return p.parseStructure(roomData)
	default:
		return nil, fmt.Errorf("unknown room type: %s", roomType)
	}
}

// parseArea parses an area from TOML data
func (p *TOMLParser) parseArea(data map[string]interface{}) (*domain.Area, error) {
	area := &domain.Area{
		ID:          domain.LocationId(uuid.New()),
		Name:        getString(data, "name", "Unnamed Area"),
		Description: getString(data, "description", ""),
		Boundaries:  []domain.Boundary{},
		Entrances:   []domain.Entrance{},
		Items:       []domain.ItemInstance{},
		NPCs:        []domain.NPCInstance{},
		Players:     []domain.CharacterId{},
	}

	// Parse boundaries
	if boundariesData, ok := data["boundaries"].([]interface{}); ok {
		for _, boundaryData := range boundariesData {
			if boundaryMap, ok := boundaryData.(map[string]interface{}); ok {
				boundary := domain.Boundary{
					Direction:    getString(boundaryMap, "direction", ""),
					TargetAreaId: domain.LocationId(uuid.New()), // TODO: resolve actual ID
					Description:  getString(boundaryMap, "description", ""),
				}
				area.Boundaries = append(area.Boundaries, boundary)
			}
		}
	}

	// Parse entrances
	if entrancesData, ok := data["entrances"].([]interface{}); ok {
		for _, entranceData := range entrancesData {
			if entranceMap, ok := entranceData.(map[string]interface{}); ok {
				entrance := domain.Entrance{
					Name:              getString(entranceMap, "name", ""),
					TargetStructureId: domain.LocationId(uuid.New()), // TODO: resolve actual ID
					Description:       getString(entranceMap, "description", ""),
				}
				area.Entrances = append(area.Entrances, entrance)
			}
		}
	}

	return area, nil
}

// parseStructure parses a structure from TOML data
func (p *TOMLParser) parseStructure(data map[string]interface{}) (*domain.Structure, error) {
	structure := &domain.Structure{
		ID:          domain.LocationId(uuid.New()),
		Name:        getString(data, "name", "Unnamed Structure"),
		Description: getString(data, "description", ""),
		Exits:       []domain.Exit{},
		Items:       []domain.ItemInstance{},
		NPCs:        []domain.NPCInstance{},
		Players:     []domain.CharacterId{},
	}

	// Parse exits
	if exitsData, ok := data["exits"].([]interface{}); ok {
		for _, exitData := range exitsData {
			if exitMap, ok := exitData.(map[string]interface{}); ok {
				exit := domain.Exit{
					Direction:      getString(exitMap, "direction", ""),
					TargetLocation: domain.LocationId(uuid.New()), // TODO: resolve actual ID
					Description:    getString(exitMap, "description", ""),
				}
				structure.Exits = append(structure.Exits, exit)
			}
		}
	}

	return structure, nil
}

// parseItems parses item TOML files
func (p *TOMLParser) parseItems(ctx context.Context, contentPath string, world *domain.World) error {
	itemsPath := filepath.Join(contentPath, "items")
	if _, err := os.Stat(itemsPath); os.IsNotExist(err) {
		p.logger.Info("No items directory found, skipping item parsing", map[string]interface{}{})
		return nil
	}

	return filepath.Walk(itemsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".toml") {
			return nil
		}

		item, err := p.parseItemFile(ctx, path)
		if err != nil {
			return fmt.Errorf("failed to parse item file %s: %w", path, err)
		}

		world.Items[item.ID] = item
		return nil
	})
}

// parseItemFile parses a single item TOML file
func (p *TOMLParser) parseItemFile(ctx context.Context, filePath string) (*domain.ItemInstance, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var itemData map[string]interface{}
	if err := toml.Unmarshal(data, &itemData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal TOML: %w", err)
	}

	item := &domain.ItemInstance{
		ID:          domain.ItemId(uuid.New()),
		TemplateID:  getString(itemData, "template_id", ""),
		Name:        getString(itemData, "name", "Unnamed Item"),
		Description: getString(itemData, "description", ""),
		Type:        getString(itemData, "type", "misc"),
		Weight:      int32(getInt(itemData, "weight", 1)),
		Properties:  make(map[string]interface{}),
		Quantity:    1,
	}

	// Parse properties
	if propertiesData, ok := itemData["properties"].(map[string]interface{}); ok {
		item.Properties = propertiesData
	}

	return item, nil
}

// parseNPCs parses NPC TOML files
func (p *TOMLParser) parseNPCs(ctx context.Context, contentPath string, world *domain.World) error {
	npcsPath := filepath.Join(contentPath, "npcs")
	if _, err := os.Stat(npcsPath); os.IsNotExist(err) {
		p.logger.Info("No NPCs directory found, skipping NPC parsing", map[string]interface{}{})
		return nil
	}

	return filepath.Walk(npcsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".toml") {
			return nil
		}

		npc, err := p.parseNPCFile(ctx, path)
		if err != nil {
			return fmt.Errorf("failed to parse NPC file %s: %w", path, err)
		}

		world.NPCs[npc.ID] = npc
		return nil
	})
}

// parseNPCFile parses a single NPC TOML file
func (p *TOMLParser) parseNPCFile(ctx context.Context, filePath string) (*domain.NPCInstance, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var npcData map[string]interface{}
	if err := toml.Unmarshal(data, &npcData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal TOML: %w", err)
	}

	npc := &domain.NPCInstance{
		ID:          domain.NPCId(uuid.New()),
		TemplateID:  getString(npcData, "template_id", ""),
		Name:        getString(npcData, "name", "Unnamed NPC"),
		Description: getString(npcData, "description", ""),
		Stats: domain.Stats{
			Health:     int32(getInt(npcData, "health", 100)),
			MaxHealth:  int32(getInt(npcData, "max_health", 100)),
			Mana:       int32(getInt(npcData, "mana", 50)),
			MaxMana:    int32(getInt(npcData, "max_mana", 50)),
			Attributes: make(map[string]int32),
		},
		Inventory: domain.MUDInventory{
			Items: []domain.ItemInstance{},
		},
		Abilities: []domain.MUDAbility{},
		AI: domain.AIBehavior{
			Type:       getString(npcData, "ai_type", "passive"),
			Properties: make(map[string]interface{}),
		},
	}

	// Parse attributes
	if attributesData, ok := npcData["attributes"].(map[string]interface{}); ok {
		for key, value := range attributesData {
			if intValue, ok := value.(int64); ok {
				npc.Stats.Attributes[key] = int32(intValue)
			}
		}
	}

	return npc, nil
}

// parseQuests parses quest TOML files
func (p *TOMLParser) parseQuests(ctx context.Context, contentPath string, world *domain.World) error {
	questsPath := filepath.Join(contentPath, "quests")
	if _, err := os.Stat(questsPath); os.IsNotExist(err) {
		p.logger.Info("No quests directory found, skipping quest parsing", map[string]interface{}{})
		return nil
	}

	return filepath.Walk(questsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".toml") {
			return nil
		}

		quest, err := p.parseQuestFile(ctx, path)
		if err != nil {
			return fmt.Errorf("failed to parse quest file %s: %w", path, err)
		}

		world.Quests[quest.ID] = quest
		return nil
	})
}

// parseQuestFile parses a single quest TOML file
func (p *TOMLParser) parseQuestFile(ctx context.Context, filePath string) (*domain.Quest, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var questData map[string]interface{}
	if err := toml.Unmarshal(data, &questData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal TOML: %w", err)
	}

	quest := &domain.Quest{
		ID:          domain.QuestId(uuid.New()),
		Name:        getString(questData, "name", "Unnamed Quest"),
		Description: getString(questData, "description", ""),
		Status:      domain.QuestStatusAvailable,
		Objectives:  []domain.QuestObjective{},
		Rewards:     []domain.QuestReward{},
		Properties:  make(map[string]interface{}),
	}

	// Parse objectives
	if objectivesData, ok := questData["objectives"].([]interface{}); ok {
		for _, objectiveData := range objectivesData {
			if objectiveMap, ok := objectiveData.(map[string]interface{}); ok {
				objective := domain.QuestObjective{
					ID:          getString(objectiveMap, "id", ""),
					Description: getString(objectiveMap, "description", ""),
					Type:        getString(objectiveMap, "type", ""),
					Target:      getString(objectiveMap, "target", ""),
					Quantity:    int32(getInt(objectiveMap, "quantity", 1)),
					Completed:   false,
					Properties:  make(map[string]interface{}),
				}
				quest.Objectives = append(quest.Objectives, objective)
			}
		}
	}

	return quest, nil
}

// Helper functions
func getString(data map[string]interface{}, key, defaultValue string) string {
	if value, ok := data[key].(string); ok {
		return value
	}
	return defaultValue
}

func getInt(data map[string]interface{}, key string, defaultValue int) int {
	if value, ok := data[key].(int64); ok {
		return int(value)
	}
	if value, ok := data[key].(int); ok {
		return value
	}
	return defaultValue
}
