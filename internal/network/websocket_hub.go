package network

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/solo-seven/platform.drifter.solo7.media/generated/proto"
	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
)

// WebSocketHub manages WebSocket connections for real-time events
type WebSocketHub struct {
	// Registered connections
	connections map[string]*WebSocketConnection
	// Register requests from connections
	register chan *WebSocketConnection
	// Unregister requests from connections
	unregister chan *WebSocketConnection
	// Broadcast channel for sending messages to all connections
	broadcast chan *proto.GameEvent
	// Channel for sending messages to specific connections
	sendToConnection chan *ConnectionMessage
	// Logger
	logger domain.Logger
	// Mutex for thread safety
	mu sync.RWMutex
}

// ConnectionMessage represents a message to be sent to a specific connection
type ConnectionMessage struct {
	ConnectionID string
	Message      *proto.GameEvent
}

// WebSocketConnection represents a WebSocket connection
type WebSocketConnection struct {
	// The websocket connection
	conn *websocket.Conn
	// Buffered channel of outbound messages
	send chan []byte
	// Connection ID
	id string
	// Player ID
	playerID string
	// Last heartbeat time
	lastHeartbeat time.Time
	// Hub reference
	hub *WebSocketHub
	// Logger
	logger domain.Logger
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub(logger domain.Logger) *WebSocketHub {
	return &WebSocketHub{
		connections:      make(map[string]*WebSocketConnection),
		register:         make(chan *WebSocketConnection),
		unregister:       make(chan *WebSocketConnection),
		broadcast:        make(chan *proto.GameEvent),
		sendToConnection: make(chan *ConnectionMessage),
		logger:           logger,
	}
}

// Start starts the WebSocket hub
func (h *WebSocketHub) Start(ctx context.Context) {
	go h.run()
}

// run handles the hub's main loop
func (h *WebSocketHub) run() {
	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			h.connections[conn.id] = conn
			h.mu.Unlock()
			h.logger.Info("WebSocket connection registered", map[string]interface{}{
				"connection_id": conn.id,
				"player_id":     conn.playerID,
			})

		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.connections[conn.id]; ok {
				delete(h.connections, conn.id)
				close(conn.send)
			}
			h.mu.Unlock()
			h.logger.Info("WebSocket connection unregistered", map[string]interface{}{
				"connection_id": conn.id,
				"player_id":     conn.playerID,
			})

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, conn := range h.connections {
				select {
				case conn.send <- h.messageToBytes(message):
				default:
					close(conn.send)
					delete(h.connections, conn.id)
				}
			}
			h.mu.RUnlock()

		case msg := <-h.sendToConnection:
			h.mu.RLock()
			if conn, ok := h.connections[msg.ConnectionID]; ok {
				select {
				case conn.send <- h.messageToBytes(msg.Message):
				default:
					close(conn.send)
					delete(h.connections, conn.id)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all connected clients
func (h *WebSocketHub) Broadcast(message *proto.GameEvent) {
	h.broadcast <- message
}

// SendToConnection sends a message to a specific connection
func (h *WebSocketHub) SendToConnection(connectionID string, message *proto.GameEvent) {
	h.sendToConnection <- &ConnectionMessage{
		ConnectionID: connectionID,
		Message:      message,
	}
}

// SendToPlayer sends a message to a specific player
func (h *WebSocketHub) SendToPlayer(playerID string, message *proto.GameEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, conn := range h.connections {
		if conn.playerID == playerID {
			select {
			case conn.send <- h.messageToBytes(message):
			default:
				// Connection is full, skip
			}
		}
	}
}

// GetConnectionCount returns the number of active connections
func (h *WebSocketHub) GetConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.connections)
}

// messageToBytes converts a GameEvent to bytes
func (h *WebSocketHub) messageToBytes(message *proto.GameEvent) []byte {
	data, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("Failed to marshal message", map[string]interface{}{
			"error": err,
		})
		return nil
	}
	return data
}

// HandleWebSocket handles WebSocket connections
func (h *WebSocketHub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for development
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade WebSocket connection", map[string]interface{}{
			"error": err,
		})
		return
	}

	// Extract player ID from query parameters or headers
	playerID := r.URL.Query().Get("player_id")
	if playerID == "" {
		playerID = r.Header.Get("X-Player-ID")
	}
	if playerID == "" {
		playerID = uuid.New().String()
	}

	wsConn := &WebSocketConnection{
		conn:          conn,
		send:          make(chan []byte, 256),
		id:            uuid.New().String(),
		playerID:      playerID,
		lastHeartbeat: time.Now(),
		hub:           h,
		logger:        h.logger,
	}

	h.register <- wsConn

	// Start goroutines for reading and writing
	go wsConn.writePump()
	go wsConn.readPump()
}

// writePump pumps messages from the hub to the websocket connection
func (c *WebSocketConnection) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump pumps messages from the websocket connection to the hub
func (c *WebSocketConnection) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		c.lastHeartbeat = time.Now()
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Error("WebSocket error", map[string]interface{}{
					"error": err,
				})
			}
			break
		}
	}
}

// GetPlayerID returns the player ID for this connection
func (c *WebSocketConnection) GetPlayerID() string {
	return c.playerID
}

// GetConnectionID returns the connection ID
func (c *WebSocketConnection) GetConnectionID() string {
	return c.id
}

// IsConnected returns true if the connection is still active
func (c *WebSocketConnection) IsConnected() bool {
	return c.conn != nil
}

// GetLastHeartbeat returns the last heartbeat time
func (c *WebSocketConnection) GetLastHeartbeat() time.Time {
	return c.lastHeartbeat
}

// UpdateHeartbeat updates the last heartbeat time
func (c *WebSocketConnection) UpdateHeartbeat() {
	c.lastHeartbeat = time.Now()
}
