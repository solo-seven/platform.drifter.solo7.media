package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
)

// GameServerImpl implements the GameServer interface
type GameServerImpl struct {
	config            domain.Configuration
	connectionManager domain.ConnectionManager
	networkProtocol   domain.NetworkProtocol
	logger            domain.Logger

	// Server state
	running bool
	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc

	// Game state
	world          *domain.World
	gameLoop       *GameLoop
	dslInterpreter domain.DSLInterpreter
	actionService  domain.ActionService
	eventBus       domain.EventBus
}

// NewGameServer creates a new game server instance
func NewGameServer(
	config domain.Configuration,
	connectionManager domain.ConnectionManager,
	networkProtocol domain.NetworkProtocol,
	logger domain.Logger,
) *GameServerImpl {
	// Create action service
	actionService := NewActionService(logger)

	// Create event bus
	eventBus := NewEventBus(logger)

	// Create DSL interpreter
	dslInterpreter := NewDSLInterpreter(logger)

	// Create game loop
	gameLoop := NewGameLoop(10, actionService, eventBus, logger) // 10 ticks per second

	return &GameServerImpl{
		config:            config,
		connectionManager: connectionManager,
		networkProtocol:   networkProtocol,
		logger:            logger,
		gameLoop:          gameLoop,
		dslInterpreter:    dslInterpreter,
		actionService:     actionService,
		eventBus:          eventBus,
	}
}

// Start starts the game server
func (gs *GameServerImpl) Start(ctx context.Context) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if gs.running {
		return fmt.Errorf("server is already running")
	}

	gs.ctx, gs.cancel = context.WithCancel(ctx)
	gs.running = true

	// Start background tasks
	go gs.heartbeatMonitor()
	go gs.connectionCleanup()
	go gs.runGameLoop()

	gs.logger.Info("Game server started", map[string]interface{}{
		"max_connections":    gs.config.GetMaxConnections(),
		"heartbeat_interval": gs.config.GetHeartbeatInterval(),
	})

	return nil
}

// Stop stops the game server
func (gs *GameServerImpl) Stop(ctx context.Context) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if !gs.running {
		return fmt.Errorf("server is not running")
	}

	gs.cancel()
	gs.running = false

	gs.logger.Info("Game server stopped", nil)
	return nil
}

// IsRunning returns whether the server is running
func (gs *GameServerImpl) IsRunning() bool {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.running
}

// GetConnectionManager returns the connection manager
func (gs *GameServerImpl) GetConnectionManager() domain.ConnectionManager {
	return gs.connectionManager
}

// GetEntityManager returns a placeholder entity manager
func (gs *GameServerImpl) GetEntityManager() domain.EntityManager {
	return nil
}

// GetRulesEngine returns a placeholder rules engine
func (gs *GameServerImpl) GetRulesEngine() domain.RulesEngine {
	return nil
}

// GetWorldStateManager returns a placeholder world state manager
func (gs *GameServerImpl) GetWorldStateManager() domain.WorldStateManager {
	return nil
}

// GetNetworkProtocol returns the network protocol
func (gs *GameServerImpl) GetNetworkProtocol() domain.NetworkProtocol {
	return gs.networkProtocol
}

// HandleWebSocket handles WebSocket connections
func (gs *GameServerImpl) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for development
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		gs.logger.Error("Failed to upgrade WebSocket connection", map[string]interface{}{
			"error": err,
		})
		return
	}
	defer conn.Close()

	// Generate player ID for new connection
	playerID := uuid.New()

	// The WebSocket hub will handle the connection
	// This is just a placeholder for now
	gs.logger.Info("WebSocket connection established", map[string]interface{}{
		"player_id": playerID,
	})

	gs.logger.Info("New player connected", map[string]interface{}{
		"player_id":     playerID,
		"total_players": gs.connectionManager.GetPlayerCount(),
	})

	// The WebSocket hub will handle messages
	// This is just a placeholder for now
}

