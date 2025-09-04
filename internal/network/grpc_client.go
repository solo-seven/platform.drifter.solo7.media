package network

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/solo-seven/platform.drifter.solo7.media/generated/proto"
	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
)

// GRPCClient handles gRPC communication with the server
type GRPCClient struct {
	conn   *grpc.ClientConn
	client proto.MUDServiceClient
	logger domain.Logger
}

// NewGRPCClient creates a new gRPC client
func NewGRPCClient(serverAddr string, logger domain.Logger) (*GRPCClient, error) {
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	client := proto.NewMUDServiceClient(conn)

	return &GRPCClient{
		conn:   conn,
		client: client,
		logger: logger,
	}, nil
}

// Close closes the gRPC connection
func (c *GRPCClient) Close() error {
	return c.conn.Close()
}

// Login logs in a character
func (c *GRPCClient) Login(ctx context.Context, playerName, characterName string) (*proto.LoginResponse, error) {
	req := &proto.LoginRequest{
		PlayerName:    playerName,
		CharacterName: characterName,
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return c.client.Login(ctx, req)
}

// Move moves a character
func (c *GRPCClient) Move(ctx context.Context, characterID, direction string) (*proto.MoveResponse, error) {
	req := &proto.MoveRequest{
		CharacterId: characterID,
		Direction:   direction,
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.Move(ctx, req)
}

// Look looks at the current room or a specific target
func (c *GRPCClient) Look(ctx context.Context, characterID, target string) (*proto.LookResponse, error) {
	req := &proto.LookRequest{
		CharacterId: characterID,
		Target:      target,
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.Look(ctx, req)
}

// Get picks up an item
func (c *GRPCClient) Get(ctx context.Context, characterID, itemName string) (*proto.GetResponse, error) {
	req := &proto.GetRequest{
		CharacterId: characterID,
		ItemName:    itemName,
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.Get(ctx, req)
}

// Drop drops an item
func (c *GRPCClient) Drop(ctx context.Context, characterID, itemName string) (*proto.DropResponse, error) {
	req := &proto.DropRequest{
		CharacterId: characterID,
		ItemName:    itemName,
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.Drop(ctx, req)
}

// Say says something
func (c *GRPCClient) Say(ctx context.Context, characterID, message, channel, target string) (*proto.SayResponse, error) {
	req := &proto.SayRequest{
		CharacterId: characterID,
		Message:     message,
		Channel:     channel,
		Target:      target,
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.Say(ctx, req)
}

// Attack attacks a target
func (c *GRPCClient) Attack(ctx context.Context, characterID, target, weapon string) (*proto.AttackResponse, error) {
	req := &proto.AttackRequest{
		CharacterId: characterID,
		Target:      target,
		Weapon:      weapon,
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.Attack(ctx, req)
}

// Use uses an item
func (c *GRPCClient) Use(ctx context.Context, characterID, itemName, target string) (*proto.UseResponse, error) {
	req := &proto.UseRequest{
		CharacterId: characterID,
		ItemName:    itemName,
		Target:      target,
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.Use(ctx, req)
}

// Inventory gets character inventory
func (c *GRPCClient) Inventory(ctx context.Context, characterID string) (*proto.InventoryResponse, error) {
	req := &proto.InventoryRequest{
		CharacterId: characterID,
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.Inventory(ctx, req)
}

// Talk talks to an NPC
func (c *GRPCClient) Talk(ctx context.Context, characterID, npcName, topic string) (*proto.TalkResponse, error) {
	req := &proto.TalkRequest{
		CharacterId: characterID,
		NpcName:     npcName,
		Topic:       topic,
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return c.client.Talk(ctx, req)
}
