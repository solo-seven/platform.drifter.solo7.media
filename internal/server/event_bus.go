package server

import (
	"context"

	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
)

// EventBus implements domain.EventBus
type EventBus struct {
	logger domain.Logger
}

// NewEventBus creates a new event bus
func NewEventBus(logger domain.Logger) *EventBus {
	return &EventBus{
		logger: logger,
	}
}

// Publish publishes an event
func (eb *EventBus) Publish(ctx context.Context, eventType domain.EventType, data map[string]interface{}) error {
	// TODO: Implement actual event publishing
	eb.logger.Debug("Event published", map[string]interface{}{
		"event_type": eventType,
	})
	return nil
}

// Subscribe subscribes to an event type
func (eb *EventBus) Subscribe(ctx context.Context, eventType string, handler domain.EventHandler) error {
	// TODO: Implement actual subscription
	return nil
}

// Unsubscribe unsubscribes from an event type
func (eb *EventBus) Unsubscribe(ctx context.Context, eventType string, handler domain.EventHandler) error {
	// TODO: Implement actual unsubscription
	return nil
}
