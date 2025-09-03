package network

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
	"github.com/solo-seven/platform.drifter.solo7.media/internal/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectionManager_AddRemoveConnections(t *testing.T) {
	t.Run("should add connection successfully", func(t *testing.T) {
		// Given
		logger := &mockLogger{}
		cm := network.NewConnectionManager(logger)
		conn := &mockConnection{
			playerID:     uuid.New(),
			connectionID: uuid.New().String(),
		}

		// When
		err := cm.AddConnection(context.Background(), conn)

		// Then
		require.NoError(t, err)
		assert.Equal(t, 1, cm.GetConnectionCount())
		assert.Equal(t, 1, cm.GetPlayerCount())
	})

	t.Run("should remove connection successfully", func(t *testing.T) {
		// Given
		logger := &mockLogger{}
		cm := network.NewConnectionManager(logger)
		conn := &mockConnection{
			playerID:     uuid.New(),
			connectionID: uuid.New().String(),
		}

		err := cm.AddConnection(context.Background(), conn)
		require.NoError(t, err)

		// When
		err = cm.RemoveConnection(context.Background(), conn.GetConnectionID())

		// Then
		require.NoError(t, err)
		assert.Equal(t, 0, cm.GetConnectionCount())
		assert.Equal(t, 0, cm.GetPlayerCount())
	})

	t.Run("should replace existing player connection", func(t *testing.T) {
		// Given
		logger := &mockLogger{}
		cm := network.NewConnectionManager(logger)
		playerID := uuid.New()

		oldConn := &mockConnection{
			playerID:     playerID,
			connectionID: uuid.New().String(),
		}

		newConn := &mockConnection{
			playerID:     playerID,
			connectionID: uuid.New().String(),
		}

		// Add first connection
		err := cm.AddConnection(context.Background(), oldConn)
		require.NoError(t, err)
		assert.Equal(t, 1, cm.GetConnectionCount())

		// When - Add second connection for same player
		err = cm.AddConnection(context.Background(), newConn)

		// Then
		require.NoError(t, err)
		assert.Equal(t, 1, cm.GetConnectionCount()) // Still only one connection
		assert.Equal(t, 1, cm.GetPlayerCount())

		// Verify old connection was closed
		assert.True(t, oldConn.closed)

		// Verify new connection is active
		retrievedConn, err := cm.GetPlayerConnection(context.Background(), playerID)
		require.NoError(t, err)
		assert.Equal(t, newConn.GetConnectionID(), retrievedConn.GetConnectionID())
	})

	t.Run("should return error for non-existent connection", func(t *testing.T) {
		// Given
		logger := &mockLogger{}
		cm := network.NewConnectionManager(logger)
		nonExistentID := uuid.New().String()

		// When
		err := cm.RemoveConnection(context.Background(), nonExistentID)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestConnectionManager_GetConnections(t *testing.T) {
	t.Run("should get connection by ID", func(t *testing.T) {
		// Given
		logger := &mockLogger{}
		cm := network.NewConnectionManager(logger)
		conn := &mockConnection{
			playerID:     uuid.New(),
			connectionID: uuid.New().String(),
		}

		err := cm.AddConnection(context.Background(), conn)
		require.NoError(t, err)

		// When
		retrievedConn, err := cm.GetConnection(context.Background(), conn.GetConnectionID())

		// Then
		require.NoError(t, err)
		assert.Equal(t, conn.GetConnectionID(), retrievedConn.GetConnectionID())
		assert.Equal(t, conn.GetPlayerID(), retrievedConn.GetPlayerID())
	})

	t.Run("should get connection by player ID", func(t *testing.T) {
		// Given
		logger := &mockLogger{}
		cm := network.NewConnectionManager(logger)
		playerID := uuid.New()
		conn := &mockConnection{
			playerID:     playerID,
			connectionID: uuid.New().String(),
		}

		err := cm.AddConnection(context.Background(), conn)
		require.NoError(t, err)

		// When
		retrievedConn, err := cm.GetPlayerConnection(context.Background(), playerID)

		// Then
		require.NoError(t, err)
		assert.Equal(t, conn.GetConnectionID(), retrievedConn.GetConnectionID())
		assert.Equal(t, playerID, retrievedConn.GetPlayerID())
	})

	t.Run("should return error for non-existent player", func(t *testing.T) {
		// Given
		logger := &mockLogger{}
		cm := network.NewConnectionManager(logger)
		nonExistentPlayerID := uuid.New()

		// When
		_, err := cm.GetPlayerConnection(context.Background(), nonExistentPlayerID)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no connection found")
	})
}

func TestConnectionManager_Broadcast(t *testing.T) {
	t.Run("should broadcast to all connections", func(t *testing.T) {
		// Given
		logger := &mockLogger{}
		cm := network.NewConnectionManager(logger)

		conn1 := &mockConnection{
			playerID:     uuid.New(),
			connectionID: uuid.New().String(),
		}
		conn2 := &mockConnection{
			playerID:     uuid.New(),
			connectionID: uuid.New().String(),
		}

		err := cm.AddConnection(context.Background(), conn1)
		require.NoError(t, err)
		err = cm.AddConnection(context.Background(), conn2)
		require.NoError(t, err)

		message := &domain.NetworkMessage{
			Type:      domain.SystemNotification,
			ID:        uuid.New().String(),
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"message": "test"},
		}

		// When
		err = cm.BroadcastToAll(context.Background(), message)

		// Then
		require.NoError(t, err)
		assert.Len(t, conn1.sentMessages, 1)
		assert.Len(t, conn2.sentMessages, 1)
		assert.Equal(t, message.Type, conn1.sentMessages[0].Type)
		assert.Equal(t, message.Type, conn2.sentMessages[0].Type)
	})

	t.Run("should broadcast to specific player", func(t *testing.T) {
		// Given
		logger := &mockLogger{}
		cm := network.NewConnectionManager(logger)

		playerID := uuid.New()
		conn := &mockConnection{
			playerID:     playerID,
			connectionID: uuid.New().String(),
		}

		err := cm.AddConnection(context.Background(), conn)
		require.NoError(t, err)

		message := &domain.NetworkMessage{
			Type:      domain.SystemNotification,
			ID:        uuid.New().String(),
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"message": "test"},
		}

		// When
		err = cm.BroadcastToPlayer(context.Background(), playerID, message)

		// Then
		require.NoError(t, err)
		assert.Len(t, conn.sentMessages, 1)
		assert.Equal(t, message.Type, conn.sentMessages[0].Type)
	})

	t.Run("should handle broadcast to non-existent player", func(t *testing.T) {
		// Given
		logger := &mockLogger{}
		cm := network.NewConnectionManager(logger)
		nonExistentPlayerID := uuid.New()

		message := &domain.NetworkMessage{
			Type:      domain.SystemNotification,
			ID:        uuid.New().String(),
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"message": "test"},
		}

		// When
		err := cm.BroadcastToPlayer(context.Background(), nonExistentPlayerID, message)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no connection found")
	})
}

