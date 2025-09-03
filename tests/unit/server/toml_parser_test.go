package server_test

import (
	"testing"

	"github.com/solo7.media/platform.drifter.solo7.media/internal/domain"
	"github.com/solo7.media/platform.drifter.solo7.media/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTOMLParser_ParseTOML(t *testing.T) {
	mockLogger := &MockLogger{}
	parser := server.NewTOMLParser(mockLogger)

	t.Run("successful parsing", func(t *testing.T) {
		tomlData := []byte(`
id = "class.fighter"
name = "Fighter"
type = "class"
hit_die = "d10"
primary_attributes = ["str", "con"]

[abilities.second_wind]
name = "Second Wind"
uses = { per = "short_rest", count = 1 }
effect = 'heal(self, roll("1d10") + self.level)'
`)

		mockLogger.On("Debug", "Parsing TOML data", mock.AnythingOfType("map[string]interface {}")).Return()
		mockLogger.On("Info", "Successfully parsed TOML content", mock.AnythingOfType("map[string]interface {}")).Return()

		options := domain.TOMLParseOptions{
			ValidateSchema:    false,
			ResolveReferences: false,
		}

		result, err := parser.ParseTOML(tomlData, options)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.NotNil(t, result.Content)
		assert.Equal(t, "class.fighter", result.Content.ID)
		assert.Equal(t, domain.ContentTypeClass, result.Content.Type)
		assert.Equal(t, "1.0.0", result.Content.Version)

		mockLogger.AssertExpectations(t)
	})

	t.Run("invalid TOML syntax", func(t *testing.T) {
		tomlData := []byte(`
id = "class.fighter"
name = "Fighter"
type = "class"
hit_die = "d10"
primary_attributes = ["str", "con"
`)

		mockLogger.On("Debug", "Parsing TOML data", mock.AnythingOfType("map[string]interface {}")).Return()

		options := domain.TOMLParseOptions{
			ValidateSchema:    false,
			ResolveReferences: false,
		}

		result, err := parser.ParseTOML(tomlData, options)

		assert.NoError(t, err)
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "Failed to parse TOML")

		mockLogger.AssertExpectations(t)
	})

	t.Run("missing ID field", func(t *testing.T) {
		tomlData := []byte(`
name = "Fighter"
type = "class"
hit_die = "d10"
`)

		mockLogger.On("Debug", "Parsing TOML data", mock.AnythingOfType("map[string]interface {}")).Return()

		options := domain.TOMLParseOptions{
			ValidateSchema:    false,
			ResolveReferences: false,
		}

		result, err := parser.ParseTOML(tomlData, options)

		assert.NoError(t, err)
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "Failed to extract ID")

		mockLogger.AssertExpectations(t)
	})

	t.Run("content type not allowed", func(t *testing.T) {
		tomlData := []byte(`
id = "class.fighter"
name = "Fighter"
type = "class"
hit_die = "d10"
`)

		mockLogger.On("Debug", "Parsing TOML data", mock.AnythingOfType("map[string]interface {}")).Return()

		options := domain.TOMLParseOptions{
			ValidateSchema:    false,
			ResolveReferences: false,
			AllowedTypes:      []domain.ContentType{domain.ContentTypeSpecies},
		}

		result, err := parser.ParseTOML(tomlData, options)

		assert.NoError(t, err)
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "Content type class is not allowed")

		mockLogger.AssertExpectations(t)
	})

	t.Run("with validation", func(t *testing.T) {
		tomlData := []byte(`
id = "class.fighter"
name = "Fighter"
type = "class"
hit_die = "d10"
`)

		mockLogger.On("Debug", "Parsing TOML data", mock.AnythingOfType("map[string]interface {}")).Return()
		mockLogger.On("Info", "Successfully parsed TOML content", mock.AnythingOfType("map[string]interface {}")).Return()

		options := domain.TOMLParseOptions{
			ValidateSchema:    true,
			ResolveReferences: false,
		}

		result, err := parser.ParseTOML(tomlData, options)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.NotNil(t, result.Content)

		mockLogger.AssertExpectations(t)
	})
}

func TestTOMLParser_ExtractExpressions(t *testing.T) {
	mockLogger := &MockLogger{}
	parser := server.NewTOMLParser(mockLogger)

	t.Run("extract expressions from content", func(t *testing.T) {
		content := &domain.TOMLContent{
			Data: map[string]interface{}{
				"effect": "heal(self, roll(\"1d10\") + self.level)",
				"damage": "roll(\"2d6\") + 3",
				"name":   "Basic Attack",
			},
		}

		expressions, err := parser.ExtractExpressions(content)

		assert.NoError(t, err)
		assert.NotNil(t, expressions)
		assert.Contains(t, expressions, "effect")
		assert.Contains(t, expressions, "damage")
		assert.Equal(t, "heal(self, roll(\"1d10\") + self.level)", expressions["effect"])
		assert.Equal(t, "roll(\"2d6\") + 3", expressions["damage"])
	})

	t.Run("extract expressions from nested data", func(t *testing.T) {
		content := &domain.TOMLContent{
			Data: map[string]interface{}{
				"abilities": map[string]interface{}{
					"second_wind": map[string]interface{}{
						"effect": "heal(self, roll(\"1d10\") + self.level)",
					},
				},
			},
		}

		expressions, err := parser.ExtractExpressions(content)

		assert.NoError(t, err)
		assert.NotNil(t, expressions)
		assert.Contains(t, expressions, "abilities.second_wind.effect")
	})

	t.Run("no expressions found", func(t *testing.T) {
		content := &domain.TOMLContent{
			Data: map[string]interface{}{
				"name":        "Basic Attack",
				"description": "A simple attack",
				"damage":      "1d8",
			},
		}

		expressions, err := parser.ExtractExpressions(content)

		assert.NoError(t, err)
		assert.NotNil(t, expressions)
		// "1d8" should be detected as an expression due to dice notation
		assert.Contains(t, expressions, "damage")
	})

	t.Run("nil content", func(t *testing.T) {
		expressions, err := parser.ExtractExpressions(nil)

		assert.NoError(t, err)
		assert.NotNil(t, expressions)
		assert.Empty(t, expressions)
	})
}

