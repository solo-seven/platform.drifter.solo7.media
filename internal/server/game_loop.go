package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
)

// GameLoop implements a fixed-timestep game loop
type GameLoop struct {
	// Configuration
	tickRate     int           // Ticks per second
	tickDuration time.Duration // Duration of each tick

	// State
	running bool
	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc

	// Channels for communication
	inputChan  chan domain.MUDAction
	outputChan chan domain.MUDGameEvent

	// Game state
	world       *domain.World
	actionQueue []domain.MUDAction
	eventQueue  []domain.MUDGameEvent

	// Services
	actionService domain.ActionService
	eventBus      domain.EventBus
	logger        domain.Logger
}

// NewGameLoop creates a new game loop
func NewGameLoop(tickRate int, actionService domain.ActionService, eventBus domain.EventBus, logger domain.Logger) *GameLoop {
	return &GameLoop{
		tickRate:      tickRate,
		tickDuration:  time.Second / time.Duration(tickRate),
		inputChan:     make(chan domain.MUDAction, 1000),
		outputChan:    make(chan domain.MUDGameEvent, 1000),
		actionQueue:   make([]domain.MUDAction, 0, 100),
		eventQueue:    make([]domain.MUDGameEvent, 0, 100),
		actionService: actionService,
		eventBus:      eventBus,
		logger:        logger,
	}
}

// Start starts the game loop
func (gl *GameLoop) Start(ctx context.Context) error {
	gl.mu.Lock()
	defer gl.mu.Unlock()

	if gl.running {
		return fmt.Errorf("game loop is already running")
	}

	gl.ctx, gl.cancel = context.WithCancel(ctx)
	gl.running = true

	// Start the main game loop goroutine
	go gl.run()

	gl.logger.Info("Game loop started", map[string]interface{}{
		"tick_rate":     gl.tickRate,
		"tick_duration": gl.tickDuration,
	})

	return nil
}

// Stop stops the game loop
func (gl *GameLoop) Stop(ctx context.Context) error {
	gl.mu.Lock()
	defer gl.mu.Unlock()

	if !gl.running {
		return fmt.Errorf("game loop is not running")
	}

	gl.cancel()
	gl.running = false

	gl.logger.Info("Game loop stopped", map[string]interface{}{})
	return nil
}

// IsRunning returns whether the game loop is running
func (gl *GameLoop) IsRunning() bool {
	gl.mu.RLock()
	defer gl.mu.RUnlock()
	return gl.running
}

// GetTickRate returns the current tick rate
func (gl *GameLoop) GetTickRate() int {
	return gl.tickRate
}

// SetTickRate sets the tick rate
func (gl *GameLoop) SetTickRate(rate int) {
	gl.mu.Lock()
	defer gl.mu.Unlock()

	gl.tickRate = rate
	gl.tickDuration = time.Second / time.Duration(rate)
}

// GetInputChannel returns the input channel for actions
func (gl *GameLoop) GetInputChannel() chan<- domain.MUDAction {
	return gl.inputChan
}

// GetOutputChannel returns the output channel for events
func (gl *GameLoop) GetOutputChannel() <-chan domain.MUDGameEvent {
	return gl.outputChan
}

// SetWorld sets the game world
func (gl *GameLoop) SetWorld(world *domain.World) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.world = world
}

// run is the main game loop
func (gl *GameLoop) run() {
	ticker := time.NewTicker(gl.tickDuration)
	defer ticker.Stop()

	for {
		select {
		case <-gl.ctx.Done():
			return
		case <-ticker.C:
			gl.tick()
		case action := <-gl.inputChan:
			gl.queueAction(action)
		}
	}
}

// tick processes one game tick
func (gl *GameLoop) tick() {
	start := time.Now()

	// Phase 1: Process Input Queue
	gl.processInputQueue()

	// Phase 2: Update Game State
	gl.updateGameState()

	// Phase 3: Process Output Queue
	gl.processOutputQueue()

	// Check if we're falling behind
	elapsed := time.Since(start)
	if elapsed > gl.tickDuration {
		gl.logger.Warn("Game tick took longer than target duration", map[string]interface{}{
			"elapsed": elapsed,
			"target":  gl.tickDuration,
			"overrun": elapsed - gl.tickDuration,
		})
	}
}

// processInputQueue processes all queued actions
func (gl *GameLoop) processInputQueue() {
	gl.mu.Lock()
	actions := make([]domain.MUDAction, len(gl.actionQueue))
	copy(actions, gl.actionQueue)
	gl.actionQueue = gl.actionQueue[:0] // Clear the queue
	gl.mu.Unlock()

	for _, action := range actions {
		gl.processAction(action)
	}
}

