package server

import (
	"fmt"
	"regexp"

	"github.com/solo7.media/platform.drifter.solo7.media/internal/domain"
)

// ReferenceResolverImpl implements the ReferenceResolver interface
type ReferenceResolverImpl struct {
	logger domain.Logger
}

// NewReferenceResolver creates a new reference resolver
func NewReferenceResolver(logger domain.Logger) domain.ReferenceResolver {
	return &ReferenceResolverImpl{
		logger: logger,
	}
}

// ResolveReferences resolves cross-references for content
func (rr *ReferenceResolverImpl) ResolveReferences(content *domain.TOMLContent, repository *domain.ContentRepository) error {
	if content == nil {
		return fmt.Errorf("content cannot be nil")
	}

	if repository == nil {
		return fmt.Errorf("repository cannot be nil")
	}

	// Extract references from content
	references, err := rr.extractReferencesFromContent(content)
	if err != nil {
		return fmt.Errorf("failed to extract references: %v", err)
	}

	// Validate references
	refValidation, err := rr.ValidateReferences(references, repository)
	if err != nil {
		return fmt.Errorf("failed to validate references: %v", err)
	}

	// Update content with resolved references
	content.References = references

	// Log resolution results
	rr.logger.Info("References resolved", map[string]interface{}{
		"content_id":         content.ID,
		"total_references":   len(references),
		"valid_references":   len(refValidation.ValidReferences),
		"invalid_references": len(refValidation.InvalidReferences),
		"missing_references": len(refValidation.MissingReferences),
	})

	return nil
}

// ValidateReferences validates a list of references against a repository
func (rr *ReferenceResolverImpl) ValidateReferences(references []string, repository *domain.ContentRepository) (*domain.ReferenceValidation, error) {
	if repository == nil {
		return nil, fmt.Errorf("repository cannot be nil")
	}

	validation := &domain.ReferenceValidation{
		ValidReferences:   make([]string, 0),
		InvalidReferences: make([]string, 0),
		MissingReferences: make([]string, 0),
	}

	for _, ref := range references {
		if ref == "" {
			continue
		}

		// Check if reference is valid format
		if !rr.isValidReferenceFormat(ref) {
			validation.InvalidReferences = append(validation.InvalidReferences, ref)
			continue
		}

		// Check if reference exists in repository
		if _, exists := repository.Content[ref]; exists {
			validation.ValidReferences = append(validation.ValidReferences, ref)
		} else {
			validation.MissingReferences = append(validation.MissingReferences, ref)
		}
	}

	rr.logger.Debug("References validated", map[string]interface{}{
		"total_references":   len(references),
		"valid_references":   len(validation.ValidReferences),
		"invalid_references": len(validation.InvalidReferences),
		"missing_references": len(validation.MissingReferences),
	})

	return validation, nil
}

// FindMissingReferences finds references that are missing from the repository
func (rr *ReferenceResolverImpl) FindMissingReferences(content *domain.TOMLContent, repository *domain.ContentRepository) ([]string, error) {
	if content == nil {
		return nil, fmt.Errorf("content cannot be nil")
	}

	if repository == nil {
		return nil, fmt.Errorf("repository cannot be nil")
	}

	// Extract references from content
	references, err := rr.extractReferencesFromContent(content)
	if err != nil {
		return nil, fmt.Errorf("failed to extract references: %v", err)
	}

	// Find missing references
	missing := make([]string, 0)
	for _, ref := range references {
		if _, exists := repository.Content[ref]; !exists {
			missing = append(missing, ref)
		}
	}

	rr.logger.Debug("Missing references found", map[string]interface{}{
		"content_id":         content.ID,
		"missing_references": len(missing),
	})

	return missing, nil
}

// Helper methods

func (rr *ReferenceResolverImpl) extractReferencesFromContent(content *domain.TOMLContent) ([]string, error) {
	references := make([]string, 0)

	// Extract from existing references
	references = append(references, content.References...)

	// Extract from data fields
	if content.Data != nil {
		dataRefs, err := rr.extractReferencesFromData(content.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to extract references from data: %v", err)
		}
		references = append(references, dataRefs...)
	}

	// Extract from expressions
	if content.Expressions != nil {
		exprRefs, err := rr.extractReferencesFromExpressions(content.Expressions)
		if err != nil {
			return nil, fmt.Errorf("failed to extract references from expressions: %v", err)
		}
		references = append(references, exprRefs...)
	}

	// Remove duplicates
	references = rr.removeDuplicateReferences(references)

	return references, nil
}

func (rr *ReferenceResolverImpl) extractReferencesFromData(data map[string]interface{}) ([]string, error) {
	references := make([]string, 0)

	// Recursively search for reference patterns
	rr.extractReferencesRecursive(data, &references)

	return references, nil
}

func (rr *ReferenceResolverImpl) extractReferencesFromExpressions(expressions map[string]string) ([]string, error) {
	references := make([]string, 0)

	for _, expression := range expressions {
		exprRefs := rr.extractReferencesFromExpression(expression)
		references = append(references, exprRefs...)
	}

	return references, nil
}

func (rr *ReferenceResolverImpl) extractReferencesRecursive(data interface{}, references *[]string) {
	switch v := data.(type) {
	case map[string]interface{}:
		for _, value := range v {
			rr.extractReferencesRecursive(value, references)
		}
	case []interface{}:
		for _, value := range v {
			rr.extractReferencesRecursive(value, references)
		}
	case string:
		// Check if this looks like a reference
		if rr.isReferenceString(v) {
			*references = append(*references, v)
		}
	}
}

func (rr *ReferenceResolverImpl) extractReferencesFromExpression(expression string) []string {
	references := make([]string, 0)

	// Find @reference patterns
	atRefRegex := regexp.MustCompile(`@(\w+/\w+)`)
	matches := atRefRegex.FindAllStringSubmatch(expression, -1)
	for _, match := range matches {
		if len(match) > 1 {
			references = append(references, match[1])
		}
	}

	// Find ID-like patterns
	idRefRegex := regexp.MustCompile(`\b(\w+\.\w+)\b`)
	matches = idRefRegex.FindAllStringSubmatch(expression, -1)
	for _, match := range matches {
		if len(match) > 1 {
			references = append(references, match[1])
		}
	}

	return references
}

func (rr *ReferenceResolverImpl) isReferenceString(str string) bool {
	// Check for @reference pattern
	if matched, _ := regexp.MatchString(`@\w+/\w+`, str); matched {
		return true
	}

	// Check for ID-like pattern
	if matched, _ := regexp.MatchString(`\w+\.\w+`, str); matched {
		return true
	}

	return false
}

func (rr *ReferenceResolverImpl) isValidReferenceFormat(ref string) bool {
	// Check for @reference pattern
	if matched, _ := regexp.MatchString(`@\w+/\w+`, ref); matched {
		return true
	}

	// Check for ID-like pattern
	if matched, _ := regexp.MatchString(`\w+\.\w+`, ref); matched {
		return true
	}

	return false
}

func (rr *ReferenceResolverImpl) removeDuplicateReferences(references []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)

	for _, ref := range references {
		if !seen[ref] {
			seen[ref] = true
			result = append(result, ref)
		}
	}

	return result
}
