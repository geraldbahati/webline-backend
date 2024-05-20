package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

// Init initializes the logger
func Init() {
	// Set up custom log configuration
	config := zap.Config{
		Encoding:         "json", // or "console"
		Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
		OutputPaths:      []string{"stdout", "app.log"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}

	var err error
	logger, err = config.Build()
	if err != nil {
		panic(err)
	}
}

// Info logs a message at Info level
func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

// Warn logs a message at Warn level
func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

// Error logs a message at Error level
func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

// Debug logs a message at Debug level
func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

// Fatal logs a message at Fatal level and exits
func Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}

// Sync flushes any buffered log entries
func Sync() {
	_ = logger.Sync()
}
