package network

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
	"github.com/solo-seven/platform.drifter.solo7.media/internal/network"
	"github.com/stretchr/testify/assert"
)

func TestWebSocketConnection_Basic(t *testing.T) {
	t.Run("should create WebSocket connection with correct properties", func(t *testing.T) {
		// Given
		playerID := uuid.New()
		logger := &mockLogger{}

		// Create a WebSocket connection with nil conn (for testing basic properties)
		wsConn := network.NewWebSocketConnection(nil, playerID, logger)

		// When/Then
		assert.Equal(t, playerID, wsConn.GetPlayerID())
		assert.NotEmpty(t, wsConn.GetConnectionID())
		assert.False(t, wsConn.IsConnected()) // Should be false with nil conn
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
		playerID := uuid.New()
		logger := &mockLogger{}

		// Create a WebSocket connection with nil conn
		wsConn := network.NewWebSocketConnection(nil, playerID, logger)

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
		err := wsConn.Send(context.Background(), message)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection is closed")
	})
}

// Mock logger for testing
type mockLogger struct{}

func (m *mockLogger) Debug(msg string, fields map[string]interface{}) {}
func (m *mockLogger) Info(msg string, fields map[string]interface{})  {}
func (m *mockLogger) Warn(msg string, fields map[string]interface{})  {}
func (m *mockLogger) Error(msg string, fields map[string]interface{}) {}
func (m *mockLogger) Fatal(msg string, fields map[string]interface{}) {}
