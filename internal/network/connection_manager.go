package network

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
)

// ConnectionManagerImpl implements the ConnectionManager interface
type ConnectionManagerImpl struct {
	connections       map[string]domain.ClientConnection
	playerConnections map[domain.PlayerId]domain.ClientConnection
	regionConnections map[domain.RegionId][]domain.ClientConnection
	mutex             sync.RWMutex
	logger            domain.Logger
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(logger domain.Logger) *ConnectionManagerImpl {
	return &ConnectionManagerImpl{
		connections:       make(map[string]domain.ClientConnection),
		playerConnections: make(map[domain.PlayerId]domain.ClientConnection),
		regionConnections: make(map[domain.RegionId][]domain.ClientConnection),
		logger:            logger,
	}
}

// AddConnection adds a new client connection
func (cm *ConnectionManagerImpl) AddConnection(ctx context.Context, conn domain.ClientConnection) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	connectionID := conn.GetConnectionID()
	playerID := conn.GetPlayerID()

	// Check if connection already exists
	if existingConn, exists := cm.connections[connectionID]; exists {
		cm.logger.Warn("Connection already exists", map[string]interface{}{
			"connection_id":      connectionID,
			"existing_player_id": existingConn.GetPlayerID(),
		})
		return fmt.Errorf("connection %s already exists", connectionID)
	}

	// Check if player already has a connection
	if existingConn, exists := cm.playerConnections[playerID]; exists {
		cm.logger.Info("Player already connected, closing old connection", map[string]interface{}{
			"player_id":         playerID,
			"old_connection_id": existingConn.GetConnectionID(),
		})
		// Close the old connection
		existingConn.Close()
		delete(cm.connections, existingConn.GetConnectionID())
	}

	// Add the new connection
	cm.connections[connectionID] = conn
	cm.playerConnections[playerID] = conn

	cm.logger.Info("Connection added", map[string]interface{}{
		"connection_id": connectionID,
		"player_id":     playerID,
	})

	return nil
}

// RemoveConnection removes a client connection
func (cm *ConnectionManagerImpl) RemoveConnection(ctx context.Context, connectionID string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	conn, exists := cm.connections[connectionID]
	if !exists {
		cm.logger.Warn("Connection not found for removal", map[string]interface{}{
			"connection_id": connectionID,
		})
		return fmt.Errorf("connection %s not found", connectionID)
	}

	playerID := conn.GetPlayerID()

	// Remove from all maps
	delete(cm.connections, connectionID)
	delete(cm.playerConnections, playerID)

	// Remove from region connections
	for regionID, connections := range cm.regionConnections {
		for i, c := range connections {
			if c.GetConnectionID() == connectionID {
				cm.regionConnections[regionID] = append(connections[:i], connections[i+1:]...)
				break
			}
		}
	}

	// Close the connection
	conn.Close()

	cm.logger.Info("Connection removed", map[string]interface{}{
		"connection_id": connectionID,
		"player_id":     playerID,
	})

	return nil
}

// GetConnection retrieves a connection by ID
func (cm *ConnectionManagerImpl) GetConnection(ctx context.Context, connectionID string) (domain.ClientConnection, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	conn, exists := cm.connections[connectionID]
	if !exists {
		return nil, fmt.Errorf("connection %s not found", connectionID)
	}

	return conn, nil
}

// GetPlayerConnection retrieves a connection by player ID
func (cm *ConnectionManagerImpl) GetPlayerConnection(ctx context.Context, playerID domain.PlayerId) (domain.ClientConnection, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	conn, exists := cm.playerConnections[playerID]
	if !exists {
		return nil, fmt.Errorf("no connection found for player %s", playerID)
	}

	return conn, nil
}

// GetConnectionsInRegion retrieves all connections in a specific region
func (cm *ConnectionManagerImpl) GetConnectionsInRegion(ctx context.Context, regionID domain.RegionId) ([]domain.ClientConnection, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	connections, exists := cm.regionConnections[regionID]
	if !exists {
		return []domain.ClientConnection{}, nil
	}

	// Return a copy to avoid race conditions
	result := make([]domain.ClientConnection, len(connections))
	copy(result, connections)
	return result, nil
}

// BroadcastToAll broadcasts a message to all connected clients
func (cm *ConnectionManagerImpl) BroadcastToAll(ctx context.Context, message *domain.NetworkMessage) error {
	cm.mutex.RLock()
	connections := make([]domain.ClientConnection, 0, len(cm.connections))
	for _, conn := range cm.connections {
		connections = append(connections, conn)
	}
	cm.mutex.RUnlock()

	// Send to all connections concurrently
	var wg sync.WaitGroup
	errors := make(chan error, len(connections))

	for _, conn := range connections {
		wg.Add(1)
		go func(c domain.ClientConnection) {
			defer wg.Done()
			if err := c.Send(ctx, message); err != nil {
				errors <- fmt.Errorf("failed to send to connection %s: %w", c.GetConnectionID(), err)
			}
		}(conn)
	}

	wg.Wait()
	close(errors)

	// Collect any errors
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		cm.logger.Error("Errors during broadcast", map[string]interface{}{
			"error_count": len(errs),
			"errors":      errs,
		})
		return fmt.Errorf("broadcast failed with %d errors", len(errs))
	}

	cm.logger.Debug("Broadcast completed", map[string]interface{}{
		"message_type":    message.Type,
		"recipient_count": len(connections),
	})

	return nil
}

