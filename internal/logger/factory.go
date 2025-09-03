package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
)

// LoggerFactory creates logger instances
type LoggerFactory struct{}

// NewLoggerFactory creates a new logger factory
func NewLoggerFactory() *LoggerFactory {
	return &LoggerFactory{}
}

// CreateLogger creates a logger based on configuration
func (f *LoggerFactory) CreateLogger(serviceName string, options ...LoggerOption) (domain.Logger, error) {
	config := Config{
		Level:       InfoLevel,
		Output:      ConsoleOutput,
		Verbose:     false,
		ServiceName: serviceName,
	}

	// Apply options
	for _, option := range options {
		option(&config)
	}

	return NewOTelLogger(config)
}

// CreateDefaultLogger creates a logger with default settings
func (f *LoggerFactory) CreateDefaultLogger(serviceName string) (domain.Logger, error) {
	return f.CreateLogger(serviceName)
}

// CreateVerboseLogger creates a verbose logger for development
func (f *LoggerFactory) CreateVerboseLogger(serviceName string) (domain.Logger, error) {
	return f.CreateLogger(serviceName, WithVerbose(true), WithLevel(DebugLevel))
}

// CreateProductionLogger creates a logger suitable for production
func (f *LoggerFactory) CreateProductionLogger(serviceName string) (domain.Logger, error) {
	return f.CreateLogger(serviceName,
		WithLevel(InfoLevel),
		WithOutput(JSONOutput),
		WithFilePath("/var/log/game-server/app.log"))
}

// LoggerOption is a function that configures a logger
type LoggerOption func(*Config)

// WithLevel sets the logging level
func WithLevel(level LogLevel) LoggerOption {
	return func(c *Config) {
		c.Level = level
	}
}

// WithOutput sets the output type
func WithOutput(output OutputType) LoggerOption {
	return func(c *Config) {
		c.Output = output
	}
}

// WithVerbose sets the verbose flag
func WithVerbose(verbose bool) LoggerOption {
	return func(c *Config) {
		c.Verbose = verbose
	}
}

// WithFilePath sets the file path for file output
func WithFilePath(filePath string) LoggerOption {
	return func(c *Config) {
		c.FilePath = filePath
	}
}

// WithServiceName sets the service name
func WithServiceName(serviceName string) LoggerOption {
	return func(c *Config) {
		c.ServiceName = serviceName
	}
}

// ParseLogLevel parses a string log level
func ParseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DebugLevel
	case "INFO":
		return InfoLevel
	case "WARN", "WARNING":
		return WarnLevel
	case "ERROR":
		return ErrorLevel
	case "FATAL":
		return FatalLevel
	default:
		return InfoLevel
	}
}

// ParseOutputType parses a string output type
func ParseOutputType(output string) OutputType {
	switch strings.ToUpper(output) {
	case "CONSOLE":
		return ConsoleOutput
	case "JSON":
		return JSONOutput
	case "FILE":
		return FileOutput
	default:
		return ConsoleOutput
	}
}

// CreateLoggerFromEnv creates a logger based on environment variables
func (f *LoggerFactory) CreateLoggerFromEnv(serviceName string) (domain.Logger, error) {
	config := Config{
		Level:       ParseLogLevel(os.Getenv("LOG_LEVEL")),
		Output:      ParseOutputType(os.Getenv("LOG_OUTPUT")),
		Verbose:     os.Getenv("LOG_VERBOSE") == "true",
		ServiceName: serviceName,
		FilePath:    os.Getenv("LOG_FILE_PATH"),
	}

	// Set default file path if using file output but no path specified
	if config.Output == FileOutput && config.FilePath == "" {
		config.FilePath = fmt.Sprintf("logs/%s.log", serviceName)
	}

	return NewOTelLogger(config)
}

// CreateLoggerFromConfig creates a logger based on application configuration
func (f *LoggerFactory) CreateLoggerFromConfig(config domain.Configuration, verboseOverride bool) (domain.Logger, error) {
	logConfig := Config{
		Level:       ParseLogLevel(config.GetLogLevel()),
		Output:      ConsoleOutput, // Default to console
		Verbose:     verboseOverride || config.GetLogVerbose(),
		ServiceName: config.GetLogServiceName(),
		FilePath:    config.GetLogFilePath(),
	}

	// Parse output type from config
	outputType := config.GetLogOutput()
	switch strings.ToUpper(outputType) {
	case "JSON":
		logConfig.Output = JSONOutput
	case "FILE":
		logConfig.Output = FileOutput
	default:
		logConfig.Output = ConsoleOutput
	}

	// Set default file path if using file output but no path specified
	if logConfig.Output == FileOutput && logConfig.FilePath == "" {
		logConfig.FilePath = fmt.Sprintf("logs/%s.log", logConfig.ServiceName)
	}

	return NewOTelLogger(logConfig)
}