func TestConnectionManager_CleanupInactiveConnections(t *testing.T) {
	t.Run("should cleanup inactive connections", func(t *testing.T) {
		// Given
		logger := &mockLogger{}
		cm := network.NewConnectionManager(logger)

		// Create connections with different heartbeat times
		activeConn := &mockConnection{
			playerID:      uuid.New(),
			connectionID:  uuid.New().String(),
			lastHeartbeat: time.Now(),
		}

		inactiveConn := &mockConnection{
			playerID:      uuid.New(),
			connectionID:  uuid.New().String(),
			lastHeartbeat: time.Now().Add(-2 * time.Hour), // 2 hours ago
		}

		err := cm.AddConnection(context.Background(), activeConn)
		require.NoError(t, err)
		err = cm.AddConnection(context.Background(), inactiveConn)
		require.NoError(t, err)

		assert.Equal(t, 2, cm.GetConnectionCount())

		// When - Cleanup connections inactive for more than 1 hour
		err = cm.CleanupInactiveConnections(context.Background(), 1*time.Hour)

		// Then
		require.NoError(t, err)
		assert.Equal(t, 1, cm.GetConnectionCount()) // Only active connection remains
		assert.True(t, inactiveConn.closed)         // Inactive connection was closed
		assert.False(t, activeConn.closed)          // Active connection remains open
	})
}

// Mock connection for testing
type mockConnection struct {
	playerID      domain.PlayerId
	connectionID  string
	lastHeartbeat time.Time
	closed        bool
	sentMessages  []*domain.NetworkMessage
}

func (m *mockConnection) GetPlayerID() domain.PlayerId {
	return m.playerID
}

func (m *mockConnection) GetConnectionID() string {
	return m.connectionID
}

func (m *mockConnection) Send(ctx context.Context, message *domain.NetworkMessage) error {
	m.sentMessages = append(m.sentMessages, message)
	return nil
}

func (m *mockConnection) Receive(ctx context.Context) (*domain.NetworkMessage, error) {
	return nil, nil
}

func (m *mockConnection) Close() error {
	m.closed = true
	return nil
}

func (m *mockConnection) IsConnected() bool {
	return !m.closed
}

func (m *mockConnection) GetLastHeartbeat() time.Time {
	return m.lastHeartbeat
}

func (m *mockConnection) UpdateHeartbeat() {
	m.lastHeartbeat = time.Now()
}
