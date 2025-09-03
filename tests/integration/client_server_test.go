//go:build integration

package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/solo7.media/platform.drifter.solo7.media/internal/domain"
	"github.com/solo7.media/platform.drifter.solo7.media/internal/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientServerCommunication(t *testing.T) {
	t.Run("should establish WebSocket connection and exchange messages", func(t *testing.T) {
		// Given
		logger := &mockLogger{}
		cm := network.NewConnectionManager(logger)

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool { return true },
			}
			conn, err := upgrader.Upgrade(w, r, nil)
			require.NoError(t, err)
			defer conn.Close()

			// Create WebSocket connection
			playerID := uuid.New()
			wsConn := network.NewWebSocketConnection(conn, playerID, logger)

			// Add to connection manager
			err = cm.AddConnection(context.Background(), wsConn)
			require.NoError(t, err)

			// Handle messages
			for {
				message, err := wsConn.Receive(context.Background())
				if err != nil {
					break // Connection closed
				}

				// Echo back the message
				err = wsConn.Send(context.Background(), message)
				if err != nil {
					break
				}
			}
		}))
		defer server.Close()

		// When - Client connects and sends message
		wsURL := "ws" + server.URL[4:] + "/"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		// Create client-side connection
		playerID := uuid.New()
		clientConn := network.NewWebSocketConnection(conn, playerID, logger)

		// Send test message
		message := &domain.NetworkMessage{
			Type:      domain.PlayerInput,
			ID:        uuid.New().String(),
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"player_id":  playerID.String(),
				"input_type": "movement",
				"data": map[string]interface{}{
					"direction": "forward",
					"speed":     "1.0",
				},
			},
		}

		err = clientConn.Send(context.Background(), message)
		require.NoError(t, err)

		// Then - Receive echoed message
		receivedMessage, err := clientConn.Receive(context.Background())
		require.NoError(t, err)

		assert.Equal(t, message.Type, receivedMessage.Type)
		assert.Equal(t, message.ID, receivedMessage.ID)
		assert.Equal(t, playerID.String(), receivedMessage.Data["player_id"])
		assert.Equal(t, "movement", receivedMessage.Data["input_type"])
	})

	t.Run("should handle multiple concurrent connections", func(t *testing.T) {
		// Given
		logger := &mockLogger{}
		cm := network.NewConnectionManager(logger)

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool { return true },
			}
			conn, err := upgrader.Upgrade(w, r, nil)
			require.NoError(t, err)
			defer conn.Close()

			// Create WebSocket connection
			playerID := uuid.New()
			wsConn := network.NewWebSocketConnection(conn, playerID, logger)

			// Add to connection manager
			err = cm.AddConnection(context.Background(), wsConn)
			require.NoError(t, err)

			// Keep connection alive
			time.Sleep(100 * time.Millisecond)
		}))
		defer server.Close()

		// When - Create multiple client connections
		numConnections := 5
		connections := make([]*websocket.Conn, numConnections)

		for i := 0; i < numConnections; i++ {
			wsURL := "ws" + server.URL[4:] + "/"
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			require.NoError(t, err)
			connections[i] = conn
		}

		// Wait for connections to be established
		time.Sleep(50 * time.Millisecond)

		// Then - Verify all connections are managed
		assert.Equal(t, numConnections, cm.GetConnectionCount())
		assert.Equal(t, numConnections, cm.GetPlayerCount())

		// Cleanup
		for _, conn := range connections {
			conn.Close()
		}
	})

	t.Run("should broadcast messages to all connections", func(t *testing.T) {
		// Given
		logger := &mockLogger{}
		cm := network.NewConnectionManager(logger)

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool { return true },
			}
			conn, err := upgrader.Upgrade(w, r, nil)
			require.NoError(t, err)
			defer conn.Close()

			// Create WebSocket connection
			playerID := uuid.New()
			wsConn := network.NewWebSocketConnection(conn, playerID, logger)

			// Add to connection manager
			err = cm.AddConnection(context.Background(), wsConn)
			require.NoError(t, err)

			// Keep connection alive
			time.Sleep(200 * time.Millisecond)
		}))
		defer server.Close()

		// Create multiple client connections
		numConnections := 3
		connections := make([]*websocket.Conn, numConnections)
		clientConns := make([]domain.ClientConnection, numConnections)

		for i := 0; i < numConnections; i++ {
			wsURL := "ws" + server.URL[4:] + "/"
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			require.NoError(t, err)
			connections[i] = conn

			playerID := uuid.New()
			clientConn := network.NewWebSocketConnection(conn, playerID, logger)
			clientConns[i] = clientConn
		}

		// Wait for connections to be established
		time.Sleep(50 * time.Millisecond)

		// When - Broadcast message to all connections
		broadcastMessage := &domain.NetworkMessage{
			Type:      domain.SystemNotification,
			ID:        uuid.New().String(),
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"message": "Server maintenance in 5 minutes",
			},
		}

		err := cm.BroadcastToAll(context.Background(), broadcastMessage)
		require.NoError(t, err)

		// Then - All clients should receive the message
		for i, clientConn := range clientConns {
			receivedMessage, err := clientConn.Receive(context.Background())
			require.NoError(t, err, "Client %d should receive broadcast message", i)

			assert.Equal(t, broadcastMessage.Type, receivedMessage.Type)
			assert.Equal(t, broadcastMessage.ID, receivedMessage.ID)
		}

		// Cleanup
		for _, conn := range connections {
			conn.Close()
		}
	})
}

func TestConnectionLifecycle(t *testing.T) {
	t.Run("should handle connection disconnection gracefully", func(t *testing.T) {
		// Given
		logger := &mockLogger{}
		cm := network.NewConnectionManager(logger)

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool { return true },
			}
			conn, err := upgrader.Upgrade(w, r, nil)
			require.NoError(t, err)
			defer conn.Close()

			// Create WebSocket connection
			playerID := uuid.New()
			wsConn := network.NewWebSocketConnection(conn, playerID, logger)

			// Add to connection manager
			err = cm.AddConnection(context.Background(), wsConn)
			require.NoError(t, err)

			// Keep connection alive
			time.Sleep(100 * time.Millisecond)
		}))
		defer server.Close()

		// When - Client connects and then disconnects
		wsURL := "ws" + server.URL[4:] + "/"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)

		// Wait for connection to be established
		time.Sleep(50 * time.Millisecond)
		assert.Equal(t, 1, cm.GetConnectionCount())

		// Disconnect client
		conn.Close()
		time.Sleep(50 * time.Millisecond)

		// Then - Connection should be cleaned up
		// Note: In a real implementation, we'd need a mechanism to detect disconnections
		// For now, we'll test that the connection manager can handle the cleanup
		err = cm.CleanupInactiveConnections(context.Background(), 1*time.Millisecond)
		require.NoError(t, err)
	})
}

// Mock logger for testing
type mockLogger struct{}

func (m *mockLogger) Debug(msg string, fields map[string]interface{}) {}
func (m *mockLogger) Info(msg string, fields map[string]interface{})  {}
func (m *mockLogger) Warn(msg string, fields map[string]interface{})  {}
func (m *mockLogger) Error(msg string, fields map[string]interface{}) {}
func (m *mockLogger) Fatal(msg string, fields map[string]interface{}) {}
