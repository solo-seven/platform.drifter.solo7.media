package server

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
)

// TOMLParserImpl implements the TOMLParser interface
type TOMLParserImpl struct {
	logger domain.Logger
}

// NewTOMLParser creates a new TOML parser instance
func NewTOMLParser(logger domain.Logger) domain.TOMLParser {
	return &TOMLParserImpl{
		logger: logger,
	}
}

// ParseTOML parses TOML data into structured content
func (tp *TOMLParserImpl) ParseTOML(data []byte, options domain.TOMLParseOptions) (*domain.TOMLParseResult, error) {
	tp.logger.Debug("Parsing TOML data", map[string]interface{}{
		"data_size": len(data),
		"options":   options,
	})

	// Parse TOML data
	var rawData map[string]interface{}
	if err := toml.Unmarshal(data, &rawData); err != nil {
		return &domain.TOMLParseResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse TOML: %v", err),
		}, nil
	}

	// Extract content type and ID
	contentType, err := tp.extractContentType(rawData)
	if err != nil {
		return &domain.TOMLParseResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to extract content type: %v", err),
		}, nil
	}

	// Check if content type is allowed
	if len(options.AllowedTypes) > 0 {
		if !tp.isContentTypeAllowed(contentType, options.AllowedTypes) {
			return &domain.TOMLParseResult{
				Success: false,
				Error:   fmt.Sprintf("Content type %s is not allowed", contentType),
			}, nil
		}
	}

	// Extract ID
	id, err := tp.extractID(rawData)
	if err != nil {
		return &domain.TOMLParseResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to extract ID: %v", err),
		}, nil
	}

	// Extract version
	version := tp.extractVersion(rawData)

	// Extract metadata
	metadata := tp.extractMetadata(rawData)

	// Extract expressions
	expressions, err := tp.ExtractExpressions(&domain.TOMLContent{Data: rawData})
	if err != nil {
		tp.logger.Warn("Failed to extract expressions", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Extract references
	references, err := tp.ExtractReferences(&domain.TOMLContent{Data: rawData})
	if err != nil {
		tp.logger.Warn("Failed to extract references", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Create content object
	content := &domain.TOMLContent{
		Type:        contentType,
		ID:          id,
		Version:     version,
		Metadata:    metadata,
		Data:        rawData,
		Expressions: expressions,
		References:  references,
	}

	// Validate content if requested
	var validation *domain.ContentValidation
	if options.ValidateSchema {
		validation, err = tp.ValidateContent(content, options)
		if err != nil {
			tp.logger.Warn("Failed to validate content", map[string]interface{}{
				"error": err.Error(),
			})
		}
		content.Validation = *validation
	}

	tp.logger.Info("Successfully parsed TOML content", map[string]interface{}{
		"content_type": contentType,
		"id":           id,
		"version":      version,
		"expressions":  len(expressions),
		"references":   len(references),
	})

	return &domain.TOMLParseResult{
		Content: content,
		Success: true,
		Metadata: map[string]interface{}{
			"parsed_at": time.Now(),
			"size":      len(data),
		},
	}, nil
}

// ParseTOMLFile parses a TOML file into structured content
func (tp *TOMLParserImpl) ParseTOMLFile(filePath string, options domain.TOMLParseOptions) (*domain.TOMLParseResult, error) {
	tp.logger.Debug("Parsing TOML file", map[string]interface{}{
		"file_path": filePath,
		"options":   options,
	})

	// Read file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return &domain.TOMLParseResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to read file: %v", err),
		}, nil
	}

	// Parse TOML data
	result, err := tp.ParseTOML(data, options)
	if err != nil {
		return result, err
	}

	// Add file metadata
	if result.Success && result.Content != nil {
		result.Content.Metadata["file_path"] = filePath
		result.Content.Metadata["file_name"] = filepath.Base(filePath)
		result.Metadata["file_path"] = filePath
	}

	return result, nil
}

// ValidateContent validates parsed content against schema and rules
func (tp *TOMLParserImpl) ValidateContent(content *domain.TOMLContent, options domain.TOMLParseOptions) (*domain.ContentValidation, error) {
	validation := &domain.ContentValidation{
		IsValid:     true,
		SchemaValid: true,
	}

	// Basic validation
	if content.ID == "" {
		validation.Errors = append(validation.Errors, domain.ValidationError{
			Field:   "id",
			Message: "ID is required",
			Code:    "MISSING_ID",
		})
		validation.IsValid = false
	}

	if content.Type == "" {
		validation.Errors = append(validation.Errors, domain.ValidationError{
			Field:   "type",
			Message: "Type is required",
			Code:    "MISSING_TYPE",
		})
		validation.IsValid = false
	}

	// Validate expressions if any
	for field, expression := range content.Expressions {
		if err := tp.validateExpression(expression); err != nil {
			validation.Errors = append(validation.Errors, domain.ValidationError{
				Field:   field,
				Message: fmt.Sprintf("Invalid expression: %v", err),
				Code:    "INVALID_EXPRESSION",
			})
			validation.IsValid = false
		}
	}

	// Validate references if any
	if len(content.References) > 0 {
		refValidation := &domain.ReferenceValidation{
			ValidReferences:   content.References,
			InvalidReferences: []string{},
			MissingReferences: []string{},
		}
		validation.References = *refValidation
	}

	return validation, nil
}

// ExtractExpressions extracts expression strings from content data
func (tp *TOMLParserImpl) ExtractExpressions(content *domain.TOMLContent) (map[string]string, error) {
	expressions := make(map[string]string)

	if content == nil || content.Data == nil {
		return expressions, nil
	}

	// Recursively search for expression patterns
	tp.extractExpressionsRecursive(content.Data, "", expressions)

	return expressions, nil
}

// ExtractReferences extracts cross-references from content data
func (tp *TOMLParserImpl) ExtractReferences(content *domain.TOMLContent) ([]string, error) {
	references := make([]string, 0)

	if content == nil || content.Data == nil {
		return references, nil
	}

	// Recursively search for reference patterns
	references = tp.extractReferencesRecursive(content.Data, references)

	return references, nil
}

// Helper methods

func (tp *TOMLParserImpl) extractContentType(data map[string]interface{}) (domain.ContentType, error) {
	// Try to extract from explicit type field
	if typeValue, exists := data["type"]; exists {
		if typeStr, ok := typeValue.(string); ok {
			return domain.ContentType(typeStr), nil
		}
	}

	// Try to infer from ID prefix
	if idValue, exists := data["id"]; exists {
		if idStr, ok := idValue.(string); ok {
			parts := strings.Split(idStr, ".")
			if len(parts) > 0 {
				switch parts[0] {
				case "class":
					return domain.ContentTypeClass, nil
				case "species":
					return domain.ContentTypeSpecies, nil
				case "item":
					return domain.ContentTypeItem, nil
				case "spell":
					return domain.ContentTypeSpell, nil
				case "ability":
					return domain.ContentTypeAbility, nil
				case "monster":
					return domain.ContentTypeMonster, nil
				case "encounter":
					return domain.ContentTypeEncounter, nil
				case "location":
					return domain.ContentTypeLocation, nil
				case "mechanic":
					return domain.ContentTypeMechanic, nil
				case "action":
					return domain.ContentTypeAction, nil
				case "resolution":
					return domain.ContentTypeResolution, nil
				case "modifier":
					return domain.ContentTypeModifier, nil
				case "randomization":
					return domain.ContentTypeRandomization, nil
				}
			}
		}
	}

	// Try to infer from section names (for shattered-realms format)
	for key, value := range data {
		if strings.HasPrefix(key, "npcs.") || strings.HasPrefix(key, "creatures.") {
			return domain.ContentTypeMonster, nil
		}
		if strings.HasPrefix(key, "items.") || strings.HasPrefix(key, "equipment.") {
			return domain.ContentTypeItem, nil
		}
		if key == "items" {
			// Check if this is a map containing item data
			if itemsMap, ok := value.(map[string]interface{}); ok {
				// Check if any of the values have item-like structure
				for _, item := range itemsMap {
					if itemMap, ok := item.(map[string]interface{}); ok {
						if _, hasID := itemMap["id"]; hasID {
							if idStr, ok := itemMap["id"].(string); ok && strings.HasPrefix(idStr, "item.") {
								return domain.ContentTypeItem, nil
							}
						}
					}
				}
			}
		}
		if strings.HasPrefix(key, "classes.") {
			return domain.ContentTypeClass, nil
		}
		if strings.HasPrefix(key, "species.") {
			return domain.ContentTypeSpecies, nil
		}
		if strings.HasPrefix(key, "locations.") || strings.HasPrefix(key, "regions.") {
			return domain.ContentTypeLocation, nil
		}
		if strings.HasPrefix(key, "world_book") {
			return domain.ContentTypeLocation, nil
		}
		if strings.HasPrefix(key, "behavior_profiles.") {
			return domain.ContentTypeMechanic, nil
		}
		if strings.HasPrefix(key, "encounters.") {
			return domain.ContentTypeEncounter, nil
		}
		if strings.HasPrefix(key, "rules.") || strings.HasPrefix(key, "mechanics.") {
			return domain.ContentTypeMechanic, nil
		}
		// Check if value is a map with nested content
		if valueMap, ok := value.(map[string]interface{}); ok {
			if id, exists := valueMap["id"]; exists {
				if idStr, ok := id.(string); ok {
					parts := strings.Split(idStr, ".")
					if len(parts) > 0 {
						switch parts[0] {
						case "npc", "creature":
							return domain.ContentTypeMonster, nil
						case "item", "equipment":
							return domain.ContentTypeItem, nil
						case "class":
							return domain.ContentTypeClass, nil
						case "species":
							return domain.ContentTypeSpecies, nil
						case "location", "district", "region":
							return domain.ContentTypeLocation, nil
						case "behavior", "profile":
							return domain.ContentTypeMechanic, nil
						case "encounter":
							return domain.ContentTypeEncounter, nil
						case "rule", "mechanic":
							return domain.ContentTypeMechanic, nil
						}
					}
				}
			}
		}
	}

	// If we still can't determine the type, try to infer from the file path or content structure
	// This is a fallback for files that don't follow standard patterns
	return domain.ContentTypeMechanic, nil // Default to mechanic for unknown content
}

func (tp *TOMLParserImpl) extractID(data map[string]interface{}) (string, error) {
	// Try to get ID from top-level field first
	if idValue, exists := data["id"]; exists {
		if idStr, ok := idValue.(string); ok {
			return idStr, nil
		}
	}

	// For shattered-realms format, look for ID in nested content
	for _, value := range data {
		if valueMap, ok := value.(map[string]interface{}); ok {
			if id, exists := valueMap["id"]; exists {
				if idStr, ok := id.(string); ok {
					return idStr, nil
				}
			}
		}
	}

	// Special case for multi-item files (like equipment.toml)
	// If we have an "items" section, we can't extract a single ID
	// Return a placeholder that will be handled by the content loader
	if items, exists := data["items"]; exists {
		if itemsMap, ok := items.(map[string]interface{}); ok {
			// Check if any of the items have IDs
			for _, item := range itemsMap {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if id, exists := itemMap["id"]; exists {
						if idStr, ok := id.(string); ok {
							// Return the first item's ID as a placeholder
							// The content loader will handle creating separate content items
							return idStr, nil
						}
					}
				}
			}
		}
	}

	// If no ID found, generate one from the first section key
	for key := range data {
		if strings.Contains(key, ".") {
			// Use the section key as ID
			return key, nil
		}
	}

	return "", fmt.Errorf("ID field not found or invalid")
}