// BroadcastToRegion broadcasts a message to all clients in a specific region
func (cm *ConnectionManagerImpl) BroadcastToRegion(ctx context.Context, regionID domain.RegionId, message *domain.NetworkMessage) error {
	connections, err := cm.GetConnectionsInRegion(ctx, regionID)
	if err != nil {
		return fmt.Errorf("failed to get connections in region: %w", err)
	}

	if len(connections) == 0 {
		cm.logger.Debug("No connections in region for broadcast", map[string]interface{}{
			"region_id": regionID,
		})
		return nil
	}

	// Send to all connections in region concurrently
	var wg sync.WaitGroup
	errors := make(chan error, len(connections))

	for _, conn := range connections {
		wg.Add(1)
		go func(c domain.ClientConnection) {
			defer wg.Done()
			if err := c.Send(ctx, message); err != nil {
				errors <- fmt.Errorf("failed to send to connection %s: %w", c.GetConnectionID(), err)
			}
		}(conn)
	}

	wg.Wait()
	close(errors)

	// Collect any errors
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		cm.logger.Error("Errors during region broadcast", map[string]interface{}{
			"region_id":   regionID,
			"error_count": len(errs),
			"errors":      errs,
		})
		return fmt.Errorf("region broadcast failed with %d errors", len(errs))
	}

	cm.logger.Debug("Region broadcast completed", map[string]interface{}{
		"region_id":       regionID,
		"message_type":    message.Type,
		"recipient_count": len(connections),
	})

	return nil
}

// BroadcastToPlayer broadcasts a message to a specific player
func (cm *ConnectionManagerImpl) BroadcastToPlayer(ctx context.Context, playerID domain.PlayerId, message *domain.NetworkMessage) error {
	conn, err := cm.GetPlayerConnection(ctx, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player connection: %w", err)
	}

	if err := conn.Send(ctx, message); err != nil {
		return fmt.Errorf("failed to send message to player %s: %w", playerID, err)
	}

	cm.logger.Debug("Player broadcast completed", map[string]interface{}{
		"player_id":    playerID,
		"message_type": message.Type,
	})

	return nil
}

// UpdatePlayerRegion updates the region association for a player
func (cm *ConnectionManagerImpl) UpdatePlayerRegion(ctx context.Context, playerID domain.PlayerId, regionID domain.RegionId) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	conn, exists := cm.playerConnections[playerID]
	if !exists {
		return fmt.Errorf("no connection found for player %s", playerID)
	}

	// Remove from old region
	for oldRegionID, connections := range cm.regionConnections {
		for i, c := range connections {
			if c.GetConnectionID() == conn.GetConnectionID() {
				cm.regionConnections[oldRegionID] = append(connections[:i], connections[i+1:]...)
				break
			}
		}
	}

	// Add to new region
	cm.regionConnections[regionID] = append(cm.regionConnections[regionID], conn)

	cm.logger.Debug("Player region updated", map[string]interface{}{
		"player_id": playerID,
		"region_id": regionID,
	})

	return nil
}

// GetConnectionCount returns the total number of active connections
func (cm *ConnectionManagerImpl) GetConnectionCount() int {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return len(cm.connections)
}

// GetPlayerCount returns the total number of connected players
func (cm *ConnectionManagerImpl) GetPlayerCount() int {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return len(cm.playerConnections)
}

// CleanupInactiveConnections removes connections that haven't sent heartbeats
func (cm *ConnectionManagerImpl) CleanupInactiveConnections(ctx context.Context, maxInactiveTime time.Duration) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	now := time.Now()
	var inactiveConnections []string

	for connectionID, conn := range cm.connections {
		if now.Sub(conn.GetLastHeartbeat()) > maxInactiveTime {
			inactiveConnections = append(inactiveConnections, connectionID)
		}
	}

	for _, connectionID := range inactiveConnections {
		cm.logger.Info("Removing inactive connection", map[string]interface{}{
			"connection_id": connectionID,
		})

		conn := cm.connections[connectionID]
		playerID := conn.GetPlayerID()

		delete(cm.connections, connectionID)
		delete(cm.playerConnections, playerID)

		// Remove from region connections
		for regionID, connections := range cm.regionConnections {
			for i, c := range connections {
				if c.GetConnectionID() == connectionID {
					cm.regionConnections[regionID] = append(connections[:i], connections[i+1:]...)
					break
				}
			}
		}

		conn.Close()
	}

	if len(inactiveConnections) > 0 {
		cm.logger.Info("Cleaned up inactive connections", map[string]interface{}{
			"removed_count": len(inactiveConnections),
		})
	}

	return nil
}
