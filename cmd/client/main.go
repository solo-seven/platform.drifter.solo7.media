package main

import (
	"fmt"
	"log"

	"github.com/google/uuid"
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
}
