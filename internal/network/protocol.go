package network

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	platform_drifter_solo7_media "github.com/solo-seven/platform.drifter.solo7.media/generated/proto"
	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
	protobuf "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// WebSocketProtocol implements the NetworkProtocol interface using WebSockets
type WebSocketProtocol struct {
	upgrader          websocket.Upgrader
	logger            domain.Logger
	connectionManager domain.ConnectionManager
}

// NewWebSocketProtocol creates a new WebSocket protocol instance
func NewWebSocketProtocol(logger domain.Logger) *WebSocketProtocol {
	return &WebSocketProtocol{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
		logger: logger,
	}
}

// SetConnectionManager sets the connection manager for the protocol
func (p *WebSocketProtocol) SetConnectionManager(cm domain.ConnectionManager) {
	p.connectionManager = cm
}

// WebSocketConnection implements ClientConnection interface
type WebSocketConnection struct {
	conn          *websocket.Conn
	playerID      domain.PlayerId
	connectionID  string
	lastHeartbeat time.Time
	logger        domain.Logger
}

// NewWebSocketConnection creates a new WebSocket connection
func NewWebSocketConnection(conn *websocket.Conn, playerID domain.PlayerId, logger domain.Logger) *WebSocketConnection {
	return &WebSocketConnection{
		conn:          conn,
		playerID:      playerID,
		connectionID:  uuid.New().String(),
		lastHeartbeat: time.Now(),
		logger:        logger,
	}
}

// GetPlayerID returns the player ID for this connection
func (c *WebSocketConnection) GetPlayerID() domain.PlayerId {
	return c.playerID
}

// GetConnectionID returns the unique connection ID
func (c *WebSocketConnection) GetConnectionID() string {
	return c.connectionID
}

// Send sends a message to the client
func (c *WebSocketConnection) Send(ctx context.Context, message *domain.NetworkMessage) error {
	if c.conn == nil {
		return fmt.Errorf("connection is nil")
	}

	// Convert domain message to protobuf
	pbMessage, err := c.domainToProtoMessage(message)
	if err != nil {
		return fmt.Errorf("failed to convert message to protobuf: %w", err)
	}

	// Serialize protobuf message
	data, err := protobuf.Marshal(pbMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal protobuf message: %w", err)
	}

	// Apply a write deadline to avoid indefinite blocking
	// Prefer context deadline if provided; otherwise, use a sensible default.
	if dl, ok := ctx.Deadline(); ok {
		_ = c.conn.SetWriteDeadline(dl)
	} else {
		_ = c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	}
	defer func() { _ = c.conn.SetWriteDeadline(time.Time{}) }()

	// Send via WebSocket
	err = c.conn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		return fmt.Errorf("failed to send WebSocket message: %w", err)
	}

	return nil
}

// Receive receives a message from the client
func (c *WebSocketConnection) Receive(ctx context.Context) (*domain.NetworkMessage, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("connection is nil")
	}

	// Set read deadline
	deadline := time.Now().Add(30 * time.Second)
	err := c.conn.SetReadDeadline(deadline)
	if err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	// Read message
	messageType, data, err := c.conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to read WebSocket message: %w", err)
	}

	if messageType != websocket.BinaryMessage {
		return nil, fmt.Errorf("unexpected message type: %d", messageType)
	}

	// Deserialize protobuf message
	var pbMessage platform_drifter_solo7_media.NetworkMessage
	err = protobuf.Unmarshal(data, &pbMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal protobuf message: %w", err)
	}

	// Convert protobuf to domain message
	domainMessage, err := c.protoToDomainMessage(&pbMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to convert protobuf to domain message: %w", err)
	}

	return domainMessage, nil
}

// Close closes the connection
func (c *WebSocketConnection) Close() error {
	if c.conn == nil {
		return nil
	}
	err := c.conn.Close()
	c.conn = nil
	return err
}

// IsConnected checks if the connection is still active
func (c *WebSocketConnection) IsConnected() bool {
	return c.conn != nil
}

// GetLastHeartbeat returns the last heartbeat time
func (c *WebSocketConnection) GetLastHeartbeat() time.Time {
	return c.lastHeartbeat
}

