package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/solo-seven/platform.drifter.solo7.media/generated/proto"
	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
	"github.com/solo-seven/platform.drifter.solo7.media/internal/network"
)

// MUDClient represents a MUD client using both gRPC and WebSocket
type MUDClient struct {
	grpcClient    *network.GRPCClient
	wsConn        *websocket.Conn
	characterID   string
	characterName string
	logger        domain.Logger
}

// NewMUDClient creates a new MUD client
func NewMUDClient(grpcAddr, wsAddr string, logger domain.Logger) (*MUDClient, error) {
	// Create gRPC client
	grpcClient, err := network.NewGRPCClient(grpcAddr, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	// Connect to WebSocket
	u, err := url.Parse(wsAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid WebSocket URL: %w", err)
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	return &MUDClient{
		grpcClient: grpcClient,
		wsConn:     conn,
		logger:     logger,
	}, nil
}

// Login logs in a character
func (c *MUDClient) Login(ctx context.Context, playerName, characterName string) error {
	resp, err := c.grpcClient.Login(ctx, playerName, characterName)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("login failed: %s", resp.Message)
	}

	c.characterID = resp.CharacterId
	c.characterName = characterName

	fmt.Printf("Logged in as %s (ID: %s)\n", characterName, c.characterID)
	fmt.Printf("Current location: %s\n", resp.CurrentLocation)

	return nil
}

// Start starts the client
func (c *MUDClient) Start(ctx context.Context) error {
	// Start WebSocket message receiver
	go c.receiveMessages(ctx)

	// Start interactive command loop
	go c.interactiveLoop(ctx)

	// Wait for context cancellation
	<-ctx.Done()
	return nil
}

// Close closes the client
func (c *MUDClient) Close() error {
	if c.grpcClient != nil {
		c.grpcClient.Close()
	}
	if c.wsConn != nil {
		c.wsConn.Close()
	}
	return nil
}

// receiveMessages receives messages from WebSocket
func (c *MUDClient) receiveMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, message, err := c.wsConn.ReadMessage()
			if err != nil {
				if ctx.Err() == nil {
					fmt.Printf("Error receiving message: %v\n", err)
				}
				return
			}

			var gameEvent proto.GameEvent
			if err := json.Unmarshal(message, &gameEvent); err != nil {
				fmt.Printf("Error parsing message: %v\n", err)
				continue
			}

			c.handleGameEvent(&gameEvent)
		}
	}
}

// handleGameEvent handles a game event
func (c *MUDClient) handleGameEvent(event *proto.GameEvent) {
	switch e := event.Event.(type) {
	case *proto.GameEvent_ChatMessage:
		msg := e.ChatMessage
		fmt.Printf("💬 [%s] %s: %s\n", msg.Channel, msg.PlayerName, msg.Message)
	case *proto.GameEvent_PlayerJoined:
		player := e.PlayerJoined
		fmt.Printf("👋 %s joined the game\n", player.PlayerName)
	case *proto.GameEvent_PlayerLeft:
		player := e.PlayerLeft
		fmt.Printf("👋 %s left the game\n", player.PlayerName)
	case *proto.GameEvent_RoomDescription:
		room := e.RoomDescription
		fmt.Printf("🏠 %s\n", room.Description)
		if len(room.Exits) > 0 {
			fmt.Printf("Exits: %s\n", strings.Join(room.Exits, ", "))
		}
		if len(room.Items) > 0 {
			fmt.Printf("Items: %s\n", strings.Join(room.Items, ", "))
		}
		if len(room.Npcs) > 0 {
			fmt.Printf("NPCs: %s\n", strings.Join(room.Npcs, ", "))
		}
		if len(room.Players) > 0 {
			fmt.Printf("Players: %s\n", strings.Join(room.Players, ", "))
		}
	case *proto.GameEvent_PlayerStateUpdate:
		state := e.PlayerStateUpdate
		fmt.Printf("📊 Health: %d/%d, Mana: %d/%d\n", state.Health, state.MaxHealth, state.Mana, state.MaxMana)
	case *proto.GameEvent_CombatLog:
		combat := e.CombatLog
		fmt.Printf("⚔️ %s attacks %s for %d damage (%s)\n", combat.AttackerId, combat.TargetId, combat.Damage, combat.Result)
	case *proto.GameEvent_SystemNotification:
		notif := e.SystemNotification
		fmt.Printf("🔔 %s: %s\n", notif.Type, notif.Message)
	case *proto.GameEvent_AestheticEvent:
		aesthetic := e.AestheticEvent
		fmt.Printf("🎭 %s: %s\n", aesthetic.Type, aesthetic.Description)
	default:
		fmt.Printf("📨 Unknown event type: %T\n", event.Event)
	}
}

// interactiveLoop handles user input
func (c *MUDClient) interactiveLoop(ctx context.Context) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("\nAvailable commands:")
	fmt.Println("  look [target]     - Look at room or target")
	fmt.Println("  move <direction>  - Move in direction (north, south, east, west, up, down)")
	fmt.Println("  get <item>        - Pick up an item")
	fmt.Println("  drop <item>       - Drop an item")
	fmt.Println("  inventory         - Show inventory")
	fmt.Println("  say <message>     - Say something")
	fmt.Println("  attack <target>   - Attack a target")
	fmt.Println("  use <item> [target] - Use an item")
	fmt.Println("  talk <npc> [topic] - Talk to an NPC")
	fmt.Println("  quit              - Quit the game")
	fmt.Println("\nType a command:")

	for {
		select {
		case <-ctx.Done():
			return
		default:
			fmt.Print("> ")
			if !scanner.Scan() {
				return
			}

			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				continue
			}

			if err := c.handleCommand(ctx, input); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		}
	}
}

