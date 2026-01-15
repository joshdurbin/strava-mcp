package logging

import (
	"encoding/json"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Level represents the logging verbosity level
type Level int

const (
	// LevelNormal shows INFO and above (default)
	LevelNormal Level = 0
	// LevelVerbose shows DEBUG and above (-v)
	LevelVerbose Level = 1
	// LevelTrace shows DEBUG and above plus HTTP headers (-vv)
	LevelTrace Level = 2
)

var currentLevel Level

// Logger is the global zerolog logger instance
var Logger zerolog.Logger

// Setup initializes zerolog with a console writer to stderr.
// The level parameter controls verbosity:
//   - 0: INFO and above (default)
//   - 1: DEBUG and above (-v)
//   - 2+: DEBUG and above with HTTP headers (-vv)
func Setup(level Level) {
	currentLevel = level

	var zerologLevel zerolog.Level
	switch {
	case level >= LevelVerbose:
		zerologLevel = zerolog.DebugLevel
	default:
		zerologLevel = zerolog.InfoLevel
	}

	output := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}

	Logger = zerolog.New(output).
		Level(zerologLevel).
		With().
		Timestamp().
		Logger()
}

// GetLevel returns the current logging level
func GetLevel() Level {
	return currentLevel
}

// IsVerbose returns true if verbose/debug logging is enabled
func IsVerbose() bool {
	return currentLevel >= LevelVerbose
}

// IsTraceEnabled returns true if trace-level logging (HTTP headers) is enabled
func IsTraceEnabled() bool {
	return currentLevel >= LevelTrace
}

// ToJSON converts any value to JSON string for debug logging
func ToJSON(v any) string {
	if v == nil {
		return "null"
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "<marshal error>"
	}
	// Truncate very long responses
	s := string(b)
	if len(s) > 2000 {
		return s[:2000] + "...(truncated)"
	}
	return s
}

// LeveledLogger implements retryablehttp.LeveledLogger using zerolog
type LeveledLogger struct{}

func (l *LeveledLogger) Error(msg string, keysAndValues ...interface{}) {
	Logger.Error().Fields(keysAndValues).Msg(msg)
}

func (l *LeveledLogger) Info(msg string, keysAndValues ...interface{}) {
	Logger.Info().Fields(keysAndValues).Msg(msg)
}

func (l *LeveledLogger) Debug(msg string, keysAndValues ...interface{}) {
	Logger.Debug().Fields(keysAndValues).Msg(msg)
}

func (l *LeveledLogger) Warn(msg string, keysAndValues ...interface{}) {
	Logger.Warn().Fields(keysAndValues).Msg(msg)
}

// Info logs at info level with key-value pairs (slog-compatible API)
func Info(msg string, keysAndValues ...interface{}) {
	Logger.Info().Fields(keysAndValues).Msg(msg)
}

// Debug logs at debug level with key-value pairs (slog-compatible API)
func Debug(msg string, keysAndValues ...interface{}) {
	Logger.Debug().Fields(keysAndValues).Msg(msg)
}

// Warn logs at warn level with key-value pairs (slog-compatible API)
func Warn(msg string, keysAndValues ...interface{}) {
	Logger.Warn().Fields(keysAndValues).Msg(msg)
}

// Error logs at error level with key-value pairs (slog-compatible API)
func Error(msg string, keysAndValues ...interface{}) {
	Logger.Error().Fields(keysAndValues).Msg(msg)
}
