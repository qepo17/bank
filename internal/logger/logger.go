package logger

import (
	"context"
	"os"

	"github.com/rs/zerolog"
)

type Logger struct {
	logger zerolog.Logger
}

func NewLogger(level string) *Logger {
	logLevel := parseLogLevel(level)

	logger := zerolog.New(os.Stdout).
		Level(logLevel).
		With().
		Timestamp().
		Logger()

	return &Logger{
		logger: logger,
	}
}

func parseLogLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

func (l *Logger) Info(ctx context.Context, message string, args ...interface{}) {
	l.logger.Info().Ctx(ctx).Msgf(message, args...)
}

func (l *Logger) Error(ctx context.Context, message string, args ...interface{}) {
	l.logger.Error().Ctx(ctx).Msgf(message, args...)
}

func (l *Logger) Debug(ctx context.Context, message string, args ...interface{}) {
	l.logger.Debug().Ctx(ctx).Msgf(message, args...)
}

func (l *Logger) Warn(ctx context.Context, message string, args ...interface{}) {
	l.logger.Warn().Ctx(ctx).Msgf(message, args...)
}

func (l *Logger) Fatal(ctx context.Context, message string, args ...interface{}) {
	l.logger.Fatal().Ctx(ctx).Msgf(message, args...)
}

// WithFields returns a new logger instance with additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newLogger := l.logger.With()
	for k, v := range fields {
		newLogger = newLogger.Interface(k, v)
	}

	return &Logger{
		logger: newLogger.Logger(),
	}
}

// WithField returns a new logger instance with an additional field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		logger: l.logger.With().Interface(key, value).Logger(),
	}
}
