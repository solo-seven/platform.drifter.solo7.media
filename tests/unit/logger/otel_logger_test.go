package logger

import (
	"testing"
	"time"

	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOTelLogger_Creation(t *testing.T) {
	// Test default logger creation
	logger, err := NewDefaultOTelLogger("test-service")
	require.NoError(t, err)
	require.NotNil(t, logger)
	defer logger.Close()

	// Test that it implements the domain.Logger interface
	var _ domain.Logger = logger
}

func TestOTelLogger_Logging(t *testing.T) {
	logger, err := NewDefaultOTelLogger("test-service")
	require.NoError(t, err)
	defer logger.Close()

	// Test all log levels
	logger.Debug("Debug message", map[string]interface{}{"key": "value"})
	logger.Info("Info message", map[string]interface{}{"key": "value"})
	logger.Warn("Warning message", map[string]interface{}{"key": "value"})
	logger.Error("Error message", map[string]interface{}{"key": "value"})

	// Note: We don't test Fatal as it would exit the process
}

func TestOTelLogger_WithDifferentConfigs(t *testing.T) {
	// Test with verbose enabled
	logger, err := NewOTelLogger(Config{
		Level:       DebugLevel,
		Output:      ConsoleOutput,
		Verbose:     true,
		ServiceName: "test-service",
	})
	require.NoError(t, err)
	defer logger.Close()

	logger.Debug("Debug message with verbose", map[string]interface{}{"test": true})
	logger.Info("Info message with verbose", map[string]interface{}{"test": true})
}

func TestOTelLogger_LevelFiltering(t *testing.T) {
	// Test with Info level - Debug should be filtered out
	logger, err := NewOTelLogger(Config{
		Level:       InfoLevel,
		Output:      ConsoleOutput,
		Verbose:     false,
		ServiceName: "test-service",
	})
	require.NoError(t, err)
	defer logger.Close()

	// These should all work
	logger.Info("Info message", nil)
	logger.Warn("Warning message", nil)
	logger.Error("Error message", nil)
}

func TestLoggerFactory(t *testing.T) {
	factory := NewLoggerFactory()

	// Test default logger
	logger, err := factory.CreateDefaultLogger("test-service")
	require.NoError(t, err)
	require.NotNil(t, logger)
	defer logger.(*OTelLogger).Close()

	// Test verbose logger
	verboseLogger, err := factory.CreateVerboseLogger("test-service")
	require.NoError(t, err)
	require.NotNil(t, verboseLogger)
	defer verboseLogger.(*OTelLogger).Close()

	// Test production logger
	prodLogger, err := factory.CreateProductionLogger("test-service")
	require.NoError(t, err)
	require.NotNil(t, prodLogger)
	defer prodLogger.(*OTelLogger).Close()
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"DEBUG", DebugLevel},
		{"INFO", InfoLevel},
		{"WARN", WarnLevel},
		{"WARNING", WarnLevel},
		{"ERROR", ErrorLevel},
		{"FATAL", FatalLevel},
		{"unknown", InfoLevel}, // default
		{"", InfoLevel},        // default
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := ParseLogLevel(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestParseOutputType(t *testing.T) {
	tests := []struct {
		input    string
		expected OutputType
	}{
		{"CONSOLE", ConsoleOutput},
		{"JSON", JSONOutput},
		{"FILE", FileOutput},
		{"unknown", ConsoleOutput}, // default
		{"", ConsoleOutput},        // default
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := ParseOutputType(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestOTelLogger_WithFields(t *testing.T) {
	logger, err := NewDefaultOTelLogger("test-service")
	require.NoError(t, err)
	defer logger.Close()

	// Test with various field types
	fields := map[string]interface{}{
		"string":  "test",
		"int":     42,
		"int64":   int64(123),
		"float64": 3.14,
		"bool":    true,
		"strings": []string{"a", "b", "c"},
		"ints":    []int{1, 2, 3},
	}

	logger.Info("Message with various field types", fields)
}

// Mock configuration for testing
type mockConfig struct {
	logLevel       string
	logVerbose     bool
	logOutput      string
	logFilePath    string
	logServiceName string
}

func (m *mockConfig) GetLogLevel() string       { return m.logLevel }
func (m *mockConfig) GetLogVerbose() bool       { return m.logVerbose }
func (m *mockConfig) GetLogOutput() string      { return m.logOutput }
func (m *mockConfig) GetLogFilePath() string    { return m.logFilePath }
func (m *mockConfig) GetLogServiceName() string { return m.logServiceName }

// Implement other required methods (not used in this test)
func (m *mockConfig) GetServerPort() int                  { return 8080 }
func (m *mockConfig) GetMaxConnections() int              { return 1000 }
func (m *mockConfig) GetHeartbeatInterval() time.Duration { return 30 * time.Second }
func (m *mockConfig) GetRegionSize() float64              { return 1000.0 }
func (m *mockConfig) GetMaxEntitiesPerRegion() int        { return 1000 }
func (m *mockConfig) GetDatabaseURL() string              { return "" }
func (m *mockConfig) GetRedisURL() string                 { return "" }

func TestLoggerFactory_CreateLoggerFromConfig(t *testing.T) {
	factory := NewLoggerFactory()

	tests := []struct {
		name     string
		config   *mockConfig
		verbose  bool
		expected LogLevel
	}{
		{
			name: "Default config",
			config: &mockConfig{
				logLevel:       "info",
				logVerbose:     false,
				logOutput:      "console",
				logServiceName: "test-service",
			},
			verbose:  false,
			expected: InfoLevel,
		},
		{
			name: "Debug level with verbose",
			config: &mockConfig{
				logLevel:       "debug",
				logVerbose:     true,
				logOutput:      "console",
				logServiceName: "test-service",
			},
			verbose:  false,
			expected: DebugLevel,
		},
		{
			name: "Verbose override",
			config: &mockConfig{
				logLevel:       "info",
				logVerbose:     false,
				logOutput:      "console",
				logServiceName: "test-service",
			},
			verbose:  true,
			expected: InfoLevel,
		},
		{
			name: "JSON output",
			config: &mockConfig{
				logLevel:       "warn",
				logVerbose:     false,
				logOutput:      "json",
				logServiceName: "test-service",
			},
			verbose:  false,
			expected: WarnLevel,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger, err := factory.CreateLoggerFromConfig(test.config, test.verbose)
			require.NoError(t, err)
			require.NotNil(t, logger)
			defer logger.(*OTelLogger).Close()

			// Test that the logger works
			logger.Info("Test message", map[string]interface{}{"test": true})
		})
	}
}