func TestTOMLParser_ExtractReferences(t *testing.T) {
	mockLogger := &MockLogger{}
	parser := server.NewTOMLParser(mockLogger)

	t.Run("extract references from content", func(t *testing.T) {
		content := &domain.TOMLContent{
			Data: map[string]interface{}{
				"weapon":      "item.iron_sword",
				"armor":       "item.leather_armor",
				"description": "Uses @Item/iron_sword for combat",
			},
		}

		references, err := parser.ExtractReferences(content)

		assert.NoError(t, err)
		assert.NotNil(t, references)

		// The current implementation should detect these as references
		assert.Contains(t, references, "item.iron_sword")
		assert.Contains(t, references, "item.leather_armor")
	})

	t.Run("extract references from nested data", func(t *testing.T) {
		content := &domain.TOMLContent{
			Data: map[string]interface{}{
				"equipment": map[string]interface{}{
					"weapon": "item.iron_sword",
					"armor":  "item.leather_armor",
				},
			},
		}

		references, err := parser.ExtractReferences(content)

		assert.NoError(t, err)
		assert.NotNil(t, references)
		// The current implementation should detect these as references
		assert.Contains(t, references, "item.iron_sword")
		assert.Contains(t, references, "item.leather_armor")
	})

	t.Run("no references found", func(t *testing.T) {
		content := &domain.TOMLContent{
			Data: map[string]interface{}{
				"name":        "Basic Attack",
				"description": "A simple attack",
				"damage":      "1d8",
			},
		}

		references, err := parser.ExtractReferences(content)

		assert.NoError(t, err)
		assert.NotNil(t, references)
		assert.Empty(t, references)
	})

	t.Run("nil content", func(t *testing.T) {
		references, err := parser.ExtractReferences(nil)

		assert.NoError(t, err)
		assert.NotNil(t, references)
		assert.Empty(t, references)
	})
}

func TestTOMLParser_ValidateContent(t *testing.T) {
	mockLogger := &MockLogger{}
	parser := server.NewTOMLParser(mockLogger)

	t.Run("valid content", func(t *testing.T) {
		content := &domain.TOMLContent{
			ID:   "class.fighter",
			Type: domain.ContentTypeClass,
			Data: map[string]interface{}{
				"name": "Fighter",
			},
		}

		options := domain.TOMLParseOptions{
			ValidateSchema: false,
		}

		validation, err := parser.ValidateContent(content, options)

		assert.NoError(t, err)
		assert.NotNil(t, validation)
		assert.True(t, validation.IsValid)
		assert.True(t, validation.SchemaValid)
	})

	t.Run("missing ID", func(t *testing.T) {
		content := &domain.TOMLContent{
			Type: domain.ContentTypeClass,
			Data: map[string]interface{}{
				"name": "Fighter",
			},
		}

		options := domain.TOMLParseOptions{
			ValidateSchema: false,
		}

		validation, err := parser.ValidateContent(content, options)

		assert.NoError(t, err)
		assert.NotNil(t, validation)
		assert.False(t, validation.IsValid)
		assert.Contains(t, validation.Errors[0].Message, "ID is required")
	})

	t.Run("missing type", func(t *testing.T) {
		content := &domain.TOMLContent{
			ID: "class.fighter",
			Data: map[string]interface{}{
				"name": "Fighter",
			},
		}

		options := domain.TOMLParseOptions{
			ValidateSchema: false,
		}

		validation, err := parser.ValidateContent(content, options)

		assert.NoError(t, err)
		assert.NotNil(t, validation)
		assert.False(t, validation.IsValid)
		assert.Contains(t, validation.Errors[0].Message, "Type is required")
	})

	t.Run("invalid expression", func(t *testing.T) {
		content := &domain.TOMLContent{
			ID:   "class.fighter",
			Type: domain.ContentTypeClass,
			Data: map[string]interface{}{
				"name": "Fighter",
			},
			Expressions: map[string]string{
				"effect": "heal(self, roll(\"1d10\") + self.level", // Missing closing parenthesis
			},
		}

		options := domain.TOMLParseOptions{
			ValidateSchema: false,
		}

		validation, err := parser.ValidateContent(content, options)

		assert.NoError(t, err)
		assert.NotNil(t, validation)
		assert.False(t, validation.IsValid)
		assert.Contains(t, validation.Errors[0].Message, "Invalid expression")
	})
}
