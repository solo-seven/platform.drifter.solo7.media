package network

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"

	"github.com/solo-seven/platform.drifter.solo7.media/generated/proto/proto"
	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
)

// GRPCServer implements the MUDService gRPC server
type GRPCServer struct {
	proto.UnimplementedMUDServiceServer
	gameServer domain.GameServer
	logger     domain.Logger
}

// NewGRPCServer creates a new gRPC server
func NewGRPCServer(gameServer domain.GameServer, logger domain.Logger) *GRPCServer {
	return &GRPCServer{
		gameServer: gameServer,
		logger:     logger,
	}
}

// Start starts the gRPC server
func (s *GRPCServer) Start(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterMUDServiceServer(grpcServer, s)

	s.logger.Info("Starting gRPC server", map[string]interface{}{
		"port": port,
	})

	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve gRPC: %w", err)
	}

	return nil
}

// Login handles player login
func (s *GRPCServer) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	s.logger.Info("Login request", map[string]interface{}{
		"player_name":    req.PlayerName,
		"character_name": req.CharacterName,
	})

	// TODO: Implement actual login logic
	// For now, return a mock response
	return &proto.LoginResponse{
		Success:         true,
		Message:         "Login successful",
		CharacterId:     "char_123",
		CurrentLocation: "room.starting_area",
	}, nil
}

// Move handles character movement
func (s *GRPCServer) Move(ctx context.Context, req *proto.MoveRequest) (*proto.MoveResponse, error) {
	s.logger.Info("Move request", map[string]interface{}{
		"character_id": req.CharacterId,
		"direction":    req.Direction,
	})

	// TODO: Implement actual movement logic
	// For now, return a mock response
	return &proto.MoveResponse{
		Success:     true,
		Message:     fmt.Sprintf("Moved %s", req.Direction),
		NewLocation: "room.new_location",
	}, nil
}

// Look handles looking at the current room or specific objects
func (s *GRPCServer) Look(ctx context.Context, req *proto.LookRequest) (*proto.LookResponse, error) {
	s.logger.Info("Look request", map[string]interface{}{
		"character_id": req.CharacterId,
		"target":       req.Target,
	})

	// TODO: Implement actual look logic
	// For now, return a mock response
	return &proto.LookResponse{
		Description: "You are in a dimly lit room. Dust motes dance in the air.",
		Exits:       []string{"north", "south", "east"},
		Items:       []string{"rusty sword", "old book"},
		Npcs:        []string{"mysterious figure"},
		Players:     []string{"other_player"},
	}, nil
}

// Get handles picking up items
func (s *GRPCServer) Get(ctx context.Context, req *proto.GetRequest) (*proto.GetResponse, error) {
	s.logger.Info("Get request", map[string]interface{}{
		"character_id": req.CharacterId,
		"item_name":    req.ItemName,
	})

	// TODO: Implement actual get logic
	return &proto.GetResponse{
		Success: true,
		Message: fmt.Sprintf("You pick up the %s", req.ItemName),
		ItemId:  "item_" + req.ItemName,
	}, nil
}

// Drop handles dropping items
func (s *GRPCServer) Drop(ctx context.Context, req *proto.DropRequest) (*proto.DropResponse, error) {
	s.logger.Info("Drop request", map[string]interface{}{
		"character_id": req.CharacterId,
		"item_name":    req.ItemName,
	})

	// TODO: Implement actual drop logic
	return &proto.DropResponse{
		Success: true,
		Message: fmt.Sprintf("You drop the %s", req.ItemName),
	}, nil
}

// Say handles speaking
func (s *GRPCServer) Say(ctx context.Context, req *proto.SayRequest) (*proto.SayResponse, error) {
	s.logger.Info("Say request", map[string]interface{}{
		"character_id": req.CharacterId,
		"message":      req.Message,
		"channel":      req.Channel,
	})

	// TODO: Implement actual say logic and broadcast to other players
	return &proto.SayResponse{
		Success: true,
		Message: "Message sent",
	}, nil
}

// Attack handles combat
func (s *GRPCServer) Attack(ctx context.Context, req *proto.AttackRequest) (*proto.AttackResponse, error) {
	s.logger.Info("Attack request", map[string]interface{}{
		"character_id": req.CharacterId,
		"target":       req.Target,
		"weapon":       req.Weapon,
	})

	// TODO: Implement actual combat logic
	return &proto.AttackResponse{
		Success:      true,
		Message:      fmt.Sprintf("You attack %s with %s", req.Target, req.Weapon),
		Damage:       5,
		TargetStatus: "wounded",
	}, nil
}

// Use handles using items
func (s *GRPCServer) Use(ctx context.Context, req *proto.UseRequest) (*proto.UseResponse, error) {
	s.logger.Info("Use request", map[string]interface{}{
		"character_id": req.CharacterId,
		"item_name":    req.ItemName,
		"target":       req.Target,
	})

	// TODO: Implement actual use logic
	return &proto.UseResponse{
		Success: true,
		Message: fmt.Sprintf("You use the %s", req.ItemName),
		Effects: []string{"You feel refreshed"},
	}, nil
}

// Inventory handles inventory requests
func (s *GRPCServer) Inventory(ctx context.Context, req *proto.InventoryRequest) (*proto.InventoryResponse, error) {
	s.logger.Info("Inventory request", map[string]interface{}{
		"character_id": req.CharacterId,
	})

	// TODO: Implement actual inventory logic
	items := []*proto.Item{
		{
			Id:          "item_sword",
			Name:        "Rusty Sword",
			Description: "A rusty but serviceable sword",
			Type:        "weapon",
			Weight:      3,
			Properties:  map[string]string{"damage": "1d6+1"},
		},
		{
			Id:          "item_potion",
			Name:        "Health Potion",
			Description: "A red liquid that smells of herbs",
			Type:        "consumable",
			Weight:      1,
			Properties:  map[string]string{"healing": "2d4+2"},
		},
	}

	return &proto.InventoryResponse{
		Items:        items,
		Capacity:     20,
		UsedCapacity: 4,
	}, nil
}

// Talk handles NPC conversations
func (s *GRPCServer) Talk(ctx context.Context, req *proto.TalkRequest) (*proto.TalkResponse, error) {
	s.logger.Info("Talk request", map[string]interface{}{
		"character_id": req.CharacterId,
		"npc_name":     req.NpcName,
		"topic":        req.Topic,
	})

	// TODO: Implement actual NPC conversation logic
	return &proto.TalkResponse{
		Response:          "The mysterious figure looks at you with ancient eyes and says, 'Welcome, traveler.'",
		Topics:            []string{"quest", "rumors", "history"},
		ConversationEnded: false,
	}, nil
}

// AdminCommand handles admin commands
func (s *GRPCServer) AdminCommand(ctx context.Context, req *proto.AdminCommandRequest) (*proto.AdminCommandResponse, error) {
	s.logger.Info("Admin command request", map[string]interface{}{
		"admin_id":   req.AdminId,
		"command":    req.Command,
		"parameters": req.Parameters,
	})

	// TODO: Implement actual admin command logic
	return &proto.AdminCommandResponse{
		Success: true,
		Message: "Admin command executed",
		Result:  "Command result here",
	}, nil
}