// UpdateHeartbeat updates the heartbeat timestamp
func (c *WebSocketConnection) UpdateHeartbeat() {
	c.lastHeartbeat = time.Now()
}

// domainToProtoMessage converts a domain NetworkMessage to protobuf
func (c *WebSocketConnection) domainToProtoMessage(msg *domain.NetworkMessage) (*platform_drifter_solo7_media.NetworkMessage, error) {
	pbMsg := &platform_drifter_solo7_media.NetworkMessage{
		Id:        msg.ID,
		Timestamp: timestamppb.New(msg.Timestamp),
	}

	// Convert message type
	switch msg.Type {
	case domain.PlayerInput:
		pbMsg.Type = platform_drifter_solo7_media.MessageType_PLAYER_INPUT
	case domain.ChatMessage:
		pbMsg.Type = platform_drifter_solo7_media.MessageType_CHAT_MESSAGE
	case domain.AdminCommand:
		pbMsg.Type = platform_drifter_solo7_media.MessageType_ADMIN_COMMAND
	case domain.StateUpdate:
		pbMsg.Type = platform_drifter_solo7_media.MessageType_STATE_UPDATE
	case domain.AestheticEvent:
		pbMsg.Type = platform_drifter_solo7_media.MessageType_AESTHETIC_EVENT
	case domain.SystemNotification:
		pbMsg.Type = platform_drifter_solo7_media.MessageType_SYSTEM_NOTIFICATION
	case domain.Heartbeat:
		pbMsg.Type = platform_drifter_solo7_media.MessageType_HEARTBEAT
	case domain.ConnectionNegotiation:
		pbMsg.Type = platform_drifter_solo7_media.MessageType_CONNECTION_NEGOTIATION
	default:
		return nil, fmt.Errorf("unknown message type: %v", msg.Type)
	}

	// Convert data based on message type
	data, err := c.convertMessageData(msg.Type, msg.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert message data: %w", err)
	}

	pbMsg.Data = data
	return pbMsg, nil
}

// protoToDomainMessage converts a protobuf NetworkMessage to domain
func (c *WebSocketConnection) protoToDomainMessage(pbMsg *platform_drifter_solo7_media.NetworkMessage) (*domain.NetworkMessage, error) {
	msg := &domain.NetworkMessage{
		ID:        pbMsg.Id,
		Timestamp: pbMsg.Timestamp.AsTime(),
	}

	// Convert message type
	switch pbMsg.Type {
	case platform_drifter_solo7_media.MessageType_PLAYER_INPUT:
		msg.Type = domain.PlayerInput
	case platform_drifter_solo7_media.MessageType_CHAT_MESSAGE:
		msg.Type = domain.ChatMessage
	case platform_drifter_solo7_media.MessageType_ADMIN_COMMAND:
		msg.Type = domain.AdminCommand
	case platform_drifter_solo7_media.MessageType_STATE_UPDATE:
		msg.Type = domain.StateUpdate
	case platform_drifter_solo7_media.MessageType_AESTHETIC_EVENT:
		msg.Type = domain.AestheticEvent
	case platform_drifter_solo7_media.MessageType_SYSTEM_NOTIFICATION:
		msg.Type = domain.SystemNotification
	case platform_drifter_solo7_media.MessageType_HEARTBEAT:
		msg.Type = domain.Heartbeat
	case platform_drifter_solo7_media.MessageType_CONNECTION_NEGOTIATION:
		msg.Type = domain.ConnectionNegotiation
	default:
		return nil, fmt.Errorf("unknown protobuf message type: %v", pbMsg.Type)
	}

	// Convert data based on message type
	data, err := c.convertProtoData(pbMsg.Type, pbMsg.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert protobuf data: %w", err)
	}

	msg.Data = data
	return msg, nil
}

// convertMessageData converts domain message data to protobuf Any
func (c *WebSocketConnection) convertMessageData(msgType domain.MessageType, data map[string]interface{}) (*anypb.Any, error) {
	switch msgType {
	case domain.PlayerInput:
		return c.convertPlayerInputData(data)
	case domain.ChatMessage:
		return c.convertChatMessageData(data)
	case domain.AdminCommand:
		return c.convertAdminCommandData(data)
	case domain.StateUpdate:
		return c.convertStateUpdateData(data)
	case domain.Heartbeat:
		return c.convertHeartbeatData(data)
	case domain.ConnectionNegotiation:
		return c.convertConnectionNegotiationData(data)
	default:
		// For unknown types, serialize as JSON
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data as JSON: %w", err)
		}
		return anypb.New(&platform_drifter_solo7_media.SystemNotification{
			Type:      "unknown",
			Message:   string(jsonData),
			Timestamp: timestamppb.Now(),
		})
	}
}