// handleCommand handles a user command
func (c *MUDClient) handleCommand(ctx context.Context, input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	command := parts[0]
	args := parts[1:]

	switch command {
	case "look":
		target := ""
		if len(args) > 0 {
			target = strings.Join(args, " ")
		}
		return c.handleLook(ctx, target)
	case "move":
		if len(args) == 0 {
			fmt.Println("Usage: move <direction>")
			return nil
		}
		return c.handleMove(ctx, args[0])
	case "get":
		if len(args) == 0 {
			fmt.Println("Usage: get <item>")
			return nil
		}
		return c.handleGet(ctx, strings.Join(args, " "))
	case "drop":
		if len(args) == 0 {
			fmt.Println("Usage: drop <item>")
			return nil
		}
		return c.handleDrop(ctx, strings.Join(args, " "))
	case "inventory":
		return c.handleInventory(ctx)
	case "say":
		if len(args) == 0 {
			fmt.Println("Usage: say <message>")
			return nil
		}
		return c.handleSay(ctx, strings.Join(args, " "))
	case "attack":
		if len(args) == 0 {
			fmt.Println("Usage: attack <target>")
			return nil
		}
		return c.handleAttack(ctx, strings.Join(args, " "))
	case "use":
		if len(args) == 0 {
			fmt.Println("Usage: use <item> [target]")
			return nil
		}
		target := ""
		if len(args) > 1 {
			target = strings.Join(args[1:], " ")
		}
		return c.handleUse(ctx, args[0], target)
	case "talk":
		if len(args) == 0 {
			fmt.Println("Usage: talk <npc> [topic]")
			return nil
		}
		topic := ""
		if len(args) > 1 {
			topic = strings.Join(args[1:], " ")
		}
		return c.handleTalk(ctx, args[0], topic)
	case "quit":
		fmt.Println("Goodbye!")
		os.Exit(0)
		return nil // This will never be reached, but satisfies the compiler
	default:
		// Treat as say command
		return c.handleSay(ctx, input)
	}
}

// handleLook handles the look command
func (c *MUDClient) handleLook(ctx context.Context, target string) error {
	resp, err := c.grpcClient.Look(ctx, c.characterID, target)
	if err != nil {
		return err
	}

	fmt.Println(resp.Description)
	return nil
}

// handleMove handles the move command
func (c *MUDClient) handleMove(ctx context.Context, direction string) error {
	resp, err := c.grpcClient.Move(ctx, c.characterID, direction)
	if err != nil {
		return err
	}

	fmt.Println(resp.Message)
	return nil
}

// handleGet handles the get command
func (c *MUDClient) handleGet(ctx context.Context, itemName string) error {
	resp, err := c.grpcClient.Get(ctx, c.characterID, itemName)
	if err != nil {
		return err
	}

	fmt.Println(resp.Message)
	return nil
}

// handleDrop handles the drop command
func (c *MUDClient) handleDrop(ctx context.Context, itemName string) error {
	resp, err := c.grpcClient.Drop(ctx, c.characterID, itemName)
	if err != nil {
		return err
	}

	fmt.Println(resp.Message)
	return nil
}

// handleInventory handles the inventory command
func (c *MUDClient) handleInventory(ctx context.Context) error {
	resp, err := c.grpcClient.Inventory(ctx, c.characterID)
	if err != nil {
		return err
	}

	fmt.Printf("Inventory (%d/%d):\n", resp.UsedCapacity, resp.Capacity)
	for _, item := range resp.Items {
		fmt.Printf("  - %s: %s\n", item.Name, item.Description)
	}
	return nil
}

// handleSay handles the say command
func (c *MUDClient) handleSay(ctx context.Context, message string) error {
	resp, err := c.grpcClient.Say(ctx, c.characterID, message, "say", "")
	if err != nil {
		return err
	}

	fmt.Println(resp.Message)
	return nil
}

// handleAttack handles the attack command
func (c *MUDClient) handleAttack(ctx context.Context, target string) error {
	resp, err := c.grpcClient.Attack(ctx, c.characterID, target, "")
	if err != nil {
		return err
	}

	fmt.Println(resp.Message)
	return nil
}

// handleUse handles the use command
func (c *MUDClient) handleUse(ctx context.Context, itemName, target string) error {
	resp, err := c.grpcClient.Use(ctx, c.characterID, itemName, target)
	if err != nil {
		return err
	}

	fmt.Println(resp.Message)
	for _, effect := range resp.Effects {
		fmt.Printf("  %s\n", effect)
	}
	return nil
}

// handleTalk handles the talk command
func (c *MUDClient) handleTalk(ctx context.Context, npcName, topic string) error {
	resp, err := c.grpcClient.Talk(ctx, c.characterID, npcName, topic)
	if err != nil {
		return err
	}

	fmt.Printf("%s: %s\n", npcName, resp.Response)
	if len(resp.Topics) > 0 {
		fmt.Printf("Available topics: %s\n", strings.Join(resp.Topics, ", "))
	}
	return nil
}
