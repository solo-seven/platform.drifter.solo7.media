package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
	"github.com/solo-seven/platform.drifter.solo7.media/internal/logger"
	"github.com/solo-seven/platform.drifter.solo7.media/internal/network"
	"github.com/spf13/cobra"
)

var (
	serverURL string
	playerID  string
	verbose   bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "drifter-client",
		Short: "CLI client for Drifter Platform RPG",
		Long:  "A command-line client for connecting to and interacting with the Drifter Platform RPG server.",
		Run:   runClient,
	}

	rootCmd.Flags().StringVarP(&serverURL, "server", "s", "ws://localhost:8080/ws", "WebSocket server URL")
	rootCmd.Flags().StringVarP(&playerID, "player", "p", "", "Player ID (generated if not provided)")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runClient(cmd *cobra.Command, args []string) {
	// Generate player ID if not provided
	if playerID == "" {
		playerID = uuid.New().String()
	}

	fmt.Printf("Connecting to server: %s\n", serverURL)
	fmt.Printf("Player ID: %s\n", playerID)

	// Parse server URL
	u, err := url.Parse(serverURL)
	if err != nil {
		log.Fatalf("Invalid server URL: %v", err)
	}

	// Connect to WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	fmt.Println("Connected to server successfully!")

	// Create client connection
	loggerFactory := logger.NewLoggerFactory()
	gameLogger, err := loggerFactory.CreateLogger("game-client",
		logger.WithVerbose(verbose),
		logger.WithLevel(logger.InfoLevel))
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer gameLogger.(*logger.OTelLogger).Close()
	clientConn := network.NewWebSocketConnection(conn, uuid.MustParse(playerID), gameLogger)

	// Set up context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start message receiver goroutine
	go receiveMessages(ctx, clientConn)

	// Start heartbeat goroutine
	go sendHeartbeat(ctx, clientConn)

	// Start interactive command loop
	go interactiveLoop(ctx, clientConn)

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nShutting down client...")
	cancel()
}

func receiveMessages(ctx context.Context, conn domain.ClientConnection) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			message, err := conn.Receive(ctx)
			if err != nil {
				if ctx.Err() == nil {
					fmt.Printf("Error receiving message: %v\n", err)
				}
				return
			}

			handleReceivedMessage(message)
		}
	}
}

func handleReceivedMessage(message *domain.NetworkMessage) {
	switch message.Type {
	case domain.StateUpdate:
		fmt.Printf("📊 State Update: %+v\n", message.Data)
	case domain.AestheticEvent:
		fmt.Printf("🎭 Aesthetic Event: %+v\n", message.Data)
	case domain.SystemNotification:
		fmt.Printf("🔔 System: %+v\n", message.Data)
	case domain.ChatMessage:
		playerID := message.Data["player_id"]
		msg := message.Data["message"]
		channel := message.Data["channel"]
		fmt.Printf("💬 [%s] %s: %s\n", channel, playerID, msg)
	case domain.Heartbeat:
		// Heartbeat responses are handled silently
	default:
		fmt.Printf("📨 Unknown message type: %v\n", message.Data)
	}
}

func sendHeartbeat(ctx context.Context, conn domain.ClientConnection) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			heartbeat := &domain.NetworkMessage{
				Type:      domain.Heartbeat,
				ID:        uuid.New().String(),
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"connection_id": conn.GetConnectionID(),
				},
			}

			if err := conn.Send(ctx, heartbeat); err != nil {
				if ctx.Err() == nil {
					fmt.Printf("Error sending heartbeat: %v\n", err)
				}
				return
			}
		}
	}
}

func interactiveLoop(ctx context.Context, conn domain.ClientConnection) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("\nAvailable commands:")
	fmt.Println("  /help     - Show this help message")
	fmt.Println("  /move <direction> - Move character (forward, back, left, right)")
	fmt.Println("  /attack <target> - Attack a target")
	fmt.Println("  /chat <message> - Send chat message")
	fmt.Println("  /quit     - Disconnect from server")
	fmt.Println("\nType a command or message:")

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

			if err := handleCommand(ctx, conn, input); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		}
	}
}

func handleCommand(ctx context.Context, conn domain.ClientConnection, input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	command := parts[0]

	switch command {
	case "/help":
		showHelp()
	case "/move":
		return handleMoveCommand(ctx, conn, parts[1:])
	case "/attack":
		return handleAttackCommand(ctx, conn, parts[1:])
	case "/chat":
		return handleChatCommand(ctx, conn, parts[1:])
	case "/quit":
		fmt.Println("Disconnecting...")
		os.Exit(0)
	default:
		// Treat as chat message
		return handleChatCommand(ctx, conn, parts)
	}

	return nil
}

func showHelp() {
	fmt.Println("\nAvailable commands:")
	fmt.Println("  /help     - Show this help message")
	fmt.Println("  /move <direction> - Move character (forward, back, left, right)")
	fmt.Println("  /attack <target> - Attack a target")
	fmt.Println("  /chat <message> - Send chat message")
	fmt.Println("  /quit     - Disconnect from server")
	fmt.Println("\nYou can also just type a message to send it as a chat message.")
}

func handleMoveCommand(ctx context.Context, conn domain.ClientConnection, args []string) error {
	if len(args) == 0 {
		fmt.Println("Usage: /move <direction>")
		fmt.Println("Directions: forward, back, left, right")
		return nil
	}

	direction := args[0]
	validDirections := map[string]bool{
		"forward": true,
		"back":    true,
		"left":    true,
		"right":   true,
	}

	if !validDirections[direction] {
		fmt.Println("Invalid direction. Use: forward, back, left, right")
		return nil
	}

	message := &domain.NetworkMessage{
		Type:      domain.PlayerInput,
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"player_id":  conn.GetPlayerID().String(),
			"input_type": "movement",
			"data": map[string]interface{}{
				"direction": direction,
				"speed":     "1.0",
			},
		},
	}

	return conn.Send(ctx, message)
}

func handleAttackCommand(ctx context.Context, conn domain.ClientConnection, args []string) error {
	if len(args) == 0 {
		fmt.Println("Usage: /attack <target>")
		return nil
	}

	target := strings.Join(args, " ")

	message := &domain.NetworkMessage{
		Type:      domain.PlayerInput,
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"player_id":  conn.GetPlayerID().String(),
			"input_type": "attack",
			"data": map[string]interface{}{
				"target": target,
				"weapon": "sword",
			},
		},
	}

	return conn.Send(ctx, message)
}

func handleChatCommand(ctx context.Context, conn domain.ClientConnection, args []string) error {
	if len(args) == 0 {
		fmt.Println("Usage: /chat <message>")
		return nil
	}

	message := strings.Join(args, " ")

	chatMessage := &domain.NetworkMessage{
		Type:      domain.ChatMessage,
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"player_id": conn.GetPlayerID().String(),
			"message":   message,
			"channel":   "general",
		},
	}

	return conn.Send(ctx, chatMessage)
}