// handleConnection handles messages from a specific connection
func (gs *GameServerImpl) handleConnection(conn domain.ClientConnection, playerID domain.PlayerId) {
	defer func() {
		// Remove connection when done
		if err := gs.connectionManager.RemoveConnection(gs.ctx, conn.GetConnectionID()); err != nil {
			gs.logger.Error("Failed to remove connection", map[string]interface{}{
				"player_id": playerID,
				"error":     err,
			})
		}

		gs.logger.Info("Player disconnected", map[string]interface{}{
			"player_id":     playerID,
			"total_players": gs.connectionManager.GetPlayerCount(),
		})
	}()

	for {
		select {
		case <-gs.ctx.Done():
			return
		default:
			message, err := conn.Receive(gs.ctx)
			if err != nil {
				gs.logger.Debug("Connection receive error", map[string]interface{}{
					"player_id": playerID,
					"error":     err,
				})
				return
			}

			// Process the message
			if err := gs.processMessage(conn, playerID, message); err != nil {
				gs.logger.Error("Failed to process message", map[string]interface{}{
					"player_id":    playerID,
					"message_type": message.Type,
					"error":        err,
				})
			}
		}
	}
}

// processMessage processes incoming messages
func (gs *GameServerImpl) processMessage(conn domain.ClientConnection, playerID domain.PlayerId, message *domain.NetworkMessage) error {
	switch message.Type {
	case domain.PlayerInput:
		return gs.handlePlayerInput(conn, playerID, message)
	case domain.ChatMessage:
		return gs.handleChatMessage(conn, playerID, message)
	case domain.Heartbeat:
		return gs.handleHeartbeat(conn, playerID, message)
	case domain.AdminCommand:
		return gs.handleAdminCommand(conn, playerID, message)
	default:
		gs.logger.Warn("Unknown message type", map[string]interface{}{
			"player_id":    playerID,
			"message_type": message.Type,
		})
		return nil
	}
}

// handlePlayerInput processes player input messages
func (gs *GameServerImpl) handlePlayerInput(conn domain.ClientConnection, playerID domain.PlayerId, message *domain.NetworkMessage) error {
	inputType := message.Data["input_type"]

	switch inputType {
	case "movement":
		return gs.handleMovementInput(conn, playerID, message)
	case "attack":
		return gs.handleAttackInput(conn, playerID, message)
	default:
		gs.logger.Debug("Unknown input type", map[string]interface{}{
			"player_id":  playerID,
			"input_type": inputType,
		})
	}

	return nil
}

// handleMovementInput processes movement input
func (gs *GameServerImpl) handleMovementInput(conn domain.ClientConnection, playerID domain.PlayerId, message *domain.NetworkMessage) error {
	// Get movement data
	data := message.Data["data"].(map[string]interface{})
	direction := data["direction"].(string)
	speed := data["speed"].(string)

	gs.logger.Debug("Player movement", map[string]interface{}{
		"player_id": playerID,
		"direction": direction,
		"speed":     speed,
	})

	// Create state update message
	stateUpdate := &domain.NetworkMessage{
		Type:      domain.StateUpdate,
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"entity_id": playerID.String(),
			"component": "transform",
			"changes": map[string]interface{}{
				"position": map[string]interface{}{
					"x": 1.0,
					"y": 0.0,
					"z": 1.0,
				},
			},
		},
	}

	// Send update back to player
	return conn.Send(gs.ctx, stateUpdate)
}

// handleAttackInput processes attack input
func (gs *GameServerImpl) handleAttackInput(conn domain.ClientConnection, playerID domain.PlayerId, message *domain.NetworkMessage) error {
	// Get attack data
	data := message.Data["data"].(map[string]interface{})
	target := data["target"].(string)
	weapon := data["weapon"].(string)

	gs.logger.Debug("Player attack", map[string]interface{}{
		"player_id": playerID,
		"target":    target,
		"weapon":    weapon,
	})

	// Create aesthetic event for other players
	aestheticEvent := &domain.NetworkMessage{
		Type:      domain.AestheticEvent,
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"type":      "player_action",
			"player_id": playerID.String(),
			"action":    "attack",
			"data": map[string]interface{}{
				"target": target,
				"weapon": weapon,
			},
		},
	}

	// Broadcast to all other players
	return gs.connectionManager.BroadcastToAll(gs.ctx, aestheticEvent)
}

