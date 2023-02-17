package log_test

import (
	"errors"
	"testing"

	"github.com/hexastack-dev/devkit-go/log"
	"github.com/stretchr/testify/assert"
)

func BenchmarkLogger(b *testing.B) {
	logger := &log.NoOpLogger{}

	b.ResetTimer()
	b.Run("log", func(b *testing.B) {
		testLog(b, logger)
	})

	b.ResetTimer()
	b.Run("log with 10 fields", func(b *testing.B) {
		testLogWithArguments(b, logger)
	})

	b.ResetTimer()
	b.Run("log error with 10 fields", func(b *testing.B) {
		testError(b, logger)
	})
}

func testLog(b *testing.B, logger log.Logger) {
	for i := 0; i < b.N; i++ {
		logger.Info("info message")
	}
}

func testLogWithArguments(b *testing.B, logger log.Logger) {
	for i := 0; i < b.N; i++ {
		logger.Info("info message", generateField(i)...)
	}
}

func testError(b *testing.B, logger log.Logger) {
	for i := 0; i < b.N; i++ {
		err := errors.New("oops")
		logger.Error("something went wrong", err, generateField(i)...)
	}
}

func generateField(i int) []log.LogField {
	return []log.LogField{
		log.Field("a", "1"),
		log.Field("b", "2"),
		log.Field("c", "3"),
		log.Field("d", "4"),
		log.Field("e", "5"),
		log.Field("f", "s1"),
		log.Field("g", "s2"),
		log.Field("h", "s3"),
		log.Field("i", "s4"),
		log.Field("j", "s5"),
	}
}

type logObserver struct {
	entries []string
}

func (l *logObserver) Write(m []byte) (n int, err error) {
	l.entries = append(l.entries, string(m))
	return len(m), nil
}

func TestSimpleLogger_Debug(t *testing.T) {
	observer := &logObserver{}
	logger := log.NewSimpleLogger(observer)

	writeLog(logger, log.DebugLogLevel)
	assert.Equal(t, observer.entries, 1)
	assert.Equal(t, "Hello", observer.entries[0])
}

func writeLog(logger log.Logger, lv log.LogLevel, fields ...log.LogField) {
	switch lv {
	case log.DebugLogLevel:
		logger.Debug("Hello", fields...)
	case log.InfoLogLevel:
		logger.Info("Hello", fields...)
	case log.WarnLogLevel:
		logger.Warn("Hello", fields...)
	}
}

type noopWriter struct{}

func (w *noopWriter) Write(m []byte) (n int, err error) {
	return len(m), nil
}

func BenchmarkSimpleLogger(b *testing.B) {
	logger := log.NewSimpleLogger(&noopWriter{})

	b.ResetTimer()
	b.Run("log", func(b *testing.B) {
		testLog(b, logger)
	})

	b.ResetTimer()
	b.Run("log with 10 fields", func(b *testing.B) {
		testLogWithArguments(b, logger)
	})

	b.ResetTimer()
	b.Run("log error with 10 fields", func(b *testing.B) {
		testError(b, logger)
	})
}