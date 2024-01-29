package log

import (
	"io"
	"time"

	"cosmossdk.io/log"
)

// Logger is the interface for the logger. It's based on cosmossdk.io/log.
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

// loggerImpl is the implementation of the Logger interface.
type loggerImpl struct {
	log.Logger
}

// With returns a new wrapped logger with additional context provided by a set.
func (l *loggerImpl) With(keyVals ...any) Logger {
	logger := l.Logger.With(keyVals...)
	return &loggerImpl{logger}
}

// NewLogger creates a new logger with the given writer and runner name.
// The logger is a wrapper around cosmossdk logger.
func NewLogger(dst io.Writer, runner string) Logger {
	opts := []log.Option{
		log.ColorOption(true),
		log.TimeFormatOption(time.RFC3339),
	}
	logger := log.NewLogger(dst, opts...)
	return &loggerImpl{logger.With("namespace", runner)}
}

// NewJsonLogger creates a new logger with the given writer and runner name.
// It sets the output of the logger to JSON.
func NewJSONLogger(dst io.Writer, runner string) Logger {
	opts := []log.Option{
		log.OutputJSONOption(),
		log.TimeFormatOption(time.RFC3339),
	}

	logger := log.NewLogger(dst, opts...)
	return &loggerImpl{logger.With("namespace", runner)}
}

// NewBlankLogger creates a new logger with the given writer.
// The logger is a wrapper around cosmossdk logger.
func NewBlankLogger(dst io.Writer) Logger {
	opts := []log.Option{
		log.ColorOption(true),
		log.TimeFormatOption(time.Kitchen),
	}

	logger := log.NewLogger(dst, opts...)
	return &loggerImpl{logger}
}