// processAction processes a single action
func (gl *GameLoop) processAction(action domain.MUDAction) {
	if gl.world == nil {
		gl.logger.Warn("Cannot process action: world not set", map[string]interface{}{})
		return
	}

	// Validate the action
	if gl.actionService != nil {
		// Convert MUDAction to Action for validation
		// TODO: Implement proper conversion or create MUD-specific validation
		gl.logger.Debug("Validating action", map[string]interface{}{
			"action_type":  action.GetType(),
			"character_id": action.GetCharacterID(),
		})
	}

	// Execute the action
	result, err := action.Execute(gl.ctx, gl.world)
	if err != nil {
		gl.logger.Error("Action execution failed", map[string]interface{}{
			"action_type":  action.GetType(),
			"character_id": action.GetCharacterID(),
			"error":        err,
		})
		return
	}

	// Queue state changes
	for _, change := range result.StateChanges {
		gl.applyStateChange(change)
	}

	// Queue events
	for _, event := range result.Events {
		gl.queueEvent(event)
	}

	// Queue notifications
	for _, notification := range result.Notifications {
		gl.queueNotification(notification)
	}
}

// applyStateChange applies a state change to the world
func (gl *GameLoop) applyStateChange(change domain.MUDStateChange) {
	// TODO: Implement state change application
	// This would update the world state based on the change
	gl.logger.Debug("Applying state change", map[string]interface{}{
		"entity_id": change.EntityID,
		"component": change.Component,
		"changes":   change.Changes,
	})
}

// queueEvent queues an event for output
func (gl *GameLoop) queueEvent(event domain.MUDGameEvent) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.eventQueue = append(gl.eventQueue, event)
}

// queueNotification queues a notification for output
func (gl *GameLoop) queueNotification(notification domain.MUDClientNotification) {
	// Convert notification to event
	event := domain.MUDGameEvent{
		ID:   fmt.Sprintf("notification_%d", time.Now().UnixNano()),
		Type: "client_notification",
		Data: map[string]interface{}{
			"player_id":  notification.PlayerID,
			"type":       notification.Type,
			"message":    notification.Message,
			"properties": notification.Properties,
		},
		Timestamp: time.Now(),
	}
	gl.queueEvent(event)
}

// queueAction queues an action for processing
func (gl *GameLoop) queueAction(action domain.MUDAction) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.actionQueue = append(gl.actionQueue, action)
}

// updateGameState updates the game state
func (gl *GameLoop) updateGameState() {
	if gl.world == nil {
		return
	}

	// Update NPCs
	gl.updateNPCs()

	// Update status effects
	gl.updateStatusEffects()

	// Update quests
	gl.updateQuests()

	// Update world time
	gl.updateWorldTime()
}

// updateNPCs updates all NPCs
func (gl *GameLoop) updateNPCs() {
	// TODO: Implement NPC AI updates
	// This would process NPC behaviors, movement, etc.
}

// updateStatusEffects updates all status effects
func (gl *GameLoop) updateStatusEffects() {
	// TODO: Implement status effect updates
	// This would process ongoing effects like poison, buffs, etc.
}

// updateQuests updates all quests
func (gl *GameLoop) updateQuests() {
	// TODO: Implement quest updates
	// This would check quest objectives, trigger quest events, etc.
}

// updateWorldTime updates the world time
func (gl *GameLoop) updateWorldTime() {
	// TODO: Implement world time updates
	// This would advance the game world's time
}

// processOutputQueue processes all queued events
func (gl *GameLoop) processOutputQueue() {
	gl.mu.Lock()
	events := make([]domain.MUDGameEvent, len(gl.eventQueue))
	copy(events, gl.eventQueue)
	gl.eventQueue = gl.eventQueue[:0] // Clear the queue
	gl.mu.Unlock()

	for _, event := range events {
		// Send to output channel
		select {
		case gl.outputChan <- event:
		default:
			gl.logger.Warn("Output channel full, dropping event", map[string]interface{}{
				"event_type": event.Type,
			})
		}

		// Publish to event bus
		if gl.eventBus != nil {
			if err := gl.eventBus.Publish(gl.ctx, domain.EventType(event.Type), event.Data); err != nil {
				gl.logger.Error("Failed to publish event", map[string]interface{}{
					"event_type": event.Type,
					"error":      err,
				})
			}
		}
	}
}
