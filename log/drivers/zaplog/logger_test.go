package zaplog_test

import (
	"bufio"
	"context"
	"errors"
	"os"
	"testing"

	"github.com/hexastack-dev/devkit-go/log"
	"github.com/hexastack-dev/devkit-go/log/drivers/zaplog"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

type noopWriter struct{}

func (*noopWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func TestLogger_Debug(t *testing.T) {
	core, observedLogs := observer.New(zap.DebugLevel)
	logger := zaplog.New(zap.New(core))
	writeLog(logger, log.DebugLogLevel)

	assert.Equal(t, 1, observedLogs.Len())
	assert.Equal(t, "Hello", observedLogs.All()[0].Message)
	assert.Equal(t, zap.DebugLevel, observedLogs.All()[0].Level)
}

func TestLogger_Info(t *testing.T) {
	core, observedLogs := observer.New(zap.InfoLevel)
	logger := zaplog.New(zap.New(core))
	writeLog(logger, log.DebugLogLevel)
	writeLog(logger, log.InfoLogLevel)

	assert.Equal(t, observedLogs.Len(), 1)
	assert.Equal(t, "Hello", observedLogs.All()[0].Message)
	assert.Equal(t, zap.InfoLevel, observedLogs.All()[0].Level)
}

func TestLogger_Warn(t *testing.T) {
	core, observedLogs := observer.New(zap.InfoLevel)
	logger := zaplog.New(zap.New(core))
	writeLog(logger, log.WarnLogLevel)

	assert.Equal(t, observedLogs.Len(), 1)
	assert.Equal(t, "Hello", observedLogs.All()[0].Message)
	assert.Equal(t, zap.WarnLevel, observedLogs.All()[0].Level)
}

func TestLogger_Error(t *testing.T) {
	core, observedLogs := observer.New(zap.InfoLevel)
	logger := zaplog.New(zap.New(core))
	writeErrorLog(logger, errors.New("oopsie"))

	assert.Equal(t, observedLogs.Len(), 1)
	assert.Equal(t, "Something went wrong", observedLogs.All()[0].Message)
	assert.Equal(t, zap.ErrorLevel, observedLogs.All()[0].Level)
	assert.Equal(t, "oopsie", observedLogs.All()[0].ContextMap()["error"])
}

func TestLogger_WithFields(t *testing.T) {
	core, observedLogs := observer.New(zap.InfoLevel)
	logger := zaplog.New(zap.New(core))
	writeLog(logger, log.InfoLogLevel, log.Field("v1", "value1"), log.Field("v2", "value2"))

	assert.Equal(t, observedLogs.Len(), 1)
	assert.Equal(t, "Hello", observedLogs.All()[0].Message)
	assert.Equal(t, zap.InfoLevel, observedLogs.All()[0].Level)
	assert.Equal(t, "value1", observedLogs.All()[0].ContextMap()["v1"])
	assert.Equal(t, "value2", observedLogs.All()[0].ContextMap()["v2"])
}

func TestLogger_WithContext(t *testing.T) {
	ctx := context.Background()
	tp := trace.NewTracerProvider()
	ctx, span := tp.Tracer("").Start(ctx, "testWithContext")
	defer span.End()

	core, observedLogs := observer.New(zap.InfoLevel)
	logger := zaplog.New(zap.New(core))
	logger.WithContext(ctx).Info("With span", log.Field("v1", "value1"))

	assert.Equal(t, observedLogs.Len(), 1)
	assert.Equal(t, "With span", observedLogs.All()[0].Message)
	assert.Equal(t, zap.InfoLevel, observedLogs.All()[0].Level)
	assert.Equal(t, "value1", observedLogs.All()[0].ContextMap()["v1"])
	assert.NotEmpty(t, observedLogs.All()[0].ContextMap()["spanId"])
	assert.NotEmpty(t, observedLogs.All()[0].ContextMap()["traceId"])
	assert.NotEmpty(t, observedLogs.All()[0].ContextMap()["traceFlags"])
}

func TestLogger_WriteToFile(t *testing.T) {
	err := os.RemoveAll("./log")
	if err != nil {
		t.Fatal(err)
	}

	logger := zaplog.NewDefaultLogger(zaplog.Config{
		RootLogLevel: log.InfoLogLevel,
		FileLogConfig: zaplog.FileLogConfig{
			Enabled:    true,
			Filename:   "./log/test.log",
			MaxSize:    1, // 1MB
			MaxBackups: 5,
		},
	})
	logger.Debug("Some debug")
	logger.Info("Some info")
	logger.Warn("Some warn")
	logger.Error("some error", errors.New("oopsie"))

	file, err := os.Open("./log/test.log")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	sc := bufio.NewScanner(file)
	sc.Split(bufio.ScanLines)

	var lines []string
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}

	// RootLogLevel set to INFO, thus Debug should not be logged
	assert.Equal(t, 3, len(lines))
	for _, line := range lines {
		assert.Greater(t, len(line), 1)
	}
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

func writeErrorLog(logger log.Logger, err error, fields ...log.LogField) {
	logger.Error("Something went wrong", err, fields...)
}

func BenchmarkLogger(b *testing.B) {
	logger := zaplog.NewDefaultLogger(zaplog.Config{
		RootLogLevel: log.InfoLogLevel,
		Output:       &noopWriter{},
	})

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
