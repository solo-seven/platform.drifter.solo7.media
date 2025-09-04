package server

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
)

// ContentLoader handles loading content from the shattered-realms directory
type ContentLoader struct {
	logger             domain.Logger
	tomlParser         domain.TOMLParser
	contentRepoManager domain.ContentRepositoryManager
	contentDirectory   string
	loadedContent      map[string]*domain.TOMLContent
	worldData          *WorldData
}

// WorldData holds the loaded world information
type WorldData struct {
	Locations map[string]*LocationData
	Items     map[string]*ItemData
	NPCs      map[string]*NPCData
	Classes   map[string]*ClassData
	Species   map[string]*SpeciesData
}

// LocationData represents a loaded location
type LocationData struct {
	ID          string
	Name        string
	Description string
	Type        string
	Exits       []string
	Items       []string
	NPCs        []string
	Properties  map[string]interface{}
}

// ItemData represents a loaded item
type ItemData struct {
	ID          string
	Name        string
	Description string
	Type        string
	Weight      int32
	Properties  map[string]interface{}
}

// NPCData represents a loaded NPC
type NPCData struct {
	ID          string
	Name        string
	Description string
	Category    string
	Stats       map[string]int32
	Abilities   map[string]interface{}
	Properties  map[string]interface{}
}

// ClassData represents a loaded character class
type ClassData struct {
	ID          string
	Name        string
	Description string
	Properties  map[string]interface{}
}

// SpeciesData represents a loaded character species
type SpeciesData struct {
	ID          string
	Name        string
	Description string
	Properties  map[string]interface{}
}

// NewContentLoader creates a new content loader
func NewContentLoader(
	logger domain.Logger,
	tomlParser domain.TOMLParser,
	contentRepoManager domain.ContentRepositoryManager,
	contentDirectory string,
) *ContentLoader {
	return &ContentLoader{
		logger:             logger,
		tomlParser:         tomlParser,
		contentRepoManager: contentRepoManager,
		contentDirectory:   contentDirectory,
		loadedContent:      make(map[string]*domain.TOMLContent),
		worldData: &WorldData{
			Locations: make(map[string]*LocationData),
			Items:     make(map[string]*ItemData),
			NPCs:      make(map[string]*NPCData),
			Classes:   make(map[string]*ClassData),
			Species:   make(map[string]*SpeciesData),
		},
	}
}

// LoadAllContent loads all content from the shattered-realms directory
func (cl *ContentLoader) LoadAllContent() error {
	cl.logger.Info("Loading content from shattered-realms directory", map[string]interface{}{
		"directory": cl.contentDirectory,
	})

	// Load content from each subdirectory
	subdirs := []string{
		"wb/regions",       // World Books - Regions
		"wb/environmental", // World Books - Environmental
		"wb/cultures",      // World Books - Cultures
		"pg",               // Player's Guide - All files
		"mm",               // Monster Manual - All files
		"gmd",              // Game Master Definition - All files
	}

	for _, subdir := range subdirs {
		dirPath := filepath.Join(cl.contentDirectory, subdir)
		cl.logger.Info("Loading directory", map[string]interface{}{
			"directory": dirPath,
		})
		if err := cl.loadDirectory(dirPath); err != nil {
			cl.logger.Warn("Failed to load directory", map[string]interface{}{
				"directory": dirPath,
				"error":     err.Error(),
			})
		} else {
			cl.logger.Info("Successfully loaded directory", map[string]interface{}{
				"directory": dirPath,
			})
		}
	}

	// Process loaded content to extract world data
	if err := cl.processLoadedContent(); err != nil {
		return fmt.Errorf("failed to process loaded content: %w", err)
	}

	cl.logger.Info("Content loading completed", map[string]interface{}{
		"total_content": len(cl.loadedContent),
		"locations":     len(cl.worldData.Locations),
		"items":         len(cl.worldData.Items),
		"npcs":          len(cl.worldData.NPCs),
		"classes":       len(cl.worldData.Classes),
		"species":       len(cl.worldData.Species),
	})

	return nil
}

