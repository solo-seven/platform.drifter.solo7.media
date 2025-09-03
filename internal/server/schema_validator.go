package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/solo7.media/platform.drifter.solo7.media/internal/domain"
)

// SchemaValidatorImpl implements the SchemaValidator interface
type SchemaValidatorImpl struct {
	schemas map[domain.ContentType]interface{}
	logger  domain.Logger
}

// NewSchemaValidator creates a new schema validator
func NewSchemaValidator(logger domain.Logger) domain.SchemaValidator {
	return &SchemaValidatorImpl{
		schemas: make(map[domain.ContentType]interface{}),
		logger:  logger,
	}
}

// ValidateAgainstSchema validates content against a JSON schema
func (sv *SchemaValidatorImpl) ValidateAgainstSchema(content *domain.TOMLContent, schemaPath string) (*domain.ContentValidation, error) {
	if content == nil {
		return &domain.ContentValidation{
			IsValid: false,
			Errors: []domain.ValidationError{
				{
					Field:   "content",
					Message: "Content cannot be nil",
					Code:    "NIL_CONTENT",
				},
			},
		}, nil
	}

	// Load schema if not already loaded
	if err := sv.LoadSchema(schemaPath); err != nil {
		return &domain.ContentValidation{
			IsValid: false,
			Errors: []domain.ValidationError{
				{
					Field:   "schema",
					Message: fmt.Sprintf("Failed to load schema: %v", err),
					Code:    "SCHEMA_LOAD_ERROR",
				},
			},
		}, nil
	}

	// Get schema for content type
	schema, err := sv.GetSchema(content.Type)
	if err != nil {
		return &domain.ContentValidation{
			IsValid: false,
			Errors: []domain.ValidationError{
				{
					Field:   "schema",
					Message: fmt.Sprintf("No schema found for content type %s: %v", content.Type, err),
					Code:    "MISSING_SCHEMA",
				},
			},
		}, nil
	}

	// Validate content against schema
	validation := &domain.ContentValidation{
		IsValid:     true,
		SchemaValid: true,
	}

	// Convert content to JSON for validation
	contentJSON, err := sv.contentToJSON(content)
	if err != nil {
		validation.Errors = append(validation.Errors, domain.ValidationError{
			Field:   "content",
			Message: fmt.Sprintf("Failed to convert content to JSON: %v", err),
			Code:    "JSON_CONVERSION_ERROR",
		})
		validation.IsValid = false
		validation.SchemaValid = false
		return validation, nil
	}

	// Validate against schema
	if err := sv.validateAgainstSchema(contentJSON, schema); err != nil {
		validation.Errors = append(validation.Errors, domain.ValidationError{
			Field:   "schema",
			Message: fmt.Sprintf("Schema validation failed: %v", err),
			Code:    "SCHEMA_VALIDATION_ERROR",
		})
		validation.IsValid = false
		validation.SchemaValid = false
	}

	return validation, nil
}

// LoadSchema loads a JSON schema from file
func (sv *SchemaValidatorImpl) LoadSchema(schemaPath string) error {
	if schemaPath == "" {
		return fmt.Errorf("schema path cannot be empty")
	}

	// Read schema file
	data, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %v", err)
	}

	// Parse schema
	var schema interface{}
	if err := json.Unmarshal(data, &schema); err != nil {
		return fmt.Errorf("failed to parse schema JSON: %v", err)
	}

	// Determine content type from filename
	contentType := sv.contentTypeFromPath(schemaPath)
	if contentType == "" {
		return fmt.Errorf("unable to determine content type from schema path: %s", schemaPath)
	}

	// Store schema
	sv.schemas[contentType] = schema

	sv.logger.Info("Schema loaded", map[string]interface{}{
		"schema_path":  schemaPath,
		"content_type": contentType,
	})

	return nil
}

// GetSchema retrieves a schema for a content type
func (sv *SchemaValidatorImpl) GetSchema(contentType domain.ContentType) (interface{}, error) {
	if contentType == "" {
		return nil, fmt.Errorf("content type cannot be empty")
	}

	schema, exists := sv.schemas[contentType]
	if !exists {
		return nil, fmt.Errorf("no schema found for content type %s", contentType)
	}

	return schema, nil
}

// ValidateField validates a specific field against a schema
func (sv *SchemaValidatorImpl) ValidateField(field string, value interface{}, schema interface{}) error {
	if field == "" {
		return fmt.Errorf("field name cannot be empty")
	}

	if schema == nil {
		return fmt.Errorf("schema cannot be nil")
	}

	// This is a placeholder implementation
	// In a real implementation, this would validate the field against the schema
	// using a JSON schema validation library

	sv.logger.Debug("Field validation", map[string]interface{}{
		"field": field,
		"value": value,
	})

	return nil
}

// Helper methods

func (sv *SchemaValidatorImpl) contentTypeFromPath(schemaPath string) domain.ContentType {
	filename := filepath.Base(schemaPath)

	// Remove .json extension
	if len(filename) > 5 && filename[len(filename)-5:] == ".json" {
		filename = filename[:len(filename)-5]
	}

	// Map filename to content type
	switch filename {
	case "class":
		return domain.ContentTypeClass
	case "race":
		return domain.ContentTypeRace
	case "item":
		return domain.ContentTypeItem
	case "spell":
		return domain.ContentTypeSpell
	case "ability":
		return domain.ContentTypeAbility
	case "monster":
		return domain.ContentTypeMonster
	case "encounter":
		return domain.ContentTypeEncounter
	case "location":
		return domain.ContentTypeLocation
	case "mechanic":
		return domain.ContentTypeMechanic
	case "action":
		return domain.ContentTypeAction
	case "resolution":
		return domain.ContentTypeResolution
	case "modifier":
		return domain.ContentTypeModifier
	case "randomization":
		return domain.ContentTypeRandomization
	default:
		return ""
	}
}

func (sv *SchemaValidatorImpl) contentToJSON(content *domain.TOMLContent) (interface{}, error) {
	// Convert TOML content to JSON structure
	jsonContent := map[string]interface{}{
		"type":        content.Type,
		"id":          content.ID,
		"version":     content.Version,
		"metadata":    content.Metadata,
		"data":        content.Data,
		"expressions": content.Expressions,
		"references":  content.References,
	}

	return jsonContent, nil
}

func (sv *SchemaValidatorImpl) validateAgainstSchema(content interface{}, schema interface{}) error {
	// This is a placeholder implementation
	// In a real implementation, this would use a JSON schema validation library
	// like github.com/santhosh-tekuri/jsonschema/v5

	sv.logger.Debug("Schema validation", map[string]interface{}{
		"content": content,
		"schema":  schema,
	})

	return nil
}
