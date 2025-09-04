package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	"github.com/solo-seven/platform.drifter.solo7.media/internal/logger"
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

	rootCmd.Flags().StringVarP(&serverURL, "server", "s", "localhost:8081", "gRPC server address")
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

	// Create logger
	loggerFactory := logger.NewLoggerFactory()
	gameLogger, err := loggerFactory.CreateLogger("game-client",
		logger.WithVerbose(verbose),
		logger.WithLevel(logger.InfoLevel))
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer gameLogger.(*logger.OTelLogger).Close()

	// Create MUD client
	grpcAddr := serverURL
	wsAddr := "ws://localhost:8080/ws"
	client, err := NewMUDClient(grpcAddr, wsAddr, gameLogger)
	if err != nil {
		log.Fatalf("Failed to create MUD client: %v", err)
	}
	defer client.Close()

	fmt.Println("Connected to server successfully!")

	// Login
	characterName := "TestCharacter"
	if err := client.Login(context.Background(), playerID, characterName); err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	// Set up context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the client
	go func() {
		if err := client.Start(ctx); err != nil {
			fmt.Printf("Client error: %v\n", err)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nShutting down client...")
	cancel()
}