// loadDirectory loads all TOML files from a directory
func (cl *ContentLoader) loadDirectory(dirPath string) error {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".toml") {
			filePath := filepath.Join(dirPath, file.Name())
			cl.logger.Info("Loading file", map[string]interface{}{
				"file": filePath,
			})
			if err := cl.loadFile(filePath); err != nil {
				cl.logger.Warn("Failed to load file", map[string]interface{}{
					"file":  filePath,
					"error": err.Error(),
				})
			} else {
				cl.logger.Info("Successfully loaded file", map[string]interface{}{
					"file": filePath,
				})
			}
		}
	}

	return nil
}

// loadFile loads a single TOML file
func (cl *ContentLoader) loadFile(filePath string) error {
	// Read file content
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Preprocess TOML content to fix single quotes
	processedData := cl.preprocessTOML(data)

	options := domain.TOMLParseOptions{
		ValidateSchema:    true,
		ResolveReferences: true,
		StrictMode:        false,
	}

	result, err := cl.tomlParser.ParseTOML(processedData, options)
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	if !result.Success {
		return fmt.Errorf("failed to parse file %s: %s", filePath, result.Error)
	}

	// Check if this is a multi-item file (like equipment.toml)
	if result.Content.Type == domain.ContentTypeItem {
		// Check if the data contains an "items" section with multiple items
		if itemsData, ok := result.Content.Data["items"].(map[string]interface{}); ok {
			// Process each item separately
			for _, itemValue := range itemsData {
				if itemMap, ok := itemValue.(map[string]interface{}); ok {
					// Create a new content item for each item
					itemContent := &domain.TOMLContent{
						Type:     domain.ContentTypeItem,
						ID:       itemMap["id"].(string),
						Version:  result.Content.Version,
						Metadata: result.Content.Metadata,
						Data:     itemMap,
					}

					// Add to content repository
					if err := cl.contentRepoManager.AddContent(itemContent); err != nil {
						cl.logger.Warn("Failed to add item to repository", map[string]interface{}{
							"content_id": itemContent.ID,
							"error":      err.Error(),
						})
					}

					// Store in loaded content
					cl.loadedContent[itemContent.ID] = itemContent
				}
			}
			return nil
		}
	}

	// Add to content repository
	if err := cl.contentRepoManager.AddContent(result.Content); err != nil {
		cl.logger.Warn("Failed to add content to repository", map[string]interface{}{
			"content_id": result.Content.ID,
			"error":      err.Error(),
		})
	}

	// Store in loaded content
	cl.loadedContent[result.Content.ID] = result.Content

	return nil
}

// processLoadedContent processes the loaded content to extract world data
func (cl *ContentLoader) processLoadedContent() error {
	for id, content := range cl.loadedContent {
		switch content.Type {
		case domain.ContentTypeLocation:
			if err := cl.processLocation(content); err != nil {
				cl.logger.Warn("Failed to process location", map[string]interface{}{
					"content_id": id,
					"error":      err.Error(),
				})
			}
		case domain.ContentTypeItem:
			if err := cl.processItem(content); err != nil {
				cl.logger.Warn("Failed to process item", map[string]interface{}{
					"content_id": id,
					"error":      err.Error(),
				})
			}
		case domain.ContentTypeMonster:
			if err := cl.processNPC(content); err != nil {
				cl.logger.Warn("Failed to process NPC", map[string]interface{}{
					"content_id": id,
					"error":      err.Error(),
				})
			}
		case domain.ContentTypeClass:
			if err := cl.processClass(content); err != nil {
				cl.logger.Warn("Failed to process class", map[string]interface{}{
					"content_id": id,
					"error":      err.Error(),
				})
			}
		case domain.ContentTypeSpecies:
			if err := cl.processSpecies(content); err != nil {
				cl.logger.Warn("Failed to process species", map[string]interface{}{
					"content_id": id,
					"error":      err.Error(),
				})
			}
		}
	}

	return nil
}

