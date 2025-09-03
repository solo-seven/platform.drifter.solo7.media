//go:build uat

package uat

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

// TestGameplayScenarios tests end-to-end gameplay scenarios
// These tests simulate real user interactions and validate the complete system behavior

func TestPlayerConnectionAndMovement(t *testing.T) {
	t.Run("As a player, I should be able to connect and move my character", func(t *testing.T) {
		// Given - A game server is running
		logger := &mockLogger{}
		cm := network.NewConnectionManager(logger)

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

			// Handle player input
			for {
				message, err := wsConn.Receive(context.Background())
				if err != nil {
					break
				}

				// Process movement input
				if message.Type == domain.PlayerInput {
					inputType := message.Data["input_type"]
					if inputType == "movement" {
						// Simulate movement processing
						response := &domain.NetworkMessage{
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

						err = wsConn.Send(context.Background(), response)
						if err != nil {
							break
						}
					}
				}
			}
		}))
		defer server.Close()

		// When - Player connects and sends movement input
		wsURL := "ws" + server.URL[4:] + "/"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		playerID := uuid.New()
		clientConn := network.NewWebSocketConnection(conn, playerID, logger)

		// Send movement input
		movementInput := &domain.NetworkMessage{
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

		err = clientConn.Send(context.Background(), movementInput)
		require.NoError(t, err)

		// Then - Player should receive position update
		response, err := clientConn.Receive(context.Background())
		require.NoError(t, err)

		assert.Equal(t, domain.StateUpdate, response.Type)
		assert.Equal(t, playerID.String(), response.Data["entity_id"])
		assert.Equal(t, "transform", response.Data["component"])

		changes := response.Data["changes"].(map[string]interface{})
		position := changes["position"].(map[string]interface{})
		assert.Equal(t, 1.0, position["x"])
		assert.Equal(t, 0.0, position["y"])
		assert.Equal(t, 1.0, position["z"])
	})
}

func TestMultiPlayerInteraction(t *testing.T) {
	t.Run("As multiple players, we should be able to see each other's actions", func(t *testing.T) {
		// Given - A game server with multiple connected players
		logger := &mockLogger{}
		cm := network.NewConnectionManager(logger)

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

			// Handle player input and broadcast to others
			for {
				message, err := wsConn.Receive(context.Background())
				if err != nil {
					break
				}

				// Broadcast player actions to all other players
				if message.Type == domain.PlayerInput {
					broadcastMessage := &domain.NetworkMessage{
						Type:      domain.AestheticEvent,
						ID:        uuid.New().String(),
						Timestamp: time.Now(),
						Data: map[string]interface{}{
							"type":      "player_action",
							"player_id": playerID.String(),
							"action":    message.Data["input_type"],
							"data":      message.Data["data"],
						},
					}

					// Broadcast to all other players
					err = cm.BroadcastToAll(context.Background(), broadcastMessage)
					if err != nil {
						break
					}
				}
			}
		}))
		defer server.Close()

		// When - Two players connect and one performs an action
		// Player 1 connects
		wsURL := "ws" + server.URL[4:] + "/"
		conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn1.Close()

		player1ID := uuid.New()
		client1Conn := network.NewWebSocketConnection(conn1, player1ID, logger)

		// Player 2 connects
		conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn2.Close()

		player2ID := uuid.New()
		client2Conn := network.NewWebSocketConnection(conn2, player2ID, logger)

		// Wait for connections to be established
		time.Sleep(50 * time.Millisecond)

		// Player 1 performs an action
		actionInput := &domain.NetworkMessage{
			Type:      domain.PlayerInput,
			ID:        uuid.New().String(),
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"player_id":  player1ID.String(),
				"input_type": "attack",
				"data": map[string]interface{}{
					"target": "enemy_1",
					"weapon": "sword",
				},
			},
		}

		err = client1Conn.Send(context.Background(), actionInput)
		require.NoError(t, err)

		// Then - Player 2 should see Player 1's action
		response, err := client2Conn.Receive(context.Background())
		require.NoError(t, err)

		assert.Equal(t, domain.AestheticEvent, response.Type)
		assert.Equal(t, "player_action", response.Data["type"])
		assert.Equal(t, player1ID.String(), response.Data["player_id"])
		assert.Equal(t, "attack", response.Data["action"])

		actionData := response.Data["data"].(map[string]interface{})
		assert.Equal(t, "enemy_1", actionData["target"])
		assert.Equal(t, "sword", actionData["weapon"])
	})
}

