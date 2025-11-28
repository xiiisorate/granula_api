// Package logger provides structured logging for Granula microservices.
//
// Uses uber-go/zap for high-performance, structured JSON logging.
// Supports different log levels, context propagation, and request tracing.
//
// Example:
//
//	log := logger.New(logger.Config{Level: "info", Format: "json"})
//	log.Info("server started", logger.F("port", 8080))
//	log.WithContext(ctx).Error("request failed", logger.Err(err))
package logger

import (
	"context"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ctxKey is the context key for logger.
type ctxKey struct{}

// F creates a zap.Field for structured logging.
// Shorthand for zap.Any.
func F(key string, value any) zap.Field {
	return zap.Any(key, value)
}

// Err creates a zap.Field for error logging.
func Err(err error) zap.Field {
	return zap.Error(err)
}

// String creates a string field.
func String(key, value string) zap.Field {
	return zap.String(key, value)
}

// Int creates an int field.
func Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}

// Int64 creates an int64 field.
func Int64(key string, value int64) zap.Field {
	return zap.Int64(key, value)
}

// Bool creates a bool field.
func Bool(key string, value bool) zap.Field {
	return zap.Bool(key, value)
}

// Duration creates a duration field.
func Duration(key string, value int64) zap.Field {
	return zap.Int64(key+"_ms", value)
}

// Config holds logger configuration.
type Config struct {
	// Level is the minimum log level (debug, info, warn, error).
	Level string `mapstructure:"level"`

	// Format is the output format (json, console).
	Format string `mapstructure:"format"`

	// ServiceName is added to all log entries.
	ServiceName string `mapstructure:"service_name"`

	// Development enables development mode (more verbose).
	Development bool `mapstructure:"development"`

	// OutputPaths are the log output destinations.
	OutputPaths []string `mapstructure:"output_paths"`
}

// DefaultConfig returns sensible defaults for production.
func DefaultConfig() Config {
	return Config{
		Level:       "info",
		Format:      "json",
		Development: false,
		OutputPaths: []string{"stdout"},
	}
}

// Logger wraps zap.Logger with additional functionality.
type Logger struct {
	zap    *zap.Logger
	config Config
}

var (
	globalLogger *Logger
	once         sync.Once
)

// New creates a new Logger instance.
func New(cfg Config) (*Logger, error) {
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if cfg.Format == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Configure output
	outputPaths := cfg.OutputPaths
	if len(outputPaths) == 0 {
		outputPaths = []string{"stdout"}
	}

	var writers []zapcore.WriteSyncer
	for _, path := range outputPaths {
		switch path {
		case "stdout":
			writers = append(writers, zapcore.AddSync(os.Stdout))
		case "stderr":
			writers = append(writers, zapcore.AddSync(os.Stderr))
		default:
			file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return nil, err
			}
			writers = append(writers, zapcore.AddSync(file))
		}
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.NewMultiWriteSyncer(writers...),
		level,
	)

	opts := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	}

	if cfg.Development {
		opts = append(opts, zap.Development())
	}

	if cfg.ServiceName != "" {
		opts = append(opts, zap.Fields(zap.String("service", cfg.ServiceName)))
	}

	zapLogger := zap.New(core, opts...)

	return &Logger{
		zap:    zapLogger,
		config: cfg,
	}, nil
}

// MustNew creates a new Logger and panics on error.
func MustNew(cfg Config) *Logger {
	l, err := New(cfg)
	if err != nil {
		panic(err)
	}
	return l
}

// Global returns the global logger instance.
// Creates a default logger if not initialized.
func Global() *Logger {
	once.Do(func() {
		var err error
		globalLogger, err = New(DefaultConfig())
		if err != nil {
			panic(err)
		}
	})
	return globalLogger
}

// SetGlobal sets the global logger instance.
func SetGlobal(l *Logger) {
	globalLogger = l
}

// parseLevel converts string level to zapcore.Level.
func parseLevel(level string) (zapcore.Level, error) {
	var l zapcore.Level
	err := l.UnmarshalText([]byte(level))
	return l, err
}

// Debug logs a message at debug level.
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, fields...)
}

// Info logs a message at info level.
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}

// Warn logs a message at warn level.
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}

// Error logs a message at error level.
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}

// Fatal logs a message at fatal level and exits.
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.zap.Fatal(msg, fields...)
}

// With creates a child logger with additional fields.
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{
		zap:    l.zap.With(fields...),
		config: l.config,
	}
}

// WithContext extracts trace information from context.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// Extract request ID, trace ID, user ID from context
	fields := []zap.Field{}

	if requestID := ctx.Value("request_id"); requestID != nil {
		fields = append(fields, zap.String("request_id", requestID.(string)))
	}

	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields = append(fields, zap.String("trace_id", traceID.(string)))
	}

	if userID := ctx.Value("user_id"); userID != nil {
		fields = append(fields, zap.String("user_id", userID.(string)))
	}

	if len(fields) == 0 {
		return l
	}

	return l.With(fields...)
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() error {
	return l.zap.Sync()
}

// Named adds a sub-scope to the logger's name.
func (l *Logger) Named(name string) *Logger {
	return &Logger{
		zap:    l.zap.Named(name),
		config: l.config,
	}
}

// ToContext adds logger to context.
func ToContext(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// FromContext extracts logger from context.
// Returns global logger if not found.
func FromContext(ctx context.Context) *Logger {
	if l, ok := ctx.Value(ctxKey{}).(*Logger); ok {
		return l
	}
	return Global()
}

// Zap returns the underlying zap.Logger.
func (l *Logger) Zap() *zap.Logger {
	return l.zap
}