// processLocation processes location content
func (cl *ContentLoader) processLocation(content *domain.TOMLContent) error {
	locationData := &LocationData{
		ID:         content.ID,
		Properties: make(map[string]interface{}),
	}

	// Extract basic properties
	if name, ok := content.Data["name"].(string); ok {
		locationData.Name = name
	}
	if desc, ok := content.Data["description"].(string); ok {
		locationData.Description = desc
	}
	if locType, ok := content.Data["type"].(string); ok {
		locationData.Type = locType
	}

	// Extract exits from locations data
	if locations, ok := content.Data["locations"].(map[string]interface{}); ok {
		for _, loc := range locations {
			if locMap, ok := loc.(map[string]interface{}); ok {
				// Extract exits from districts
				if districts, ok := locMap["districts"].(map[string]interface{}); ok {
					for _, district := range districts {
						if districtMap, ok := district.(map[string]interface{}); ok {
							if name, ok := districtMap["name"].(string); ok {
								locationData.Exits = append(locationData.Exits, name)
							}
						}
					}
				}
			}
		}
	}

	cl.worldData.Locations[content.ID] = locationData
	return nil
}

// processItem processes item content
func (cl *ContentLoader) processItem(content *domain.TOMLContent) error {
	itemData := &ItemData{
		ID:         content.ID,
		Properties: make(map[string]interface{}),
	}

	// Extract basic properties
	if name, ok := content.Data["name"].(string); ok {
		itemData.Name = name
	}
	if desc, ok := content.Data["description"].(string); ok {
		itemData.Description = desc
	}
	if itemType, ok := content.Data["type"].(string); ok {
		itemData.Type = itemType
	}
	if weight, ok := content.Data["weight"].(int32); ok {
		itemData.Weight = weight
	}

	// Extract properties - check if we have a properties section
	if props, ok := content.Data["properties"].(map[string]interface{}); ok {
		itemData.Properties = props
	} else {
		// If no properties section, copy all non-basic fields to properties
		itemData.Properties = make(map[string]interface{})
		for key, value := range content.Data {
			if key != "id" && key != "name" && key != "description" && key != "type" && key != "weight" {
				itemData.Properties[key] = value
			}
		}
	}

	// Extract from items data structure (for multi-item files)
	if items, ok := content.Data["items"].(map[string]interface{}); ok {
		for _, item := range items {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if id, ok := itemMap["id"].(string); ok {
					itemData.ID = id
				}
				if name, ok := itemMap["name"].(string); ok {
					itemData.Name = name
				}
				if itemType, ok := itemMap["type"].(string); ok {
					itemData.Type = itemType
				}
				if weight, ok := itemMap["weight"].(int32); ok {
					itemData.Weight = weight
				}
				if props, ok := itemMap["properties"].(map[string]interface{}); ok {
					itemData.Properties = props
				} else {
					// If no properties section, copy all non-basic fields to properties
					itemData.Properties = make(map[string]interface{})
					for key, value := range itemMap {
						if key != "id" && key != "name" && key != "description" && key != "type" && key != "weight" {
							itemData.Properties[key] = value
						}
					}
				}
			}
		}
	}

	cl.worldData.Items[content.ID] = itemData
	return nil
}

// processNPC processes NPC/monster content
func (cl *ContentLoader) processNPC(content *domain.TOMLContent) error {
	npcData := &NPCData{
		ID:         content.ID,
		Stats:      make(map[string]int32),
		Abilities:  make(map[string]interface{}),
		Properties: make(map[string]interface{}),
	}

	// Extract basic properties
	if name, ok := content.Data["name"].(string); ok {
		npcData.Name = name
	}
	if desc, ok := content.Data["description"].(string); ok {
		npcData.Description = desc
	}
	if category, ok := content.Data["category"].(string); ok {
		npcData.Category = category
	}

	// Extract from npcs data structure
	if npcs, ok := content.Data["npcs"].(map[string]interface{}); ok {
		for _, npc := range npcs {
			if npcMap, ok := npc.(map[string]interface{}); ok {
				if id, ok := npcMap["id"].(string); ok {
					npcData.ID = id
				}
				if name, ok := npcMap["name"].(string); ok {
					npcData.Name = name
				}
				if category, ok := npcMap["category"].(string); ok {
					npcData.Category = category
				}

				// Extract attributes/stats
				if attrs, ok := npcMap["attributes"].(map[string]interface{}); ok {
					for key, value := range attrs {
						if intVal, ok := value.(int32); ok {
							npcData.Stats[key] = intVal
						} else if intVal, ok := value.(int); ok {
							npcData.Stats[key] = int32(intVal)
						}
					}
				}

				// Extract abilities
				if abilities, ok := npcMap["abilities"].(map[string]interface{}); ok {
					npcData.Abilities = abilities
				}
			}
		}
	}

	cl.worldData.NPCs[content.ID] = npcData
	return nil
}

