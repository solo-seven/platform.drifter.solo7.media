package server_test

import (
	"testing"

	"github.com/solo7.media/platform.drifter.solo7.media/internal/domain"
	"github.com/solo7.media/platform.drifter.solo7.media/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLogger is a mock implementation of the Logger interface
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, fields map[string]interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields map[string]interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields map[string]interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields map[string]interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Fatal(msg string, fields map[string]interface{}) {
	m.Called(msg, fields)
}

func TestReferenceResolver_ResolveReferences(t *testing.T) {
	mockLogger := &MockLogger{}
	resolver := server.NewReferenceResolver(mockLogger)

	t.Run("successful reference resolution", func(t *testing.T) {
		content := &domain.TOMLContent{
			ID:   "class.fighter",
			Type: domain.ContentTypeClass,
			Data: map[string]interface{}{
				"weapon": "item.iron_sword",
				"armor":  "item.leather_armor",
			},
			Expressions: map[string]string{
				"effect": "heal(self, roll(\"1d10\") + self.level)",
			},
		}

		repository := &domain.ContentRepository{
			Content: map[string]*domain.TOMLContent{
				"item.iron_sword": {
					ID:   "item.iron_sword",
					Type: domain.ContentTypeItem,
				},
				"item.leather_armor": {
					ID:   "item.leather_armor",
					Type: domain.ContentTypeItem,
				},
			},
		}

		mockLogger.On("Debug", "References validated", mock.AnythingOfType("map[string]interface {}")).Return()
		mockLogger.On("Info", "References resolved", mock.AnythingOfType("map[string]interface {}")).Return()

		err := resolver.ResolveReferences(content, repository)

		assert.NoError(t, err)
		assert.NotNil(t, content.References)
		assert.Contains(t, content.References, "item.iron_sword")
		assert.Contains(t, content.References, "item.leather_armor")

		mockLogger.AssertExpectations(t)
	})

	t.Run("nil content", func(t *testing.T) {
		repository := &domain.ContentRepository{}

		err := resolver.ResolveReferences(nil, repository)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "content cannot be nil")
	})

	t.Run("nil repository", func(t *testing.T) {
		content := &domain.TOMLContent{
			ID:   "class.fighter",
			Type: domain.ContentTypeClass,
		}

		err := resolver.ResolveReferences(content, nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository cannot be nil")
	})

	t.Run("missing references", func(t *testing.T) {
		content := &domain.TOMLContent{
			ID:   "class.fighter",
			Type: domain.ContentTypeClass,
			Data: map[string]interface{}{
				"weapon": "item.iron_sword",
				"armor":  "item.nonexistent_armor",
			},
		}

		repository := &domain.ContentRepository{
			Content: map[string]*domain.TOMLContent{
				"item.iron_sword": {
					ID:   "item.iron_sword",
					Type: domain.ContentTypeItem,
				},
			},
		}

		mockLogger.On("Debug", "References validated", mock.AnythingOfType("map[string]interface {}")).Return()
		mockLogger.On("Info", "References resolved", mock.AnythingOfType("map[string]interface {}")).Return()

		err := resolver.ResolveReferences(content, repository)

		assert.NoError(t, err)
		assert.NotNil(t, content.References)
		assert.Contains(t, content.References, "item.iron_sword")
		assert.Contains(t, content.References, "item.nonexistent_armor")

		mockLogger.AssertExpectations(t)
	})
}

func TestReferenceResolver_ValidateReferences(t *testing.T) {
	mockLogger := &MockLogger{}
	resolver := server.NewReferenceResolver(mockLogger)

	t.Run("validate valid references", func(t *testing.T) {
		references := []string{
			"item.iron_sword",
			"item.leather_armor",
			"spell.heal",
		}

		repository := &domain.ContentRepository{
			Content: map[string]*domain.TOMLContent{
				"item.iron_sword": {
					ID:   "item.iron_sword",
					Type: domain.ContentTypeItem,
				},
				"item.leather_armor": {
					ID:   "item.leather_armor",
					Type: domain.ContentTypeItem,
				},
				"spell.heal": {
					ID:   "spell.heal",
					Type: domain.ContentTypeSpell,
				},
			},
		}

		mockLogger.On("Debug", "References validated", mock.AnythingOfType("map[string]interface {}")).Return()

		validation, err := resolver.ValidateReferences(references, repository)

		assert.NoError(t, err)
		assert.NotNil(t, validation)
		assert.Len(t, validation.ValidReferences, 3)
		assert.Empty(t, validation.InvalidReferences)
		assert.Empty(t, validation.MissingReferences)

		mockLogger.AssertExpectations(t)
	})

	t.Run("validate mixed references", func(t *testing.T) {
		references := []string{
			"item.iron_sword",
			"item.nonexistent_armor",
			"invalid-reference",
			"",
		}

		repository := &domain.ContentRepository{
			Content: map[string]*domain.TOMLContent{
				"item.iron_sword": {
					ID:   "item.iron_sword",
					Type: domain.ContentTypeItem,
				},
			},
		}

		mockLogger.On("Debug", "References validated", mock.AnythingOfType("map[string]interface {}")).Return()

		validation, err := resolver.ValidateReferences(references, repository)

		assert.NoError(t, err)
		assert.NotNil(t, validation)
		assert.Len(t, validation.ValidReferences, 1)
		assert.Len(t, validation.InvalidReferences, 1)
		assert.Len(t, validation.MissingReferences, 1)

		mockLogger.AssertExpectations(t)
	})

	t.Run("nil repository", func(t *testing.T) {
		references := []string{"item.iron_sword"}

		validation, err := resolver.ValidateReferences(references, nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository cannot be nil")
		assert.Nil(t, validation)
	})

	t.Run("empty references", func(t *testing.T) {
		repository := &domain.ContentRepository{}

		mockLogger.On("Debug", "References validated", mock.AnythingOfType("map[string]interface {}")).Return()

		validation, err := resolver.ValidateReferences([]string{}, repository)

		assert.NoError(t, err)
		assert.NotNil(t, validation)
		assert.Empty(t, validation.ValidReferences)
		assert.Empty(t, validation.InvalidReferences)
		assert.Empty(t, validation.MissingReferences)

		mockLogger.AssertExpectations(t)
	})
}

func TestReferenceResolver_FindMissingReferences(t *testing.T) {
	mockLogger := &MockLogger{}
	resolver := server.NewReferenceResolver(mockLogger)

	t.Run("find missing references", func(t *testing.T) {
		content := &domain.TOMLContent{
			ID:   "class.fighter",
			Type: domain.ContentTypeClass,
			Data: map[string]interface{}{
				"weapon": "item.iron_sword",
				"armor":  "item.nonexistent_armor",
			},
			Expressions: map[string]string{
				"effect": "heal(self, roll(\"1d10\") + self.level)",
			},
		}

		repository := &domain.ContentRepository{
			Content: map[string]*domain.TOMLContent{
				"item.iron_sword": {
					ID:   "item.iron_sword",
					Type: domain.ContentTypeItem,
				},
			},
		}

		mockLogger.On("Debug", "Missing references found", mock.AnythingOfType("map[string]interface {}")).Return()

		missing, err := resolver.FindMissingReferences(content, repository)

		assert.NoError(t, err)
		assert.NotNil(t, missing)
		assert.Contains(t, missing, "item.nonexistent_armor")

		mockLogger.AssertExpectations(t)
	})

	t.Run("no missing references", func(t *testing.T) {
		content := &domain.TOMLContent{
			ID:   "class.fighter",
			Type: domain.ContentTypeClass,
			Data: map[string]interface{}{
				"weapon": "item.iron_sword",
			},
		}

		repository := &domain.ContentRepository{
			Content: map[string]*domain.TOMLContent{
				"item.iron_sword": {
					ID:   "item.iron_sword",
					Type: domain.ContentTypeItem,
				},
			},
		}

		mockLogger.On("Debug", "Missing references found", mock.AnythingOfType("map[string]interface {}")).Return()

		missing, err := resolver.FindMissingReferences(content, repository)

		assert.NoError(t, err)
		assert.NotNil(t, missing)
		assert.Empty(t, missing)

		mockLogger.AssertExpectations(t)
	})

	t.Run("nil content", func(t *testing.T) {
		repository := &domain.ContentRepository{}

		missing, err := resolver.FindMissingReferences(nil, repository)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "content cannot be nil")
		assert.Nil(t, missing)
	})

	t.Run("nil repository", func(t *testing.T) {
		content := &domain.TOMLContent{
			ID:   "class.fighter",
			Type: domain.ContentTypeClass,
		}

		missing, err := resolver.FindMissingReferences(content, nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository cannot be nil")
		assert.Nil(t, missing)
	})
}
