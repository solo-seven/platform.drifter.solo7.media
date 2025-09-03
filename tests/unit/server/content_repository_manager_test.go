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

func TestContentRepositoryManager_AddContent(t *testing.T) {
	mockLogger := &MockLogger{}
	manager := server.NewContentRepositoryManager(mockLogger)

	t.Run("successful addition", func(t *testing.T) {
		content := &domain.TOMLContent{
			ID:   "class.fighter",
			Type: domain.ContentTypeClass,
			Data: map[string]interface{}{
				"name": "Fighter",
			},
			References: []string{"item.iron_sword"},
		}

		mockLogger.On("Info", "Content added to repository", mock.AnythingOfType("map[string]interface {}")).Return()

		err := manager.AddContent(content)

		assert.NoError(t, err)

		// Verify content was added
		retrieved, err := manager.GetContent("class.fighter")
		assert.NoError(t, err)
		assert.Equal(t, content, retrieved)

		mockLogger.AssertExpectations(t)
	})

	t.Run("nil content", func(t *testing.T) {
		err := manager.AddContent(nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "content cannot be nil")
	})

	t.Run("empty ID", func(t *testing.T) {
		content := &domain.TOMLContent{
			Type: domain.ContentTypeClass,
			Data: map[string]interface{}{
				"name": "Fighter",
			},
		}

		err := manager.AddContent(content)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "content ID cannot be empty")
	})

	t.Run("duplicate content", func(t *testing.T) {
		// Create a fresh manager for this test
		freshManager := server.NewContentRepositoryManager(mockLogger)
		content := &domain.TOMLContent{
			ID:   "class.fighter",
			Type: domain.ContentTypeClass,
			Data: map[string]interface{}{
				"name": "Fighter",
			},
		}

		mockLogger.On("Info", "Content added to repository", mock.AnythingOfType("map[string]interface {}")).Return()

		// Add first time
		err := freshManager.AddContent(content)
		assert.NoError(t, err)

		// Try to add again - this should fail
		err = freshManager.AddContent(content)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")

		mockLogger.AssertExpectations(t)
	})
}

