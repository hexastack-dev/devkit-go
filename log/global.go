package log

import (
	"context"
	"log"
)

var globalLogger Logger

func init() {
	globalLogger = NewSimpleLogger(log.Default().Writer())
}

// GetLogger get global logger, by default global logger use SimpleLogger and use
// standard log default writer (log.Default().Writer()) as it's writer.
// Use SetLogger to set global logger.
func GetLogger() Logger {
	return globalLogger
}

// SetLogger set global logger.
func SetLogger(logger Logger) {
	globalLogger = logger
}

// Fatal logs a message at FatalLevel using global logger, then calls os.Exit(1). This should only be use with extra care, ideally Fatal
// should only be used in main where the apps encountered an error and have noway to continue.
// optfields is optional, when supplied it will be added as new field using
// Key as field name, and Value as it's value.
func Fatal(msg string, err error, optfields ...LogField) {
	GetLogger().Fatal(msg, err, optfields...)
}

// Error logs a message at ErrorLevel using global logger, put the passed error in "error" field.
// optfields is optional, when supplied it will be added as new field using
// Key as field name, and Value as it's value.
func Error(msg string, err error, optfields ...LogField) {
	GetLogger().Error(msg, err, optfields...)
}

// Warn logs a message at WarnLevel using global logger.
// optfields is optional, when supplied it will be added as new field using
// Key as field name, and Value as it's value.
func Warn(msg string, optfields ...LogField) {
	GetLogger().Warn(msg, optfields...)
}

// Info logs a message at InfoLevel using global logger.
// optfields is optional, when supplied it will be added as new field using
// Key as field name, and Value as it's value.
func Info(msg string, optfields ...LogField) {
	GetLogger().Info(msg, optfields...)
}

// Debug logs a message at DebugLevel using global logger.
// optfields is optional, when supplied it will be added as new field using
// Key as field name, and Value as it's value.
func Debug(msg string, optfields ...LogField) {
	GetLogger().Debug(msg, optfields...)
}

// WithContext return Logger instance that will use passed context to log additional info,
// such as opentelemetry's SpanID and TraceID if applicable.
func WithContext(ctx context.Context) Logger {
	return GetLogger().WithContext(ctx)
}
