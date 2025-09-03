package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
)

// ContentRepositoryManagerImpl implements the ContentRepositoryManager interface
type ContentRepositoryManagerImpl struct {
	repository *domain.ContentRepository
	mu         sync.RWMutex
	logger     domain.Logger
}

// NewContentRepositoryManager creates a new content repository manager
func NewContentRepositoryManager(logger domain.Logger) domain.ContentRepositoryManager {
	return &ContentRepositoryManagerImpl{
		repository: &domain.ContentRepository{
			Content:     make(map[string]*domain.TOMLContent),
			ByType:      make(map[domain.ContentType][]string),
			References:  make(map[string][]string),
			LastUpdated: time.Now(),
			Version:     "1.0.0",
		},
		logger: logger,
	}
}

// AddContent adds content to the repository
func (crm *ContentRepositoryManagerImpl) AddContent(content *domain.TOMLContent) error {
	crm.mu.Lock()
	defer crm.mu.Unlock()

	if content == nil {
		return fmt.Errorf("content cannot be nil")
	}

	if content.ID == "" {
		return fmt.Errorf("content ID cannot be empty")
	}

	// Check if content already exists
	if _, exists := crm.repository.Content[content.ID]; exists {
		return fmt.Errorf("content with ID %s already exists", content.ID)
	}

	// Add content
	crm.repository.Content[content.ID] = content

	// Update type index
	if crm.repository.ByType[content.Type] == nil {
		crm.repository.ByType[content.Type] = make([]string, 0)
	}
	crm.repository.ByType[content.Type] = append(crm.repository.ByType[content.Type], content.ID)

	// Update references index
	for _, ref := range content.References {
		if crm.repository.References[ref] == nil {
			crm.repository.References[ref] = make([]string, 0)
		}
		crm.repository.References[ref] = append(crm.repository.References[ref], content.ID)
	}

	// Update metadata
	crm.repository.LastUpdated = time.Now()

	crm.logger.Info("Content added to repository", map[string]interface{}{
		"content_id":   content.ID,
		"content_type": content.Type,
		"references":   len(content.References),
	})

	return nil
}

// GetContent retrieves content by ID
func (crm *ContentRepositoryManagerImpl) GetContent(id string) (*domain.TOMLContent, error) {
	crm.mu.RLock()
	defer crm.mu.RUnlock()

	if id == "" {
		return nil, fmt.Errorf("content ID cannot be empty")
	}

	content, exists := crm.repository.Content[id]
	if !exists {
		return nil, fmt.Errorf("content with ID %s not found", id)
	}

	return content, nil
}

// GetContentByType retrieves all content of a specific type
func (crm *ContentRepositoryManagerImpl) GetContentByType(contentType domain.ContentType) ([]*domain.TOMLContent, error) {
	crm.mu.RLock()
	defer crm.mu.RUnlock()

	if contentType == "" {
		return nil, fmt.Errorf("content type cannot be empty")
	}

	contentIDs, exists := crm.repository.ByType[contentType]
	if !exists {
		return []*domain.TOMLContent{}, nil
	}

	content := make([]*domain.TOMLContent, 0, len(contentIDs))
	for _, id := range contentIDs {
		if c, exists := crm.repository.Content[id]; exists {
			content = append(content, c)
		}
	}

	return content, nil
}

// RemoveContent removes content from the repository
func (crm *ContentRepositoryManagerImpl) RemoveContent(id string) error {
	crm.mu.Lock()
	defer crm.mu.Unlock()

	if id == "" {
		return fmt.Errorf("content ID cannot be empty")
	}

	content, exists := crm.repository.Content[id]
	if !exists {
		return fmt.Errorf("content with ID %s not found", id)
	}

	// Remove from content map
	delete(crm.repository.Content, id)

	// Remove from type index
	if contentIDs, exists := crm.repository.ByType[content.Type]; exists {
		for i, contentID := range contentIDs {
			if contentID == id {
				crm.repository.ByType[content.Type] = append(contentIDs[:i], contentIDs[i+1:]...)
				break
			}
		}
	}

	// Remove from references index
	for _, ref := range content.References {
		if refs, exists := crm.repository.References[ref]; exists {
			for i, refID := range refs {
				if refID == id {
					crm.repository.References[ref] = append(refs[:i], refs[i+1:]...)
					break
				}
			}
		}
	}

	// Update metadata
	crm.repository.LastUpdated = time.Now()

	crm.logger.Info("Content removed from repository", map[string]interface{}{
		"content_id":   id,
		"content_type": content.Type,
	})

	return nil
}

// UpdateContent updates existing content in the repository
func (crm *ContentRepositoryManagerImpl) UpdateContent(content *domain.TOMLContent) error {
	crm.mu.Lock()
	defer crm.mu.Unlock()

	if content == nil {
		return fmt.Errorf("content cannot be nil")
	}

	if content.ID == "" {
		return fmt.Errorf("content ID cannot be empty")
	}

	// Check if content exists
	existingContent, exists := crm.repository.Content[content.ID]
	if !exists {
		return fmt.Errorf("content with ID %s not found", content.ID)
	}

	// Remove old references
	for _, ref := range existingContent.References {
		if refs, exists := crm.repository.References[ref]; exists {
			for i, refID := range refs {
				if refID == content.ID {
					crm.repository.References[ref] = append(refs[:i], refs[i+1:]...)
					break
				}
			}
		}
	}

	// Update content
	crm.repository.Content[content.ID] = content

	// Update type index if type changed
	if existingContent.Type != content.Type {
		// Remove from old type
		if contentIDs, exists := crm.repository.ByType[existingContent.Type]; exists {
			for i, contentID := range contentIDs {
				if contentID == content.ID {
					crm.repository.ByType[existingContent.Type] = append(contentIDs[:i], contentIDs[i+1:]...)
					break
				}
			}
		}

		// Add to new type
		if crm.repository.ByType[content.Type] == nil {
			crm.repository.ByType[content.Type] = make([]string, 0)
		}
		crm.repository.ByType[content.Type] = append(crm.repository.ByType[content.Type], content.ID)
	}

	// Update references index
	for _, ref := range content.References {
		if crm.repository.References[ref] == nil {
			crm.repository.References[ref] = make([]string, 0)
		}
		crm.repository.References[ref] = append(crm.repository.References[ref], content.ID)
	}

	// Update metadata
	crm.repository.LastUpdated = time.Now()

	crm.logger.Info("Content updated in repository", map[string]interface{}{
		"content_id":   content.ID,
		"content_type": content.Type,
		"references":   len(content.References),
	})

	return nil
}