func (tp *TOMLParserImpl) extractVersion(data map[string]interface{}) string {
	if versionValue, exists := data["version"]; exists {
		if versionStr, ok := versionValue.(string); ok {
			return versionStr
		}
	}
	return "1.0.0" // Default version
}

func (tp *TOMLParserImpl) extractMetadata(data map[string]interface{}) map[string]interface{} {
	metadata := make(map[string]interface{})

	// Extract common metadata fields
	metadataFields := []string{"name", "description", "author", "created_at", "updated_at", "tags"}

	for _, field := range metadataFields {
		if value, exists := data[field]; exists {
			metadata[field] = value
		}
	}

	return metadata
}

func (tp *TOMLParserImpl) isContentTypeAllowed(contentType domain.ContentType, allowedTypes []domain.ContentType) bool {
	for _, allowedType := range allowedTypes {
		if contentType == allowedType {
			return true
		}
	}
	return false
}

func (tp *TOMLParserImpl) validateExpression(expression string) error {
	// Basic expression validation
	// This is a placeholder - proper validation would use the expression parser

	// Check for basic syntax issues
	if strings.TrimSpace(expression) == "" {
		return fmt.Errorf("expression cannot be empty")
	}

	// Check for balanced parentheses
	openCount := strings.Count(expression, "(")
	closeCount := strings.Count(expression, ")")
	if openCount != closeCount {
		return fmt.Errorf("unbalanced parentheses")
	}

	return nil
}