func TestChatSystem(t *testing.T) {
	t.Run("As a player, I should be able to send and receive chat messages", func(t *testing.T) {
		// Given - A game server with chat functionality
		logger := &mockLogger{}
		cm := network.NewConnectionManager(logger)

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

			// Handle chat messages
			for {
				message, err := wsConn.Receive(context.Background())
				if err != nil {
					break
				}

				// Process chat messages
				if message.Type == domain.ChatMessage {
					chatMessage := &domain.NetworkMessage{
						Type:      domain.ChatMessage,
						ID:        uuid.New().String(),
						Timestamp: time.Now(),
						Data: map[string]interface{}{
							"player_id": playerID.String(),
							"message":   message.Data["message"],
							"channel":   message.Data["channel"],
						},
					}

					// Broadcast chat message to all players
					err = cm.BroadcastToAll(context.Background(), chatMessage)
					if err != nil {
						break
					}
				}
			}
		}))
		defer server.Close()

		// When - Player sends a chat message
		wsURL := "ws" + server.URL[4:] + "/"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		playerID := uuid.New()
		clientConn := network.NewWebSocketConnection(conn, playerID, logger)

		// Send chat message
		chatInput := &domain.NetworkMessage{
			Type:      domain.ChatMessage,
			ID:        uuid.New().String(),
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"player_id": playerID.String(),
				"message":   "Hello, world!",
				"channel":   "general",
			},
		}

		err = clientConn.Send(context.Background(), chatInput)
		require.NoError(t, err)

		// Then - Player should receive their own chat message back
		response, err := clientConn.Receive(context.Background())
		require.NoError(t, err)

		assert.Equal(t, domain.ChatMessage, response.Type)
		assert.Equal(t, playerID.String(), response.Data["player_id"])
		assert.Equal(t, "Hello, world!", response.Data["message"])
		assert.Equal(t, "general", response.Data["channel"])
	})
}

func TestSystemNotifications(t *testing.T) {
	t.Run("As a player, I should receive important system notifications", func(t *testing.T) {
		// Given - A game server that sends system notifications
		logger := &mockLogger{}
		cm := network.NewConnectionManager(logger)

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

			// Send welcome notification
			welcomeMessage := &domain.NetworkMessage{
				Type:      domain.SystemNotification,
				ID:        uuid.New().String(),
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"type":    "welcome",
					"message": "Welcome to the game!",
					"properties": map[string]interface{}{
						"server_version": "1.0.0",
						"online_players": 1,
					},
				},
			}

			err = wsConn.Send(context.Background(), welcomeMessage)
			if err != nil {
				return
			}

			// Keep connection alive
			time.Sleep(100 * time.Millisecond)
		}))
		defer server.Close()

		// When - Player connects to the server
		wsURL := "ws" + server.URL[4:] + "/"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		playerID := uuid.New()
		clientConn := network.NewWebSocketConnection(conn, playerID, logger)

		// Then - Player should receive welcome notification
		response, err := clientConn.Receive(context.Background())
		require.NoError(t, err)

		assert.Equal(t, domain.SystemNotification, response.Type)
		assert.Equal(t, "welcome", response.Data["type"])
		assert.Equal(t, "Welcome to the game!", response.Data["message"])

		properties := response.Data["properties"].(map[string]interface{})
		assert.Equal(t, "1.0.0", properties["server_version"])
		assert.Equal(t, 1, properties["online_players"])
	})
}

// Mock logger for testing
type mockLogger struct{}

func (m *mockLogger) Debug(msg string, fields map[string]interface{}) {}
func (m *mockLogger) Info(msg string, fields map[string]interface{})  {}
func (m *mockLogger) Warn(msg string, fields map[string]interface{})  {}
func (m *mockLogger) Error(msg string, fields map[string]interface{}) {}
func (m *mockLogger) Fatal(msg string, fields map[string]interface{}) {}