// convertProtoData converts protobuf Any to domain message data
func (c *WebSocketConnection) convertProtoData(msgType platform_drifter_solo7_media.MessageType, data *anypb.Any) (map[string]interface{}, error) {
	switch msgType {
	case platform_drifter_solo7_media.MessageType_PLAYER_INPUT:
		return c.convertProtoPlayerInputData(data)
	case platform_drifter_solo7_media.MessageType_CHAT_MESSAGE:
		return c.convertProtoChatMessageData(data)
	case platform_drifter_solo7_media.MessageType_ADMIN_COMMAND:
		return c.convertProtoAdminCommandData(data)
	case platform_drifter_solo7_media.MessageType_STATE_UPDATE:
		return c.convertProtoStateUpdateData(data)
	case platform_drifter_solo7_media.MessageType_HEARTBEAT:
		return c.convertProtoHeartbeatData(data)
	case platform_drifter_solo7_media.MessageType_CONNECTION_NEGOTIATION:
		return c.convertProtoConnectionNegotiationData(data)
	default:
		// For unknown types, return empty map
		return make(map[string]interface{}), nil
	}
}

// Helper methods for data conversion
func (c *WebSocketConnection) convertPlayerInputData(data map[string]interface{}) (*anypb.Any, error) {
	playerID, ok := data["player_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid player_id")
	}

	inputType, ok := data["input_type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid input_type")
	}

	// Convert properties to string map
	properties := make(map[string]string)
	if props, ok := data["data"].(map[string]interface{}); ok {
		for k, v := range props {
			properties[k] = fmt.Sprintf("%v", v)
		}
	}

	pbData := &platform_drifter_solo7_media.PlayerInputMessage{
		PlayerId:  playerID,
		InputType: inputType,
		Data:      properties,
	}

	return anypb.New(pbData)
}

func (c *WebSocketConnection) convertProtoPlayerInputData(data *anypb.Any) (map[string]interface{}, error) {
	var pbData platform_drifter_solo7_media.PlayerInputMessage
	err := data.UnmarshalTo(&pbData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal PlayerInputMessage: %w", err)
	}

	result := map[string]interface{}{
		"player_id":  pbData.PlayerId,
		"input_type": pbData.InputType,
		"data":       pbData.Data,
	}

	return result, nil
}

// Additional conversion methods would be implemented here for other message types
// For brevity, I'll implement a few key ones

func (c *WebSocketConnection) convertChatMessageData(data map[string]interface{}) (*anypb.Any, error) {
	playerID, _ := data["player_id"].(string)
	message, _ := data["message"].(string)
	channel, _ := data["channel"].(string)

	pbData := &platform_drifter_solo7_media.ChatMessage{
		PlayerId:  playerID,
		Message:   message,
		Channel:   channel,
		Timestamp: timestamppb.Now(),
	}

	return anypb.New(pbData)
}

func (c *WebSocketConnection) convertProtoChatMessageData(data *anypb.Any) (map[string]interface{}, error) {
	var pbData platform_drifter_solo7_media.ChatMessage
	err := data.UnmarshalTo(&pbData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal ChatMessage: %w", err)
	}

	result := map[string]interface{}{
		"player_id": pbData.PlayerId,
		"message":   pbData.Message,
		"channel":   pbData.Channel,
		"timestamp": pbData.Timestamp.AsTime(),
	}

	return result, nil
}

func (c *WebSocketConnection) convertHeartbeatData(data map[string]interface{}) (*anypb.Any, error) {
	connectionID, _ := data["connection_id"].(string)

	pbData := &platform_drifter_solo7_media.Heartbeat{
		ConnectionId: connectionID,
		Timestamp:    timestamppb.Now(),
	}

	return anypb.New(pbData)
}

