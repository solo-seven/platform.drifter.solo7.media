package network

import (
	"context"
	"fmt"
	"net"
	"strings"

	"google.golang.org/grpc"

	"github.com/solo-seven/platform.drifter.solo7.media/generated/proto"
	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
	"github.com/solo-seven/platform.drifter.solo7.media/internal/server"
)

// GRPCServer implements the MUDService gRPC server
type GRPCServer struct {
	proto.UnimplementedMUDServiceServer
	gameServer    domain.GameServer
	logger        domain.Logger
	grpcServer    *grpc.Server
	listener      net.Listener
	contentLoader *server.ContentLoader
}

// NewGRPCServer creates a new gRPC server
func NewGRPCServer(gameServer domain.GameServer, logger domain.Logger, contentLoader *server.ContentLoader) *GRPCServer {
	return &GRPCServer{
		gameServer:    gameServer,
		logger:        logger,
		contentLoader: contentLoader,
	}
}

// Start starts the gRPC server
func (s *GRPCServer) Start(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	s.listener = lis
	s.grpcServer = grpc.NewServer()
	proto.RegisterMUDServiceServer(s.grpcServer, s)

	s.logger.Info("Starting gRPC server", map[string]interface{}{
		"port": port,
	})

	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve gRPC: %w", err)
	}

	return nil
}

// Stop gracefully stops the gRPC server
func (s *GRPCServer) Stop() error {
	if s.grpcServer == nil {
		s.logger.Warn("gRPC server is not running", map[string]interface{}{})
		return nil
	}

	s.logger.Info("Stopping gRPC server", map[string]interface{}{})

	// Gracefully stop the gRPC server
	s.grpcServer.GracefulStop()

	// Close the listener if it's still open
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			// Check if the error is because the listener is already closed
			if !strings.Contains(err.Error(), "use of closed network connection") {
				s.logger.Error("Failed to close listener", map[string]interface{}{
					"error": err.Error(),
				})
				return fmt.Errorf("failed to close listener: %w", err)
			}
			// If it's already closed, that's fine - just log it
			s.logger.Info("Listener was already closed", map[string]interface{}{})
		}
	}

	s.logger.Info("gRPC server stopped successfully", map[string]interface{}{})
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

	// If no target specified, show current location
	if req.Target == "" {
		// For now, default to Nexus City - Stable Quarter
		location := s.contentLoader.GetLocation("district.stable_quarter")
		if location == nil {
			// Fallback to any available location
			locations := s.contentLoader.GetAllLocations()
			for _, loc := range locations {
				location = loc
				break
			}
		}

		if location != nil {
			return &proto.LookResponse{
				Description: location.Description,
				Exits:       location.Exits,
				Items:       []string{}, // TODO: Get actual items in location
				Npcs:        []string{}, // TODO: Get actual NPCs in location
				Players:     []string{}, // TODO: Get actual players in location
			}, nil
		}
	}

	// Look at specific target
	// TODO: Implement looking at specific items, NPCs, etc.
	return &proto.LookResponse{
		Description: "You don't see that here.",
		Exits:       []string{},
		Items:       []string{},
		Npcs:        []string{},
		Players:     []string{},
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

	// Get items from content loader
	allItems := s.contentLoader.GetAllItems()
	items := make([]*proto.Item, 0, len(allItems))

	// Convert loaded items to proto format
	for _, itemData := range allItems {
		// Convert properties to string map
		properties := make(map[string]string)
		for key, value := range itemData.Properties {
			if str, ok := value.(string); ok {
				properties[key] = str
			} else {
				properties[key] = fmt.Sprintf("%v", value)
			}
		}

		items = append(items, &proto.Item{
			Id:          itemData.ID,
			Name:        itemData.Name,
			Description: itemData.Description,
			Type:        itemData.Type,
			Weight:      itemData.Weight,
			Properties:  properties,
		})
	}

	// Calculate capacity
	totalWeight := int32(0)
	for _, item := range items {
		totalWeight += item.Weight
	}

	return &proto.InventoryResponse{
		Items:        items,
		Capacity:     100, // Default capacity
		UsedCapacity: totalWeight,
	}, nil
}

// Talk handles NPC conversations
func (s *GRPCServer) Talk(ctx context.Context, req *proto.TalkRequest) (*proto.TalkResponse, error) {
	s.logger.Info("Talk request", map[string]interface{}{
		"character_id": req.CharacterId,
		"npc_name":     req.NpcName,
		"topic":        req.Topic,
	})

	// Find NPC by name
	var npcData *server.NPCData
	allNPCs := s.contentLoader.GetAllNPCs()
	for _, npc := range allNPCs {
		if npc.Name == req.NpcName {
			npcData = npc
			break
		}
	}

	if npcData == nil {
		return &proto.TalkResponse{
			Response:          "You don't see anyone by that name here.",
			Topics:            []string{},
			ConversationEnded: true,
		}, nil
	}

	// Generate response based on NPC and topic
	response := fmt.Sprintf("The %s looks at you and says, 'Greetings, traveler.'", npcData.Name)
	topics := []string{"quest", "rumors", "history", "trade"}

	if req.Topic != "" {
		switch req.Topic {
		case "quest":
			response = "The " + npcData.Name + " mentions there might be work available for brave adventurers."
		case "rumors":
			response = "The " + npcData.Name + " whispers about strange happenings in the Fracture Wastes."
		case "history":
			response = "The " + npcData.Name + " speaks of the ancient times before the reality fractures."
		case "trade":
			response = "The " + npcData.Name + " offers to trade goods, though prices vary with reality stability."
		default:
			response = "The " + npcData.Name + " doesn't seem interested in discussing that topic."
		}
	}

	return &proto.TalkResponse{
		Response:          response,
		Topics:            topics,
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