// handleChatMessage processes chat messages
func (gs *GameServerImpl) handleChatMessage(conn domain.ClientConnection, playerID domain.PlayerId, message *domain.NetworkMessage) error {
	msg := message.Data["message"].(string)
	channel := message.Data["channel"].(string)

	gs.logger.Debug("Chat message", map[string]interface{}{
		"player_id": playerID,
		"channel":   channel,
		"message":   msg,
	})

	// Create chat message for broadcast
	chatMessage := &domain.NetworkMessage{
		Type:      domain.ChatMessage,
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"player_id": playerID.String(),
			"message":   msg,
			"channel":   channel,
		},
	}

	// Broadcast to all players
	return gs.connectionManager.BroadcastToAll(gs.ctx, chatMessage)
}

// handleHeartbeat processes heartbeat messages
func (gs *GameServerImpl) handleHeartbeat(conn domain.ClientConnection, playerID domain.PlayerId, message *domain.NetworkMessage) error {
	conn.UpdateHeartbeat()
	return nil
}

// handleAdminCommand processes admin commands
func (gs *GameServerImpl) handleAdminCommand(conn domain.ClientConnection, playerID domain.PlayerId, message *domain.NetworkMessage) error {
	// TODO: Implement admin command handling
	gs.logger.Debug("Admin command received", map[string]interface{}{
		"player_id": playerID,
		"command":   message.Data,
	})
	return nil
}

// sendWelcomeMessage sends a welcome message to a new player
func (gs *GameServerImpl) sendWelcomeMessage(conn domain.ClientConnection, playerID domain.PlayerId) {
	welcomeMessage := &domain.NetworkMessage{
		Type:      domain.SystemNotification,
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"type":    "welcome",
			"message": "Welcome to Drifter Platform RPG!",
			"properties": map[string]interface{}{
				"server_version": "1.0.0",
				"online_players": gs.connectionManager.GetPlayerCount(),
			},
		},
	}

	if err := conn.Send(gs.ctx, welcomeMessage); err != nil {
		gs.logger.Error("Failed to send welcome message", map[string]interface{}{
			"player_id": playerID,
			"error":     err,
		})
	}
}

// heartbeatMonitor monitors connection heartbeats
func (gs *GameServerImpl) heartbeatMonitor() {
	ticker := time.NewTicker(gs.config.GetHeartbeatInterval())
	defer ticker.Stop()

	for {
		select {
		case <-gs.ctx.Done():
			return
		case <-ticker.C:
			// Cleanup inactive connections
			maxInactiveTime := gs.config.GetHeartbeatInterval() * 3
			if err := gs.connectionManager.CleanupInactiveConnections(gs.ctx, maxInactiveTime); err != nil {
				gs.logger.Error("Failed to cleanup inactive connections", map[string]interface{}{
					"error": err,
				})
			}
		}
	}
}

// connectionCleanup periodically cleans up inactive connections
func (gs *GameServerImpl) connectionCleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-gs.ctx.Done():
			return
		case <-ticker.C:
			// Log connection statistics
			gs.logger.Debug("Connection statistics", map[string]interface{}{
				"total_connections": gs.connectionManager.GetConnectionCount(),
				"total_players":     gs.connectionManager.GetPlayerCount(),
			})
		}
	}
}

// runGameLoop runs the main game loop
func (gs *GameServerImpl) runGameLoop() {
	ticker := time.NewTicker(100 * time.Millisecond) // 10 FPS
	defer ticker.Stop()

	for {
		select {
		case <-gs.ctx.Done():
			return
		case <-ticker.C:
			// Update game state
			gs.updateGameState()
		}
	}
}

// updateGameState updates the game state
func (gs *GameServerImpl) updateGameState() {
	// TODO: Implement game state updates
	// This would include:
	// - Entity position updates
	// - Physics simulation
	// - AI behavior
	// - Status effect processing
	// - etc.
}

// initializeWorldState creates the initial world state
func initializeWorldState() *domain.WorldState {
	return &domain.WorldState{
		Regions: make(map[domain.RegionId]domain.RegionState),
		GlobalState: domain.GlobalGameState{
			GameTime:   time.Now(),
			GamePhase:  "active",
			Properties: make(map[string]interface{}),
		},
		PlayerStates: make(map[domain.PlayerId]domain.PlayerState),
	}
}
