package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/solo7.media/platform.drifter.solo7.media/internal/domain"
	"github.com/solo7.media/platform.drifter.solo7.media/internal/network"
	"github.com/solo7.media/platform.drifter.solo7.media/internal/server"
	"github.com/spf13/viper"
)

var (
	configFile = flag.String("config", "config.yaml", "Configuration file path")
	port       = flag.Int("port", 8080, "Server port")
	verbose    = flag.Bool("verbose", false, "Enable verbose logging")
)

func main() {
	flag.Parse()

	// Load configuration
	config := loadConfig()

	// Create logger
	logger := &consoleLogger{verbose: *verbose}

	// Create server components
	cm := network.NewConnectionManager(logger)
	protocol := network.NewWebSocketProtocol(logger)
	protocol.SetConnectionManager(cm)

	// Create game server
	gameServer := server.NewGameServer(config, cm, protocol, logger)

	// Set up HTTP server with WebSocket endpoint
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", gameServer.HandleWebSocket)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: mux,
	}

	// Start server
	go func() {
		logger.Info("Starting server", map[string]interface{}{
			"port":        *port,
			"config_file": *configFile,
		})

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", map[string]interface{}{
				"error": err,
			})
		}
	}()

	// Start game server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := gameServer.Start(ctx); err != nil {
		logger.Fatal("Failed to start game server", map[string]interface{}{
			"error": err,
		})
	}

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	logger.Info("Shutting down server...", nil)

	// Shutdown game server
	if err := gameServer.Stop(ctx); err != nil {
		logger.Error("Error stopping game server", map[string]interface{}{
			"error": err,
		})
	}

	// Shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("Error shutting down HTTP server", map[string]interface{}{
			"error": err,
		})
	}

	logger.Info("Server shutdown complete", nil)
}

func loadConfig() domain.Configuration {
	viper.SetConfigFile(*configFile)
	viper.SetConfigType("yaml")

	// Set default values
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.max_connections", 1000)
	viper.SetDefault("server.heartbeat_interval", "30s")
	viper.SetDefault("server.region_size", 1000.0)
	viper.SetDefault("server.max_entities_per_region", 1000)
	viper.SetDefault("server.log_level", "info")
	viper.SetDefault("server.database_url", "")
	viper.SetDefault("server.redis_url", "")

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("Error reading config file: %v", err)
		}
	}

	return &configImpl{
		port:                 viper.GetInt("server.port"),
		maxConnections:       viper.GetInt("server.max_connections"),
		heartbeatInterval:    viper.GetDuration("server.heartbeat_interval"),
		regionSize:           viper.GetFloat64("server.region_size"),
		maxEntitiesPerRegion: viper.GetInt("server.max_entities_per_region"),
		logLevel:             viper.GetString("server.log_level"),
		databaseURL:          viper.GetString("server.database_url"),
		redisURL:             viper.GetString("server.redis_url"),
	}
}

// Configuration implementation
type configImpl struct {
	port                 int
	maxConnections       int
	heartbeatInterval    time.Duration
	regionSize           float64
	maxEntitiesPerRegion int
	logLevel             string
	databaseURL          string
	redisURL             string
}

func (c *configImpl) GetServerPort() int {
	return c.port
}

func (c *configImpl) GetMaxConnections() int {
	return c.maxConnections
}

func (c *configImpl) GetHeartbeatInterval() time.Duration {
	return c.heartbeatInterval
}

func (c *configImpl) GetRegionSize() float64 {
	return c.regionSize
}

func (c *configImpl) GetMaxEntitiesPerRegion() int {
	return c.maxEntitiesPerRegion
}

func (c *configImpl) GetLogLevel() string {
	return c.logLevel
}

func (c *configImpl) GetDatabaseURL() string {
	return c.databaseURL
}

func (c *configImpl) GetRedisURL() string {
	return c.redisURL
}

// Console logger implementation
type consoleLogger struct {
	verbose bool
}

func (l *consoleLogger) Debug(msg string, fields map[string]interface{}) {
	if l.verbose {
		fmt.Printf("DEBUG: %s %+v\n", msg, fields)
	}
}

func (l *consoleLogger) Info(msg string, fields map[string]interface{}) {
	fmt.Printf("INFO: %s %+v\n", msg, fields)
}

func (l *consoleLogger) Warn(msg string, fields map[string]interface{}) {
	fmt.Printf("WARN: %s %+v\n", msg, fields)
}

func (l *consoleLogger) Error(msg string, fields map[string]interface{}) {
	fmt.Printf("ERROR: %s %+v\n", msg, fields)
}

func (l *consoleLogger) Fatal(msg string, fields map[string]interface{}) {
	fmt.Printf("FATAL: %s %+v\n", msg, fields)
	os.Exit(1)
}
