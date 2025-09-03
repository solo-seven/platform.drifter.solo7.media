package logger

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/solo-seven/platform.drifter.solo7.media/internal/domain"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

// OTelLogger implements the domain.Logger interface using OpenTelemetry
type OTelLogger struct {
	logger   log.Logger
	provider *sdklog.LoggerProvider
	level    LogLevel
	output   OutputType
	verbose  bool
	ctx      context.Context
}

// LogLevel represents the logging level
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// OutputType represents the output destination
type OutputType int

const (
	ConsoleOutput OutputType = iota
	JSONOutput
	FileOutput
)

// Config holds the logger configuration
type Config struct {
	Level       LogLevel
	Output      OutputType
	Verbose     bool
	ServiceName string
	FilePath    string // Only used when Output is FileOutput
}

// NewOTelLogger creates a new OpenTelemetry logger
func NewOTelLogger(config Config) (*OTelLogger, error) {
	ctx := context.Background()

	// Create resource with service name
	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", config.ServiceName),
			attribute.String("service.version", "1.0.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create exporter based on output type
	var exporter sdklog.Exporter
	switch config.Output {
	case ConsoleOutput:
		exporter, err = stdoutlog.New(stdoutlog.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("failed to create console exporter: %w", err)
		}
	case JSONOutput:
		exporter, err = stdoutlog.New()
		if err != nil {
			return nil, fmt.Errorf("failed to create JSON exporter: %w", err)
		}
	case FileOutput:
		if config.FilePath == "" {
			config.FilePath = "logs/app.log"
		}
		file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		exporter, err = stdoutlog.New(stdoutlog.WithWriter(file))
		if err != nil {
			return nil, fmt.Errorf("failed to create file exporter: %w", err)
		}
	default:
		exporter, err = stdoutlog.New(stdoutlog.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("failed to create default exporter: %w", err)
		}
	}

	// Create logger provider
	provider := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
	)

	// Get logger instance
	logger := provider.Logger("game-server")

	return &OTelLogger{
		logger:   logger,
		provider: provider,
		level:    config.Level,
		output:   config.Output,
		verbose:  config.Verbose,
		ctx:      ctx,
	}, nil
}

// NewDefaultOTelLogger creates a logger with default console output
func NewDefaultOTelLogger(serviceName string) (*OTelLogger, error) {
	return NewOTelLogger(Config{
		Level:       InfoLevel,
		Output:      ConsoleOutput,
		Verbose:     false,
		ServiceName: serviceName,
	})
}

// Debug logs a debug message
func (l *OTelLogger) Debug(msg string, fields map[string]interface{}) {
	if l.level > DebugLevel {
		return
	}
	l.log(DebugLevel, msg, fields)
}

// Info logs an info message
func (l *OTelLogger) Info(msg string, fields map[string]interface{}) {
	if l.level > InfoLevel {
		return
	}
	l.log(InfoLevel, msg, fields)
}

// Warn logs a warning message
func (l *OTelLogger) Warn(msg string, fields map[string]interface{}) {
	if l.level > WarnLevel {
		return
	}
	l.log(WarnLevel, msg, fields)
}

// Error logs an error message
func (l *OTelLogger) Error(msg string, fields map[string]interface{}) {
	if l.level > ErrorLevel {
		return
	}
	l.log(ErrorLevel, msg, fields)
}

// Fatal logs a fatal message and exits
func (l *OTelLogger) Fatal(msg string, fields map[string]interface{}) {
	l.log(FatalLevel, msg, fields)
	os.Exit(1)
}

// log is the internal logging method
func (l *OTelLogger) log(level LogLevel, msg string, fields map[string]interface{}) {
	// Convert fields to OpenTelemetry attributes
	attrs := make([]log.KeyValue, 0, len(fields)+2)

	// Add timestamp
	attrs = append(attrs, log.String("timestamp", time.Now().Format(time.RFC3339)))

	// Add level
	attrs = append(attrs, log.String("level", l.levelToString(level)))

	// Add custom fields
	for key, value := range fields {
		attrs = append(attrs, l.valueToAttribute(key, value))
	}

	// Create log record
	record := log.Record{}
	record.SetTimestamp(time.Now())
	record.SetBody(log.StringValue(msg))
	record.SetSeverity(l.levelToSeverity(level))
	record.SetSeverityText(l.levelToString(level))

	// Add attributes
	for _, attr := range attrs {
		record.AddAttributes(attr)
	}

	// Emit the log
	l.logger.Emit(l.ctx, record)

	// For console output, also print to stdout/stderr for immediate visibility
	if l.output == ConsoleOutput {
		l.printToConsole(level, msg, fields)
	}
}

// printToConsole prints logs to console for immediate visibility
func (l *OTelLogger) printToConsole(level LogLevel, msg string, fields map[string]interface{}) {
	levelStr := strings.ToUpper(l.levelToString(level))
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	var output string
	if len(fields) > 0 {
		output = fmt.Sprintf("[%s] %s %s: %s %+v\n", timestamp, levelStr, "GAME-SERVER", msg, fields)
	} else {
		output = fmt.Sprintf("[%s] %s %s: %s\n", timestamp, levelStr, "GAME-SERVER", msg)
	}

	// Write to appropriate output stream
	switch level {
	case DebugLevel, InfoLevel:
		if l.verbose || level == InfoLevel {
			fmt.Print(output)
		}
	case WarnLevel:
		fmt.Print(output)
	case ErrorLevel, FatalLevel:
		fmt.Fprint(os.Stderr, output)
	}
}

// levelToString converts LogLevel to string
func (l *OTelLogger) levelToString(level LogLevel) string {
	switch level {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// levelToSeverity converts LogLevel to OpenTelemetry severity
func (l *OTelLogger) levelToSeverity(level LogLevel) log.Severity {
	switch level {
	case DebugLevel:
		return log.SeverityDebug
	case InfoLevel:
		return log.SeverityInfo
	case WarnLevel:
		return log.SeverityWarn
	case ErrorLevel:
		return log.SeverityError
	case FatalLevel:
		return log.SeverityFatal
	default:
		return log.SeverityInfo
	}
}

// valueToAttribute converts a value to an OpenTelemetry attribute
func (l *OTelLogger) valueToAttribute(key string, value interface{}) log.KeyValue {
	switch v := value.(type) {
	case string:
		return log.String(key, v)
	case int:
		return log.Int(key, v)
	case int64:
		return log.Int64(key, v)
	case float64:
		return log.Float64(key, v)
	case bool:
		return log.Bool(key, v)
	case []string:
		values := make([]log.Value, len(v))
		for i, val := range v {
			values[i] = log.StringValue(val)
		}
		return log.Slice(key, values...)
	case []int:
		values := make([]log.Value, len(v))
		for i, val := range v {
			values[i] = log.IntValue(val)
		}
		return log.Slice(key, values...)
	default:
		// Convert to string for unknown types
		return log.String(key, fmt.Sprintf("%v", v))
	}
}

// SetLevel changes the logging level
func (l *OTelLogger) SetLevel(level LogLevel) {
	l.level = level
}

// SetVerbose sets the verbose flag
func (l *OTelLogger) SetVerbose(verbose bool) {
	l.verbose = verbose
}

// Close shuts down the logger
func (l *OTelLogger) Close() error {
	if l.provider != nil {
		return l.provider.Shutdown(l.ctx)
	}
	return nil
}

// Ensure OTelLogger implements domain.Logger interface
var _ domain.Logger = (*OTelLogger)(nil)
