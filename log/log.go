package log

import (
	"io"
	"time"

	"github.com/rs/zerolog"
)

// Logger is the interface for the logger. It's a wrapper around zerolog.
type Logger interface {
	// Info takes a message and a set of key/value pairs and logs with level INFO.
	// The key of the tuple must be a string.
	Info(msg string, keyVals ...any)

	// Warn takes a message and a set of key/value pairs and logs with level WARN.
	// The key of the tuple must be a string.
	Warn(msg string, keyVals ...any)

	// Error takes a message and a set of key/value pairs and logs with level ERR.
	// The key of the tuple must be a string.
	Error(msg string, keyVals ...any)

	// Debug takes a message and a set of key/value pairs and logs with level DEBUG.
	// The key of the tuple must be a string.
	Debug(msg string, keyVals ...any)

	// With returns a new wrapped logger with additional context provided by a set
	With(keyVals ...any) Logger

	// Impl returns the underlying logger implementation
	// It is used to access the full functionalities of the underlying logger
	// Advanced users can type cast the returned value to the actual logger
	Impl() any
}

// zeroLogWrapper is the implementation of the Logger interface.
// It wraps a zerolog logger.
type zeroLogWrapper struct {
	*zerolog.Logger
}

// Info takes a message and a set of key/value pairs and logs with level INFO.
// The key of the tuple must be a string.
func (l zeroLogWrapper) Info(msg string, keyVals ...interface{}) {
	l.Logger.Info().Fields(keyVals).Msg(msg)
}

// Error takes a message and a set of key/value pairs and logs with level DEBUG.
// The key of the tuple must be a string.
func (l zeroLogWrapper) Error(msg string, keyVals ...interface{}) {
	l.Logger.Error().Fields(keyVals).Msg(msg)
}

// Warn takes a message and a set of key/value pairs and logs with level WARN.
// The key of the tuple must be a string.
func (l zeroLogWrapper) Warn(msg string, keyVals ...interface{}) {
	l.Logger.Warn().Fields(keyVals).Msg(msg)
}

// Debug takes a message and a set of key/value pairs and logs with level ERR.
// The key of the tuple must be a string.
func (l zeroLogWrapper) Debug(msg string, keyVals ...interface{}) {
	l.Logger.Debug().Fields(keyVals).Msg(msg)
}

// With returns a new wrapped logger with additional context provided by a set.
func (l zeroLogWrapper) With(keyVals ...interface{}) Logger {
	logger := l.Logger.With().Fields(keyVals).Logger()
	return zeroLogWrapper{&logger}
}

// Impl returns the underlying zerolog logger.
// It can be used to used zerolog structured API directly instead of the wrapper.
func (l zeroLogWrapper) Impl() interface{} {
	return l.Logger
}

// NewLogger creates a new logger with the given writer and runner name.
// The logger is a wrapper around zerolog.
func NewLogger(dst io.Writer, runner string) Logger {
	output := zerolog.ConsoleWriter{Out: dst, TimeFormat: time.Kitchen}
	logger := zerolog.New(output).With().Timestamp().Str("namespace", runner).Logger()
	return zeroLogWrapper{&logger}
}

// NewBlankLogger creates a new logger with the given writer and runner name.
// The logger is a wrapper around zerolog.
func NewBlankLogger(dst io.Writer) Logger {
	output := zerolog.ConsoleWriter{Out: dst, TimeFormat: time.Kitchen}
	logger := zerolog.New(output).With().Timestamp().Logger()
	return zeroLogWrapper{&logger}
}