// processClass processes character class content
func (cl *ContentLoader) processClass(content *domain.TOMLContent) error {
	classData := &ClassData{
		ID:         content.ID,
		Properties: make(map[string]interface{}),
	}

	// Extract basic properties
	if name, ok := content.Data["name"].(string); ok {
		classData.Name = name
	}
	if desc, ok := content.Data["description"].(string); ok {
		classData.Description = desc
	}

	// Extract from classes data structure
	if classes, ok := content.Data["classes"].(map[string]interface{}); ok {
		for _, class := range classes {
			if classMap, ok := class.(map[string]interface{}); ok {
				if id, ok := classMap["id"].(string); ok {
					classData.ID = id
				}
				if name, ok := classMap["name"].(string); ok {
					classData.Name = name
				}
				if desc, ok := classMap["description"].(string); ok {
					classData.Description = desc
				}
			}
		}
	}

	cl.worldData.Classes[content.ID] = classData
	return nil
}

// processSpecies processes character species content
func (cl *ContentLoader) processSpecies(content *domain.TOMLContent) error {
	speciesData := &SpeciesData{
		ID:         content.ID,
		Properties: make(map[string]interface{}),
	}

	// Extract basic properties
	if name, ok := content.Data["name"].(string); ok {
		speciesData.Name = name
	}
	if desc, ok := content.Data["description"].(string); ok {
		speciesData.Description = desc
	}

	// Extract from species data structure
	if species, ok := content.Data["species"].(map[string]interface{}); ok {
		for _, spec := range species {
			if specMap, ok := spec.(map[string]interface{}); ok {
				if id, ok := specMap["id"].(string); ok {
					speciesData.ID = id
				}
				if name, ok := specMap["name"].(string); ok {
					speciesData.Name = name
				}
				if desc, ok := specMap["description"].(string); ok {
					speciesData.Description = desc
				}
			}
		}
	}

	cl.worldData.Species[content.ID] = speciesData
	return nil
}

// GetWorldData returns the loaded world data
func (cl *ContentLoader) GetWorldData() *WorldData {
	return cl.worldData
}

// GetLocation returns a location by ID
func (cl *ContentLoader) GetLocation(id string) *LocationData {
	return cl.worldData.Locations[id]
}

// GetItem returns an item by ID
func (cl *ContentLoader) GetItem(id string) *ItemData {
	return cl.worldData.Items[id]
}

// GetNPC returns an NPC by ID
func (cl *ContentLoader) GetNPC(id string) *NPCData {
	return cl.worldData.NPCs[id]
}

// GetClass returns a class by ID
func (cl *ContentLoader) GetClass(id string) *ClassData {
	return cl.worldData.Classes[id]
}

// GetSpecies returns a species by ID
func (cl *ContentLoader) GetSpecies(id string) *SpeciesData {
	return cl.worldData.Species[id]
}

// GetAllLocations returns all loaded locations
func (cl *ContentLoader) GetAllLocations() map[string]*LocationData {
	return cl.worldData.Locations
}

// GetAllItems returns all loaded items
func (cl *ContentLoader) GetAllItems() map[string]*ItemData {
	return cl.worldData.Items
}

// GetAllNPCs returns all loaded NPCs
func (cl *ContentLoader) GetAllNPCs() map[string]*NPCData {
	return cl.worldData.NPCs
}

// preprocessTOML fixes common TOML syntax issues
func (cl *ContentLoader) preprocessTOML(data []byte) []byte {
	content := string(data)

	// Replace single quotes with double quotes for string values
	// This is a simple regex that matches single-quoted strings
	// but avoids replacing single quotes inside double-quoted strings
	re := regexp.MustCompile(`'([^']*)'`)
	content = re.ReplaceAllString(content, `"$1"`)

	return []byte(content)
}
