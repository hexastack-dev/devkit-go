package log

import (
	"context"
	"os"
)

type LogLevel int8

const (
	FatalLogLevel LogLevel = 3
	ErrorLogLevel LogLevel = 2
	WarnLogLevel  LogLevel = 1
	InfoLogLevel  LogLevel = 0
	DebugLogLevel LogLevel = -1
)

// Logger log message to underlying logger. Any logging driver should
// implements Logger interface. The optfield should be optional, and
// when it's not nil/empty it should be logged as structured key/value.
type Logger interface {
	// Fatal logs a message at FatalLevel, then calls os.Exit(1). This should only be use with extra care, ideally Fatal
	// should only be used in main where the apps encountered an error and have noway to continue.
	// optfields is optional, when supplied it will be added as new field using
	// Key as field name, and Value as it's value.
	Fatal(msg string, err error, optfields ...LogField)
	// Error logs a message at ErrorLevel, put the passed error in "error" field.
	// optfields is optional, when supplied it will be added as new field using
	// Key as field name, and Value as it's value.
	Error(msg string, err error, optfields ...LogField)
	// Warn logs a message at WarnLevel.
	// optfields is optional, when supplied it will be added as new field using
	// Key as field name, and Value as it's value.
	Warn(msg string, optfields ...LogField)
	// Info logs a message at InfoLevel.
	// optfields is optional, when supplied it will be added as new field using
	// Key as field name, and Value as it's value.
	Info(msg string, optfields ...LogField)
	// Debug logs a message at DebugLevel.
	// optfields is optional, when supplied it will be added as new field using
	// Key as field name, and Value as it's value.
	Debug(msg string, optfields ...LogField)
	// WithContext return Logger instance that will use passed context to log additional info,
	// such as opentelemetry's SpanID and TraceID if applicable.
	WithContext(ctx context.Context) Logger
}

// NoOpLogger will not writes out logs to any output. All NoOpLogger method basically doesn't do anything
// except for Fatal, it will simply call os.Exit(1).
type NoOpLogger struct{}

var _ Logger = &NoOpLogger{}

// Fatal call os.Exit(1)
func (l *NoOpLogger) Fatal(msg string, err error, optfields ...LogField) {
	os.Exit(1)
}
func (l *NoOpLogger) Error(msg string, err error, optfields ...LogField) {}
func (l *NoOpLogger) Warn(msg string, optfields ...LogField)             {}
func (l *NoOpLogger) Info(msg string, optfields ...LogField)             {}
func (l *NoOpLogger) Debug(msg string, optfields ...LogField)            {}
func (l *NoOpLogger) WithContext(ctx context.Context) Logger {
	return l
}

type Writer func(msg string, optfields ...LogField)

func (w Writer) Write(m []byte) (n int, err error) {
	w(string(m))
	return len(m), nil
}