func TestContentRepositoryManager_GetContent(t *testing.T) {
	mockLogger := &MockLogger{}
	manager := server.NewContentRepositoryManager(mockLogger)

	t.Run("successful retrieval", func(t *testing.T) {
		content := &domain.TOMLContent{
			ID:   "class.fighter",
			Type: domain.ContentTypeClass,
			Data: map[string]interface{}{
				"name": "Fighter",
			},
		}

		mockLogger.On("Info", "Content added to repository", mock.AnythingOfType("map[string]interface {}")).Return()

		// Add content
		err := manager.AddContent(content)
		assert.NoError(t, err)

		// Retrieve content
		retrieved, err := manager.GetContent("class.fighter")
		assert.NoError(t, err)
		assert.Equal(t, content, retrieved)

		mockLogger.AssertExpectations(t)
	})

	t.Run("empty ID", func(t *testing.T) {
		_, err := manager.GetContent("")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "content ID cannot be empty")
	})

	t.Run("content not found", func(t *testing.T) {
		_, err := manager.GetContent("nonexistent.id")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestContentRepositoryManager_GetContentByType(t *testing.T) {
	mockLogger := &MockLogger{}
	manager := server.NewContentRepositoryManager(mockLogger)

	t.Run("successful retrieval by type", func(t *testing.T) {
		content1 := &domain.TOMLContent{
			ID:   "class.fighter",
			Type: domain.ContentTypeClass,
			Data: map[string]interface{}{
				"name": "Fighter",
			},
		}

		content2 := &domain.TOMLContent{
			ID:   "class.wizard",
			Type: domain.ContentTypeClass,
			Data: map[string]interface{}{
				"name": "Wizard",
			},
		}

		content3 := &domain.TOMLContent{
			ID:   "race.human",
			Type: domain.ContentTypeRace,
			Data: map[string]interface{}{
				"name": "Human",
			},
		}

		mockLogger.On("Info", "Content added to repository", mock.AnythingOfType("map[string]interface {}")).Return().Times(3)

		// Add content
		err := manager.AddContent(content1)
		assert.NoError(t, err)
		err = manager.AddContent(content2)
		assert.NoError(t, err)
		err = manager.AddContent(content3)
		assert.NoError(t, err)

		// Retrieve by type
		classes, err := manager.GetContentByType(domain.ContentTypeClass)
		assert.NoError(t, err)
		assert.Len(t, classes, 2)

		races, err := manager.GetContentByType(domain.ContentTypeRace)
		assert.NoError(t, err)
		assert.Len(t, races, 1)

		mockLogger.AssertExpectations(t)
	})

	t.Run("empty content type", func(t *testing.T) {
		_, err := manager.GetContentByType("")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "content type cannot be empty")
	})

	t.Run("no content of type", func(t *testing.T) {
		content, err := manager.GetContentByType(domain.ContentTypeSpell)

		assert.NoError(t, err)
		assert.Empty(t, content)
	})
}

func TestContentRepositoryManager_RemoveContent(t *testing.T) {
	mockLogger := &MockLogger{}
	manager := server.NewContentRepositoryManager(mockLogger)

	t.Run("successful removal", func(t *testing.T) {
		content := &domain.TOMLContent{
			ID:   "class.fighter",
			Type: domain.ContentTypeClass,
			Data: map[string]interface{}{
				"name": "Fighter",
			},
			References: []string{"item.iron_sword"},
		}

		mockLogger.On("Info", "Content added to repository", mock.AnythingOfType("map[string]interface {}")).Return()
		mockLogger.On("Info", "Content removed from repository", mock.AnythingOfType("map[string]interface {}")).Return()

		// Add content
		err := manager.AddContent(content)
		assert.NoError(t, err)

		// Remove content
		err = manager.RemoveContent("class.fighter")
		assert.NoError(t, err)

		// Verify content was removed
		_, err = manager.GetContent("class.fighter")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		mockLogger.AssertExpectations(t)
	})

	t.Run("empty ID", func(t *testing.T) {
		err := manager.RemoveContent("")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "content ID cannot be empty")
	})

	t.Run("content not found", func(t *testing.T) {
		err := manager.RemoveContent("nonexistent.id")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestContentRepositoryManager_UpdateContent(t *testing.T) {
	mockLogger := &MockLogger{}
	manager := server.NewContentRepositoryManager(mockLogger)

	t.Run("successful update", func(t *testing.T) {
		originalContent := &domain.TOMLContent{
			ID:   "class.fighter",
			Type: domain.ContentTypeClass,
			Data: map[string]interface{}{
				"name": "Fighter",
			},
			References: []string{"item.iron_sword"},
		}

		updatedContent := &domain.TOMLContent{
			ID:   "class.fighter",
			Type: domain.ContentTypeClass,
			Data: map[string]interface{}{
				"name":    "Fighter",
				"hit_die": "d10",
			},
			References: []string{"item.iron_sword", "item.plate_armor"},
		}

		mockLogger.On("Info", "Content added to repository", mock.AnythingOfType("map[string]interface {}")).Return()
		mockLogger.On("Info", "Content updated in repository", mock.AnythingOfType("map[string]interface {}")).Return()

		// Add original content
		err := manager.AddContent(originalContent)
		assert.NoError(t, err)

		// Update content
		err = manager.UpdateContent(updatedContent)
		assert.NoError(t, err)

		// Verify content was updated
		retrieved, err := manager.GetContent("class.fighter")
		assert.NoError(t, err)
		assert.Equal(t, updatedContent, retrieved)

		mockLogger.AssertExpectations(t)
	})

	t.Run("nil content", func(t *testing.T) {
		err := manager.UpdateContent(nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "content cannot be nil")
	})

	t.Run("empty ID", func(t *testing.T) {
		content := &domain.TOMLContent{
			Type: domain.ContentTypeClass,
			Data: map[string]interface{}{
				"name": "Fighter",
			},
		}

		err := manager.UpdateContent(content)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "content ID cannot be empty")
	})

	t.Run("content not found", func(t *testing.T) {
		content := &domain.TOMLContent{
			ID:   "nonexistent.id",
			Type: domain.ContentTypeClass,
			Data: map[string]interface{}{
				"name": "Fighter",
			},
		}

		err := manager.UpdateContent(content)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestContentRepositoryManager_ListContent(t *testing.T) {
	mockLogger := &MockLogger{}
	manager := server.NewContentRepositoryManager(mockLogger)

	t.Run("list all content", func(t *testing.T) {
		content1 := &domain.TOMLContent{
			ID:   "class.fighter",
			Type: domain.ContentTypeClass,
			Data: map[string]interface{}{
				"name": "Fighter",
			},
		}

		content2 := &domain.TOMLContent{
			ID:   "race.human",
			Type: domain.ContentTypeRace,
			Data: map[string]interface{}{
				"name": "Human",
			},
		}

		mockLogger.On("Info", "Content added to repository", mock.AnythingOfType("map[string]interface {}")).Return().Times(2)

		// Add content
		err := manager.AddContent(content1)
		assert.NoError(t, err)
		err = manager.AddContent(content2)
		assert.NoError(t, err)

		// List content
		content, err := manager.ListContent()
		assert.NoError(t, err)
		assert.Len(t, content, 2)

		mockLogger.AssertExpectations(t)
	})

	t.Run("empty repository", func(t *testing.T) {
		// Create a fresh manager for this test
		freshManager := server.NewContentRepositoryManager(mockLogger)
		content, err := freshManager.ListContent()
		assert.NoError(t, err)
		assert.Empty(t, content)
	})
}

func TestContentRepositoryManager_ValidateRepository(t *testing.T) {
	mockLogger := &MockLogger{}
	manager := server.NewContentRepositoryManager(mockLogger)

	t.Run("valid repository", func(t *testing.T) {
		content := &domain.TOMLContent{
			ID:   "class.fighter",
			Type: domain.ContentTypeClass,
			Data: map[string]interface{}{
				"name": "Fighter",
			},
		}

		mockLogger.On("Info", "Content added to repository", mock.AnythingOfType("map[string]interface {}")).Return()

		// Add content
		err := manager.AddContent(content)
		assert.NoError(t, err)

		// Validate repository
		validation, err := manager.ValidateRepository()
		assert.NoError(t, err)
		assert.NotNil(t, validation)
		assert.True(t, validation.IsValid)

		mockLogger.AssertExpectations(t)
	})

	t.Run("empty repository", func(t *testing.T) {
		validation, err := manager.ValidateRepository()
		assert.NoError(t, err)
		assert.NotNil(t, validation)
		assert.True(t, validation.IsValid)
	})
}

func TestContentRepositoryManager_GetRepository(t *testing.T) {
	mockLogger := &MockLogger{}
	manager := server.NewContentRepositoryManager(mockLogger)

	t.Run("get repository", func(t *testing.T) {
		content := &domain.TOMLContent{
			ID:   "class.fighter",
			Type: domain.ContentTypeClass,
			Data: map[string]interface{}{
				"name": "Fighter",
			},
		}

		mockLogger.On("Info", "Content added to repository", mock.AnythingOfType("map[string]interface {}")).Return()

		// Add content
		err := manager.AddContent(content)
		assert.NoError(t, err)

		// Get repository
		repository := manager.GetRepository()
		assert.NotNil(t, repository)
		assert.Len(t, repository.Content, 1)
		assert.Contains(t, repository.Content, "class.fighter")

		mockLogger.AssertExpectations(t)
	})
}
