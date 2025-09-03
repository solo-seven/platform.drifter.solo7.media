package network

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
	"github.com/solo-seven/platform.drifter.solo7.media/internal/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebSocketConnection_SendReceive(t *testing.T) {
	t.Run("should send and receive player input message", func(t *testing.T) {
		// Given
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{}
			conn, err := upgrader.Upgrade(w, r, nil)
			require.NoError(t, err)
			defer conn.Close()

			// Read message from client
			_, data, err := conn.ReadMessage()
			require.NoError(t, err)

			// Echo back the message
			err = conn.WriteMessage(websocket.BinaryMessage, data)
			require.NoError(t, err)
		}))
		defer server.Close()

		// Connect to the test server
		wsURL := "ws" + server.URL[4:] + "/"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		playerID := uuid.New()
		logger := &mockLogger{}
		wsConn := network.NewWebSocketConnection(conn, playerID, logger)

		// Create test message
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

		// When - Send message
		err = wsConn.Send(context.Background(), message)
		require.NoError(t, err)

		// Then - Receive echoed message
		receivedMessage, err := wsConn.Receive(context.Background())
		require.NoError(t, err)

		assert.Equal(t, message.Type, receivedMessage.Type)
		assert.Equal(t, message.ID, receivedMessage.ID)
		assert.Equal(t, playerID.String(), receivedMessage.Data["player_id"])
		assert.Equal(t, "movement", receivedMessage.Data["input_type"])
	})

	t.Run("should send and receive heartbeat message", func(t *testing.T) {
		// Given
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{}
			conn, err := upgrader.Upgrade(w, r, nil)
			require.NoError(t, err)
			defer conn.Close()

			// Read message from client
			_, data, err := conn.ReadMessage()
			require.NoError(t, err)

			// Echo back the message
			err = conn.WriteMessage(websocket.BinaryMessage, data)
			require.NoError(t, err)
		}))
		defer server.Close()

		// Connect to the test server
		wsURL := "ws" + server.URL[4:] + "/"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		playerID := uuid.New()
		logger := &mockLogger{}
		wsConn := network.NewWebSocketConnection(conn, playerID, logger)

		// Create heartbeat message
		message := &domain.NetworkMessage{
			Type:      domain.Heartbeat,
			ID:        uuid.New().String(),
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"connection_id": wsConn.GetConnectionID(),
			},
		}

		// When - Send heartbeat
		err = wsConn.Send(context.Background(), message)
		require.NoError(t, err)

		// Then - Receive echoed heartbeat
		receivedMessage, err := wsConn.Receive(context.Background())
		require.NoError(t, err)

		assert.Equal(t, domain.Heartbeat, receivedMessage.Type)
		assert.Equal(t, wsConn.GetConnectionID(), receivedMessage.Data["connection_id"])
	})

	t.Run("should handle connection properties correctly", func(t *testing.T) {
		// Given
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{}
			conn, err := upgrader.Upgrade(w, r, nil)
			require.NoError(t, err)
			defer conn.Close()

			// Keep connection open for testing
			time.Sleep(100 * time.Millisecond)
		}))
		defer server.Close()

		// Connect to the test server
		wsURL := "ws" + server.URL[4:] + "/"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		playerID := uuid.New()
		logger := &mockLogger{}
		wsConn := network.NewWebSocketConnection(conn, playerID, logger)

		// When/Then
		assert.Equal(t, playerID, wsConn.GetPlayerID())
		assert.NotEmpty(t, wsConn.GetConnectionID())
		assert.True(t, wsConn.IsConnected())
		assert.NotZero(t, wsConn.GetLastHeartbeat())

		// Test heartbeat update
		originalHeartbeat := wsConn.GetLastHeartbeat()
		time.Sleep(1 * time.Millisecond)
		wsConn.UpdateHeartbeat()
		assert.True(t, wsConn.GetLastHeartbeat().After(originalHeartbeat))
	})
}

func TestWebSocketConnection_ErrorHandling(t *testing.T) {
	t.Run("should handle closed connection", func(t *testing.T) {
		// Given
		playerID := uuid.New()
		logger := &mockLogger{}

		// Create a closed connection (nil conn)
		wsConn := network.NewWebSocketConnection(nil, playerID, logger)

		message := &domain.NetworkMessage{
			Type:      domain.Heartbeat,
			ID:        uuid.New().String(),
			Timestamp: time.Now(),
			Data:      make(map[string]interface{}),
		}

		// When
		err := wsConn.Send(context.Background(), message)

		// Then
		assert.Error(t, err)
		assert.False(t, wsConn.IsConnected())
	})

	t.Run("should handle invalid message data", func(t *testing.T) {
		// Given
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{}
			conn, err := upgrader.Upgrade(w, r, nil)
			require.NoError(t, err)
			defer conn.Close()
		}))
		defer server.Close()

		// Connect to the test server
		wsURL := "ws" + server.URL[4:] + "/"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		playerID := uuid.New()
		logger := &mockLogger{}
		wsConn := network.NewWebSocketConnection(conn, playerID, logger)

		// Create message with invalid data
		message := &domain.NetworkMessage{
			Type:      domain.PlayerInput,
			ID:        uuid.New().String(),
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				// Missing required player_id field
				"input_type": "movement",
			},
		}

		// When
		err = wsConn.Send(context.Background(), message)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing or invalid player_id")
	})
}

// Mock logger for testing
type mockLogger struct{}

func (m *mockLogger) Debug(msg string, fields map[string]interface{}) {}
func (m *mockLogger) Info(msg string, fields map[string]interface{})  {}
func (m *mockLogger) Warn(msg string, fields map[string]interface{})  {}
func (m *mockLogger) Error(msg string, fields map[string]interface{}) {}
func (m *mockLogger) Fatal(msg string, fields map[string]interface{}) {}
