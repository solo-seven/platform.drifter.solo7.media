package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
)

// ResolutionManagerImpl implements the ResolutionManager interface
type ResolutionManagerImpl struct {
	resolutionMethods map[domain.ResolutionId]*domain.ResolutionMethod
	mu                sync.RWMutex
	logger            domain.Logger
}

// NewResolutionManager creates a new resolution manager
func NewResolutionManager(logger domain.Logger) *ResolutionManagerImpl {
	return &ResolutionManagerImpl{
		resolutionMethods: make(map[domain.ResolutionId]*domain.ResolutionMethod),
		logger:            logger,
	}
}

// RegisterResolutionMethod registers a new resolution method
func (rm *ResolutionManagerImpl) RegisterResolutionMethod(ctx context.Context, method *domain.ResolutionMethod) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if method.ID == "" {
		return fmt.Errorf("resolution method ID cannot be empty")
	}

	if method.Name == "" {
		return fmt.Errorf("resolution method name cannot be empty")
	}

	// Validate the resolution method
	if err := rm.validateResolutionMethod(method); err != nil {
		return fmt.Errorf("resolution method validation failed: %w", err)
	}

	// Check if resolution method already exists
	if _, exists := rm.resolutionMethods[method.ID]; exists {
		return fmt.Errorf("resolution method with ID %s already exists", method.ID)
	}

	rm.resolutionMethods[method.ID] = method

	rm.logger.Info("Resolution method registered", map[string]interface{}{
		"resolution_id":      method.ID,
		"resolution_name":    method.Name,
		"applicable_actions": len(method.ApplicableActions),
	})

	return nil
}

// UnregisterResolutionMethod removes a resolution method
func (rm *ResolutionManagerImpl) UnregisterResolutionMethod(ctx context.Context, resolutionID domain.ResolutionId) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	method, exists := rm.resolutionMethods[resolutionID]
	if !exists {
		return fmt.Errorf("resolution method with ID %s not found", resolutionID)
	}

	delete(rm.resolutionMethods, resolutionID)

	rm.logger.Info("Resolution method unregistered", map[string]interface{}{
		"resolution_id":   resolutionID,
		"resolution_name": method.Name,
	})

	return nil
}

// GetResolutionMethod retrieves a resolution method by ID
func (rm *ResolutionManagerImpl) GetResolutionMethod(ctx context.Context, resolutionID domain.ResolutionId) (*domain.ResolutionMethod, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	method, exists := rm.resolutionMethods[resolutionID]
	if !exists {
		return nil, fmt.Errorf("resolution method with ID %s not found", resolutionID)
	}

	// Return a copy to prevent external modification
	methodCopy := *method
	return &methodCopy, nil
}

// GetResolutionMethodsForAction retrieves all resolution methods applicable to a specific action
func (rm *ResolutionManagerImpl) GetResolutionMethodsForAction(ctx context.Context, actionID domain.ActionTypeId) ([]*domain.ResolutionMethod, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var methods []*domain.ResolutionMethod
	for _, method := range rm.resolutionMethods {
		// Check if this resolution method is applicable to the action
		for _, applicableAction := range method.ApplicableActions {
			if applicableAction == actionID {
				// Return a copy to prevent external modification
				methodCopy := *method
				methods = append(methods, &methodCopy)
				break
			}
		}
	}

	return methods, nil
}

// ValidateResolutionMethod validates a resolution method
func (rm *ResolutionManagerImpl) ValidateResolutionMethod(ctx context.Context, method *domain.ResolutionMethod) error {
	return rm.validateResolutionMethod(method)
}

// ListResolutionMethods returns all registered resolution methods
func (rm *ResolutionManagerImpl) ListResolutionMethods(ctx context.Context) ([]*domain.ResolutionMethod, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var methods []*domain.ResolutionMethod
	for _, method := range rm.resolutionMethods {
		// Return a copy to prevent external modification
		methodCopy := *method
		methods = append(methods, &methodCopy)
	}

	return methods, nil
}

// validateResolutionMethod performs internal validation of a resolution method
func (rm *ResolutionManagerImpl) validateResolutionMethod(method *domain.ResolutionMethod) error {
	if method == nil {
		return fmt.Errorf("resolution method cannot be nil")
	}

	if method.ID == "" {
		return fmt.Errorf("resolution method ID is required")
	}

	if method.Name == "" {
		return fmt.Errorf("resolution method name is required")
	}

	if len(method.ApplicableActions) == 0 {
		return fmt.Errorf("resolution method must have at least one applicable action")
	}

	// Validate required inputs
	for i, input := range method.RequiredInputs {
		if input.Name == "" {
			return fmt.Errorf("required input %d name cannot be empty", i)
		}
		if input.Type == "" {
			return fmt.Errorf("required input %d type cannot be empty for input: %s", i, input.Name)
		}
	}

	// Validate calculation steps
	if len(method.CalculationSteps) == 0 {
		return fmt.Errorf("resolution method must have at least one calculation step")
	}

	for i, step := range method.CalculationSteps {
		if step.Step <= 0 {
			return fmt.Errorf("calculation step %d step number must be positive", i)
		}
		if step.Type == "" {
			return fmt.Errorf("calculation step %d type cannot be empty", i)
		}
		if step.Expression == "" {
			return fmt.Errorf("calculation step %d expression cannot be empty", i)
		}
	}

	// Validate possible outcomes
	for i, outcome := range method.PossibleOutcomes {
		if outcome.Type == "" {
			return fmt.Errorf("possible outcome %d type cannot be empty", i)
		}
	}

	return nil
}