// ListContent lists all content in the repository
func (crm *ContentRepositoryManagerImpl) ListContent() ([]*domain.TOMLContent, error) {
	crm.mu.RLock()
	defer crm.mu.RUnlock()

	content := make([]*domain.TOMLContent, 0, len(crm.repository.Content))
	for _, c := range crm.repository.Content {
		content = append(content, c)
	}

	return content, nil
}

// ValidateRepository validates the entire repository
func (crm *ContentRepositoryManagerImpl) ValidateRepository() (*domain.ContentValidation, error) {
	crm.mu.RLock()
	defer crm.mu.RUnlock()

	validation := &domain.ContentValidation{
		IsValid:     true,
		SchemaValid: true,
	}

	// Validate all content
	for id, content := range crm.repository.Content {
		if content.ID != id {
			validation.Errors = append(validation.Errors, domain.ValidationError{
				Field:   "id",
				Message: fmt.Sprintf("Content ID mismatch: expected %s, got %s", id, content.ID),
				Code:    "ID_MISMATCH",
			})
			validation.IsValid = false
		}

		if content.Type == "" {
			validation.Errors = append(validation.Errors, domain.ValidationError{
				Field:   "type",
				Message: fmt.Sprintf("Content type is empty for ID %s", id),
				Code:    "MISSING_TYPE",
			})
			validation.IsValid = false
		}
	}

	// Validate type index
	for contentType, contentIDs := range crm.repository.ByType {
		for _, contentID := range contentIDs {
			if content, exists := crm.repository.Content[contentID]; !exists {
				validation.Errors = append(validation.Errors, domain.ValidationError{
					Field:   "type_index",
					Message: fmt.Sprintf("Content ID %s in type index but not in content map", contentID),
					Code:    "MISSING_CONTENT",
				})
				validation.IsValid = false
			} else if content.Type != contentType {
				validation.Errors = append(validation.Errors, domain.ValidationError{
					Field:   "type_index",
					Message: fmt.Sprintf("Content type mismatch for ID %s: expected %s, got %s", contentID, contentType, content.Type),
					Code:    "TYPE_MISMATCH",
				})
				validation.IsValid = false
			}
		}
	}

	// Validate references index
	for ref, contentIDs := range crm.repository.References {
		for _, contentID := range contentIDs {
			if content, exists := crm.repository.Content[contentID]; !exists {
				validation.Errors = append(validation.Errors, domain.ValidationError{
					Field:   "references_index",
					Message: fmt.Sprintf("Content ID %s in references index but not in content map", contentID),
					Code:    "MISSING_CONTENT",
				})
				validation.IsValid = false
			} else {
				// Check if reference is actually in content
				found := false
				for _, contentRef := range content.References {
					if contentRef == ref {
						found = true
						break
					}
				}
				if !found {
					validation.Errors = append(validation.Errors, domain.ValidationError{
						Field:   "references_index",
						Message: fmt.Sprintf("Reference %s in index but not in content %s", ref, contentID),
						Code:    "MISSING_REFERENCE",
					})
					validation.IsValid = false
				}
			}
		}
	}

	return validation, nil
}

// ResolveReferences resolves cross-references for content
func (crm *ContentRepositoryManagerImpl) ResolveReferences(content *domain.TOMLContent) error {
	crm.mu.RLock()
	defer crm.mu.RUnlock()

	if content == nil {
		return fmt.Errorf("content cannot be nil")
	}

	// This is a placeholder implementation
	// In a real implementation, this would resolve references to other content
	// and validate that referenced content exists

	for _, ref := range content.References {
		if _, exists := crm.repository.Content[ref]; !exists {
			crm.logger.Warn("Reference not found", map[string]interface{}{
				"content_id": content.ID,
				"reference":  ref,
			})
		}
	}

	return nil
}

// GetRepository returns the current repository state
func (crm *ContentRepositoryManagerImpl) GetRepository() *domain.ContentRepository {
	crm.mu.RLock()
	defer crm.mu.RUnlock()

	// Return a copy to prevent external modification
	repo := &domain.ContentRepository{
		Content:     make(map[string]*domain.TOMLContent),
		ByType:      make(map[domain.ContentType][]string),
		References:  make(map[string][]string),
		LastUpdated: crm.repository.LastUpdated,
		Version:     crm.repository.Version,
	}

	// Copy content
	for id, content := range crm.repository.Content {
		repo.Content[id] = content
	}

	// Copy type index
	for contentType, contentIDs := range crm.repository.ByType {
		repo.ByType[contentType] = make([]string, len(contentIDs))
		copy(repo.ByType[contentType], contentIDs)
	}

	// Copy references index
	for ref, contentIDs := range crm.repository.References {
		repo.References[ref] = make([]string, len(contentIDs))
		copy(repo.References[ref], contentIDs)
	}

	return repo
}