func (c *WebSocketConnection) convertProtoHeartbeatData(data *anypb.Any) (map[string]interface{}, error) {
	var pbData platform_drifter_solo7_media.Heartbeat
	err := data.UnmarshalTo(&pbData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal Heartbeat: %w", err)
	}

	result := map[string]interface{}{
		"connection_id": pbData.ConnectionId,
		"timestamp":     pbData.Timestamp.AsTime(),
	}

	return result, nil
}

// Placeholder implementations for other conversion methods
func (c *WebSocketConnection) convertAdminCommandData(data map[string]interface{}) (*anypb.Any, error) {
	// Implementation would go here
	return anypb.New(&platform_drifter_solo7_media.AdminCommand{})
}

func (c *WebSocketConnection) convertProtoAdminCommandData(data *anypb.Any) (map[string]interface{}, error) {
	// Implementation would go here
	return make(map[string]interface{}), nil
}

func (c *WebSocketConnection) convertStateUpdateData(data map[string]interface{}) (*anypb.Any, error) {
	// Implementation would go here
	return anypb.New(&platform_drifter_solo7_media.StateUpdateMessage{})
}

func (c *WebSocketConnection) convertProtoStateUpdateData(data *anypb.Any) (map[string]interface{}, error) {
	// Implementation would go here
	return make(map[string]interface{}), nil
}

func (c *WebSocketConnection) convertConnectionNegotiationData(data map[string]interface{}) (*anypb.Any, error) {
	// Implementation would go here
	return anypb.New(&platform_drifter_solo7_media.ConnectionNegotiation{})
}

func (c *WebSocketConnection) convertProtoConnectionNegotiationData(data *anypb.Any) (map[string]interface{}, error) {
	// Implementation would go here
	return make(map[string]interface{}), nil
}

// NetworkProtocol interface methods for WebSocketProtocol

// SendMessage sends a message to a specific connection
func (p *WebSocketProtocol) SendMessage(ctx context.Context, message *domain.NetworkMessage) error {
	// This method is not directly applicable to WebSocketProtocol
	// as it needs a specific connection. Use connection.Send() instead.
	return fmt.Errorf("SendMessage requires a specific connection, use connection.Send() instead")
}

// ReceiveMessage receives a message from a specific connection
func (p *WebSocketProtocol) ReceiveMessage(ctx context.Context) (*domain.NetworkMessage, error) {
	// This method is not directly applicable to WebSocketProtocol
	// as it needs a specific connection. Use connection.Receive() instead.
	return nil, fmt.Errorf("ReceiveMessage requires a specific connection, use connection.Receive() instead")
}

// BroadcastToRegion broadcasts a message to all connections in a region
func (p *WebSocketProtocol) BroadcastToRegion(ctx context.Context, regionID domain.RegionId, message *domain.NetworkMessage) error {
	if p.connectionManager == nil {
		return fmt.Errorf("connection manager not set")
	}
	return p.connectionManager.BroadcastToRegion(ctx, regionID, message)
}

// BroadcastToPlayer broadcasts a message to a specific player
func (p *WebSocketProtocol) BroadcastToPlayer(ctx context.Context, playerID domain.PlayerId, message *domain.NetworkMessage) error {
	if p.connectionManager == nil {
		return fmt.Errorf("connection manager not set")
	}
	return p.connectionManager.BroadcastToPlayer(ctx, playerID, message)
}

// BroadcastToArea broadcasts a message to connections in a specific area
func (p *WebSocketProtocol) BroadcastToArea(ctx context.Context, regionID domain.RegionId, center domain.Vector3, radius float64, message *domain.NetworkMessage) error {
	if p.connectionManager == nil {
		return fmt.Errorf("connection manager not set")
	}

	// Get connections in the region
	connections, err := p.connectionManager.GetConnectionsInRegion(ctx, regionID)
	if err != nil {
		return fmt.Errorf("failed to get connections in region: %w", err)
	}

	// Send to all connections in region (simplified implementation)
	// TODO: Implement proper area-based filtering
	for _, conn := range connections {
		if err := conn.Send(ctx, message); err != nil {
			p.logger.Error("Failed to send message to connection", map[string]interface{}{
				"connection_id": conn.GetConnectionID(),
				"error":         err,
			})
		}
	}

	return nil
}