func (tp *TOMLParserImpl) extractExpressionsRecursive(data interface{}, path string, expressions map[string]string) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			currentPath := key
			if path != "" {
				currentPath = path + "." + key
			}
			tp.extractExpressionsRecursive(value, currentPath, expressions)
		}
	case []interface{}:
		for i, value := range v {
			currentPath := fmt.Sprintf("%s[%d]", path, i)
			tp.extractExpressionsRecursive(value, currentPath, expressions)
		}
	case string:
		// Check if this looks like an expression
		if tp.isExpressionString(v) {
			expressions[path] = v
		}
	}
}

func (tp *TOMLParserImpl) isExpressionString(str string) bool {
	// Simple heuristic to identify expression strings
	// Look for common expression patterns

	// Check for function calls
	if matched, _ := regexp.MatchString(`\w+\(`, str); matched {
		return true
	}

	// Check for dice notation
	if matched, _ := regexp.MatchString(`\d+d\d+`, str); matched {
		return true
	}

	// Check for variable references
	if matched, _ := regexp.MatchString(`\$\w+`, str); matched {
		return true
	}

	// Check for self/target references
	if matched, _ := regexp.MatchString(`(self|target|party|terrain)\.`, str); matched {
		return true
	}

	return false
}

func (tp *TOMLParserImpl) extractReferencesRecursive(data interface{}, references []string) []string {
	switch v := data.(type) {
	case map[string]interface{}:
		for _, value := range v {
			references = tp.extractReferencesRecursive(value, references)
		}
	case []interface{}:
		for _, value := range v {
			references = tp.extractReferencesRecursive(value, references)
		}
	case string:
		// Check if this looks like a reference
		if tp.isReferenceString(v) {
			references = append(references, v)
		}
	}
	return references
}

func (tp *TOMLParserImpl) isReferenceString(str string) bool {
	// Simple heuristic to identify reference strings
	// Look for common reference patterns

	// Check for @reference pattern
	if matched, _ := regexp.MatchString(`@\w+/\w+`, str); matched {
		return true
	}

	// Check for ID-like pattern (e.g., item.iron_sword, class.fighter)
	if matched, _ := regexp.MatchString(`^\w+\.\w+$`, str); matched {
		return true
	}

	return false
}
