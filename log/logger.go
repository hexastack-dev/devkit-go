package log

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"
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

var _ Logger = &NoOpLogger{}

// NoOpLogger will not writes out logs to any output. All NoOpLogger method basically doesn't do anything
// except for Fatal, it will simply call os.Exit(1).
type NoOpLogger struct{}

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

// WriterFunc takes Logger's log method signature to implement io.Writer,
// this is useful when you want to use Logger as standard log's output.
// ie. stdlog.SetOutput(log.WriterFunc(logger.Debug))
type WriterFunc func(msg string, optfields ...LogField)

func (w WriterFunc) Write(m []byte) (n int, err error) {
	w(string(m))
	return len(m), nil
}

var _ Logger = &SimpleLogger{}

type SimpleLogger struct {
	l  *log.Logger
	lv LogLevel
}

func NewSimpleLogger(w io.Writer, lv LogLevel) *SimpleLogger {
	if w == nil {
		w = log.Default().Writer()
	}
	l := log.New(w, "", 0)
	return &SimpleLogger{
		l:  l,
		lv: lv,
	}
}

// Fatal call os.Exit(1)
func (l *SimpleLogger) Fatal(msg string, err error, optfields ...LogField) {
	l.writeLog(FatalLogLevel, msg, err, optfields...)
	os.Exit(1)
}

func (l *SimpleLogger) Error(msg string, err error, optfields ...LogField) {
	l.writeLog(ErrorLogLevel, msg, err, optfields...)
}

func (l *SimpleLogger) Warn(msg string, optfields ...LogField) {
	l.writeLog(WarnLogLevel, msg, nil, optfields...)
}

func (l *SimpleLogger) Info(msg string, optfields ...LogField) {
	l.writeLog(InfoLogLevel, msg, nil, optfields...)
}

func (l *SimpleLogger) Debug(msg string, optfields ...LogField) {
	l.writeLog(DebugLogLevel, msg, nil, optfields...)
}

func (l *SimpleLogger) WithContext(ctx context.Context) Logger {
	return l
}

func (l *SimpleLogger) writeLog(lv LogLevel, msg string, err error, optfields ...LogField) {
	if l.lv > lv {
		return
	}

	tf := "2006-01-02T15:04:05.000Z0700"
	now := time.Now()
	var b []byte
	b = append(b, "timestamp:"...)
	b = now.AppendFormat(b, tf)
	b = append(b, "\tlevel:"...)

	switch lv {
	case FatalLogLevel:
		b = append(b, "fatal"...)
		b = append(b, "\terror:"...)
		b = append(b, err.Error()...)
	case ErrorLogLevel:
		b = append(b, "error"...)
		b = append(b, "\terror:"...)
		b = append(b, err.Error()...)
	case WarnLogLevel:
		b = append(b, "warn"...)
	case InfoLogLevel:
		b = append(b, "info"...)
	default:
		b = append(b, "debug"...)
	}

	b = append(b, "\tmessage:"...)
	b = append(b, msg...)

	for _, field := range optfields {
		b = append(b, '\t')
		b = append(b, field.Key...)
		b = append(b, ':')
		b = append(b, fmt.Sprint(field.Value)...)
	}
	l.l.Println(string(b))
}
